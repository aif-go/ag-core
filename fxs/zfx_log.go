package fxs

import (
	"ag-core/ag/ag_log/agslog"
	"ag-core/ag/ag_log/slogzap"

	"go.uber.org/fx"
)

var FxAgSlogMode = fx.Module("ag_log.agslog",
	agslog.FxAgSlogProvide,
)

var FxAgSlogZapMode = fx.Module("ag_log.agslogzap",
	slogzap.FxAgSlogZapProvide,
)
