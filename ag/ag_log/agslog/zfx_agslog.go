package agslog

import (
	"log/slog"

	slogmulti "github.com/samber/slog-multi"
	"go.uber.org/fx"
)

type FxInAgSlogBuilderParams struct {
	fx.In

	Props *AgSlogProperties
	// 直接注册的handler
	Handlers  []slog.Handler   `group:"agslog.handler",optional:"true"`
	Handlerss [][]slog.Handler `group:"agslog.handlers",optional:"true"`

	Factorys  []*HandlerFactory   `group:"agslog.factory",optional:"true"`
	Factoryss [][]*HandlerFactory `group:"agslog.factorys",optional:"true"`

	Middlewares []slogmulti.Middleware `group:"agslog.middleware",optional:"true"`
}

var FxAgSlogProvide = fx.Provide(
	BindAgSlogProperties,
	FxBuildAgSlogBuilder,
	FxBuildTopLog,
)

// var FxAgSlogMode = fx.Module("ag_log.agslog",
// 	fx.Provide(
// 		BindAgSlogProperties,
// 		FxBuildAgSlogBuilder,
// 		// BuildAgSlog,
// 	),
// )

// FxBuildTopLog 构建顶层slog logger
func FxBuildTopLog(builder *Builder) (*slog.Logger, error) {
	return builder.Build()
}

// FxBuildAgSlogBuilder 构建slog logger builder
func FxBuildAgSlogBuilder(params FxInAgSlogBuilderParams) (*Builder, error) {
	builder := NewBuilder()

	builder.WithProperties(params.Props)

	builder.AddHandlers(params.Handlers)
	builder.AddHandlerss(params.Handlerss)

	builder.AddHandlerFactorys(params.Factorys)
	builder.AddHandlerFactoryss(params.Factoryss)

	// builder.AddHandlerDefs(params.HandlerDefs...)
	builder.AddMiddlewares(params.Middlewares...)

	return builder, nil
}
