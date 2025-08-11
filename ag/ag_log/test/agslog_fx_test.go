package test

import (
	"ag-core/ag/ag_log/agslog"
	"ag-core/ag/ag_log/logzap"
	"ag-core/ag/ag_log/slogzap"
	"context"
	"log/slog"
	"os"
	"testing"

	"go.uber.org/fx"
)

type FxInSlog struct {
	fx.In

	Handlers []slog.Handler `group:"agslog.handler"`
	// Handlerss [][]slog.Handler `group:"agslog.handlers"`
}

func TestAgLogFx(t *testing.T) {
	fxapp := fx.New(
		fx.Provide(
			// 构建zaplog
			logzap.NewZapLog,
			// logzap.NewZapLog2,

			fx.Annotate(
				slogzap.NewSlogHandler4ZapProps,
				fx.ResultTags(`group:"agslog.handlers"`),
			),

			fx.Annotate(
				slogzap.NewSlog4Zap,
				fx.ResultTags(`group:"agslog.handler"`),
			),

			// 构建slog handler console实现
			fx.Annotate(
				func() slog.Handler {
					return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})
				},
				fx.ResultTags(`group:"ag.slog.handler"`),
			),

			// 初始化 agslog.SlogOption
			func(in FxInSlog) *agslog.Builder {
				builder := agslog.NewBuilder()
				builder.RegTopHandler(in.Handlers...)
				return builder
			},

			// 构建slog
			agslog.BuildAgSlog,
		),
		fx.Invoke(func(logger *slog.Logger) {
			log := logger

			log = log.With("k1", "v1")
			_ag_log_fx_log(log)

			log = log.WithGroup("hzw")
			log = log.With("k2", "v2")
			_ag_log_fx_log(log)
		}),
	)

	fxapp.Start(context.Background())
}

func _ag_log_fx_log(logger *slog.Logger) {
	logger.Info("loginfo")
	logger.Debug("logdebug")
	logger.Warn("logwarn")
	logger.Error("logerror")
}
