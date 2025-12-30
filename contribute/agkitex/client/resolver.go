package client

import (
	"log/slog"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/kitex-contrib/registry-nacos/resolver"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
)

// BuildKitexResolver 构建服务发现解析器
func BuildKitexResolver(config *KitexClientConfig, namingClient naming_client.INamingClient) discovery.Resolver {
	if !config.Resolver.Enable {
		slog.Debug("Resolver is disabled")
		return nil
	}

	if namingClient == nil {
		slog.Warn("Naming client is nil, resolver will not be created")
		return nil
	}

	switch config.Resolver.Type {
	case "agnacos":
		slog.Debug("Creating Ag Nacos resolver",
			slog.String("group", config.Resolver.Nacos.Group),
			slog.String("cluster", config.Resolver.Nacos.Cluster))

		return NewAgNacosResolver(namingClient,
			WithGroup(config.Resolver.Nacos.Group),
			WithCluster(config.Resolver.Nacos.Cluster))
	case "nacos":
		slog.Debug("Creating Nacos resolver",
			slog.String("group", config.Resolver.Nacos.Group),
			slog.String("cluster", config.Resolver.Nacos.Cluster))

		return resolver.NewNacosResolver(namingClient,
			resolver.WithGroup(config.Resolver.Nacos.Group),
			resolver.WithCluster(config.Resolver.Nacos.Cluster))

	default:
		slog.Warn("Unknown resolver type", slog.String("type", config.Resolver.Type))
		return nil
	}
}
