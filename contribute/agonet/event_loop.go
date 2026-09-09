package agonet

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
	goroutine "github.com/aif-go/ag-core/contribute/agonet/pkg/pool/goroutline"

	"github.com/petermattis/goid"
)

type eventloop struct {
	ch           chan any           // channel for event-loop
	idx          int                // index of event-loop in event-loops
	eng          *engine            // engine in loop
	connCount    int32              // number of active connections in event-loop
	connections  map[*conn]struct{} // TCP connection map: fd -> conn
	eventHandler EventHandler       // user eventHandler

	goroutineId atomic.Int64
}

// safeHandle 包装事件处理（D6 层 2：func/写等无连接上下文的事件——panic 记录 + 返回
// 非致命错误——loop 继续；连接事件已由 read 层 recover 覆盖）
func (el *eventloop) safeHandle(i any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("event-loop panic recovered", "event", fmt.Sprintf("%T", i),
				"panic", r, "stack", string(debug.Stack()))
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()
	return el.handleEvent(i)
}

func (el *eventloop) run() (err error) {
	defer func() {
		el.eng.shutdown(err)
		for c := range el.connections {
			_ = el.close(c, nil)
		}
	}()

	// 绑定事件循环到当前线程
	if el.eng.opts.LockOSThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	// 获取协程id
	id := goid.Get()
	el.goroutineId.Store(id)
	slog.Debug(fmt.Sprintf("event-loop(%d) is running, gid: %d", el.idx, el.goroutineId.Load()))

	for {
		select {
		case i := <-el.ch:
			err = el.safeHandle(i) // D6：panic 隔离（记录——loop 继续）
			if errors.Is(err, aerrors.ErrEngineShutdown) {
				// el.getLogger().Debugf("event-loop(%d) is exiting in terms of the demand from user, %v", el.idx, err)
				slog.Error(fmt.Sprintf("event-loop(%d) is exiting in terms of the demand from user, %v", el.idx, err))
				return nil
			} else if err != nil {
				// el.getLogger().Debugf("event-loop(%d) got a nonlethal error: %v", el.idx, err)
				slog.Error(fmt.Sprintf("event-loop(%d) got a nonlethal error: %v", el.idx, err))
			}
		case <-el.eng.concurrency.ctx.Done():
			// A1 优雅关闭：引擎关闭信号直达 loop（不依赖 ch 信号投递——裸投递在
			// ch 满时阻塞 → Stop 卡死）。先完成在途（drain ch 存量），超时兜底。
			slog.Debug(fmt.Sprintf("event-loop(%d) is draining on engine shutdown", el.idx))
			el.drain()
			return nil
		}
	}
}

// handleEvent 处理单个事件（run 与 drain 共用）。
func (el *eventloop) handleEvent(v any) error {
	switch i := v.(type) {
	case error:
		return i
	case *netErr:
		return el.close(i.c, i.err)
	case *openConn:
		return el.open(i)
	case *tcpConn:
		return el.read(unpackTCPConn(i))
	case func() error:
		return i()
	}
	return nil
}

// drain 同步处理队列在途事件（loop goroutine 内调用——run 主循环已退出，loop 空闲——
// 自身处理天然无并发消费者）；队列空返回（优雅完成）或超时返回（丢弃剩余——强关）。
// P1（review 修正）：同步化替代原 goroutine 版——修复前超时只让外层返回、内部 goroutine
// 继续 handleEvent，与 run 的 defer 清理并发（破坏单线程所有权 + use-after-free）。
// 边界：deadline 检查在每事件间——正在执行的 handler 无法打断（模型固有——文档化）。
func (el *eventloop) drain() {
	deadline := time.Now().Add(el.eng.shutdownTimeout())
	for {
		select {
		case v := <-el.ch:
			el.safeHandle(v) // D6 对称：run 与 drain 消费同一 ch——panic 隔离（记录 + 继续），逃逸会击穿 run() 无 recover defer
		default:
			return // 队列空——优雅完成
		}
		if time.Now().After(deadline) {
			el.rejectRemaining() // review 修正：超时丢弃剩余——openConn 必须完成通知 + 关连接（防 Dial 永久挂起）
			return               // 超时——丢弃剩余（强关语义——ShutdownTimeout 生效）
		}
	}
}

