package naming

import "ag-core/ag/ag_nacos/common"

const (
	// NacosNamingPropertiesPrefix nacos naming properties prefix
	NacosNamingPropertiesPrefix string = "nacos.naming"
)

// NacosNamingProperties nacos naming properties
type NacosNamingProperties struct {
	Enable bool `value:"${enable:true}"`

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
