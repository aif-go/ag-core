package testtls

import (
	"github.com/aif-go/ag-core/contribute/agonet"
	"fmt"
	"log/slog"
	"net"
	"testing"
	"time"
)

func TestAgTlsTlcp_Client(t *testing.T) {
	handler := &TestClientEventHandler{}

	tlsCfg, err := ersa_LoadClient_MTLS_AuthConfig()
	if err != nil {
		t.Fatalf("ersa_LoadClient_MTLS_AuthConfig failed: %v", err)
	}
	// tlsCfg.InsecureSkipVerify = true // 关闭SAN校验

	tlcpCfg, err := egms_SingleSideAuthConfig()
	if err != nil {
		t.Fatalf("egms_LoadClientMutualTLCPAuthConfig failed: %v", err)
	}
	tlcpCfg.InsecureSkipVerify = true // 关闭SAN校验

	opts := &agonet.Options{
		Multicore:    true,
		NumEventLoop: 1,
		KeepAlive: agonet.KeepAlive{
			Enable:   true,
			Idle:     time.Duration(3) * time.Second,
			Interval: time.Duration(3) * time.Second,
			Count:    3,
		},
		// TLSType: agonet.TLSType_TLCP,
		// TLSType: agonet.TLSType_TLS,
		// TLSType:    agonet.TLSTypeNone,
		CLI_TLSType: agonet.TLSType_TLS,
		// CLI_TLSType: agonet.TLSType_NONE,
		// TLSConfig:  tlsCfg,
		// TLCPConfig: tlcpCfg,
	}

	agonet.ExtendOptions(opts,
		// agonet.WithTLSType(agonet.TLSType_TLCP),
		agonet.WithTLSType(agonet.TLSType_TLS),
		agonet.WithTLSConfig(tlsCfg),
		agonet.WithTLCPConfig(tlcpCfg),
	)

	cli, err := agonet.NewClientWithOptions(handler, opts)
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

type TestClientEventHandler struct {
	agonet.BuiltinEventEngine
}

func (e *TestClientEventHandler) OnTraffic(c agonet.Conn) (action agonet.Action) {
	msg := make([]byte, 50)
	_, err := c.Read(msg)
	if err != nil {
		return agonet.Close
	}

	slog.Info(fmt.Sprintf("client Received reply message: %s", msg))

	return
}
