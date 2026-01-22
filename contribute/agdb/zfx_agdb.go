package agdb

import (
	"ag-core/ag/ag_service"
	"ag-core/contribute/agdb/agdao"

	"go.uber.org/fx"
)

var FxAgDbModule = fx.Module(
	"fx_agdb_module",
	fx.Provide(
		ag_service.NewFxAgGlobalMiddleware(
			NewTransactionMiddlewareProvider,
		),
		// 基础Dao的基础增强能力
		agdao.FxNewAgGormBaseDao,
	),
)
