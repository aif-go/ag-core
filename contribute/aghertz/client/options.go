package client

import (
	"github.com/cloudwego/hertz/pkg/common/config"
)

type (
	// ClientSuite is a suite of client options.
	ClientSuite interface {
		Options() []*config.ClientOption
	}

	// SimpleClientSuite is a simple implementation of ClientSuite.
	SimpleClientSuite struct {
		opts []*config.ClientOption
	}
)

// Options implements ClientSuite.
func (s *SimpleClientSuite) Options() []*config.ClientOption {
	return s.opts
}
func (s *SimpleClientSuite) AddOptionsPtr(opts ...*config.ClientOption) {
	s.opts = append(s.opts, opts...)
}
func (s *SimpleClientSuite) AddOptions(opts ...config.ClientOption) {
	optPtrs := make([]*config.ClientOption, 0, len(opts))
	for _, opt := range opts {
		optPtrs = append(optPtrs, &opt)
	}
	s.AddOptionsPtr(optPtrs...)
}

// WithClientSuite applies the client options suite to the client options.
func WithClientSuite(suite ClientSuite) config.ClientOption {
	copt := config.ClientOption{
		F: func(o *config.ClientOptions) {
			opts := []config.ClientOption{}
			for _, opt := range suite.Options() {
				opts = append(opts, *opt)
			}
			o.Apply(opts)
			// o.Apply(suite.Options())
		},
	}
	return copt
}
