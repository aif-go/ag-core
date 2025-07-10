package ag_nacos

const (
	// NacosConfigPropertiesPrefix nacos config properties prefix
	NacosConfigPropertiesPrefix string = "nacos.config"
)

// NacosConfigProperties nacos config properties
type NacosConfigProperties struct {
	EnableConfig bool `value:"${enableconfig:true}"`
	EnableNaming bool `value:"${enablenaming:true}"`
	// server
	Schema      string `value:"${schema:http}"`
	ContextPath string `value:"${contextpath:/nacos}"`
	ServerAddr  string `value:"${serveraddr}"`

	// client
	NameSpace string `value:"${namespace}"`
	UserName  string `value:"${username}"`
	Password  string `value:"${password}"`

	DataIDs []DataIDInfo `value:"${dataids}"`
}

// DataIDInfo nacso dataid相关的配置
type DataIDInfo struct {
	DataID string `value:"${dataid}"` //required
	Group  string `value:"${group}"`  //required
	Type   string `value:"${type}"`   //required

	AutoRefresh bool `value:"${autorefresh:true}"`
	// First       bool `value:"${first:false}"`
	// Before      bool `value:"${before:false}"`
	// After       bool `value:"${after:false}"`
}
