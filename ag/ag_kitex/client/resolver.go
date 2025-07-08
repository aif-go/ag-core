package client

import (
	"fmt"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/kitex-contrib/registry-nacos/resolver"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"go.uber.org/fx"
)

type FxInKitexResolverParams struct {
	fx.In
	NamingClient naming_client.INamingClient `optional:"true"`
}

func BuildKitexResolver(params FxInKitexResolverParams, cpri *KitexClientProperties) (discovery.Resolver, error) {
	enable := cpri.Resolver.Enable
	if !enable {
		return nil, nil
	}

	rtype := cpri.Resolver.Type

	switch rtype {
	case "agnacos":
		return buildAgNacosResolver(params, cpri)
	case "nacos":
		return buildNacosResolver(params, cpri)
	default:
		return nil, fmt.Errorf("unsupported resolver type: %s", rtype)
	}

}

func buildAgNacosResolver(params FxInKitexResolverParams, cpri *KitexClientProperties) (discovery.Resolver, error) {
	if params.NamingClient == nil {
		return nil, fmt.Errorf("nacos naming client is nil")
	}
	opts := make([]Option, 0)

	group := cpri.Resolver.Nacos.Group
	cluster := cpri.Resolver.Nacos.Cluster

	if group != "" {
		opts = append(opts, WithGroup(group))
	}

	if cluster != "" {
		opts = append(opts, WithCluster(cluster))
	}

	return NewAgNacosResolver(params.NamingClient, opts...), nil
}

func buildNacosResolver(params FxInKitexResolverParams, cpri *KitexClientProperties) (discovery.Resolver, error) {
	if params.NamingClient == nil {
		return nil, fmt.Errorf("nacos naming client is nil")
	}
	opts := make([]resolver.Option, 0)

	group := cpri.Resolver.Nacos.Group
	cluster := cpri.Resolver.Nacos.Cluster

	if group != "" {
		opts = append(opts, resolver.WithGroup(group))
	}

	if cluster != "" {
		opts = append(opts, resolver.WithCluster(cluster))
	}

	return resolver.NewNacosResolver(params.NamingClient, opts...), nil
}
