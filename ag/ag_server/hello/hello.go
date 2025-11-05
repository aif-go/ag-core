package hello

import (
	// "ag-core/ag/ag_netty/client"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	httpSrv *http.Server
	// suite   *client.NettyOptionSuite
	logger *slog.Logger
}
type Option func(s *Server)

// func NewHelloServer(engine *gin.Engine, logger *log.Logger, opts ...Option) *Server {
func NewHelloServer(
	// suite *client.NettyOptionSuite,
	logger *slog.Logger,
) *Server {
	s := &Server{
		// suite:  suite,
		logger: logger,
	}
	return s
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", "0.0.0.0", 8888)
	slog.Info("hello server start", "addr", addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 打印r中所有的head
		for k, v := range r.Header {
			fmt.Printf("%s:%s\n", k, v)
		}

		body := r.Body
		// 读取body内容
		buf := make([]byte, 2048)
		n, _ := body.Read(buf)

		// clientWithSuite := client.NewNettyClientWithSuite(s.suite, s.logger)
		// err2 := clientWithSuite.Connect()
		// if err2 != nil {
		// 	panic(err2)
		// }
		// _, err2 = clientWithSuite.SendAndGet(buf[:n])
		// if err2 != nil {
		// 	println("err2", err2.(error).Error())
		// 	panic(err2)
		// }

		bbuf := buf[:n]
		var bmap map[string]any
		err := json.Unmarshal(bbuf, &bmap)
		if err != nil {
			slog.Error("unmarshal", "err", err)
		} else {
			bijson, _ := json.MarshalIndent(bmap, " ", " ")
			slog.Info(fmt.Sprintf("%s\n", bijson))
		}

		slog.Info("hello world", "addr", addr, "time", time.Now().Format("2006-01-02 15:04:05.000"))
		fmt.Fprintf(w, "Hello, World!")
	})

	s.httpSrv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// log.Fatalf("listen: %s\n", err)
		// slog.Error("hello server", "err", err)
		slog.Info("hello server", "err", err)
		return err
	}
	return nil
}
func (s *Server) Stop(ctx context.Context) error {
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown: ", err)
	}

	slog.Info("Server exiting")
	return nil
}
