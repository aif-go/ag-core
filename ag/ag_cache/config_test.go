package ag_cache_test

import (
	"errors"
	"testing"

	"github.com/aif-go/ag-core/ag/ag_cache"
	"github.com/aif-go/ag-core/ag/ag_conf"
)

// newBinder 构建 ag_conf.IBinder，用扁平小写属性 map 填充。
func newBinder(t *testing.T, source map[string]any) ag_conf.IBinder {
	t.Helper()
	ps := &ag_conf.MapPropertySource{}
	ps.Name = "test"
	ps.Source = source
	env, err := ag_conf.NewStandardEnvironment()
	if err != nil {
		t.Fatalf("NewStandardEnvironment: %v", err)
	}
	env.GetPropertySources().AddFirst(ps)
	return ag_conf.NewConfigurationPropertiesBinder(env)
}

func TestDefaultAgCacheProperties(t *testing.T) {
	p := ag_cache.DefaultAgCacheProperties()
	if p.DefaultEngine != "ristretto" {
		t.Fatalf("DefaultEngine = %q, want ristretto", p.DefaultEngine)
	}
}

func TestBindAgCacheProperties(t *testing.T) {
	binder := newBinder(t, map[string]any{
		"agcache.defaultengine": "custom-engine",
	})
	props, err := ag_cache.BindAgCacheProperties(binder)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if props.DefaultEngine != "custom-engine" {
		t.Fatalf("DefaultEngine = %q, want custom-engine", props.DefaultEngine)
	}
}

func TestBindAgCacheProperties_MissingKeepsDefault(t *testing.T) {
	binder := newBinder(t, map[string]any{})
	props, err := ag_cache.BindAgCacheProperties(binder)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if props.DefaultEngine != "ristretto" {
		t.Fatalf("DefaultEngine = %q, want ristretto (default)", props.DefaultEngine)
	}
}

// ──────── Bind 错误分支 ────────

// failBinder 是返回固定错误的 ag_conf.IBinder 测试替身。
type failBinder struct{}

func (failBinder) GetEnv() ag_conf.IConfigurableEnvironment { return nil }
func (failBinder) Bind(any, ...string) error                { return errors.New("bind failed") }

func TestBindAgCacheProperties_Error(t *testing.T) {
	if _, err := ag_cache.BindAgCacheProperties(failBinder{}); err == nil {
		t.Fatal("BindAgCacheProperties with failing binder should error")
	}
}
