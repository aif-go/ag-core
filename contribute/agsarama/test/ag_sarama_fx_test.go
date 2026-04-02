package test

import (
	"ag-core/ag/ag_conf"
	"ag-core/contribute/agsarama"
	"ag-core/fxs"
	"context"
	"os"
	"testing"

	"github.com/IBM/sarama"
	"go.uber.org/fx"
)

func TestFxAgsarama(t *testing.T) {
	os.Setenv(ag_conf.AppConfKey, "conf.yaml")

	fxapp := fx.New(

		fxs.FxAgConfModule,

		agsarama.FxAgsaramaConfigModule,

		fx.Invoke(func(client sarama.Client) error {
			return _testProducerByClient(client)
		}),
		fx.Invoke(func(client sarama.Client) error {
			return _testConsumerByClient(client)
		}),
	)

	err := fxapp.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}
