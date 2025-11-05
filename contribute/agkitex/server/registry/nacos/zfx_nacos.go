package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"go.uber.org/fx"
)

type FxInParam struct {
	fx.In

	NamingClient naming_client.INamingClient `optional:"true"`
}

var FxNacosRegistyModule = fx.Module("fx_kitex_registry_nacos",
	fx.Provide(
		NewProperties,
		NewParam,
		NewRegisty,
	),
)

func NewParam(in FxInParam) *Param {
	return &Param{
		NamingClient: in.NamingClient,
	}
}
