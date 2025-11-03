package config

import "ag-core/contribute/agnacos/common"

const (
	// NacosConfigPropertiesPrefix nacos config properties prefix
	NacosConfigPropertiesPrefix string = "nacos.config"
)

// NacosConfigProperties nacos config properties
type NacosConfigProperties struct {
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

	DataIDs []DataIDInfo `value:"${dataids}"`
}

// DataIDInfo nacso dataid相关的配置
type DataIDInfo struct {
	DataID string `value:"${dataid}"`              //required
	Group  string `value:"${group:DEFAULT_GROUP}"` //required
	Type   string `value:"${type:yaml}"`           //required

	AutoRefresh bool `value:"${autorefresh:true}"`
}
