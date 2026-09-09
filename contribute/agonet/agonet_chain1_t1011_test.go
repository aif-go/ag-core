package agonet

// 链1 低风险收尾批测试（white-box，无 build tag 常驻）：
//   T10（A4）：eventloop 内 Dial 重入 → 快速失败（ErrDialInEventLoop），非死锁
//   T11（D7）见 simple 包测试（simple_chain1_t11_test.go——agonet 测试无法 import simple，循环依赖）
//
// 红态：T10 修复前 func 卡在 connOpened（loop 死锁，测试超时）。

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
)

// ==================== T10：A4 eventloop 内 Dial 快速失败 ====================

func TestChain1_T10_DialInEventLoopFailsFast(t *testing.T) {
	// 单 loop engine + client（white-box 直接构造，避免 server 启动的注册竞争）
	rootCtx, shutdown := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(rootCtx)
	eng := &engine{
		opts:         &Options{},
		eventHandler: &BuiltinEventEngine{},
		turnOff:      shutdown,
		concurrency: struct {
			*errgroup.Group
			ctx context.Context
		}{eg, ctx},
	}
	eng.eventLoops = new(roundRobinLoadBalancer)
	el := &eventloop{ch: make(chan any, 1024), eng: eng, eventHandler: eng.eventHandler, connections: make(map[*conn]struct{})}
	eng.eventLoops.register(el)
	eg.Go(el.run)
	t.Cleanup(func() {
		eng.shutdown(nil)
		eng.closeEventLoops()
		_ = eg.Wait()
	})

	cli := &client{opts: &Options{}, eng: eng, eventHandler: eng.eventHandler}

	// 可达目标：真实 listener（红态：net.Dial 成功 → EnrollContext 等 connOpened → loop 死锁）
	ln, err := net.Listen("tcp", "127.0.0.1:18099")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	// 在 loop 内调 Dial（投递 func 事件——func 由 loop 执行 = handler 上下文）
	dialErr := make(chan error, 1)
	if !el.send(func() error {
		_, err := cli.Dial("tcp", ln.Addr().String())
		dialErr <- err
		return nil
	}) {
		t.Fatal("send failed")
	}

	select {
	case err := <-dialErr:
		if !errors.Is(err, aerrors.ErrDialInEventLoop) {
			t.Fatalf("err=%v, want ErrDialInEventLoop（A4 守卫：loop 内 Dial 快速失败）", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("RED: Dial hung in eventloop (A4 deadlock)")
	}
}
