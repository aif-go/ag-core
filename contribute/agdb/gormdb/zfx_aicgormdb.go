package gormdb

import (
	"go.uber.org/fx"
)

// var FxAicGromdbModule = fx.Module(
// 	"fx_aic_gormdb",
// 	fx.Provide(
// 		NewDB,
// 		NewRepository,
// 		NewTransactionManager, // TODO db模块,还需进一步进行抽象设计
// 		// NewZapGormLog,
// 		NewSLogGormLog,
// 		// NewTmMiddlewareContext,
// 	),
// )

var FxAicGromdbModule = fx.Module(
	"fx_aic_gormdb",
	fx.Provide(
		NewAggormDbConfig,
		NewDB_V2,
		NewRepository,
		NewTransactionManager,
		FindGormLoggerFromAgslog,
	),
)
