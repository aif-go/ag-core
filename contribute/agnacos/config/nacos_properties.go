package config

import "github.com/aif-go/ag-core/contribute/agnacos/common"

const (
	// NacosConfigPropertiesPrefix nacos config properties prefix
	NacosConfigPropertiesPrefix string = "nacos.config"
)

// NacosConfigProperties nacos config properties
type NacosConfigProperties struct {
	Enable bool

	common.SCProperties

	DataIDs []DataIDInfo
}

// DataIDInfo nacso dataid相关的配置
type DataIDInfo struct {
	DataID string `required:"true"`                           //required
	Group  string `value:"${:DEFAULT_GROUP}" required:"true"` //required
	Type   string `value:"${:yaml}" required:"true"`          //required

	AutoRefresh bool `value:"${:true}"`
}

func defaultNacosConfigProperties() *NacosConfigProperties {
	return &NacosConfigProperties{
		Enable:       true,
		SCProperties: common.DefaultSCProperties(),
	}
}
