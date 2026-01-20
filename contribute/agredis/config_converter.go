package agredis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

func NewUniversalOptionsWithAgUniversalProperties(props AgUniversalOptionsProperties) *redis.UniversalOptions {
	uniopt := &redis.UniversalOptions{
		// 地址配置
		Addrs: props.Addrs,

		// 客户端名称
		ClientName: props.ClientName,

		// 数据库配置
		DB: props.DB,

		// 协议和认证配置
		Protocol: props.Protocol,
		Username: props.Username,
		Password: props.Password,

		// Sentinel认证配置
		SentinelUsername: props.SentinelUsername,
		SentinelPassword: props.SentinelPassword,

		// 重试配置
		MaxRetries:      props.MaxRetries,
		MinRetryBackoff: unitToMillisecond(props.MinRetryBackoff),
		MaxRetryBackoff: unitToMillisecond(props.MaxRetryBackoff),

		// 超时配置
		DialTimeout:           unitToMillisecond(props.DialTimeout),
		ReadTimeout:           unitToMillisecond(props.ReadTimeout),
		WriteTimeout:          unitToMillisecond(props.WriteTimeout),
		ContextTimeoutEnabled: props.ContextTimeoutEnabled,

		// 缓冲区配置
		ReadBufferSize:  props.ReadBufferSize,
		WriteBufferSize: props.WriteBufferSize,

		// 连接池配置
		PoolFIFO:        props.PoolFIFO,
		PoolSize:        props.PoolSize,
		PoolTimeout:     unitToMillisecond(props.PoolTimeout),
		MinIdleConns:    props.MinIdleConns,
		MaxIdleConns:    props.MaxIdleConns,
		MaxActiveConns:  props.MaxActiveConns,
		ConnMaxIdleTime: unitToMillisecond(props.ConnMaxIdleTime),
		ConnMaxLifetime: unitToMillisecond(props.ConnMaxLifetime),

		// 集群相关配置
		MaxRedirects:   props.MaxRedirects,
		ReadOnly:       props.ReadOnly,
		RouteByLatency: props.RouteByLatency,
		RouteRandomly:  props.RouteRandomly,

		// 哨兵相关配置
		MasterName: props.MasterName,

		// 客户端标识配置
		DisableIdentity:  props.DisableIdentity,
		DisableIndentity: props.DisableIndentity, // 兼容旧版本拼写
		IdentitySuffix:   props.IdentitySuffix,

		// 集群故障处理配置
		FailingTimeoutSeconds: props.FailingTimeoutSeconds,

		// RESP3协议配置
		UnstableResp3: props.UnstableResp3,

		// 集群模式配置
		IsClusterMode: props.IsClusterMode,

		// 以下字段在AgUniversalOptionsProperties中未定义，但可以根据需要添加
		// Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 	// 自定义连接拨号函数
		// 	return net.DialTimeout(network, addr, props.DialTimeout)
		// },
		// OnConnect: nil, // 可以添加连接回调函数
		// CredentialsProvider: nil, // 可以添加凭据提供者
		// CredentialsProviderContext: nil, // 可以添加上下文凭据提供者
		// StreamingCredentialsProvider: nil, // 可以添加流凭据提供者
		// TLSConfig: nil, // 可以添加TLS配置
		// MaintNotificationsConfig: nil, // 可以添加维护通知配置

	}
	return uniopt
}

// 单位毫秒
func unitToMillisecond(d time.Duration) time.Duration {
	return d * time.Millisecond
}
