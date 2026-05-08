package server

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_ext/ip"
	"ag-core/contribute/aghertz/server/consts"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
)

// HertzServerProperties is the properties of the Hertz server.
type HertzServerProperties struct {
	Host          string `value:"${host}"`            // Host to bind the Hertz server to.
	Port          int    `value:"${port}"`            // Port to bind the Hertz server to.
	AdaptivePort  bool   `value:"${adaptive-port}"`   // Whether to enable adaptive port.
	EnableIPRange string `value:"${enable-ip-range}"` // Whether to enable IP range.

	KeepAlive        bool          `value:"${keep-alive}"`         // Whether to enable keep alive.
	KeepAliveTimeout time.Duration `value:"${keep-alive-timeout}"` // Keep alive timeout in seconds, default 60s.
	Pprof            bool          `value:"${pprof}"`              // Whether to enable pprof.
	PprofPath        string        `value:"${pprof-path}"`         // Pprof path.
	EnableH2C        bool          `value:"${enable-h2c}"`         // Whether to enable H2C.

	// Service info
	ServiceName string            `value:"${service-name}"` // Service name.
	Cluster     string            `value:"${cluster}"`      // Cluster name.
	Group       string            `value:"${group}"`        // Group name.
	Tags        map[string]string `value:"${tags}"`         // Tags.
}

// NewHertzServerProperties creates and bind a new HertzServerProperties.
func NewHertzServerProperties(binder ag_conf.IBinder) (*HertzServerProperties, error) {
	p := defaultHertzServerProperties()
	err := binder.Bind(p, consts.HertzServerPropertiesPrefix)
	return p, err
}

func findHertzHostPort(hconf *HertzServerProperties) (host string, port int, rerr error) {
	// 服务ip、端口配置
	host = hconf.Host
	if host == "" {
		host = "0.0.0.0"
	}

	if !ip.IsHostAvailable(host) {
		return "", 0, fmt.Errorf("hertz host unavailable: %s", host)
	}

	port = hconf.Port
	if hconf.AdaptivePort {
		slog.Info("hertz server enable adaptive port")
		if port == 0 {
			port = consts.DefaultHertzOriginPort
		}
		port, rerr = ip.GetAvailablePort(host, port)
		if rerr != nil {
			return
		}
	} else {
		if port == 0 {
			return host, port, fmt.Errorf("hertz port invalid:%v", port)
		}
	}

	slog.Info(fmt.Sprintf("found hertz host:%s, port:%d", host, port))
	return
}

func buildHertzRegInfo(props *HertzServerProperties, port int) (*registry.Info, error) {
	regInfo := &registry.Info{}
	regInfo.Weight = 1

	ipranger, err := ip.NewIPRanger(props.EnableIPRange)
	if err != nil {
		return nil, err
	}

	host, ok, err := ipranger.GetLocalIP()
	if err != nil {
		return nil, err
	}
	if ok {
		slog.Info("hertz server enable ip range", "regAddr", fmt.Sprintf("%s:%d", host, port))
		regInfo.Addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			return nil, err
		}
	}

	sname := props.ServiceName
	if sname == "" {
		sname = "hertz-server"
	}
	regInfo.ServiceName = sname

	// 服务元信息配置，可在配置中配置，兼容并行阶段的spring-grpc网关调用
	tags := make(map[string]string)
	if props.Tags != nil {
		tags = props.Tags
	}
	tags["ag_core"] = "All rights reserved"
	tags["lang_type"] = "Golang"
	regInfo.Tags = tags

	return regInfo, nil
}

func defaultHertzServerProperties() *HertzServerProperties {
	return &HertzServerProperties{
		Host:          "0.0.0.0",
		Port:          consts.DefaultHertzOriginPort,
		AdaptivePort:  false,
		EnableIPRange: "",

		KeepAlive:        true,
		KeepAliveTimeout: 60,
		Pprof:            false,
		PprofPath:        consts.DefaultPprofPath,
		EnableH2C:        false,

		ServiceName: "hertz-server",
	}
}
