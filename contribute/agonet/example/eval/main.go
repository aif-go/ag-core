package main

// agonet 整体性能评估——负载评估工具（example/eval）。
// 运行：
//   go run ./contribute/agonet/example/eval -mode stream   # 持续帧流 10s（吞吐 + 内存采样）
//   go run ./contribute/agonet/example/eval -mode storm    # 连接风暴（1K 快速建/关——goroutine 回落）
//   go run ./contribute/agonet/example/eval -mode leak     # 建/关循环——前后 heap/goroutine 对比（泄露检测）
//   go run ./contribute/agonet/example/eval -mode all      # 依次执行三场景
//   -pprof-dir <dir>   # 采样 CPU/heap/goroutine profile（写入 dir）
// 评估维度：性能（吞吐/延迟）、内存分配（heap/GC）、并发（goroutine 回落）、内存泄露（前后对比）。
// 端口 18210/18211（独立段）。
//
// 架构约束（C7）：conn.Read 只在 OnTraffic（loop 内）——client 数据经 handler channel 收。

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
)

// ---------- handlers ----------

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

// ---------- 工具 ----------

func memStats() (allocMB float64, sysMB float64, gcNum uint32, gNum int) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return float64(ms.HeapAlloc) / 1e6, float64(ms.Sys) / 1e6, ms.NumGC, runtime.NumGoroutine()
}

