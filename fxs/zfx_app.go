package fxs

import (
	"github.com/aif-go/ag-core/ag/ag_app"
	"github.com/aif-go/ag-core/ag/ag_server"
	"context"
	"log/slog"

	"go.uber.org/fx"
)

// FxAppParams APP构造参数，自动注入
type FxAppParams struct {
	fx.In
	// Name    string          `name:"application.name"`
	Servers []ag_server.Server `group:"ag_servers"`
	Logger  *slog.Logger
	// Servers []server.Server
	// S1 server.Server `optional:"true"` // 可选参数 TEST
}

// FxApp 根APP fx构造器
func FxApp(params FxAppParams, lc fx.Lifecycle) *ag_app.App {
	app := &ag_app.App{
		// name:    params.Name,
		Servers: params.Servers,
		Logger:  params.Logger,
	}
	// 定义生命周期钩子，用于启动和停止APP
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := app.Start(ctx)
			// 此ctx会自动done，可以通过fx.StartTimeout选项延长启动超时
			return err
		},
		OnStop: func(ctx context.Context) error {
			app.Stop(ctx)
			return nil
		},
	})

	return app
}

var FxAppMode = fx.Module("app",
	fx.Provide(
		FxApp,
	),
)
