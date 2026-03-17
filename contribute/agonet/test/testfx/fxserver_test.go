package test

import (
	"ag-core/ag/ag_conf"
	"ag-core/contribute/agonet"
	"fmt"
	"log/slog"
	"testing"

	"go.uber.org/fx"
)

func TestFxServer_Start(t *testing.T) {
	fxapp := fx.New(
		agonet.FxAgonetServerModule,
		fx.Provide(
			MockIBinder,
		),
		fx.Invoke(func(s agonet.Server) {
			go s.Start()
		}),

		// fx.Supply(&TestEventHandler{}),
		fx.Provide(func() agonet.EventHandler {
			return &TestEventHandler{}
		}),
	)

	fxapp.Run()
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

func MockIBinder() ag_conf.IBinder {
	return &MockBinder{}
}

type MockBinder struct {
}

func (m *MockBinder) GetEnv() ag_conf.IConfigurableEnvironment {
	return nil
}

func (m *MockBinder) Bind(i any, name ...string) error {
	return nil
}
