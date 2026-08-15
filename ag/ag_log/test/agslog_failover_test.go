package test

import (
	"context"
	"os"
	"testing"

	"github.com/aif-go/ag-core/ag/ag_conf"
	ag_log "github.com/aif-go/ag-core/ag/ag_log"
	"github.com/aif-go/ag-core/ag/ag_log/agslog"
	"github.com/aif-go/ag-core/fxs"

	"go.uber.org/fx"
)

func TestAgSlogFailover(t *testing.T) {
	os.Setenv(ag_conf.AppConfKey, "agslog_failover.yaml")

	fxapp := fx.New(
		fxs.FxAgConfModule,
		ag_log.FxAglogMode,
	)

	if err := fxapp.Start(context.Background()); err != nil {
		t.Fatalf("fx start error: %v", err)
	}
	defer fxapp.Stop(context.Background())

	logger := agslog.GetSlogByName("fo1")
	if logger == nil {
		t.Fatal("failover logger fo1 should not be nil")
	}

	_test_log(logger, "failover_test")
}
