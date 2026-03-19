package agonet

import (
	"go.uber.org/fx"
)

var FxAgonetServerModule = fx.Module("fx_agonet_server",
	fx.Provide(
		NewServerConfig,
		NewServer,
		WarpServer,
	),
)

var FxAgonetClientModule = fx.Module("fx_agonet_client",
	fx.Provide(
		NewClientConfig,
		NewClient,
	),
)
