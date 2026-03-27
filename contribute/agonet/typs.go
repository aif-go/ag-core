package agonet

import (
	"context"
	"io"
	"net"
	"time"
)

// Conn is an interface of underlying connection.
type Conn interface {
	Reader // all methods in Reader are not concurrency-safe.
	Writer // some methods in Writer are concurrency-safe, some are not.
	// Socket // all methods in Socket are concurrency-safe.

	// Context returns a user-defined context, it's not concurrency-safe,
	// you must invoke it within any method in EventHandler.
	Context() (ctx any)

	// EventLoop returns the event-loop that the connection belongs to.
	// The returned EventLoop is concurrency-safe.
	EventLoop() EventLoop

	// SetContext sets a user-defined context, it's not concurrency-safe,
	// you must invoke it within any method in EventHandler.
	SetContext(ctx any)

	// Close implements net.Conn.
	Close() error

	// LocalAddr implements net.Conn.
	LocalAddr() net.Addr

	// RemoteAddr  implements net.Conn.
	RemoteAddr() net.Addr

	// SetDeadline implements net.Conn.
	SetDeadline(time.Time) error

	// SetReadDeadline implements net.Conn.
	SetReadDeadline(time.Time) error

	// SetWriteDeadline implements net.Conn.
	SetWriteDeadline(time.Time) error

	// NetConn returns the underlying net.Conn like net.TCPConn.
	// 注意直接读写该连接对象将会导致无法预估后果
	NetConn() net.Conn
}

// EventLoop provides a set of methods for manipulating the event-loop.
type EventLoop interface {
	// Register connects to the given address and registers the connection to the current event-loop,
	// it's concurrency-safe.
	Register(ctx context.Context, addr net.Addr) (<-chan RegisteredResult, error)
	// Enroll is like Register, but it accepts an established net.Conn instead of a net.Addr,
	// it's concurrency-safe.
	Enroll(ctx context.Context, c net.Conn) (<-chan RegisteredResult, error)

	// Close closes the given Conn that belongs to the current event-loop.
	// It must be called on the same event-loop that the connection belongs to.
	// This method is not concurrency-safe, you must invoke it on the event loop.
	Close(Conn, error) error

	// Deprecated
	InEventLoop() bool

	Execute(ctx context.Context, runnable Runnable) error
}

// Runnable defines the common protocol of an execution on an event-loop.
// This interface should be implemented and passed to an event-loop in some way,
// then the event-loop will invoke Run to perform the execution.
// !!!Caution: Run must not contain any blocking operations like heavy disk or
// network I/O, or else it will block the event-loop.
type Runnable interface {
	// Run is about to be executed by the event-loop.
	Run(ctx context.Context) error
}

// RunnableFunc is an adapter to allow the use of ordinary function as a Runnable.
type RunnableFunc func(ctx context.Context) error

// Run executes the RunnableFunc itself.
func (fn RunnableFunc) Run(ctx context.Context) error {
	return fn(ctx)
}

// RegisteredResult is the result of a Register call.
type RegisteredResult struct {
	Conn Conn
	Err  error
}

// Reader is an interface that consists of a number of methods for reading that Conn must implement.
type Reader interface {
	io.Reader
	io.WriterTo

	// Next returns the next n bytes and advances the inbound buffer.
	Next(n int) (buf []byte, err error)

	// Peek returns the next n bytes without advancing the inbound buffer,
	Peek(n int) (buf []byte, err error)

	// Discard advances the inbound buffer with next n bytes, returning the number of bytes discarded.
	Discard(n int) (discarded int, err error)

	// ReadableBytes returns the number of bytes that are available to read.
	ReadableBytes() int
}

// Writer is an interface that consists of a number of methods for writing that Conn must implement.
type Writer interface {
	io.Writer     // not concurrency-safe
	io.ReaderFrom // not concurrency-safe

	Flush() error // not concurrency-safe

	AsyncWrite(buf []byte, callback AsyncCallback) (err error)
}

// AsyncCallback is a callback that will be invoked after the asynchronous function finishes.
type AsyncCallback func(c Conn, err error) error

type KeepAliveAbility interface {
	// SetKeepAliveConfig sets the keep-alive config.
	SetKeepAliveConfig(config net.KeepAliveConfig) error
}
