package fxs

import (
	"github.com/aif-go/ag-core/ag/ag_server"
	"github.com/aif-go/ag-core/ag/ag_server/hello"

	"go.uber.org/fx"
)

var FxHelloServerMode = fx.Module("helloServer",

	fx.Provide(
		// NewHelloServer,
		fx.Annotate(
			hello.NewHelloServer,
			fx.As(new(ag_server.Server)), // 类型不匹配时，可以使用As指定接口类型
			fx.ResultTags(`group:"ag_servers"`),
		),
	),
)
