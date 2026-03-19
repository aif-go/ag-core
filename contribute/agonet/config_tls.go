package agonet

import (
	"crypto/tls"
	"fmt"
)

type (
	TLSType    = string
	TLSVersion = string
	// TLCPVersion = string
	// TLSClientAuthType = string
)

const (
	// #### TLSType ####
	TLSType_UNSET    TLSType = ""         // 未设置
	TLSType_NONE     TLSType = "none"     // 不使用TLS
	TLSType_TLS      TLSType = "tls"      // TLS
	TLSType_TLCP     TLSType = "tlcp"     // TLCP
	TLSTYPE_TLS_TLCP TLSType = "tls_tlcp" // 同时支持TLS和TLCP，仅服务端配置有效

	// #### TLSClientAuthType ####
	// TLSClientAuthType_NoClientCert TLSClientAuthType = "NoClientCert" // 不要求客户端证书

	// #### TLSVersion ####
	TLSVersionTLS13 TLSVersion = "TLS 1.3" // TLS 1.3
	TLSVersionTLS12 TLSVersion = "TLS 1.2" // TLS 1.2
	TLSVersionTLS11 TLSVersion = "TLS 1.1" // TLS 1.1
	TLSVersionTLS10 TLSVersion = "TLS 1.0" // TLS 1.0
)

type SecurityConfig struct {
	Type    TLSType
	CliType TLSType

	CertsDir string // 证书基础路径

	// CaPaths []string

	TLS  TLSConfig
	TLCP TLCPConfig
}

type TLSConfig struct {
	// TLS
	CaPath string
	// CaPaths      []string
	AuthCertPath string
	AuthKeyPath  string
	SignCertPath string
	SignKeyPath  string

	// TLS 扩展配置
	ServerName         string
	InsecureSkipVerify bool
	// NextProtos         []string

	// CipherSuites []uint16 // 密码套件 ID 列表
	// ClientAuthType TLSClientAuthType // 客户端认证类型
	// SessionCacheSize int // 会话缓存大小

	// MinVersion TLSVersion // 最低支持的最小TLS版本
	// MaxVersion TLSVersion // 最高支持的最大TLS版本
}

type TLCPConfig struct {
	// TLCP
	CaPath string
	// CaPaths      []string
	AuthCertPath string
	AuthKeyPath  string
	SignCertPath string
	SignKeyPath  string
	EncCertPath  string
	EncKeyPath   string

	// TLS 扩展配置
	ServerName         string
	InsecureSkipVerify bool

	// NextProtos []string

	// TLCPMinVersion TLCPVersion // 最低支持的最小TLCP版本，目前TLCP只有一个版本
	// TLCPMaxVersion TLCPVersion // 最高支持的最大TLCP版本, 目前TLCP只有一个版本
}

// 支持的协议版本列表
var supportedVersions = []TLSVersion{
	TLSVersionTLS13,
	TLSVersionTLS12,
	TLSVersionTLS11,
	TLSVersionTLS10,
}

func TransformTLSVersion(version TLSVersion) (uint16, error) {
	switch version {
	case TLSVersionTLS13:
		return tls.VersionTLS13, nil
	case TLSVersionTLS12:
		return tls.VersionTLS12, nil
	case TLSVersionTLS11:
		return tls.VersionTLS11, nil
	case TLSVersionTLS10:
		return tls.VersionTLS10, nil
	default:
		return 0, fmt.Errorf("unsupported TLS version: %s", version)
	}
}

func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		Type:     TLSType_NONE,
		CertsDir: "certs",
	}
}
