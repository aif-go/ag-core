package test

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_log/agslog"
	"ag-core/ag/ag_log/fanout"
	"ag-core/ag/ag_log/slogzap"
	"context"
	"log/slog"
	"testing"

	"go.uber.org/fx"
)

func TestAgSlogFx2(t *testing.T) {
	fxapp := fx.New(
		fx.Provide(
			func() ag_conf.IBinder {
				env, _ := ag_conf.NewStandardEnvironment()
				ag_conf.LoadConfigFile(env, "agslog_fxbyconf_test.yaml")
				binder := ag_conf.NewConfigurationPropertiesBinder(env)
				return binder
			},
		),

		slogzap.FxAgSlogZapProvide,

		fanout.FxAgSlogFanoutProvide,

		agslog.FxAgSlogProvide,

		fx.Invoke(func(logger *slog.Logger) {
			log := logger

			// log = log.With("k1", "v1")
			_ag_log_fx_log(log)

			// log = log.WithGroup("hzw")
			// log = log.With("k2", "v2")
			// _ag_log_fx_log(log)
		}),
	)

	fxapp.Start(context.Background())
}
