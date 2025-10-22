package registy

import (
	"ag-core/contribute/aghertz/server/registy/nacos"

	"go.uber.org/fx"
)

var FxHertzRegistyModule = fx.Module("fx_hertz_registry",
	nacos.FxNacosRegistyModule, // nacos
	// TODO 其他服务发现实现
)
