package test

import (
	"github.com/aif-go/ag-core/contribute/agonet"
	"fmt"
	"log/slog"
	"testing"
)

func TestServer_Start(t *testing.T) {
	// TODO 测试启动服务器
	handler := &TestEventHandler{}
	server, err := agonet.NewServer(handler, &agonet.ServerConfig{
		Addr: "tcp://:9000",
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
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

	// _, err = c.Write([]byte(resp))
	err = c.AsyncWrite([]byte(resp), nil)
	if err != nil {
		return agonet.Close
	}
	c.Flush()

	return
}
