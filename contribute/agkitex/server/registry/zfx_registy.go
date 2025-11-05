package registry

import (
	"ag-core/contribute/agkitex/server/registry/nacos"

	"go.uber.org/fx"
)

var FxKitexRegistyModule = fx.Module("fx_kitex_registry",
	nacos.FxNacosRegistyModule, // nacos
	// TODO 其他服务发现实现
)
