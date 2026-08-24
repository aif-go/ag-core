package ag_cache_test

import (
	"context"
	"testing"

	"github.com/aif-go/ag-core/ag/ag_cache"
	"github.com/aif-go/ag-core/ag/ag_cache/agristretto"
	"github.com/aif-go/ag-core/ag/ag_conf"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// ──────── Fx group 联合装配 ────────

func TestFxAgCacheMode(t *testing.T) {
	binder := newBinder(t, map[string]any{}) // agcache.defaultEngine 默认 ristretto
	var got string
	fxtest.New(t,
		fx.Provide(func() ag_conf.IBinder { return binder }),
		ag_cache.FxAgCacheMode,
		agristretto.FxAgCacheRistrettoMode,
		fx.Invoke(func(m *ag_cache.Manager) {
			c := ag_cache.New[string]("test", func(ctx context.Context, key string) (string, error) {
				return "v", nil
			})
			v, err := c.Get(context.Background(), "k")
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			got = v
		}),
	).RequireStart().RequireStop()

	if got != "v" {
		t.Fatalf("got %q, want v", got)
	}
}

func TestFxAgCacheMode_IdempotentRegister(t *testing.T) {
	binder := newBinder(t, map[string]any{})
	opts := []fx.Option{
		fx.Provide(func() ag_conf.IBinder { return binder }),
		ag_cache.FxAgCacheMode,
		agristretto.FxAgCacheRistrettoMode,
	}
	// Two fx apps in one process: core must skip re-registration (idempotent).
	app1 := fxtest.New(t, opts...)
	app1.RequireStart().RequireStop()
	app2 := fxtest.New(t, opts...)
	app2.RequireStart().RequireStop()
}

// ──────── 生命周期统计 ────────

func TestLogStats(t *testing.T) {
	registerMockEngine()
	props := ag_cache.DefaultAgCacheProperties()
	props.DefaultEngine = "mock"
	m, err := ag_cache.NewManager(props)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	ag_cache.SetDefault(m)
	defer ag_cache.CloseAll()
	ctx := context.Background()

	ag_cache.New[string]("users", strLoader("U1")).GetOrElse(ctx, "u1", strLoader("U1"))
	ag_cache.Get[string]("params").Get(ctx, "missing")

	count := 0
	m.Visit(func(name string, s ag_cache.Stats) { count++ })
	if count != 2 {
		t.Fatalf("expected 2 namespaces, got %d", count)
	}
	ag_cache.LogStats(m) // must not panic; stats verified via Visit
}
