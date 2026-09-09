package agonet

// 链1 T7：R6 Options.MaxConn 连接上限（护栏语义）。
// 概要设计 [[agonet-链1修复设计]] §R6：MaxConn>0 时超限连接静默拒绝（不触发 OnOpen/OnClose）。
// 默认 0 = 不限制（现有测试全绿即验证无行为变化）。
//
// 注意：测试清单（详细设计 §五 T1-T6）未含 R6——实施 review 补覆盖缺口。

import (
	"net"
	"sync/atomic"
	"testing"
	"time"
)

type maxConnHandler struct {
	*BuiltinEventEngine
	opens atomic.Int32
}

func (h *maxConnHandler) OnOpen(c Conn) ([]byte, Action) {
	h.opens.Add(1)
	return nil, None
}

func TestChain1_T7_MaxConnLimit(t *testing.T) {
	h := &maxConnHandler{BuiltinEventEngine: &BuiltinEventEngine{}}
	const addr = "tcp://127.0.0.1:18095"

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
	startErr := make(chan error, 1)
	go func() { startErr <- srv.Start() }()
	defer srv.Stop()

	// 等就绪
	host := "127.0.0.1:18095"
	deadline := time.Now().Add(3 * time.Second)
	var c1, c2 net.Conn
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", host, 200*time.Millisecond)
		if err == nil {
			c1 = c
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if c1 == nil {
		t.Fatal("server not ready")
	}
	defer c1.Close()

	// 连接 2：应被静默拒绝（rawConn 关闭，OnOpen 不触发）
	c2, err = net.DialTimeout("tcp", host, time.Second)
	if err != nil {
		t.Fatal(err) // TCP 层面仍可建立（护栏是 accept 后关闭）
	}
	defer c2.Close()

	// 等两个连接的 open 处理完成（读 goroutine 退出 + OnOpen 计数稳定）
	time.Sleep(500 * time.Millisecond)
	if got := h.opens.Load(); got != 1 {
		t.Fatalf("expected 1 OnOpen (MaxConn=1 超限静默拒绝), got %d", got)
	}
}
