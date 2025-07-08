package server

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_ext/ip"
	"context"
	"fmt"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
	"log/slog"
	"net"
	"time"

	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/grpc"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/registry-nacos/registry"

	kregistry "github.com/cloudwego/kitex/pkg/registry"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/spf13/cast"

	_ "github.com/cloudwego/kitex/pkg/remote/codec/protobuf/encoding/gzip"
)

type Server struct {
	ksvr   server.Server
	logger *slog.Logger
}

// NewServer (engine *gin.Engine, logger *log.Logger, opts ...Option) *Server {
func NewServer(
	svr server.Server,
	logger *slog.Logger,
	// creporter *HzwKCReporter,
) *Server {
	s := &Server{
		ksvr:   svr,
		logger: logger,
	}
	// connpool.SetReporter(creporter) // 注册自定义连接池监控，TODO是否有用？
	return s
}

func (s *Server) Start(ctx context.Context) error {
	sinfos := s.ksvr.GetServiceInfos()
	// 格式化打印服务信息
	for sname, sinfo := range sinfos {
		s.logger.Info("Kitex服务信息",
			"服务名称", sname,
			"方法数量", len(sinfo.Methods),
			"方法列表", sinfo.Methods,
			"额外信息", sinfo.Extra,
		)
	}

	s.logger.Info("Kitex服务启动",
		"总服务数", len(sinfos),
	)
	return s.ksvr.Run()
}
func (s *Server) Stop(ctx context.Context) error {
	return s.ksvr.Stop()
}

// // HzwKCReporter 自定义连接池监控
// type HzwKCReporter struct{}

// func NewHzwKCReporter() *HzwKCReporter {
// 	return &HzwKCReporter{}
// }

// func (r *HzwKCReporter) ConnSucceed(poolType connpool.ConnectionPoolType, serviceName string, addr net.Addr) {
// 	fmt.Printf("ConnSucceed poolType:%d, serviceName:%s, addr:%v\n", poolType, serviceName, addr)
// }
// func (r *HzwKCReporter) ConnFailed(poolType connpool.ConnectionPoolType, serviceName string, addr net.Addr) {
// 	fmt.Printf("ConnFailed poolType:%d, serviceName:%s, addr:%v\n", poolType, serviceName, addr)
// }
// func (r *HzwKCReporter) ReuseSucceed(poolType connpool.ConnectionPoolType, serviceName string, addr net.Addr) {
// 	fmt.Printf("ReuseSucceed poolType:%d, serviceName:%s, addr:%v\n", poolType, serviceName, addr)
// }

// NewKitexOriginalServer 创建原始的kitex服务实例
func NewKitexOriginalServer(
	logger *slog.Logger,
	env ag_conf.IConfigurableEnvironment,
	client naming_client.INamingClient,
) server.Server {
	host := env.GetProperty("kitex.host")
	port, err := cast.ToIntE(env.GetProperty("kitex.port"))
	if err != nil {
		panic(err)
	}
	endpoint := env.GetProperty("kitex.endpoint")

	kitexHostStr := fmt.Sprintf("%s:%d", host, port)
	logger.Info("kitex", "host", kitexHostStr)

	addr, _ := net.ResolveTCPAddr("tcp", kitexHostStr)
	/*
		svr := server.NewServer(
			server.WithServiceAddr(addr),
			// 配合 SkipDecoder https://www.cloudwego.io/zh/docs/kitex/tutorials/code-gen/skip_decoder/
			server.WithPayloadCodec(thrift.NewThriftCodecWithConfig(thrift.FastWrite|thrift.FastRead|thrift.EnableSkipDecoder)),
			//server.WithMuxTransport(),                               // 服务端开启多路复用；Server 开启连接多路复用对 Client 没有限制，可以接受短连接、长连接池、连接多路复用的请求
			//server.WithMetaHandler(transmeta.ServerHTTP2Handler),    // 指定基于HTTP2 协议header的信息透传
			server.WithMetaHandler(transmeta.ServerTTHeaderHandler), // 指定基于TTHeader 协议header的信息透传
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: endpoint}),
			server.WithRegistry(registry.NewNacosRegistry(client)),
		)
	*/

	options := []server.Option{
		server.WithServiceAddr(addr),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: endpoint}),
		// server.WithExitWaitTime(30 * time.Second), // 优雅停机等待时间
		// server.WithStatsLevel(stats.LevelDetailed),
		// server.WithMaxConnIdleTime(time.Second * 30),
		server.WithGRPCKeepaliveParams(grpc.ServerKeepalive{ // 设置GRPC Keepalive参数 TODO 修改为配置配置方式
			MaxConnectionIdle: time.Second * 50, // 最大空闲连接时间
		}),
		server.WithStatsLevel(stats.LevelDisabled),
		// server.WithMiddleware()
		// server.WithSuite(...), // TODO 调整修改为Suite的方式统一管理
	}
	if client != nil {
		slog.Info("kitex server enable nacos naming")
		options = append(options, server.WithRegistry(registry.NewNacosRegistry(client)))
	}
	svr := server.NewServer(options...)

	return svr
}

