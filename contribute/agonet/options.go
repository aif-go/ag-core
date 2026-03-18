package agonet

import (
	"crypto/tls"
	"net"
	"time"

	// "github.com/tjfoc/gmsm/gmtls"

	"gitee.com/Trisia/gotlcp/tlcp"
)

// // Option is a function that will set up option.
// type Option func(opts *Options)

// func loadOptions(options ...Option) *Options {
// 	opts := new(Options)
// 	for _, option := range options {
// 		option(opts)
// 	}
// 	return opts
// }

// Options are configurations for the gnet application.
type Options struct {
	Multicore bool

	NumEventLoop int

	// LockOSThread bool

	// Ticker bool

	KeepAlive KeepAlive

	TLSType TLSType

	TLSConfig  *tls.Config
	TLCPConfig *tlcp.Config
	// TLCPConfig *gmtls.Config

	// TLS  tlsConfig
	// TLCP tlcpConfig
}

type KeepAlive struct {
	Enable   bool
	Idle     time.Duration
	Interval time.Duration
	Count    int
}

// type tlsConfig struct {
// 	Cert    tls.Certificate
// 	CaCerts []tlsx509.Certificate
// }

// type tlcpConfig struct {
// 	SigCert gmtls.Certificate
// 	EncCert gmtls.Certificate
// 	CaCerts []x509.Certificate
// }

// buildOptionsWithConfig builds options with given config.
func buildOptionsWithConfig(conf CommonConfig) (*Options, error) {
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