// rejectRemaining 强关超时后的剩余队列：openConn 未处理 → 关 rawConn + cb(ErrEngineShutdown)
// 完成通知（防 client EnrollContext 的 connOpened 永久等待——Stop 已返回但 Dial 挂起）；
// 其余事件类型（netErr/tcpConn/func）无外部等待者，直接丢弃。
func (el *eventloop) rejectRemaining() {
	for {
		select {
		case v := <-el.ch:
			if oc, ok := v.(*openConn); ok {
				_ = oc.c.rawConn.Close()
				oc.c.release()
				if oc.cb != nil {
					oc.cb(aerrors.ErrEngineShutdown)
				}
			}
		default:
			return
		}
	}
}

func (el *eventloop) open(oc *openConn) error {
	c := oc.c

	// R6：连接上限前置检查（0 = 不限）。超限静默拒绝（护栏语义，Netty 社区
	// channelActive 计数 + close 同款）：关 rawConn → 读 goroutine Read 返回 err 自然退出；
	// 不注册、不触发 OnOpen/OnClose。
	// cb 仍需调用——客户端 EnrollContext 的 connOpened 边界（open 不执行 → 否则 Dial 挂起）。
	// P1（review 修正）：全局原子配额——Add(+1) 后判断（检查与递增原子合一——
	// 修复前 totalConns() 读时求和 + incConn 分离，多 loop 并发 open 竞态超限）
	if el.eng.maxConns() > 0 {
		if cur := atomic.AddInt32(&el.eng.totalConn, 1); cur > el.eng.maxConns() {
			atomic.AddInt32(&el.eng.totalConn, -1) // 超限回滚
			_ = c.rawConn.Close()
			c.release() // review 修正：释放被拒连接的缓冲资源
			if oc.cb != nil {
				oc.cb(aerrors.ErrMaxConnRejected) // review 修正：拒绝必须携带错误（防 client 把配额拒绝当连接成功）
			}
			return nil
		}
	}

	var openErr error
	if oc.cb != nil {
		defer func() { oc.cb(openErr) }() // 成功 nil；panic 时携带错误（Dial 返回失败而非伪成功）
	}

	el.connections[c] = struct{}{}
	el.incConn(1)

	// OnOpen panic → 关闭肇事连接（对齐 read() 的 D6 语义——半初始化连接不可信），
	// 经 close() 对称回收配额/注册/缓冲；cb 收到错误
	defer func() {
		if r := recover(); r != nil {
			openErr = fmt.Errorf("panic in OnOpen: %v", r)
			slog.Error("OnOpen panic recovered (conn closed)", "remote", c.rawConn.RemoteAddr(),
				"panic", r, "stack", string(debug.Stack()))
			_ = el.close(c, openErr)
		}
	}()

	out, action := el.eventHandler.OnOpen(c)
	if out != nil {
		if _, err := c.rawConn.Write(out); err != nil {
			return err
		}
	}

	return el.handleAction(c, action)
}

// resolveInboundLimit 解析入站滞留上限（Options → 实际值 + 防呆钳制）：
// 默认 16MB（正常滞留 = 半包帧 ≤ 解码器 maxFrameLength + 消费积压——16MB 为异常阈值）；
// 下限 1MB（太小误杀正常业务——业务消费慢也累积滞留）
func resolveInboundLimit(opts *Options) int {
	limit := defaultInboundLimit // 16MB
	if opts.InboundBufferLimit > 0 {
		limit = opts.InboundBufferLimit
	}
	if limit < minInboundLimit { // 1MB
		limit = minInboundLimit
	}
	return limit
}

