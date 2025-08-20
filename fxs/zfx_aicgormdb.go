package fxs

import (
	"ag-core/ag/ag_db/gormdb"

	"go.uber.org/fx"
)

var FxAicGromdbModule = fx.Module(
	"fx_aic_gormdb",
	fx.Provide(
		gormdb.NewDB,
		gormdb.NewRepository,
		gormdb.NewTransactionManager, // TODO db模块,还需进一步进行抽象设计
		gormdb.NewZapGormLog,
		gormdb.NewTmMiddlewareContext,
	),
)
