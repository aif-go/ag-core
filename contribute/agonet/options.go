package agonet

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	// "github.com/tjfoc/gmsm/gmtls"

	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
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

// ValidateServer 校验服务端配置自洽：TLSType 声明与对应 config 必须匹配。
// 在 NewServerWithOptions 构造入口调用，配置错误 fail-fast，避免启动后静默明文/协议错配。
func (opt *Options) ValidateServer() error {
	switch opt.TLSType {
	case TLSType_UNSET, TLSType_NONE:
		return nil // 未声明安全类型：明文监听合法
	case TLSType_TLS:
		if opt.TLSConfig == nil {
			return aerrors.ErrTLSConfigIsNil
		}
	case TLSType_TLCP:
		if opt.TLCPConfig == nil {
			return aerrors.ErrTLCPConfigIsNil
		}
	case TLSTYPE_TLS_TLCP:
		if opt.TLSConfig == nil || opt.TLCPConfig == nil {
			return fmt.Errorf("agonet: tls_tlcp requires both TLSConfig and TLCPConfig")
		}
	default:
		return fmt.Errorf("agonet: unknown TLSType %q", opt.TLSType)
	}
	return nil
}

// ValidateClient 校验客户端配置自洽：对 CliTLSType() fallback 解析后的类型校验。
// 必须用 getter（而非裸 CLI_* 字段）以保留"客户端复用服务端配置"的用法：
// 手工 Options{TLSType: tls} 共用配置场景，CLI_* 为空时 fallback 借服务端字段。
func (opt *Options) ValidateClient() error {
	switch t := opt.CliTLSType(); t {
	case TLSType_UNSET, TLSType_NONE:
		return nil // 明文是客户端合法默认
	case TLSTYPE_TLS_TLCP:
		// tls_tlcp 仅服务端有效；客户端 fallback 到它时 DialContext 无对应分支会静默明文。
		// 客户端须显式 CliType=tls/tlcp（或由 WithAgClientTLSConfig 归一化）。
		return fmt.Errorf("agonet: tls_tlcp invalid for client, set CliType to tls or tlcp")
	case TLSType_TLS:
		if opt.CliTLSConfig() == nil {
			return aerrors.ErrTLSConfigIsNil
		}
	case TLSType_TLCP:
		if opt.CliTLCPConfig() == nil {
			return aerrors.ErrTLCPConfigIsNil
		}
	default:
		return fmt.Errorf("agonet: unknown TLSType %q", t)
	}
	return nil
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
