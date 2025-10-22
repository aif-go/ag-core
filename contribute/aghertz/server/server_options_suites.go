package server

import (
	"fmt"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/http2/factory"
	"github.com/hertz-contrib/pprof"
)

// WithServerSuite 将 ServerSuite 中的选项应用到 Hertz 服务器配置中。
func WithServerSuite(suite ServerSuite) ServerOption {
	return ServerOption{
		F: func(h *server.Hertz) error {
			for _, op := range suite.Options() {
				err := op.F(h)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// WithPprof 配置pprof
func WithPprof(pprofUrl string) ServerOption {
	return ServerOption{
		F: func(h *server.Hertz) error {
			if pprofUrl == "" {
				return fmt.Errorf("pprofUrl is empty")
			}
			slog.Info("enable pprof", "url", pprofUrl)
			pprof.Register(h, pprofUrl)
			return nil
		},
	}
}

// WithH2C 配置H2C
func WithH2C() ServerOption {
	return ServerOption{
		F: func(h *server.Hertz) error {
			h.AddProtocol("h2", factory.NewServerFactory())
			return nil
		},
	}
}
