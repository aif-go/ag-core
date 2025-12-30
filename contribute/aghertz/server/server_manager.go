package server

import (
	"fmt"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// ServerConfigurator 服务器组件及路由配置器
type ServerConfigurator struct {
	Server   *server.Hertz     // Hertz 服务
	Opts     []*ServerOption   // 服务器选项
	Routes   []*Route          // 路由
	Mws      []Middleware      // 全局中间件
	MwsHFunc []app.HandlerFunc // 全局中间件
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
	err = m.ApplyMiddleware()
	if err != nil {
		return err
	}

	// 应用路由
	err = m.ApplyRoute()
	if err != nil {
		return err
	}
	// TODO 路由级别的中间件应用
	return nil
}

func (m *ServerConfigurator) ApplyRoute() error {
	// 应用路由
	if m.Server == nil {
		return fmt.Errorf("hertz server is nil")
	}

	if len(m.Routes) == 0 {
		return nil
	}

	for _, route := range m.Routes {
		slog.Info("apply hertz server route", "httpMethod", route.HttpMethod, "relativePath", route.RelativePath)
		m.Server.Handle(route.HttpMethod, route.RelativePath, route.Handlers...)
	}
	return nil
}

// ApplyServerOptions 应用服务器选项
func (m *ServerConfigurator) ApplyServerOptions() error {
	if m.Server == nil {
		return fmt.Errorf("hertz server is nil")
	}

	if len(m.Opts) == 0 {
		return nil
	}

	suite := &SimpleServerSuite{}
	suite.AddPtr(m.Opts...)
	return OptionHertzServerSuite(m.Server, suite)
}

func (m *ServerConfigurator) ApplyMiddleware() error {
	// 应用全局中间件
	if m.Server == nil {
		return fmt.Errorf("hertz server is nil")
	}

	for _, mw := range m.Mws {
		m.Server.Use(app.HandlerFunc(mw))
	}

	for _, mw := range m.MwsHFunc {
		m.Server.Use(mw)
	}
	return nil
}
