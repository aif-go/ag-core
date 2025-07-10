package fxs

import (
	"ag-core/ag/ag_nacos"
	"ag-core/ag/ag_nacos/config"
	"ag-core/ag/ag_nacos/naming"

	"go.uber.org/fx"
)

var FxConfNacoMode = fx.Module(
	"fx_conf_nacos",
	fx.Provide(
		// TODO nacos通用类配置不应该在配置模块定义
		ag_nacos.NewNacosProperties,
		ag_nacos.NewNacosServerConfig,
		ag_nacos.NewNacosClientConfig,

		ag_nacos.NewNacosNamingClient, // TODO Naming部分不应该在此配置模块定义
		ag_nacos.NewNacosConfigClient,
	),
	fx.Invoke(ag_nacos.NewNacosRemoteConfig),
)

var FxNacosConfigMode = fx.Module(
	"fx_nacos_config",
	fx.Provide(
		config.NewNacosConfigProperties,
		config.NewNacosConfigClient,
	),
)

var FxNacosNamingMode = fx.Module(
	"fx_nacos_naming",
	fx.Provide(
		naming.NewNacosNamingProperties,
		naming.NewNacosNamingClient,
	),
)

var FxEnableNacosRemoteConfigMode = fx.Module(
	"fx_nacos_remote_configenable",
	fx.Invoke(
		config.EnableNacosRemoteConfig,
	),
)
