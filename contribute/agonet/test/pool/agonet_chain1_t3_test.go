package pool_test

// 链1 P1 T3：C 类池 worker 无滞留（红测）。
// 独立包（M2：全局池替换类测试须独立目录 + -p 1 防污染同包并行测试）。
//
// 背景：connection.go AsyncWrite/Close/Wake + Execute 的转池 goroutine 在 loop 死后
// 永久阻塞 ants worker（A2 变体，链1 分析盲区 C 类）。
// 红（修复前）：转池 goroutine 裸发送 c.loop.ch <- fn → ctx 取消不唤醒 → worker 滞留 → Pool.Running() 不回落
// 绿（修复后 R2）：el.send 的 ctx.Done → 丢弃 → worker 归还 → Pool.Running() 回落 0
//
// 对应详细设计：[[agonet-链1修复详细设计]] §五 T3 / §3.5 C 类说明。

import (
	"net"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/aif-go/ag-core/contribute/agonet"
	goroutline "github.com/aif-go/ag-core/contribute/agonet/pkg/pool/goroutline" // 目录 goroutline，包名 goroutine——显式别名
)

// slowHandler 慢消费 handler：OnTraffic sleep 拖慢 eventloop 消费 → el.ch（1024）积压打满。
type slowHandler struct {
	*agonet.BuiltinEventEngine
	conns chan agonet.Conn
}

func (h *slowHandler) OnOpen(c agonet.Conn) ([]byte, agonet.Action) {
	h.conns <- c
	return nil, agonet.None
}

func (h *slowHandler) OnTraffic(agonet.Conn) agonet.Action {
	time.Sleep(5 * time.Millisecond) // 慢消费 → 读 goroutine 投递积压 → ch 满
	return agonet.None
}

func TestChain1_T3_PoolWorkerReclaimAfterShutdown(t *testing.T) {
	// ① 替换全局池为小容量（4）+ 可观察（Running()）
	oldPool := goroutline.DefaultWorkerPool
	small, err := ants.NewPool(4, ants.WithOptions(ants.Options{Nonblocking: true}))
	if err != nil {
		t.Fatal(err)
	}
	goroutline.DefaultWorkerPool = small
	defer func() { goroutline.DefaultWorkerPool = oldPool }()
	defer small.Release()

	// ② 起 server（慢 handler）
	handler := &slowHandler{BuiltinEventEngine: &agonet.BuiltinEventEngine{}, conns: make(chan agonet.Conn, 4)}
	const addr = "127.0.0.1:18094"
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = "tcp://" + addr
	srv, err := agonet.NewServer(handler, &cfg)
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

	// ③ 轮询等 server 就绪（消除"Dial 早于 listener bind"的启动竞态）→ 建立连接，拿 agonet 侧 conn
	var cliConn net.Conn
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			cliConn = c
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	close(ready)
	if cliConn == nil {
		t.Fatal("server not ready")
	}
	defer cliConn.Close()
	var srvConn agonet.Conn
	select {
	case srvConn = <-handler.conns:
	case <-time.After(2 * time.Second):
		t.Fatal("server OnOpen not fired")
	}

	// ④ 打满 el.ch：客户端狂发数据（读 goroutine 投递 2048 包 > 1024 容量）
	//    + 慢 handler 拖慢消费 → ch 满 → AsyncWrite trySend 失败 → 转池 Submit → worker 阻塞 send
	go func() {
		buf := make([]byte, 1<<20) // 1MB
		for i := 0; i < 128; i++ { // 128MB → 2048 次 Read(64KB)
			if _, err := cliConn.Write(buf); err != nil {
				return
			}
		}
	}()
	awDone := make(chan struct{})
	go func() {
		defer close(awDone)
		for i := 0; i < 20000; i++ {
			// 池满 Submit 失败（③层）——正常，忽略；继续触发转池路径
			_ = srvConn.AsyncWrite(nil, nil)
		}
	}()

	// ⑤ 等 ch 满 + 转池任务进入阻塞（保守等待）
	time.Sleep(4 * time.Second)
	<-awDone

	// ⑥ 引擎关闭 → ctx.Done → 转池任务归还（绿）或滞留（红）
	if err := srv.Stop(); err != nil {
		t.Fatal(err)
	}

	// ⑦ 断言池 worker 回落 0
	deadline = time.Now().Add(5 * time.Second)
	reclaimed := false
	for time.Now().Before(deadline) {
		if small.Running() == 0 {
			reclaimed = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !reclaimed {
		t.Fatalf("RED: pool workers stuck after shutdown (running=%d)", small.Running())
	}
}
