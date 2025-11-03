package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"go.uber.org/fx"
)

type FxInParam struct {
	fx.In

	NamingClient naming_client.INamingClient `optional:"true"`
}

var FxNacosResolverModule = fx.Module("fx_hertz_resolver_nacos",
	fx.Provide(
		NewProperties,
		NewParam,
		NewResolver,
	),
)

func NewParam(in FxInParam) *Param {
	return &Param{
		NamingClient: in.NamingClient,
	}
}
