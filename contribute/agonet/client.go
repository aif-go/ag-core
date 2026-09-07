package agonet

import (
	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"

	// "github.com/tjfoc/gmsm/gmtls"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/valyala/bytebufferpool"
	"golang.org/x/sync/errgroup"
)

type Client interface {
	Start() error
	Stop() error
	Dial(network, addr string) (Conn, error)
	DialContext(network, addr string, ctx any) (Conn, error)
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
	opts, err := BuildOptionsWithConfig(config.Config)
	if err != nil {
		return nil, err
	}

	// 配置TLS
	secCfg := config.Config.Security
	// 注意：Type=tls_tlcp 也必须进入客户端 TLS 配置分支，
	// 由 WithAgClientTLSConfig 内部将 tls_tlcp 归一化为客户端可用的 TLS/TLCP（见 options_tls.go）。
	if secCfg.Type != TLSType_NONE && secCfg.Type != TLSType_UNSET {
		err := ExtendOptions(opts, WithAgClientTLSConfig(&secCfg))
		if err != nil {
			return nil, err
		}
	}

	return NewClientWithOptions(handler, opts)
}

func NewClientWithOptions(handler EventHandler, opts *Options) (Client, error) {
	// 配置自洽校验：对 CliTLSType() fallback 解析后的类型校验（保留客户端复用服务端配置的用法）
	if err := opts.ValidateClient(); err != nil {
		return nil, err
	}
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

		isClient: true,
	}

	eng.eventLoops = new(leastConnectionsLoadBalancer)
	cli.eng = &eng

	return cli, nil
}

func (cli *client) Start() error {
	numEventLoop := determineEventLoops(cli.opts)
	slog.Info(fmt.Sprintf("Starting agonet client with %d event loops", numEventLoop))

	cli.eng.isClient = true

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
	cliTlsType := cli.opts.CliTLSType()

	switch cliTlsType {
	case TLSType_NONE:
		c, err = net.Dial(network, addr)
	case TLSType_TLS:
		tlsCfg := cli.opts.CliTLSConfig()
		if tlsCfg == nil {
			return nil, aerrors.ErrTLSConfigIsNil
		}
		c, err = tls.Dial(network, addr, tlsCfg)
	case TLSType_TLCP:
		tlcpCfg := cli.opts.CliTLCPConfig()
		if tlcpCfg == nil {
			return nil, aerrors.ErrTLCPConfigIsNil
		}
		c, err = tlcp.Dial(network, addr, tlcpCfg)

	default:
		c, err = net.Dial(network, addr)
		// return nil, aerrors.ErrUnsupportedProtocol
	}

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
	if el == nil {
		// 客户端未 Start（eventloops 为空）时 next 返回 nil，返回明确错误而非越界 panic
		return nil, aerrors.ErrInvalidNetConn
	}
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

	if cli.opts.KeepAlive.Enable {
		err := cli.applyKeepAlive(nc)
		if err != nil {
			return nil, err
		}
	}

	c := newStreamConn(el, nc, ctx)

	if !el.send(&openConn{c: c, cb: func() { close(connOpened) }}) { // R2：openConn send 化
		nc.Close()
		close(connOpened) // 关键边界：open 不会执行 → cb 不会被调 → 防 Dial 永久挂起（A3）
		return nil, aerrors.ErrEngineShutdown
	}

	// R1：读 goroutine 出池（原生 goroutine，不再占全局池）。
	// A7 触发链消失：不再有 Submit 失败 → 客户端永久挂起。
	go func() {
		var buffer [0x10000]byte // B2 另案（64KB 缓冲实为逃逸堆，非栈空间）
		for {
			// 监听连接读取数据
			n, err := nc.Read(buffer[:])

			if err != nil {
				// 处理读取错误
				el.send(&netErr{c, err}) // R2：错误路径 send 化
				return
			}
			// 6. 触发连接读取事件
			tc2 := packTCPConn(c, buffer[:n])
			if !el.send(tc2) { // R2：数据路径 send 化，引擎关闭时归还 ByteBuffer
				bytebufferpool.Put(tc2.b)
				return
			}
		}
	}()
	gc = c

	<-connOpened

	return
}

func (cli *client) applyKeepAlive(nc net.Conn) error {
	keepOpt := cli.opts.KeepAlive
	if !keepOpt.Enable || keepOpt.Idle <= 0 {
		return nil
	}

	tc := nc

	switch bc := nc.(type) {
	case *net.UnixConn: // 支持 Unix 域套接字连接
		return nil
	case *net.TCPConn: // 支持 TCP 连接
	case *tls.Conn: // 支持 TLS 连接
		tc = bc.NetConn()
	case *tlcp.Conn: // 支持 TLCP 连接
		tc = bc.NetConn()
	default:
		return aerrors.ErrUnsupportedProtocol
	}

	if kpAblity, ok := tc.(KeepAliveAbility); ok {
		keepAlive := buildKeepAliveWithConfig(keepOpt)
		if keepAlive != nil {
			err := kpAblity.SetKeepAliveConfig(*keepAlive)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
