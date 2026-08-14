package fanout

import (
	"errors"
	"testing"

	"github.com/aif-go/ag-core/ag/ag_conf"
)

// failingBinder Bind 总是返回 error
type failingBinder struct{}

func (failingBinder) GetEnv() ag_conf.IConfigurableEnvironment { return nil }
func (failingBinder) Bind(i any, name ...string) error          { return errors.New("bind fail") }

func TestBindAgSLogFanoutPropertiesBindError(t *testing.T) {
	prop, err := BindAgSLogFanoutProperties(failingBinder{})
	if prop == nil {
		t.Fatal("expected non-nil prop on bind error")
	}
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestNewFanoutHandlerFactorysNilSafe(t *testing.T) {
	factories, err := NewFanoutHandlerFactorys(nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(factories) != 0 {
		t.Fatalf("expected empty factories for nil props, got %d", len(factories))
	}

	factories, err = NewFanoutHandlerFactorys(&AgSlogFanoutProperties{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(factories) != 0 {
		t.Fatalf("expected empty factories for nil Logs, got %d", len(factories))
	}
}
