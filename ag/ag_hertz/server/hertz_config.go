package server

import "ag-core/ag/ag_conf"

const (
	hertzServerPropertiesPrefix = "hertz.server"
	DefaultHertzOriginPort      = 7000
)

type HertzServerProperties struct {
	Host          string            `value:"${host:0.0.0.0}"`
	Port          int               `value:"${port:0}"`
	AdaptivePort  bool              `value:"${adaptive-port:false}"`
	ServiceName   string            `value:"${service-name:}"`
	Cluster       string            `value:"${cluster:DEFAULT}"`
	Group         string            `value:"${group:DEFAULT_GROUP}"`
	EnableIPRange string            `value:"${enable-ip-range:}"`
	EnableH2C     bool              `value:"${enable-h2c:false}"`
	Tags          map[string]string `value:"${tags:}"`
}

func NewHertzServerProperties(binder ag_conf.IBinder) (*HertzServerProperties, error) {
	p := &HertzServerProperties{}
	err := binder.Bind(p, hertzServerPropertiesPrefix)
	return p, err
}
