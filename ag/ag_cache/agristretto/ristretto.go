// Package agristretto 提供 ag_cache 的 Ristretto 本地引擎。
// 这是唯一允许 import Ristretto 的包。
package agristretto

import (
	"context"
	"fmt"
	"time"

	"github.com/aif-go/ag-core/ag/ag_cache"
	"github.com/dgraph-io/ristretto/v2"
)

// RistrettoConfig 是引擎自身的配置。MaxCost 在这里——
// 内存管理是引擎的责任。
type RistrettoConfig struct {
	MaxCost     int64
	NumCounters int64
}

// DefaultRistrettoConfig 返回 100MB / 10M-counter 的默认配置。
func DefaultRistrettoConfig() RistrettoConfig {
	return RistrettoConfig{MaxCost: 100_000_000, NumCounters: 10_000_000}
}

// String 实现 fmt.Stringer。
func (c RistrettoConfig) String() string {
	return fmt.Sprintf("RistrettoConfig{MaxCost=%d, NumCounters=%d}", c.MaxCost, c.NumCounters)
}

// ristrettoEngine 实现由单个 Ristretto 实例支撑的 ag_cache.Engine。
// 每个实例独立——无共享状态、无 key 索引。
type ristrettoEngine struct {
	cache      *ristretto.Cache[string, []byte]
	defaultTTL time.Duration // 引擎内部默认 TTL（配置 defaultTtl）
}

// NewRistrettoEngine 从配置创建本地引擎。
// 零值回退到默认（MaxCost）或推导（NumCounters）。
func NewRistrettoEngine(cfg RistrettoConfig) (ag_cache.Engine, error) {
	return newRistrettoEngine(cfg, 0)
}

func newRistrettoEngine(cfg RistrettoConfig, defaultTTL time.Duration) (ag_cache.Engine, error) {
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
	return &ristrettoEngine{cache: cache, defaultTTL: defaultTTL}, nil
}

// Get 命中返回 (data, nil)，未命中返回 (nil, ag_cache.ErrCacheMiss)。
func (e *ristrettoEngine) Get(ctx context.Context, key string) ([]byte, error) {
	v, ok := e.cache.Get(key)
	if !ok {
		return nil, ag_cache.ErrCacheMiss
	}
	return v, nil
}

// Set 使用引擎内部默认 TTL。
// Set 是异步的（无 Wait）。cost 由值字节长度内部计算——
// 泛型 SPI 不携带 cost 概念。
func (e *ristrettoEngine) Set(ctx context.Context, key string, value []byte) error {
	return e.setWithTTL(key, value, e.defaultTTL)
}

// SetWithTTL 实现 ag_cache.TTLSetter——显式 per-entry TTL。
func (e *ristrettoEngine) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return e.setWithTTL(key, value, ttl)
}

func (e *ristrettoEngine) setWithTTL(key string, value []byte, ttl time.Duration) error {
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

// Sync 实现 ag_cache.syncer——阻塞直到待写写入对读可见。
func (e *ristrettoEngine) Sync() { e.cache.Wait() }

// Del 实现 ag_cache.Engine。
func (e *ristrettoEngine) Del(ctx context.Context, key string) error {
	e.cache.Del(key)
	return nil
}

// Clear 实现 ag_cache.Engine——清空本独立实例，忽略 prefix。
func (e *ristrettoEngine) Clear(ctx context.Context, prefix string) error {
	e.cache.Clear()
	return nil
}

// Close 实现 ag_cache.Engine。
func (e *ristrettoEngine) Close() error {
	e.cache.Close()
	return nil
}

var _ ag_cache.Engine = (*ristrettoEngine)(nil)

// ──────── Engine factory ────────

// agristrettoFactory 实现 ag_cache.EngineFactory，持有引擎配置
// 与引擎声明的默认 TTL（自包含）。
type agristrettoFactory struct {
	cfg RistrettoConfig
	ttl time.Duration
}

// Name 返回注册的引擎名。
func (f agristrettoFactory) Name() string { return "ristretto" }

// Create 从工厂持有的配置构建引擎，用引擎声明的默认值
// 播种引擎内部默认 TTL。name 是 namespace 上下文
// （agristretto 忽略——每个缓存名获取全新实例）。
func (f agristrettoFactory) Create(name string) (ag_cache.Engine, error) {
	return newRistrettoEngine(f.cfg, f.ttl)
}

var _ ag_cache.EngineFactory = agristrettoFactory{}
