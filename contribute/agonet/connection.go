package agonet

import (
	"ag-core/contribute/agonet/pkg/aerrors"
	"ag-core/contribute/agonet/pkg/pool/byteslice"
	goroutine "ag-core/contribute/agonet/pkg/pool/goroutline"
	"errors"
	"io"
	"net"
	"time"

	"github.com/smallnest/ringbuffer"
	"github.com/valyala/bytebufferpool"
)

type netErr struct {
	c   *conn
	err error
}

type tcpConn struct {
	c *conn
	// b *bbPool.ByteBuffer
	b *bytebufferpool.ByteBuffer
}

type openConn struct {
	c  *conn
	cb func()
}

type conn struct {
	ctx    any // user-defined context
	loop   *eventloop
	buffer *bytebufferpool.ByteBuffer // reuse memory of inbound data as a temporary buffer
	cache  []byte                     // temporary cache for the inbound data
	// net.Conn            // original connection
	rawConn    net.Conn // original connection
	localAddr  net.Addr // local server addr
	remoteAddr net.Addr // remote addr
	// inboundBuffer elastic.RingBuffer // buffer for data from the remote
	// inboundBuffer *bytebufferpool.ByteBuffer
	inboundBytes  []byte
	inboundBuffer *ringbuffer.RingBuffer
	// inboundBuffer *ring.Ring

	// add by sirius
	// rawReader *bufio.Reader
	// rawWriter *bufio.Writer
	// *bufio.Reader
	// *bufio.Writer
}

func newStreamConn(el *eventloop, nc net.Conn, ctx any) (c *conn) {
	// inboundBuffer := ring.New(0x10000) // 64KB 环形缓冲区
	// rawReader := bufio.NewReader(nc)
	// rawWriter := bufio.NewWriter(nc)
	inboundBytes := byteslice.Get(1024 * 1024 * 64)
	// ringBuffer := ringbuffer.New(1024) // TODO 环形缓冲区大小
	ringBuffer := ringbuffer.NewBuffer(inboundBytes)
	return &conn{
		ctx:     ctx,
		loop:    el,
		buffer:  bytebufferpool.Get(),
		rawConn: nc,
		// Conn:          nc,
		localAddr:  nc.LocalAddr(),
		remoteAddr: nc.RemoteAddr(),
		// inboundBuffer: bytebufferpool.Get(), // TODO RingBuffer 实现
		// inboundBuffer: inboundBuffer,
		inboundBytes:  inboundBytes,
		inboundBuffer: ringBuffer,
		// rawReader:     rawReader,
		// rawWriter:     rawWriter,
		// Reader: rawReader,
		// Writer: rawWriter,
	}
}

func packTCPConn(c *conn, buf []byte) *tcpConn {
	// b := bbPool.Get()
	if buf == nil {
		return &tcpConn{c: c, b: nil}
	}

	b := bytebufferpool.Get()
	_, _ = b.Write(buf)
	return &tcpConn{c: c, b: b}
}

func unpackTCPConn(tc *tcpConn) *conn {
	if tc.c.buffer == nil { // the connection has been closed
		return nil
	}
	_, _ = tc.c.buffer.Write(tc.b.B)
	bytebufferpool.Put(tc.b) // 归还 ByteBuffer 到池
	tc.b = nil

	// if tc.c.rawReader == nil { // the connection has been closed
	// 	// if tc.c.Reader == nil { // the connection has been closed
	// 	return nil
	// }
	return tc.c
}

// ### conn implements Reader ###
var _ Reader = (*conn)(nil)

func (c *conn) resetBuffer() {
	c.buffer.Reset()
	c.inboundBuffer.Reset()
}

