package agonet

import (
	"ag-core/contribute/agonet/pkg/aerrors"
	goroutine "ag-core/contribute/agonet/pkg/pool/goroutline"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"

	// "github.com/tjfoc/gmsm/gmtls"

	"gitee.com/Trisia/gotlcp/tlcp"
	"golang.org/x/sync/errgroup"
)

type Client interface {
	Start() error
	Stop() error
	Dial(network, addr string) (Conn, error)
	// DialContext(network, addr string, ctx any) (Conn, error)
	// Enroll(nc net.Conn) (Conn, error)
	// EnrollContext(nc net.Conn, ctx any) (Conn, error)
}

type client struct {
	// config       ClientConfig
	opts         *Options
	eng          *engine
	eventHandler EventHandler
}

func NewClient(handler EventHandler, config *ClientConfig) (Client, error) {
	opts, err := buildOptionsWithConfig(config.Config)
	if err != nil {
		return nil, err
	}

	return NewClientWithOptions(handler, opts)
}

func NewClientWithOptions(handler EventHandler, opts *Options) (Client, error) {
	cli := &client{
		eventHandler: handler,
	}

	cli.opts = opts

	rootCtx, shutdown := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(rootCtx)

	eng := engine{
		addrs:        []string{},
		opts:         opts,
		listeners:    []*listener{},
		eventHandler: cli.eventHandler,
		turnOff:      shutdown,
		concurrency: struct {
			*errgroup.Group
			ctx context.Context
		}{eg, ctx},
	}

	eng.eventLoops = new(leastConnectionsLoadBalancer)
	cli.eng = &eng

	return cli, nil
}

func (cli *client) Start() error {
	numEventLoop := determineEventLoops(cli.opts)
	slog.Info(fmt.Sprintf("Starting gnet client with %d event loops", numEventLoop))

	cli.eng.eventHandler.OnBoot(Engine{cli.eng})

	for i := 0; i < numEventLoop; i++ {
		el := eventloop{
			ch:           make(chan any, 1024),
			eng:          cli.eng,
			connections:  make(map[*conn]struct{}),
			eventHandler: cli.eng.eventHandler,
		}
		cli.eng.eventLoops.register(&el)
		cli.eng.concurrency.Go(el.run)
	}

	return nil
}

func (cli *client) Stop() error {
	cli.eng.shutdown(nil)

	cli.eng.eventHandler.OnShutdown(Engine{cli.eng})

	// Notify all event-loops to exit.
	cli.eng.closeEventLoops()

	// Wait for all event-loops to exit.
	err := cli.eng.concurrency.Wait()

	// Put the engine into the shutdown state.
	cli.eng.inShutdown.Store(true)

	return err
}

func (cli *client) Dial(network, addr string) (Conn, error) {
	return cli.DialContext(network, addr, nil)
}

func (cli *client) DialContext(network, addr string, ctx any) (Conn, error) {
	var (
		c   net.Conn
		err error
	)
	// c, err = net.Dial(network, addr)
	switch cli.opts.TLSType {
	case TLSTypeNone:
		c, err = net.Dial(network, addr)
	case TLSTypeTLS:
		if cli.opts.TLSConfig == nil {
			return nil, aerrors.ErrTLSConfigIsNil
		}
		c, err = tls.Dial(network, addr, cli.opts.TLSConfig)
	case TLSTypeTLCP:
		if cli.opts.TLCPConfig == nil {
			return nil, aerrors.ErrTLCPConfigIsNil
		}
		c, err = tlcp.Dial(network, addr, cli.opts.TLCPConfig)
	default:
		return nil, aerrors.ErrUnsupportedProtocol
	}

	// c, err = tls.Dial(network, addr, cli.opts.TLSConfig)
	if err != nil {
		return nil, err
	}

	return cli.EnrollContext(c, ctx)
}

func (cli *client) Enroll(nc net.Conn) (gc Conn, err error) {
	return cli.EnrollContext(nc, nil)
}

func (cli *client) EnrollContext(nc net.Conn, ctx any) (gc Conn, err error) {
	el := cli.eng.eventLoops.next(nil)
	connOpened := make(chan struct{})

	// 不支持的协议判断，支持tpc4、tls、tlcp 等
	// switch v := nc.(type) {
	switch nc.(type) {
	case *net.UnixConn: // 支持 Unix 域套接字连接
	case *net.TCPConn: // 支持 TCP 连接
	case *tls.Conn: // 支持 TLS 连接
		// case *gmtls.Conn: // 支持 gmtls实现的国密TLCP连接
	case *tlcp.Conn: // 支持 TLCP 连接
	default:
		return nil, aerrors.ErrUnsupportedProtocol
	}

	c := newStreamConn(el, nc, ctx)

	// if tcpc, ok := nc.(*net.TCPConn); ok {
	if kpAblity, ok := nc.(KeepAliveAbility); ok {
		options := cli.opts
		if options.KeepAlive.Enable && options.KeepAlive.Idle > 0 {
			keepAlive := buildKeepAliveWithConfig(options.KeepAlive)
			if keepAlive != nil {
				err := kpAblity.SetKeepAliveConfig(*keepAlive)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	el.ch <- &openConn{c: c, cb: func() { close(connOpened) }}

	goroutine.DefaultWorkerPool.Submit(func() {
		var buffer [0x10000]byte // 64KB 栈空间，不使用堆内存
		for {
			// 监听连接读取数据
			n, err := nc.Read(buffer[:])

			if err != nil {
				// 处理读取错误
				el.ch <- &netErr{c, err}
				return
			}
			// 6. 触发连接读取事件
			el.ch <- packTCPConn(c, buffer[:n])
		}

		// for {
		// 	_, err := c.rawReader.Peek(0)
		// 	// _, err := c.Reader.Peek(0)
		// 	if err != nil {
		// 		// 处理读取错误
		// 		el.ch <- &netErr{c, err}
		// 		return
		// 	}
		// 	// el.ch <- packTCPConn(c, buffer[:n])
		// 	fmt.Println("read data=====")
		// 	el.ch <- packTCPConn(c, nil)
		// }

		// var buffer [0x10000]byte
		// for {
		// 	n, err := nc.Read(buffer[:])
		// 	if err != nil {
		// 		el.ch <- &netErr{c, err}
		// 		return
		// 	}
		// 	el.ch <- packTCPConn(c, buffer[:n])
		// }
	})
	gc = c

	<-connOpened

	return
}
