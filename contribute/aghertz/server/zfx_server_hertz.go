package server

import (
	"ag-core/ag/ag_server"
	"ag-core/contribute/aghertz/server/registy"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/config"
	"go.uber.org/fx"
)

// FxAgHertzServerModule 创建HTTP服务，并注册到注册中心
var FxAgHertzServerModule = fx.Module("fx_aghertz_server",
	// 注册中心
	registy.FxHertzRegistyModule,

	fx.Provide(
		// 配置文件
		NewHertzServerProperties,

		// 注入构建HertzServerParam
		FxNewHertzServerConfigParam,

		// Hertz 服务配置选项
		NewFxServerConfigOptionsProvider(BuildHertzServerConfigOption),

		// Hertz 服务配置套件，注入聚合所有配置选项
		FxNewHertsServerConfigSuite,

		// 原始Hertz服务
		NewHertzServerWithSuite,

		// Ag Hertz服务，包装原始Hertz服务
		NewAGHertzServer,

		// Hertz 服务选项
		NewFxServerOptionsProvider(BuildHertzServerOptions),

		// Hertz 服务选项
		// FxNewHertzServerSuite,

		// Hertz 路由选项
		// FxNewHertzRouteOptions,
		FxNewServerConfiguratorParam,
		NewServerConfiguratorWithParam,

		fx.Annotate(
			hertzServerWrapper,                  // Hertz服务包装器，将AgHertzServer转换为ag_server.Server
			fx.ResultTags(`group:"ag_servers"`), // 标记为ag_servers组，被ag-app发现以启动
		),
	),
	fx.Invoke(
		func(sc *ServerConfigurator) error {
			return sc.InitHertzServer()
		},
	),
)

func hertzServerWrapper(s *AgHertzServer) ag_server.Server {
	return s
}

// FxInHertzServerConfigSuiteParam HertzServerSuiteParam 内容注入器
type FxInHertzServerConfigSuiteParam struct {
	fx.In
	Opts []*config.Option `group:"aghertz_server_config_options" ,optional:"true"`
}

// FxNewHertsServerConfigSuite 通过注入器组装Hertz服务配置套件
func FxNewHertsServerConfigSuite(fp FxInHertzServerConfigSuiteParam) ConfigSuite {
	suite := &SimpleSuite{}
	if fp.Opts == nil {
		return suite
	}
	suite.AddPtr(fp.Opts...)
	return suite
}

// FxNewHertzServerParam HertzServerParam 内容注入器
type FxInHertzServerConfigParam struct {
	fx.In
	HertzServerProperties *HertzServerProperties
	Register              registry.Registry `optional:"true"`
}

// FxNewHertzServerConfigParam 通过注入器构建 HertzServerParam
func FxNewHertzServerConfigParam(fp FxInHertzServerConfigParam) *HertzServerConfigParam {
	return &HertzServerConfigParam{
		HertzServerProperties: fp.HertzServerProperties,
		Register:              fp.Register,
	}
}

type FxInServerConfiguratorParam struct {
	fx.In

	Server *server.Hertz
	Opts   []*ServerOption `group:"aghertz_server_options" ,optional:"true"`
	Routes []*Route        `group:"aghertz_route" ,optional:"true"`
	Mws    []Middleware    `group:"aghertz_middleware" ,optional:"true"`
}

func FxNewServerConfiguratorParam(fp FxInServerConfiguratorParam) *ServerConfiguratorParam {
	return &ServerConfiguratorParam{
		Server: fp.Server,
		Opts:   fp.Opts,
		Routes: fp.Routes,
		Mws:    fp.Mws,
	}
}

// NewFxServerOptionsProvider 为 HertzServerOptions 提供 FX 选项
func NewFxServerOptionsProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"aghertz_server_options"`),
	)
}

// NewFxServerConfigOptionsProvider 为 HertzServerConfigOptions 提供 FX 选项
func NewFxServerConfigOptionsProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"aghertz_server_config_options"`),
	)
}

// NewFxServerRouteProvider 为 Route 提供 FX 选项
func NewFxServerRouteProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"aghertz_route"`),
	)
}

func NewFxServerMiddlewareProvider(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"aghertz_middleware"`),
	)
}
