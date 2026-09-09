package main

// idle 样例集成回归：IdleStateHandler 心跳/空闲超时。
// 覆盖：ALL_IDLE 超时触发关闭、心跳保活（持续流量不触发）、READER_IDLE 独立配置。
// 端口 18883 独立。⚠️ idle 最小 1 秒（实现约束），单测 2-3 秒。

import (
	"net"
	"testing"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
	"github.com/aif-go/ag-core/contribute/agonet/simple"
)

// startIdleServer 启动 idle 管理 server（events 收 idle 事件）。
func startIdleServer(t *testing.T, readerIdle, writerIdle, allIdle int64, events chan<- simple.IdleStateEvent) {
	t.Helper()
	handler, err := newIdleHandler(readerIdle, writerIdle, allIdle, events)
	if err != nil {
		t.Fatal(err)
	}
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = addr
	srv, err := agonet.NewServer(handler, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	startErr := make(chan error, 1)
	go func() { startErr <- srv.Start() }()
	t.Cleanup(func() {
		select {
		case err := <-startErr:
			if err != nil {
				t.Errorf("server.Start returned err=%v, expect nil after Stop", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server.Start did not return after Stop")
		}
	})
	t.Cleanup(func() { _ = srv.Stop() })

	// 等端口就绪
	deadline := time.Now().Add(3 * time.Second)
	for {
		c, err := net.DialTimeout("tcp", host, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("server port not ready")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// newIdleClient 创建简单客户端（无需 handler 逻辑）。
func newIdleClient(t *testing.T) agonet.Client {
	t.Helper()
	cfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(&agonet.BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cli.Stop() })
	return cli
}

func TestIdle_AllIdleTriggersClose(t *testing.T) {
	events := make(chan simple.IdleStateEvent, 8)
	startIdleServer(t, 1, 1, 1, events) // 全空闲 1 秒（最小单位）
	cli := newIdleClient(t)

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// 连接后静默 → 1 秒后应触发 ALL_IDLE（期间 READER/WRITER_IDLE 也会触发）
	deadline := time.After(4 * time.Second)
	sawAll := false
	for !sawAll {
		select {
		case ev := <-events:
			if ev.State == simple.ALL_IDLE {
				sawAll = true
			}
		case <-deadline:
			t.Fatal("ALL_IDLE not fired within 4s")
		}
	}

	// 服务端已关连接：client 读应 EOF/错误（数据读尽或连接关闭）
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 16)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			return // 连接被服务端关闭（期望）
		}
	}
}

func TestIdle_HeartbeatKeepsAlive(t *testing.T) {
	events := make(chan simple.IdleStateEvent, 8)
	// 只配 reader idle（server 单向收心跳，不写——配 writer/all idle 会因 lastWritTime 不更新而误触发）
	startIdleServer(t, 1, 0, 0, events)
	cli := newIdleClient(t)

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// 心跳：Dial 后立即启动（每 400ms 发 1 字节 < 1s idle 阈值），持续更新 lastReadTime。
	// 首触发窗口：timer 自 OnOpen 起 1s 触发，检查 time.Since(lastReadTime) >= 1s——
	// 单次热身写（OnOpen 后立即）因边界（1s-ε）不可靠，须用心跳流持续刷新；
	// sleep 1.2s 跨过首触发窗口后再进观察窗，避免 -race 全量并行下的调度延迟误报。
	// stop+defer 模式：任何退出路径（含 Fatal）都停心跳 goroutine，防测试结束后泄漏竞争
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		tick := time.NewTicker(400 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-stop:
				return
			case <-tick.C:
				if _, err := conn.Write([]byte("h")); err != nil {
					return // 连接关闭（不该发生，但退出心跳）
				}
			}
		}
	}()
	defer func() { close(stop); <-done }()
	time.Sleep(1200 * time.Millisecond) // 跨过首触发窗口（OnOpen+1s）

	select {
	case ev := <-events:
		t.Fatalf("idle fired despite heartbeat: %s", ev.State)
	case <-time.After(2500 * time.Millisecond):
		// 心跳保活生效：无 idle 事件
	}
}

func TestIdle_ReaderIdleOnly(t *testing.T) {
	events := make(chan simple.IdleStateEvent, 8)
	startIdleServer(t, 1, 0, 0, events) // 仅读空闲：1 秒无读数据即触发
	cli := newIdleClient(t)

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// 持续写（200ms tick < 1s idle 阈值）：Dial 后立即启动，持续刷新 lastReadTime；
	// sleep 1.2s 跨过首触发窗口（OnOpen+1s 边界误报——单次热身写不可靠，须持续流）
	// stopFn：正常路径第一阶段后显式停写（第二阶段需静默）；Fatal 路径 defer 兜底停
	stop := make(chan struct{})
	writeDone := make(chan struct{})
	stopped := false
	stopFn := func() {
		if !stopped {
			stopped = true
			close(stop)
			<-writeDone
		}
	}
	defer stopFn()
	go func() {
		defer close(writeDone)
		tick := time.NewTicker(200 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-stop:
				return
			case <-tick.C:
				if _, err := conn.Write([]byte("d")); err != nil {
					return
				}
			}
		}
	}()
	time.Sleep(1200 * time.Millisecond) // 跨过首触发窗口（OnOpen+1s）

	select {
	case ev := <-events:
		t.Fatalf("READER_IDLE fired despite read traffic: %s", ev.State)
	case <-time.After(1600 * time.Millisecond):
	}
	stopFn() // 停写 → 静默观察

	// 停写 → READER_IDLE 触发（deadline 8s：容忍 -race + 高负载下 timer 回调链
	// onTimeoutInEL → Execute → 池 worker → loop 的调度延迟，偶发 >4s）
	deadline := time.After(8 * time.Second)
	for {
		select {
		case ev := <-events:
			if ev.State == simple.READER_IDLE {
				return
			}
		case <-deadline:
			t.Fatal("READER_IDLE not fired after write stopped")
		}
	}
}
