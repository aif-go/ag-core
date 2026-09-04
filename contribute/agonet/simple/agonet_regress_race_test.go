package simple

// 并发竞态回归测试（white-box, package simple）：D1/D2（simple 包部分）。
// 对应跟踪清单：agonet-问题总报告与跟踪.md -> D1 / D2。
// 语义：期望"无 data race"（链5 修复后转正，删 agonet_regress tag）。
// 必须 -race 运行才有效：go test -race ./simple/ -run 'TestD1|TestD2' -v

import (
	"sync"
	"testing"
)

// TestD1_PipelineConcurrentModifyShouldBeSafe 缺陷复现
//
// 缺陷：pipeline 双向链表（pipeline.go:86-101）无锁；业务 goroutine 在 active 期 AddLast（如
//       RequestSync 每次 AddLast promiseHandler）修改 prev/next，与 eventloop 内 FireChannelRead
//       遍历（context.go:131-144）并发 -> 链表指针读写 data race / 链表损坏。
// 期望：pipeline 读写加锁（或禁止活跃期外部修改），并发遍历+修改无 race。
// 注：遍历用指针级 walk（等价于 FireChannelRead 的链表访问），避免 handler 传播的 O(n^2) 开销。
func TestD1_PipelineConcurrentModifyShouldBeSafe(t *testing.T) {
	p := NewPipeline().(*pipeline) // white-box：访问具体类型的 head/tail
	h := NewSimpleInboundHandler(func(ctx InboundContext, msg []byte) {})
	p.AddLast(h)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() { // 模拟 eventloop：遍历链表（读 next/prev 指针，贯穿到 tail 触及插入区）
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			for c := p.head.next.Load(); c != nil; c = c.next.Load() {
				_ = c.prev.Load()
			}
		}
	}()

	wg.Add(1)
	go func() { // 模拟业务 goroutine：修改链表（AddLast 插入到 tail 之前，写 prev/next）
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			p.AddLast(h)
		}
	}()

	wg.Wait()
	t.Log("PASS(D1 已修复): pipeline 并发遍历+AddLast 无 data race（当前 -race 下若 race/panic 则为红）")
}

// TestD2_IdleHandlerStateShouldBeSynchronized 缺陷复现
//
// 缺陷：idleHandler 状态字段（handlerCtx/lastReadTime/timers，handler_idle.go:61-83）无锁；
//       HandleActive/HandleInactive（eventloop 侧）写 handlerCtx、HandleRead/HandleWrite 写 lastReadTime，
//       time.AfterFunc 回调 onTimeoutInEL（timer goroutine）读 handlerCtx/timers -> data race；
//       且 Timer.Reset 在 timer goroutine 与 eventloop 两侧并发调用（Go 文档禁止）。
// 期望：状态字段加锁（或状态读写整体入 eventloop），跨 goroutine 访问无 race。
func TestD2_IdleHandlerStateShouldBeSynchronized(t *testing.T) {
	h := &idleHandler{} // 全部 idleTime=0 -> initialize 不创建定时器，仅测字段同步
	// D2：handlerCtx 为 atomic.Pointer[handlerContext]，用真实 pipeline 头节点作为 context（stub 类型不符）
	p := NewPipeline().(*pipeline)
	ctx := p.head

	var wg sync.WaitGroup

	wg.Add(1)
	go func() { // 模拟 eventloop：HandleActive + HandleRead 写状态
		defer wg.Done()
		h.HandleActive(ctx)
		for i := 0; i < 50000; i++ {
			h.HandleRead(ctx, []byte{1})
		}
	}()

	wg.Add(1)
	go func() { // 模拟 timer goroutine/关闭路径：HandleInactive 写 handlerCtx + onTimeoutInEL 读
		defer wg.Done()
		for i := 0; i < 50000; i++ {
			h.HandleInactive(ctx, nil)
			_ = h.handlerCtx.Load() // 模拟 onTimeoutInEL 的 handlerCtx 读（D2：原子读）
		}
	}()

	wg.Wait()
	t.Log("PASS(D2 已修复): idleHandler 状态跨 goroutine 访问无 data race（当前 -race 下若 race 则为红）")
}
