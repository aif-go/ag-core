package ag_cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Engine 是缓存引擎 SPI。它操作原始 []byte，不感知类型、序列化或 namespace。
// 每个 Engine 是独立实例；内存管理由引擎自身负责
// （Ristretto：MaxCost；Redis：服务端 maxmemory）。
//
// 所有方法接收 ctx 用于单次调用的超时/取消（网络引擎如 Redis 需要；
// 本地引擎可忽略）。
//
// 错误契约：
//   - Get 在 key 不存在时返回 ErrCacheMiss；任何其他错误表示后端故障
//     （调用方不应将其当作未命中）。
//   - Set/Del/Clear 成功返回 nil，后端故障返回错误。
//
// Set 不携带 ttl：默认 TTL 由引擎自身决定（其配置如 Ristretto defaultTtl，
// 或自然行为）。外部 TTL 经可选能力 TTLSetter 应用。ttl=0 表示永不过期。
type Engine interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Del(ctx context.Context, key string) error
	// Clear 清空本引擎实例的所有条目。prefix 携带 namespace key 前缀
	// （"agcache::<name>::"），供共享后端引擎（如 Redis SCAN+DEL）按前缀清理；
	// 本地引擎可忽略。
	Clear(ctx context.Context, prefix string) error
	Close() error
}

// cachePrefix 构建 namespace key 前缀 "agcache::<name>::"。
func cachePrefix(name string) string {
	return "agcache::" + name + "::"
}

// TTLSetter 是可选能力：支持显式 per-entry TTL 的引擎实现它。
// 业务代码经 ICache.SetWithTTL 或 WithDefaultTTL 指定 TTL；
// typedCache 探测该接口并委托给 SetWithTTL。
// 未实现 TTLSetter 的引擎忽略外部 TTL（等同 Set）。
type TTLSetter interface {
	SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

// BulkDelEngine 是可选能力：支持一次删除多个 key 的引擎实现它。
// typedCache.Del 探测该接口，可用时用 DelMany，否则回退到循环单 key Del。
type BulkDelEngine interface {
	DelMany(ctx context.Context, keys ...string) error
}

// EngineFactory 创建具名引擎实例。Create 接收缓存名作为
// namespace 上下文（引擎可忽略或按名查 per-name 配置，对齐 Spring
// createCache(name)）；每次 Create 返回全新独立实例——复用与生命周期归 Manager。
type EngineFactory interface {
	Name() string
	Create(name string) (Engine, error)
}

// syncer 是可选能力：异步写引擎实现它，使 GetOrElse 能保证
// 未命中加载的值在返回前可见。不属于 Engine SPI——纯可选接口。
type syncer interface{ Sync() }

// errBackend 将引擎错误包装为 ErrBackend。ErrCacheMiss 原样透传。
func errBackend(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrCacheMiss) {
		return err
	}
	return fmt.Errorf("%w: %v", ErrBackend, err)
}
