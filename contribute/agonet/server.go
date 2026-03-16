package agonet

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Server interface {
	Start() error
	Stop() error
}

func NewServer(handler EventHandler, config ServerConfig) Server {
	return &server{
		config:       config,
		eventHandler: handler,
	}
}

type server struct {
	config       ServerConfig
	opts         *Options
	eng          *engine
	eventHandler EventHandler
}

func (s *server) Start() error {
	addrs := make([]string, 0)
	addrs = append(addrs, s.config.Address)
	return s.run(addrs)
}

func (s *server) Stop() error {
	s.eng.shutdown(nil)
	return nil
}

func (s *server) run(addrs []string) error {

	opts := s.buildOptionsWithConfig()
	s.opts = opts

	// createListeners
	lns, err := createListeners(addrs, opts)
	if err != nil {
		return err
	}

	defer func() {
		for _, ln := range lns {
			ln.close()
		}
	}()

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

	defer eng.stop(rootCtx, e) // 等待上下文取消，触发关闭操作

	return nil
}

func (s *server) buildOptionsWithConfig() *Options {

	// opts := &Options{
	// 	NumEventLoop: s.config.Engine.NumEventLoop,
	// 	Multicore:    s.config.Engine.Multicore,
	// 	Ticker:       s.config.Engine.Ticker,
	// 	KeepAlive: struct {
	// 		Enable   bool
	// 		Idle     time.Duration
	// 		Interval time.Duration
	// 		Count    int
	// 	}{
	// 		Enable:   s.config.KeepAlive.Enable,
	// 		Idle:     time.Duration(s.config.KeepAlive.Idle) * time.Second,
	// 		Interval: time.Duration(s.config.KeepAlive.Interval) * time.Second,
	// 		Count:    s.config.KeepAlive.Count,
	// 	},
	// }
	opts := buildOptionsWithConfig(s.config.Config)
	return opts
}