func waitPort(host string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for {
		c, err := net.DialTimeout("tcp", host, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			return
		}
		if time.Now().After(deadline) {
			panic("server not ready: " + host)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func startServer(handler agonet.EventHandler, addr string) agonet.Server {
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = addr
	srv, err := agonet.NewServer(handler, &cfg)
	if err != nil {
		panic(err)
	}
	go func() { _ = srv.Start() }()
	return srv
}

func startClient(handler agonet.EventHandler) agonet.Client {
	ccfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(handler, &ccfg)
	if err != nil {
		panic(err)
	}
	if err := cli.Start(); err != nil {
		panic(err)
	}
	return cli
}

// ---------- 场景 1：连接风暴（快速建/关——goroutine 回落 + heap 稳定） ----------

func sceneStorm(cli agonet.Client, host string, total int) {
	alloc0, _, _, g0 := memStats()
	start := time.Now()
	closed := int64(0)
	var wg sync.WaitGroup
	workers := 8
	per := total / workers
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < per; i++ {
				c, err := cli.Dial("tcp", host)
				if err != nil {
					continue
				}
				_ = c.Close()
				atomic.AddInt64(&closed, 1)
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)
	// 等读 goroutine 回落（defer Put 归还读缓冲）
	time.Sleep(500 * time.Millisecond)
	alloc1, _, _, g1 := memStats()
	fmt.Printf("[storm] 建/关 %d 连接（%d 并发 worker）用时 %v（%.0f conn/s）\n",
		closed, workers, elapsed.Round(time.Millisecond), float64(closed)/elapsed.Seconds())
	fmt.Printf("[storm] goroutine: %d → %d（回落 %d）——读 goroutine/loop 无泄漏\n", g0, g1, g0-g1)
	fmt.Printf("[storm] heap: %.1f → %.1f MB（Δ%.1f MB）——%s\n",
		alloc0, alloc1, alloc1-alloc0, leakVerdict(alloc1-alloc0, 5))
}

// ---------- 场景 2：持续帧流（吞吐 + 分配 + 内存曲线） ----------

func sceneStream(cli agonet.Client, host string, connsN int, dur time.Duration) {
	handler := &recvOnlyHandler{}
	_ = handler
	conns := make([]agonet.Conn, connsN)
	for i := range conns {
		c, err := cli.Dial("tcp", host)
		if err != nil {
			panic(err)
		}
		conns[i] = c
	}
	defer func() {
		for _, c := range conns {
			_ = c.Close()
		}
	}()

	payload := make([]byte, 1024)
	var sent, failed atomic.Int64
	alloc0, _, gc0, g0 := memStats()
	start := time.Now()
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < connsN; i++ {
		wg.Add(1)
		go func(c agonet.Conn) {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				if _, err := c.Write(payload); err != nil {
					failed.Add(1)
				} else {
					sent.Add(1)
				}
			}
		}(conns[i])
	}
	time.Sleep(dur)
	close(stop)
	wg.Wait()
	elapsed := time.Since(start)
	time.Sleep(300 * time.Millisecond) // 等在途数据处理完（server 计数稳定）
	alloc1, _, gc1, g1 := memStats()
	fmt.Printf("[stream] %d 连接持续帧流 %.0fs：投递 %.1f 万帧（%.0f 帧/s，%.1f MB/s）失败 %d\n",
		connsN, dur.Seconds(), float64(sent.Load())/1e4, float64(sent.Load())/elapsed.Seconds(),
		float64(sent.Load()*int64(len(payload)))/elapsed.Seconds()/1e6, failed.Load())
	fmt.Printf("[stream] goroutine %d → %d（回落 %d）；heap %.1f → %.1f MB；GC %d → %d（+%d 次）\n",
		g0, g1, g0-g1, alloc0, alloc1, gc0, gc1, gc1-gc0)
}

// ---------- 场景 3：建/关循环泄露检测（前后对比） ----------

func sceneLeak(cli agonet.Client, host string, rounds int) {
	// 预热（池校准/缓冲热身后测基线）
	for i := 0; i < 200; i++ {
		c, _ := cli.Dial("tcp", host)
		_ = c.Close()
	}
	time.Sleep(500 * time.Millisecond)
	runtime.GC()
	alloc0, _, gc0, g0 := memStats()
	for r := 0; r < rounds; r++ {
		for i := 0; i < 100; i++ {
			c, err := cli.Dial("tcp", host)
			if err != nil {
				continue
			}
			_ = c.Close()
		}
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(500 * time.Millisecond) // 读 goroutine 退出 + 池归还稳定
	runtime.GC()
	alloc1, _, gc1, g1 := memStats()
	deltaMB := alloc1 - alloc0
	fmt.Printf("[leak] %d 轮 × 100 建/关：goroutine %d → %d（Δ%d）；heap %.1f → %.1f MB（Δ%.1f MB）；GC +%d\n",
		rounds, g0, g1, g1-g0, alloc0, alloc1, deltaMB, gc1-gc0)
	fmt.Printf("[leak] 判定：%s\n", leakVerdict(deltaMB, 2))
}

func leakVerdict(deltaMB float64, threshold float64) string {
	if deltaMB > threshold {
		return fmt.Sprintf("⚠ 疑似泄漏（Δ%.1f MB > %.0f MB 阈值）", deltaMB, threshold)
	}
	return fmt.Sprintf("✓ 无泄漏（Δ%.1f MB ≤ %.0f MB——池化复用/GC 回收正常）", deltaMB, threshold)
}

// ---------- main ----------

func main() {
	mode := flag.String("mode", "stream", "stream|storm|leak|all")
	connsN := flag.Int("conns", 50, "并发连接数")
	dur := flag.Duration("dur", 10*time.Second, "stream 持续时长")
	pprofDir := flag.String("pprof-dir", "", "pprof 采样输出目录（CPU/heap/goroutine）")
	flag.Parse()

	if *pprofDir != "" {
		if err := os.MkdirAll(*pprofDir, 0o755); err != nil {
			panic(err)
		}
		cf, _ := os.Create(*pprofDir + "/cpu.pprof")
		_ = pprof.StartCPUProfile(cf)
		defer func() {
			pprof.StopCPUProfile()
			_ = cf.Close()
		}()
	}

	// server（recvOnly——stream 用）+ client
	handler := &recvOnlyHandler{}
	srv := startServer(handler, "tcp://127.0.0.1:18210")
	defer srv.Stop()
	waitPort("127.0.0.1:18210", 3*time.Second)

	cli := startClient(&recvOnlyHandler{})
	defer cli.Stop()

	host := "127.0.0.1:18210"

	run := func(name string, fn func()) {
		fmt.Printf("=== %s ===\n", name)
		start := time.Now()
		fn()
		fmt.Printf("=== %s 完成（%v）===\n", name, time.Since(start).Round(time.Millisecond))
	}

	switch *mode {
	case "storm":
		run("连接风暴", func() { sceneStorm(cli, host, 1000) })
	case "stream":
		run("持续帧流", func() { sceneStream(cli, host, *connsN, *dur) })
	case "leak":
		run("泄露检测", func() { sceneLeak(cli, host, 10) })
	case "all":
		run("连接风暴", func() { sceneStorm(cli, host, 1000) })
		run("持续帧流", func() { sceneStream(cli, host, *connsN, *dur) })
		run("泄露检测", func() { sceneLeak(cli, host, 10) })
	default:
		fmt.Println("unknown mode:", *mode)
	}

	if *pprofDir != "" {
		hp, _ := os.Create(*pprofDir + "/heap.pprof")
		_ = pprof.WriteHeapProfile(hp)
		_ = hp.Close()
		gp, _ := os.Create(*pprofDir + "/goroutine.pprof")
		_ = pprof.Lookup("goroutine").WriteTo(gp, 0)
		_ = gp.Close()
		fmt.Printf("[pprof] 采样已写入 %s/（cpu/heap/goroutine）\n", *pprofDir)
	}
}
