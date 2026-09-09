// Package main tls 样例：TLS 加密通道（自签证书）。
// 运行：go run ./contribute/agonet/example/tls
// 回归：go test -race ./contribute/agonet/example/tls（含 TLCP 国密通道测试）
//
// 演示：Options 直接配置（WithTLSConfig/WithTLSType 等价）——
//
//	server: TLSType=TLS + TLSConfig{Certificates: 自签证书}
//	client: CLITLSType=TLS + CLITLSConfig{RootCAs: 自签 CA}
//
// TLCP（国密）配置见 example_test.go（复用仓库 test/certs/fgmsm 证书）。
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/aif-go/ag-core/contribute/agonet"
)

// genSelfSignedCert 生成自签 CA + localhost 服务端证书（测试内，无外部依赖）。
func genSelfSignedCert() (caPEM, certPEM, keyPEM []byte, err error) {
	now := time.Now()

	// CA
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "example-ca"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, err
	}
	caPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})

	// 服务端证书（CA 签发，SAN=localhost/127.0.0.1）
	srvKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}
	srvTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	srvDER, err := x509.CreateCertificate(rand.Reader, srvTmpl, caTmpl, &srvKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, err
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: srvDER})
	keyDER, err := x509.MarshalECPrivateKey(srvKey)
	if err != nil {
		return nil, nil, nil, err
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return caPEM, certPEM, keyPEM, nil
}

const addr = "tcp://127.0.0.1:18884"
const host = "127.0.0.1:18884"

// readAllInLoop 在 eventloop goroutine 内读出当前全部入站数据（C7：loop 内读）。
func readAllInLoop(c agonet.Conn) []byte {
	buf := make([]byte, 4096)
	var out []byte
	for {
		n, err := c.Read(buf)
		if n > 0 {
			out = append(out, buf[:n]...)
		}
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
	}
	return out
}

// echoHandler 服务端回显 handler。
type echoHandler struct{}

func (h *echoHandler) OnBoot(agonet.Engine) agonet.Action         { return agonet.None }
func (h *echoHandler) OnShutdown(agonet.Engine)                   {}
func (h *echoHandler) OnOpen(agonet.Conn) ([]byte, agonet.Action) { return nil, agonet.None }
func (h *echoHandler) OnClose(agonet.Conn, error) agonet.Action   { return agonet.None }
func (h *echoHandler) OnTraffic(c agonet.Conn) agonet.Action {
	if data := readAllInLoop(c); len(data) > 0 {
		if _, werr := c.Write(data); werr != nil {
			return agonet.Close
		}
	}
	return agonet.None
}

// recvHandler 客户端接收 handler（loop 内读 → channel）。
type recvHandler struct {
	got chan []byte
}

func (h *recvHandler) OnBoot(agonet.Engine) agonet.Action         { return agonet.None }
func (h *recvHandler) OnShutdown(agonet.Engine)                   {}
func (h *recvHandler) OnOpen(agonet.Conn) ([]byte, agonet.Action) { return nil, agonet.None }
func (h *recvHandler) OnClose(agonet.Conn, error) agonet.Action   { return agonet.None }
func (h *recvHandler) OnTraffic(c agonet.Conn) agonet.Action {
	if data := readAllInLoop(c); len(data) > 0 {
		h.got <- data
	}
	return agonet.None
}

func main() {
	caPEM, certPEM, keyPEM, err := genSelfSignedCert()
	if err != nil {
		panic(err)
	}
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}

	// 服务端（TLS）
	sh := &echoHandler{}
	srvOpts := &agonet.Options{
		TLSType:   agonet.TLSType_TLS,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
	}
	srv, err := agonet.NewServerWithOptions(sh, []string{addr}, srvOpts)
	if err != nil {
		panic(err)
	}
	go func() { _ = srv.Start() }()
	defer srv.Stop()

	// 等端口就绪
	deadline := time.Now().Add(3 * time.Second)
	for {
		c, err := net.DialTimeout("tcp", host, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			break
		}
		if time.Now().After(deadline) {
			panic("server not ready")
		}
		time.Sleep(20 * time.Millisecond)
	}

	// 客户端（TLS，信任自签 CA）
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		panic("append ca failed")
	}
	cliOpts := &agonet.Options{
		CLI_TLSType:   agonet.TLSType_TLS,
		CLI_TLSConfig: &tls.Config{RootCAs: pool, ServerName: "localhost"},
	}
	rh := &recvHandler{got: make(chan []byte, 4)}
	cli, err := agonet.NewClientWithOptions(rh, cliOpts)
	if err != nil {
		panic(err)
	}
	if err := cli.Start(); err != nil {
		panic(err)
	}
	defer cli.Stop()

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	payload := []byte("hello over tls")
	if _, err := conn.Write(payload); err != nil {
		panic(err)
	}
	select {
	case echoed := <-rh.got:
		if string(echoed) != string(payload) {
			panic(fmt.Sprintf("tls echo mismatch: got %q want %q", echoed, payload))
		}
		fmt.Printf("tls ok: %q\n", echoed)
	case <-time.After(3 * time.Second):
		panic("tls echo timeout")
	}
}
