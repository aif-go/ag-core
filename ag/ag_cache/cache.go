// Package ag_cache provides a generic cache abstraction for ag-core.
// Business code uses ICache[T] with zero framework concepts;
// engine implementations live in sub-packages (e.g. agristretto) and are
// contributed via EngineFactory through fx group injection.
package ag_cache

import (
	"context"
	"errors"
	"time"
)

// Sentinel errors for cache operations.
var (
	// ErrCacheMiss is returned by Get / GetOrElse when the key is not present.
	// It is the ONLY error that triggers a loader call (read-through).
	ErrCacheMiss = errors.New("agcache: key not found")
	// ErrBackend wraps any backend engine failure that is NOT a miss.
	// Business code can test with errors.Is(err, agcache.ErrBackend).
	ErrBackend = errors.New("agcache: backend engine error")
)

// LoaderFunc is called by GetOrElse when the key is not in cache.
// Receives both ctx and key so loaders can be reused across keys
// and degrade decorators can reconstruct the original query.
type LoaderFunc[T any] func(ctx context.Context, key string) (T, error)

// ICache is the generic cache interface for business code.
// Intentionally minimal: only the methods used in 90% of call sites.
type ICache[T any] interface {
	// Get reads from cache. Returns (value, nil) on hit, (zero, ErrCacheMiss) on miss.
	Get(ctx context.Context, key string) (T, error)

	// TryGet reads from cache without a miss error: (value, true, nil) on hit,
	// (zero, false, nil) on miss, (zero, false, err) on backend failure.
	// Use it for existence checks and probing (e.g. monitoring expected hits).
	TryGet(ctx context.Context, key string) (T, bool, error)

	// GetOrElse reads from cache; on miss, calls loader, caches the result, and returns it.
	GetOrElse(ctx context.Context, key string, loader LoaderFunc[T]) (T, error)

	// Set writes a value. The TTL is the namespace default (WithDefaultTTL or the
	// engine's internal default); the business interface does not expose TTL here.
	Set(ctx context.Context, key string, value T) error

	// SetWithTTL writes a value with an explicit per-entry TTL (highest priority
	// in the TTL chain: SetWithTTL > WithDefaultTTL > engine internal default).
	// ttl=0 means never expire. Engines without TTLSetter treat it as Set.
	SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Clear(ctx context.Context) error
}

// Stats holds cache metrics (reserved for future StatsProvider; not part of Engine).
type Stats struct {
	Hits       int64
	Misses     int64
	Evictions  int64
	EntryCount int64
}