const (
	defaultInboundLimit = 16 * 1024 * 1024
	minInboundLimit     = 1024 * 1024
)

// errInboundOverflow 入站滞留超限（F3：半包/慢速客户端——关闭连接释放内存）
var errInboundOverflow = errors.New("inbound buffer overflow (F3)")

// read 处理连接读事件（OnTraffic 调度 + 滞留写入）。
// D6：read 层 recover——handler panic 隔离（记录 + 关闭肇事连接——panic 时连接
// 处于半状态：buffer 半消费/缓存脏——留着后续事件会基于脏状态错误处理——必须关闭）。
// 修复前无 recover——handler panic 传播到进程——整服崩溃（T19 红态实证）。
func (el *eventloop) read(c *conn) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("read panic recovered (conn closed)", "remote", c.rawConn.RemoteAddr(),
				"panic", r, "stack", string(debug.Stack()))
			err = el.close(c, fmt.Errorf("panic in handler: %v", r))
		}
	}()
	if _, ok := el.connections[c]; !ok {
		return nil // ignore stale wakes.
	}
	// 调用消息处理函数
	action := el.eventHandler.OnTraffic(c)
	switch action {
	case None:
	case Close:
		return el.close(c, nil)
	case Shutdown:
		return aerrors.ErrEngineShutdown
	}

	// F3：入站滞留超限（半包/慢速客户端——inboundBuffer 无上限增长内存 DoS）→ 关闭连接
	// 滞留量 = inboundBuffer（跨包累积）+ buffer（当期未消费）——写前检查（防一次大写入超限）
	if c.inboundBuffer.Buffered()+c.buffer.Len() > resolveInboundLimit(el.eng.opts) {
		return el.close(c, errInboundOverflow)
	}

	// 剩余未处理的字节写入缓存
	_, err = c.inboundBuffer.Write(c.buffer.B) // FIXME elastic.RingBuffer 自动实现扩容

	if err != nil {
		// return el.close(c, err)
		// 判断异常，长度不够的要扩容inboundBuffer
		// FIXME elastic.RingBuffer 自动实现扩容
	}

	c.buffer.Reset()

	return nil
}

func (el *eventloop) wake(c *conn) error {
	if _, ok := el.connections[c]; !ok {
		return nil // ignore stale wakes.
	}
	action := el.eventHandler.OnTraffic(c)
	return el.handleAction(c, action)
}

func (el *eventloop) close(c *conn, err error) (retErr error) {
	_, ok := el.connections[c]
	if c.rawConn == nil || !ok {
		return nil // ignore stale wakes.
	}

	delete(el.connections, c)
	el.incConn(-1)
	atomic.AddInt32(&el.eng.totalConn, -1) // R6 全局配额递减（与 open Add 判断对称）

	// 资源收尾兜底：OnClose/handleAction 抛 panic 也保证 rawConn 关闭 + 缓冲归还
	//（review 修正：修复前 OnClose panic 中断后续 rawConn.Close/release——fd/缓冲泄漏——
	// 且连接已出 map，run defer 不再兜底）
	defer func() {
		_ = c.rawConn.Close()
		c.release()
	}()

	// OnClose 单独 recover：panic 记录 + 继续资源释放（对齐 read() 的 D6 隔离语义）
	action := Action(None)
	func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("OnClose panic recovered (conn closed)", "remote", c.rawConn.RemoteAddr(),
					"panic", r, "stack", string(debug.Stack()))
			}
		}()
		action = el.eventHandler.OnClose(c, err)
	}()

	return el.handleAction(c, action)
}

func (el *eventloop) incConn(delta int32) {
	atomic.AddInt32(&el.connCount, delta)
}

func (el *eventloop) countConn() int32 {
	return atomic.LoadInt32(&el.connCount)
}

func (el *eventloop) handleAction(c *conn, action Action) error {
	switch action {
	case None:
		return nil
	case Close:
		return el.close(c, nil)
	case Shutdown:
		return aerrors.ErrEngineShutdown
	default:
		return nil
	}
}

