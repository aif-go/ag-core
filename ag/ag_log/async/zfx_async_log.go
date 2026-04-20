package async

import "go.uber.org/fx"

// FxAglogAsyncProvide 异步日志提供器
var FxAglogAsyncProvide = fx.Provide(
	BindAsyncLogConfig,
	fx.Annotate(
		BuildAsyncHandlerFactorys,
		fx.ResultTags(`group:"agslog.factorys"`),
	),
)
