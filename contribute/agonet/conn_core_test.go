package agonet

// 核心功能正确性测试（默认构建、绿测）：conn 缓冲语义。
// 覆盖：Read/Next/Peek/Discard/WriteTo/InboundBuffered、空帧、超界、关闭后 ErrClosed。

import (
	"bytes"
	"io"
	"net"
	"testing"
)

// newTestConn 构造带指定缓冲数据的 conn（el=nil，仅测缓冲路径，不触 eventloop）。
func newTestConn(data []byte) *conn {
	c := newStreamConn(nil, mockNetConn{}, nil)
	if len(data) > 0 {
		_, _ = c.buffer.Write(data)
	}
	return c
}

func TestConn_Next_Advance(t *testing.T) {
	c := newTestConn([]byte("hello"))
	buf, err := c.Next(3)
	if err != nil {
		t.Fatalf("Next err: %v", err)
	}
	if string(buf) != "hel" {
		t.Fatalf("Next(3)=%q, expect hel", buf)
	}
	if c.InboundBuffered() != 2 {
		t.Fatalf("InboundBuffered=%d, expect 2", c.InboundBuffered())
	}
}

func TestConn_Next_Exceeds(t *testing.T) {
	c := newTestConn([]byte("hi"))
	_, err := c.Next(10)
	if err != io.ErrShortBuffer {
		t.Fatalf("Next(10) err=%v, expect io.ErrShortBuffer", err)
	}
}

func TestConn_Next_ZeroEmptyFrame(t *testing.T) {
	// 空帧：Next(0) 返回空且不消费（F1 修复后语义）。
	c := newTestConn([]byte("hello"))
	buf, err := c.Next(0)
	if err != nil {
		t.Fatalf("Next(0) err: %v", err)
	}
	if len(buf) != 0 {
		t.Fatalf("Next(0) len=%d, expect 0", len(buf))
	}
	if c.InboundBuffered() != 5 {
		t.Fatalf("Next(0) must not consume, InboundBuffered=%d, expect 5", c.InboundBuffered())
	}
}

func TestConn_Peek_NoAdvance(t *testing.T) {
	c := newTestConn([]byte("hello"))
	buf, err := c.Peek(3)
	if err != nil {
		t.Fatalf("Peek err: %v", err)
	}
	if string(buf) != "hel" {
		t.Fatalf("Peek(3)=%q, expect hel", buf)
	}
	if c.InboundBuffered() != 5 {
		t.Fatalf("Peek must not advance, InboundBuffered=%d, expect 5", c.InboundBuffered())
	}
}

func TestConn_Peek_AllWhenNonPositive(t *testing.T) {
	// Peek(0)/负值返回全部有效数据（不消费）。
	c := newTestConn([]byte("hello"))
	buf, err := c.Peek(0)
	if err != nil {
		t.Fatalf("Peek(0) err: %v", err)
	}
	if string(buf) != "hello" {
		t.Fatalf("Peek(0)=%q, expect hello", buf)
	}
	if c.InboundBuffered() != 5 {
		t.Fatalf("Peek must not advance, InboundBuffered=%d", c.InboundBuffered())
	}
}

func TestConn_Discard_Advance(t *testing.T) {
	c := newTestConn([]byte("hello"))
	n, err := c.Discard(3)
	if err != nil || n != 3 {
		t.Fatalf("Discard(3)=%d,%v, expect 3,nil", n, err)
	}
	if c.InboundBuffered() != 2 {
		t.Fatalf("InboundBuffered=%d, expect 2", c.InboundBuffered())
	}
}

func TestConn_Discard_AllOrMore(t *testing.T) {
	// n>=total：清空缓冲。
	c := newTestConn([]byte("hello"))
	n, _ := c.Discard(100)
	if n != 5 {
		t.Fatalf("Discard(100)=%d, expect 5", n)
	}
	if c.InboundBuffered() != 0 {
		t.Fatalf("InboundBuffered=%d, expect 0", c.InboundBuffered())
	}
}

func TestConn_Read_Basic(t *testing.T) {
	c := newTestConn([]byte("hello"))
	p := make([]byte, 3)
	n, err := c.Read(p)
	if err != nil || n != 3 {
		t.Fatalf("Read=%d,%v, expect 3,nil", n, err)
	}
	if string(p) != "hel" {
		t.Fatalf("Read got %q, expect hel", p)
	}
	if c.InboundBuffered() != 2 {
		t.Fatalf("InboundBuffered=%d, expect 2", c.InboundBuffered())
	}
}

func TestConn_Read_Empty(t *testing.T) {
	c := newTestConn(nil)
	p := make([]byte, 4)
	n, err := c.Read(p)
	if n != 0 || err != io.ErrShortBuffer {
		t.Fatalf("Read empty=%d,%v, expect 0,io.ErrShortBuffer", n, err)
	}
}

func TestConn_WriteTo_Drains(t *testing.T) {
	c := newTestConn([]byte("hello"))
	var w bytes.Buffer
	n, err := c.WriteTo(&w)
	if err != nil || n != 5 {
		t.Fatalf("WriteTo=%d,%v, expect 5,nil", n, err)
	}
	if w.String() != "hello" {
		t.Fatalf("WriteTo wrote %q, expect hello", w.String())
	}
}

func TestConn_InboundBuffered_Empty(t *testing.T) {
	c := newTestConn(nil)
	if c.InboundBuffered() != 0 {
		t.Fatalf("InboundBuffered=%d, expect 0", c.InboundBuffered())
	}
}

func TestConn_Closed_AllReaderMethodsErrClosed(t *testing.T) {
	// release 后：Read/Next/Peek/WriteTo 均返回 net.ErrClosed 而非 panic（C6 修复）。
	c := newTestConn([]byte("hello"))
	c.release()

	if _, err := c.Next(1); err != net.ErrClosed {
		t.Fatalf("Next after release err=%v, expect net.ErrClosed", err)
	}
	if _, err := c.Peek(1); err != net.ErrClosed {
		t.Fatalf("Peek after release err=%v, expect net.ErrClosed", err)
	}
	if _, err := c.Discard(1); err != net.ErrClosed {
		t.Fatalf("Discard after release err=%v, expect net.ErrClosed", err)
	}
	p := make([]byte, 4)
	if _, err := c.Read(p); err != net.ErrClosed {
		t.Fatalf("Read after release err=%v, expect net.ErrClosed", err)
	}
	var w bytes.Buffer
	if _, err := c.WriteTo(&w); err != net.ErrClosed {
		t.Fatalf("WriteTo after release err=%v, expect net.ErrClosed", err)
	}
}
