// Package ag_cache provides a generic local cache abstraction for ag-core.
// Business code uses ICache[T] / AdminCache[T] with zero framework concepts;
// engine implementations live in sub-packages (e.g. agristretto) and are
// registered via EngineFactory through fx group injection.
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
	// Stats counters (Hits/Misses) are updated.
	Get(ctx context.Context, key string) (T, error)

	// GetOrElse reads from cache; on miss, calls loader, caches the result, and returns it.
	GetOrElse(ctx context.Context, key string, loader LoaderFunc[T]) (T, error)

	// Set writes a value. ttl omitted uses the default TTL (engine-declared or core fallback);
	// 0 = never expire; >0 = explicit TTL.
	Set(ctx context.Context, key string, value T, ttl ...time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Clear(ctx context.Context) error
}

// AdminCache extends ICache with diagnostics and admin operations.
// Business code injects ICache[T]; monitoring/admin code injects AdminCache[T].
type AdminCache[T any] interface {
	ICache[T]
	// Peek is a pure read: does not trigger a loader.
	Peek(ctx context.Context, key string) (T, bool, error)
	Stats() Stats
}

// Stats holds cache metrics.
type Stats struct {
	Hits       int64
	Misses     int64
	Evictions  int64
	EntryCount int64
}
