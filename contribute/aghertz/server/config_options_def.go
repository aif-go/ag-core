package server

import (
	"github.com/cloudwego/hertz/pkg/common/config"
)

// ConfigSuite 是一个选项组，用于配置 Hertz 服务器。
type ConfigSuite interface {
	Options() []*config.Option
}

// WithConfigSuite 将 Suite 中的选项应用到 Hertz 服务器配置中。
func WithConfigSuite(suite ConfigSuite) config.Option {
	return config.Option{F: func(o *config.Options) {
		for _, op := range suite.Options() {
			op.F(o)
		}
	}}
}

type SimpleSuite struct {
	opts []*config.Option
}

func (s *SimpleSuite) Options() []*config.Option {
	return s.opts
}

func (s *SimpleSuite) Add(opts ...config.Option) {
	opt := make([]*config.Option, len(opts))
	for i, o := range opts {
		opt[i] = &o
	}
	s.opts = append(s.opts, opt...)
}

func (s *SimpleSuite) AddPtr(opts ...*config.Option) {
	s.opts = append(s.opts, opts...)
}
