package agonet

import (
	"net"

	"github.com/valyala/bytebufferpool"
)

type tcpConn struct {
	c *conn
	// b *bbPool.ByteBuffer
	b *bytebufferpool.ByteBuffer
}

type conn struct {
	ctx any // user-defined context
	// cache []byte // temporary cache for the inbound data
	// rawConn    net.Conn // original connection
	net.Conn            // original connection
	localAddr  net.Addr // local server addr
	remoteAddr net.Addr // remote addr
}

func newStreamConn(nc net.Conn, ctx any) (c *conn) {
	return &conn{
		ctx: ctx,
		// rawConn:    nc,
		Conn:       nc,
		localAddr:  nc.LocalAddr(),
		remoteAddr: nc.RemoteAddr(),
	}
}

func packTCPConn(c *conn, buf []byte) *tcpConn {
	// b := bbPool.Get()
	b := bytebufferpool.Get()
	_, _ = b.Write(buf)
	return &tcpConn{c: c, b: b}
}
