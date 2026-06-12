package nacos

import (
	"github.com/aif-go/ag-core/ag/ag_conf"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	rnacos "github.com/hertz-contrib/registry/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
)

const (
	// HertzClientPropertiesPrefix = client.HertzClientPropertiesPrefix + ".discovery"
	HertzClientPropertiesPrefix = "hertz.client" + ".discovery"
)

type (
	Param struct {
		NamingClient naming_client.INamingClient
	}

	Properties struct {
		Enabled bool
		Type    string
		Nacos   NacosProperties
	}

	NacosProperties struct {
		Cluster string // Cluster name.
		Group   string // Group name.
	}
)

func NewProperties(binder ag_conf.IBinder) (*Properties, error) {
	p := defaultProperties()
	err := binder.Bind(p, HertzClientPropertiesPrefix)
	return p, err
}

func NewResolver(param *Param, props *Properties) discovery.Resolver {
	if !props.Enabled {
		return nil
	}

	if props.Type != "nacos" {
		return nil
	}

	if param.NamingClient != nil {
		return rnacos.NewNacosResolver(
			param.NamingClient,
			rnacos.WithResolverCluster(props.Nacos.Cluster),
			rnacos.WithResolverGroup(props.Nacos.Group),
		)
	}
	return nil
}

func defaultProperties() *Properties {
	return &Properties{
		Enabled: true,
		Type:    "nacos",
		Nacos: NacosProperties{
			Cluster: "DEFAULT",
			Group:   "DEFAULT_GROUP",
		},
	}
}
