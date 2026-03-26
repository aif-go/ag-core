package simple

import (
	"ag-core/contribute/agonet"
	"errors"
	"sync/atomic"
	"time"
)

var _ Channel = (*channel)(nil)

type channel struct {
	id       int64
	conn     agonet.Conn
	executor agonet.EventLoop
	pipeline Pipeline

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
		id:       id,
		conn:     conn,
		executor: executor,
		pipeline: pipeline,
	}

	return channel
}

func (c *channel) ID() int64 {
	return c.id
}

func (c *channel) Write(message Message) error {
	if !c.IsActive() {
		return errors.New("channel is closed")
	}
	return c.invokeMethod(func() {
		c.pipeline.FireChannelWrite(message)
	})
}

// LocalAddr local address
func (c *channel) LocalAddr() string {
	return c.conn.LocalAddr().String()
}

// RemoteAddr remote address
func (c *channel) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *channel) Write1(message []byte) (n int, err error) {
	if c.EventLoop().InEventLoop() {
		n, err = c.conn.Write(message)
	} else {
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
		<-writesig // TODO 此处若在事件循环中,会阻塞事件循环
	}
	return
}

func (c *channel) EventLoop() agonet.EventLoop {
	return c.executor
}

func (c *channel) Pipeline() Pipeline {
	return c.pipeline
}

// Close through the Pipeline
func (c *channel) Close(err error) {
	// c.EventLoop().Register(context.Background(), c.conn.LocalAddr())
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		// TODO 处理关闭逻辑

	}
}

func (c *channel) IsActive() bool {
	return atomic.LoadInt32(&c.closed) == 0
}

func (c *channel) invokeMethod(fn func()) (err error) {

	defer func() {
		if err := recover(); nil != err && 0 == atomic.LoadInt32(&c.closed) {
			c.pipeline.FireChannelException(AsException(err))

			// if e, ok := err.(error); ok {
			// 	var ne net.Error
			// 	if errors.As(e, &ne) && !ne.Timeout() {
			// 		c.Close(e)
			// 	}
			// }
		}
	}()

	fn()
	return nil
}
