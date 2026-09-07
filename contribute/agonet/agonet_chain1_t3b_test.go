package agonet

// 链1 P1 T3b：C 类转池 worker 无滞留（white-box 精确构造红态）。
// T3a（test/pool 集成场景）因动态平衡（loop 消费一个、读 goroutine 补一个，ch 恒有空位）
// 无法复现"转池任务永久滞留"——本测试直接构造 loop 死（无 run 循环）+ ch 满，
// 精确命中 C 类：AsyncWrite trySend 失败 → 转池 Submit → worker 阻塞投递。
// 红（裸发送）：worker 永久阻塞 → Running() 滞留
// 绿（R2 send）：ctx.Done → 丢弃 → worker 归还
//
// 对应详细设计：[[agonet-链1修复详细设计]] §五 T3 / §3.5 C 类说明。

import (
	"context"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"

	goroutine "github.com/aif-go/ag-core/contribute/agonet/pkg/pool/goroutline"
	"golang.org/x/sync/errgroup"
)

func TestChain1_T3b_PoolWorkerReclaimWhiteBox(t *testing.T) {
	oldPool := goroutine.DefaultWorkerPool
	small, err := ants.NewPool(4, ants.WithOptions(ants.Options{Nonblocking: true}))
	if err != nil {
		t.Fatal(err)
	}
	goroutine.DefaultWorkerPool = small
	defer func() { goroutine.DefaultWorkerPool = oldPool }()
	defer small.Release()

	// 构造 loop 死（无 run 循环）+ ch 满（容量 2）
	ctx, cancel := context.WithCancel(context.Background())
	eng := &engine{concurrency: struct {
		*errgroup.Group
		ctx context.Context
	}{ctx: ctx}}
	el := &eventloop{ch: make(chan any, 2), eng: eng}
	el.ch <- struct{}{}
	el.ch <- struct{}{} // ch 满
	c := &conn{loop: el}

	// AsyncWrite → trySend 失败（ch 满）→ Submit → worker 阻塞（send 等空位 / 裸发送永久）
	if err := c.AsyncWrite(nil, nil); err != nil {
		t.Fatalf("AsyncWrite err=%v (expect nil: Submit 成功，worker 阻塞中)", err)
	}
	time.Sleep(100 * time.Millisecond) // worker 进入阻塞
	if small.Running() != 1 {
		t.Fatalf("expected 1 worker busy (转池任务阻塞中), got %d", small.Running())
	}

	// 引擎关闭：红（裸发送）→ 永久滞留；绿（send）→ ctx.Done 归还
	cancel()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if small.Running() == 0 {
			return // 绿：worker 归还
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("RED: pool worker stuck after shutdown (running=%d)", small.Running())
}
