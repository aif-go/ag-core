package agonet

// 链1 T16（PR review P1 修正）：MaxConn 全局原子配额——并发接入时 OnOpen ≤ MaxConn。
// 红态（修复前）：totalConns() 读时求和 + incConn 分离——多 loop 并发 open 竞态超限
// （reviewer 实测 MaxConn=1 触发 2 次 OnOpen）
// 绿态（修复后）：engine.totalConn Add(+1) 后判断——原子——恒 ≤ MaxConn

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestChain1_T16_MaxConnAtomicQuota(t *testing.T) {
	h := &maxConnHandler{BuiltinEventEngine: &BuiltinEventEngine{}}
	const addr = "tcp://127.0.0.1:18096"
	const host = "127.0.0.1:18096"

	cfg := DefaultServerConfig()
	opts, err := BuildOptionsWithConfig(cfg.Config)
	if err != nil {
		t.Fatal(err)
	}
	opts.MaxConn = 1

	srv, err := NewServerWithOptions(h, []string{addr}, opts)
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = srv.Start() }()
	defer srv.Stop()

	// 等就绪
	deadline := time.Now().Add(3 * time.Second)
	for {
		c, err := net.DialTimeout("tcp", host, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("server not ready")
		}
		time.Sleep(50 * time.Millisecond)
	}
	// 等探测连接关闭处理完（配额释放——totalConn 归零），记录 OnOpen 基线
	time.Sleep(300 * time.Millisecond)
	base := h.opens.Load()

	// 并发 10 连接同时接入并【保持打开】（建立即关会让配额快速释放——连接依次通过——
	// 测不出原子配额；保持打开才验证"同时存活 ≤ MaxConn"）
	var wg sync.WaitGroup
	var mu sync.Mutex
	conns := make([]net.Conn, 0, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := net.Dial("tcp", host)
			if err == nil {
				mu.Lock()
				conns = append(conns, c) // 保持打开（不 Close——配额不释放）
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// 等 open 处理稳定
	time.Sleep(500 * time.Millisecond)
	if got := h.opens.Load() - base; got > 1 {
		t.Fatalf("MaxConn=1 并发接入: 增量 OnOpen=%d（原子配额被破坏——修复前竞态超限）", got)
	}
	// 清理（被拒连接已 EOF——Close 幂等无害）
	for _, c := range conns {
		_ = c.Close()
	}
}
