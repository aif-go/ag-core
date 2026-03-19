package agonet

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/emmansun/gmsm/smx509"
)

// WithTLSConfig sets up TLS config.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(opts *Options) error {
		opts.TLSConfig = tlsConfig
		return nil
	}
}

// WithCliTLSConfig sets up client TLS config.
func WithCliTLSConfig(tlsConfig *tls.Config) Option {
	return func(opts *Options) error {
		opts.CLI_TLSConfig = tlsConfig
		return nil
	}
}

// WithTLCPConfig sets up TLCP config.
func WithTLCPConfig(tlcpConfig *tlcp.Config) Option {
	return func(opts *Options) error {
		opts.TLCPConfig = tlcpConfig
		return nil
	}
}

// WithCliTLCPConfig sets up client TLCP config.
func WithCliTLCPConfig(tlcpConfig *tlcp.Config) Option {
	return func(opts *Options) error {
		opts.CLI_TLCPConfig = tlcpConfig
		return nil
	}
}

// WithTLSType sets up TLS type.
func WithTLSType(tlsType TLSType) Option {
	return func(opts *Options) error {
		opts.TLSType = tlsType
		return nil
	}
}

// WithCliTLSType sets up client TLS type.
func WithCliTLSType(tlsType TLSType) Option {
	return func(opts *Options) error {
		opts.CLI_TLSType = tlsType
		return nil
	}
}

func WithAgClientTLSConfig(secCfg *SecurityConfig) Option {
	return func(opts *Options) error {
		cliType := secCfg.CliType
		if cliType == TLSType_UNSET {
			cliType = secCfg.Type
		}

		if cliType != TLSType_TLS && cliType != TLSType_TLCP {
			// 客户端只有TLS和TLCP类型
			return nil
		}

		opts.CLI_TLSType = cliType

		if err := SecurityOptions(opts, secCfg, true); err != nil {
			return err
		}

		return nil
	}
}

func WithAgTLSConfig(secCfg *SecurityConfig) Option {
	return func(opts *Options) error {
		if secCfg.Type == TLSType_NONE || secCfg.Type == TLSType_UNSET {
			return nil
		}
		opts.TLSType = secCfg.Type

		if err := SecurityOptions(opts, secCfg, false); err != nil {
			return err
		}

		return nil
	}
}

// SecurityOptions sets up security options.
func SecurityOptions(opts *Options, secCfg *SecurityConfig, iscli bool) error {
	ttype := secCfg.Type

	certsDir := secCfg.CertsDir

	tlsCfg := &secCfg.TLS
	tlcpCfg := &secCfg.TLCP

	if ttype == TLSType_TLS || ttype == TLSTYPE_TLS_TLCP {
		tls, err := buildTLSConfig(certsDir, tlsCfg, iscli)
		if err != nil {
			return err
		}

		if iscli {
			opts.CLI_TLSConfig = tls
		} else {
			opts.TLSConfig = tls
		}
	}

	if ttype == TLSType_TLCP || ttype == TLSTYPE_TLS_TLCP {
		tlcp, err := buildTLCPConfig(certsDir, tlcpCfg, iscli)
		if err != nil {
			return err
		}

		if iscli {
			opts.CLI_TLCPConfig = tlcp
		} else {
			opts.TLCPConfig = tlcp
		}
	}

	return nil
}

