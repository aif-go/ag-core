// Package main lifecycle 样例：连接生命周期（短连接风暴 + 长连接会话）。
// 运行：go run ./contribute/agonet/example/lifecycle
// 回归：go test -race ./contribute/agonet/example/lifecycle
//
// 演示：
//
//	短连接风暴：批量连接快速建立→关闭（无数据），验证建/关循环与资源回收（goroutine 回落）
//	长连接会话：单连接持续收发（echo），期间新连接不受影响
package main

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
)

// statHandler 统计 handler：记录打开/流量次数（长连接 echo 用 recvHandler 见测试）。
type statHandler struct {
	opens atomic.Int32
}

func (h *statHandler) OnBoot(agonet.Engine) agonet.Action { return agonet.None }
func (h *statHandler) OnShutdown(agonet.Engine)           {}
func (h *statHandler) OnOpen(agonet.Conn) ([]byte, agonet.Action) {
	h.opens.Add(1)
	return nil, agonet.None
}
func (h *statHandler) OnClose(agonet.Conn, error) agonet.Action { return agonet.None }
func (h *statHandler) OnTraffic(agonet.Conn) agonet.Action      { return agonet.None }

const addr = "tcp://127.0.0.1:18887"
const host = "127.0.0.1:18887"

func main() {
	sh := &statHandler{}
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = addr
	srv, err := agonet.NewServer(sh, &cfg)
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

	ccfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(&agonet.BuiltinEventEngine{}, &ccfg)
	if err != nil {
		panic(err)
	}
	if err := cli.Start(); err != nil {
		panic(err)
	}
	defer cli.Stop()

	// 短连接风暴：50 个连接快速建立→关闭
	for i := 0; i < 50; i++ {
		conn, err := cli.Dial("tcp", host)
		if err != nil {
			panic(err)
		}
		_ = conn.Close()
	}
	time.Sleep(300 * time.Millisecond) // 等服务端完成打开/关闭事件处理
	fmt.Printf("lifecycle ok: %d short-lived connections opened/closed\n", sh.opens.Load())
}
