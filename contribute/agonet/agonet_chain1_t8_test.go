package agonet

// 链1 A1 优雅关闭测试（white-box，无 build tag 常驻）：
//   T8a server 侧 / T8b client 侧——背压下 Stop 不卡死 + 在途事件零丢失。
//
// 复现背压：填充 goroutine 用 el.send 持续投递事件（模拟读 goroutine），ch 满后
// 阻塞等空位（send 内部 select）——与修复前 closeEventLoops 的裸投递竞争空位，
// 填充方总赢 → 退出信号饿死 → Stop 卡死（红）；修复后（trySend + run 监听 ctx +
// drain）→ 不卡 + 在途（ch 存量）全部执行（绿）。
//
// 在途零丢失断言：executed（事件执行数）== fillOK（send 成功数）——每个投成的
// 事件要么被 loop 正常消费、要么被 drain 消费，无一丢失。

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

// testEventLoopOf 取 engine 的第一个 eventloop（white-box）。
func testEventLoopOf(t *testing.T, eng *engine) *eventloop {
	t.Helper()
	el := eng.eventLoops.index(0)
	if el == nil {
		t.Fatal("no eventloop")
	}
	return el
}

// testServerEventLoop server 侧取 eventloop：s.eng 由 run goroutine 赋值，
// 须经 engMu 锁读（裸读与赋值竞争，-race 必现）。
func testServerEventLoop(t *testing.T, srv Server) *eventloop {
	t.Helper()
	sv := srv.(*server)
	sv.engMu.Lock()
	eng := sv.eng
	sv.engMu.Unlock()
	return testEventLoopOf(t, eng)
}

// startStopWatch Stop 超时保护：Stop 必须在 timeout 内返回（红态：卡死超时）。
func startStopWatch(t *testing.T, timeout time.Duration, what string, stop func()) {
	t.Helper()
	done := make(chan struct{})
	go func() { stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("RED: %s hung under backpressure (A1)", what)
	}
}

// fillAndStop 填充 ch 至满 → Stop → 断言不卡 + 在途零丢失。
func fillAndStop(t *testing.T, el *eventloop, stop func()) {
	t.Helper()
	var executed, fillOK atomic.Int32

	// 填充 goroutine（模拟读 goroutine）：send 投递，ch 满后阻塞等空位；
	// ctx 取消（引擎关闭）后 send 返回 false → 停投（与真实读 goroutine 行为一致）
	stopFill := make(chan struct{})
	defer close(stopFill)
	go func() {
		for {
			select {
			case <-stopFill:
				return
			default:
			}
			if !el.send(func() error { executed.Add(1); return nil }) {
				return // 引擎关闭：停投
			}
			fillOK.Add(1)
		}
	}()

	// 等 ch 满（填充 goroutine 阻塞在 send）
	deadline := time.Now().Add(3 * time.Second)
	for len(el.ch) < cap(el.ch) {
		if time.Now().After(deadline) {
			t.Fatal("fill timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}

	// 引擎关闭：修复前（裸投递退出信号）→ 与填充 goroutine 抢空位 → 饿死卡死；
	// 修复后（trySend + run ctx.Done + drain）→ 不卡 + 在途完成
	startStopWatch(t, 3*time.Second, "Stop", stop)

	// 在途零丢失：投成的每个事件都被执行（正常消费 + drain 消费）
	deadline = time.Now().Add(2 * time.Second)
	for executed.Load() != fillOK.Load() {
		if time.Now().After(deadline) {
			t.Fatalf("in-flight events lost: executed=%d fillOK=%d", executed.Load(), fillOK.Load())
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestChain1_T8a_StopUnderBackpressureServer(t *testing.T) {
	// 直接构造 engine（等价 server 启动后的内部状态）：经真实 server 启动时
	// eventLoops 由 active() goroutine 注册，测试读 index(0) 与其无 happens-before
	// 竞争（-race 必现）；直接构造在测试 goroutine 内完成注册，零竞争，
	// 且关闭路径与 server.Stop 内部一致（shutdown → closeEventLoops → Wait）。
	rootCtx, shutdown := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(rootCtx)
	eng := &engine{
		opts:         &Options{},
		eventHandler: &BuiltinEventEngine{},
		turnOff:      shutdown,
		concurrency: struct {
			*errgroup.Group
			ctx context.Context
		}{eg, ctx},
	}
	eng.eventLoops = new(roundRobinLoadBalancer)
	el := &eventloop{ch: make(chan any, 1024), eng: eng, connections: make(map[*conn]struct{})}
	eng.eventLoops.register(el)
	eg.Go(el.run)

	fillAndStop(t, el, func() {
		eng.shutdown(nil)     // server.Stop 等价
		eng.closeEventLoops() // 含 trySend 退出信号（修复前裸投递）
		_ = eg.Wait()
	})
}

func TestChain1_T8b_StopUnderBackpressureClient(t *testing.T) {
	// client 同构验证（无 listener：client 的 Stop 走 closeEventLoops + Wait，同样修复）
	cfg := DefaultClientConfig()
	cli, err := NewClient(&BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cli.Stop() })

	el := testEventLoopOf(t, cli.(*client).eng)
	fillAndStop(t, el, func() { _ = cli.Stop() })
}
