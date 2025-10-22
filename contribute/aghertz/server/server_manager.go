package server

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// ServerConfigurator 服务器组件及路由配置器
type ServerConfigurator struct {
	Server *server.Hertz   // Hertz 服务
	Opts   []*ServerOption // 服务器选项
	Routes []*Route        // 路由
	Mws    []Middleware    // 全局中间件
}

type ServerConfiguratorParam struct {
	Server *server.Hertz
	Opts   []*ServerOption
	Routes []*Route
}

func NewServerConfiguratorWithParam(param *ServerConfiguratorParam) *ServerConfigurator {
	return &ServerConfigurator{
		Server: param.Server,
		Opts:   param.Opts,
		Routes: param.Routes,
	}
}

func (m *ServerConfigurator) AddRoute(route *Route) {
	m.Routes = append(m.Routes, route)
}

func (m *ServerConfigurator) InitHertzServer() error {
	// 应用服务器选项
	err := m.ApplyServerOptions()
	if err != nil {
		return err
	}

	// 应用全局中间件
	m.ApplyMiddleware()

	// 应用路由
	err = m.ApplyRoute()
	if err != nil {
		return err
	}
	// TODO 路由级别的中间件应用
	return nil
}

func (m *ServerConfigurator) ApplyRoute() error {
	for _, route := range m.Routes {
		m.Server.Handle(route.HttpMethod, route.RelativePath, route.Handlers...)
	}
	return nil
}

// ApplyServerOptions 应用服务器选项
func (m *ServerConfigurator) ApplyServerOptions() error {
	suite := &SimpleServerSuite{}
	suite.AddPtr(m.Opts...)
	return OptionHertzServerSuite(m.Server, suite)
}

func (m *ServerConfigurator) ApplyMiddleware() {
	// 应用全局中间件
	for _, mw := range m.Mws {
		m.Server.Use(app.HandlerFunc(mw))
	}
}
