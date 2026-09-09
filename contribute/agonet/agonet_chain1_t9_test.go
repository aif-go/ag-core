package agonet

// 链1 A1 收口（R4 跳满）测试（white-box，无 build tag 常驻）：
//   T9——accept 投递非阻塞化（sendNonBlocking）：业务 loop 的 ch 满时，
//   ① 新连接被快速失败（accept 后立即关闭，不静默挂起；OnOpen 不触发）
//   ② accept 循环不死（ch 空出后新连接正常接入，opens+1）
//
// 红态对照：阻塞投递（el.send）时，accept 卡在投递 → 阶段② Dial 挂起/超时。
// 注：直接构造 engine + 手动 listener + listenStream——经真实 server 启动时
// eventLoops 由 active() goroutine 注册，测试读 index(0) 与其无 happens-before 竞争。

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

// openCounter 统计 OnOpen 次数（快速失败连接不触发 OnOpen——openConn 未投递）。
type openCounter struct {
	BuiltinEventEngine
	opens atomic.Int32
}

func (h *openCounter) OnOpen(c Conn) ([]byte, Action) {
	h.opens.Add(1)
	return nil, None
}

func TestChain1_T9_AcceptSurvivesBackpressure(t *testing.T) {
	const addr = "127.0.0.1:18098"
	rootCtx, shutdown := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(rootCtx)
	eng := &engine{
		opts:         &Options{},
		eventHandler: &openCounter{},
		turnOff:      shutdown,
		concurrency: struct {
			*errgroup.Group
			ctx context.Context
		}{eg, ctx},
	}
	eng.eventLoops = new(roundRobinLoadBalancer)
	el := &eventloop{ch: make(chan any, 1024), eng: eng, eventHandler: eng.eventHandler, connections: make(map[*conn]struct{})}
	eng.eventLoops.register(el)

	ln, err := createListener("tcp", addr, &Options{})
	if err != nil {
		t.Fatal(err)
	}
	eng.listeners = []*listener{ln}
	eg.Go(el.run)
	eg.Go(func() error { return eng.listenStream(ln.ln) })
	t.Cleanup(func() {
		eng.shutdown(nil)
		eng.closeEventLoops()
		_ = eg.Wait()
	})

	// 等就绪
	deadline := time.Now().Add(3 * time.Second)
	for {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("server port not ready")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// 填充 ch 保持满（模拟业务积压）：send 阻塞等空位 → loop 消费一个补一个
	// 注：不 defer close(stopFill)——阶段② 显式 close；提前 Fatal 时 Cleanup 的
	// shutdown → ctx 取消 → 填充 goroutine 的 send 返回 false 自动退出
	stopFill := make(chan struct{})
	go func() {
		for {
			select {
			case <-stopFill:
				return
			default:
			}
			if !el.send(func() error { return nil }) {
				return
			}
		}
	}()
	for len(el.ch) < cap(el.ch) {
		time.Sleep(5 * time.Millisecond)
	}

	// 阶段①：ch 满时新连接被快速失败——多次尝试（填充补位有调度窗口，单次 Dial
	// 可能撞上空位投成；只要出现快速失败即验证 R4：连接被关且 OnOpen 不触发）
	rejected := false
	for i := 0; i < 10 && !rejected; i++ {
		c, err := net.DialTimeout("tcp", addr, time.Second)
		if err != nil {
			continue // listener 应仍 accept；偶发失败重试
		}
		_ = c.SetReadDeadline(time.Now().Add(time.Second))
		buf := make([]byte, 16)
		if _, rerr := c.Read(buf); rerr != nil {
			rejected = true // 快速失败：连接被服务端立即关闭（sendNonBlocking false → tc.Close）
		}
		_ = c.Close()
	}
	if !rejected {
		t.Fatal("RED: no conn rejected under backpressure (R4 跳满未生效——连接挂起不关)")
	}
	baseOpens := eng.eventHandler.(*openCounter).opens.Load()

	// 阶段②：停止填充 → ch 空出 → 新连接正常接入（accept 循环未被拖死）
	close(stopFill)
	deadline = time.Now().Add(3 * time.Second)
	for eng.eventHandler.(*openCounter).opens.Load() != baseOpens+1 {
		c2, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			_ = c2.Close()
		}
		if time.Now().After(deadline) {
			t.Fatalf("RED: accept loop dead after backpressure (opens=%d)", eng.eventHandler.(*openCounter).opens.Load())
		}
		time.Sleep(50 * time.Millisecond)
	}
}
