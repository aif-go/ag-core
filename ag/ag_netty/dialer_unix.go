//go:build !windows
// +build !windows

package ag_netty

import (
	"github.com/panjf2000/gnet"
	"log/slog"
	"time"
)

// Dial 连接到服务器
func Dial(
	addr string,
	connTimeout time.Duration,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	idleTimeout time.Duration,
	looper EventLooper,
) (*Channel, error) {
	eventHandler := &clientEventHandler{
		channel: nil,
		looper:  looper,
	}

	client, err := gnet.NewClient(eventHandler, gnet.WithTCPKeepAlive(idleTimeout))
	if err != nil {
		return nil, err
	}

	if err := client.Start(); err != nil {
		return nil, err
	}

	var conn gnet.Conn
	var dialErr error

	connChan := make(chan struct{})

	go func() {
		conn, dialErr = client.Dial("tcp", addr)
		close(connChan)
	}()

	select {
	case <-connChan:
		if dialErr != nil {
			client.Stop()
			return nil, dialErr
		}
	case <-time.After(connTimeout):
		client.Stop()
		return nil, &timeoutError{op: "dial", addr: addr, timeout: connTimeout}
	}

	connAdapter := NewGnetConnAdapter(conn)

	channel := NewChannel(connAdapter, looper)

	eventHandler.channel = channel

	if clientLooper, ok := looper.(*ClientEventLoop); ok {
		if clientLooper.initFunc != nil {
			clientLooper.initFunc(channel)
		}
	}

	looper.Post(func() {
		channel.Pipeline.FireActive()
	})

	slog.Info("Connected to server", "addr", addr)
	return channel, nil
}

type clientEventHandler struct {
	gnet.EventServer
	channel *Channel
	looper  EventLooper
}

func (h *clientEventHandler) React(packet []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	if h.channel != nil && len(packet) > 0 {
		data := make([]byte, len(packet))
		copy(data, packet)

		h.looper.Post(func() {
			h.channel.Pipeline.FireRead(data)
		})
	}
	return nil, gnet.None
}

func (h *clientEventHandler) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	if h.channel != nil {
		h.looper.Post(func() {
			h.channel.Pipeline.FireError(err)
			h.channel.Close()
		})
	}
	return gnet.None
}

type timeoutError struct {
	op      string
	addr    string
	timeout time.Duration
}

func (e *timeoutError) Error() string {
	return "timeout: " + e.op + " " + e.addr + " " + e.timeout.String()
}

func (e *timeoutError) Timeout() bool {
	return true
}

func (e *timeoutError) Temporary() bool {
	return true
}
