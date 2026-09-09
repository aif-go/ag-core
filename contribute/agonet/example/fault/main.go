// Package main fault 样例：故障域套件（背压/关闭/池满的集成回归）。
// 运行：go run ./contribute/agonet/example/fault
// 回归：go test -race -p 1 ./contribute/agonet/example/fault（池替换类测试独立包 + -p 1，M2）
//
// 演示：慢消费 handler（OnTraffic sleep 模拟慢业务）→ 客户端大量写入 → 读路径背压
// 但不整服关闭（链1 B1 修复后的行为：读 goroutine 出池，池满 ≠ 服务宕机）。
// 其余故障域行为（引擎关闭回落/Dial 不挂起/worker 归还/MaxConn）见 example_test.go。
package main

import (
	"fmt"
	"net"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
)

// slowHandler 慢消费 handler：模拟慢业务处理，制造读路径背压。
type slowHandler struct{}

func (h *slowHandler) OnBoot(agonet.Engine) agonet.Action         { return agonet.None }
func (h *slowHandler) OnShutdown(agonet.Engine)                   {}
func (h *slowHandler) OnOpen(agonet.Conn) ([]byte, agonet.Action) { return nil, agonet.None }
func (h *slowHandler) OnClose(agonet.Conn, error) agonet.Action   { return agonet.None }
func (h *slowHandler) OnTraffic(agonet.Conn) agonet.Action {
	time.Sleep(10 * time.Millisecond) // 慢业务：读路径背压
	return agonet.None
}

const addr = "tcp://127.0.0.1:18885"
const host = "127.0.0.1:18885"

func main() {
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = addr
	srv, err := agonet.NewServer(&slowHandler{}, &cfg)
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

	// 客户端大量写入（1MB×16）→ 慢 handler 制造背压 → 服务端读路径阻塞但不宕机
	ccfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(&agonet.BuiltinEventEngine{}, &ccfg)
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

	buf := make([]byte, 1<<20) // 1MB
	for i := 0; i < 16; i++ {
		if _, err := conn.Write(buf); err != nil {
			panic(err)
		}
	}

	// 背压下服务端仍可接受新连接（池满 ≠ 整服关）
	c2, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		panic(fmt.Sprintf("server died under backpressure: %v", err))
	}
	_ = c2.Close()
	fmt.Println("fault ok: server survives backpressure (pool-full no longer kills server)")
}
