package ag_cache_test

import (
	"testing"

	"github.com/aif-go/ag-core/ag/ag_cache"
	"github.com/aif-go/ag-core/ag/ag_cache/agristretto"
	"github.com/aif-go/ag-core/ag/ag_conf"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// newBinder builds an ag_conf.IBinder seeded with a flat lowercase property map,
// mimicking YAML flattened keys (agcache.defaultEngine, agcache.ristretto.maxCost, ...).
func newBinder(src map[string]any) ag_conf.IBinder {
	ps := &ag_conf.MapPropertySource{}
	ps.Name = "test"
	ps.Source = src
	env, err := ag_conf.NewStandardEnvironment()
	if err != nil {
		panic(err)
	}
	env.GetPropertySources().AddFirst(ps)
	return ag_conf.NewConfigurationPropertiesBinder(env)
}

// startFx boots the real assembly (core + agristretto engine) with the given YAML
// property map, returns a stop function. Registration is idempotent so multiple
// apps may be booted in one test process.
func startFx(t *testing.T, yaml map[string]any) (stop func()) {
	t.Helper()
	binder := newBinder(yaml)
	app := fxtest.New(t,
		fx.Provide(func() ag_conf.IBinder { return binder }),
		ag_cache.FxAgCacheMode,
		agristretto.FxAgCacheRistrettoMode,
	)
	app.RequireStart()
	return func() {
		app.RequireStop()
		ag_cache.CloseAll() // clear default manager to avoid cross-test leakage
	}
}

// dflt returns the default manager (set by startFx's SetDefault).
func dflt() *ag_cache.Manager {
	m := ag_cache.DefaultManager()
	if m == nil {
		panic("no default manager")
	}
	return m
}
