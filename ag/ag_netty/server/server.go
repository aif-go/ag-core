package server

import (
	"ag-core/ag/ag_ext/ip"
	"ag-core/ag/ag_netty"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Server struct {
	*ag_netty.Server
	addr     string
	handlers []ag_netty.ChannelHandler
	logger   *slog.Logger
}

type Option struct {
	opt func(*Server)
}

func WithAddr(addr string) Option {
	return Option{
		opt: func(s *Server) {
			s.addr = addr
		},
	}
}
func AppendHandler(ch ag_netty.ChannelHandler) Option {
	return Option{
		opt: func(s *Server) {
			s.handlers = append(s.handlers, ch)
		},
	}
}

func NewServer(logger *slog.Logger, opts ...Option) *Server {
	s := &Server{
		handlers: make([]ag_netty.ChannelHandler, 0),
		logger:   logger,
	}

	for _, opt := range opts {
		opt.opt(s)
	}

	initFunc := func(ch *ag_netty.Channel) {
		pipeline := ch.Pipeline
		if pipeline != nil {
			for i, handler := range s.handlers {
				pipeline.AddLast(fmt.Sprintf("handler%d", i), handler)
			}
		}
	}

	server, err := ag_netty.NewServer(s.addr, initFunc)
	if err != nil {
		panic(err)
	}
	s.Server = server
	return s
}

func NewNettyServerWithSuite(
	suite *NettyOptionSuite,
	logger *slog.Logger,
) *Server {
	return NewServer(logger, suite.Options()...)
}

type NettyOptionSuite struct {
	Opts []Option
}

func (s *NettyOptionSuite) Options() []Option { return s.Opts }

type NettySuiteBuilder struct {
	NSP           *NettyServerProperties
	CustomOptions []Option
}

func (builder *NettySuiteBuilder) BuildSuite() (*NettyOptionSuite, error) {
	suite := &NettyOptionSuite{
		Opts: make([]Option, 0),
	}

	suite.Opts = append(suite.Opts, builder.CustomOptions...)

	host, port, err := findHostPort(builder.NSP)
	if err != nil {
		panic(err)
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	slog.Info("ag_netty", "host", addr)
	suite.Opts = append(suite.Opts, WithAddr(addr))

	return suite, nil
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("ag_netty server start")
	s.Server.Start()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("ag_netty server shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.Server.Shutdown()

	s.logger.Info("Shutting down ag_netty server...")
	return nil
}

func findHostPort(conf *NettyServerProperties) (host string, port int, rerr error) {
	// 服务ip、端口配置
	host = conf.Host
	if host == "" {
		host = "0.0.0.0"
	}

	if !ip.IsHostAvailable(host) {
		return "", 0, fmt.Errorf("ag_netty host unavailable: %s", host)
	}

	port = conf.Port
	if conf.AdaptivePort {
		slog.Info("ag_netty server enable adaptive port")
		if port == 0 {
			port = DefaultNettyOriginPort
		}
		port, rerr = ip.GetAvailablePort(host, port)
		if rerr != nil {
			return
		}
	} else {
		if port == 0 {
			return host, port, fmt.Errorf("ag_netty port invalid:%v", port)
		}
	}

	slog.Info(fmt.Sprintf("found ag_netty host:%s, port:%d", host, port))
	return
}
