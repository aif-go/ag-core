package ag_cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"
)

// typedCache 用类型安全序列化与 singleflight 去重包装 Engine。
// 每个 typedCache 持有自己的 Engine（独立实例）。
type typedCache[T any] struct {
	engine     Engine
	name       string // 缓存名（namespace）
	prefix     string // key 前缀 "agcache::<name>::"
	serializer Serializer[T]
	defaultTTL time.Duration
	ttlSet     bool // 是否经 Option 设置了 defaultTTL
	sf         singleflight.Group
}

// NewWithEngine 创建由显式 Engine 支撑的 ICache[T]。
// 在测试（用 MockEngine）或 Manager 之外构建缓存时使用。
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

// Option 配置一个 typedCache。
type Option[T any] func(*typedCache[T])

// WithDefaultTTL 覆盖 Set 使用的 namespace 默认 TTL。设置后，
// Set 委托给引擎的 TTLSetter（业务 per-cache 默认）；未设置时，
// Set 用引擎内部默认。负 TTL 被拒绝。
func WithDefaultTTL[T any](ttl time.Duration) Option[T] {
	if ttl < 0 {
		panic("agcache: WithDefaultTTL: ttl must not be negative")
	}
	return func(c *typedCache[T]) {
		c.defaultTTL = ttl
		c.ttlSet = true
	}
}

// WithSerializer 覆盖默认序列化器。
func WithSerializer[T any](s Serializer[T]) Option[T] {
	return func(c *typedCache[T]) { c.serializer = s }
}

// Get 实现 ICache——纯读，不调 loader。
// 未命中返回 ErrCacheMiss；后端故障返回 ErrBackend 包装错误。
func (c *typedCache[T]) Get(ctx context.Context, key string) (val T, err error) {
	defer c.recoverPanic(&err)
	data, err := c.engine.Get(ctx, c.prefix+key)
	if err != nil {
		var zero T
		return zero, errBackend(err) // ErrCacheMiss 透传，其他包装 ErrBackend
	}
	return c.unmarshal(data)
}

// TryGet 实现 ICache——无未命中错误的宽松读。
// 命中返回 (value, true, nil)；未命中 (zero, false, nil)；后端故障 (zero, false, err)。
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

// GetOrElse 实现 ICache。
// 仅当 loader 返回时返回 ErrCacheMiss；后端故障包装为 ErrBackend
// （不当未命中——后端故障时不触发 loader，防缓存击穿）。
func (c *typedCache[T]) GetOrElse(ctx context.Context, key string, loader LoaderFunc[T]) (val T, err error) {
	defer c.recoverPanic(&err)

	ekey := c.prefix + key // engine key（namespaced）
	data, err := c.engine.Get(ctx, ekey)
	if err == nil {
		return c.unmarshal(data)
	}
	if !errors.Is(err, ErrCacheMiss) {
		// 后端故障——不调 loader（否则会冲击数据源）。
		var zero T
		return zero, errBackend(err)
	}

	type result struct {
		val T
		err error
	}
	res, _, _ := c.sf.Do(ekey, func() (any, error) {
		// 双重检查用 WithoutCancel，避免首个调用者的取消
		// 污染并发等待者的检查。
		if data, err := c.engine.Get(context.WithoutCancel(ctx), ekey); err == nil {
			v, verr := c.unmarshal(data)
			return result{v, verr}, nil
		} else if !errors.Is(err, ErrCacheMiss) {
			var zero T
			return result{zero, errBackend(err)}, nil
		}

		// loader 不能被首个调用者的 ctx 取消——用 WithoutCancel
		// 使所有并发等待者共享一次加载。
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

// Set 实现 ICache——用 namespace 默认 TTL 写入。
// 设置 WithDefaultTTL → 委托给引擎 TTLSetter 用默认；
// 否则 → engine.Set（引擎内部默认）。
// 注意：对异步写引擎 Set 是异步的（可见性不立即保证）；读穿透用 GetOrElse。
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

// SetWithTTL 实现 ICache——单条显式 TTL（最高优先级）。
// 探测引擎 TTLSetter；无此能力的引擎等同 Set。
func (c *typedCache[T]) SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) (err error) {
	defer c.recoverPanic(&err)

	data, err := c.serializer.Marshal(value)
	if err != nil {
		return err
	}

	return errBackend(c.setWithTTL(ctx, c.prefix+key, data, ttl))
}

// setWithTTL 探测 TTLSetter；无此能力的引擎回退到 Set（忽略 ttl）。
func (c *typedCache[T]) setWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if s, ok := c.engine.(TTLSetter); ok {
		return s.SetWithTTL(ctx, key, value, ttl)
	}
	return c.engine.Set(ctx, key, value)
}

// Del 实现 ICache——可用 BulkDelEngine 时用 DelMany，否则循环。
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

// Clear 实现 ICache——清空本独立实例的所有条目。
func (c *typedCache[T]) Clear(ctx context.Context) (err error) {
	defer c.recoverPanic(&err)

	return errBackend(c.engine.Clear(ctx, c.prefix))
}

// closeEngine 关闭底层引擎（供 Manager.Close 使用）。
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

// recoverPanic 捕获引擎调用的 panic 并转为 ErrBackend。
func (c *typedCache[T]) recoverPanic(err *error) {
	if r := recover(); r != nil {
		*err = errBackend(fmt.Errorf("engine panic: %v", r))
	}
}

var (
	_ ICache[string] = (*typedCache[string])(nil)
	_ engineCloser   = (*typedCache[string])(nil)
)
