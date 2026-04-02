package agonet

import (
	"crypto/tls"
	"net"
	"time"

	// "github.com/tjfoc/gmsm/gmtls"

	"gitee.com/Trisia/gotlcp/tlcp"
)

// Option is a function that will set up option.
type Option func(opts *Options) error

// func loadOptions(options ...Option) *Options {
// 	opts := new(Options)
// 	for _, option := range options {
// 		option(opts)
// 	}
// 	return opts
// }

// ExtendOptions extends options with given options.
func ExtendOptions(opts *Options, options ...Option) error {
	for _, option := range options {
		if err := option(opts); err != nil {
			return err
		}
	}
	return nil
}

// Options are configurations for the gnet application.
type Options struct {
	Multicore bool

	NumEventLoop int

	LockOSThread bool

	// Ticker bool

	KeepAlive KeepAlive

	TLSType    TLSType
	TLSConfig  *tls.Config
	TLCPConfig *tlcp.Config

	CLI_TLSType    TLSType
	CLI_TLSConfig  *tls.Config
	CLI_TLCPConfig *tlcp.Config

	// TLCPConfig *gmtls.Config

	// TLS  tlsConfig
	// TLCP tlcpConfig
}

func (opt *Options) CliTLSType() TLSType {
	cliTlsType := opt.CLI_TLSType
	if cliTlsType == TLSType_UNSET {
		cliTlsType = opt.TLSType
	}
	return cliTlsType
}

func (opt *Options) CliTLSConfig() *tls.Config {
	if opt.CLI_TLSConfig != nil {
		return opt.CLI_TLSConfig
	}
	return opt.TLSConfig
}
func (opt *Options) CliTLCPConfig() *tlcp.Config {
	if opt.CLI_TLCPConfig != nil {
		return opt.CLI_TLCPConfig
	}
	return opt.TLCPConfig
}

type KeepAlive struct {
	Enable   bool
	Idle     time.Duration
	Interval time.Duration
	Count    int
}

// BuildOptionsWithConfig builds options with given config.
func BuildOptionsWithConfig(conf OptionsConfig) (*Options, error) {
	opts := &Options{
		NumEventLoop: conf.Engine.NumEventLoop,
		Multicore:    conf.Engine.Multicore,
		// Ticker:       conf.Engine.Ticker,
		KeepAlive: KeepAlive{
			Enable:   conf.KeepAlive.Enable,
			Idle:     time.Duration(conf.KeepAlive.Idle) * time.Second,
			Interval: time.Duration(conf.KeepAlive.Interval) * time.Second,
			Count:    conf.KeepAlive.Count,
		},
	}

	return opts, nil
}

// buildKeepAliveWithConfig builds keep-alive config with given config.
func buildKeepAliveWithConfig(cnf KeepAlive) *net.KeepAliveConfig {
	if !cnf.Enable || cnf.Idle <= 0 {
		return nil
	}

	idle := cnf.Idle
	interval := cnf.Interval
	if interval <= 0 {
		interval = idle / 5 // 和count配合5 次检测，一个Idel周期内keep失败则认为连接已断开
	}
	count := cnf.Count
	if count <= 0 {
		count = 5
	}

	keepAliveConfig := &net.KeepAliveConfig{
		Enable:   true,
		Idle:     idle,
		Interval: interval,
		Count:    count,
	}
	return keepAliveConfig
}

// WithOptions sets up all options.
func WithOptions(options Options) Option {
	return func(opts *Options) error {
		*opts = options
		return nil
	}
}
