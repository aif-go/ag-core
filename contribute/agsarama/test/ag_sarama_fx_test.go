package test

import (
	"ag-core/ag/ag_conf"
	"ag-core/contribute/agsarama"
	"ag-core/fxs"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/IBM/sarama"
	"go.uber.org/fx"
)

func TestFxAgsarama(t *testing.T) {
	os.Setenv(ag_conf.AppConfKey, "conf.yaml")

	fxapp := fx.New(

		fxs.FxAgConfModule,

		agsarama.FxAgsaramaModule,

		fx.Invoke(func(client sarama.Client) error {
			return _testProducerByClient(client)
		}),
		fx.Invoke(func(client sarama.Client) error {
			return _testConsumerByClient(client)
		}),

		fx.Provide(
			agsarama.FxAgsaramaGroupTag(
				func() agsarama.ConfigOption {
					return agsarama.ConfigOption(func(conf *sarama.Config) error {
						conf.Producer.Partitioner = sarama.NewRoundRobinPartitioner
						fmt.Printf("=1111======conf: %v\n", conf)
						return nil
					})
				},
			),
		),
		fx.Supply(
			agsarama.FxAgsaramaGroupTag(
				agsarama.ConfigOption(func(conf *sarama.Config) error {
					fmt.Printf("=2222======conf: %v\n", conf)
					return nil
				}),
			),
		),
	)

	err := fxapp.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}
