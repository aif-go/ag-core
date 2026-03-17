package test

import (
	"ag-core/contribute/agonet"
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"go.uber.org/fx"
)

func TestFxClient(t *testing.T) {
	fxapp := fx.New(
		agonet.FxAgonetClientModule,
		fx.Provide(
			MockIBinder,
		),

		// fx.Supply(&TestEventHandler{}),
		fx.Provide(func() agonet.EventHandler {
			return &TestClientEventHandler{}
		}),
		fx.Invoke(func(c agonet.Client) error {
			err := c.Start()
			if err != nil {
				return err
			}
			con, err := c.Dial("tcp", "localhost:9000")
			if err != nil {
				return err
			}
			defer con.Close()

			con.Write([]byte("hello"))
			time.Sleep(time.Millisecond)

			con.Write([]byte("hello2"))
			time.Sleep(time.Millisecond)

			con.Write([]byte("hello3"))
			time.Sleep(time.Millisecond)

			return nil
		}),
	)

	fxapp.Start(context.Background())

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
