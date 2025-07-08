package client

import (
	"ag-core/ag/ag_conf"
	"log/slog"
	"time"
)

const (
	KitexClientPropertiesPrefix = "kitex.client"
)

type KitexClientProperties struct {
	RpcTimeout    time.Duration `value:"${:30}"`
	TransportType string        `value:"${:grpc}"`

	Resolver ResolverProperties
}

type Grpc struct {
	Enable            bool `value:"${:false}"`
	MaxConnectionIdle int  `value:"${:0}"`
}

type ResolverProperties struct {
	Enable bool
	Type   string

	Nacos NacosGC `value:"${Nacos}"`
}

type NacosGC struct {
	Group   string `value:"${group:DEFAULT_GROUP}"`
	Cluster string `value:"${cluster:DEFAULT}"`
}

func FxInitKitexClientProperties(binder ag_conf.IBinder) *KitexClientProperties {
	p := &KitexClientProperties{}
	binder.Bind(p, KitexClientPropertiesPrefix)
	slog.Debug("KitexClientProperties", slog.Any("KitexClientProperties", p))
	return p
}
