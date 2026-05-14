package common

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

type SCProperties struct {
	// server
	Schema      string
	ContextPath string

	// client
	ServerAddr string
	NameSpace  string
	UserName   string
	Password   string
	LogLevel   string
}

func BuildServerConfig(p SCProperties) ([]constant.ServerConfig, error) {
	adds := p.ServerAddr
	if adds == "" {
		return nil, fmt.Errorf("nacos server addr is empty")
	}
	// ipports, err := ag_nacos.parseIPPort(adds)
	ipports, err := ParseIPPort(adds)
	if err != nil {
		return nil, err
	}

	// schema
	schema := p.Schema
	if schema == "" {
		schema = "http"
	}

	// contextPath
	contextPath := p.ContextPath
	if contextPath == "" {
		contextPath = "/nacos"
	}

	opts := []constant.ServerOption{}
	if schema != "" {
		opts = append(opts, constant.WithScheme(schema))
	}
	if contextPath != "" {
		opts = append(opts, constant.WithContextPath(contextPath))
	}

	sc := []constant.ServerConfig{}

	for _, ipport := range ipports {
		sc = append(sc, *constant.NewServerConfig(ipport.Ip, ipport.Port, opts...))
	}

	return sc, nil
}

// NewNacosClientConfig 初始化nacos client配置
// func namingClientConfig(p *NacosNamingProperties) (*constant.ClientConfig, error) {
func BuildClientConfig(p SCProperties) (*constant.ClientConfig, error) {
	namespace := p.NameSpace
	username := p.UserName
	password := p.Password

	opts := []constant.ClientOption{}

	// namespace
	if namespace != "" {
		opts = append(opts, constant.WithNamespaceId(namespace))
	}
	// username
	if username != "" {
		opts = append(opts, constant.WithUsername(username))
	}
	// password
	if password != "" {
		opts = append(opts, constant.WithPassword(password))
	}

	if p.LogLevel != "" {
		opts = append(opts, constant.WithLogLevel(p.LogLevel))
	}

	// TODO 其他配置

	clientConfig := constant.NewClientConfig(opts...)

	return clientConfig, nil
}

func DefaultSCProperties() SCProperties {
	return SCProperties{
		Schema:      "http",
		ContextPath: "/nacos",
		// ServerAddr:  "127.0.0.1:8848",
		// NameSpace:   "public",
		// UserName:    "nacos",
		// Password:    "nacos",
		LogLevel: "error",
	}
}
