package ag_cache

import (
	"context"
	"time"
)

// LoaderCache 将默认 loader 绑定到缓存，使 Get 成为读穿透
// （miss → loader → 缓存 → 返回），无需在每次调用点传 loader。
//
// 语义：LoaderCache.Get 是带副作用的读（可能写缓存），不同于 ICache.Get（纯读）。
// 类型名承载该语义——类似 Spring 的 @Cacheable 方法。
// 需要不带 loader 的纯读时用 TryGet。
type LoaderCache[T any] struct {
	inner  ICache[T]
	loader LoaderFunc[T]
}

// WithLoader 用默认 loader 包装现有缓存。
// 适用于缓存来自显式实例的场景。
func WithLoader[T any](c ICache[T], loader LoaderFunc[T]) *LoaderCache[T] {
	return &LoaderCache[T]{inner: c, loader: loader}
}

// Get 是读穿透：miss → 默认 loader → 缓存 → 返回。
func (c *LoaderCache[T]) Get(ctx context.Context, key string) (T, error) {
	return c.inner.GetOrElse(ctx, key, c.loader)
}

// GetOrElse 转发；可传自定义 loader 做一次性覆盖。
func (c *LoaderCache[T]) GetOrElse(ctx context.Context, key string, loader LoaderFunc[T]) (T, error) {
	return c.inner.GetOrElse(ctx, key, loader)
}

// Set 转发。
func (c *LoaderCache[T]) Set(ctx context.Context, key string, value T) error {
	return c.inner.Set(ctx, key, value)
}

// SetWithTTL 转发——单条显式 TTL。
func (c *LoaderCache[T]) SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) error {
	return c.inner.SetWithTTL(ctx, key, value, ttl)
}

// Del 转发。
func (c *LoaderCache[T]) Del(ctx context.Context, keys ...string) error {
	return c.inner.Del(ctx, keys...)
}

// Clear 转发。
func (c *LoaderCache[T]) Clear(ctx context.Context) error {
	return c.inner.Clear(ctx)
}

// TryGet 转发——无未命中错误的纯读。
func (c *LoaderCache[T]) TryGet(ctx context.Context, key string) (T, bool, error) {
	return c.inner.TryGet(ctx, key)
}

var _ ICache[string] = (*LoaderCache[string])(nil)
