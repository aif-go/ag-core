package server

import (
	"ag-core/ag/ag_server"
	"ag-core/contribute/agkitex/metadata"
	agkitexReg "ag-core/contribute/agkitex/server/registry"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/server"
	"go.uber.org/fx"
)

// FxKitexServerBaseModule 服务器基础模块
var FxKitexServerBaseModule = fx.Module("fx_kitex_server_base",
	// 注册中心
	agkitexReg.FxKitexRegistyModule,

	fx.Provide(
		// 配置
		NewKitexServerProperties,
		// 服务器套件构建器
		FxNewKitexServerSuiteBuilder,
		// 服务器套件构建器，构建服务器套件
		FxBuilderKitexServerSuite,
		// 使用服务器套件创建kitex服务器
		NewKitexServerWithSuite,
		// 包装为kitex服务为agkitex服务，提供ag_app服务启动入口
		fx.Annotate(
			NewAgKitexServer,
			fx.As(new(ag_server.Server)),
			fx.ResultTags(`group:"ag_servers"`),
		),

		// 服务注册器持有者
		FxNewKitexServiceRegistryHolder,

		// AgMetadate Kitex HTTP2 服务端 元数据处理
		NewFxServerOptionsProvider(
			metadata.NewAgKitexServerAgMetadataHTTP2HandlerOption,
		),
	),

	// 调用服务注册器持有者注册服务
	fx.Invoke(
		FxInvokerKitexServiceRegistryHolder,
	),
)

// FxInKitexServerParams fx注入的服务器参数
type FxInKitexServerParams struct {
	fx.In

	// 自定义选项
	ServerOptions []*server.Option `group:"kitex_server_options",optional:"true"`

	// 配置
	Config *KitexServerProperties

	// 服务注册器
	Registry registry.Registry

	// 中间件(原始)
	Middlewares []endpoint.Middleware `group:"kitex_server_middlewares",optional:"true"`

	// 中间件(带优先级)
	PrioritizedMiddlewares []PrioritizedServerMiddleware `group:"kitex_server_prioritized_middlewares",optional:"true"`
}

// FxNewKitexServerSuiteBuilder 构建服务器套件构建器
func FxNewKitexServerSuiteBuilder(params FxInKitexServerParams) (*KitexServerSuiteBuilder, error) {
	builder := &KitexServerSuiteBuilder{
		// ServerOptions:          params.ServerOptions,
		Properties:             params.Config,
		Registry:               params.Registry,
		Middlewares:            params.Middlewares,
		PrioritizedMiddlewares: params.PrioritizedMiddlewares,
	}

	// 处理自定义选项
	svrOpt := make([]server.Option, 0)
	for _, opt := range params.ServerOptions {
		svrOpt = append(svrOpt, *opt)
	}
	builder.ServerOptions = svrOpt

	return builder, nil
}

// FxBuilderKitexServerSuite 构建服务器套件
func FxBuilderKitexServerSuite(builder *KitexServerSuiteBuilder) (*KitexServerSuite, error) {
	return builder.BuildServerSuite()
}

// NewFxServerOptionsProvider 创建fx服务器选项提供者
func NewFxServerOptionsProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"kitex_server_options"`),
	)
}

// NewFxServerMiddlewareProvider 创建fx中间件提供者
func NewFxServerMiddlewareProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"kitex_server_middlewares"`),
	)
}

// NewFxServerPrioritizedMiddlewareProvider 创建fx带优先级中间件提供者
func NewFxServerPrioritizedMiddlewareProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"kitex_server_prioritized_middlewares"`),
	)
}

// FxInKitexServiceRegistryParams fx注入的服务注册器参数
type FxInKitexServiceRegistryParams struct {
	fx.In

	// 服务注册器
	Registries []*AgKitexServiceRegistry `group:"ag_kitex_service_registry"`
}

// FxNewKitexServiceRegistryHolder 创建fx服务注册器持有者
func FxNewKitexServiceRegistryHolder(params FxInKitexServiceRegistryParams) *AgKitexServiceRegistryHolder {
	return &AgKitexServiceRegistryHolder{
		Registries: params.Registries,
	}
}

// FxInvokerKitexServiceRegistryHolder 调用fx服务注册器持有者注册服务
func FxInvokerKitexServiceRegistryHolder(holder *AgKitexServiceRegistryHolder, kitexServer server.Server) error {
	return holder.RegisterService(kitexServer)
}

// NewFxAgKitexServiceRegistry 创建fx服务注册器提供者
func NewFxAgKitexServiceRegistry(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"ag_kitex_service_registry"`),
	)
}
