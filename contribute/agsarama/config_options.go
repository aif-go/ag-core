package agsarama

import "github.com/IBM/sarama"

type ConfigOption func(conf *sarama.Config) error

// ExtendSaramaConfigWithOptions extends sarama config with given options.
func ExtendSaramaConfigWithOptions(conf *sarama.Config, options ...ConfigOption) error {
	for _, option := range options {
		if err := option(conf); err != nil {
			return err
		}
	}
	return nil
}

// TODO  add more config options
