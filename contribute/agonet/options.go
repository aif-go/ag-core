package agonet

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	// "github.com/tjfoc/gmsm/gmtls"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
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

	// MaxConn 连接上限（0 = 不限制，默认）。链1 R1 出池后 goroutine 数 = 连接数，
	// 远程连接耗尽防护（应用层配额，非精确边界；超限连接静默拒绝，不触发 OnOpen/OnClose）。
	MaxConn int32

	// ShutdownTimeout 优雅关闭 drain 超时（0 = 默认 5s）。A1：引擎关闭时在途事件
	// （ch 存量）的处理上限，防慢 handler 无限拖延关闭；超时后 loop 强制退出。
	ShutdownTimeout time.Duration
	// ReadBufferMinSize 读缓冲最小容量（0 = 默认 4KB）：防池冷启动 cap=0 → 空 Read 忙等；
	// 小包场景可调小省内存（地板——cap 只增不减，仅初始化生效）
	ReadBufferMinSize int
	// ReadBufferMaxSize 读缓冲扩展上限（0 = 默认 64KB）：满读扩展封顶（对端持续大流量
	// → cap 无限翻倍 = 内存 DoS）；大帧服务可调大（段大小效率）、小包服务可调小（内存上界）
	ReadBufferMaxSize int
	// InboundBufferLimit 入站滞留上限（字节，0 = 默认 16MB）：半包/慢速客户端滞留
	// 超限 → 关闭连接（F3：防 inboundBuffer 无上限增长内存 DoS）
	InboundBufferLimit int

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
		// 引擎级配置映射（链1/ A1 新增字段：Options 与 Config 配置面一致）
		ShutdownTimeout: time.Duration(conf.Engine.ShutdownTimeout) * time.Second,
		MaxConn:         conf.Engine.MaxConn,
		// 读缓冲边界（0 = 默认 4096/65536——读循环解析 + 防呆钳制）
		ReadBufferMinSize: conf.Engine.ReadBufferMinSize,
		ReadBufferMaxSize: conf.Engine.ReadBufferMaxSize,
		// F3 入站滞留上限（0 = 默认 16MB——el.read 解析 + 防呆下限 1MB）
		InboundBufferLimit: conf.Engine.InboundBufferLimit,
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
