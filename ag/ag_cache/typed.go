package ag_cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"
)

// typedCache wraps an Engine with type-safe serialization and singleflight
// deduplication. Each typedCache owns its Engine (standalone instance).
type typedCache[T any] struct {
	engine     Engine
	name       string // cache name (namespace)
	prefix     string // key prefix "agcache::<name>::"
	serializer Serializer[T]
	defaultTTL time.Duration
	ttlSet     bool // whether defaultTTL was set via Option
	sf         singleflight.Group
}

// NewWithEngine creates an ICache[T] backed by an explicit Engine.
// Use in tests (with MockEngine) or when a cache is built outside a Manager.
func NewWithEngine[T any](engine Engine, opts ...Option[T]) ICache[T] {
	c := &typedCache[T]{
		engine:     engine,
		serializer: DefaultSerializer[T](),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Option configures a typedCache.
type Option[T any] func(*typedCache[T])

// WithDefaultTTL overrides the namespace default TTL used by Set. When set,
// Set delegates to the engine's TTLSetter (business per-cache default);
// when omitted, Set uses the engine's internal default. Negative TTL is rejected.
func WithDefaultTTL[T any](ttl time.Duration) Option[T] {
	if ttl < 0 {
		panic("agcache: WithDefaultTTL: ttl must not be negative")
	}
	return func(c *typedCache[T]) {
		c.defaultTTL = ttl
		c.ttlSet = true
	}
}

// WithSerializer overrides the default serializer.
func WithSerializer[T any](s Serializer[T]) Option[T] {
	return func(c *typedCache[T]) { c.serializer = s }
}

// Get implements ICache — pure read, no loader.
// Returns ErrCacheMiss on miss; ErrBackend-wrapped error on backend failure.
func (c *typedCache[T]) Get(ctx context.Context, key string) (val T, err error) {
	defer c.recoverPanic(&err)
	data, err := c.engine.Get(ctx, c.prefix+key)
	if err != nil {
		var zero T
		return zero, errBackend(err) // ErrCacheMiss passes through, others wrap ErrBackend
	}
	return c.unmarshal(data)
}

// TryGet implements ICache — relaxed read without a miss error.
// (value, true, nil) on hit; (zero, false, nil) on miss; (zero, false, err) on backend failure.
func (c *typedCache[T]) TryGet(ctx context.Context, key string) (T, bool, error) {
	data, err := c.engine.Get(ctx, c.prefix+key)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			var zero T
			return zero, false, nil
		}
		var zero T
		return zero, false, errBackend(err)
	}
	v, err := c.unmarshal(data)
	if err != nil {
		var zero T
		return zero, false, err
	}
	return v, true, nil
}

// GetOrElse implements ICache.
// Returns ErrCacheMiss only if loader returns it; backend failures are
// ErrBackend-wrapped (NOT treated as a miss — no loader storm on backend down).
func (c *typedCache[T]) GetOrElse(ctx context.Context, key string, loader LoaderFunc[T]) (val T, err error) {
	defer c.recoverPanic(&err)

	ekey := c.prefix + key // engine key (namespaced)
	data, err := c.engine.Get(ctx, ekey)
	if err == nil {
		return c.unmarshal(data)
	}
	if !errors.Is(err, ErrCacheMiss) {
		// Backend failure — do NOT call loader (would storm the source).
		var zero T
		return zero, errBackend(err)
	}

	type result struct {
		val T
		err error
	}
	res, _, _ := c.sf.Do(ekey, func() (any, error) {
		// Double-check uses WithoutCancel so the first caller's cancellation
		// cannot poison the check for concurrent waiters.
		if data, err := c.engine.Get(context.WithoutCancel(ctx), ekey); err == nil {
			v, verr := c.unmarshal(data)
			return result{v, verr}, nil
		} else if !errors.Is(err, ErrCacheMiss) {
			var zero T
			return result{zero, errBackend(err)}, nil
		}

		// loader must NOT be cancelled by the first caller's ctx — use
		// WithoutCancel so all concurrent waiters share one load.
		v, lerr := func() (v T, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("agcache: loader panic: %v", r)
				}
			}()
			return loader(context.WithoutCancel(ctx), key)
		}()
		if lerr != nil {
			return result{v, lerr}, lerr
		}

		data, serr := c.serializer.Marshal(v)
		if serr != nil {
			return result{v, serr}, serr
		}

		var setErr error
		if c.ttlSet {
			setErr = c.setWithTTL(context.WithoutCancel(ctx), ekey, data, c.defaultTTL)
		} else {
			setErr = c.engine.Set(context.WithoutCancel(ctx), ekey, data)
		}
		if setErr != nil {
			werr := errBackend(setErr)
			return result{v, werr}, werr
		}
		if s, ok := c.engine.(syncer); ok {
			s.Sync()
		}
		return result{v, nil}, nil
	})

	r := res.(result)
	return r.val, r.err
}

