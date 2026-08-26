package ag_cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
// Set carries a ttl as an execution-layer parameter: engines that support TTL
// apply it (Ristretto SetWithTTL / Redis SETEX); engines without TTL ignore it
// (entries never expire, rely on Del/Clear). ttl=0 means never expire.
type Engine interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Close() error
}

// BulkDelEngine is an optional capability: engines that can delete multiple keys
// in one operation implement it. typedCache.Del probes it and uses DelMany when
// available, otherwise falls back to looping single-key Del.
type BulkDelEngine interface {
	DelMany(ctx context.Context, keys ...string) error
}

// EngineFactory creates named engine instances. Create takes no parameters:
// the engine configuration is held by the factory itself (bound at
// construction time by the engine package). Core never passes engine config.
type EngineFactory interface {
	Name() string
	Create() (Engine, error)
}

// DefaultTTLProvider is an optional capability: engines with their own default
// TTL implement it so core uses that value when Set is called without an
// explicit TTL. Engines without TTL support simply don't implement it; core
// falls back to a 5-minute default.
type DefaultTTLProvider interface {
	DefaultTTL() time.Duration
}

// syncer is an optional capability: engines with async writes implement it
// so GetOrElse can guarantee the miss-loaded value is visible before returning.
// Not part of the Engine SPI — a plain optional interface.
type syncer interface{ Sync() }

// ──────── Engine factory (registry) ────────

var (
	factories   = map[string]EngineFactory{}
	factoriesMu sync.RWMutex
)

// RegisterEngine registers an engine factory by Name. Duplicate registration panics.
func RegisterEngine(f EngineFactory) {
	factoriesMu.Lock()
	defer factoriesMu.Unlock()
	if _, dup := factories[f.Name()]; dup {
		panic("agcache: duplicate engine factory: " + f.Name())
	}
	factories[f.Name()] = f
}

// EngineRegistered reports whether an engine factory with the given name is registered.
func EngineRegistered(name string) bool {
	factoriesMu.RLock()
	defer factoriesMu.RUnlock()
	_, ok := factories[name]
	return ok
}

func getFactory(name string) (EngineFactory, error) {
	factoriesMu.RLock()
	defer factoriesMu.RUnlock()
	f, ok := factories[name]
	if !ok {
		return nil, fmt.Errorf("agcache: engine %q not registered", name)
	}
	return f, nil
}

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