// ### eventloop implements EventLoop ###
var _ EventLoop = (*eventloop)(nil)

func (el *eventloop) Register(ctx context.Context, addr net.Addr) (<-chan RegisteredResult, error) {
	if el.eng.isShutdown() {
		return nil, aerrors.ErrEngineInShutdown
	}
	if addr == nil {
		return nil, aerrors.ErrInvalidNetworkAddress
	}

	// TODO
	return nil, nil
}

func (el *eventloop) Enroll(ctx context.Context, c net.Conn) (<-chan RegisteredResult, error) {
	if el.eng.isShutdown() {
		return nil, aerrors.ErrEngineInShutdown
	}
	// TODO
	return nil, nil
}

func (el *eventloop) Close(c Conn, err error) error {
	return el.close(c.(*conn), err)
}

// Deprecated
func (el *eventloop) InEventLoop() bool {
	// check goroutine id
	cid := goid.Get()

	return el.goroutineId.Load() == cid
}

// send 阻塞投递事件到事件循环：等待空位或引擎关闭信号。
// 返回 false 表示引擎已关闭，投递被放弃（调用方应自行清理资源）。
// R2/R3：统一投递抽象；复用 engine.concurrency.ctx（零新增 channel），R7=a 粒度已足够。
// 注意：先做非阻塞 ctx 检查再进入阻塞 select——避免"ch 有空位 + ctx 已关"时 select 随机
// 选中 ch 分支，把任务投进已退出的 loop（openConn 类任务将永不处理 → 调用方永久等待）。
func (el *eventloop) send(v any) bool {
	// 先查 ctx（非阻塞）再进入阻塞 select——避免"ch 有空位 + ctx 已关"时 select 随机
	// 选中 ch 分支，把任务投进已退出的 loop（openConn 类任务将永不处理 → 调用方永久等待）。
	// ctx.Err() 等价于非阻塞 select 查 Done（已取消返回非 nil），可读性更优。
	if el.eng.concurrency.ctx.Err() != nil {
		return false
	}
	select {
	case el.ch <- v:
		return true
	case <-el.eng.concurrency.ctx.Done():
		return false
	}
}

// sendNonBlocking 非阻塞投递：ch 满或引擎关闭 → false（调用方自行降级，如 accept 跳满）。
// 与 trySend 的区别：trySend 不检查引擎关闭（投成进已退 loop 由 ctx 兜底方负责）；
// sendNonBlocking 先查 ctx——accept 投递用此保证引擎关闭时快速失败、不滞留连接。
func (el *eventloop) sendNonBlocking(v any) bool {
	if el.eng.concurrency.ctx.Err() != nil {
		return false
	}
	select {
	case el.ch <- v:
		return true
	default:
		return false
	}
}

// trySend 非阻塞投递事件到事件循环。
// 返回 false 表示通道已满，调用方应降级处理（如转池 send）。
// 注意：不检查引擎关闭（非阻塞快速路径）——引擎关闭时若 ch 有空位仍会投成，
// 由失败后的转池 send（ctx.Done 兜底）保证 worker 不滞留。
func (el *eventloop) trySend(v any) bool {
	select {
	case el.ch <- v:
		return true
	default:
		return false
	}
}

// Execute executes the Runnable in the event-loop.
// eg :
//
//	  Execute(
//			context.Background(),
//			RunnableFunc(fn),
//		)
func (el *eventloop) Execute(ctx context.Context, runnable Runnable) error {

	if el.eng.isShutdown() {
		return aerrors.ErrEngineInShutdown
	}
	if runnable == nil {
		return aerrors.ErrNilRunnable
	}
	return goroutine.DefaultWorkerPool.Submit(func() {
		el.send(func() error { return runnable.Run(ctx) }) // R2：裸 el.ch <- → el.send（引擎关闭丢弃，worker 不滞留）
	})
}