// Read implements io.Reader.
func (c *conn) Read(p []byte) (n int, err error) {
	if c.inboundBuffer.IsEmpty() {
		n = copy(p, c.buffer.B)
		c.buffer.B = c.buffer.B[n:]
		if n == 0 && len(p) > 0 {
			err = io.ErrShortBuffer
		}
		return
	}
	n, _ = c.inboundBuffer.Read(p)
	if n == len(p) {
		return
	}
	m := copy(p[n:], c.buffer.B)
	n += m
	c.buffer.B = c.buffer.B[m:]
	return
}

// WriteTo implements io.WriterTo.
func (c *conn) WriteTo(w io.Writer) (n int64, err error) {
	// return c.rawReader.WriteTo(w)

	if !c.inboundBuffer.IsEmpty() {
		if n, err = c.inboundBuffer.WriteTo(w); err != nil {
			return
		}
	}

	if c.buffer == nil {
		return 0, nil
	}
	defer c.buffer.Reset()
	return c.buffer.WriteTo(w)
}

func (c *conn) Next(n int) (buf []byte, err error) {
	inBufferLen := c.inboundBuffer.Length()
	if totalLen := inBufferLen + c.buffer.Len(); n > totalLen {
		return nil, io.ErrShortBuffer
	} else if n <= 0 {
		n = totalLen
	}
	if c.inboundBuffer.IsEmpty() {
		buf = c.buffer.B[:n]
		c.buffer.B = c.buffer.B[n:]
		return
	}

	// buf = make([]byte, n) // TODO 考虑从池中获取
	buf = byteslice.Get(n)
	_, err = c.Read(buf)
	return
}

func (c *conn) Peek(n int) (buf []byte, err error) {
	inBufferLen := c.inboundBuffer.Length()
	if totalLen := inBufferLen + c.buffer.Len(); n > totalLen {
		// 若有效数据长度小于 n，则返回错误
		return nil, io.ErrShortBuffer
	} else if n <= 0 {
		// 若 n 小于等于 0，则返回所有有效数据
		n = totalLen
	}
	if c.inboundBuffer.IsEmpty() {
		return c.buffer.B[:n], err
	}

	// head := make([]byte, 0, n)
	buf = byteslice.Get(n)

	// head, tail := c.inboundBuffer.Peek()
	pn, err := c.inboundBuffer.Peek(buf)
	if err != nil {
		return nil, err
	}

	if len(buf) == pn {
		return buf, err
	}

	if inBufferLen >= n {
		return
	}

	remaining := n - inBufferLen
	buf = append(buf, c.buffer.B[:remaining]...)

	c.cache = buf
	return
}

func (c *conn) Discard(n int) (discarded int, err error) {
	// discarded, err = c.rawReader.Discard(n)
	// return
	if len(c.cache) > 0 {
		byteslice.Put(c.cache)
		c.cache = nil
	}

	inBufferLen := c.inboundBuffer.Length()
	if totalLen := inBufferLen + c.buffer.Len(); n >= totalLen || n <= 0 {
		c.resetBuffer()
		return totalLen, nil
	}

	if c.inboundBuffer.IsEmpty() {
		c.buffer.B = c.buffer.B[n:]
		return n, nil
	}

	// discarded, _ := c.inboundBuffer.Discard(n)
	discarded, _ = discardRB(c.inboundBuffer, n)
	if discarded < inBufferLen {
		return discarded, nil
	}

	remaining := n - inBufferLen
	c.buffer.B = c.buffer.B[remaining:]
	return n, nil
}

// ### conn implements Writer ###Q
var _ Writer = (*conn)(nil)

// Write implements io.Writer.
func (c *conn) Write(p []byte) (n int, err error) {

	if c.rawConn == nil {
		return 0, net.ErrClosed
	}
	return c.rawConn.Write(p)
}

// ReadFrom implements io.ReaderFrom.
func (c *conn) ReadFrom(r io.Reader) (n int64, err error) {
	if c.rawConn != nil {
		return io.Copy(c.rawConn, r)
	}
	return 0, net.ErrClosed
}

