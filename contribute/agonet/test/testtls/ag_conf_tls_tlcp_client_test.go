package testtls

import (
	"ag-core/contribute/agonet"
	"net"
	"testing"
	"time"
)

var _agonet_tls_cli = `
agonet:
  client:
    config:
      security:
        type: tlcp
        type1: tls
        type2: none
        certsDir: ./certs
        TLS:
          CaPath: fgmsm/RSA_CA.cer
          AuthCertPath: fgmsm/rsa_auth_cert.cer
          AuthKeyPath: fgmsm/rsa_auth_key.pem
          SignCertPath: fgmsm/rsa_sign.cer
          SignKeyPath: fgmsm/rsa_sign_key.pem
          InsecureSkipVerify: true
        TLCP:
          CaPath: fgmsm/SM2_CA.cer
          AuthCertPath: fgmsm/sm2_auth_cert.cer
          AuthKeyPath: fgmsm/sm2_auth_key.pem
          SignCertPath: fgmsm/sm2_sign_cert.cer
          SignKeyPath: fgmsm/sm2_sign_key.pem
          EncCertPath: fgmsm/sm2_enc_cert.cer
          EncKeyPath: fgmsm/sm2_enc_key.pem
          InsecureSkipVerify: true
`

func TestAgCfgTlsTlcp_Client(t *testing.T) {

	binder, err := buildCfgBinder(_agonet_tls_cli)
	if err != nil {
		t.Fatalf("buildCfgBinder failed: %v", err)
	}

	cliCfg, err := agonet.NewClientConfig(binder)
	if err != nil {
		t.Fatalf("NewClientConfig failed: %v", err)
	}

	handler := &TestClientEventHandler{}

	// cli, err := agonet.NewClientWithOptions(handler, opts)
	cli, err := agonet.NewClient(handler, cliCfg)
	if err != nil {
		t.Fatalf("NewClientWithOptions failed: %v", err)
	}

	err = cli.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	tcon, err := cli.Dial("tcp", "localhost:8443")
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	var con net.Conn
	con = tcon
	defer con.Close()

	_, err = con.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	time.Sleep(time.Millisecond)

	con.Write([]byte("hello2"))
	time.Sleep(time.Millisecond)

	// con.Write([]byte("hello3"))
	// time.Sleep(time.Millisecond)

	time.Sleep(time.Second)
	// time.Sleep(time.Second * 10)

}
