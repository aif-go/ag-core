package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app/server"
)

// AgHertzServer is the server of Hertz.
type AgHertzServer struct {
	hertz *server.Hertz
}

// NewAGHertzServer creates a new AgHertzServer.
func NewAGHertzServer(hertz *server.Hertz) *AgHertzServer {
	s := &AgHertzServer{
		hertz: hertz,
	}
	return s
}

// Start starts the AgHertzServer implement ag_server.Server.
func (s *AgHertzServer) Start(context.Context) error {
	if s.hertz == nil {
		return fmt.Errorf("hertz server is nil")
	}

	slog.Info("start hertz server")
	s.hertz.Spin()

	return nil
}

// Stop stops the AgHertzServer implement ag_server.Server.
func (s *AgHertzServer) Stop(ctx context.Context) error {
	slog.Info("stop hertz server")
	return s.hertz.Shutdown(ctx)
}