// Set implements ICache — writes using the namespace default TTL.
// WithDefaultTTL set → delegates to the engine's TTLSetter with that default;
// otherwise → engine.Set (engine internal default).
// Note: Set is asynchronous for async-write engines (visibility is not
// immediately guaranteed); use GetOrElse for read-through semantics.
func (c *typedCache[T]) Set(ctx context.Context, key string, value T) (err error) {
	defer c.recoverPanic(&err)

	data, err := c.serializer.Marshal(value)
	if err != nil {
		return err
	}

	ekey := c.prefix + key
	if c.ttlSet {
		return errBackend(c.setWithTTL(ctx, ekey, data, c.defaultTTL))
	}
	return errBackend(c.engine.Set(ctx, ekey, data))
}

// SetWithTTL implements ICache — single-entry explicit TTL (highest priority).
// Probes the engine's TTLSetter; engines without it treat this as Set.
func (c *typedCache[T]) SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) (err error) {
	defer c.recoverPanic(&err)

	data, err := c.serializer.Marshal(value)
	if err != nil {
		return err
	}

	return errBackend(c.setWithTTL(ctx, c.prefix+key, data, ttl))
}

// setWithTTL probes TTLSetter; engines without it fall back to Set (ignore ttl).
func (c *typedCache[T]) setWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if s, ok := c.engine.(TTLSetter); ok {
		return s.SetWithTTL(ctx, key, value, ttl)
	}
	return c.engine.Set(ctx, key, value)
}

// Del implements ICache — uses BulkDelEngine when available, else loops.
func (c *typedCache[T]) Del(ctx context.Context, keys ...string) (err error) {
	defer c.recoverPanic(&err)

	if len(keys) == 0 {
		return nil
	}
	if b, ok := c.engine.(BulkDelEngine); ok {
		ekeys := make([]string, len(keys))
		for i, k := range keys {
			ekeys[i] = c.prefix + k
		}
		return errBackend(b.DelMany(ctx, ekeys...))
	}
	for _, key := range keys {
		if err := c.engine.Del(ctx, c.prefix+key); err != nil {
			return errBackend(err)
		}
	}
	return nil
}

// Clear implements ICache — clears all entries of this standalone instance.
func (c *typedCache[T]) Clear(ctx context.Context) (err error) {
	defer c.recoverPanic(&err)

	return errBackend(c.engine.Clear(ctx, c.prefix))
}

// closeEngine closes the underlying engine (used by Manager.Close).
func (c *typedCache[T]) closeEngine() {
	if c.engine != nil {
		_ = c.engine.Close()
	}
}

func (c *typedCache[T]) unmarshal(data []byte) (T, error) {
	v, err := c.serializer.Unmarshal(data)
	if err != nil {
		var zero T
		return zero, err
	}
	return *v, nil
}

// recoverPanic catches panics from engine calls and converts them to ErrBackend.
func (c *typedCache[T]) recoverPanic(err *error) {
	if r := recover(); r != nil {
		*err = errBackend(fmt.Errorf("engine panic: %v", r))
	}
}

var (
	_ ICache[string] = (*typedCache[string])(nil)
	_ engineCloser   = (*typedCache[string])(nil)
)
