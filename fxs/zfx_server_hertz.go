package fxs

import (
	hertz "ag-core/ag/ag_hertz/server"
	"ag-core/ag/ag_server"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"go.uber.org/fx"
)

// FxHertzWithRegistryServerBaseModule 创建HTTP服务，并注册到注册中心
var FxHertzWithRegistryServerBaseModule = fx.Module("fx_hertz_with_registry_server",
	fx.Provide(
		hertz.NewHertzServerProperties,
		FxBuilderHertzSuite,
		hertz.NewHertzServerWithSuit,
	),
	fx.Provide(
		fx.Annotate(
			hertzServerWrapper,
			fx.ResultTags(`group:"ag_servers"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			FxHertzH2COption,
			fx.ResultTags(`group:"hertz_options"`),
		)),
)

func hertzServerWrapper(s *hertz.Server) ag_server.Server {
	return s
}

type FxHertzServerInParam struct {
	fx.In
	HertzServerProperties *hertz.HertzServerProperties
	CustOptions           []config.Option             `group:"hertz_options" ,optional:"true"`
	RouterOptions         []hertz.Option              `group:"hertz_router_options" ,optional:"true"`
	NamingClient          naming_client.INamingClient `optional:"true"`
}

func FxBuilderHertzSuite(params FxHertzServerInParam) (*hertz.HertzOptionSuite, error) {
	build := &hertz.HertzSuiteBuilder{
		HCP:           params.HertzServerProperties,
		NamingClient:  params.NamingClient,
		CustOptions:   params.CustOptions,
		RouterOptions: params.RouterOptions,
	}

	return build.BuildSuite()
}

func FxHertzH2COption(p *hertz.HertzServerProperties) config.Option {
	return server.WithH2C(p.EnableH2C)
}
