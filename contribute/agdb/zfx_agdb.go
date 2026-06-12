package agdb

import (
	"github.com/aif-go/ag-core/ag/ag_service"
	"github.com/aif-go/ag-core/contribute/agdb/agdao"

	"go.uber.org/fx"
)

var FxAgDbModule = fx.Module(
	"fx_agdb_module",
	fx.Provide(
		ag_service.NewFxAgGlobalMiddleware(
			NewTransactionMiddlewareProvider,
		),
		ag_service.NewFxAgCallInfoOpt(
			func() ag_service.CallInfoOpt {
				return TransactionPreOpt
			},
		),
		// 基础Dao的基础增强能力
		agdao.FxNewAgGormBaseDao,
	),
	// fx.Invoke(func(rtm TransactionManager) {
	// 	setTransactionManager(rtm)
	// }),
)
