// Package main echo 样例：core 层基本链路（server/client 数据往返）。
// 运行：go run ./contribute/agonet/example/echo
// 回归：go test -race ./contribute/agonet/example/echo
//
// 演示：EventHandler 直接使用（core 层）。
// 正确用法（架构约束，C7）：conn.Read 只能在 eventloop goroutine 内调用
// （OnTraffic 回调中）——数据由 loop 写入 conn.buffer，外部 goroutine 直接
// Read 会与 loop 并发竞争（-race 触发）。因此 server 用 echoHandler 回显，
// client 用 recvHandler 在 OnTraffic 内读数据并经 channel 交给业务 goroutine。
package main

import (
	"fmt"
	"net"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
)

// readAllInLoop 在 eventloop goroutine 内读出当前全部入站数据。
// agonet conn.Read 契约（C7）：数据读尽返回 io.ErrShortBuffer（非 EOF）。
func readAllInLoop(c agonet.Conn) []byte {
	buf := make([]byte, 4096)
	var out []byte
	for {
		n, err := c.Read(buf)
		if n > 0 {
			out = append(out, buf[:n]...)
		}
		if err != nil {
			break // 读尽（ErrShortBuffer）或连接错误——框架负责关闭
		}
		if n == 0 {
			break
		}
	}
	return out
}

// echoHandler 服务端回显 handler：OnTraffic（loop 内）读出全部入站数据并写回。
type echoHandler struct{}

func (h *echoHandler) OnBoot(agonet.Engine) agonet.Action         { return agonet.None }
func (h *echoHandler) OnShutdown(agonet.Engine)                   {}
func (h *echoHandler) OnOpen(agonet.Conn) ([]byte, agonet.Action) { return nil, agonet.None }
func (h *echoHandler) OnClose(agonet.Conn, error) agonet.Action   { return agonet.None }
func (h *echoHandler) OnTraffic(c agonet.Conn) agonet.Action {
	data := readAllInLoop(c)
	if len(data) > 0 {
		if _, werr := c.Write(data); werr != nil {
			return agonet.Close
		}
	}
	return agonet.None
}

// recvHandler 客户端接收 handler：OnTraffic（loop 内）读出数据并经 channel 交给业务侧。
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

const addr = "tcp://127.0.0.1:18881"
const host = "127.0.0.1:18881"

func main() {
	// 服务端（阻塞式启动，须 goroutine 包裹）
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = addr
	srv, err := agonet.NewServer(&echoHandler{}, &cfg)
	if err != nil {
		panic(err)
	}
	go func() { _ = srv.Start() }()
	defer srv.Stop()

	// 等端口就绪
	deadline := time.Now().Add(3 * time.Second)
	for {
		c, err := net.DialTimeout("tcp", host, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			break
		}
		if time.Now().After(deadline) {
			panic("server not ready")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// 客户端：连接 → 发送 → 经 recvHandler 收回声
	rh := &recvHandler{got: make(chan []byte, 4)}
	ccfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(rh, &ccfg)
	if err != nil {
		panic(err)
	}
	if err := cli.Start(); err != nil {
		panic(err)
	}
	defer cli.Stop()

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	payload := []byte("hello agonet echo")
	if _, err := conn.Write(payload); err != nil {
		panic(err)
	}
	select {
	case echoed := <-rh.got:
		if string(echoed) != string(payload) {
			panic(fmt.Sprintf("echo mismatch: got %q want %q", echoed, payload))
		}
		fmt.Printf("echo ok: %q\n", echoed)
	case <-time.After(3 * time.Second):
		panic("echo timeout")
	}
}
