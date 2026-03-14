package agonet

import (
	"ag-core/contribute/agonet/pkg/aerrors"
	"fmt"
	"sync/atomic"
)

// TODO 一期暂不实现
type eventloop struct {
	ch           chan any           // channel for event-loop
	idx          int                // index of event-loop in event-loops
	eng          *engine            // engine in loop
	connCount    int32              // number of active connections in event-loop
	connections  map[*conn]struct{} // TCP connection map: fd -> conn
	eventHandler EventHandler       // user eventHandler
}

func (el *eventloop) run() (err error) {
	defer func() {
		el.eng.shutdown(err)
		for c := range el.connections {
			_ = el.close(c, nil)
		}
	}()
	// TODO 实现事件循环
	return nil
}

func (el *eventloop) close(c *conn, err error) error {
	if _, ok := el.connections[c]; c.Conn == nil || !ok {
		return nil // ignore stale wakes.
	}

	delete(el.connections, c)
	el.incConn(-1)
	action := el.eventHandler.OnClose(c, err)
	err = c.Conn.Close()
	c.release()
	if err != nil {
		return fmt.Errorf("failed to close connection=%s in event-loop(%d): %v", c.remoteAddr, el.idx, err)
	}

	return el.handleAction(c, action)
}

func (el *eventloop) incConn(delta int32) {
	atomic.AddInt32(&el.connCount, delta)
}

func (c *conn) release() {
	c.ctx = nil
	c.localAddr = nil
	if c.Conn != nil {
		c.Conn = nil
		c.remoteAddr = nil
	}
	// c.inboundBuffer.Done()
	// bbPool.Put(c.buffer)
	// c.buffer = nil
}

func (el *eventloop) handleAction(c *conn, action Action) error {
	switch action {
	case None:
		return nil
	case Close:
		return el.close(c, nil)
	case Shutdown:
		return aerrors.ErrEngineShutdown
	default:
		return nil
	}
}
