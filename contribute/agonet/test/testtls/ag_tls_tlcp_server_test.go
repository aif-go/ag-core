package testtls

import (
	"ag-core/contribute/agonet"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func TestAgTlsTlcp_Server(t *testing.T) {
	handler := &TestEventHandler{}

	tlsCfg, err := ersa_LoadServer_MTLS_SigConfig()
	if err != nil {
		t.Fatalf("ersa_LoadServer_MTLS_SigConfig failed: %v", err)
	}

	tlcpCfg, err := egms_LoadServerMutualTLCPAuthConfig()
	if err != nil {
		t.Fatalf("egms_LoadServerMutualTLCPAuthConfig failed: %v", err)
	}

	// tlcpCfg.InsecureSkipVerify = true // 禁用证书验证（仅用于测试）

	opts := &agonet.Options{
		Multicore:    true,
		NumEventLoop: 1,
		KeepAlive: agonet.KeepAlive{
			Enable:   true,
			Idle:     time.Duration(5) * time.Second,
			Interval: time.Duration(5) * time.Second,
			Count:    3,
		},
		TLSType: agonet.TLSTYPE_TLS_TLCP,
		// TLSType:    agonet.TLSTypeNone,
		TLSConfig:  tlsCfg,
		TLCPConfig: tlcpCfg,
	}

	server, err := agonet.NewServerWithOptions(handler, []string{"tcp://:8443"}, opts)
	if err != nil {
		t.Fatalf("NewServerWithOptions failed: %v", err)
	}

	err = server.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}

type TestEventHandler struct {
	agonet.BuiltinEventEngine
}

func (e *TestEventHandler) OnTraffic(c agonet.Conn) (action agonet.Action) {
	// msg := make([]byte, 50)
	msg := make([]byte, 50)
	_, err := c.Read(msg)
	if err != nil {
		return agonet.Close
	}

	slog.Info(fmt.Sprintf("Received message: %s", msg))

	resp := fmt.Sprintf("Echo:%s", msg)
	// resp := fmt.Sprintf("Echo: %d", len(msg))

	_, err = c.Write([]byte(resp))
	if err != nil {
		return agonet.Close
	}
	c.Flush()

	return
}

func (e *TestEventHandler) OnOpen(c agonet.Conn) (out []byte, action agonet.Action) {
	return
}

func (e *TestEventHandler) OnClose(c agonet.Conn, err error) (action agonet.Action) {
	slog.Info(fmt.Sprintf("OnClose: %v", err))
	return
}
