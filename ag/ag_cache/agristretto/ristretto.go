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

// RistrettoOptions 是运行时层——实际创建 Ristretto 实例的完整参数（已解析）。
type RistrettoOptions struct {
	MaxCost     int64         // 内存预算（字节）
	NumCounters int64         // TinyLFU 频率 sketch 大小
	BufferItems int64         // Ristretto 读缓冲环大小
	DefaultTTL  time.Duration // 引擎默认 TTL（0=永不过期）
}

// Validate 校验 Options 非负性。导出供手动构造 Options 的场景显式校验。
func (o RistrettoOptions) Validate() error {
	if o.MaxCost < 0 {
		return fmt.Errorf("agcache: MaxCost must be >= 0, got %d", o.MaxCost)
	}
	if o.NumCounters < 0 {
		return fmt.Errorf("agcache: NumCounters must be >= 0, got %d", o.NumCounters)
	}
	if o.BufferItems < 0 {
		return fmt.Errorf("agcache: BufferItems must be >= 0, got %d", o.BufferItems)
	}
	return nil
}

// String 实现 fmt.Stringer。
func (o RistrettoOptions) String() string {
	return fmt.Sprintf("RistrettoOptions{MaxCost=%d, NumCounters=%d, BufferItems=%d, DefaultTTL=%v}", o.MaxCost, o.NumCounters, o.BufferItems, o.DefaultTTL)
}

// ristrettoEngine 实现由单个 Ristretto 实例支撑的 ag_cache.Engine。
// 每个实例独立——无共享状态、无 key 索引。
type ristrettoEngine struct {
	cache      *ristretto.Cache[string, []byte]
	defaultTTL time.Duration // 引擎内部默认 TTL
}

// NewRistrettoEngine 从已解析 Options 创建本地引擎。
// 先 Validate（负值报错），再做零值兜底（0→默认/固定值，非 MaxCost 推导）。
func NewRistrettoEngine(opts RistrettoOptions) (ag_cache.Engine, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	if opts.MaxCost <= 0 {
		opts.MaxCost = defaultMaxCost
	}
	if opts.NumCounters <= 0 {
		opts.NumCounters = defaultNumCounters
	}
	if opts.BufferItems <= 0 {
		opts.BufferItems = defaultBufferItems
	}

	cache, err := ristretto.NewCache[string, []byte](&ristretto.Config[string, []byte]{
		NumCounters: opts.NumCounters,
		MaxCost:     opts.MaxCost,
		BufferItems: opts.BufferItems,
		Metrics:     false, // v3: Stats 后置，无统计消费，关闭 Metrics 省开销
	})
	if err != nil {
		return nil, err
	}
	return &ristrettoEngine{cache: cache, defaultTTL: opts.DefaultTTL}, nil
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

// agristrettoFactory 实现 ag_cache.EngineFactory。
// 装配期预解析全部配置（Default + 所有 Namespaces）缓存进 map；
// Create(name) 运行期纯查表（未命中用全局默认 ""）。
type agristrettoFactory struct {
	cfg  RistrettoConfigs
	opts map[string]RistrettoOptions // "" = 全局默认 + 每 per-name
}

// NewAgristrettoFactory 由绑定容器构造工厂。
// 启动期预解析校验：Default + 所有 Namespaces 逐一 ToOptions + Validate，
// 任一非法（含永不使用的 name）→ 返回 error（fail-fast，含 name 定位）。
func NewAgristrettoFactory(cfg *RistrettoConfigs) (ag_cache.EngineFactory, error) {
	if cfg == nil {
		cfg = DefaultRistrettoConfigs()
	}
	opts := make(map[string]RistrettoOptions, 1+len(cfg.Namespaces))

	def, err := cfg.Default.ToOptions()
	if err != nil {
		return nil, err
	}
	if err := def.Validate(); err != nil {
		return nil, err
	}
	opts[""] = def

	for name, nc := range cfg.Namespaces {
		merged := mergeConfig(cfg.Default, nc)
		o, err := merged.ToOptions()
		if err != nil {
			return nil, fmt.Errorf("agcache: namespace %q: %w", name, err)
		}
		if err := o.Validate(); err != nil {
			return nil, fmt.Errorf("agcache: namespace %q: %w", name, err)
		}
		opts[name] = o
	}
	return agristrettoFactory{cfg: *cfg, opts: opts}, nil
}

// Name 返回注册的引擎名。
func (f agristrettoFactory) Name() string { return "ristretto" }

// Create 查预解析 map 构建引擎：命中 name 用该配置，未命中用全局默认。
// 每个 name 独立实例。运行期零解析零报错。
func (f agristrettoFactory) Create(name string) (ag_cache.Engine, error) {
	o, ok := f.opts[name]
	if !ok {
		o = f.opts[""]
	}
	return NewRistrettoEngine(o)
}

var _ ag_cache.EngineFactory = agristrettoFactory{}
