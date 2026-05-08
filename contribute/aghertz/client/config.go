package client

import (
	"ag-core/ag/ag_conf"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/config"
)

const (
	HertzClientPropertiesPrefix = "hertz.client"
)

type (
	HertzClientProperties struct {
		KeepAlive bool

		// Timeout for establishing a connection to server. unit: ms
		DialTimeout int

		// The max connection nums for each host
		MaxConnsPerHost int

		// The max duration before idle keep-alive connection closed. unit: ms
		MaxIdleConnDuration int
	}
)

func NewClientProperties(binder ag_conf.IBinder) (*HertzClientProperties, error) {
	props := defaultClientProperties()
	err := binder.Bind(props, HertzClientPropertiesPrefix)
	if err != nil {
		return nil, err
	}
	return props, nil
}

// BuildClientOptionWithConfig builds a client option with the given properties.
func BuildClientOptionWithConfig(props *HertzClientProperties) *config.ClientOption {
	optSuite := &SimpleClientSuite{}

	optSuite.AddOptions(
		// keep-alive
		client.WithKeepAlive(props.KeepAlive),
		// dial timeout
		client.WithDialTimeout(time.Duration(props.DialTimeout)*time.Millisecond),
		// max conns per host
		client.WithMaxConnsPerHost(props.MaxConnsPerHost),
		// max idle conn duration
		client.WithMaxIdleConnDuration(time.Duration(props.MaxIdleConnDuration)*time.Millisecond),
	)

	opt := WithClientSuite(optSuite)
	return &opt
}

func BuildMiddlewareOptionWithConfig(props *HertzClientProperties) PrioritizedClientMiddlewareSuite {
	optSuite := &SimplePrioritizedClientMiddlewareSuite{}

	// optSuite.AddMiddleware(
	// 	sd.Discovery(res),
	// )

	return optSuite
}

func defaultClientProperties() *HertzClientProperties {
	return &HertzClientProperties{
		KeepAlive:           false,
		DialTimeout:         1000, // 1s
		MaxConnsPerHost:     512,
		MaxIdleConnDuration: 10000, // 10s
	}
}
