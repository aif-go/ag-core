package test

import (
	"ag-core/contribute/agonet"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	handler := &TestClientEventHandler{}
	client := agonet.NewClient(handler, &agonet.ClientConfig{})

	err := client.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	con, err := client.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer con.Close()

	con.Write([]byte("hello"))
	time.Sleep(time.Millisecond)

	con.Write([]byte("hello2"))
	time.Sleep(time.Millisecond)

	con.Write([]byte("hello3"))
	time.Sleep(time.Millisecond)

	time.Sleep(time.Second * 3)

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
