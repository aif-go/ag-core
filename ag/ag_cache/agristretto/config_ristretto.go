package agristretto

import (
	"fmt"
	"time"

	"github.com/aif-go/ag-core/ag/ag_conf"
)

const (
	// RistrettoConfPrefix 是引擎自身的配置前缀（由引擎绑定）。
	RistrettoConfPrefix = "agcache.ristretto"

	defaultMaxCost     = 100_000_000 // 100MB 内容预算
	defaultNumCounters = 131_072     // 2^17 = 100K 档（sketch ~0.26MB）
	defaultBufferItems = 64
)

// RistrettoConfig 是引擎绑定层叶子——单个 name 的 YAML 绑定配置。
// 无 value tag：ag_conf 按字段名大小写不敏感匹配配置键。
// DefaultTTL 保留 string：经 parseTTL 严格解析（纯数字报错，防纳秒陷阱）。
type RistrettoConfig struct {
	// MaxCost 是内存预算（字节），0=默认 100MB。
	MaxCost int64
	// NumCounters 是 TinyLFU 频率 sketch 大小（影响淘汰精度、非条数上限），
	// 0=默认 131072（2^17，100K 档）。预分配内存随 NumCounters 线性。
	NumCounters int64
	// BufferItems 是 Ristretto 读缓冲环大小，0=默认 64。
	BufferItems int64
	// DefaultTTL 是引擎默认 TTL：""或"0"→永不过期；"60s"→60 秒；非法→ToOptions 报错。
	DefaultTTL string
}

// RistrettoConfigs 是绑定层容器——全局默认 + per-name 覆盖。
// Default 为全局限量默认；Namespaces 按缓存名覆盖（键如 "users"）。
// 叶子与容器分离，避免 map[string]RistrettoConfig 递归嵌套。
type RistrettoConfigs struct {
	Default    RistrettoConfig
	Namespaces map[string]RistrettoConfig
}

// DefaultRistrettoConfig 显式给出全部四字段默认值（YAML 缺省起点）。
func DefaultRistrettoConfig() RistrettoConfig {
	return RistrettoConfig{MaxCost: defaultMaxCost, NumCounters: defaultNumCounters, BufferItems: defaultBufferItems, DefaultTTL: "0"}
}

// DefaultRistrettoConfigs 返回携带默认配置的空容器（Namespaces 空 map）。
func DefaultRistrettoConfigs() *RistrettoConfigs {
	return &RistrettoConfigs{Default: DefaultRistrettoConfig(), Namespaces: map[string]RistrettoConfig{}}
}

// BindRistrettoConfig 从配置 binder 绑定 agcache.ristretto.* 容器，
// 以 DefaultRistrettoConfig() 为默认起点，Namespaces 缺省为空 map。
func BindRistrettoConfig(binder ag_conf.IBinder) (*RistrettoConfigs, error) {
	cfg := DefaultRistrettoConfigs()
	if err := binder.Bind(cfg, RistrettoConfPrefix); err != nil {
		return nil, err
	}
	return cfg, nil
}

// mergeConfig 非零覆盖继承：per-name 指定字段（非零/非空）覆盖 Default，
// 未指定字段继承 Default。
func mergeConfig(def, nc RistrettoConfig) RistrettoConfig {
	if nc.MaxCost != 0 {
		def.MaxCost = nc.MaxCost
	}
	if nc.NumCounters != 0 {
		def.NumCounters = nc.NumCounters
	}
	if nc.BufferItems != 0 {
		def.BufferItems = nc.BufferItems
	}
	if nc.DefaultTTL != "" {
		def.DefaultTTL = nc.DefaultTTL
	}
	return def
}

// ToOptions 将绑定配置转换为运行时 Options：
// 非负校验 + parseTTL + 固定默认填充（不做 MaxCost×10% 推导，避免 next2Power 跳档浪费）。
func (c RistrettoConfig) ToOptions() (RistrettoOptions, error) {
	if c.MaxCost < 0 || c.NumCounters < 0 || c.BufferItems < 0 {
		return RistrettoOptions{}, fmt.Errorf("agcache: negative value in RistrettoConfig (MaxCost=%d NumCounters=%d BufferItems=%d)", c.MaxCost, c.NumCounters, c.BufferItems)
	}
	ttl, err := parseTTL(c.DefaultTTL)
	if err != nil {
		return RistrettoOptions{}, err
	}
	opts := RistrettoOptions{DefaultTTL: ttl}
	if c.MaxCost <= 0 {
		opts.MaxCost = defaultMaxCost
	} else {
		opts.MaxCost = c.MaxCost
	}
	if c.NumCounters <= 0 {
		opts.NumCounters = defaultNumCounters
	} else {
		opts.NumCounters = c.NumCounters
	}
	if c.BufferItems <= 0 {
		opts.BufferItems = defaultBufferItems
	} else {
		opts.BufferItems = c.BufferItems
	}
	return opts, nil
}

// parseTTL 将 DefaultTTL 字符串转为时长。
// "" 或 "0" → 0（永不过期，引擎默认）；"60s" → 60s。
// 非法字符串（含纯数字 "60"，防纳秒陷阱）报错。
func parseTTL(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("agcache: invalid defaultTtl %q: %w", s, err)
	}
	return d, nil
}