func buildTLSConfig(certsDir string, cfg *TLSConfig, iscli bool) (*tls.Config, error) {
	tlsCfg := &tls.Config{}

	// 证书配置
	if iscli {
		// 客户端配置,证书使用Auth
		if cfg.AuthCertPath != "" && cfg.AuthKeyPath != "" {
			authAbsPath := path.Join(certsDir, cfg.AuthCertPath)
			authKeyAbsPath := path.Join(certsDir, cfg.AuthKeyPath)

			authCert, err := getTlsCertByPath(authAbsPath, authKeyAbsPath)
			if err != nil {
				return nil, err
			}
			tlsCfg.Certificates = []tls.Certificate{*authCert}
		} else {
			// 客户端：若不需要客户端身份认证则可以为空，否则至少1对密钥对。
			// return nil, fmt.Errorf("auth cert path or key path is empty")
		}
	} else {
		// 服务端配置,证书使用Sign
		if cfg.SignCertPath != "" && cfg.SignKeyPath != "" {
			signAbsPath := path.Join(certsDir, cfg.SignCertPath)
			signKeyAbsPath := path.Join(certsDir, cfg.SignKeyPath)

			signCert, err := getTlsCertByPath(signAbsPath, signKeyAbsPath)
			if err != nil {
				return nil, err
			}
			// tlsCfg.Certificates = append(tlsCfg.Certificates, *signCert)
			tlsCfg.Certificates = []tls.Certificate{*signCert}
		} else {
			return nil, fmt.Errorf("sign cert path or key path is empty")
		}
	}

	// CA证书配置
	if cfg.CaPath != "" {
		caAbsPath := path.Join(certsDir, cfg.CaPath)
		caCont, err := os.ReadFile(caAbsPath)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()

		if !caCertPool.AppendCertsFromPEM(caCont) {
			return nil, fmt.Errorf("append ca cert to pool failed")
		}
		tlsCfg.RootCAs = caCertPool
	}

	// ServerName
	if cfg.ServerName != "" {
		tlsCfg.ServerName = cfg.ServerName
	}

	// InsecureSkipVerify 是否跳过证书验证(不建议在生产环境中使用)
	if cfg.InsecureSkipVerify {
		slog.Warn("InsecureSkipVerify is true, it may cause security issues")
		tlsCfg.InsecureSkipVerify = cfg.InsecureSkipVerify
	}

	// TODO ClientAuthType and ClientCAs

	return tlsCfg, nil
}

func buildTLCPConfig(certsDir string, cfg *TLCPConfig, iscli bool) (*tlcp.Config, error) {
	tlcpCfg := &tlcp.Config{}

	if iscli {
		// 客户端配置,证书使用Auth
		if cfg.AuthCertPath != "" && cfg.AuthKeyPath != "" {
			authAbsPath := path.Join(certsDir, cfg.AuthCertPath)
			authKeyAbsPath := path.Join(certsDir, cfg.AuthKeyPath)

			authCert, err := getTlcpCertByPath(authAbsPath, authKeyAbsPath)
			if err != nil {
				return nil, err
			}
			// tlcpCfg.Certificates = append(tlcpCfg.Certificates, *authCert)
			tlcpCfg.Certificates = []tlcp.Certificate{*authCert}
		} else {
			// 客户端：若不需要客户端身份认证则可以为空，否则至少1对密钥对。
			// return nil, fmt.Errorf("auth cert path or key path is empty")
		}
	} else {
		if cfg.SignCertPath == "" || cfg.SignKeyPath == "" {
			return nil, fmt.Errorf("sign cert path or key path is empty")
		}
		if cfg.EncCertPath == "" || cfg.EncKeyPath == "" {
			return nil, fmt.Errorf("enc cert path or key path is empty")
		}

		signAbsPath := path.Join(certsDir, cfg.SignCertPath)
		signKeyAbsPath := path.Join(certsDir, cfg.SignKeyPath)
		encAbsPath := path.Join(certsDir, cfg.EncCertPath)
		encKeyAbsPath := path.Join(certsDir, cfg.EncKeyPath)

		signCert, err := getTlcpCertByPath(signAbsPath, signKeyAbsPath)
		if err != nil {
			return nil, err
		}
		encCert, err := getTlcpCertByPath(encAbsPath, encKeyAbsPath)
		if err != nil {
			return nil, err
		}

		// tlcpCfg.Certificates = append(tlcpCfg.Certificates, *signCert, *encCert)
		// 服务端：至少2对密钥对和证书，按照顺序[签名密钥对, 加密密钥对]
		tlcpCfg.Certificates = []tlcp.Certificate{*signCert, *encCert}
	}

	// CA证书配置
	if cfg.CaPath != "" {
		caAbsPath := path.Join(certsDir, cfg.CaPath)
		caCont, err := os.ReadFile(caAbsPath)
		if err != nil {
			return nil, err
		}

		caCertPool := smx509.NewCertPool()

		if !caCertPool.AppendCertsFromPEM(caCont) {
			return nil, fmt.Errorf("append ca cert to pool failed")
		}
		tlcpCfg.RootCAs = caCertPool
	}

	// ServerName
	if cfg.ServerName != "" {
		tlcpCfg.ServerName = cfg.ServerName
	}

	// InsecureSkipVerify 是否跳过证书验证(不建议在生产环境中使用)
	if cfg.InsecureSkipVerify {
		slog.Warn("InsecureSkipVerify is true, it may cause security issues")
		tlcpCfg.InsecureSkipVerify = cfg.InsecureSkipVerify
	}

	// TODO ClientAuthType and ClientCAs

	return tlcpCfg, nil
}

