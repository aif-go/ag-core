package agonet

// 链1 T19（D6）：loop 级 panic 隔离——handler panic 不崩进程。
// 红态（修复前）：核心代码无 recover——panic 传播到进程——整服崩溃（probe 实证 0.06s）
// 绿态（修复后）：read 层 recover（记录 + 关肇事连接）+ safeHandle 层 recover（记录）——
// 引擎存活 + 服务继续 + 连接清理

import (
	"net"
	"sync/atomic"
	"testing"
	"time"
)

type panicCloseHandler struct {
	*BuiltinEventEngine
	closed atomic.Int32
}

func (h *panicCloseHandler) OnTraffic(c Conn) Action { panic("T19 handler panic") }

func (h *panicCloseHandler) OnClose(c Conn, err error) Action {
	h.closed.Add(1)
	return None
}

func TestChain1_T19_HandlerPanicIsolated(t *testing.T) {
	h := &panicCloseHandler{BuiltinEventEngine: &BuiltinEventEngine{}}
	const addr = "tcp://127.0.0.1:18124"
	const host = "127.0.0.1:18124"

	cfg := DefaultServerConfig()
	opts, err := BuildOptionsWithConfig(cfg.Config)
	if err != nil {
		t.Fatal(err)
	}
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

	// 触发 panic 连接（OnTraffic panic——红态下此连接的处理直接炸进程）
	nc, err := net.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Close()
	_, _ = nc.Write([]byte("ping"))

	// 等 panic 处理 + 连接清理
	time.Sleep(500 * time.Millisecond)

	// ① 引擎存活：probe Dial 成功（accept loop 活着）
	probe, err := net.DialTimeout("tcp", host, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("D6: engine died after handler panic (probe dial failed: %v)——panic 未隔离——整服崩溃", err)
	}
	_ = probe.Close()

	// ② 肇事连接被关（read 层 recover → el.close——OnClose 计数）
	if got := h.closed.Load(); got == 0 {
		t.Fatalf("D6: panic conn not closed (closed=%d)——半状态连接残留", got)
	}
}
