// Package ag_cache 提供 ag-core 的通用缓存抽象。
// 业务代码通过 ICache[T] 使用缓存，不感知框架细节；
// 引擎实现在子包（如 agristretto），经 EngineFactory 通过 fx group 注入。
package ag_cache

import (
	"context"
	"errors"
	"time"
)

// 缓存操作的哨兵错误。
var (
	// ErrCacheMiss 是 Get/GetOrElse 未命中时返回的错误。
	// 它是唯一会触发 loader 调用（读穿透）的错误。
	ErrCacheMiss = errors.New("agcache: key not found")
	// ErrBackend 包装任何非未命中的后端引擎错误。
	// 业务代码可用 errors.Is(err, agcache.ErrBackend) 判断。
	ErrBackend = errors.New("agcache: backend engine error")
)

// LoaderFunc 是 GetOrElse 在缓存未命中时调用的加载函数。
// 同时接收 ctx 和 key，便于 loader 跨 key 复用，
// 也便于降级装饰器重建原始查询。
type LoaderFunc[T any] func(ctx context.Context, key string) (T, error)

// ICache 是面向业务代码的泛型缓存接口。
// 刻意保持精简：只含 90% 调用点用到的方法。
type ICache[T any] interface {
	// Get 从缓存读取。命中返回 (value, nil)；未命中返回 (zero, ErrCacheMiss)。
	Get(ctx context.Context, key string) (T, error)

	// TryGet 是无未命中错误的宽松读：
	// 命中返回 (value, true, nil)；未命中 (zero, false, nil)；后端故障 (zero, false, err)。
	// 用于存在性检查与探活（如监控预期命中）。
	TryGet(ctx context.Context, key string) (T, bool, error)

	// GetOrElse 从缓存读取；未命中时调用 loader，写入缓存后返回。
	GetOrElse(ctx context.Context, key string, loader LoaderFunc[T]) (T, error)

	// Set 写入值。TTL 用 namespace 默认（WithDefaultTTL 或引擎内部默认）。
	Set(ctx context.Context, key string, value T) error

	// SetWithTTL 以显式 per-entry TTL 写入（TTL 链中最高优先级：
	// SetWithTTL > WithDefaultTTL > 引擎内部默认）。
	// ttl=0 表示永不过期。引擎无 TTLSetter 时等同 Set。
	SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Clear(ctx context.Context) error
}

// Stats 保存缓存指标（为未来 StatsProvider 预留；不属于 Engine）。
type Stats struct {
	Hits       int64
	Misses     int64
	Evictions  int64
	EntryCount int64
}
