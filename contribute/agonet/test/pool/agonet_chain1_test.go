package pool_test

// 链1 P1 T1：B1 池满不整服关（红测）。
// 独立包（M2：全局池替换类测试须独立目录 + -p 1 防污染同包并行测试）。
//
// 红（当前）：池满 → listenStream Submit 读 goroutine 失败 → return err → defer eng.shutdown(err)
//            → 整服关闭 → listener 关闭 → 新连接建立失败
// 绿（修复后 R1）：读 goroutine 原生 go 不走池 → 池满不影响 accept → 服务存活
//
// 对应详细设计：[[agonet-链1修复详细设计]] §五 T1。

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/aif-go/ag-core/contribute/agonet"
	goroutline "github.com/aif-go/ag-core/contribute/agonet/pkg/pool/goroutline" // 目录 goroutline，包名 goroutine——显式别名
)

func TestChain1_T1_PoolFull_ServerSurvives(t *testing.T) {
	// ① 替换全局池为小容量（4）+ Nonblocking（池满即 Submit 失败）
	oldPool := goroutline.DefaultWorkerPool
	small, err := ants.NewPool(4, ants.WithOptions(ants.Options{Nonblocking: true}))
	if err != nil {
		t.Fatal(err)
	}
	goroutline.DefaultWorkerPool = small
	defer func() { goroutline.DefaultWorkerPool = oldPool }()
	defer small.Release()

	// ② 打满池：4 个阻塞任务占满全部 worker（不归还）
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
	wg.Wait() // 4 个 worker 全忙
	defer close(block)

	// ③ 起 server：当前 accept 后 Submit 读 goroutine 失败 → shutdown（红）；修复后原生 go 正常（绿）
	const addr = "127.0.0.1:18093"
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = "tcp://" + addr
	srv, err := agonet.NewServer(&agonet.BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	startErr := make(chan error, 1)
	go func() { startErr <- srv.Start() }()
	// 同步：等 server 就绪（listener 绑定 = run() 已完成 s.eng 赋值）后再 Stop，
	// 规避 agonet 既有竞态（server.go s.eng 无同步写读，-race 下 Stop 与 Start 竞争）
	ready := make(chan struct{})
	defer func() {
		select {
		case <-ready:
		case <-time.After(3 * time.Second):
		}
		srv.Stop()
	}()

	// ④ 断言：连接能建立【且服务持续存活】
	// 轮询等 server 就绪（消除"Dial 早于 listener bind"的启动竞态）：
	// 红（当前 B1）：Submit 失败 → shutdown → listener 关 → 轮询超时红
	// 绿（R1 后）：Dial 成功 → 继续
	dialOK := func() bool {
		c, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err != nil {
			return false
		}
		_ = c.Close()
		return true
	}

	deadline := time.Now().Add(3 * time.Second)
	ready2 := false
	for time.Now().Before(deadline) {
		if dialOK() {
			ready2 = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	close(ready)
	if !ready2 {
		t.Fatal("RED: server not accepting (B1: 整服 shutdown)")
	}
	time.Sleep(300 * time.Millisecond) // 给 shutdown 留出时间（红测时引擎已关）
	if !dialOK() {
		t.Fatal("RED: server died after pool full (B1: 整服 shutdown)")
	}
}
