package fxs

import (
	"github.com/aif-go/ag-core/ag/ag_server"
	"github.com/aif-go/ag-core/ag/ag_server/http"

	"go.uber.org/fx"
)

var FxHttpServerBaseModule = fx.Module("fx_http_server_base",
	fx.Provide(
		http.NewHttpGinServer,
	),

	// fx.Annotate(
	// 	NewHttpGinServer,
	// 	fx.As(new(server.Server)), // 类型不匹配时，可以使用As指定接口类型
	// 	fx.ResultTags(`group:"ag_servers"`),
	// ),

	fx.Provide(
		fx.Annotate(
			// fx.Decorate(func(s *Server) server.Server {
			// 	return s
			// }),
			httpserverWrapper,
			// 	fx.As(new(server.Server)), // 类型不匹配时，可以使用As指定接口类型
			fx.ResultTags(`group:"ag_servers"`),
		),
	),
)

func httpserverWrapper(s *http.Server) ag_server.Server {
	return s
}
