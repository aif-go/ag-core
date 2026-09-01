package ag_cache

import (
	"context"
	"fmt"

	"go.uber.org/fx"
)

// EngineFactoryParams collects all engine factories contributed by engine
// modules (e.g. agristretto.FxAgCacheRistrettoMode) via fx value groups.
type EngineFactoryParams struct {
	fx.In
	Factories []EngineFactory `group:"agcache.engine"`
}

// NewAgCacheManager builds a Manager from the contributed engine factories and
// the core assembly config: fill the factory map, select the default engine via
// config, and fail fast if the default engine is not registered.
func NewAgCacheManager(p EngineFactoryParams, props *AgCacheProperties) (*Manager, error) {
	m, err := NewManager(props)
	if err != nil {
		return nil, err
	}
	for _, f := range p.Factories {
		m.SetEngineFactory(f.Name(), f)
	}
	if f := m.EngineFactory(m.DefaultEngine()); f == nil {
		return nil, fmt.Errorf("agcache: engine %q not registered (modules: %v)", m.DefaultEngine(), factoryNames(p.Factories))
	}
	return m, nil
}

func factoryNames(fs []EngineFactory) []string {
	names := make([]string, 0, len(fs))
	for _, f := range fs {
		names = append(names, f.Name())
	}
	return names
}

// registerHooks sets the default manager (so DefaultManager() works at runtime)
// and registers the OnStop hook: just Close (stats are not tracked in v3).
func registerHooks(lc fx.Lifecycle, m *Manager) {
	SetDefault(m)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return m.Close()
		},
	})
}

// FxAgCacheMode is the self-contained Fx module for the cache core:
// it collects all engine factories (fx group), registers them, builds the
// Manager from agcache.* config, and wires lifecycle hooks.
var FxAgCacheMode = fx.Module("ag_cache",
	fx.Provide(
		BindAgCacheProperties,
		NewAgCacheManager,
	),
	fx.Invoke(registerHooks),
)
