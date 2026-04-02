package agsarama

import "go.uber.org/fx"

var FxAgsaramaConfigModule = fx.Module("fx_agsarama_base",
	fx.Provide(
		// 提供agsarama配置
		NewAgsaramaConfig,
		// 提供agsarama配置转换函数
		TransConfigToSaramaConfig,
		// 提供agsarama客户端创建函数
		NewClientWithAgConfig,
	),
)
