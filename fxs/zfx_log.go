package fxs

import (
	"github.com/aif-go/ag-core/ag/ag_log/agslog"
	"github.com/aif-go/ag-core/ag/ag_log/slogzap"

	"go.uber.org/fx"
)

// Deprecated: use ag_log.FxAglogMode instead
var FxAgSlogMode = fx.Module("ag_log.agslog",
	agslog.FxAgSlogProvide,
)

// Deprecated: use ag_log.FxAglogMode instead
var FxAgSlogZapMode = fx.Module("ag_log.agslogzap",
	slogzap.FxAgSlogZapProvide,
)
