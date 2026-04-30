package client

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/loadbalance"
	"github.com/cloudwego/kitex/transport"
)

// KitexClientSuite 客户端配置套件
type KitexClientSuite struct {
	opts []client.Option
}

// Options 返回客户端选项
func (s *KitexClientSuite) Options() []client.Option {
	return s.opts
}

// KitexSuiteBuilder 客户端套件构建器
type KitexSuiteBuilder struct {
	// 自定义选项
	ClientOptions []client.Option

	// 配置
	Config *KitexClientConfig

	// 服务发现解析器
	Resolver discovery.Resolver

	// 负载均衡器
	LoadBalancer loadbalance.Loadbalancer

	// 中间件
	Middleware []endpoint.Middleware

	// 中间件
	PrioritizedMiddlewares []PrioritizedClientMiddleware
}

// BuildSuite 构建客户端套件
func (builder *KitexSuiteBuilder) BuildSuite() (*KitexClientSuite, error) {
	config := builder.Config
	// if config == nil {
	// 	return nil, fmt.Errorf("client config is required")
	// }

	opts := make([]client.Option, 0)

	// 1. 添加自定义选项
	opts = append(opts, builder.ClientOptions...)

	// 2. 配置传输协议
	switch config.TransportType {
	case "grpc":
		opts = append(opts, client.WithTransportProtocol(transport.GRPC))
	case "grpcstream":
		opts = append(opts, client.WithTransportProtocol(transport.GRPCStreaming))
	default:
		slog.Warn(fmt.Sprintf("invalid transport type: [%s], using default GRPC", config.TransportType))
		opts = append(opts, client.WithTransportProtocol(transport.GRPC))
	}

	// 3. 配置服务发现
	if builder.Resolver != nil {
		opts = append(opts, client.WithResolver(builder.Resolver))
	}

	// 4. 配置负载均衡器
	if builder.LoadBalancer != nil {
		opts = append(opts, client.WithLoadBalancer(builder.LoadBalancer))
	}

	// 4. 配置RPC超时时间
	if config.RpcTimeout > 0 {
		opts = append(opts, client.WithRPCTimeout(config.RpcTimeout*time.Second))
	}

	// 5. 添加中间件
	pmws := append([]PrioritizedClientMiddleware{}, builder.PrioritizedMiddlewares...) // 复制有优先级的中间件
	for _, mw := range builder.Middleware {
		pmws = append(pmws, NewSimplePrioritizedClientMiddleware(ClientMiddlewarePriorityNormal, mw)) // 原始中间件转换成普通优先级的中间件，添加到中间件列表
	}

	if len(pmws) > 0 {
		middlewareOpts := BuildMiddlewareOptions(pmws)
		opts = append(opts, middlewareOpts...)
	}

	connConf := config.Conn

	// 配置GRPC连接池大小
	if connConf.GRPCConnPoolSize > 0 {
		opts = append(opts, client.WithGRPCConnPoolSize(connConf.GRPCConnPoolSize))
	}

	suite := &KitexClientSuite{
		opts: opts,
	}
	return suite, nil
}

// NewKitexClientSuite 创建新的客户端套件
// func NewKitexClientSuite(config *KitexClientConfig, resolver discovery.Resolver, middlewares ...PrioritizedClientMiddleware) (*KitexClientSuite, error) {
// 	builder := &KitexSuiteBuilder{
// 		Config:                 config,
// 		Resolver:               resolver,
// 		PrioritizedMiddlewares: middlewares,
// 	}
// 	return builder.BuildSuite()
// }
