package testtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/emmansun/gmsm/smx509"
)

const (
	egms_sm2CaCertPath   = "certs/fgmsm/SM2_CA.cer"
	egms_sm2AuthCertPath = "certs/fgmsm/sm2_auth_cert.cer"
	egms_sm2AuthKeyPath  = "certs/fgmsm/sm2_auth_key.pem"
	egms_sm2SignCertPath = "certs/fgmsm/sm2_sign_cert.cer"
	egms_sm2SignKeyPath  = "certs/fgmsm/sm2_sign_key.pem"
	egms_sm2EncCertPath  = "certs/fgmsm/sm2_enc_cert.cer"
	egms_sm2EncKeyPath   = "certs/fgmsm/sm2_enc_key.pem"

	egms_rsaCaCertPath   = "certs/fgmsm/RSA_CA.cer"
	egms_rsaAuthCertPath = "certs/fgmsm/rsa_auth_cert.cer"
	egms_rsaAuthKeyPath  = "certs/fgmsm/rsa_auth_key.pem"
	egms_rsaSignCertPath = "certs/fgmsm/rsa_sign.cer"
	egms_rsaSignKeyPath  = "certs/fgmsm/rsa_sign_key.pem"
)

// #### RSA Load ####

func ersa_SingleSideAuthConfig() (*tls.Config, error) {
	caPool, err := ersa_LoadCertPool()
	if err != nil {
		return nil, err
	}

	tlsCfg := &tls.Config{
		RootCAs: caPool,
	}
	return tlsCfg, nil

}

func ersa_LoadClient_MTLS_AuthConfig() (*tls.Config, error) {
	caPool, err := ersa_LoadCertPool()
	if err != nil {
		return nil, err
	}

	authCert, err := ersa_LoadAuthCert()
	if err != nil {
		return nil, err
	}

	tlsCfg := &tls.Config{
		RootCAs:      caPool,
		Certificates: []tls.Certificate{*authCert},
	}
	return tlsCfg, nil
}

func ersa_LoadServer_MTLS_SigConfig() (*tls.Config, error) {
	caPool, err := ersa_LoadCertPool()
	if err != nil {
		return nil, err
	}

	sigCert, err := ersa_LoadSigCert()
	if err != nil {
		return nil, err
	}

	tlsCfg := &tls.Config{
		RootCAs:      caPool,
		Certificates: []tls.Certificate{*sigCert},
	}
	return tlsCfg, nil
}

// #### 国密SM2 Load ####
// 双向身份认证 服务端配置
func egms_LoadServerMutualTLCPAuthConfig() (*tlcp.Config, error) {

	sigCert, err := egms_LoadSigCert()
	if err != nil {
		return nil, err
	}
	encCert, err := egms_LoadEncCert()
	if err != nil {
		return nil, err
	}

	caPool, err := egms_LoadCertPool()
	if err != nil {
		return nil, err
	}

	tlcpCfg := &tlcp.Config{
		Certificates: []tlcp.Certificate{*sigCert, *encCert},
		RootCAs:      caPool,
	}
	return tlcpCfg, nil
}

// 获取 单向身份认证（只认证服务端） 配置
func egms_SingleSideAuthConfig() (*tlcp.Config, error) {
	// 信任的根证书
	caPool, err := egms_LoadCertPool()
	if err != nil {
		return nil, err
	}

	tlcpCfg := &tlcp.Config{
		RootCAs: caPool,
	}
	return tlcpCfg, nil
}

// #### 国密SM2 ####
func egms_LoadSigCert() (*tlcp.Certificate, error) {
	cert, err := tlcp.LoadX509KeyPair(egms_sm2SignCertPath, egms_sm2SignKeyPath)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func egms_LoadEncCert() (*tlcp.Certificate, error) {
	cert, err := tlcp.LoadX509KeyPair(egms_sm2EncCertPath, egms_sm2EncKeyPath)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func egms_LoadAuthCert() (*tlcp.Certificate, error) {
	cert, err := tlcp.LoadX509KeyPair(egms_sm2AuthCertPath, egms_sm2AuthKeyPath)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func egms_LoadCertPool() (*smx509.CertPool, error) {
	caCertPool := smx509.NewCertPool()

	caCert, err := os.ReadFile(egms_sm2CaCertPath)
	if err != nil {
		return nil, err
	}
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("添加CA证书失败")
	}

	ca2, err := os.ReadFile(egms_rsaCaCertPath)
	if err != nil {
		return nil, err
	}
	if !caCertPool.AppendCertsFromPEM(ca2) {
		return nil, fmt.Errorf("添加CA证书失败")
	}

	return caCertPool, nil
}

// #### RSA ####
func ersa_LoadSigCert() (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(egms_rsaSignCertPath, egms_rsaSignKeyPath)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func ersa_LoadAuthCert() (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(egms_rsaAuthCertPath, egms_rsaAuthKeyPath)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func ersa_LoadCertPool() (*x509.CertPool, error) {
	caCert, err := os.ReadFile(egms_rsaCaCertPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("添加CA证书失败")
	}
	return caCertPool, nil
}
