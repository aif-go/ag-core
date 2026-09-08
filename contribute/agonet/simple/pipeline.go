package simple

import (
	"fmt"
	"github.com/aif-go/ag-core/contribute/agonet/simple/utils"
	"log/slog"
	"sync"
	"sync/atomic"
)

var _ Pipeline = (*pipeline)(nil)

// NewPipeline create a pipeline.
func NewPipeline() Pipeline {

	p := &pipeline{}
	p.head = newHandlerContext(p, headHandler{}, nil, nil)
	p.tail = newHandlerContext(p, tailHandler{}, nil, nil)

	p.head.next.Store(p.tail)
	p.tail.prev.Store(p.head)

	// head + tail
	p.size = 2
	return p
}

// pipeline to implement Pipeline
type pipeline struct {
	mu      sync.Mutex // D1：结构变更互斥（AddFirst/AddLast），遍历无锁（原子读）
	head    *handlerContext
	tail    *handlerContext
	channel Channel
	size    int
	// handlingException D7：异常链防重入标志。
	// 异常处理器自身 panic 时保持 true（连接将关闭，后续异常无意义）；
	// 仅正常返回复位（不可用 defer 复位——panic 打断会错误复位导致循环）。
	handlingException atomic.Bool
}

func (p *pipeline) AddFirst(handlers ...Handler) Pipeline {
	// checking handler.
	checkHandler(handlers...)

	for _, h := range handlers {
		p.addFirst(h)
	}
	return p
}

func (p *pipeline) AddLast(handlers ...Handler) Pipeline {
	// checking handler.
	checkHandler(handlers...)

	for _, h := range handlers {
		p.addLast(h)
	}
	return p
}

func (p *pipeline) Channel() Channel {
	return p.channel
}

func (p *pipeline) ServeChannel(channel Channel) {
	p.channel = channel
}

func (p *pipeline) FireChannelActive() {
	p.head.FireActive()
}

func (p *pipeline) FireChannelInactive(ex error) {
	p.head.FireInactive(ex)
}

func (p *pipeline) FireChannelRead(message any) {
	p.head.FireRead(message)
}

func (p *pipeline) FireChannelWrite(message any) {
	p.tail.FireWrite(message)
}

// FireChannelException 触发异常链（D7 防重入：重入时忽略）。
func (p *pipeline) FireChannelException(ex error) {
	_ = p.TryFireException(ex)
}

// TryFireException 防重入异常链触发（返回语义：false = 重入被拒或异常链未正常完成，
// 调用方应强制关闭连接——如 OnTraffic 收尾 action=Close / invokeMethod Close）。
//
// 关键（D7 两个实现要点）：
//  1. panic 打断必须保持 flag=true——不可用 defer Store(false)（panic 触发 defer 复位
//     → recover 再进时 flag 变 false → 循环依旧）；只有正常返回才复位
//  2. 异常处理器自身 panic 在此 recover 接住（不崩 loop）——flag 保持 true，返回 false
func (p *pipeline) TryFireException(ex error) (ok bool) {
	if p.handlingException.Swap(true) {
		slog.Warn("exception chain re-entry ignored", "err", ex)
		return false
	}
	ok = true
	defer func() {
		if r := recover(); r != nil {
			// 异常处理器自身 panic：接住（不崩 eventloop）；flag 保持 true（连接将关闭）
			slog.Error("exception handler panicked", "err", r)
			ok = false
		}
	}()
	p.head.FireExceptionCaught(ex)
	p.handlingException.Store(false)
	return
}

func (p *pipeline) FireChannelEvent(event any) {
	p.head.FireEvent(event)
}

// bindHandler handler 入链前的共享性校验（D2）。
// 实现 UnshareableHandler 的 handler 只允许绑定一个 pipeline；重复入链 panic。
// 在 mu 锁内调用：TryBind 是有副作用的状态变更（CAS 置位），与链入须原子一致。
func (p *pipeline) bindHandler(handler Handler) {
	if uh, ok := handler.(UnshareableHandler); ok {
		if !uh.TryBind() {
			panic("handler not sharable: 已绑定其他 pipeline")
		}
	}
}

// addFirst to add handlers head
func (p *pipeline) addFirst(handler Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bindHandler(handler)
	oldNext := p.head.next.Load()
	ctx := newHandlerContext(p, handler, p.head, oldNext)
	p.head.next.Store(ctx)
	oldNext.prev.Store(ctx)
	p.size++
}

// addLast to add handlers tail
func (p *pipeline) addLast(handler Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bindHandler(handler)
	oldPrev := p.tail.prev.Load()
	ctx := newHandlerContext(p, handler, oldPrev, p.tail)
	p.tail.prev.Store(ctx)
	oldPrev.next.Store(ctx)
	p.size++
}

// checkHandler to checking handlers
func checkHandler(handlers ...Handler) {

	for index, h := range handlers {
		switch h.(type) {
		case InboundHandler:
		case OutboundHandler:
		case ExceptionHandler:
		case ActiveHandler:
		case InactiveHandlerFunc:
		case EventHandler:
		default:
			utils.Assert(fmt.Errorf("unrecognized Handler: %d:%T", index, h))
		}
	}
}
