package agonet

// 链1 T18（PR review P2 修正）：OnBoot 提前退出释放监听端口。
// 红态（修复前）：OnBoot 返回 Close/Shutdown → Start 直接 return——listener 未关——
// 端口泄漏——同一端口无法重新监听
// 绿态（修复后）：提前返回分支补 ln.close()——端口释放

import (
	"net"
	"testing"
	"time"
)

type bootCloseHandler struct {
	*BuiltinEventEngine
}

func (h *bootCloseHandler) OnBoot(Engine) Action { return Close }

func TestChain1_T18_OnBootEarlyExitFreesPort(t *testing.T) {
	const addr = "tcp://127.0.0.1:18097"
	const host = "127.0.0.1:18097"

	cfg := DefaultServerConfig()
	opts, err := BuildOptionsWithConfig(cfg.Config)
	if err != nil {
		t.Fatal(err)
	}
	srv, err := NewServerWithOptions(&bootCloseHandler{BuiltinEventEngine: &BuiltinEventEngine{}}, []string{addr}, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Start 阻塞式——但 OnBoot 返回 Close → 提前返回（goroutine 包裹）
	startErr := make(chan error, 1)
	go func() { startErr <- srv.Start() }()

	// 等 Start 返回
	select {
	case err := <-startErr:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return after OnBoot Close")
	}

	// 同一端口应可重新监听（修复前泄漏——Listen 失败）
	ln, err := net.Listen("tcp", host)
	if err != nil {
		t.Fatalf("port still held after OnBoot early exit: %v（P2 泄漏）", err)
	}
	_ = ln.Close()
}
