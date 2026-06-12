package discovery

import (
	"github.com/aif-go/ag-core/contribute/aghertz/client/discovery/nacos"

	"go.uber.org/fx"
)

var FxHertzResolverModule = fx.Module("fx_hertz_resolver",
	nacos.FxNacosResolverModule, // nacos
	// TODO 其他服务发现实现
)
