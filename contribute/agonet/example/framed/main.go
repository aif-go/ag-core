// Package main framed 样例：simple 层 lengthField 编解码 + handler 链。
// 运行：go run ./contribute/agonet/example/framed
// 回归：go test -race ./contribute/agonet/example/framed
//
// 演示：SimpleEventHandler + WithChannelInitializer 配置 pipeline：
//   - NewLengthFieldDecoder：入站按 2 字节大端长度字段拆帧（粘包/半包自动处理）
//   - NewLengthFieldEncoder：出站自动加长度头
//   - 业务 inbound handler：收解码后的完整帧
//
// 帧格式：2 字节长度（= payload 长度，不含长度字段）+ payload。
package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
	"github.com/aif-go/ag-core/contribute/agonet/simple"
)

// newEchoFrameHandler 服务端帧 handler：解码入站帧 → 回显（出站 encoder 编码）。
func newEchoFrameHandler() (agonet.EventHandler, error) {
	return simple.NewSimpleEventHandlerWithOptions(
		simple.WithChannelInitializer(func(ch simple.Channel) error {
			ch.Pipeline().AddLast(
				simple.NewLengthFieldDecoder(binary.BigEndian, 65535, 0, 2, 0, 2),
				simple.NewLengthFieldEncoder(binary.BigEndian, 2, 0, false),
				simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
					// 业务：收到完整帧，原样回显（ctx.Write 走 pipeline 出站 → encoder 编码）
					ctx.Write(msg)
				}),
			)
			return nil
		}),
	)
}

// newRecvFrameHandler 客户端帧 handler：解码入站帧 → channel 交给业务侧。
func newRecvFrameHandler(got chan []byte) (agonet.EventHandler, error) {
	return simple.NewSimpleEventHandlerWithOptions(
		simple.WithChannelInitializer(func(ch simple.Channel) error {
			ch.Pipeline().AddLast(
				simple.NewLengthFieldDecoder(binary.BigEndian, 65535, 0, 2, 0, 2),
				simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
					got <- msg
				}),
			)
			return nil
		}),
	)
}

// encodeFrame 手工编码一帧（2 字节长度 = payload 长度，不含长度字段自身）。
// 与 NewLengthFieldEncoder(binary.BigEndian, 2, 0, false) 语义一致；
// decoder 侧 frameLength = offset(0) + fieldLength(2) + adjustment(0) + L = 2 + L。
// 粘包/半包测试需要直接操纵字节流，故手工编码而非走 client pipeline。
func encodeFrame(payload []byte) []byte {
	buf := make([]byte, 2+len(payload))
	binary.BigEndian.PutUint16(buf[:2], uint16(len(payload)))
	copy(buf[2:], payload)
	return buf
}

const addr = "tcp://127.0.0.1:18882"
const host = "127.0.0.1:18882"

func main() {
	// 服务端（阻塞式启动，须 goroutine 包裹）
	srvHandler, err := newEchoFrameHandler()
	if err != nil {
		panic(err)
	}
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = addr
	srv, err := agonet.NewServer(srvHandler, &cfg)
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

	// 客户端：simple 帧 handler 收帧；手工发帧字节（2 帧一次发送 = 粘包场景）
	got := make(chan []byte, 4)
	cliHandler, err := newRecvFrameHandler(got)
	if err != nil {
		panic(err)
	}
	ccfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(cliHandler, &ccfg)
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

	f1 := encodeFrame([]byte("frame-1"))
	f2 := encodeFrame([]byte("frame-2"))
	if _, err := conn.Write(append(f1, f2...)); err != nil { // 粘包：两帧合并一次发送
		panic(err)
	}

	received := make([]string, 0, 2)
	deadline = time.Now().Add(3 * time.Second)
	for len(received) < 2 {
		select {
		case msg := <-got:
			received = append(received, string(msg))
		case <-time.After(time.Until(deadline)):
			panic(fmt.Sprintf("frame recv timeout: got %v", received))
		}
	}
	if received[0] != "frame-1" || received[1] != "frame-2" {
		panic(fmt.Sprintf("frame mismatch: %v", received))
	}
	fmt.Printf("framed ok: %v\n", received)
}
