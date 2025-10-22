package server

import (
	"github.com/cloudwego/hertz/pkg/app/server"
)

// ServerOption 是一个选项，用于配置 增强 Hertz 服务器。
type ServerOption struct {
	F func(h *server.Hertz) error
}

// ServerSuite 是一个选项集合，用于配置 增强 Hertz 服务器。
type ServerSuite interface {
	Options() []*ServerOption
}

// OptionHertzServer 将多个 ServerOption 应用到 Hertz 服务器配置中。
func OptionHertzServer(hertz *server.Hertz, opts ...ServerOption) error {
	for _, opt := range opts {
		err := opt.F(hertz)
		if err != nil {
			return err
		}
	}
	return nil
}

// OptionHertzServerSuite 将服务器设置应用到 Hertz 服务器配置中。
func OptionHertzServerSuite(hertz *server.Hertz, suite ServerSuite) error {
	return OptionHertzServer(
		hertz,
		WithServerSuite(suite),
	)
}

// SimpleServerSuite 是一个简单的 ServerSuite 实现。
type SimpleServerSuite struct {
	opts []*ServerOption
}

// Options 返回 SimpleServerSuite 中的所有 ServerOption。
func (s *SimpleServerSuite) Options() []*ServerOption {
	return s.opts
}

// Add 添加多个 ServerOption 到 Suite 中。
func (s *SimpleServerSuite) Add(opts ...ServerOption) {
	opt := make([]*ServerOption, len(opts))
	for i, o := range opts {
		opt[i] = &o
	}
	s.opts = append(s.opts, opt...)
}

// AddPtr 添加多个 ServerOption 到 Suite 中。
func (s *SimpleServerSuite) AddPtr(opts ...*ServerOption) {
	s.opts = append(s.opts, opts...)
}
