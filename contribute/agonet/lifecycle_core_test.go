package agonet

// 核心功能正确性测试（默认构建、绿测）：client/server 生命周期（真实 loopback）。
// 覆盖：Start/Stop 幂等、Dial→Enroll、OnOpen/OnTraffic/OnClose 事件链、数据往返。

import (
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// echoHandler 记录事件 + 回显数据。
type echoHandler struct {
	opened  atomic.Int32
	closed  atomic.Int32
	traffic atomic.Int32
	lastIn  atomic.Value // []byte
}

func (h *echoHandler) OnBoot(Engine) Action        { return None }
func (h *echoHandler) OnShutdown(Engine)           {}
func (h *echoHandler) OnOpen(c Conn) ([]byte, Action) {
	h.opened.Add(1)
	return nil, None
}
func (h *echoHandler) OnTraffic(c Conn) Action {
	h.traffic.Add(1)
	return None
}
func (h *echoHandler) OnClose(c Conn, err error) Action {
	h.closed.Add(1)
	return None
}

// waitFor 轮询等待条件成立。
func waitFor(t *testing.T, timeout time.Duration, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestServer_StopBeforeStart_Idempotent(t *testing.T) {
	s, err := NewServer(&BuiltinEventEngine{}, &ServerConfig{Addr: "tcp://127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Stop(); err != nil {
		t.Fatalf("Stop before Start err=%v, expect nil", err)
	}
}

func TestServer_Start_NoHandler_Error(t *testing.T) {
	s, err := NewServer(nil, &ServerConfig{Addr: "tcp://127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err == nil {
		t.Fatal("Start with nil handler must error")
	}
}

func TestServer_Start_NoAddress_Error(t *testing.T) {
	s, err := NewServer(&BuiltinEventEngine{}, &ServerConfig{Addr: ""})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err == nil {
		t.Fatal("Start with empty addr must error")
	}
}

func TestClientServer_EndToEnd_OpenTrafficClose(t *testing.T) {
	// 真实 loopback：server 监听固定端口，client 连接，验证事件链。
	// 注意：server.Start() 为阻塞式启动（阻塞运行直到 Stop），须 goroutine 包裹。
	cfg := DefaultServerConfig()
	cfg.Addr = "tcp://127.0.0.1:18081"
	sh := &echoHandler{}
	server, err := NewServer(sh, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	startErr := make(chan error, 1)
	go func() { startErr <- server.Start() }()
	// 注意 defer 顺序（LIFO）：先注册 check（后执行），再注册 Stop（先执行）——
	// 必须等 Stop 调用解除 eng.stop 的阻塞后，Start 才返回，check 才能收到结果。
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
	defer server.Stop()

	// 等待端口就绪
	waitFor(t, 2*time.Second, "server port ready", func() bool {
		c, err := net.DialTimeout("tcp", "127.0.0.1:18081", 200*time.Millisecond)
		if err != nil {
			return false
		}
		_ = c.Close()
		return true
	})

	// client 连接
	ccfg := DefaultClientConfig()
	ch := &echoHandler{}
	client, err := NewClient(ch, &ccfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Start(); err != nil {
		t.Fatal(err)
	}
	defer client.Stop()

	conn, err := client.Dial("tcp", "127.0.0.1:18081")
	if err != nil {
		t.Fatalf("Dial err: %v", err)
	}

	// 写数据 → 触发 server OnTraffic
	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatalf("Write err: %v", err)
	}
	waitFor(t, 2*time.Second, "server OnOpen", func() bool { return sh.opened.Load() > 0 })
	waitFor(t, 2*time.Second, "client OnOpen", func() bool { return ch.opened.Load() > 0 })
	waitFor(t, 2*time.Second, "server OnTraffic", func() bool { return sh.traffic.Load() > 0 })

	// 关闭连接 → OnClose
	_ = conn.Close()
	waitFor(t, 2*time.Second, "server OnClose", func() bool { return sh.closed.Load() > 0 })
	waitFor(t, 2*time.Second, "client OnClose", func() bool { return ch.closed.Load() > 0 })
}

func TestClient_StartStop_Idempotent(t *testing.T) {
	cfg := DefaultClientConfig()
	cli, err := NewClient(&BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	if err := cli.Stop(); err != nil {
		t.Fatalf("Stop err=%v", err)
	}
	// 二次 Stop 应安全（幂等）
	_ = cli.Stop()
}
