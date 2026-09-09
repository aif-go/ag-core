package main

// echo 样例集成回归：core 层基本链路。
// 覆盖：单连接往返一致性、多连接并发、大数据量往返、优雅关闭（Start 返回 nil + goroutine 回落）。
// 端口 18881 独立（与 agonet 既有测试 18081-18095 及样例段 18882+ 不冲突）。
//
// 架构约束（C7）：conn.Read 只能在 eventloop goroutine 内（OnTraffic）调用——
// client 侧数据一律经 recvHandler（channel 传递）读取，测试 goroutine 不直接 Read。

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
)

// startEchoServer 启动 echo server（阻塞式 Start 须 goroutine 包裹）。
// defer 检查 startErr（Stop 后 Start 应返回 nil）。
func startEchoServer(t *testing.T) {
	t.Helper()
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = addr
	srv, err := agonet.NewServer(&echoHandler{}, &cfg)
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

// newClient 创建并启动 agonet 客户端（带接收 handler）。
func newClient(t *testing.T, h agonet.EventHandler) agonet.Client {
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

// recvN 从 handler channel 累积接收 n 字节（大 payload 回包分多次 OnTraffic 到达）。
func recvN(t *testing.T, got chan []byte, n int) []byte {
	t.Helper()
	out := make([]byte, 0, n)
	deadline := time.After(5 * time.Second)
	for len(out) < n {
		select {
		case chunk := <-got:
			out = append(out, chunk...)
		case <-deadline:
			t.Fatalf("recv timeout: got %d of %d bytes", len(out), n)
		}
	}
	return out
}

// echoOnce 单次往返：写 payload，经 handler channel 收满 len(payload) 字节，断言一致。
func echoOnce(t *testing.T, conn agonet.Conn, rh *recvHandler, payload []byte) {
	t.Helper()
	if _, err := conn.Write(payload); err != nil {
		t.Fatalf("Write err: %v", err)
	}
	echoed := recvN(t, rh.got, len(payload))
	if !bytes.Equal(echoed, payload) {
		t.Fatalf("echo mismatch: got %d bytes want %d", len(echoed), len(payload))
	}
}

func TestEcho_SingleRoundTrip(t *testing.T) {
	startEchoServer(t)
	rh := &recvHandler{got: make(chan []byte, 8)}
	cli := newClient(t, rh)

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// 随机 payload 往返一致性
	payload := make([]byte, 4096)
	if _, err := rand.Read(payload); err != nil {
		t.Fatal(err)
	}
	echoOnce(t, conn, rh, payload)
}

func TestEcho_ConcurrentClients(t *testing.T) {
	startEchoServer(t)

	const n = 8
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		go func(seed byte) {
			rh := &recvHandler{got: make(chan []byte, 8)}
			cfg := agonet.DefaultClientConfig()
			cli, err := agonet.NewClient(rh, &cfg)
			if err != nil {
				errCh <- err
				return
			}
			if err := cli.Start(); err != nil {
				errCh <- err
				return
			}
			defer cli.Stop()

			conn, err := cli.Dial("tcp", host)
			if err != nil {
				errCh <- err
				return
			}
			defer conn.Close()

			payload := bytes.Repeat([]byte{seed}, 1024)
			if _, err := conn.Write(payload); err != nil {
				errCh <- err
				return
			}
			out := make([]byte, 0, len(payload))
			deadline := time.After(5 * time.Second)
			for len(out) < len(payload) {
				select {
				case chunk := <-rh.got:
					out = append(out, chunk...)
				case <-deadline:
					errCh <- fmt.Errorf("seed=%d recv timeout", seed)
					return
				}
			}
			if !bytes.Equal(out, payload) {
				errCh <- fmt.Errorf("seed=%d mismatch", seed)
				return
			}
			errCh <- nil
		}(byte(i))
	}
	for i := 0; i < n; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("concurrent echo err: %v", err)
		}
	}
}

func TestEcho_LargePayload(t *testing.T) {
	startEchoServer(t)
	rh := &recvHandler{got: make(chan []byte, 8)}
	cli := newClient(t, rh)

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// 256KB：server 多次 Read(4KB) 循环写回 → client 多次 OnTraffic 到达
	payload := make([]byte, 256<<10)
	if _, err := rand.Read(payload); err != nil {
		t.Fatal(err)
	}
	echoOnce(t, conn, rh, payload)
}

func TestEcho_GracefulShutdown(t *testing.T) {
	startEchoServer(t)
	rh := &recvHandler{got: make(chan []byte, 8)}
	cli := newClient(t, rh)

	// 建 2 个连接并保持活跃
	conns := make([]agonet.Conn, 0, 2)
	for i := 0; i < 2; i++ {
		conn, err := cli.Dial("tcp", host)
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, conn)
	}
	time.Sleep(200 * time.Millisecond)
	peak := runtime.NumGoroutine()

	// 优雅关闭：所有连接关闭 + 读 goroutine 回落
	cli.Stop()
	for _, c := range conns {
		_ = c.Close()
	}

	deadline := time.Now().Add(5 * time.Second)
	for runtime.NumGoroutine() > peak-2 {
		if time.Now().After(deadline) {
			t.Fatalf("goroutines not reclaimed after shutdown: %d (peak %d)", runtime.NumGoroutine(), peak)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
