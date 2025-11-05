package client

import (
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"go.uber.org/fx"
)

// FxKitexClientBaseModule 客户端基础模块
var FxKitexClientBaseModule = fx.Module(
	"fx_kitex_client_base",
	fx.Provide(
		InitKitexClientConfig,
		BuildKitexResolver,
		FxKitexClientSuiteBuilder,
		FxNewKitexClientSuite,
	),
)

// FxInKitexClientParams fx注入的客户端参数
type FxInKitexClientParams struct {
	fx.In

	Config                 *KitexClientConfig
	ClientOptions          []*client.Option `group:"kitex_client_options",optional:"true"`
	Resolver               discovery.Resolver
	Middleware             []endpoint.Middleware         `group:"kitex_client_middlewares",optional:"true"`
	PrioritizedMiddlewares []PrioritizedClientMiddleware `group:"kitex_client_prioritized_middlewares",optional:"true"`
}

// FxBuilderKitexClientSuite 创建fx客户端套件构建器
func FxKitexClientSuiteBuilder(params FxInKitexClientParams) (*KitexSuiteBuilder, error) {
	builder := &KitexSuiteBuilder{
		Resolver:               params.Resolver,
		Config:                 params.Config,
		Middleware:             params.Middleware,
		PrioritizedMiddlewares: params.PrioritizedMiddlewares,
	}

	// 处理自定义选项
	cliOpt := make([]client.Option, 0)
	for _, opt := range params.ClientOptions {
		cliOpt = append(cliOpt, *opt)
	}
	builder.ClientOptions = cliOpt

	return builder, nil
}

// FxNewKitexClientSuite 创建新的客户端套件
func FxNewKitexClientSuite(builder *KitexSuiteBuilder) (*KitexClientSuite, error) {
	return builder.BuildSuite()
}

// NewFxClientOptionsProvider 创建fx客户端选项提供者
func NewFxClientOptionsProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"kitex_client_options"`),
	)
}

// NewFxClientMiddlewareProvider 创建fx中间件提供者
func NewFxClientMiddlewareProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"kitex_client middlewares"`),
	)
}

// NewFxClientPrioritizedMiddlewareProvider 创建fx优先级中间件提供者
func NewFxClientPrioritizedMiddlewareProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"kitex_client_prioritized_middlewares"`),
	)
}

// // FxInKitexClientMiddlewareParams fx注入的客户端中间件参数
// type FxInKitexClientMiddlewareParams struct {
// 	fx.In

// 	Middlewares []PrioritizedClientMiddleware `group:"kitex_client_middlewares",optional:"true"`
// }

// // FxBuildKitexClientMiddleware 构建客户端中间件
// func FxBuildKitexClientMiddleware(p FxInKitexClientMiddlewareParams) []PrioritizedClientMiddleware {
// 	return p.Middlewares
// }
