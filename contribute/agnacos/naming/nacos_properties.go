package naming

import "ag-core/contribute/agnacos/common"

const (
	// NacosNamingPropertiesPrefix nacos naming properties prefix
	NacosNamingPropertiesPrefix string = "nacos.naming"
)

// NacosNamingProperties nacos naming properties
type NacosNamingProperties struct {
	Enable bool

	common.SCProperties
	// // server
	// Schema      string `value:"${schema:http}"`
	// ContextPath string `value:"${contextpath:/nacos}"`

	// // client
	// ServerAddr string `value:"${serveraddr}"`
	// NameSpace  string `value:"${namespace}"`
	// UserName   string `value:"${username}"`
	// Password   string `value:"${password}"`
}

func defaultNacosNamingProperties() *NacosNamingProperties {
	return &NacosNamingProperties{
		Enable:       true,
		SCProperties: common.DefaultSCProperties(),
	}
}