func (c *conn) Flush() error {
	// if c.rawConn == nil || c.rawWriter == nil {
	// 	return net.ErrClosed
	// }
	// return c.rawWriter.Flush()
	return nil
}

// ### conn implements Conn ###
var _ Conn = (*conn)(nil)

func (c *conn) Context() (ctx any) {
	return c.ctx
}

func (c *conn) EventLoop() EventLoop {
	return c.loop
}

func (c *conn) SetContext(ctx any) {
	c.ctx = ctx
}

func (c *conn) Close() (err error) {
	closeFn := func() error {
		return c.loop.close(c, nil)
	}

	select {
	case c.loop.ch <- closeFn:
	default:
		// If the event-loop channel is full, asynchronize this operation to avoid blocking the eventloop.
		err = goroutine.DefaultWorkerPool.Submit(func() {
			c.loop.ch <- closeFn
		})
	}

	return
}

func (c *conn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *conn) SetDeadline(t time.Time) error {
	tcpConn, ok := c.rawConn.(*net.TCPConn)
	if !ok {
		return aerrors.ErrUnsupportedOp
	}
	return tcpConn.SetDeadline(t)
}

func (c *conn) SetReadDeadline(t time.Time) error {
	tcpConn, ok := c.rawConn.(*net.TCPConn)
	if !ok {
		return aerrors.ErrUnsupportedOp
	}
	return tcpConn.SetReadDeadline(t)
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	tcpConn, ok := c.rawConn.(*net.TCPConn)
	if !ok {
		return aerrors.ErrUnsupportedOp
	}
	return tcpConn.SetWriteDeadline(t)
}

func (c *conn) release() {
	c.ctx = nil
	c.localAddr = nil
	if c.rawConn != nil {
		c.rawConn = nil
		c.remoteAddr = nil
	}

	c.inboundBuffer.CloseWithError(nil)
	c.inboundBuffer = nil
	byteslice.Put(c.inboundBytes) // 归还缓冲区

	bytebufferpool.Put(c.buffer) // 归还缓冲区
	c.buffer = nil
}

func (c *conn) NetConn() net.Conn {
	return c.rawConn
}

// discardRB 基于ringbuffer现有API实现丢弃n个字节的逻辑（非阻塞）
// 返回实际丢弃的字节数和错误
func discardRB(rb *ringbuffer.RingBuffer, n int) (discarded int, err error) {
	if n < 0 {
		return 0, errors.New("discard count cannot be negative")
	}
	if n == 0 {
		return 0, nil
	}

	// 临时缓冲区：每次读取最多4KB（减少内存分配，也可直接用n长度）
	// tempBuf := make([]byte, min(n, 4096))
	tempBuf := make([]byte, 4096)
	totalDiscarded := 0

	//获取缓冲区当前可用数据长度（核心：只丢弃实际存在的数据）
	available := rb.Length()
	if available == 0 {
		return 0, nil // 无数据可丢弃，直接返回
	}

	if available < n {
		n = available
	}

	for totalDiscarded < n {
		// 计算剩余需要丢弃的字节数
		remaining := n - totalDiscarded
		// 本次读取的长度：不超过临时缓冲区大小，也不超过剩余需要丢弃的长度
		// readLen := min(len(tempBuf), remaining)
		readLen := min(4096, remaining)

		// 读取数据（丢弃，不使用tempBuf）
		nRead, err := rb.Read(tempBuf[:readLen])
		if nRead > 0 {
			totalDiscarded += nRead
		}

		// 处理错误：非阻塞模式下，读空会返回错误，直接终止
		if err != nil {
			// 如果已经读取了部分数据，返回已读取的数量+错误；否则直接返回错误
			if totalDiscarded > 0 {
				return totalDiscarded, err
			}
			return 0, err
		}

		// 如果读取到0字节（无数据），终止循环
		if nRead == 0 {
			break
		}
	}

	return totalDiscarded, nil
}
