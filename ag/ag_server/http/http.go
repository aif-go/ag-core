package http

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type Server struct {
	*gin.Engine
	httpSrv *http.Server
	host    string
	port    int
	//logger  *log.Logger
	logger *slog.Logger
}
type Option func(s *Server)

// func NewServer(engine *gin.Engine, logger *log.Logger, opts ...Option) *Server {
func NewServer(engine *gin.Engine, logger *slog.Logger, opts ...Option) *Server {
	s := &Server{
		Engine: engine,
		logger: logger,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithServerHost(host string) Option {
	return func(s *Server) {
		s.host = host
	}
}

func WithServerPort(port int) Option {
	return func(s *Server) {
		s.port = port
	}
}

func NewHttpGinServer(
	logger *slog.Logger,
	conf ag_conf.IConfigurableEnvironment,
) *Server {
	/*====================*/
	host := conf.GetProperty("http.host")
	port, err := cast.ToIntE(conf.GetProperty("http.port"))
	if err != nil {
		panic(err)
	}
	gin.SetMode(gin.DebugMode)
	gins := gin.Default()
	s := NewServer(
		gins,
		logger,
		WithServerHost(host),
		WithServerPort(port),
	)

	// TODO 启用pprof

	return s
	/*====================*/
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("gin server start", "host", fmt.Sprintf("http://%s:%d", s.host, s.port))
	//	app.Logger.Info("docs addr", "addr", fmt.Sprintf("http://%s:%d/swagger/index.html", conf.GetString("http.host"), conf.GetInt("http.port")))
	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: s,
	}

	if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		//s.logger.Sugar().Fatalf("listen: %s\n", err)
		s.logger.Error("listen: %s\n", err) // TODO 能达到Fatal的效果吗
		log.Fatalf("listen: %s\n", err)
	}

	return nil
}
func (s *Server) Stop(ctx context.Context) error {
	//s.logger.Sugar().Info("Shutting down server...")
	s.logger.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		//s.logger.Sugar().Fatal("Server forced to shutdown: ", err)
		s.logger.Error("Server forced to shutdown: ", err)
	}

	//s.logger.Sugar().Info("Server exiting")
	s.logger.Info("Server exiting")
	return nil
}
