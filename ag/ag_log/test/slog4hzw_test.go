package test

import (
	"ag-core/ag/ag_log"
	"log/slog"
	"testing"
)

func TestHzw(t *testing.T) {
	hzwHandler := &ag_log.HzwHandler{}

	log := slog.New(hzwHandler)
	log.Info("message", "key1", "value1", "key2", "value2")

	log2 := log.With("ak1", "av1", "ak2", "av2")
	log2.Info("message2", "key4", "value4")
}
