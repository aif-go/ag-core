package main

// framed 样例集成回归：simple 层 lengthField 编解码。
// 覆盖：帧往返、粘包（两帧合并一次发送）、半包（一帧拆两段）、handler 链顺序、大帧边界。
// 端口 18882 独立。
//
// 架构约束（C7）：client 侧帧收经 simple handler channel（loop 内解码），测试 goroutine 不直接 Read。

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
	"github.com/aif-go/ag-core/contribute/agonet/simple"
)

// startFrameServer 启动帧 echo server（阻塞式 Start 须 goroutine 包裹）。
func startFrameServer(t *testing.T, h agonet.EventHandler) {
	t.Helper()
	cfg := agonet.DefaultServerConfig()
	cfg.Addr = addr
	srv, err := agonet.NewServer(h, &cfg)
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
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("server port not ready")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// newFrameClient 创建 simple 帧 client（got 收解码后帧）。
func newFrameClient(t *testing.T, got chan []byte) agonet.Client {
	t.Helper()
	h, err := newRecvFrameHandler(got)
	if err != nil {
		t.Fatal(err)
	}
	cfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(h, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cli.Stop() })
	return cli
}

// recvNFrames 累积收 n 帧（每帧 channel 一次投递）。
func recvNFrames(t *testing.T, got chan []byte, n int) [][]byte {
	t.Helper()
	frames := make([][]byte, 0, n)
	deadline := time.After(5 * time.Second)
	for len(frames) < n {
		select {
		case msg := <-got:
			frames = append(frames, msg)
		case <-deadline:
			t.Fatalf("recv timeout: got %d of %d frames", len(frames), n)
		}
	}
	return frames
}

func TestFramed_RoundTrip(t *testing.T) {
	startFrameServer(t, echoHandlerOrFail(t))
	got := make(chan []byte, 4)
	cli := newFrameClient(t, got)

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	payload := []byte("single-frame-payload")
	if _, err := conn.Write(encodeFrame(payload)); err != nil {
		t.Fatal(err)
	}
	frames := recvNFrames(t, got, 1)
	if !bytes.Equal(frames[0], payload) {
		t.Fatalf("frame mismatch: got %q want %q", frames[0], payload)
	}
}

func TestFramed_StickyPackets(t *testing.T) {
	// 粘包：两帧合并一次发送，decoder 应拆出两帧且顺序正确
	startFrameServer(t, echoHandlerOrFail(t))
	got := make(chan []byte, 4)
	cli := newFrameClient(t, got)

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	f1 := encodeFrame([]byte("sticky-1"))
	f2 := encodeFrame([]byte("sticky-2"))
	if _, err := conn.Write(append(f1, f2...)); err != nil {
		t.Fatal(err)
	}
	frames := recvNFrames(t, got, 2)
	if string(frames[0]) != "sticky-1" || string(frames[1]) != "sticky-2" {
		t.Fatalf("sticky frames mismatch: %q %q", frames[0], frames[1])
	}
}

func TestFramed_SplitPackets(t *testing.T) {
	// 半包：一帧拆两段分次发送，decoder 应缓存拼出完整帧
	startFrameServer(t, echoHandlerOrFail(t))
	got := make(chan []byte, 4)
	cli := newFrameClient(t, got)

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	payload := []byte("split-across-writes-payload")
	frame := encodeFrame(payload)
	half := len(frame) / 2
	if _, err := conn.Write(frame[:half]); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond) // 保证两次写入被内核分成独立 TCP 段
	if _, err := conn.Write(frame[half:]); err != nil {
		t.Fatal(err)
	}
	frames := recvNFrames(t, got, 1)
	if !bytes.Equal(frames[0], payload) {
		t.Fatalf("split frame mismatch: got %q want %q", frames[0], payload)
	}
}

func TestFramed_HandlerChainOrder(t *testing.T) {
	// handler 链顺序：入站传播 head→tail，按 AddLast 顺序执行（decoder 拆帧后逐级接力）
	var order []string
	h, err := simple.NewSimpleEventHandlerWithOptions(
		simple.WithChannelInitializer(func(ch simple.Channel) error {
			ch.Pipeline().AddLast(
				simple.NewLengthFieldDecoder(binary.BigEndian, 65535, 0, 2, 0, 2),
				simple.NewLengthFieldEncoder(binary.BigEndian, 2, 0, false), // 出站编码：C 的 echo 才有帧头
				simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
					order = append(order, "A")
					ctx.FireRead(msg) // SimpleInboundHandler 是终端消费者语义，接力须显式 FireRead
				}),
				simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
					order = append(order, "B")
					ctx.FireRead(msg)
				}),
				simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
					order = append(order, "C")
					ctx.Write(msg) // 链尾消费 + 回显（出站编码）
				}),
			)
			return nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	startFrameServer(t, h)
	got := make(chan []byte, 4)
	cli := newFrameClient(t, got)

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if _, err := conn.Write(encodeFrame([]byte("chain"))); err != nil {
		t.Fatal(err)
	}
	recvNFrames(t, got, 1)
	if len(order) != 3 || order[0] != "A" || order[1] != "B" || order[2] != "C" {
		t.Fatalf("handler chain order mismatch: %v", order)
	}
}

func TestFramed_LargeFrame(t *testing.T) {
	// 大帧边界：60KB payload（帧总长 < maxFrameLength 65535）
	startFrameServer(t, echoHandlerOrFail(t))
	got := make(chan []byte, 4)
	cli := newFrameClient(t, got)

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	payload := bytes.Repeat([]byte("L"), 60000)
	if _, err := conn.Write(encodeFrame(payload)); err != nil {
		t.Fatal(err)
	}
	frames := recvNFrames(t, got, 1)
	if !bytes.Equal(frames[0], payload) {
		t.Fatalf("large frame mismatch: got %d bytes want %d", len(frames[0]), len(payload))
	}
}

// echoHandlerOrFail 构造服务端帧 handler（失败即 Fatal）。
func echoHandlerOrFail(t *testing.T) agonet.EventHandler {
	t.Helper()
	h, err := newEchoFrameHandler()
	if err != nil {
		t.Fatal(err)
	}
	return h
}
