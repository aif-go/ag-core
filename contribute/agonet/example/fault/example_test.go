package main

// fault 样例集成回归：故障域套件（链1 锁定行为的组合回归）。
// 覆盖：池满服务存活（B1）、引擎关闭 goroutine 回落（A6）、引擎关闭后 Dial 不挂起（A3）、
//       转池 worker 归还（C 类集成形态）、MaxConn 超限拒绝（R6）。
// 端口 18885/18886 独立；池替换类测试独立包 + -p 1（M2）。
// 注：C 类精确红测（loop 死 + ch 满构造）在 agonet 包 T3b；此处为集成形态回归保护。

import (
	"net"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/aif-go/ag-core/contribute/agonet"
	goroutline "github.com/aif-go/ag-core/contribute/agonet/pkg/pool/goroutline" // 目录 goroutline，包名 goroutine
)

// startServer 启动 server（阻塞式 Start 须 goroutine 包裹），返回 srv 供显式 Stop。
func startServer(t *testing.T, h agonet.EventHandler, opts *agonet.Options) agonet.Server {
	t.Helper()
	var srv agonet.Server
	var err error
	if opts != nil {
		srv, err = agonet.NewServerWithOptions(h, []string{addr}, opts)
	} else {
		cfg := agonet.DefaultServerConfig()
		cfg.Addr = addr
		srv, err = agonet.NewServer(h, &cfg)
	}
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

// replacePool 替换全局池为小容量 Nonblocking 池（defer 恢复）。
func replacePool(t *testing.T, size int) *ants.Pool {
	t.Helper()
	old := goroutline.DefaultWorkerPool
	p, err := ants.NewPool(size, ants.WithOptions(ants.Options{Nonblocking: true}))
	if err != nil {
		t.Fatal(err)
	}
	goroutline.DefaultWorkerPool = p
	t.Cleanup(func() { goroutline.DefaultWorkerPool = old })
	t.Cleanup(p.Release)
	return p
}

func dialOK() bool {
	c, err := net.DialTimeout("tcp", host, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

// TestFault_PoolFullServerSurvives B1：池满（4 阻塞任务占满 Nonblocking 池）→ 服务端读路径已出池 → 服务存活。
func TestFault_PoolFullServerSurvives(t *testing.T) {
	replacePool(t, 4)
	block := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		if err := goroutline.DefaultWorkerPool.Submit(func() {
			wg.Done()
			<-block
		}); err != nil {
			t.Fatalf("fill pool submit err: %v", err)
		}
	}
	wg.Wait() // 池满
	defer close(block)

	startServer(t, &agonet.BuiltinEventEngine{}, nil)

	deadline := time.Now().Add(3 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		if dialOK() {
			ready = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !ready {
		t.Fatal("RED: server not accepting (B1: 整服 shutdown)")
	}
	time.Sleep(300 * time.Millisecond)
	if !dialOK() {
		t.Fatal("RED: server died after pool full (B1: 整服 shutdown)")
	}
}

// TestFault_GoroutineReclaim A6：引擎关闭后读 goroutine 回落。
func TestFault_GoroutineReclaim(t *testing.T) {
	srv := startServer(t, &agonet.BuiltinEventEngine{}, nil)

	conns := make([]net.Conn, 0, 3)
	for i := 0; i < 3; i++ {
		c, err := net.DialTimeout("tcp", host, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, c)
	}
	time.Sleep(200 * time.Millisecond)
	peak := runtime.NumGoroutine()

	if err := srv.Stop(); err != nil {
		t.Fatal(err)
	}
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

// TestFault_DialAfterShutdown A3：引擎关闭后 Dial 不挂起（快速返回错误）。
func TestFault_DialAfterShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:18886")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(&agonet.BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	if err := cli.Stop(); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := cli.Dial("tcp", ln.Addr().String())
		done <- err
	}()
	select {
	case <-done:
		// 绿：快速返回错误（不挂起）
	case <-time.After(2 * time.Second):
		t.Fatal("RED: Dial hung after engine shutdown (A3)")
	}
}

// TestFault_PoolWorkerReclaim C 类集成形态：慢消费打满 ch → AsyncWrite 转池 → 引擎关闭 → 池 worker 归还。
func TestFault_PoolWorkerReclaim(t *testing.T) {
	p := replacePool(t, 4)
	srv := startServer(t, &slowHandler{}, nil)

	cliCfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(&agonet.BuiltinEventEngine{}, &cliCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	defer cli.Stop()

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// 大数据流（打满 server 侧 el.ch）+ 高频 AsyncWrite 触发转池路径。
	// 注意：AsyncWrite 传 nil——fn 内 c.Write(nil) 不阻塞（写真实数据会在 client loop 内
	// 阻塞于 TCP 发送缓冲满（server 慢消费），连锁导致 ch 不消费 → 误判 worker 滞留）。
	awDone := make(chan struct{})
	go func() {
		defer close(awDone)
		for i := 0; i < 5000; i++ {
			_ = conn.AsyncWrite(nil, nil)
		}
	}()

	time.Sleep(3 * time.Second) // 等 ch 满 + 转池任务进入阻塞
	<-awDone

	if err := srv.Stop(); err != nil { // 引擎关闭 → ctx.Done → 转池任务归还（绿）/ 滞留（红）
		t.Fatal(err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if p.Running() == 0 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("RED: pool workers stuck (running=%d)", p.Running())
}

// TestFault_MaxConnLimit R6：MaxConn=1 时第二个连接被静默拒绝（连接被服务端关闭）。
func TestFault_MaxConnLimit(t *testing.T) {
	startServer(t, &agonet.BuiltinEventEngine{}, &agonet.Options{MaxConn: 1})

	c1, err := net.DialTimeout("tcp", host, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()
	c2, err := net.DialTimeout("tcp", host, time.Second)
	if err != nil {
		t.Fatal(err) // TCP 层仍可建立（护栏是 accept 后关闭）
	}
	defer c2.Close()

	// 第二个连接被静默关闭（rawConn.Close → 读 EOF/错误）
	_ = c2.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 16)
	for {
		if _, err := c2.Read(buf); err != nil {
			return // 连接被服务端关闭（MaxConn 静默拒绝生效）
		}
	}
}
