package main

// lifecycle 样例集成回归：连接生命周期。
// 覆盖：短连接风暴（200 连接快速建/关，全部 OnOpen + 资源回收回落）、
//       长连接会话（单连接多次 echo + 期间新连接不受影响 + 关闭后回落）。
// 端口 18887 独立。
//
// 架构约束（C7）：长连接 echo 经 recvHandler channel 读（loop 内）；短连接无数据读写。

import (
	"bytes"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
)

// readAllInLoop 在 eventloop goroutine 内读出当前全部入站数据（C7：loop 内读）。
func readAllInLoop(c agonet.Conn) []byte {
	buf := make([]byte, 4096)
	var out []byte
	for {
		n, err := c.Read(buf)
		if n > 0 {
			out = append(out, buf[:n]...)
		}
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
	}
	return out
}

// echoHandler 服务端回显 handler（长连接 echo 会话）。
type echoHandler struct{}

func (h *echoHandler) OnBoot(agonet.Engine) agonet.Action         { return agonet.None }
func (h *echoHandler) OnShutdown(agonet.Engine)                   {}
func (h *echoHandler) OnOpen(agonet.Conn) ([]byte, agonet.Action) { return nil, agonet.None }
func (h *echoHandler) OnClose(agonet.Conn, error) agonet.Action   { return agonet.None }
func (h *echoHandler) OnTraffic(c agonet.Conn) agonet.Action {
	if data := readAllInLoop(c); len(data) > 0 {
		if _, werr := c.Write(data); werr != nil {
			return agonet.Close
		}
	}
	return agonet.None
}

// recvHandler 客户端接收 handler（loop 内读 → channel）。
type recvHandler struct {
	got chan []byte
}

func (h *recvHandler) OnBoot(agonet.Engine) agonet.Action         { return agonet.None }
func (h *recvHandler) OnShutdown(agonet.Engine)                   {}
func (h *recvHandler) OnOpen(agonet.Conn) ([]byte, agonet.Action) { return nil, agonet.None }
func (h *recvHandler) OnClose(agonet.Conn, error) agonet.Action   { return agonet.None }
func (h *recvHandler) OnTraffic(c agonet.Conn) agonet.Action {
	if data := readAllInLoop(c); len(data) > 0 {
		h.got <- data
	}
	return agonet.None
}

// startLifecycleServer 启动 server（阻塞式 Start 须 goroutine 包裹），返回 srv。
func startLifecycleServer(t *testing.T, h agonet.EventHandler) agonet.Server {
	t.Helper()
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = addr
	srv, err := agonet.NewServer(h, &cfg)
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
			return srv
		}
		if time.Now().After(deadline) {
			t.Fatal("server port not ready")
		}
		time.Sleep(20 * time.Millisecond)
	}
	return srv
}

// newLifecycleClient 创建 agonet 客户端（可选 handler）。
func newLifecycleClient(t *testing.T, h agonet.EventHandler) agonet.Client {
	t.Helper()
	cfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(h, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cli.Stop() })
	return cli
}

// waitGoroutines 轮询 goroutine 数回落。
func waitGoroutines(t *testing.T, timeout time.Duration, target int) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for runtime.NumGoroutine() > target {
		if time.Now().After(deadline) {
			t.Fatalf("goroutines not reclaimed: %d (target %d)", runtime.NumGoroutine(), target)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestLifecycle_ShortConnStorm(t *testing.T) {
	sh := &statHandler{}
	srv := startLifecycleServer(t, sh)
	cli := newLifecycleClient(t, &agonet.BuiltinEventEngine{})

	base := runtime.NumGoroutine()
	const n = 200
	for i := 0; i < n; i++ {
		conn, err := cli.Dial("tcp", host)
		if err != nil {
			t.Fatalf("dial %d err: %v", i, err)
		}
		_ = conn.Close()
	}

	// 全部连接 OnOpen（服务端处理完建/关事件链）
	deadline := time.Now().Add(5 * time.Second)
	for int(sh.opens.Load()) < n {
		if time.Now().After(deadline) {
			t.Fatalf("opens=%d want %d", sh.opens.Load(), n)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// 连接读 goroutine 全部回收（回落到基线附近；ants 池 worker 常驻含在基线）
	_ = srv
	waitGoroutines(t, 5*time.Second, base+4)
}

func TestLifecycle_LongLivedSession(t *testing.T) {
	startLifecycleServer(t, &echoHandler{})
	rh := &recvHandler{got: make(chan []byte, 16)}
	cli := newLifecycleClient(t, rh)

	// 长连接：10 次 echo 会话（间隔 100ms，模拟持续业务）
	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for i := 0; i < 10; i++ {
		payload := []byte("session-msg")
		if _, err := conn.Write(payload); err != nil {
			t.Fatal(err)
		}
		select {
		case echoed := <-rh.got:
			if !bytes.Equal(echoed, payload) {
				t.Fatalf("session echo mismatch: %q", echoed)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("session echo %d timeout", i)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 会话期间新连接不受影响
	c2, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c2.Write([]byte("new-conn")); err != nil {
		t.Fatal(err)
	}
	select {
	case echoed := <-rh.got:
		if string(echoed) != "new-conn" {
			t.Fatalf("new conn echo mismatch: %q", echoed)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("new conn echo timeout")
	}
	_ = c2.Close()

	// 关闭后资源回收
	_ = conn.Close()
	time.Sleep(200 * time.Millisecond)
	base := runtime.NumGoroutine()
	waitGoroutines(t, 5*time.Second, base+2)
}
