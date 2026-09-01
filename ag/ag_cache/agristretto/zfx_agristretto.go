package agristretto

import (
	"github.com/aif-go/ag-core/ag/ag_cache"
	"github.com/aif-go/ag-core/ag/ag_conf"

	"go.uber.org/fx"
)

// ProvideAgristrettoFactory binds agcache.ristretto.* and returns an
// ag_cache.EngineFactory (Name="ristretto") carrying the engine config and
// engine-declared default TTL. The factory is injected into the fx group
// "agcache.engine" and consumed by core — the engine never registers globally.
func ProvideAgristrettoFactory(binder ag_conf.IBinder) (ag_cache.EngineFactory, error) {
	props, err := BindRistrettoConfigProperties(binder)
	if err != nil {
		return nil, err
	}
	ttl, err := parseTTL(props.DefaultTTL)
	if err != nil {
		return nil, err
	}
	return agristrettoFactory{
		cfg: RistrettoConfig{MaxCost: props.MaxCost, NumCounters: props.NumCounters},
		ttl: ttl,
	}, nil
}

// FxAgCacheRistrettoMode contributes the Ristretto engine factory to the
// core's "agcache.engine" fx group.
var FxAgCacheRistrettoMode = fx.Module("ag_cache.agristretto",
	fx.Provide(
		fx.Annotate(ProvideAgristrettoFactory, fx.ResultTags(`group:"agcache.engine"`)),
	),
)
