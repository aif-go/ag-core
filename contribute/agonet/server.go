package agonet

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type Server interface {
	Start() error
	Stop() error
}

func NewServer(handler EventHandler, config *ServerConfig) (Server, error) {
	opts, err := BuildOptionsWithConfig(config.Config)
	if err != nil {
		return nil, err
	}

	// 配置TLS
	secCfg := config.Config.Security
	if secCfg.Type != TLSType_NONE && secCfg.Type != TLSType_UNSET {
		err := ExtendOptions(opts, WithAgTLSConfig(&secCfg))
		if err != nil {
			return nil, err
		}
	}

	addrs := make([]string, 0)
	addrs = append(addrs, config.Addr)

	return NewServerWithOptions(handler, addrs, opts)
}

func NewServerWithOptions(handler EventHandler, addr []string, opts *Options) (Server, error) {
	// 配置自洽校验：Type 声明与 config 对应，错误在构造期暴露（fail-fast）
	if err := opts.ValidateServer(); err != nil {
		return nil, err
	}

	ser := &server{
		addrs:        addr,
		opts:         opts,
		eventHandler: handler,
	}

	return ser, nil
}

type server struct {
	// config       *ServerConfig
	addrs        []string
	opts         *Options
	eng          *engine
	eventHandler EventHandler
}

func (s *server) Start() error {
	return s.run()
}

func (s *server) Stop() error {
	if s.eng == nil {
		return nil // 未启动则幂等返回，避免 nil 解引用 panic
	}
	s.eng.shutdown(nil)
	return nil
}

func (s *server) run() error {
	if s.eventHandler == nil {
		return fmt.Errorf("agonet: no event handler")
	}

	addrs := s.addrs
	opts := s.opts

	if addrs == nil || len(addrs) == 0 {
		return fmt.Errorf("agonet: no address")
	}

	// createListeners
	lns, err := createListeners(addrs, opts)
	if err != nil {
		return err
	}

	// lns := make([]net.Listener, 0, len(listeners))
	// for _, ln := range listeners {
	// 	lns = append(lns, ln)
	// }

	rootCtx, shutdown := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(rootCtx)

	eng := engine{
		addrs:        addrs,
		opts:         opts,
		listeners:    lns,
		eventHandler: s.eventHandler,
		turnOff:      shutdown,
		concurrency: struct {
			*errgroup.Group
			ctx context.Context
		}{eg, ctx},
	}

	// create event-loops
	eng.eventLoops = new(roundRobinLoadBalancer)

	s.eng = &eng

	e := Engine{
		eng: &eng,
	}

	switch eng.eventHandler.OnBoot(e) {
	case None:
	case Close:
	case Shutdown:
		return nil // 引导事件返回关闭或关闭引擎，直接返回
	}

	err = eng.start(ctx)
	if err != nil {
		// FXME 启动失败操作
		return err
	}

	// 阻塞运行直到 Stop()（eng.shutdown → turnOff → ctx 取消）：
	// 保持阻塞式启动语义（与 ag_app 的 `go srv.Start()` 用法配合）。
	// 注意不能用 defer：defer 在 return 前执行，启动失败路径也会触发阻塞等待。
	eng.stop(rootCtx, e)

	return nil
}
