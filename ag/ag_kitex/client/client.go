package client

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/transport"

	"github.com/cloudwego/kitex/pkg/discovery"
	_ "github.com/cloudwego/kitex/pkg/remote/codec/protobuf/encoding/gzip"
)

type KitexClientSuite struct {
	opts []client.Option
}

func (s *KitexClientSuite) Options() []client.Option {
	return s.opts
}

type KitexSuiteBuilder struct {
	CustOptions []client.Option

	KCProps  *KitexClientProperties
	Resolver discovery.Resolver
}

// func (builder *KitexSuiteBuilder) BuildSuite() (client.Suite, error) {
func (builder *KitexSuiteBuilder) BuildSuite() (*KitexClientSuite, error) {

	p := builder.KCProps

	opts := make([]client.Option, 0)

	// 自定义的配置项
	opts = append(opts, builder.CustOptions...)

	// 指定传输协议 TODO 某个client若要单独配置怎么办？
	switch p.TransportType {
	case "grpc":
		opts = append(opts, client.WithTransportProtocol(transport.GRPC))
	case "grpcstream":
		opts = append(opts, client.WithTransportProtocol(transport.GRPCStreaming))
	default:
		// opts = append(opts, client.WithTransportProtocol(transport.GRPC))
		slog.Warn(fmt.Sprintf("invalid transport type: [%s]", p.TransportType))
		// return nil, fmt.Errorf("invalid transport type: %s", p.TransportType)
	}

	// 注册中心配置
	if builder.Resolver != nil {
		opts = append(opts, client.WithResolver(builder.Resolver))
	}

	// RPC超时时间参数设置
	if p.RpcTimeout > 0 {
		opts = append(opts, client.WithRPCTimeout(p.RpcTimeout*time.Second))
	}

	// opts = append(opts, client.WithLoadBalancer())

	suite := &KitexClientSuite{
		opts: opts,
	}
	return suite, nil
}
