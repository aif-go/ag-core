package simple

import (
	"ag-core/contribute/agonet"

	"go.uber.org/fx"
)

// FxShortClientParams
type FxShortClientParams struct {
	fx.In

	Client agonet.Client

	Opts []ShortClientOption `group:"agonet_simple",optional:"true"`
}

func NewShortClientFx(param FxShortClientParams) (SimpleShortClient, error) {
	// 创建 SimpleShortClient 实例
	return NewSimpleShortClient(param.Client, param.Opts...)
}

func FxAgonetSimpleGroupTag(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"agonet_simple"`),
	)
}
