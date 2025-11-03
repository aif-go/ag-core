package nacos

import (
	"ag-core/ag/ag_conf"
	"ag-core/contribute/aghertz/client"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	rnacos "github.com/hertz-contrib/registry/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
)

const (
	HertzClientPropertiesPrefix = client.HertzClientPropertiesPrefix + ".discovery"
)

type (
	Param struct {
		NamingClient naming_client.INamingClient
	}

	Properties struct {
		Enabled bool            `value:"${enabled:true}"`
		Nacos   NacosProperties `value:"${nacos}"`
	}

	NacosProperties struct {
		Cluster string `value:"${cluster:DEFAULT}"`     // Cluster name.
		Group   string `value:"${group:DEFAULT_GROUP}"` // Group name.
	}
)

func NewProperties(binder ag_conf.IBinder) (*Properties, error) {
	p := &Properties{}
	err := binder.Bind(p, HertzClientPropertiesPrefix)
	return p, err
}

func NewResolver(param *Param, props *Properties) discovery.Resolver {
	if param.NamingClient != nil {
		return rnacos.NewNacosResolver(
			param.NamingClient,
			rnacos.WithResolverCluster(props.Nacos.Cluster),
			rnacos.WithResolverGroup(props.Nacos.Group),
		)
	}
	return nil
}
