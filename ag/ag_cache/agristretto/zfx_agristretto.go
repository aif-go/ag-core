package agristretto

import (
	"github.com/aif-go/ag-core/ag/ag_cache"
	"github.com/aif-go/ag-core/ag/ag_conf"

	"go.uber.org/fx"
)

// ProvideAgristrettoFactory 绑定 agcache.ristretto.* 并返回
// ag_cache.EngineFactory（Name="ristretto"），携带引擎配置与
// 引擎声明的默认 TTL。工厂注入 fx group "agcache.engine" 并由 core 消费——
// 引擎从不全局注册。
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

// FxAgCacheRistrettoMode 将 Ristretto 引擎工厂贡献给 core 的
// "agcache.engine" fx group。
var FxAgCacheRistrettoMode = fx.Module("ag_cache.agristretto",
	fx.Provide(
		fx.Annotate(ProvideAgristrettoFactory, fx.ResultTags(`group:"agcache.engine"`)),
	),
)
