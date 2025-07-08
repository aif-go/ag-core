/*
	对 github.com/kitex-contrib/registry-nacos/resolver.go 的增强
	添加对 spring-grpc 服务注册的元数据解析支持，spring-grpc将grpc的端口放在了metadata中
*/

package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/rpcinfo"

	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type options struct {
	cluster string
	group   string
}

// Option is nacos option.
type Option func(o *options)

// WithCluster with cluster option.
func WithCluster(cluster string) Option {
	return func(o *options) { o.cluster = cluster }
}

// WithGroup with group option.
func WithGroup(group string) Option {
	return func(o *options) { o.group = group }
}

type AgNacosResolver struct {
	cli  naming_client.INamingClient
	opts options
}

// NewDefaultNacosResolver create a default service resolver using nacos.
//func NewDefaultNacosResolver(opts ...Option) (discovery.Resolver, error) {
//	cli, err := nacos.NewDefaultNacosClient()
//	if err != nil {
//		return nil, err
//	}
//	return NewNacosResolver(cli, opts...), nil
//}

// NewAgNacosResolver create a service resolver using nacos.
func NewAgNacosResolver(cli naming_client.INamingClient, opts ...Option) discovery.Resolver {
	op := options{
		cluster: "DEFAULT",
		group:   "DEFAULT_GROUP",
	}
	for _, option := range opts {
		option(&op)
	}
	return &AgNacosResolver{cli: cli, opts: op}
}

// Target return a description for the given target that is suitable for being a key for cache.
func (n *AgNacosResolver) Target(_ context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

// Resolve a service info by desc.
func (n *AgNacosResolver) Resolve(_ context.Context, desc string) (discovery.Result, error) {
	res, err := n.cli.SelectInstances(vo.SelectInstancesParam{
		ServiceName: desc,
		HealthyOnly: true,
		GroupName:   n.opts.group,
		Clusters:    []string{n.opts.cluster},
	})
	if err != nil {
		return discovery.Result{}, err
	}
	if len(res) == 0 {
		return discovery.Result{}, fmt.Errorf("no instance remains for %v", desc)
	}
	instances := make([]discovery.Instance, 0, len(res))
	for _, in := range res {
		inst, ok := resolverInstance(in)
		if ok {
			instances = append(instances, inst)
		}
	}
	if len(instances) == 0 {
		return discovery.Result{}, fmt.Errorf("no instance remains for %v", desc)
	}
	return discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: instances,
	}, nil
}

// Diff computes the difference between two results.
func (n *AgNacosResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

// Name returns the name of the resolver.
func (n *AgNacosResolver) Name() string {
	return "nacos" + ":" + n.opts.cluster + ":" + n.opts.group
}

var _ discovery.Resolver = (*AgNacosResolver)(nil)

func resolverInstance(in model.Instance) (discovery.Instance, bool) {
	if !in.Enable {
		return nil, false
	}
	inIp := in.Ip
	inPort := in.Port

	// spring-grpc将grpc的端口放在了metadata中
	metadate := in.Metadata
	prs, ok := metadate["preserved.register.source"]
	// 兼容https://github.com/grpc-ecosystem/grpc-spring项目的服务注册实现
	if ok {
		if prs == "SPRING_CLOUD" {
			grpcPort, ok := metadate["gRPC_port"]
			if ok {
				p, err := strconv.ParseUint(grpcPort, 10, 64)
				if err == nil {
					inPort = p
				}
			}
		}
	}

	inst := discovery.NewInstance(
		"tcp",
		fmt.Sprintf("%s:%d", inIp, inPort),
		int(in.Weight),
		in.Metadata)

	return inst, true
}
