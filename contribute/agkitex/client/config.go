package client

import (
	"ag-core/ag/ag_conf"
	"log/slog"
	"runtime"
	"time"
)

const (
	// KitexClientPropertiesPrefix 客户端配置前缀
	KitexClientPropertiesPrefix = "kitex.client"
)

// KitexClientConfig 客户端配置
type KitexClientConfig struct {
	// RPC超时时间（秒）
	RpcTimeout time.Duration

	// 传输协议类型
	TransportType string

	// 服务发现配置
	Resolver ResolverConfig

	// 连接配置
	Conn ConnConfig
}

// ResolverConfig 服务发现配置
type ResolverConfig struct {
	// 是否启用服务发现
	Enable bool

	// 服务发现类型
	Type string

	// Nacos配置
	Nacos NacosConfig
}

type ConnConfig struct {
	// GRPC连接池大小
	GRPCConnPoolSize uint32
}

// NacosConfig Nacos配置
type NacosConfig struct {
	Group   string
	Cluster string
}

// GrpcConfig GRPC配置
// type GrpcConfig struct {
// 	Enable            bool `value:"${:false}"`
// 	MaxConnectionIdle int  `value:"${:0}"`
// }

// InitKitexClientConfig 初始化客户端配置
func InitKitexClientConfig(binder ag_conf.IBinder) *KitexClientConfig {
	config := defaultKitexClientConfig()
	binder.Bind(config, KitexClientPropertiesPrefix)
	slog.Debug("KitexClientConfig initialized", slog.Any("config", config))
	return config
}

func defaultKitexClientConfig() *KitexClientConfig {

	numP := runtime.GOMAXPROCS(0)
	connPoolSize := uint32(numP * 3 / 2)

	return &KitexClientConfig{
		RpcTimeout:    30,
		TransportType: "grpc",
		Resolver: ResolverConfig{
			Enable: true,
			Type:   "agnacos",
			Nacos: NacosConfig{
				Group:   "DEFAULT_GROUP",
				Cluster: "DEFAULT",
			},
		},
		Conn: ConnConfig{
			GRPCConnPoolSize: connPoolSize,
		},
	}
}
