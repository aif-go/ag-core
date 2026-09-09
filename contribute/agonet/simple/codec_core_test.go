package simple

// 核心功能正确性测试（默认构建、绿测）：长度域编解码。
// 覆盖：单帧/粘包/半包/空 body/offset/adjustment/strip/往返/超限。
// 依赖：mockReader（模拟 agonet.Reader 语义：Peek 不消费、Next 消费、Discard 前进）。

import (
	"encoding/binary"
	"io"
	"testing"
)

// mockReader 最小 Reader 桩：内部 []byte 缓冲，语义对齐 conn（Peek 不消费、Next 消费）。
type mockReader struct {
	buf []byte
}

func (m *mockReader) Read(p []byte) (int, error) {
	if len(m.buf) == 0 {
		return 0, io.EOF
	}
	n := copy(p, m.buf)
	m.buf = m.buf[n:]
	return n, nil
}

func (m *mockReader) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(m.buf)
	m.buf = nil
	return int64(n), err
}

func (m *mockReader) Next(n int) ([]byte, error) {
	if n < 0 || n > len(m.buf) {
		return nil, io.ErrShortBuffer
	}
	b := m.buf[:n]
	m.buf = m.buf[n:]
	return b, nil
}

func (m *mockReader) Peek(n int) ([]byte, error) {
	if n > len(m.buf) {
		return nil, io.ErrShortBuffer
	}
	return m.buf[:n], nil
}

func (m *mockReader) Discard(n int) (int, error) {
	if n > len(m.buf) {
		n = len(m.buf)
	}
	m.buf = m.buf[n:]
	return n, nil
}

func (m *mockReader) InboundBuffered() int { return len(m.buf) }

// captureInboundCtx 捕获 decoder FireRead 产出的消息。
type captureInboundCtx struct {
	messages []any
}

func (c *captureInboundCtx) Channel() Channel             { return nil }
func (c *captureInboundCtx) Handler() Handler             { return nil }
func (c *captureInboundCtx) Write(message any)            {}
func (c *captureInboundCtx) Trigger(event any)            {}
func (c *captureInboundCtx) FireRead(message any)         { c.messages = append(c.messages, message) }
func (c *captureInboundCtx) FireWrite(message any)        {}
func (c *captureInboundCtx) FireExceptionCaught(ex error) {}
func (c *captureInboundCtx) FireEvent(event any)          {}

// newLenDecoder 默认配置：大端、长度域 2 字节、strip=2（解码产出纯 body）。
func newLenDecoder(maxFrame int) Decoder {
	return NewLengthFieldDecoder(binary.BigEndian, maxFrame, 0, 2, 0, 2)
}

// dec 单帧解码：body "hello" → 帧 [0x00 0x05 h e l l o]
func decHandle(dec Decoder, data []byte) ([]any, error) {
	ctx := &captureInboundCtx{}
	dec.HandleRead(ctx, &mockReader{buf: data})
	return ctx.messages, nil
}

func TestDecoder_SingleFrame(t *testing.T) {
	frame := append([]byte{0x00, 0x05}, []byte("hello")...)
	msgs, err := decHandle(newLenDecoder(1024), frame)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expect 1 message, got %d", len(msgs))
	}
	if got := string(msgs[0].([]byte)); got != "hello" {
		t.Fatalf("body=%q, expect %q", got, "hello")
	}
}

func TestDecoder_MultipleFrames_NoPipeliningLoss(t *testing.T) {
	// 两帧粘包：decoder 每次 HandleRead 只产出一帧；剩余由后续事件继续（不吞包）。
	f1 := append([]byte{0x00, 0x03}, []byte("abc")...)
	f2 := append([]byte{0x00, 0x02}, []byte("xy")...)
	all := append(append([]byte{}, f1...), f2...)

	msgs, err := decHandle(newLenDecoder(1024), all)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expect 1 message (剩余由后续事件处理), got %d", len(msgs))
	}
	if got := string(msgs[0].([]byte)); got != "abc" {
		t.Fatalf("first frame=%q, expect abc", got)
	}
}

func TestDecoder_HalfPacket_NoConsume(t *testing.T) {
	// 半包：长度域完整但 body 不足（length=5，仅 3 字节数据）→ 不产出、不消费。
	r := &mockReader{buf: append([]byte{0x00, 0x05}, []byte("he")...)} // 2+2=4 字节
	ctx := &captureInboundCtx{}
	newLenDecoder(1024).HandleRead(ctx, r)
	if len(ctx.messages) != 0 {
		t.Fatalf("half-packet must not produce message, got %d", len(ctx.messages))
	}
	if r.InboundBuffered() != 4 {
		t.Fatalf("half-packet must not consume buffer, InboundBuffered=%d, expect 4", r.InboundBuffered())
	}
}

func TestDecoder_EmptyBodyFrame(t *testing.T) {
	// 空 body 帧（长度域=0，strip=0）：产出 2 字节整帧（长度域），不吞后续数据。
	dec := NewLengthFieldDecoder(binary.BigEndian, 1024, 0, 2, 0, 0)
	frame := []byte{0x00, 0x00} // length=0
	msgs, err := decHandle(dec, frame)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expect 1 frame, got %d", len(msgs))
	}
	if len(msgs[0].([]byte)) != 2 {
		t.Fatalf("expect 2-byte frame, got %d bytes", len(msgs[0].([]byte)))
	}
}

