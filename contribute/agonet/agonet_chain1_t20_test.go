package agonet

// 链1 T20（review 行内 P1 修正）：生命周期回调 panic 的连接清理。
// 红态（修复前）：safeHandle 吞 OnOpen/OnClose panic 保 loop，但——
//   - OnOpen panic：连接滞留 connections + 配额不回滚（MaxConn=1 后续连接被拒）——reviewer 实测
//   - OnClose panic：rawConn.Close/release 被 panic 中断（连接已出 map——run defer 不再兜底）——fd/缓冲泄漏
// 绿态（修复后）：open() OnOpen panic → close 肇事连接（配额/注册/缓冲对称回收）；
//   close() 资源收尾 defer 兜底 + OnClose 单独 recover——任何 panic 也释放。

import (
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

type lifecyclePanicHandler struct {
	*BuiltinEventEngine
	opens      atomic.Int32
	panicOpen  atomic.Bool
	panicClose atomic.Bool
}

func (h *lifecyclePanicHandler) OnOpen(c Conn) ([]byte, Action) {
	h.opens.Add(1)
	if h.panicOpen.Load() {
		panic("T20 OnOpen panic")
	}
	return nil, None
}

func (h *lifecyclePanicHandler) OnClose(c Conn, err error) Action {
	if h.panicClose.Load() {
		panic("T20 OnClose panic")
	}
	return None
}

// 红测（修复前）：OnOpen panic 后连接不关闭 + 配额不回滚——MaxConn=1 后续连接被拒
func TestChain1_T20_OnOpenPanicReclaimsQuota(t *testing.T) {
	h := &lifecyclePanicHandler{BuiltinEventEngine: &BuiltinEventEngine{}}
	h.panicOpen.Store(true)
	const addr = "tcp://127.0.0.1:18125"
	const host = "127.0.0.1:18125"

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

	// 就绪（探测连接会触发 OnOpen panic——等配额回滚稳定）
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
	time.Sleep(300 * time.Millisecond)
	base := h.opens.Load()

	// 连接 1：OnOpen panic → 引擎应关闭该连接（对端 EOF）
	c1, err := net.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()
	_ = c1.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := c1.Read(make([]byte, 1)); err == nil {
		t.Fatal("RED: conn not closed after OnOpen panic (expect EOF)")
	}
	time.Sleep(300 * time.Millisecond) // 等 close 处理完（配额回滚）

	// 连接 2：panicOpen 已关——配额回滚则 OnOpen 正常执行（opens 增长）；未回滚则被拒
	h.panicOpen.Store(false)
	before := h.opens.Load()
	c2, err := net.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()
	time.Sleep(300 * time.Millisecond)
	if got := h.opens.Load(); got <= before {
		t.Fatalf("RED: quota not reclaimed after OnOpen panic — subsequent conn rejected (opens %d→%d, base %d)", base, got, before)
	}
}

// 红测（修复前）：OnClose panic 中断 rawConn.Close/release——对端读不到 EOF（连接悬挂）
func TestChain1_T20_OnClosePanicStillClosesConn(t *testing.T) {
	h := &lifecyclePanicHandler{BuiltinEventEngine: &BuiltinEventEngine{}}
	h.panicClose.Store(true)
	const addr = "tcp://127.0.0.1:18126"
	const host = "127.0.0.1:18126"

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

	// 就绪
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

	conn, err := net.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	time.Sleep(300 * time.Millisecond) // 等 OnOpen 完成

	// 引擎关闭 → 所有连接 close → OnClose panic（recover）→ rawConn 仍须关闭
	_ = srv.Stop()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Read(make([]byte, 1)); err == nil {
		t.Fatal("RED: conn not closed despite OnClose panic (expect EOF)")
	} else if err != io.EOF {
		t.Fatalf("expect EOF, got %v", err)
	}
}
