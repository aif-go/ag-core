package ag_cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Engine is the cache engine SPI. It operates on raw []byte, with no
// knowledge of types, serialization, or namespace.
// Each Engine is a standalone instance; memory management is the engine's
// own responsibility (Ristretto: MaxCost; Redis: server-side maxmemory).
//
// All methods take ctx for per-call timeout/cancellation (needed by network
// engines like Redis; local engines may ignore it).
//
// Error contract:
//   - Get returns ErrCacheMiss when the key is not present; any other error
//     means a backend failure (the caller should NOT treat it as a miss).
//   - Set/Del/Clear return nil on success, an error on backend failure.
//
// Set carries no ttl: the default TTL is decided by the engine itself (its own
// config, e.g. Ristretto defaultTtl, or natural behavior). External TTL is
// applied via the optional TTLSetter capability. ttl=0 means never expire.
type Engine interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Del(ctx context.Context, key string) error
	// Clear clears all entries of this engine instance. prefix carries the
	// namespace key prefix ("agcache::<name>::") for engines that share a
	// backend (e.g. Redis SCAN+DEL); local engines may ignore it.
	Clear(ctx context.Context, prefix string) error
	Close() error
}

// cachePrefix builds the namespace key prefix "agcache::<name>::".
func cachePrefix(name string) string {
	return "agcache::" + name + "::"
}

// TTLSetter is an optional capability: engines that support explicit per-entry
// TTL implement it. Business code specifies TTL via ICache.SetWithTTL or
// WithDefaultTTL; typedCache probes this interface and delegates to SetWithTTL.
// Engines that do not implement TTLSetter ignore external TTL (same as Set).
type TTLSetter interface {
	SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

// BulkDelEngine is an optional capability: engines that can delete multiple keys
// in one operation implement it. typedCache.Del probes it and uses DelMany when
// available, otherwise falls back to looping single-key Del.
type BulkDelEngine interface {
	DelMany(ctx context.Context, keys ...string) error
}

// EngineFactory creates named engine instances. Create takes the cache name as
// a namespace context (the engine may ignore it or look up per-name config,
// aligning with Spring createCache(name)); each Create returns a fresh
// standalone instance — reuse and lifecycle are the Manager's job.
type EngineFactory interface {
	Name() string
	Create(name string) (Engine, error)
}

// syncer is an optional capability: engines with async writes implement it
// so GetOrElse can guarantee the miss-loaded value is visible before returning.
// Not part of the Engine SPI — a plain optional interface.
type syncer interface{ Sync() }

// errBackend wraps an engine error as ErrBackend. ErrCacheMiss passes through unchanged.
func errBackend(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrCacheMiss) {
		return err
	}
	return fmt.Errorf("%w: %v", ErrBackend, err)
}