type KitexServerRegistrar struct {
	Regs []Option
}

func (s *KitexServerRegistrar) Registrars() []Option {
	return s.Regs
}

type Option func(server.Server)

type ServiceRegistrar struct {
	*kitex.ServiceInfo
	Handler interface{}
}

func WithServiceRegistrar(sr *ServiceRegistrar) Option {
	return func(s server.Server) {
		err := s.RegisterService(sr.ServiceInfo, sr.Handler)
		if err != nil {
			panic(err)
		}
	}
}

func NewKitexServerWithSuit(
	suite server.Suite,
	registrar KitexServerRegistrar,
) server.Server {
	svr := server.NewServer(server.WithSuite(suite))
	for _, r := range registrar.Registrars() {
		r(svr)
	}
	return svr
}

type KitexServerSuite struct {
	Opts []server.Option
}

func (s *KitexServerSuite) Options() []server.Option {
	return s.Opts
}

type KitexSuiteBuilder struct {
	Env    ag_conf.IConfigurableEnvironment
	Binder ag_conf.IBinder

	CustOptions []server.Option

	NamingClient naming_client.INamingClient
}

// BuildSuite KitexServer的配置构建器，最终提供一个server.Suite 供NewKitexServerWithSuit创建kitex服务使用
func (builder *KitexSuiteBuilder) BuildSuite() (server.Suite, error) {
	suite := &KitexServerSuite{
		Opts: make([]server.Option, 0),
	}

	// 自定义的配置项
	suite.Opts = append(suite.Opts, builder.CustOptions...)

	var kconf KitexServerProperties
	err := builder.Binder.Bind(&kconf, KitexServerPropertiesPrefix)
	if err != nil {
		slog.Error("kitex server config error", "err", err)
		return nil, err
	}

	// 注册中心配置
	if builder.NamingClient != nil {
		slog.Info("kitex server enable nacos naming")
		suite.Opts = append(suite.Opts, server.WithRegistry(registry.NewNacosRegistry(builder.NamingClient)))
	}

	// 服务地址配置
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
	suite.Opts = append(suite.Opts, server.WithServiceAddr(addr))

	// 服务信息配置
	sname := kconf.ServiceName
	if sname == "" {
		sname = "kitex-server"
	}
	info := &rpcinfo.EndpointBasicInfo{
		ServiceName: sname,
	}
	suite.Opts = append(suite.Opts, server.WithServerBasicInfo(info))

	// 自定义注册信息
	regInfo := &kregistry.Info{
		Tags: map[string]string{},
	}
	regInfo.Weight = 1
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
		// 兼容https://github.com/grpc-ecosystem/grpc-spring项目的服务发现实现
		regInfo.Tags["gRPC_port"] = fmt.Sprintf("%d", port)

	}
	suite.Opts = append(suite.Opts, server.WithRegistryInfo(regInfo))

	// Grpc配置
	if kconf.Grpc.Enable {
		gskeep := grpc.ServerKeepalive{}
		if kconf.Grpc.MaxConnectionIdle > 0 {
			// 最大空闲连接时间
			gskeep.MaxConnectionIdle = time.Second * time.Duration(kconf.Grpc.MaxConnectionIdle)
		}
		suite.Opts = append(suite.Opts, server.WithGRPCKeepaliveParams(gskeep))

		server.WithMetaHandler(transmeta.ClientHTTP2Handler)
	}

	// options := []server.Option{
	// 	server.WithServiceAddr(addr),
	// 	server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
	// 	server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: sname}),
	// 	server.WithGRPCKeepaliveParams(grpc.ServerKeepalive{ // 设置GRPC Keepalive参数 TODO 修改为配置配置方式
	// 		MaxConnectionIdle: time.Second * 50, // 最大空闲连接时间
	// 	}),
	// 	server.WithStatsLevel(stats.LevelDisabled),
	// }

	return suite, nil
}

func findKitexHostPort(kconf KitexServerProperties) (host string, port int, rerr error) {
	// 服务ip、端口配置
	host = kconf.Host
	if host == "" {
		host = "0.0.0.0"
	}

	if !ip.IsHostAvailable(host) {
		return "", 0, fmt.Errorf("kitex host unavailable: %s", host)
	}

	port = kconf.Port
	if kconf.AdaptivePort {
		slog.Info("kitex server enable adaptive port")
		if port == 0 {
			port = DefaultKitexOriginPort
		}
		port, rerr = ip.GetAvailablePort(host, port)
		if rerr != nil {
			return
		}
	} else {
		if port == 0 {
			return host, port, fmt.Errorf("kitex port invalid:%v", port)
		}
	}

	slog.Info(fmt.Sprintf("finded kitex host:%s, port:%d", host, port))
	return
}
