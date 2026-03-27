package simple

import (
	"ag-core/contribute/agonet"
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var _ Channel = (*channel)(nil)

type channel struct {
	id        int64
	conn      agonet.Conn
	eventloop agonet.EventLoop
	pipeline  Pipeline

	// untilWrite     bool
	closed int32
	// running        int32
	// closeErr error
	// writeLock      sync.Mutex // for sync write
}

func newChannel(conn agonet.Conn, pipeline Pipeline) Channel {

	id := time.Now().UnixNano() // TODO 获取通道ID
	executor := conn.EventLoop()

	channel := &channel{
		id:        id,
		conn:      conn,
		eventloop: executor,
		pipeline:  pipeline,
	}

	return channel
}

// ID get channel id
func (c *channel) ID() int64 {
	return c.id
}

// Pipeline get pipeline
func (c *channel) Pipeline() Pipeline {
	return c.pipeline
}

// EventLoop get event loop
func (c *channel) EventLoop() agonet.EventLoop {
	return c.eventloop
}

// IsActive check channel is active
func (c *channel) IsActive() bool {
	return atomic.LoadInt32(&c.closed) == 0
}

// LocalAddr local address
func (c *channel) LocalAddr() string {
	return c.conn.LocalAddr().String()
}

// RemoteAddr remote address
func (c *channel) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

// Write write message to pipeline
func (c *channel) Write(message any) error {
	if !c.IsActive() {
		return errors.New("channel is closed")
	}
	return c.invokeMethod(func() {
		c.pipeline.FireChannelWrite(message)
	})
}

// Write1 write message to conn
func (c *channel) Write1(message []byte) (n int, err error) {
	if c.EventLoop().InEventLoop() {
		n, err = c.conn.Write(message) // 若在事件循环中,直接写入conn
	} else {
		// 若不在事件循环中, 在eventloop中异步写入conn,并等待写入完成
		writesig := make(chan error)
		n = len(message)
		err3 := c.conn.AsyncWrite(message, func(c agonet.Conn, err2 error) error {
			err = err2
			close(writesig)
			return nil
		})
		if nil != err3 {
			err = err3
			close(writesig)
			return
		}
		<-writesig // FIXME 此处会阻塞写入完成, 请在非事件循环中调用
	}
	return
}

// Close through the Pipeline
func (c *channel) Close(err error) {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		err2 := c.eventloop.Execute(context.Background(), agonet.RunnableFunc(func(_ context.Context) error {
			return c.EventLoop().Close(c.conn, err)
		}))

		if err2 != nil {
			// TODO 异常处理设计
			// c.pipeline.FireChannelException(AsException(err))
		}
	}
}

func (c *channel) invokeMethod(fn func()) (err error) {
	defer func() {
		if err := recover(); nil != err && 0 == atomic.LoadInt32(&c.closed) {
			c.pipeline.FireChannelException(AsException(err))
		}
	}()

	fn()
	return nil
}
