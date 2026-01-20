package ag_service

import (
	"go.uber.org/fx"
)

// type BaseFxServiceMiddlewareParams struct {
// 	fx.In
// 	GlobalMws []ServiceMiddlewareProvider `group:"fx_global_service_middleware" ,optional:"true"`
// }

// // NewServiceGlobalMiddleware 创建service全局中间件fx提供者
// func NewServiceGlobalMiddleware(t any) any {
// 	return fx.Annotate(
// 		t,
// 		fx.ResultTags(`group:"fx_global_service_middleware"`),
// 	)
// }

// FxAgServiceMode 服务模块，相关支持
var FxAgServiceMode = fx.Module("ag_service.agservice",
	fx.Provide(
		// NewAgServiceBuilder,
		FxNewAgServiceBuilder,
	),
)

// FxInAgServiceBuilder fx构建AgServiceBuilder的注入参数
type FxInAgServiceBuilder struct {
	fx.In
	CiOpts       []CallInfoOpt                   `group:"fx_service_callinfo_opt",optional:"true"`
	GlobalPMWs   []PrioritizedMiddlewareProvider `group:"fx_global_service_middleware",optional:"true"`
	GlobalMWs    []MiddlewareProvider            `group:"fx_global_service_middleware",optional:"true"`
	GlobalMWFuns []MiddlewareFunc                `group:"fx_global_service_middleware",optional:"true"`
}

// FxNewAgServiceBuilder fx构建AgServiceBuilder
func FxNewAgServiceBuilder(in FxInAgServiceBuilder) (AgServiceBuilder, error) {
	return NewAgServiceBuilder(in.CiOpts, in.GlobalMWs, in.GlobalPMWs, in.GlobalMWFuns)
}

// NewFxAgCallInfoOpt 创建fx调用信息选项提供者
func NewFxAgCallInfoOpt(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"fx_service_callinfo_opt"`),
	)
}

// NewFxAgGlobalMiddleware 创建fx全局中间件提供者
func NewFxAgGlobalMiddleware(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"fx_global_service_middleware"`),
	)
}