func getTlsCertByPath(certPth, keyPth string) (*tls.Certificate, error) {
	_, ccont, err := checkFileReadableAndReadFile(certPth)
	if err != nil {
		return nil, err
	}
	_, kcont, err := checkFileReadableAndReadFile(keyPth)
	if err != nil {
		return nil, err
	}

	tlsCert, err := tls.X509KeyPair(ccont, kcont)
	if err != nil {
		return nil, err
	}

	return &tlsCert, nil
}

func getTlcpCertByPath(certPth, keyPth string) (*tlcp.Certificate, error) {
	_, ccont, err := checkFileReadableAndReadFile(certPth)
	if err != nil {
		return nil, err
	}
	_, kcont, err := checkFileReadableAndReadFile(keyPth)
	if err != nil {
		return nil, err
	}

	tlcpCert, err := tlcp.X509KeyPair(ccont, kcont)
	if err != nil {
		return nil, err
	}

	return &tlcpCert, nil
}

// checkFileReadable 检查文件是否存在且有读权限
func checkFileReadableAndReadFile(filePath string) (bool, []byte, error) {

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, fmt.Errorf("file %s is not exist", filePath)
		}
		if os.IsPermission(err) {
			return false, nil, fmt.Errorf("file %s has no read permission", filePath)
		}
		return false, nil, fmt.Errorf("open file %s failed: %w", filePath, err)
	}
	// 成功打开后必须关闭文件
	defer file.Close()
	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return false, nil, fmt.Errorf("read file failed: %w", err)
	}
	return true, content, nil

	// // 1. 获取文件信息，判断是否存在
	// fileInfo, err := os.Stat(filePath)
	// if err != nil {
	// 	if os.IsNotExist(err) {
	// 		return false, errors.New("file is not exist")
	// 	}
	// 	// 其他错误（如权限不足但文件存在、路径是目录等）
	// 	return false, fmt.Errorf("check file readable failed: %w", err)
	// }

	// // 2. 检查是否是文件（不是目录）
	// if fileInfo.IsDir() {
	// 	return false, errors.New("path is directory, not file")
	// }

	// // 3. 检查读权限（分系统处理）
	// // Unix/Linux/Mac：通过文件模式位判断
	// if runtime.GOOS != "windows" {
	// 	// 获取文件权限位（0444 代表读权限，分别对应所有者/组/其他）
	// 	perm := fileInfo.Mode().Perm()
	// 	// 检查当前用户是否有读权限（简化版：只要任意读权限位存在即可）
	// 	// 更精准的判断需要结合用户/组，新手可先用此简化版
	// 	if (perm&0400) == 0 && (perm&0040) == 0 && (perm&0004) == 0 {
	// 		return false, errors.New("file has no read permission")
	// 	}
	// }

	// // Windows：无法通过Mode直接判断，建议用Open验证（方案2）
	// // 此处默认Windows下Stat成功即代表可访问（或后续用方案2）

	// return true, nil
}
