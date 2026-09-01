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
			c := ag_cache.GetCacheWithLoader(m, "test", func(ctx context.Context, key string) (string, error) {
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
	// 一个进程内两个 fx app：core 需跳过重复注册（幂等）。
	app1 := fxtest.New(t, opts...)
	app1.RequireStart().RequireStop()
	app2 := fxtest.New(t, opts...)
	app2.RequireStart().RequireStop()
}

// ──────── 生命周期（OnStop 只 Close，无统计）───────

// TestFxAgCacheMode_OnStopClose 验证 Fx 联合装配 + 生命周期（已含于 TestFxAgCacheMode）。
// Stats 已后置移除，无需统计断言。

// ──────── 引擎模型：fx group + config 选默认 + fail-fast ────────

func TestEngineModel_GroupDefault(t *testing.T) {
	binder := newBinder(t, map[string]any{"agcache.defaultengine": "ristretto"})
	var got string
	fxtest.New(t,
		fx.Provide(func() ag_conf.IBinder { return binder }),
		ag_cache.FxAgCacheMode,
		agristretto.FxAgCacheRistrettoMode,
		fx.Invoke(func(m *ag_cache.Manager) {
			// config defaultEngine=ristretto → ristrettoFactory 在 group 中
			if m.DefaultEngine() != "ristretto" {
				t.Fatalf("DefaultEngine = %q, want ristretto", m.DefaultEngine())
			}
			if f := m.EngineFactory("ristretto"); f == nil {
				t.Fatal("ristretto factory should be registered from fx group")
			}
			c := ag_cache.GetCacheWithLoader(m, "test", func(ctx context.Context, key string) (string, error) {
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

func TestEngineModel_FailFastUnknownDefault(t *testing.T) {
	binder := newBinder(t, map[string]any{"agcache.defaultengine": "no-such"})
	props, err := ag_cache.BindAgCacheProperties(binder)
	if err != nil {
		t.Fatalf("BindAgCacheProperties: %v", err)
	}
	factory, err := agristretto.ProvideAgristrettoFactory(binder)
	if err != nil {
		t.Fatalf("ProvideAgristrettoFactory: %v", err)
	}
	// 直接调 NewAgCacheManager：group 有 ristretto 但 defaultEngine=no-such → fail-fast
	_, err = ag_cache.NewAgCacheManager(ag_cache.EngineFactoryParams{
		Factories: []ag_cache.EngineFactory{factory},
	}, props)
	if err == nil {
		t.Fatal("NewAgCacheManager should fail when default engine is not registered")
	}
}
