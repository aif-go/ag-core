package nacos

import (
	"github.com/aif-go/ag-core/ag/ag_conf"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	rnacos "github.com/hertz-contrib/registry/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
)

const (
	HertzServerRegistryPropertiesPrefix = "hertz.server.registry"
)

type Properties struct {
	Cluster string `value:"${cluster:DEFAULT}"`     // Cluster name.
	Group   string `value:"${group:DEFAULT_GROUP}"` // Group name.
}

func NewProperties(binder ag_conf.IBinder) (*Properties, error) {
	p := &Properties{}
	err := binder.Bind(p, HertzServerRegistryPropertiesPrefix)
	return p, err
}

type Param struct {
	NamingClient naming_client.INamingClient
}

func NewRegisty(param *Param, props *Properties) registry.Registry {
	if param.NamingClient != nil {
		return rnacos.NewNacosRegistry(
			param.NamingClient,
			rnacos.WithRegistryCluster(props.Cluster),
			rnacos.WithRegistryGroup(props.Group),
		)
	}
	return nil
}
