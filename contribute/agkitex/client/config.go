package client

import (
	"ag-core/ag/ag_conf"
	"log/slog"
	"time"
)

const (
	// KitexClientPropertiesPrefix 客户端配置前缀
	KitexClientPropertiesPrefix = "kitex.client"
)

// KitexClientConfig 客户端配置
type KitexClientConfig struct {
	// RPC超时时间（秒）
	RpcTimeout time.Duration `value:"${:30}"`

	// 传输协议类型
	TransportType string `value:"${:grpc}"`

	// 服务发现配置
	Resolver ResolverConfig

	// 连接配置
	Conn ConnConfig
}

// ResolverConfig 服务发现配置
type ResolverConfig struct {
	// 是否启用服务发现
	Enable bool `value:"${:false}"`

	// 服务发现类型
	Type string `value:"${type:agnacos}"`

	// Nacos配置
	Nacos NacosConfig `value:"${Nacos}"`
}

type ConnConfig struct {
	// GRPC连接池大小
	GRPCConnPoolSize uint32 `value:"${:1}"`
}

// NacosConfig Nacos配置
type NacosConfig struct {
	Group   string `value:"${group:DEFAULT_GROUP}"`
	Cluster string `value:"${cluster:DEFAULT}"`
}

// GrpcConfig GRPC配置
type GrpcConfig struct {
	Enable            bool `value:"${:false}"`
	MaxConnectionIdle int  `value:"${:0}"`
}

// InitKitexClientConfig 初始化客户端配置
func InitKitexClientConfig(binder ag_conf.IBinder) *KitexClientConfig {
	config := &KitexClientConfig{}
	binder.Bind(config, KitexClientPropertiesPrefix)
	slog.Debug("KitexClientConfig initialized", slog.Any("config", config))
	return config
}
