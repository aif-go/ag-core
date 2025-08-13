package fanout

import "go.uber.org/fx"

var FxAgSlogFanoutProvide = fx.Provide(
	BindAgSLogFanoutProperties,
	fx.Annotate(
		NewFanoutHandlerFactorys,
		fx.ResultTags(`group:"agslog.factorys"`),
	),
)
