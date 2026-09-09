package agonet

// 链1 P1 红测/绿测（agonet 包 white-box，build tag: agonet_chain1）。
// 对应详细设计：[[agonet-链1修复详细设计]] §五 测试设计。
//
// T4（红测）：A3/connOpened——引擎关闭后 Dial 不挂起（当前挂起 → 红；修复后返回错误 → 绿）
// T2（绿测）：A2/A6 goroutine 回落（修复后回归验证）
//
// 诚实标注：T2 的红测触发条件（ch 满 + loop 永久退出）与 Stop 卡死（D 类）纠缠，
// 正常场景（ch 未满）当前代码不泄漏——T2 为修复后回归验证，非红测（M1 已标难测）。

import (
	"net"
	"runtime"
	"testing"
	"time"
)

// TestChain1_T4_DialAfterShutdown_ReturnsError 红测（当前红）：
// 客户端引擎关闭后 Dial → 当前 <-connOpened 永久阻塞（挂起）；修复后（send 失败 → close(connOpened) + 返回错误）不挂起。
// 顺带锁定 A3（引擎关闭后 Dial 永久挂起，跟踪清单 R-A3）。
func TestChain1_T4_DialAfterShutdown_ReturnsError(t *testing.T) {
	// 需要一个可达地址让 net.Dial 成功（才能走到 openConn 投递挂起点）
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := DefaultClientConfig()
	cli, err := NewClient(&BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	if err := cli.Stop(); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := cli.Dial("tcp", ln.Addr().String())
		done <- err
	}()

	select {
	case <-done:
		// 绿（修复后）：返回错误（send 失败 → close(connOpened) + return ErrEngineShutdown）
	case <-time.After(2 * time.Second):
		t.Fatal("RED: Dial hung after engine shutdown (A3/connOpened)")
	}
}

// TestChain1_T2_GoroutineReclaimAfterShutdown 绿测（修复后回归验证）：
// 引擎关闭后 goroutine 回落（读 goroutine + accept 退出）。
// 注意：非红测——正常场景（ch 未满）当前代码不泄漏（M1 已标 A2/A6 难测，时序敏感）。
func TestChain1_T2_GoroutineReclaimAfterShutdown(t *testing.T) {
	const addr = "127.0.0.1:18092"
	cfg := DefaultServerConfig()
	cfg.Addr = "tcp://" + addr
	srv, err := NewServer(&BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}

	startErr := make(chan error, 1)
	go func() { startErr <- srv.Start() }()
	// defer 顺序（LIFO）：check 先注册（后执行），Stop 后注册（先执行）
	defer func() {
		select {
		case err := <-startErr:
			if err != nil {
				t.Errorf("server.Start returned err=%v, expect nil after Stop", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server.Start did not return after Stop")
		}
	}()
	defer srv.Stop()

	waitFor(t, 2*time.Second, "server port ready", func() bool {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err != nil {
			return false
		}
		_ = c.Close()
		return true
	})

	// 建立 3 个连接，等读 goroutine 起来
	conns := make([]net.Conn, 0, 3)
	for i := 0; i < 3; i++ {
		c, err := net.DialTimeout("tcp", addr, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, c)
	}
	time.Sleep(200 * time.Millisecond)
	peak := runtime.NumGoroutine()

	// 关闭引擎
	if err := srv.Stop(); err != nil {
		t.Fatal(err)
	}
	for _, c := range conns {
		_ = c.Close()
	}

	// 断言回落（至少回落 2 个 goroutine：读 goroutine + accept 退出）
	waitFor(t, 5*time.Second, "goroutine reclaim after shutdown", func() bool {
		return runtime.NumGoroutine() <= peak-2
	})
}
