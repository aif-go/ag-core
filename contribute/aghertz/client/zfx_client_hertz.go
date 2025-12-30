package client

import (
	"ag-core/contribute/aghertz/metadata"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/common/config"
	"go.uber.org/fx"
)

var FxModuleAgHertzClient = fx.Module(
	"aghertz_client",
	fx.Provide(
		FxNewHertzClientParams,
		NewHertzClient,
		// AgMetadate Hertz 客户端 元数据处理 中间件
		NewFxClientMiddlewareProvider(
			metadata.NewAgHertzClientAgMetadataMiddleware,
		),
	),
)

// FxHertzClientMiddlewareParams fx注入的Hertz客户端中间件参数
type FxHertzClientMiddlewareParams struct {
	fx.In

	ClientOptions                    []*config.ClientOption           `group:"aghertz_client_options" ,optional:"true"`                // 客户端选项
	ClientMiddleware                 []client.Middleware              `group:"aghertz_client_middleware" ,optional:"true"`             // 普通中间件（不保证顺序）
	PrioritizedClientMiddleware      []PrioritizedClientMiddleware    `group:"aghertz_client_prioritized_middleware" ,optional:"true"` // 带优先级的中间件（保证顺序）
	PrioritizedClientMiddlewareSuite PrioritizedClientMiddlewareSuite `optional:"true"`                                                // 带优先级的中间件套件（用于组合多个中间件）
	Resolver                         discovery.Resolver               `optional:"true"`                                                // 服务发现中间件，放在param中处理
}

// NewFxClientOptionsProvider 创建fx客户端选项提供者
func NewFxClientOptionsProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"aghertz_client_options"`),
	)
}

// NewFxClientMiddlewareProvider 创建普通中间件提供者
func NewFxClientMiddlewareProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"aghertz_client_middleware"`),
	)
}

// NewFxClientMiddlewareProvider 创建fx中间件提供者
// 用于将普通中间件转换为带优先级的中间件，便于通过fx group注入
func NewFxClientPrioritizedMiddlewareProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"aghertz_client_prioritized_middleware"`),
	)
}

// NewFxClientPrioritizedMiddlewareSuiteProvider 创建带优先级的中间件套件提供者,提供一组中间件选项
func NewFxClientPrioritizedMiddlewareSuiteProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"aghertz_client_prioritized_middleware_suite"`),
	)
}

// FxNewHertzClientParams fx注入创建Hertz客户端参数
func FxNewHertzClientParams(iparam FxHertzClientMiddlewareParams) *HertzClientParams {
	return &HertzClientParams{
		ClientOptions:                    iparam.ClientOptions,
		ClientMiddleware:                 iparam.ClientMiddleware,
		PrioritizedClientMiddleware:      iparam.PrioritizedClientMiddleware,
		PrioritizedClientMiddlewareSuite: iparam.PrioritizedClientMiddlewareSuite,
		Resolver:                         iparam.Resolver,
	}
}
