package agonet

import (
	"ag-core/contribute/agonet/pkg/aerrors"
	goroutine "ag-core/contribute/agonet/pkg/pool/goroutline"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

type Engine struct {
	// eng is the internal engine struct.
	eng *engine
}

func (eng *Engine) IsClient() bool {
	return eng.eng.isClient
}

type engine struct {
	// listeners    map[string]net.Listener
	addrs []string
	// listeners  []net.Listener
	listeners  []*listener
	opts       *Options
	eventLoops loadBalancer // handling events

	inShutdown    atomic.Bool
	beingShutdown atomic.Bool
	turnOff       context.CancelFunc
	eventHandler  EventHandler
	concurrency   struct {
		*errgroup.Group

		ctx context.Context
	}

	isClient bool
}

func (eng *engine) isShutdown() bool {
	return eng.inShutdown.Load()
}

// shutdown signals the engine to shut down.
func (eng *engine) shutdown(err error) {
	if err != nil && !errors.Is(err, aerrors.ErrEngineShutdown) {
		slog.Error("engine is being shutdown with error", "err", err)
	}
	eng.turnOff() // 发送关闭信号
	eng.beingShutdown.Store(true)
}

func (e *engine) start(ctx context.Context) error {
	err := e.active(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (eng *engine) stop(ctx context.Context, engine Engine) {
	// Wait on a signal for shutdown
	<-ctx.Done()

	// 触发关闭事件
	eng.eventHandler.OnShutdown(engine)

	// 关闭事件循环
	eng.closeEventLoops()

	// 等待所有事件循环关闭
	if err := eng.concurrency.Wait(); err != nil && !errors.Is(err, aerrors.ErrEngineShutdown) {
		// eng.opts.Logger.Errorf("engine shutdown error: %v", err)
		slog.Error("engine shutdown error", "err", err)
	}

	// 标记引擎为关闭状态
	eng.inShutdown.Store(true)

}

func (eng *engine) active(ctx context.Context) error {
	// numEventLoop := eng.numEventLoop
	numEventLoop := determineEventLoops(eng.opts)

	slog.Info(fmt.Sprintf("Launching ag net with %d event-loops, listening on: %s",
		numEventLoop, strings.Join(eng.addrs, " | ")))

	// 初始化eventLoops
	for i := 0; i < numEventLoop; i++ {
		el := eventloop{
			ch:           make(chan any, 1024),
			eng:          eng,
			connections:  make(map[*conn]struct{}),
			eventHandler: eng.eventHandler,
		}
		eng.eventLoops.register(&el)
		eng.concurrency.Go(el.run)
	}

	for _, l := range eng.listeners {
		eng.concurrency.Go(func() error {
			return eng.listenStream(l.ln)
		})
	}

	return nil
}

// listenerAccept 监听并接受客户端连接
func (eng *engine) listenStream(listener net.Listener) (err error) {

	defer func() { eng.shutdown(err) }()

	// 循环接收客户端连接
	for {
		// 等待客户端连接
		tc, e := listener.Accept()
		if e != nil {
			err = e
			if !eng.beingShutdown.Load() {
				slog.Error("Accept() fails due to error", "err", err)
			} else if errors.Is(err, net.ErrClosed) {
				err = errors.Join(err, aerrors.ErrEngineShutdown) // 引擎关闭时，返回错误
			}
			return
		}

		// // 初始化连接相关参数
		// // FIXME 初始化连接相关参数
		// tcpConn, ok := tc.(*net.TCPConn)
		// if ok {
		// 	// 开启 TCP 连接的 KeepAlive 功能 TODO 参数控制
		// 	tcpConn.SetKeepAlive(true)
		// 	tcpConn.SetKeepAlivePeriod(30 * time.Second)
		// }

		el := eng.eventLoops.next(tc.RemoteAddr())

		// 组装连接对象
		c := newStreamConn(el, tc, nil)

		// // 触发连接打开事件
		oconn := &openConn{
			c: c,
		}
		el.ch <- oconn

		// 启动 goroutine 处理单个客户端连接（支持多客户端并发）
		err := goroutine.DefaultWorkerPool.Submit(func() {
			var buffer [0x10000]byte
			for {
				// 监听连接读取数据
				n, err := tc.Read(buffer[:])

				if err != nil {
					// 处理读取错误
					el.ch <- &netErr{c, err}
					return
				}
				// 触发连接读取事件
				el.ch <- packTCPConn(c, buffer[:n])

			}
		})
		if err != nil {
			return err
		}
	}
}

func (eng *engine) closeEventLoops() {
	eng.eventLoops.iterate(func(i int, el *eventloop) bool {
		// 每个eventloop发送关闭信号
		el.ch <- aerrors.ErrEngineShutdown
		return true
	})
	for _, ln := range eng.listeners {
		ln.close()
	}
}
