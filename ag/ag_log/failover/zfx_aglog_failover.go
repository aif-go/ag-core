package failover

import "go.uber.org/fx"

// FxAgSlogFailoverProvide failover 日志提供器
var FxAgSlogFailoverProvide = fx.Provide(
	BindAgSLogFailoverProperties,
	fx.Annotate(
		NewFailoverHandlerFactorys,
		fx.ResultTags(`group:"agslog.factorys"`),
	),
)
