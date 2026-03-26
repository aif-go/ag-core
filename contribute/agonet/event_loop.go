package agonet

import (
	"ag-core/contribute/agonet/pkg/aerrors"
	goroutine "ag-core/contribute/agonet/pkg/pool/goroutline"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync/atomic"

	"github.com/petermattis/goid"
)

type eventloop struct {
	ch           chan any           // channel for event-loop
	idx          int                // index of event-loop in event-loops
	eng          *engine            // engine in loop
	connCount    int32              // number of active connections in event-loop
	connections  map[*conn]struct{} // TCP connection map: fd -> conn
	eventHandler EventHandler       // user eventHandler

	goroutineId int64
}

func (el *eventloop) run() (err error) {
	defer func() {
		el.eng.shutdown(err)
		for c := range el.connections {
			_ = el.close(c, nil)
		}
	}()

	// TODO 绑定事件循环到当前线程
	// if el.eng.opts.LockOSThread {
	// 	runtime.LockOSThread()
	// 	defer runtime.UnlockOSThread()
	// }

	// 获取协程id
	id := goid.Get()
	el.goroutineId = id

	for i := range el.ch {
		switch v := i.(type) {
		case error:
			err = v
		case *netErr:
			err = el.close(v.c, v.err)
		case *openConn:
			err = el.open(v)
		case *tcpConn:
			err = el.read(unpackTCPConn(v))
		case func() error:
			err = v()
		}

		if errors.Is(err, aerrors.ErrEngineShutdown) {
			// el.getLogger().Debugf("event-loop(%d) is exiting in terms of the demand from user, %v", el.idx, err)
			slog.Debug(fmt.Sprintf("event-loop(%d) is exiting in terms of the demand from user, %v", el.idx, err))
			break
		} else if err != nil {
			// el.getLogger().Debugf("event-loop(%d) got a nonlethal error: %v", el.idx, err)
			slog.Debug(fmt.Sprintf("event-loop(%d) got a nonlethal error: %v", el.idx, err))
		}
	}

	return nil
}

func (el *eventloop) open(oc *openConn) error {
	if oc.cb != nil {
		defer oc.cb()
	}

	c := oc.c
	el.connections[c] = struct{}{}
	el.incConn(1)

	out, action := el.eventHandler.OnOpen(c)
	if out != nil {
		if _, err := c.rawConn.Write(out); err != nil {
			return err
		}
	}

	return el.handleAction(c, action)
}

func (el *eventloop) read(c *conn) error {
	if _, ok := el.connections[c]; !ok {
		return nil // ignore stale wakes.
	}
	// 调用消息处理函数
	action := el.eventHandler.OnTraffic(c)
	switch action {
	case None:
	case Close:
		return el.close(c, nil)
	case Shutdown:
		return aerrors.ErrEngineShutdown
	}

	// 剩余未处理的字节写入缓存
	_, err := c.inboundBuffer.Write(c.buffer.B)

	if err != nil {
		// return el.close(c, err)
		// TODO 判断异常，长度不够的要扩容inboundBuffer
	}

	c.buffer.Reset()

	return nil
}

func (el *eventloop) close(c *conn, err error) error {
	// if _, ok := el.connections[c]; c.Conn == nil || !ok {
	if _, ok := el.connections[c]; c.rawConn == nil || !ok {
		return nil // ignore stale wakes.
	}

	delete(el.connections, c)
	el.incConn(-1)

	action := el.eventHandler.OnClose(c, err)

	err = c.rawConn.Close()

	c.release()
	if err != nil {
		return fmt.Errorf("failed to close connection=%s in event-loop(%d): %v", c.remoteAddr, el.idx, err)
	}

	return el.handleAction(c, action)
}

func (el *eventloop) incConn(delta int32) {
	atomic.AddInt32(&el.connCount, delta)
}

func (el *eventloop) countConn() int32 {
	return atomic.LoadInt32(&el.connCount)
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

// ### eventloop implements EventLoop ###
var _ EventLoop = (*eventloop)(nil)

func (el *eventloop) Register(ctx context.Context, addr net.Addr) (<-chan RegisteredResult, error) {
	// TODO
	return nil, nil
}

func (el *eventloop) Enroll(ctx context.Context, c net.Conn) (<-chan RegisteredResult, error) {
	// TODO
	return nil, nil
}

func (el *eventloop) Close(Conn) error {
	// TODO
	return nil
}

// Deprecated
func (el *eventloop) InEventLoop() bool {
	// check goroutine id
	cid := goid.Get()

	return el.goroutineId == cid
}

func (el *eventloop) ExecuteInEventLoop(fn func() error) error {
	var err error
	select {
	case el.ch <- fn:
	default:
		// If the event-loop channel is full, asynchronize this operation to avoid blocking the eventloop.
		err = goroutine.DefaultWorkerPool.Submit(func() {
			el.ch <- fn
		})
		// return aerrors.ErrEventLoopQueueFull //TODO 事件循环队列满的情况如何处理，是否异常
	}

	return err
}
