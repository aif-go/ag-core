package server

import (
	"ag-core/contribute/aghertz/server/consts"
	"fmt"
	"log/slog"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/config"
)

// NewHertzServerWithSuite creates a new Hertz server.
func NewHertzServerWithSuite(suite ConfigSuite) *server.Hertz {
	return NewHertzServer(WithConfigSuite(suite))
}

// NewHertzServer creates a new Hertz server.
func NewHertzServer(opts ...config.Option) *server.Hertz {
	hertz := server.Default(opts...)
	return hertz
}

// HertzServerConfigParam Hertz 服务器参数
type HertzServerConfigParam struct {
	HertzServerProperties *HertzServerProperties
	Register              registry.Registry `optional:"true"`
}

// BuildHertzServerConfigOption 根据配置构建 Hertz 服务器配置选项。
func BuildHertzServerConfigOption(param *HertzServerConfigParam) (*config.Option, error) {

	props := param.HertzServerProperties
	opts := &SimpleSuite{}

	// host port
	host, port, rerr := findHertzHostPort(props)
	if rerr != nil {
		return nil, rerr
	}
	hertzHostStr := fmt.Sprintf("%s:%d", host, port)
	slog.Info("hertz", "host", hertzHostStr)
	opts.Add(server.WithHostPorts(hertzHostStr))

	// keep alive
	opts.Add(server.WithKeepAlive(props.KeepAlive))
	opts.Add(server.WithKeepAliveTimeout(props.KeepAliveTimeout * time.Second))

	// registry naming
	if param.Register != nil {
		regInfo, err := buildHertzRegInfo(props, port)
		if err != nil {
			return nil, err
		}
		opts.Add(server.WithRegistry(param.Register, regInfo))
	}

	copt := WithConfigSuite(opts) // 将所有配置最后包装成config.Option，由注入生效
	return &copt, nil
}

// BuildHertzServerOptions 根据配置构建 Hertz 服务器选项。
func BuildHertzServerOptions(param *HertzServerConfigParam) (*ServerOption, error) {
	props := param.HertzServerProperties
	suite := &SimpleServerSuite{}

	// pprof
	if props.Pprof {
		pprofpath := props.PprofPath
		if pprofpath == "" {
			pprofpath = consts.DefaultPprofPath
		}
		suite.Add(WithPprof(pprofpath))
	}

	// TODO 其他能力

	opt := WithServerSuite(suite)
	return &opt, nil
}
