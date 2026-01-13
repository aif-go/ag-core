package agredis

import (
	"context"

	"go.uber.org/fx"
)

// FxAgRedisServerMode 是一个 Fx 模块，用于初始化 Redis 客户端
var FxAgRedisServerMode = fx.Module("fx_agredis_mode",
	fx.Provide(
		// 初始化配置对象
		NewAgRedisPropertiesByBinder,
		// 注入初始化构建器
		FxNewAgRedisClientBuilder,
		// 通过构建器创建AgRedisClient
		CreateClientByBuilder,
	),
	fx.Invoke(registerHooks),
)

// registerHooks 注册生命周期钩子
// TODO 生命周期钩子是否能在框架层面进行支持，因为套件需要能自恢复能力，需要在框架层面进行支持
func registerHooks(lc fx.Lifecycle, client AgRedisClient) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// 启动时健康检查
			// return client.Ping(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// 关闭连接
			return client.Close()
		},
	})
}
