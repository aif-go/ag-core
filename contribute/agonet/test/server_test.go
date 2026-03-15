package test

import (
	"ag-core/contribute/agonet"
	"fmt"
	"log/slog"
	"testing"
)

func TestServer_Start(t *testing.T) {
	// TODO 测试启动服务器
	handler := &TestEventHandler{}
	server := agonet.NewServer(handler, agonet.ServerConfig{
		Address: "tcp://:9000",
	})

	err := server.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}

type TestEventHandler struct {
	agonet.BuiltinEventEngine
}

func (e *TestEventHandler) OnTraffic(c agonet.Conn) (action agonet.Action) {
	msg := make([]byte, 1024)
	_, err := c.Read(msg)
	if err != nil {
		return agonet.Close
	}

	slog.Info(fmt.Sprintf("Received message: %s", msg))

	resp := fmt.Sprintf("Echo: %s", msg)

	_, err = c.Write([]byte(resp))
	if err != nil {
		return agonet.Close
	}

	return
}
