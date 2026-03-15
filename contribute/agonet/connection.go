package agonet

import (
	"ag-core/contribute/agonet/pkg/aerrors"
	goroutine "ag-core/contribute/agonet/pkg/pool/goroutline"
	"bufio"
	"io"
	"net"
	"time"

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
	// cache []byte // temporary cache for the inbound data
	// net.Conn            // original connection
	rawConn    net.Conn // original connection
	localAddr  net.Addr // local server addr
	remoteAddr net.Addr // remote addr
	// inboundBuffer elastic.RingBuffer // buffer for data from the remote
	inboundBuffer *bytebufferpool.ByteBuffer
	// inboundBuffer *ring.Ring

	// add by sirius
	rawReader *bufio.Reader
	rawWriter *bufio.Writer
	// *bufio.Reader
	// *bufio.Writer
}

func newStreamConn(el *eventloop, nc net.Conn, ctx any) (c *conn) {
	// inboundBuffer := ring.New(0x10000) // 64KB 环形缓冲区
	rawReader := bufio.NewReader(nc)
	rawWriter := bufio.NewWriter(nc)
	return &conn{
		ctx:     ctx,
		loop:    el,
		buffer:  bytebufferpool.Get(),
		rawConn: nc,
		// Conn:          nc,
		localAddr:     nc.LocalAddr(),
		remoteAddr:    nc.RemoteAddr(),
		inboundBuffer: bytebufferpool.Get(), // TODO RingBuffer 实现
		// inboundBuffer: inboundBuffer,
		rawReader: rawReader,
		rawWriter: rawWriter,
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
	// if tc.c.buffer == nil { // the connection has been closed
	// 	return nil
	// }
	// _, _ = tc.c.buffer.Write(tc.b.B)
	// bytebufferpool.Put(tc.b) // 归还 ByteBuffer 到池
	// tc.b = nil
	if tc.c.rawReader == nil { // the connection has been closed
		// if tc.c.Reader == nil { // the connection has been closed
		return nil
	}
	return tc.c
}

// ### conn implements Reader ###
var _ Reader = (*conn)(nil)

// Read implements io.Reader.
func (c *conn) Read(p []byte) (n int, err error) {
	return c.rawReader.Read(p)
}

// WriteTo implements io.WriterTo.
func (c *conn) WriteTo(w io.Writer) (n int64, err error) {
	return c.rawReader.WriteTo(w)
}

func (c *conn) Next(n int) (buf []byte, err error) {
	totalLen := c.rawReader.Size()
	if totalLen < n {
		return nil, io.ErrShortBuffer
	} else if n <= 0 {
		n = totalLen
	}

	buf = make([]byte, n) // TODO 考虑是否需要从池中获取 ByteBuffer

	_, err = c.rawReader.Read(buf)

	return
}

func (c *conn) Peek(n int) (buf []byte, err error) {
	totalLen := c.rawReader.Size()
	if totalLen < n {
		return nil, io.ErrShortBuffer
	} else if n <= 0 {
		n = totalLen
	}

	// buf = make([]byte, n) // TODO 考虑是否需要从池中获取 ByteBuffer
	buf, err = c.rawReader.Peek(n)
	return
}

func (c *conn) Discard(n int) (discarded int, err error) {
	discarded, err = c.rawReader.Discard(n)
	return
}

// ### conn implements Writer ###Q
var _ Writer = (*conn)(nil)

// Write implements io.Writer.
func (c *conn) Write(p []byte) (n int, err error) {
	if c.rawConn == nil || c.rawWriter == nil {
		return 0, net.ErrClosed
	}

	n, err = c.rawWriter.Write(p)
	c.rawWriter.Flush()
	return
}

// ReadFrom implements io.ReaderFrom.
func (c *conn) ReadFrom(r io.Reader) (n int64, err error) {
	if c.rawConn == nil || c.rawWriter == nil {
		n, err = io.Copy(c.rawWriter, r)
	}

	return
}

func (c *conn) Flush() error {
	if c.rawConn == nil || c.rawWriter == nil {
		return net.ErrClosed
	}
	return c.rawWriter.Flush()
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
