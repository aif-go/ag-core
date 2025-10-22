package server

import (
	"github.com/cloudwego/hertz/pkg/app"
)

// type RouteOption ServerOption
type Middleware app.HandlerFunc

// Route is the route def of Hertz.
type Route struct {
	HttpMethod   string
	RelativePath string
	Handlers     []app.HandlerFunc
}

func NewRoute(method, path string, handlers ...app.HandlerFunc) *Route {
	return &Route{
		HttpMethod:   method,
		RelativePath: path,
		Handlers:     handlers,
	}
}

// type RouteMiddleware struct {
// 	Path       string
// 	Middleware app.HandlerFunc
// }

// RouteMw is the middleware of a route.
// type RouteMw struct {
// 	Path        string
// 	Middlewares []app.HandlerFunc
// }

// type RouteOptionParam struct {
// 	Route []*RouteOption
// }

// // WithRoute adds a route to Hertz.
// func WithRoute(route Route) RouteOption {
// 	return RouteOption{
// 		F: func(h *server.Hertz) error {
// 			h.Handle(route.HttpMethod, route.RelativePath, route.Handlers...)
// 			return nil
// 		},
// 	}
// }

// // OptionHertzServerRoute applies the routes to Hertz.
// func OptionHertzServerRoute(h *server.Hertz, fp *RouteOptionParam) error {
// 	slog.Info("apply hertz server routes")
// 	for _, route := range fp.Route {
// 		err := route.F(h)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
