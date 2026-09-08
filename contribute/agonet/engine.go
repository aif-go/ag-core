package agonet

import (
	"context"
	"errors"
	"fmt"
	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
	"log/slog"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/valyala/bytebufferpool"
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

// maxConns 连接上限（0 = 不限制）。
func (eng *engine) maxConns() int32 {
	return eng.opts.MaxConn
}

// totalConns 当前连接总数（各 loop 原子计数累加）。
// 并发 open（多 loop 同时接入）时可能瞬时超限——护栏语义（配额非精确边界），与 Netty
// 社区 channelActive 计数实现一致；open 前置检查在同一 loop 内串行，不会重复注册。
func (eng *engine) totalConns() int32 {
	var total int32
	eng.eventLoops.iterate(func(_ int, el *eventloop) bool {
		total += el.countConn()
		return true
	})
	return total
}

// shutdownTimeout 优雅关闭 drain 超时（Options.ShutdownTimeout，0 = 默认 5s）。
// A1 兜底：防慢 handler 无限拖延关闭；drain 超时后 loop 强制退出。
func (eng *engine) shutdownTimeout() time.Duration {
	if eng.opts.ShutdownTimeout <= 0 {
		return 5 * time.Second
	}
	return eng.opts.ShutdownTimeout
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
		if !el.sendNonBlocking(oconn) { // R4 跳满：ch 满或引擎关闭 → 快速失败（关连接继续 accept），
			// accept 循环不被业务背压拖死（A1 残留收口）；引擎关闭时不滞留连接
			tc.Close()
			continue
		}

		// R1：读 goroutine 出池（原生 goroutine，不再占全局池）。
		// B1 触发链根除：池满 → Submit err → shutdown 整服关 不再可能由读路径触发。
		// goroutine 数 = 连接数（fd 限制兜底；R6 MaxConn 应用层配额暂缓）。
		go func() {
			var buffer [0x10000]byte // B2 另案（本链不动）
			for {
				// 监听连接读取数据
				n, err := tc.Read(buffer[:])

				if err != nil {
					// 处理读取错误
					el.send(&netErr{c, err}) // R2：错误路径 send 化
					return
				}
				// 触发连接读取事件
				tc2 := packTCPConn(c, buffer[:n])
				if !el.send(tc2) { // R2：数据路径 send 化，引擎关闭时归还 ByteBuffer
					bytebufferpool.Put(tc2.b)
					return
				}

			}
		}()
	}
}

func (eng *engine) closeEventLoops() {
	eng.eventLoops.iterate(func(i int, el *eventloop) bool {
		// A1：loop 退出不再依赖 ch 信号（run 监听 ctx.Done + drain）——
		// 裸投递 el.ch <- ErrEngineShutdown 在 ch 满时阻塞 → Stop 卡死。
		// trySend 尽力兼容（失败无碍，loop 靠 ctx 退出）。
		el.trySend(aerrors.ErrEngineShutdown)
		return true
	})
	for _, ln := range eng.listeners {
		ln.close()
	}
}
