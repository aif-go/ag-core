// Package agristretto provides the Ristretto-backed local engine for ag_cache.
// This is the only package allowed to import Ristretto.
package agristretto

import (
	"context"
	"fmt"
	"time"

	"github.com/aif-go/ag-core/ag/ag_cache"
	"github.com/dgraph-io/ristretto/v2"
)

// RistrettoConfig is the engine's own config. MaxCost lives here —
// memory management is the engine's responsibility.
type RistrettoConfig struct {
	MaxCost     int64
	NumCounters int64
}

// DefaultRistrettoConfig returns a 100MB / 10M-counter default config.
func DefaultRistrettoConfig() RistrettoConfig {
	return RistrettoConfig{MaxCost: 100_000_000, NumCounters: 10_000_000}
}

// String implements fmt.Stringer.
func (c RistrettoConfig) String() string {
	return fmt.Sprintf("RistrettoConfig{MaxCost=%d, NumCounters=%d}", c.MaxCost, c.NumCounters)
}

// ristrettoEngine implements ag_cache.Engine backed by a single Ristretto instance.
// Each instance is standalone — no shared state, no key index.
type ristrettoEngine struct {
	cache *ristretto.Cache[string, []byte]
}

// NewRistrettoEngine creates a local engine from config.
// Zero values fall back to defaults (MaxCost) or derivation (NumCounters).
func NewRistrettoEngine(cfg RistrettoConfig) (ag_cache.Engine, error) {
	if cfg.MaxCost <= 0 {
		cfg = DefaultRistrettoConfig()
	}
	if cfg.NumCounters <= 0 {
		cfg.NumCounters = cfg.MaxCost * 10 / 100
	}

	cache, err := ristretto.NewCache[string, []byte](&ristretto.Config[string, []byte]{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxCost,
		BufferItems: 64,
		Metrics:     false, // v3: Stats 后置，无统计消费，关闭 Metrics 省开销
	})
	if err != nil {
		return nil, err
	}
	return &ristrettoEngine{cache: cache}, nil
}

// Get returns (data, nil) on hit, (nil, ag_cache.ErrCacheMiss) on miss.
func (e *ristrettoEngine) Get(ctx context.Context, key string) ([]byte, error) {
	v, ok := e.cache.Get(key)
	if !ok {
		return nil, ag_cache.ErrCacheMiss
	}
	return v, nil
}

// Set is asynchronous (no Wait). Cost is computed internally from the value
// byte length — the generic SPI carries no cost concept.
func (e *ristrettoEngine) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	cost := int64(len(value))
	if cost < 1 {
		cost = 1
	}
	ok := e.cache.SetWithTTL(key, value, cost, ttl)
	if !ok {
		return fmt.Errorf("ristretto: set dropped (buffer full)")
	}
	return nil
}

// Sync implements ag_cache.syncer — blocks until pending writes are visible to reads.
func (e *ristrettoEngine) Sync() { e.cache.Wait() }

// Del implements ag_cache.Engine.
func (e *ristrettoEngine) Del(ctx context.Context, key string) error {
	e.cache.Del(key)
	return nil
}

// Clear implements ag_cache.Engine.
func (e *ristrettoEngine) Clear(ctx context.Context) error {
	e.cache.Clear()
	return nil
}

// Close implements ag_cache.Engine.
func (e *ristrettoEngine) Close() error {
	e.cache.Close()
	return nil
}

var _ ag_cache.Engine = (*ristrettoEngine)(nil)

// ──────── Engine factory ────────

// agristrettoFactory implements ag_cache.EngineFactory and holds the engine
// config plus the engine-declared default TTL (self-contained, Create takes
// no parameters).
type agristrettoFactory struct {
	cfg RistrettoConfig
	ttl time.Duration
}

// Name returns the registered engine name.
func (f agristrettoFactory) Name() string { return "ristretto" }

// Create builds an engine from the factory-held config.
func (f agristrettoFactory) Create() (ag_cache.Engine, error) {
	return NewRistrettoEngine(f.cfg)
}

// DefaultTTL returns the engine-declared default TTL (ag_cache.DefaultTTLProvider).
func (f agristrettoFactory) DefaultTTL() time.Duration { return f.ttl }

var (
	_ ag_cache.EngineFactory      = agristrettoFactory{}
	_ ag_cache.DefaultTTLProvider = agristrettoFactory{}
)
