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

var FxAgSlogMode = fx.Module("ag_log.agslog",
	fx.Provide(
		BindAgSlogProperties,
		FxBuildAgSlogBuilder,
		// BuildAgSlog,
	),
)

func FxBuildAgSlogBuilder(params FxInAgSlogBuilderParams) (*slog.Logger, error) {
	builder := NewBuilder()
	builder.WithProperties(params.Props)
	builder.AddHandlers(params.Handlers...)
	for _, handlers := range params.Handlerss {
		builder.AddHandlers(handlers...)
	}

	builder.AddHandlerFactorys(params.Factorys...)
	for _, factorys := range params.Factoryss {
		builder.AddHandlerFactorys(factorys...)
	}

	// builder.AddHandlerDefs(params.HandlerDefs...)
	builder.AddMiddlewares(params.Middlewares...)

	logger, err := builder.Build()
	if err != nil {
		return nil, err // TODO 日志初始化是否要中断程序
	}
	return logger, nil
}
