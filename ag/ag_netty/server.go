package ag_netty

import (
	"log/slog"
	"net"
)

// Server 服务器
type Server struct {
	loop          *EventLoop
	numEventLoops int
	listener      net.Listener
	shutdown      chan struct{}
}

// NewServer 创建新服务器
func NewServer(addr string, initFunc func(ch *Channel)) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	loop, err := NewEventLoop(initFunc)
	if err != nil {
		return nil, err
	}

	return &Server{
		loop:          loop,
		numEventLoops: 4,
		listener:      listener,
		shutdown:      make(chan struct{}),
	}, nil
}

// Start 启动服务器
func (s *Server) Start() {
	go func() {
		if err := s.loop.Run(s.listener, s.numEventLoops); err != nil {
			slog.Error("EventLoop exited: ", "error", err)
		}
	}()

	slog.Info("Server started!", "port", s.listener.Addr(), "eventLoops", s.numEventLoops)

	// 等待关闭信号
	select {
	case <-s.shutdown:
		slog.Info("Server shutdown")
	}
}

// Shutdown 关闭服务器
func (s *Server) Shutdown() {
	close(s.shutdown)
	s.loop.Shutdown()
	s.listener.Close()
}
