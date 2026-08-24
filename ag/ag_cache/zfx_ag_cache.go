package ag_cache

import (
	"context"
	"log/slog"

	"go.uber.org/fx"
)

// EngineFactoryParams collects all engine factories contributed by engine
// modules (e.g. agristretto.FxAgCacheRistrettoMode) via fx value groups.
type EngineFactoryParams struct {
	fx.In
	Factories []EngineFactory `group:"agcache.engine"`
}

// NewAgCacheManager builds a Manager from the contributed engine factories and
// the core assembly config. Registration is idempotent.
func NewAgCacheManager(p EngineFactoryParams, props *AgCacheProperties) (*Manager, error) {
	for _, f := range p.Factories {
		if !EngineRegistered(f.Name()) {
			RegisterEngine(f)
		}
	}
	return NewManager(props)
}

// registerHooks sets the default manager (so package-level New/Get work)
// and registers the OnStop hook: stats then Close.
func registerHooks(lc fx.Lifecycle, m *Manager) {
	SetDefault(m)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			LogStats(m)
			return m.Close()
		},
	})
}

// LogStats outputs per-namespace cache stats via slog.
func LogStats(m *Manager) {
	m.Visit(func(name string, s Stats) {
		slog.Info("agcache stats", "namespace", name,
			"hits", s.Hits, "misses", s.Misses,
			"evictions", s.Evictions, "entries", s.EntryCount)
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
