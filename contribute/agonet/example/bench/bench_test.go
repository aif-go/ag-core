package main

// agonet 整体性能评估——基准套件（example/bench）。
// 运行：go test -bench=. -benchmem -count=5 ./contribute/agonet/example/bench
//   （-count=5 防机器波动；bench 不带 -race——race 会大幅失真）
// 评估维度：① 性能（RTT/吞吐/连接数扩展）② 内存分配（allocs/op）
// 设计：真实 server + client（端到端——含 syscall/调度/通道），非白盒。
// 端口 18201（独立段——与 agonet 测试 18081-18095 / 样例 18881+ 不冲突）。
//
// 架构约束（C7）：conn.Read 只能在 eventloop goroutine（OnTraffic）内——
// client 侧数据经 recvHandler channel 收，基准 goroutine 不直接 Read。

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
)

// ---------- handlers（复用 echo 样例模式） ----------

// readAllInLoop 在 eventloop goroutine 内读出全部入站数据（C7：读尽返回 ErrShortBuffer）。
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

// echoHandler 服务端回显：OnTraffic 内读出全部并写回。
type echoHandler struct{}

func (h *echoHandler) OnBoot(agonet.Engine) agonet.Action         { return agonet.None }
func (h *echoHandler) OnShutdown(agonet.Engine)                   {}
func (h *echoHandler) OnOpen(agonet.Conn) ([]byte, agonet.Action) { return nil, agonet.None }
func (h *echoHandler) OnClose(agonet.Conn, error) agonet.Action   { return agonet.None }
func (h *echoHandler) OnTraffic(c agonet.Conn) agonet.Action {
	if data := readAllInLoop(c); len(data) > 0 {
		if _, err := c.Write(data); err != nil {
			// 写失败（loop 背压 ch 满等）→ 关闭连接（防半开连接 + client 侧感知快速失败）
			return agonet.Close
		}
	}
	return agonet.None
}

// recvHandler 客户端接收：OnTraffic 内读出并经 channel 交基准 goroutine。
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

// ---------- 全局 setup（TestMain：一次启动 server/client） ----------

const benchAddr = "tcp://127.0.0.1:18201"
const benchHost = "127.0.0.1:18201"

var (
	benchSrv agonet.Server
	benchCli agonet.Client
	benchRH  *recvHandler
	benchRO  *recvOnlyHandler
)

func TestMain(m *testing.M) {
	// 服务端 1（echo——RTT 基准用）：阻塞式启动，goroutine 包裹
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = benchAddr
	srv, err := agonet.NewServer(&echoHandler{}, &cfg)
	if err != nil {
		panic(err)
	}
	go func() { _ = srv.Start() }()
	benchSrv = srv

	// 服务端 2（只收不答——吞吐基准用）：端口 18202
	cfg2 := agonet.DefaultServerConfig()
	cfg2.Addr = "tcp://127.0.0.1:18202"
	benchRO = &recvOnlyHandler{}
	srv2, err := agonet.NewServer(benchRO, &cfg2)
	if err != nil {
		panic(err)
	}
	go func() { _ = srv2.Start() }()

	// 等两个端口就绪
	deadline := time.Now().Add(3 * time.Second)
	for _, host := range []string{benchHost, "127.0.0.1:18202"} {
		for {
			c, err := net.DialTimeout("tcp", host, 200*time.Millisecond)
			if err == nil {
				_ = c.Close()
				break
			}
			if time.Now().After(deadline) {
				panic("server not ready: " + host)
			}
			time.Sleep(20 * time.Millisecond)
		}
	}

	// 客户端
	benchRH = &recvHandler{got: make(chan []byte, 256)}
	ccfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(benchRH, &ccfg)
	if err != nil {
		panic(err)
	}
	if err := cli.Start(); err != nil {
		panic(err)
	}
	benchCli = cli

	code := m.Run()
	_ = cli.Stop()
	_ = srv.Stop()
	_ = srv2.Stop()
	os.Exit(code)
}

// ---------- 基准 1：单连接 echo 往返延迟（RTT） ----------

func BenchmarkEchoRTT(b *testing.B) {
	conn, err := benchCli.Dial("tcp", benchHost)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	payload := bytes.Repeat([]byte{'x'}, 64) // 64B 小帧（常见消息量级）
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := conn.Write(payload); err != nil {
			b.Fatal(err)
		}
		<-benchRH.got // 等回显（同步往返）
	}
}

// recvOnlyHandler 只收不答（单向吞吐场景——避免回显路径背压失真）
type recvOnlyHandler struct {
	count atomic.Int64
}

func (h *recvOnlyHandler) OnBoot(agonet.Engine) agonet.Action         { return agonet.None }
func (h *recvOnlyHandler) OnShutdown(agonet.Engine)                   {}
func (h *recvOnlyHandler) OnOpen(agonet.Conn) ([]byte, agonet.Action) { return nil, agonet.None }
func (h *recvOnlyHandler) OnClose(agonet.Conn, error) agonet.Action   { return agonet.None }
func (h *recvOnlyHandler) OnTraffic(c agonet.Conn) agonet.Action {
	if data := readAllInLoop(c); len(data) > 0 {
		h.count.Add(int64(len(data)))
	}
	return agonet.None
}

// ---------- 基准 2：多连接单向吞吐（client 投递 + server 接收双计数） ----------
// 注：echo（回显）模式在多连接高速下会打爆 server loop ch（回显 Write 失败丢帧——
// 试跑实锤）——单向测端到端入站吞吐（背压损耗自然计入）

func BenchmarkThroughput(b *testing.B) {
	const tpHost = "127.0.0.1:18202" // recvOnly server（避免 echo 回显背压）
	for _, n := range []int{10, 100} {
		b.Run(fmt.Sprintf("N%d", n), func(b *testing.B) {
			conns := make([]agonet.Conn, n)
			for i := range conns {
				c, err := benchCli.Dial("tcp", tpHost)
				if err != nil {
					b.Fatal(err)
				}
				conns[i] = c
			}
			defer func() {
				for _, c := range conns {
					_ = c.Close()
				}
			}()

			payload := bytes.Repeat([]byte{'x'}, 64)
			var ok, fail atomic.Int64
			before := benchRO.count.Load()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				for j := 0; j < n; j++ {
					wg.Add(1)
					go func(c agonet.Conn) {
						defer wg.Done()
						if _, err := c.Write(payload); err != nil {
							fail.Add(1) // client 侧投递失败（ch 满——打满状态的背压损耗）
						} else {
							ok.Add(1)
						}
					}(conns[j])
				}
				wg.Wait()
			}
			dur := b.Elapsed().Seconds()
			got := benchRO.count.Load() - before // server 实际接收字节
			b.ReportMetric(float64(ok.Load())/dur, "client_frames/s")
			b.ReportMetric(float64(got)/float64(len(payload))/dur, "server_frames/s")
			b.ReportMetric(float64(got)/dur/1e6, "MB/s")
			b.ReportMetric(float64(fail.Load())/dur, "fail/s")
		})
	}
}

// ---------- 基准 3：连接建/关生命周期开销 ----------

func BenchmarkConnLifecycle(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn, err := benchCli.Dial("tcp", benchHost)
		if err != nil {
			b.Fatal(err)
		}
		_ = conn.Close()
	}
}
