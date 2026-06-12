package agsarama

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
	"fmt"

	"github.com/IBM/sarama"
)

const (
	AgsaramaConfigPrefix = "agsarama"
)

func NewAgsaramaConfig(binder ag_conf.IBinder) (*Config, error) {
	cfg := NewDefaultConfig()
	err := binder.Bind(cfg, AgsaramaConfigPrefix)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func TransConfigToSaramaConfig(cfg *Config) (*sarama.Config, error) {
	return cfg.ToSaramaConfig()
}

func NewClientWithAgConfig(cfg *Config) (sarama.Client, error) {
	brokers := cfg.Brokers
	if len(brokers) == 0 {
		return nil, fmt.Errorf("brokers is empty")
	}

	saramaConfig, err := cfg.ToSaramaConfig()
	if err != nil {
		return nil, err
	}
	client, err := sarama.NewClient(brokers, saramaConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}
