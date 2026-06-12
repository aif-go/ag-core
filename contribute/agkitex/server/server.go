package server

import (
	"github.com/aif-go/ag-core/ag/ag_ext/ip"
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/registry"
	kregistry "github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/grpc"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
)

type (

	// KitexServerSuite 服务器配置套件
	KitexServerSuite struct {
		opts []server.Option
	}

	// KitexServerSuiteBuilder 服务器套件构建器
	KitexServerSuiteBuilder struct {
		// 自定义选项
		ServerOptions []server.Option

		// 配置
		Properties *KitexServerProperties

		// 服务注册器
		Registry registry.Registry

		// 中间件(原始)
		Middlewares []endpoint.Middleware

		// 中间件(带优先级)
		PrioritizedMiddlewares []PrioritizedServerMiddleware
	}

	// AgKitexServer agkitex服务器，包装kitex服务器，提供ag_app服务启动入口
	AgKitexServer struct {
		KitexServer server.Server
	}
)

// NewAgKitexServer 包装kitex服务器为agkitex服务器，提供ag_app服务启动入口
func NewAgKitexServer(kitexServer server.Server) *AgKitexServer {
	// func NewAgKitexServer(kitexServer server.Server) ag_server.Server {

	return &AgKitexServer{
		KitexServer: kitexServer,
	}
}

func (s *AgKitexServer) Start(ctx context.Context) error {
	sinfos := s.KitexServer.GetServiceInfos()
	// 格式化打印服务信息
	for sname, sinfo := range sinfos {
		slog.Info("Kitex Server Service Info",
			"service_name", sname,
			"method_count", len(sinfo.Methods),
			"method_list", sinfo.Methods,
			"extra_info", sinfo.Extra,
		)
	}

	slog.Info("Kitex Server Start",
		"service_count", len(sinfos),
	)

	return s.KitexServer.Run()
}

func (s *AgKitexServer) Stop(ctx context.Context) error {
	slog.Info("Kitex Server Stoping...")
	err := s.KitexServer.Stop()
	if err != nil {
		slog.Error("Kitex Server Stop Error", "error", err)
		return err
	}
	// 打印停止日志
	slog.Info("Kitex Server Stop Success")
	return nil
}

// NewKitexServerWithSuite 创建kitex服务器，使用指定的服务器套件
func NewKitexServerWithSuite(suite *KitexServerSuite) (server.Server, error) {
	svr := server.NewServer(server.WithSuite(suite))
	return svr, nil
}

// Options 返回服务器选项
func (s *KitexServerSuite) Options() []server.Option {
	return s.opts
}

// BuildSuite 构建服务器套件
func (builder *KitexServerSuiteBuilder) BuildServerSuite() (*KitexServerSuite, error) {
	opts := make([]server.Option, 0)

	// 1. 添加自定义选项
	opts = append(opts, builder.ServerOptions...)

	// 2. 从配置构建的选项
	configOpts, err := builder.optionsFromConfig()
	if err != nil {
		return nil, err
	}
	opts = append(opts, configOpts...)

	// 4. 配置服务注册器
	mvOpts, err := builder.middlewareOptions()
	if err != nil {
		return nil, err
	}
	opts = append(opts, mvOpts...)

	// 5. 配置服务注册器
	if builder.Registry != nil {
		opts = append(opts, server.WithRegistry(builder.Registry))
	}

	suite := &KitexServerSuite{
		opts: opts,
	}
	return suite, nil
}

// 根据配置生成服务器选项
func (builder *KitexServerSuiteBuilder) optionsFromConfig() ([]server.Option, error) {
	opts := make([]server.Option, 0)

	kconf := builder.Properties
	slog.Info(fmt.Sprintf("service_name: %s", kconf.ServiceName))

	// = kitex 服务地址配置 =
	host, port, err := findKitexHostPort(kconf)
	if err != nil {
		return nil, err
	}
	kitexHostStr := fmt.Sprintf("%s:%d", host, port)
	slog.Info("kitex", "host", kitexHostStr)
	addr, err := net.ResolveTCPAddr("tcp", kitexHostStr)
	if err != nil {
		return nil, fmt.Errorf("kitex host error: %w", err)
	}
	opts = append(opts, server.WithServiceAddr(addr))

	// = kitex服务信息配置 =
	sname := kconf.ServiceName
	if sname == "" {
		sname = "kitex-server"
	}
	info := &rpcinfo.EndpointBasicInfo{
		ServiceName: sname,
	}
	opts = append(opts, server.WithServerBasicInfo(info))

	// = kitex服务注册信息 =
	regInfo := &kregistry.Info{
		Tags: map[string]string{},
	}
	regInfo.Weight = 1                                  // FIXME grpc-spring 项目默认注册权重为1，kitex默认为10
	regInfo.Tags["gRPC_port"] = fmt.Sprintf("%d", port) // FIXME 兼容https://github.com/grpc-ecosystem/grpc-spring项目的服务发现实现
	if kconf.EnableIPRange != "" {
		ipranger, err := ip.NewIPRanger(kconf.EnableIPRange)
		if err != nil {
			return nil, err
		}
		host, ok, err := ipranger.GetLocalIP()
		if err != nil {
			return nil, err
		}
		if ok {
			slog.Info("kitex server enable ip range", "regAddr", fmt.Sprintf("%s:%d", host, port))
			regInfo.SkipListenAddr = true
			regInfo.Addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
			if err != nil {
				return nil, err
			}
		}
	}
	opts = append(opts, server.WithRegistryInfo(regInfo))

	// = grpc相关配置 =
	if kconf.Grpc.Enable {
		gskeep := grpc.ServerKeepalive{}
		if kconf.Grpc.MaxConnectionIdle > 0 {
			// 最大空闲连接时间
			gskeep.MaxConnectionIdle = time.Second * time.Duration(kconf.Grpc.MaxConnectionIdle)
		}
		opts = append(opts, server.WithGRPCKeepaliveParams(gskeep))
		opts = append(opts, server.WithMetaHandler(transmeta.ServerHTTP2Handler))
	}

	return opts, nil
}

// middlewareOptions 构建服务器中间件选项
func (builder *KitexServerSuiteBuilder) middlewareOptions() ([]server.Option, error) {
	opts := make([]server.Option, 0)

	// 复制中间件
	pmws := append([]PrioritizedServerMiddleware{}, builder.PrioritizedMiddlewares...)

	// 添加普通中间件，默认优先级为 ServerMiddlewarePriorityNormal
	for _, mv := range builder.Middlewares {
		pmws = append(
			pmws,
			NewSimplePrioritizedServerMiddleware(ServerMiddlewarePriorityNormal, mv),
		)
	}

	// 构建中间件选项
	if len(pmws) > 0 {
		middlewareOpts := BuildServerMiddlewareOptions(pmws)
		opts = append(opts, middlewareOpts...)
	}
	return opts, nil
}

// parseAddr 解析地址
func parseAddr(addr string) (*net.TCPAddr, error) {
	return net.ResolveTCPAddr("tcp", addr)
}
