package agonet

// 链1 T21（review 行内 P1/P2 修正）：客户端连接打开语义。
// 红态（修复前）：
//   - cb 无参——MaxConn 超限拒绝无错误通道——Enroll 返回 (c, nil)（伪成功）——reviewer 实测
//   - drain 超时丢弃 openConn 不调 cb——Stop 已返回但 Dial 永久挂起——reviewer 实测
// 绿态（修复后）：cb 携带错误（ErrMaxConnRejected / ErrEngineShutdown / panic）——
//   Enroll/Dial 返回失败而非伪成功/挂起。

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
)

// 红测（修复前）：MaxConn=1 客户端连续 Dial——第二次返回 (c, nil) 伪成功
func TestChain1_T21_MaxConnRejectReturnsError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0") // 可达目标——net.Dial 成功进入 EnrollContext
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := DefaultClientConfig()
	opts, err := BuildOptionsWithConfig(cfg.Config)
	if err != nil {
		t.Fatal(err)
	}
	opts.MaxConn = 1
	cli, err := NewClientWithOptions(&BuiltinEventEngine{}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cli.Stop() })

	c1, err := cli.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close() // 保持打开——占用配额

	c2, err := cli.Dial("tcp", ln.Addr().String())
	if err == nil {
		_ = c2.Close()
		t.Fatal("RED: second Dial returned success despite MaxConn=1 (expect ErrMaxConnRejected)")
	}
	if !errors.Is(err, aerrors.ErrMaxConnRejected) {
		t.Fatalf("RED: second Dial err=%v (expect ErrMaxConnRejected)", err)
	}
}

// 红测（修复前）：Stop 时排队中的 Dial 永久挂起（drain 丢弃 openConn 不调 cb）
func TestChain1_T21_ShutdownRejectsQueuedDial(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0") // 可达目标
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	// 慢 OnOpen：首个 Dial 处理期间其余排队；ShutdownTimeout 短于单事件耗时——drain 必超时
	h := &slowOpenHandler{BuiltinEventEngine: &BuiltinEventEngine{}, block: 30 * time.Millisecond}
	cfg := DefaultClientConfig()
	opts, err := BuildOptionsWithConfig(cfg.Config)
	if err != nil {
		t.Fatal(err)
	}
	opts.ShutdownTimeout = 20 * time.Millisecond
	cli, err := NewClientWithOptions(h, opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}

	const n = 5
	type res struct {
		idx int
		err error
	}
	results := make(chan res, n)
	for i := 0; i < n; i++ {
		go func(i int) {
			_, err := cli.Dial("tcp", ln.Addr().String())
			results <- res{i, err}
		}(i)
	}

	time.Sleep(60 * time.Millisecond) // 首个 Dial 的 OnOpen 处理中，其余排队（ch 未满）
	go func() { _ = cli.Stop() }()    // 触发关闭 → drain → 超时 → rejectRemaining

	for i := 0; i < n; i++ {
		select {
		case r := <-results:
			_ = r // 成功或错误皆可——核心断言：全部返回不挂起
		case <-time.After(3 * time.Second):
			t.Fatalf("RED: Dial[%d] hung after client Stop (drain must reject queued openConn)", i)
		}
	}
}

// slowOpenHandler 慢 OnOpen（模拟处理耗时 > ShutdownTimeout 的事件）。
type slowOpenHandler struct {
	*BuiltinEventEngine
	block time.Duration
}

func (h *slowOpenHandler) OnOpen(c Conn) ([]byte, Action) {
	time.Sleep(h.block)
	return nil, None
}
