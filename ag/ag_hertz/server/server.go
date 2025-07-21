package server

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_ext/ip"
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"log/slog"
	"net"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/hertz-contrib/registry/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
)

type Route struct {
	HttpMethod, RelativePath string
	Handlers                 []app.HandlerFunc
}

type Server struct {
	*server.Hertz
	root   *route.RouterGroup
	logger *slog.Logger
}

type Option func(*Server)

func WithRoute(r *Route) Option {
	return func(s *Server) {
		s.root.Handle(r.HttpMethod, r.RelativePath, r.Handlers...)
	}
}
func NewServer(hertz *server.Hertz, logger *slog.Logger, opts ...Option) *Server {
	s := &Server{
		Hertz:  hertz,
		logger: logger,
		root:   hertz.Group("/", rootMw()...),
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func rootMw() []app.HandlerFunc {
	return nil
}

// NewHertzServerWithSuit 创建一个Hertz服务实例，使用配置套件，并且注册服务
// suite: 配置套件
// logger: 日志记录器
// 返回一个Server实例
func NewHertzServerWithSuit(
	suite *HertzOptionSuite,
	logger *slog.Logger,
) *Server {
	return NewServer(
		server.Default(
			suite.Options()...),
		logger,
		suite.Routers()...,
	)
}

func (s *Server) Start(context.Context) error {
	s.logger.Info("hertz server start")

	s.Hertz.Spin()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Hertz.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown hertz server", "error", err)
	}

	s.logger.Info("Shutting down hertz server...")
	return nil
}

// HertzOptionSuite 定义了Hertz服务的配置套件
type HertzOptionSuite struct {
	opts    []config.Option
	routers []Option
}

// Options 返回配置项
func (s *HertzOptionSuite) Options() []config.Option {
	return s.opts
}
func (s *HertzOptionSuite) Routers() []Option {
	return s.routers
}

// HertzSuiteBuilder 定义了Hertz服务配置套件的构建器
type HertzSuiteBuilder struct {
	Env           ag_conf.IConfigurableEnvironment
	Binder        ag_conf.IBinder
	CustOptions   []config.Option
	RouterOptions []Option
	NamingClient  naming_client.INamingClient
}

// BuildSuite 构建Hertz服务配置套件
func (builder *HertzSuiteBuilder) BuildSuite() (*HertzOptionSuite, error) {
	suite := &HertzOptionSuite{
		opts:    builder.CustOptions,
		routers: builder.RouterOptions,
	}

	var hconf HertzServerProperties
	err := builder.Binder.Bind(&hconf, hertzServerPropertiesPrefix)
	if err != nil {
		slog.Error("hertz server config error", "error", err)
		return nil, err
	}

	// 服务信息配置
	host, port, err := findHertzHostPort(hconf)
	if err != nil {
		return nil, err
	}

	hertzHostStr := fmt.Sprintf("%s:%d", host, port)
	slog.Info("hertz", "host", hertzHostStr)
	suite.opts = append(suite.opts, server.WithHostPorts(hertzHostStr))

	// 注册中心配置
	var nacosRegistry registry.Registry
	if builder.NamingClient != nil {
		slog.Info("hertz server enable nacos naming")
		nacosRegistry = nacos.NewNacosRegistry(
			builder.NamingClient,
			nacos.WithRegistryCluster(hconf.Cluster),
			nacos.WithRegistryGroup(hconf.Group),
		)

		regInfo := &registry.Info{}
		regInfo.Weight = 1
		ipranger, err := ip.NewIPRanger(hconf.EnableIPRange)
		if err != nil {
			return nil, err
		}

		host, ok, err := ipranger.GetLocalIP()
		if err != nil {
			return nil, err
		}
		if ok {
			slog.Info("hertz server enable ip range", "regAddr", fmt.Sprintf("%s:%d", host, port))
			regInfo.Addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
			if err != nil {
				return nil, err
			}
		}

		sname := hconf.ServiceName
		if sname == "" {
			sname = "hertz-server"
		}
		regInfo.ServiceName = sname

		// 服务元信息配置，可在配置中配置，兼容并行阶段的spring-grpc网关调用
		tags := make(map[string]string)
		if hconf.Tags != nil {
			tags = hconf.Tags
		}
		tags["ag_core"] = "All rights reserved"
		tags["lang_type"] = "Golang"
		regInfo.Tags = tags

		suite.opts = append(
			suite.opts,
			server.WithRegistry(
				nacosRegistry,
				regInfo,
			),
		)
	}
	return suite, nil
}

func findHertzHostPort(hconf HertzServerProperties) (host string, port int, rerr error) {
	// 服务ip、端口配置
	host = hconf.Host
	if host == "" {
		host = "0.0.0.0"
	}

	if !ip.IsHostAvailable(host) {
		return "", 0, fmt.Errorf("hertz host unavailable: %s", host)
	}

	port = hconf.Port
	if hconf.AdaptivePort {
		slog.Info("hertz server enable adaptive port")
		if port == 0 {
			port = DefaultHertzOriginPort
		}
		port, rerr = ip.GetAvailablePort(host, port)
		if rerr != nil {
			return
		}
	} else {
		if port == 0 {
			return host, port, fmt.Errorf("hertz port invalid:%v", port)
		}
	}

	slog.Info(fmt.Sprintf("found hertz host:%s, port:%d", host, port))
	return
}
