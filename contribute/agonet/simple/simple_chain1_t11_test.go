package simple

// 链1 低风险收尾批测试（D7 异常链防重入，集成形态）：
//   T11——业务 HandleException panic → 异常链防重入 → 连接被关闭（非无限循环/进程崩溃）。
// 落位 simple 包（agonet 包测试无法 import simple——循环依赖）。
//
// 红态（修复前）：异常链重入——panic 冒泡/循环，连接不关或进程崩溃。

import (
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
)

func TestChain1_T11_ExceptionChainReentry(t *testing.T) {
	// simple server：pipeline = 业务 inbound（OnMsg panic，触发异常链）
	//                + 业务异常处理器（HandleException panic，触发重入）
	// 断言：连接被关闭（TryFireException 重入拒绝 → action=Close）而非无限循环/进程崩溃
	const addr = "127.0.0.1:18100"

	handler, err := NewSimpleEventHandlerWithOptions(
		WithChannelInitializer(func(ch Channel) error {
			ch.Pipeline().AddLast(
				// 业务 handler：收到数据 panic（模拟业务 bug 触发异常链）
				// 注意：pipeline 入站传 agonet.Reader（未解码），泛型须匹配 reader 类型
				NewSimpleInboundHandler(func(ctx InboundContext, msg agonet.Reader) {
					panic("business panic")
				}),
				// 业务异常处理器：panic（模拟异常链重入——修复前循环/崩溃）
				ExceptionHandlerFunc(func(ctx ExceptionContext, ex error) {
					panic("exception handler panic")
				}),
			)
			return nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	cfg := agonet.DefaultServerConfig()
	cfg.Addr = "tcp://" + addr
	srv, err := agonet.NewServer(handler, &cfg)
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

	// 等就绪
	deadline := time.Now().Add(3 * time.Second)
	for {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("server port not ready")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// 原生 TCP 客户端：发送触发数据 → 服务端异常链 → 断言连接被关闭（标准 Read 语义，
	// 避免 agonet conn.Read 的 C7 契约干扰——ErrShortBuffer 会被误判为"关闭"）
	nc, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Close()

	if _, err := nc.Write([]byte("trigger")); err != nil {
		t.Fatal(err)
	}

	// 连接应被服务端关闭（异常链重入拒绝 → action=Close）
	_ = nc.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 16)
	for {
		if _, err := nc.Read(buf); err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				t.Fatal("RED: connection not closed (exception chain reentry not contained)")
			}
			return // EOF/RST = 连接被关闭 ✅
		}
	}
}
