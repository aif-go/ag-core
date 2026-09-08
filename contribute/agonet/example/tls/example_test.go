package main

// tls 样例集成回归：TLS（自签证书）+ TLCP（国密，复用仓库 test/certs/fgmsm 证书）。
// 覆盖：TLS 握手+往返、明文连 TLS 失败、InsecureSkipVerify；TLCP 握手+往返、明文连 TLCP 失败。
// 端口 18884 独立。

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/aif-go/ag-core/contribute/agonet"
	"github.com/emmansun/gmsm/smx509"
)

const certDir = "../../test/certs/fgmsm"

// testCert 包级共享的自签证书（server/client 必须同 CA 对——各自生成会验证失败）
var (
	testCertOnce                          sync.Once
	testCertCa, testCertCert, testCertKey []byte
	testCertErr                           error
)

func testCert() (caPEM, certPEM, keyPEM []byte, err error) {
	testCertOnce.Do(func() {
		testCertCa, testCertCert, testCertKey, testCertErr = genSelfSignedCert()
	})
	return testCertCa, testCertCert, testCertKey, testCertErr
}

// tlsServerOpts 自签 TLS server 配置。
func tlsServerOpts(t *testing.T) *agonet.Options {
	t.Helper()
	_, certPEM, keyPEM, err := testCert()
	if err != nil {
		t.Fatal(err)
	}
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	return &agonet.Options{
		TLSType:   agonet.TLSType_TLS,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
	}
}

// tlsClientOpts 自签 TLS client 配置（信任自签 CA）。
func tlsClientOpts(t *testing.T) *agonet.Options {
	t.Helper()
	caPEM, _, _, err := testCert()
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		t.Fatal("append ca failed")
	}
	return &agonet.Options{
		CLI_TLSType:   agonet.TLSType_TLS,
		CLI_TLSConfig: &tls.Config{RootCAs: pool, ServerName: "localhost"},
	}
}

// tlcpCAPool 加载仓库国密 CA 证书池。
func tlcpCAPool(t *testing.T) *smx509.CertPool {
	t.Helper()
	pool := smx509.NewCertPool()
	for _, p := range []string{"SM2_CA.cer", "RSA_CA.cer"} {
		b, err := os.ReadFile(certDir + "/" + p)
		if err != nil {
			t.Fatal(err)
		}
		if !pool.AppendCertsFromPEM(b) {
			t.Fatalf("append ca %s failed", p)
		}
	}
	return pool
}

// tlcpServerOpts TLCP server 配置（签名+加密证书 + CA）。
func tlcpServerOpts(t *testing.T) *agonet.Options {
	t.Helper()
	sig, err := tlcp.LoadX509KeyPair(certDir+"/sm2_sign_cert.cer", certDir+"/sm2_sign_key.pem")
	if err != nil {
		t.Fatal(err)
	}
	enc, err := tlcp.LoadX509KeyPair(certDir+"/sm2_enc_cert.cer", certDir+"/sm2_enc_key.pem")
	if err != nil {
		t.Fatal(err)
	}
	return &agonet.Options{
		TLSType: agonet.TLSType_TLCP,
		TLCPConfig: &tlcp.Config{
			Certificates: []tlcp.Certificate{sig, enc},
			RootCAs:      tlcpCAPool(t),
		},
	}
}

// tlcpClientOpts TLCP client 配置（单边认证：信任 CA + skip SAN）。
func tlcpClientOpts(t *testing.T) *agonet.Options {
	t.Helper()
	return &agonet.Options{
		TLSType:    agonet.TLSType_TLCP,
		TLCPConfig: &tlcp.Config{RootCAs: tlcpCAPool(t), InsecureSkipVerify: true},
	}
}

