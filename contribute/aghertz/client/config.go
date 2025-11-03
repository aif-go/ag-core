package client

import (
	"time"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/common/config"
)

const (
	HertzClientPropertiesPrefix = "hertz.client"
)

type (
	HertzClientProperties struct {
		KeepAlive bool `value:"${keepAlive:true}"`

		// Timeout for establishing a connection to server. unit: ms
		DialTimeout int `value:"${dialTimeout:1000}"`

		// The max connection nums for each host
		MaxConnsPerHost int `value:"${maxConnsPerHost:512}"`

		// The max duration before idle keep-alive connection closed. unit: ms
		MaxIdleConnDuration int `value:"${maxIdleConnDuration:10000}"`

		// FIXME other options
		Discovery DiscoveryProperties `value:"${discovery}"`
	}

	DiscoveryProperties struct {
		Enabled bool `value:"${enabled:true}"`
		// CustomizedAddrs is the customized addrs for service discovery.
		CustomizedAddrs []string `value:"${:}"`
	}
)

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
