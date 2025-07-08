package fxs

import (
	agkc "ag-core/ag/ag_kitex/client"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/discovery"
	"go.uber.org/fx"
)

var FxKitexClientBaseModule = fx.Module(
	"fx_kitex_Client_base",
	fx.Provide(
		agkc.FxInitKitexClientProperties,
		agkc.BuildKitexResolver,
		FxBuilderKitexClientSuite,
	),
)

var FxKitexAgClientBizErrorOption = fx.Provide(
	fx.Annotate(
		agkc.NewAgBizErrorMiddlewareOption,
		fx.ResultTags(`group:"kitex_client_options"`),
	),
	// fx.Annotate(
	// 	agkc.NewAgBizErrorHandler,
	// 	fx.ResultTags(`group:"kitex_client_options"`),
	// ),
)

type FxInKitexClientParams struct {
	fx.In

	KCProps     *agkc.KitexClientProperties
	CustOptions []*client.Option `group:"kitex_client_options",optional:"true"`

	Resolver discovery.Resolver
}

// func FxBuilderKitexClientSuite(params FxInKitexServerParams) (server.Suite, error) {
func FxBuilderKitexClientSuite(params FxInKitexClientParams) (*agkc.KitexClientSuite, error) {
	build := &agkc.KitexSuiteBuilder{
		Resolver: params.Resolver,
		KCProps:  params.KCProps,
	}
	// CustOptions:  params.CustOptions,
	custOpt := make([]client.Option, 0)
	for _, opt := range params.CustOptions {
		custOpt = append(custOpt, *opt)
	}
	build.CustOptions = custOpt

	return build.BuildSuite()
}