// startSecureServer 启动加密 server（阻塞式 Start 须 goroutine 包裹）。
func startSecureServer(t *testing.T, opts *agonet.Options) {
	t.Helper()
	sh := &echoHandler{}
	srv, err := agonet.NewServerWithOptions(sh, []string{addr}, opts)
	if err != nil {
		t.Fatal(err)
	}
	startErr := make(chan error, 1)
	go func() { startErr <- srv.Start() }()
	t.Cleanup(func() {
		select {
		case err := <-startErr:
			if err != nil {
				t.Errorf("server.Start returned err=%v, expect nil after Stop", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server.Start did not return after Stop")
		}
	})
	t.Cleanup(func() { _ = srv.Stop() })

	// 等端口就绪
	deadline := time.Now().Add(3 * time.Second)
	for {
		c, err := net.DialTimeout("tcp", host, 200*time.Millisecond)
		if err == nil {
			_ = c.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("server port not ready")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// secureClient 创建加密 client（带接收 handler）。
func secureClient(t *testing.T, opts *agonet.Options) (agonet.Client, *recvHandler) {
	t.Helper()
	rh := &recvHandler{got: make(chan []byte, 4)}
	cli, err := agonet.NewClientWithOptions(rh, opts)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cli.Stop() })
	return cli, rh
}

// roundTrip 写 payload → 经 handler channel 收回声，断言一致。
func roundTrip(t *testing.T, conn agonet.Conn, rh *recvHandler, payload []byte) {
	t.Helper()
	if _, err := conn.Write(payload); err != nil {
		t.Fatalf("Write err: %v", err)
	}
	deadline := time.After(5 * time.Second)
	for {
		select {
		case echoed := <-rh.got:
			if string(echoed) == string(payload) {
				return
			}
		case <-deadline:
			t.Fatalf("roundtrip timeout")
		}
	}
}

func TestTLS_SelfSignedRoundTrip(t *testing.T) {
	startSecureServer(t, tlsServerOpts(t))
	cli, rh := secureClient(t, tlsClientOpts(t))

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatalf("TLS dial err: %v", err)
	}
	defer conn.Close()

	roundTrip(t, conn, rh, []byte("hello over tls"))
}

func TestTLS_PlainClientFails(t *testing.T) {
	startSecureServer(t, tlsServerOpts(t))

	// 明文 client（不配 TLS）连 TLS server → 服务端握手失败关连接
	cfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(&agonet.BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	defer cli.Stop()

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err) // TCP 层可建立；TLS 握手失败在数据阶段
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("plaintext")); err != nil {
		return // 写即失败也符合预期
	}
	// 服务端 TLS 握手失败 → 连接被关 → 读应出错
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 16)
	for {
		if _, err := conn.Read(buf); err != nil {
			return // 连接被关（预期）
		}
	}
}

func TestTLS_InsecureSkipVerify(t *testing.T) {
	startSecureServer(t, tlsServerOpts(t))

	// 无 RootCAs + InsecureSkipVerify：跳过证书校验，握手成功
	opts := &agonet.Options{
		CLI_TLSType:   agonet.TLSType_TLS,
		CLI_TLSConfig: &tls.Config{InsecureSkipVerify: true, ServerName: "localhost"},
	}
	cli, rh := secureClient(t, opts)

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatalf("TLS dial (skip verify) err: %v", err)
	}
	defer conn.Close()

	roundTrip(t, conn, rh, []byte("skip-verify-ok"))
}

func TestTLCP_RoundTrip(t *testing.T) {
	startSecureServer(t, tlcpServerOpts(t))
	cli, rh := secureClient(t, tlcpClientOpts(t))

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatalf("TLCP dial err: %v", err)
	}
	defer conn.Close()

	roundTrip(t, conn, rh, []byte("hello over tlcp"))
}

func TestTLCP_PlainClientFails(t *testing.T) {
	startSecureServer(t, tlcpServerOpts(t))

	// 明文 client 连 TLCP server → 失败
	cfg := agonet.DefaultClientConfig()
	cli, err := agonet.NewClient(&agonet.BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Start(); err != nil {
		t.Fatal(err)
	}
	defer cli.Stop()

	conn, err := cli.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("plaintext")); err != nil {
		return
	}
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 16)
	for {
		if _, err := conn.Read(buf); err != nil {
			return
		}
	}
}