func TestDecoder_LengthFieldOffset(t *testing.T) {
	// offset=2：前 2 字节为头部占位，长度域在 [2:4]；strip=4 取纯 body。
	dec := NewLengthFieldDecoder(binary.BigEndian, 1024, 2, 2, 0, 4)
	frame := append(append([]byte{0xAA, 0xBB}, []byte{0x00, 0x03}...), []byte("abc")...)
	msgs, err := decHandle(dec, frame)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expect 1 message, got %d", len(msgs))
	}
	if got := string(msgs[0].([]byte)); got != "abc" {
		t.Fatalf("body=%q, expect abc", got)
	}
}

func TestDecoder_Str_LengthFieldOffset(t *testing.T) {
	// 字符串版 offset=2：长度域在 [2:6]（"0003"），strip=6 取纯 body。
	dec := NewLengthFieldStrDecoder(1024, 2, 4, 0, 6)
	frame := append(append([]byte{0xAA, 0xBB}, []byte("0003")...), []byte("abc")...)
	msgs, err := decHandle(dec, frame)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expect 1 message, got %d", len(msgs))
	}
	if got := string(msgs[0].([]byte)); got != "abc" {
		t.Fatalf("body=%q, expect abc", got)
	}
}

func TestDecoder_LengthAdjustment(t *testing.T) {
	// adjustment=2：总帧长 = 长度域值 + adjustment + lengthFieldEndOffset(2)。
	// body=5 时长度域写 3（=body-adjustment），总帧 = 3+2+2 = 7 = 长度域2 + body5。
	dec := NewLengthFieldDecoder(binary.BigEndian, 1024, 0, 2, 2, 2)
	frame := append([]byte{0x00, 0x03}, []byte("abcde")...)
	msgs, err := decHandle(dec, frame)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got := string(msgs[0].([]byte)); got != "abcde" {
		t.Fatalf("body=%q, expect abcde", got)
	}
}

func TestDecoder_InitialBytesToStrip(t *testing.T) {
	// strip=2：跳过长度域，业务层只收 body。
	dec := NewLengthFieldDecoder(binary.BigEndian, 1024, 0, 2, 0, 2)
	frame := append([]byte{0x00, 0x03}, []byte("abc")...)
	msgs, err := decHandle(dec, frame)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got := string(msgs[0].([]byte)); got != "abc" {
		t.Fatalf("body=%q, expect abc", got)
	}
}

func TestDecoder_ExceededMaxFrameLength(t *testing.T) {
	// 声明长度超 maxFrameLength → HandleRead panic（utils.Assert 异常链，由 simple 层 recover 断连）。
	dec := NewLengthFieldDecoder(binary.BigEndian, 10, 0, 2, 0, 0)
	frame := append([]byte{0x00, 0x64}, make([]byte, 100)...) // length=100 > 10

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expect panic on exceeded frame length, got none")
		}
	}()
	dec.HandleRead(&captureInboundCtx{}, &mockReader{buf: frame})
}

func TestEncoder_NormalLengthField(t *testing.T) {
	enc := NewLengthFieldEncoder(binary.BigEndian, 2, 0, false)
	ctx := &captureOutboundCtx{}
	enc.HandleWrite(ctx, []byte("hello"))
	if len(ctx.got) != 2+5 {
		t.Fatalf("output len=%d, expect 7", len(ctx.got))
	}
	if got := binary.BigEndian.Uint16(ctx.got[:2]); got != 5 {
		t.Fatalf("length field=%d, expect 5", got)
	}
	if string(ctx.got[2:]) != "hello" {
		t.Fatalf("body=%q, expect hello", string(ctx.got[2:]))
	}
}

func TestEncoder_Decoder_RoundTrip(t *testing.T) {
	// 编码 → 解码往返一致；多帧连续编码后解码逐帧还原。
	enc := NewLengthFieldEncoder(binary.BigEndian, 2, 0, false)
	dec := newLenDecoder(4096)

	bodies := [][]byte{[]byte("hello"), []byte("world"), []byte("x")}
	var wire []byte
	for _, b := range bodies {
		ctx := &captureOutboundCtx{}
		enc.HandleWrite(ctx, b)
		wire = append(wire, ctx.got...)
	}

	// 逐帧解码（每轮消耗一帧）
	r := &mockReader{buf: wire}
	decoded := make([][]byte, 0)
	for r.InboundBuffered() >= 2 {
		ctx := &captureInboundCtx{}
		dec.HandleRead(ctx, r)
		for _, m := range ctx.messages {
			decoded = append(decoded, m.([]byte))
		}
		if len(ctx.messages) == 0 {
			break // 防死循环（异常保护）
		}
	}
	if len(decoded) != len(bodies) {
		t.Fatalf("decoded %d frames, expect %d", len(decoded), len(bodies))
	}
	for i := range bodies {
		if string(decoded[i]) != string(bodies[i]) {
			t.Fatalf("frame[%d]=%q, expect %q", i, decoded[i], bodies[i])
		}
	}
}

func TestEncoder_StringMessage(t *testing.T) {
	// 字符串消息同样支持（utils.MustToBytes）。
	enc := NewLengthFieldEncoder(binary.BigEndian, 2, 0, false)
	ctx := &captureOutboundCtx{}
	enc.HandleWrite(ctx, "hi")
	if string(ctx.got[2:]) != "hi" {
		t.Fatalf("body=%q, expect hi", string(ctx.got[2:]))
	}
}
