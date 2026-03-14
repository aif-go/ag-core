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

	eng.eventHandler.OnShutdown(engine)

	eng.closeEventLoops()

	// TODO 关闭操作
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
		// 1. 等待客户端连接
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

		// // 2. 初始化连接相关参数
		// // FIXME 初始化连接相关参数
		// tcpConn, ok := tc.(*net.TCPConn)
		// if ok {
		// 	// 开启 TCP 连接的 KeepAlive 功能 TODO 参数控制
		// 	tcpConn.SetKeepAlive(true)
		// 	tcpConn.SetKeepAlivePeriod(30 * time.Second)
		// }

		// 3. 组装连接对象
		c := newStreamConn(tc, nil)

		// 4. 触发连接打开事件
		out, action := eng.eventHandler.OnOpen(c) // 连接打开事件
		if out != nil {
			if _, err := tc.Write(out); err != nil {
				return err
			}
		}
		// action 处理
		switch action {
		case None:
		case Close:
			eng.closeConn(tc)
		case Shutdown:
			return aerrors.ErrEngineShutdown // TODO 此操作要关闭引擎
		default:
			// return nil
		}

		// 5. 启动 goroutine 处理单个客户端连接（支持多客户端并发）

		goroutine.DefaultWorkerPool.Submit(func() {
			var buffer [0x10000]byte
			for {
				n, err := tc.Read(buffer[:])
				if err != nil {
					// TODO 处理读取错误
					return
				}
				fmt.Sprintf("Received message: %s", buffer[:n]) // TODO
				// 6. 触发连接读取事件
				packTCPConn(c, buffer[:n]) // TODO
				// el.ch <- packTCPConn(c, buffer[:n])
			}
		})
	}
}

func (eng *engine) closeConn(conn net.Conn) error {
	// TODO 关闭连接
	return nil
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
