// Package main idle 样例：IdleStateHandler 心跳/空闲超时。
// 运行：go run ./contribute/agonet/example/idle
// 回归：go test -race ./contribute/agonet/example/idle
//
// 演示：pipeline 中配置 IdleStateHandler（本样例：全空闲 2 秒即触发），
// 业务 handler 收 IdleStateEvent（经 ctx.Trigger 传播）→ 记录/关闭连接。
// ⚠️ 实现约束：idle 时间最小 1 秒（handler_idle.go max(…, time.Second)）。
package main

import (
	"fmt"
	"net"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
	"github.com/aif-go/ag-core/contribute/agonet/simple"
)

// newIdleHandler 构造 idle 管理 handler：ALL_IDLE 超时 → 记录事件 + 关闭连接。
func newIdleHandler(readerIdle, writerIdle, allIdle int64, events chan<- simple.IdleStateEvent) (agonet.EventHandler, error) {
	return simple.NewSimpleEventHandlerWithOptions(
		simple.WithChannelInitializer(func(ch simple.Channel) error {
			ch.Pipeline().AddLast(
				simple.IdleStateHandler(readerIdle, writerIdle, allIdle, time.Second),
				simple.EventHandlerFunc(func(ctx simple.EventContext, event any) {
					if ev, ok := event.(simple.IdleStateEvent); ok {
						select {
						case events <- ev:
						default:
						}
						// 业务：空闲超时视为心跳丢失，关闭连接
						ctx.Channel().Close(fmt.Errorf("idle timeout: %s", ev.State))
					}
				}),
			)
			return nil
		}),
	)
}

const addr = "tcp://127.0.0.1:18883"
const host = "127.0.0.1:18883"

func main() {
	events := make(chan simple.IdleStateEvent, 8)
	handler, err := newIdleHandler(2, 2, 2, events)
	if err != nil {
		panic(err)
	}

	// 服务端（阻塞式启动，须 goroutine 包裹）
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = addr
	srv, err := agonet.NewServer(handler, &cfg)
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

	// 客户端：连接后保持静默 → 服务端 2 秒后判定 ALL_IDLE 并关闭
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

	// 等 idle 事件（最长 4 秒）
	select {
	case ev := <-events:
		fmt.Printf("idle ok: connection closed by %s\n", ev.State)
	case <-time.After(4 * time.Second):
		panic("idle event timeout")
	}
}
