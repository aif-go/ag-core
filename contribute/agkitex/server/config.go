package server

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
	"github.com/aif-go/ag-core/ag/ag_ext/ip"
	"fmt"
	"log/slog"
)

const (
	// KitexServerPropertiesPrefix 服务器配置前缀
	KitexServerPropertiesPrefix = "kitex.server"
	DefaultKitexOriginPort      = 7000
)

type (
	// KitexServerProperties Kitex服务器配置属性
	KitexServerProperties struct {
		Host          string `value:"${:}"`
		Port          int    `value:"${:7000}"`
		AdaptivePort  bool   `value:"${:false}"`
		ServiceName   string
		EnableIPRange string `value:"${:}"`

		Grpc Grpc
	}

	// Grpc grpc配置
	Grpc struct {
		Enable            bool `value:"${:true}"` // 默认使用grpc
		MaxConnectionIdle int  `value:"${:0}"`
	}
)

// NewKitexServerProperties 创建Kitex服务器配置属性
func NewKitexServerProperties(binder ag_conf.IBinder) *KitexServerProperties {
	props := &KitexServerProperties{}
	binder.Bind(props, KitexServerPropertiesPrefix)
	return props
}

func findKitexHostPort(kconf *KitexServerProperties) (host string, port int, rerr error) {
	// 服务ip、端口配置
	host = kconf.Host
	if host == "" {
		host = "0.0.0.0"
	}

	if !ip.IsHostAvailable(host) {
		return "", 0, fmt.Errorf("kitex host unavailable: %s", host)
	}

	port = kconf.Port
	if kconf.AdaptivePort {
		slog.Info("kitex server enable adaptive port")
		if port == 0 {
			port = DefaultKitexOriginPort
		}
		port, rerr = ip.GetAvailablePort(host, port)
		if rerr != nil {
			return
		}
	} else {
		if port == 0 {
			return host, port, fmt.Errorf("kitex port invalid:%v", port)
		}
	}

	slog.Info(fmt.Sprintf("finded kitex host:%s, port:%d", host, port))
	return
}
