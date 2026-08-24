package ag_cache

import (
	"context"
	"time"
)

// LoaderCache binds a default loader to a cache, so Get becomes read-through
// (miss → loader → cache → return) without passing the loader at every call site.
//
// Semantics: LoaderCache.Get is a side-effecting read (may write cache), unlike
// ICache.Get (pure read). The type name carries the semantic — like Spring's
// @Cacheable method. For a pure read, use Peek (Admin view); for a one-off
// different loader, use GetOrElse with a custom loader.
type LoaderCache[T any] struct {
	inner  ICache[T]
	loader LoaderFunc[T]
}

// WithLoader wraps an existing cache with a default loader.
// Useful when the cache comes from an explicit instance.
func WithLoader[T any](c ICache[T], loader LoaderFunc[T]) *LoaderCache[T] {
	return &LoaderCache[T]{inner: c, loader: loader}
}

// Get is read-through: miss → default loader → cached → return.
func (c *LoaderCache[T]) Get(ctx context.Context, key string) (T, error) {
	return c.inner.GetOrElse(ctx, key, c.loader)
}

// GetOrElse forwards; pass a custom loader for a one-off override.
func (c *LoaderCache[T]) GetOrElse(ctx context.Context, key string, loader LoaderFunc[T]) (T, error) {
	return c.inner.GetOrElse(ctx, key, loader)
}

// Set forwards.
func (c *LoaderCache[T]) Set(ctx context.Context, key string, value T, ttl ...time.Duration) error {
	return c.inner.Set(ctx, key, value, ttl...)
}

// Del forwards.
func (c *LoaderCache[T]) Del(ctx context.Context, keys ...string) error {
	return c.inner.Del(ctx, keys...)
}

// Clear forwards.
func (c *LoaderCache[T]) Clear(ctx context.Context) error {
	return c.inner.Clear(ctx)
}

// Peek is the pure-read escape hatch (Admin view).
func (c *LoaderCache[T]) Peek(ctx context.Context, key string) (T, bool, error) {
	if admin, ok := c.inner.(AdminCache[T]); ok {
		return admin.Peek(ctx, key)
	}
	// Fallback: Get and translate ErrCacheMiss.
	v, err := c.inner.Get(ctx, key)
	if err != nil {
		var zero T
		if err == ErrCacheMiss {
			return zero, false, nil
		}
		return zero, false, err
	}
	return v, true, nil
}

// Stats forwards to the inner admin view if available.
func (c *LoaderCache[T]) Stats() Stats {
	if admin, ok := c.inner.(AdminCache[T]); ok {
		return admin.Stats()
	}
	return Stats{}
}

var _ ICache[string] = (*LoaderCache[string])(nil)
