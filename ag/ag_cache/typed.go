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
	engineName string // requested engine name (WithEngine); empty = default engine
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
		defaultTTL: 5 * time.Minute,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Option configures a typedCache.
type Option[T any] func(*typedCache[T])

// WithEngine selects a specific engine implementation by registered name.
// When omitted, the Manager's default engine is used.
func WithEngine[T any](engineName string) Option[T] {
	return func(c *typedCache[T]) { c.engineName = engineName }
}

// WithDefaultTTL overrides the default TTL used when Set is called.
// Takes priority over the engine's DefaultTTLProvider. Negative TTL is rejected.
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
	data, err := c.engine.Get(ctx, key)
	if err != nil {
		var zero T
		return zero, errBackend(err) // ErrCacheMiss passes through, others wrap ErrBackend
	}
	return c.unmarshal(data)
}

// TryGet implements ICache — relaxed read without a miss error.
// (value, true, nil) on hit; (zero, false, nil) on miss; (zero, false, err) on backend failure.
func (c *typedCache[T]) TryGet(ctx context.Context, key string) (T, bool, error) {
	data, err := c.engine.Get(ctx, key)
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

	data, err := c.engine.Get(ctx, key)
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
	res, _, _ := c.sf.Do(key, func() (any, error) {
		// Double-check uses WithoutCancel so the first caller's cancellation
		// cannot poison the check for concurrent waiters (robustness 2.1).
		if data, err := c.engine.Get(context.WithoutCancel(ctx), key); err == nil {
			v, verr := c.unmarshal(data)
			return result{v, verr}, nil
		} else if !errors.Is(err, ErrCacheMiss) {
			var zero T
			return result{zero, errBackend(err)}, nil
		}

		// loader must NOT be cancelled by the first caller's ctx — use
		// WithoutCancel so all concurrent waiters share one load.
		// A loader panic is labelled separately from an engine panic (2.3).
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

		if serr := c.engine.Set(context.WithoutCancel(ctx), key, data, c.defaultTTL); serr != nil {
			// Cache write failure is a backend error (robustness 2.2).
			// Wrap both the result.err (what callers read) and the Do return.
			werr := errBackend(serr)
			return result{v, werr}, werr
		}
		// Guarantee the loaded value is visible to subsequent reads
		// (optional capability for async engines like Ristretto).
		if s, ok := c.engine.(syncer); ok {
			s.Sync()
		}
		return result{v, nil}, nil
	})

	r := res.(result)
	return r.val, r.err
}

// Set implements ICache — writes using the namespace default TTL.
// Note: Set is asynchronous for async-write engines (visibility is not
// immediately guaranteed); use GetOrElse for read-through semantics.
func (c *typedCache[T]) Set(ctx context.Context, key string, value T) (err error) {
	defer c.recoverPanic(&err)

	data, err := c.serializer.Marshal(value)
	if err != nil {
		return err
	}

	return errBackend(c.engine.Set(ctx, key, data, c.defaultTTL))
}

// Del implements ICache — uses BulkDelEngine when available, else loops.
func (c *typedCache[T]) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if b, ok := c.engine.(BulkDelEngine); ok {
		return errBackend(b.DelMany(ctx, keys...))
	}
	for _, key := range keys {
		if err := c.engine.Del(ctx, key); err != nil {
			return errBackend(err)
		}
	}
	return nil
}

// Clear implements ICache — clears all entries of this standalone instance.
func (c *typedCache[T]) Clear(ctx context.Context) error {
	return errBackend(c.engine.Clear(ctx))
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
// Loader panics are labelled separately inside GetOrElse.
func (c *typedCache[T]) recoverPanic(err *error) {
	if r := recover(); r != nil {
		*err = errBackend(fmt.Errorf("engine panic: %v", r))
	}
}

var (
	_ ICache[string] = (*typedCache[string])(nil)
	_ engineCloser   = (*typedCache[string])(nil)
)
