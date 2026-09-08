package agonet

// 链3 B2-D/C3 生命周期行为锁定测试（white-box + 集成）：
//   T14——连接建/关循环后 goroutine 回落（读缓冲 defer Put + tc.b 归还兜底，
//   无 goroutine/资源泄漏）。与 T2（读 goroutine 泄漏专项）互补：
//   T2 聚焦 R1 出池；T14 聚焦 B2-D 读缓冲池化 + C3 归还（行为锁定，宽松断言）。
import (
	"net"
	"runtime"
	"testing"
	"time"
)

func TestChain3_T14_ConnLifecycleNoLeak(t *testing.T) {
	const addr = "127.0.0.1:18114"
	cfg := DefaultServerConfig()
	cfg.Addr = "tcp://" + addr
	srv, err := NewServer(&BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	// Server.Start 为阻塞式（run():137-142——阻塞直到 Stop）——必须异步启动
	go func() { _ = srv.Start() }()
	t.Cleanup(func() { _ = srv.Stop() })

	// 就绪 + 基线 goroutine
	dialOK(t, addr)
	time.Sleep(200 * time.Millisecond)
	baseline := runtime.NumGoroutine()

	// 建/关循环（每轮：Dial + 关闭——读 goroutine 启动（Get 读缓冲）→ 退出（Put））
	const loops = 30
	for i := 0; i < loops; i++ {
		nc, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		// 少量数据触发读路径
		_, _ = nc.Write([]byte("ping"))
		time.Sleep(5 * time.Millisecond)
		_ = nc.Close()
		time.Sleep(5 * time.Millisecond)
	}

	// 等待读 goroutine 全部退出（defer Put 执行）
	deadline := time.Now().Add(5 * time.Second)
	for runtime.NumGoroutine() > baseline+8 && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > baseline+8 {
		t.Fatalf("goroutine 未回落：基线=%d 当前=%d（B2-D 读缓冲/C3 归还泄漏）", baseline, n)
	}
}

func dialOK(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		nc, err := net.Dial("tcp", addr)
		if err == nil {
			_ = nc.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("server not ready: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
