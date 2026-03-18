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
	opts, err := buildOptionsWithConfig(config.Config)
	if err != nil {
		return nil, err
	}

	addrs := make([]string, 0)
	addrs = append(addrs, config.Address)

	return NewServerWithOptions(handler, addrs, opts)

}
func NewServerWithOptions(handler EventHandler, addr []string, opts *Options) (Server, error) {
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
	s.eng.shutdown(nil)
	return nil
}

func (s *server) run() error {
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
