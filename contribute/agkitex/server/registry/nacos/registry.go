package nacos

import (
	"ag-core/ag/ag_conf"
	"log/slog"

	"github.com/cloudwego/kitex/pkg/registry"
	nacosreg "github.com/kitex-contrib/registry-nacos/registry"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
)

const (
	KitexServerRegistryPropertiesPrefix = "kitex.server.registry"
	DefaultCluster                      = "DEFAULT"
	DefaultGroup                        = "DEFAULT_GROUP"
)

type (
	Properties struct {
		Cluster string `value:"${cluster:DEFAULT}"`     // Cluster name.
		Group   string `value:"${group:DEFAULT_GROUP}"` // Group name.
	}
)

func NewProperties(binder ag_conf.IBinder) (*Properties, error) {
	p := &Properties{}
	err := binder.Bind(p, KitexServerRegistryPropertiesPrefix)
	return p, err
}

type Param struct {
	NamingClient naming_client.INamingClient
}

func NewRegisty(param *Param, props *Properties) registry.Registry {

	group := props.Group
	if group == "" {
		group = DefaultGroup
	}
	cluster := props.Cluster
	if cluster == "" {
		cluster = DefaultCluster
	}

	slog.Info("create kitex nacos registry", "group", group, "cluster", cluster)
	if param.NamingClient != nil {
		return nacosreg.NewNacosRegistry(
			param.NamingClient,
			nacosreg.WithGroup(group),
			nacosreg.WithCluster(cluster),
		)
	}
	return nil
}
