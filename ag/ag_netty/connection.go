package ag_netty

import (
	"github.com/cloudwego/netpoll"
	"github.com/panjf2000/gnet"
	"io"
	"net"
	"sync"
	"time"
)

type Connection interface {
	// Read reads data from the connection
	Read() ([]byte, error)

	// Write writes data to the connection
	Write(data []byte) (int, error)

	// Close closes the connection
	Close() error

	// RemoteAddr returns the remote network address
	RemoteAddr() net.Addr

	// LocalAddr returns the local network address
	LocalAddr() net.Addr

	// SetReadTimeout sets the read timeout
	SetReadTimeout(timeout time.Duration)

	// SetWriteTimeout sets the write timeout
	SetWriteTimeout(timeout time.Duration)

	// SetIdleTimeout sets the idle timeout
	SetIdleTimeout(timeout time.Duration)
}

// NetpollConnAdapter  netpoll.Connection 适配器
type NetpollConnAdapter struct {
	conn netpoll.Connection
}

func NewNetpollConnAdapter(conn netpoll.Connection) Connection {
	return &NetpollConnAdapter{conn: conn}
}

func (a *NetpollConnAdapter) Read() ([]byte, error) {
	reader := a.conn.Reader()
	n := reader.Len()
	if n == 0 {
		return nil, io.EOF
	}
	return reader.ReadBinary(n)
}

func (a *NetpollConnAdapter) Write(data []byte) (int, error) {
	return a.conn.Write(data)
}

func (a *NetpollConnAdapter) Close() error {
	return a.conn.Close()
}

func (a *NetpollConnAdapter) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *NetpollConnAdapter) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *NetpollConnAdapter) SetReadTimeout(timeout time.Duration) {
	a.conn.SetReadTimeout(timeout)
}

func (a *NetpollConnAdapter) SetWriteTimeout(timeout time.Duration) {
	a.conn.SetWriteTimeout(timeout)
}

func (a *NetpollConnAdapter) SetIdleTimeout(timeout time.Duration) {
	a.conn.SetIdleTimeout(timeout)
}

// GnetConnAdapter  gnet.Conn 适配器
type GnetConnAdapter struct {
	conn gnet.Conn
	mu   sync.Mutex
}

func NewGnetConnAdapter(conn gnet.Conn) Connection {
	return &GnetConnAdapter{conn: conn}
}

func (a *GnetConnAdapter) Read() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	data := a.conn.Read()
	if len(data) == 0 {
		return nil, io.EOF
	}

	result := make([]byte, len(data))
	copy(result, data)

	a.conn.ResetBuffer()

	return result, nil
}

func (a *GnetConnAdapter) Write(data []byte) (int, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	err := a.conn.AsyncWrite(data)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

func (a *GnetConnAdapter) Close() error {
	return a.conn.Close()
}

func (a *GnetConnAdapter) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *GnetConnAdapter) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *GnetConnAdapter) SetReadTimeout(timeout time.Duration) {
	// gnet doesn't have a direct equivalent
}

func (a *GnetConnAdapter) SetWriteTimeout(timeout time.Duration) {
	// gnet doesn't have a direct equivalent
}

func (a *GnetConnAdapter) SetIdleTimeout(timeout time.Duration) {
	// gnet doesn't have a direct equivalent
}

// NetConnAdapter  net.Conn 适配器
type NetConnAdapter struct {
	conn     net.Conn
	buffer   []byte
	mu       sync.Mutex
	readDone bool
}

func NewNetConnAdapter(conn net.Conn) Connection {
	return &NetConnAdapter{
		conn:   conn,
		buffer: make([]byte, 0),
	}
}

func (a *NetConnAdapter) Read() ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.buffer) > 0 {
		data := make([]byte, len(a.buffer))
		copy(data, a.buffer)
		a.buffer = a.buffer[:0]
		return data, nil
	}

	if a.readDone {
		return nil, io.EOF
	}

	// Read from the connection
	buf := make([]byte, 4096)
	n, err := a.conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			a.readDone = true
		}
		return nil, err
	}

	if n == 0 {
		return nil, io.EOF
	}

	data := make([]byte, n)
	copy(data, buf[:n])
	return data, nil
}

func (a *NetConnAdapter) Write(data []byte) (int, error) {
	return a.conn.Write(data)
}

func (a *NetConnAdapter) Close() error {
	return a.conn.Close()
}

func (a *NetConnAdapter) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *NetConnAdapter) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *NetConnAdapter) SetReadTimeout(timeout time.Duration) {
	a.conn.SetReadDeadline(time.Now().Add(timeout))
}

func (a *NetConnAdapter) SetWriteTimeout(timeout time.Duration) {
	a.conn.SetWriteDeadline(time.Now().Add(timeout))
}

func (a *NetConnAdapter) SetIdleTimeout(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	a.conn.SetDeadline(deadline)
}
