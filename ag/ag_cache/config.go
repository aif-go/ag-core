package ag_cache

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
)

const (
	// AgCacheConfPrefix 是 FxAgCacheMode 绑定的配置根前缀。
	AgCacheConfPrefix = "agcache"
)

// AgCacheProperties 是 core 装配配置：仅按名选择引擎实现。
// 引擎特定参数与 TTL 由各引擎包持有（如 agcache.ristretto.*）。
//
// 无 value tag：ag_conf 按字段名大小写不敏感匹配配置键
// （DefaultEngine ↔ defaultEngine）。
type AgCacheProperties struct {
	// DefaultEngine 选择引擎实现名（如 "ristretto"）。
	DefaultEngine string
}

// DefaultAgCacheProperties 返回携带默认值的属性对象；
// binder 仅覆盖配置中存在的键。
func DefaultAgCacheProperties() *AgCacheProperties {
	return &AgCacheProperties{DefaultEngine: "ristretto"}
}

// BindAgCacheProperties 从配置 binder 绑定 agcache.*，
// 以 DefaultAgCacheProperties 为起点。
func BindAgCacheProperties(binder ag_conf.IBinder) (*AgCacheProperties, error) {
	props := DefaultAgCacheProperties()
	if err := binder.Bind(props, AgCacheConfPrefix); err != nil {
		return nil, err
	}
	return props, nil
}
