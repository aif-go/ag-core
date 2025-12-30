package agmetadata

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type agHeadKey struct{}

// var mdKeys = []string{}
var (
	mdKeys     = make(map[string]struct{}) // 用于去重检查
	mdKeysLock sync.RWMutex                // 保护并发访问
	keysCache  atomic.Value                // 存储 []string
)

// 初始化时设置空缓存
func init() {
	keysCache.Store([]string{})
}

// ParseFunc 解析函数，用于从上下文中提取元数据, 返回值为元数据值和是否存在的标志
type ParseFunc func(string) ([]string, bool)

// 获取函数，用于从元数据中获取数据，提供处理
type HandlerFunc func(k, v string)

// MD is a mapping from metadata keys to values. Users should use the following
// two convenience functions New and Pairs to generate MD.
type MD map[string]string

// New creates an MD from a given key-value map.
func New(m map[string]string) MD {
	md := MD{}
	for k, val := range m {
		// key := strings.ToLower(k)
		key := k
		md[key] = val
	}
	return md
}

// Pairs returns an MD from a list of key-value pairs.
// It error if an odd number of strings are provided.
func Pairs(kv ...string) (MD, error) {
	if len(kv)%2 == 1 {
		return nil, fmt.Errorf("metadata: Pairs got the odd number of input pairs for metadata: %d", len(kv))
	}
	md := MD{}
	var key string
	for i, s := range kv {
		if i%2 == 0 {
			// key = strings.ToLower(s)
			key = s
			continue
		}
		md[key] = s
	}
	return md, nil
}

// Len returns the number of items in md.
func (md MD) Len() int {
	return len(md)
}

// Copy returns a copy of md.
func (md MD) Copy() MD {
	result := make(MD, len(md))
	for k, v := range md {
		result[k] = v
	}
	return result
}

// Get obtains the values for a given key.
func (md MD) Get(k string) string {
	// k = strings.ToLower(k)
	return md[k]
}

// Set sets the value of a given key with a slice of values.
func (md MD) Set(k string, val string) {
	if len(val) == 0 {
		return
	}
	// k = strings.ToLower(k)
	md[k] = val
}

// AppendMdToContext 将新的元数据合并到上下文中
func AppendMdToContext(ctx context.Context, newmd MD) context.Context {
	// 从上下文提取已存在头信息
	md, _ := ctx.Value(agHeadKey{}).(MD)
	rctx := ctx

	// 将新的头信息合并到已存在头信息中
	if md == nil {
		md = make(MD, len(newmd))
		// 将合并后的头信息设置到上下文
		rctx = context.WithValue(ctx, agHeadKey{}, md)
	}
	for k, v := range newmd {
		md[k] = v
	}

	return rctx
}

// GetMdFromContext 从上下文中提取元数据
func GetMdFromContext(ctx context.Context) MD {
	if rmd, ok := ctx.Value(agHeadKey{}).(MD); ok {
		// 复制元信息给外部访问，防止外部直接修改
		return rmd.Copy()
	}
	return MD{}
}

// GetMdValueFromContext 从上下文中提取元数据值
func GetValueFromContext(ctx context.Context, key string) (string, bool) {
	if rmd, ok := ctx.Value(agHeadKey{}).(MD); ok {
		if v, ok := rmd[key]; ok {
			return v, true
		}
	}
	return "", false
}

// ParseMdToContext 解析上下文中的元数据，将其应用到新的元数据中
func ParseMdToContext(ctx context.Context, parse ParseFunc) (context.Context, error) {
	newmd := MD{}
	mdKeys := GetMdKeys()
	for _, k := range mdKeys {
		if vv, ok := parse(k); ok {
			vvlen := len(vv)
			if vvlen == 1 {
				// 当v长度等于1时，v中的数据就是元数据值
				newmd[k] = vv[0]
			} else if vvlen > 1 {
				// 当vv长度大于1时，v中的数据就是k,v模式，前一项是后一项的key
				tmd, err := Pairs(vv...)
				if err != nil {
					return nil, err
				}
				AppendMD(newmd, tmd)
			}
		}
	}
	return AppendMdToContext(ctx, newmd), nil
}

// HandlerMdFromContext 遍历 metadata 并对每个键值对调用处理函数
func HandlerMdFromContext(ctx context.Context, handler HandlerFunc) error {
	// 从上下文提取头信息
	md := GetMdFromContext(ctx)
	for key, value := range md {
		handler(key, value)
	}
	return nil
}

// AppendMD appends other into md, merging values of the same key.
func AppendMD(md, other MD) MD {
	if md == nil {
		md = make(MD, len(other))
	}
	for k, v := range other {
		md[k] = v
	}
	return md
}

// RegMdKey 注册元数据键名
func RegMdKey(key string) {
	// mdKeys = append(mdKeys, key)
	mdKeysLock.Lock()
	defer mdKeysLock.Unlock()

	if _, exists := mdKeys[key]; !exists {
		mdKeys[key] = struct{}{}
		// 更新缓存
		keys := make([]string, 0, len(mdKeys))
		for k := range mdKeys {
			keys = append(keys, k)
		}
		keysCache.Store(keys)
	}
}

// GetMdKeys 获取已注册的元数据键名
func GetMdKeys() []string {
	// return mdKeys
	if cached := keysCache.Load(); cached != nil {
		return cached.([]string)
	}
	return []string{}
}
