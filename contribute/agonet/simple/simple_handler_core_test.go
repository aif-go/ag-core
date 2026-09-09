package simple

// 核心功能正确性测试（默认构建、绿测）：SimpleEventHandler 端到端 + channel 语义。
// 覆盖：OnOpen 建 channel/pipeline、编解码数据抵达 OnMsg、OnClose 链路、channel 状态。

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
)

func TestSimpleHandler_EndToEnd_DecodeToOnMsg(t *testing.T) {
	const addr = "tcp://127.0.0.1:18082"
	got := make(chan []byte, 4)

	handler, err := NewSimpleEventHandlerWithOptions(
		WithChannelInitializer(func(ch Channel) error {
			ch.Pipeline().AddLast(
				NewLengthFieldDecoder(binary.BigEndian, 65535, 0, 2, 0, 2),
				NewSimpleInboundHandler(func(ctx InboundContext, msg []byte) {
					got <- msg
				}),
			)
			return nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	cfg := agonet.DefaultServerConfig()
	cfg.Addr = addr
	server, err := agonet.NewServer(handler, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	// server.Start() 为阻塞式启动（阻塞运行直到 Stop），须 goroutine 包裹
	startErr := make(chan error, 1)
	go func() { startErr <- server.Start() }()
	// 注意 defer 顺序（LIFO）：先注册 check（后执行），再注册 Stop（先执行）——
	// 必须等 Stop 调用解除 eng.stop 的阻塞后，Start 才返回，check 才能收到结果。
	defer func() {
		select {
		case err := <-startErr:
			if err != nil {
				t.Errorf("server.Start returned err=%v, expect nil after Stop", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server.Start did not return after Stop")
		}
	}()
	defer server.Stop()

	// 等待端口就绪
	deadline := time.Now().Add(2 * time.Second)
	for {
		c, err := net.DialTimeout("tcp", "127.0.0.1:18082", 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("server port not ready")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// 编码两帧发送
	enc := NewLengthFieldEncoder(binary.BigEndian, 2, 0, false)
	var wire []byte
	for _, body := range [][]byte{[]byte("hello"), []byte("world")} {
		ctx := &captureOutboundCtx{}
		enc.HandleWrite(ctx, body)
		wire = append(wire, ctx.got...)
	}

	conn, err := net.Dial("tcp", "127.0.0.1:18082")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Write(wire); err != nil {
		t.Fatal(err)
	}

	// 期待两帧 OnMsg
	expect := []string{"hello", "world"}
	for i, want := range expect {
		select {
		case msg := <-got:
			if string(msg) != want {
				t.Fatalf("msg[%d]=%q, expect %q", i, msg, want)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("timeout waiting msg[%d]=%q", i, want)
		}
	}
	_ = conn.Close()
}

func TestChannel_ClosedWrite_ReturnsError(t *testing.T) {
	// channel 关闭后 Write 返回错误（IsActive 守卫）。
	ch := &channel{closed: 1}
	if ch.IsActive() {
		t.Fatal("closed channel must be inactive")
	}
	if err := ch.Write([]byte("x")); err == nil {
		t.Fatal("Write on closed channel must error")
	}
}

func TestChannel_Trigger_Closed_Noop(t *testing.T) {
	// 关闭后 Trigger 静默忽略。
	ch := &channel{closed: 1}
	ch.Trigger("event") // 不应 panic
}

func TestChannel_ID_Pipeline_Accessors(t *testing.T) {
	p := NewPipeline()
	ch := newChannel(mockSimpleConn{}, p)
	if ch.ID() == 0 {
		t.Fatal("channel id must be non-zero")
	}
	if ch.Pipeline() != p {
		t.Fatal("pipeline mismatch")
	}
	if !ch.IsActive() {
		t.Fatal("fresh channel must be active")
	}
}

// mockSimpleConn 最小 agonet.Conn 桩（仅测 channel 访问器）。
type mockSimpleConn struct{}

func (mockSimpleConn) Read(p []byte) (int, error)          { return 0, nil }
func (mockSimpleConn) Write(p []byte) (int, error)         { return len(p), nil }
func (mockSimpleConn) Close() error                        { return nil }
func (mockSimpleConn) LocalAddr() net.Addr                 { return mockAddr{} }
func (mockSimpleConn) RemoteAddr() net.Addr                { return mockAddr{} }
func (mockSimpleConn) SetDeadline(t time.Time) error       { return nil }
func (mockSimpleConn) SetReadDeadline(t time.Time) error   { return nil }
func (mockSimpleConn) SetWriteDeadline(t time.Time) error  { return nil }
func (mockSimpleConn) Context() any                        { return nil }
func (mockSimpleConn) SetContext(ctx any)                  {}
func (mockSimpleConn) EventLoop() agonet.EventLoop         { return nil }
func (mockSimpleConn) NetConn() net.Conn                   { return nil }
func (mockSimpleConn) Next(n int) ([]byte, error)          { return nil, nil }
func (mockSimpleConn) ReadFrom(r io.Reader) (int64, error) { return 0, nil }
func (mockSimpleConn) WriteTo(w io.Writer) (int64, error)  { return 0, nil }
func (mockSimpleConn) Peek(n int) ([]byte, error)          { return nil, nil }
func (mockSimpleConn) Discard(n int) (int, error)          { return 0, nil }
func (mockSimpleConn) InboundBuffered() int                { return 0 }
func (mockSimpleConn) AsyncWrite(buf []byte, cb agonet.AsyncCallback) error {
	return nil
}
func (mockSimpleConn) Flush() error { return nil }
func (mockSimpleConn) Wake(callback agonet.AsyncCallback) error { return nil }

type mockAddr struct{}

func (mockAddr) Network() string { return "mock" }
func (mockAddr) String() string  { return "mock" }
