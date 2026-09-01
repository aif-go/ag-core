package agristretto

import (
	"fmt"
	"time"

	"github.com/aif-go/ag-core/ag/ag_conf"
)

const (
	// RistrettoConfPrefix is the engine's own config prefix (bound by the engine).
	RistrettoConfPrefix = "agcache.ristretto"
)

// RistrettoConfigProperties is the engine's YAML-bound config model.
// No value tags: ag_conf matches config keys case-insensitively by field name.
type RistrettoConfigProperties struct {
	// MaxCost is the memory budget in bytes (0 = engine default 100MB).
	MaxCost int64
	// NumCounters is the counter count (0 = derived from MaxCost).
	NumCounters int64
	// DefaultTTL is the engine-declared default TTL: ""→default 5min (same as core fallback);
	// "0"→never expire; "60s"→explicit.
	DefaultTTL string
}

// DefaultRistrettoConfigProperties returns a properties object carrying defaults;
// the binder only overrides keys present in the configuration.
func DefaultRistrettoConfigProperties() *RistrettoConfigProperties {
	return &RistrettoConfigProperties{MaxCost: 100_000_000}
}

// BindRistrettoConfigProperties binds agcache.ristretto.* from the config binder,
// starting from DefaultRistrettoConfigProperties.
func BindRistrettoConfigProperties(binder ag_conf.IBinder) (*RistrettoConfigProperties, error) {
	props := DefaultRistrettoConfigProperties()
	if err := binder.Bind(props, RistrettoConfPrefix); err != nil {
		return nil, err
	}
	return props, nil
}

// parseTTL converts the DefaultTTL string to a duration.
// "" → 5min (engine default, matches core fallback); "0" → 0 (never expire);
// "60s" → 60s. Invalid strings error.
func parseTTL(s string) (time.Duration, error) {
	if s == "" {
		return 5 * time.Minute, nil // TODO 是否默认不超时？
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("agcache: invalid defaultTtl %q: %w", s, err)
	}
	return d, nil
}
