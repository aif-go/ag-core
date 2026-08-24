package ag_cache

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
)

const (
	// AgCacheConfPrefix is the config root prefix bound by FxAgCacheMode.
	AgCacheConfPrefix = "agcache"
)

// AgCacheProperties is the core assembly config: it only selects which engine
// implementation to use by name. Engine-specific parameters and TTL are owned
// by each engine package (e.g. agcache.ristretto.*).
//
// No value tags: ag_conf matches config keys case-insensitively by field name
// (DefaultEngine ↔ defaultEngine).
type AgCacheProperties struct {
	// DefaultEngine selects the engine implementation name (e.g. "ristretto").
	DefaultEngine string
}

// DefaultAgCacheProperties returns a properties object carrying defaults;
// the binder only overrides keys present in the configuration.
func DefaultAgCacheProperties() *AgCacheProperties {
	return &AgCacheProperties{DefaultEngine: "ristretto"}
}

// BindAgCacheProperties binds agcache.* from the config binder,
// starting from DefaultAgCacheProperties.
func BindAgCacheProperties(binder ag_conf.IBinder) (*AgCacheProperties, error) {
	props := DefaultAgCacheProperties()
	if err := binder.Bind(props, AgCacheConfPrefix); err != nil {
		return nil, err
	}
	return props, nil
}
