package agredis

import "go.uber.org/fx"

// FxInAgRedisBuilderParams fx注入参数，agRedisBuilder构建参数
type FxInAgRedisBuilderParams struct {
	fx.In

	AgRedisConfig *AgRedisProperties

	// TODO 其他扩展能力
}

// FxNewKitexServerSuiteBuilder 构建服务器套件构建器
func FxNewAgRedisClientBuilder(params FxInAgRedisBuilderParams) (*AgRedisClientBuilder, error) {
	builder := &AgRedisClientBuilder{
		Config: params.AgRedisConfig,
	}

	return builder, nil
}
