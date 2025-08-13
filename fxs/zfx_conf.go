package fxs

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_server"

	"go.uber.org/fx"
)

var FxAgConfModule = fx.Module("ag_conf",
	fx.Provide(
		// 创建Enviroment并初始化解析环境变量和-D参数
		fx.Annotate(
			ag_conf.NewStandardEnvironment,
			fx.As(new(ag_conf.IConfigurableEnvironment)),
		),

		// 加载本地配置
		ag_conf.LoadLocalConfigToState, // 在provide阶段解析初始化本地配置，并返回一个本地初始化完成的标志，方便其他要依赖本地配置的组件控制初始化顺序
		fx.Annotate(
			func(env ag_conf.IConfigurableEnvironment, lcled ag_conf.LocalConfigLoded) ag_conf.IBinder { // 添加LocalConfigLoded依赖，控制本地配置先加载
				return ag_conf.NewConfigurationPropertiesBinder(env)
			},
			fx.As(new(ag_conf.IBinder)),
		),
	),
	fxAgConfigWatcherModule,
)

var FxAgConfigDecryptModule = fx.Module(
	"ag_conf_decrypt",
	fx.Invoke(
		ag_conf.DecryptOtherConfig,
	),
)

var fxAgConfigWatcherModule = fx.Module(
	"ag_conf_watcher",
	fx.Provide(
		ag_conf.NewConfigWatcherManager,
		fx.Annotate(
			ag_conf.NewWatcherServer,
			fx.As(new(ag_server.Server)),
			fx.ResultTags(`group:"ag_servers"`),
		),
	),
)
