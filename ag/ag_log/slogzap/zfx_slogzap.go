package slogzap

import (
	"go.uber.org/fx"
)

var FxAgSlogZapProvide = fx.Provide(
	BindSlogZapProperties,
	fx.Annotate(
		NewSlogHandler4ZapProps,
		fx.ResultTags(`group:"agslog.handlers"`),
	),
)
