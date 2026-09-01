package agristretto

import (
	"fmt"
	"time"

	"github.com/aif-go/ag-core/ag/ag_conf"
)

const (
	// RistrettoConfPrefix 是引擎自身的配置前缀（由引擎绑定）。
	RistrettoConfPrefix = "agcache.ristretto"
)

// RistrettoConfigProperties 是引擎的 YAML 绑定配置模型。
// 无 value tag：ag_conf 按字段名大小写不敏感匹配配置键。
type RistrettoConfigProperties struct {
	// MaxCost 是内存预算（字节），0=引擎默认 100MB。
	MaxCost int64
	// NumCounters 是计数器数量（0=由 MaxCost 推导）。
	NumCounters int64
	// DefaultTTL 是引擎默认 TTL：""→默认 0（永不过期）；
	// "60s"→显式。"0" 与 "" 相同（永不过期）。
	DefaultTTL string
}

// DefaultRistrettoConfigProperties 返回携带默认值的属性对象；
// binder 仅覆盖配置中存在的键。
func DefaultRistrettoConfigProperties() *RistrettoConfigProperties {
	return &RistrettoConfigProperties{MaxCost: 100_000_000}
}

// BindRistrettoConfigProperties 从配置 binder 绑定 agcache.ristretto.*，
// 以 DefaultRistrettoConfigProperties 为起点。
func BindRistrettoConfigProperties(binder ag_conf.IBinder) (*RistrettoConfigProperties, error) {
	props := DefaultRistrettoConfigProperties()
	if err := binder.Bind(props, RistrettoConfPrefix); err != nil {
		return nil, err
	}
	return props, nil
}

// parseTTL 将 DefaultTTL 字符串转为时长。
// "" 或 "0" → 0（永不过期，引擎默认）；"60s" → 60s。
// 非法字符串报错。
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
