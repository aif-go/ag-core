package registry

import (
	"ag-core/contribute/aghertz/server/registry/nacos"

	"go.uber.org/fx"
)

var FxHertzRegistyModule = fx.Module("fx_hertz_registry",
	nacos.FxNacosRegistyModule, // nacos
	// TODO 其他服务发现实现
)
