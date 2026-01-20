package agdb

import (
	"ag-core/ag/ag_service"

	"go.uber.org/fx"
)

var FxAgDbModule = fx.Module(
	"fx_agdb_module",
	fx.Provide(
		ag_service.NewFxAgGlobalMiddleware(
			NewTransactionMiddlewareProvider,
		),
	),
)
