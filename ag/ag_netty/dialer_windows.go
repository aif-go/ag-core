//go:build windows
// +build windows

package ag_netty

import (
	"log/slog"
	"net"
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
	netConn, err := net.DialTimeout("tcp", addr, connTimeout)
	if err != nil {
		return nil, err
	}

	if readTimeout > 0 {
		netConn.SetReadDeadline(time.Now().Add(readTimeout))
	}
	if writeTimeout > 0 {
		netConn.SetWriteDeadline(time.Now().Add(writeTimeout))
	}

	conn := NewNetConnAdapter(netConn)

	channel := NewChannel(conn, looper)

	if clientLooper, ok := looper.(*ClientEventLoop); ok {
		if clientLooper.initFunc != nil {
			clientLooper.initFunc(channel)
		}
	}

	looper.Post(func() {
		channel.Pipeline.FireActive()
	})

	go readLoopForNetConn(netConn, channel, looper, readTimeout, idleTimeout)

	slog.Info("Connected to server", "addr", addr)
	return channel, nil
}

func readLoopForNetConn(
	conn net.Conn,
	channel *Channel,
	looper EventLooper,
	readTimeout time.Duration,
	idleTimeout time.Duration,
) {
	buffer := make([]byte, 4096)
	for {
		if looper.IsShutdown() {
			return
		}

		if readTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(readTimeout))
		}

		n, err := conn.Read(buffer)
		if err != nil {
			looper.Post(func() {
				channel.Pipeline.FireError(err)
			})
			channel.Close()
			return
		}

		if n > 0 {
			data := make([]byte, n)
			copy(data, buffer[:n])

			looper.Post(func() {
				channel.Pipeline.FireRead(data)
			})
		} else {
			time.Sleep(10 * time.Millisecond)
		}
	}
}
