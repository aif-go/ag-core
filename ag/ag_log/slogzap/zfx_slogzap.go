package slogzap

import (
	"go.uber.org/fx"
)

var FxAgSlogZapMode = fx.Module("ag_log.agslogzap",
	fx.Provide(
		BindSlogZapProperties,
		fx.Annotate(
			NewSlogHandler4ZapProps,
			fx.ResultTags(`group:"agslog.handlers"`),
		),
	),
)
