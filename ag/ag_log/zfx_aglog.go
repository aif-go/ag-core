package ag_log

import (
	"ag-core/ag/ag_log/agslog"
	"ag-core/ag/ag_log/async"
	"ag-core/ag/ag_log/fanout"
	"ag-core/ag/ag_log/slogzap"
	"log/slog"

	"go.uber.org/fx"
)

var FxAglogMode = fx.Module("ag_log",
	// agslog
	agslog.FxAgSlogProvide,
	// fanout
	fanout.FxAgSlogFanoutProvide,
	// async
	async.FxAglogAsyncProvide,

	slogzap.FxAgSlogZapProvide,

	fx.Invoke(func(logger *slog.Logger) {
		logger.Info("slog is ready")
	}),
)
