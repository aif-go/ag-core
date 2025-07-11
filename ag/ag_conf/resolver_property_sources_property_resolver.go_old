package ag_conf

import (
	"fmt"
	"log/slog"
	"strings"
)

// PropertySourcesPropertyResolver 属性解析器
type PropertySourcesPropertyResolver struct {
	// 嵌入抽象实现
	AbstractPropertyResolver
	// PlaceholderPrefix                    string
	// PlaceholderSuffix                    string
	// ValueSeparator                       string
	// IgnoreUnresolvableNestedPlaceholders bool
	// // 必须的属性key集合
	// RequiredProperties []string

	// 属性源集合
	PropertySources IPropertySources
}

// NewPropertySourcesPropertyResolver 构造函数
func NewPropertySourcesPropertyResolver(propertySources IPropertySources) *PropertySourcesPropertyResolver {
	apr := &PropertySourcesPropertyResolver{
		AbstractPropertyResolver: AbstractPropertyResolver{
			PlaceholderPrefix:                    ConstPlaceholderPrefix,
			PlaceholderSuffix:                    ConstPlaceholderSuffix,
			ValueSeparator:                       ConstValueSeparator,
			IgnoreUnresolvableNestedPlaceholders: false,
		},
	}
	// 初始化获取属性的方法
	apr.AbstractPropertyResolver.GetProperty = apr.GetProperty

	apr.PropertySources = propertySources

	return apr
}

// GetProperty impl IPropertyResolver.GetProperty 具体实现，支持el表达式的使用,el表达式的纵向深度目前限制3,超过3不会解析到值
// 实例化时要重写赋值给AbstractPropertyResolver.GetProperty
// TODO 需要对外抛出error
func (pspr *PropertySourcesPropertyResolver) GetProperty(key string) string {
	// maxDeepLengthCheck := getPropertySupportElClosure()
	// 此处传递的第四个参数无实际意义,只是因为函数的参数结构需要
	val, err := getPropertySupportElByMaxDeepLenth(key, pspr, nil, true)
	if err != nil {
		slog.Error("getPropertySupportElByMaxDeepLenth error", "err", err)
		return ""
	}
	return val
}

// doGetProperty 实际获取env中根据key获取对应的值的内容
func doGetProperty(key string, pspr *PropertySourcesPropertyResolver) string {

	value := ""
	if pspr.PropertySources != nil {
		err := pspr.PropertySources.RangePropertySourceHandler(func(ps IPropertySource) (end bool, err error) {
			if slog.Default().Enabled(nil, slog.LevelDebug) {
				slog.Debug("Searching for key '" + key + "' in PropertySource '" + ps.GetName() + "'")
			}

			// 遍历TODO
			v := ps.GetProperty(key)
			if v != nil {
				// TODO:
				// 1. 是否需要解析嵌套占位符
				//if 需要解析 {
				//	pspr.ResolvePlaceholders(v.(string))
				//}
				// 2. 类型转换
				// 3. 记录日志
				value = v.(string) // TODO 需要类型推断，value是否有用

				return true, nil // 已经找到，结束遍历
			}
			return false, nil // 继续遍历
		})
		if err != nil {

		}
	}
	if slog.Default().Enabled(nil, slog.LevelDebug) {
		if value == "" {
			slog.Debug("Could not find key '" + key + "' in any property source")
		}
	}

	return value
}

// getPropertySupportElByMaxDeepLenth 对于el表达式的key做el解析和深度控制
// dd-->${aa}_${bb}_${cc} 会获取 aa bb cc对应key的值然后拼接为新key重新获取值
// 单个el表达式 例如 ${aa} 最多底层包含三层el,超过${aa}将返回blank
func getPropertySupportElByMaxDeepLenth(key string, pspr *PropertySourcesPropertyResolver, maxDeepLengthCheck func(key string) error, single bool) (string, error) {
	var err error
	if CheckEL(key) {
		subkeys, err := MulEl(key)
		if err != nil {
			return "", err
		}
		newkeylist := []string{}
		// 遍历subkey
		for _, subkey := range subkeys {
			val := subkey
			if CheckEL(subkey) {
				// 单个key解析el表达式的过程中可能涉及到递归，目前单个key最多递归三层
				val, maxDeepLengthCheck, err = parseEL(subkey, pspr, maxDeepLengthCheck)
				// 单个key的场景,此时直接返回
				if single {
					return val, err
				}
			}
			// 单个key的场景不做拼接处理,跳过
			if single {
				continue
			}
			// 组合key要根据获取的到的值重新拼装key之后再次获取
			if !single {
				newkeylist = append(newkeylist, val)
			}
		}
		newkey := strings.Join(newkeylist, "")
		return doGetProperty(newkey, pspr), err
	} else {
		value := doGetProperty(key, pspr)
		if CheckEL(value) {
			if maxDeepLengthCheck == nil {
				maxDeepLengthCheck = getPropertySupportElClosure()
			}
			return getPropertySupportElByMaxDeepLenth(value, pspr, maxDeepLengthCheck, singleOrNot(value))
		}
		// 说明key纵向处理完成，此时重置闭包的逻辑,保证组合key的下一个key的深度计算
		maxDeepLengthCheck = nil
		return value, err
	}
}

func parseEL(subkey string, pspr *PropertySourcesPropertyResolver, maxDeepLengthCheck func(key string) error) (res string, maxDeepFunc func(key string) error, parseErr error) {
	newkey, defaultVal := EliminatePlaceholder(subkey)
	if maxDeepLengthCheck == nil {
		maxDeepLengthCheck = getPropertySupportElClosure()
	}
	err := maxDeepLengthCheck(subkey)
	var val string
	if err != nil {
		// 当深度超过3的时候,就返回当前对应的key
		val = ""
		slog.Warn("获取env中el的值超过纵向深度2,无法获取其值", "key", subkey)
		maxDeepLengthCheck = nil
		return "", maxDeepLengthCheck, err
	} else {
		val = doGetProperty(newkey, pspr)
		if CheckEL(val) {
			val, maxDeepLengthCheck, err = parseEL(val, pspr, maxDeepLengthCheck)
		}
		// 递归调用
		if val == "" {
			val = defaultVal
		}
		maxDeepLengthCheck = nil
	}
	return val, maxDeepLengthCheck, err
}

// getPropertySupportElClosure 对EL占位符深度的检查
func getPropertySupportElClosure() func(key string) error {
	maxDeepLength := 1
	return func(key string) error {
		if maxDeepLength > 2 {
			return fmt.Errorf("当前占位符的处理已经超过最大深度2")
		}
		if CheckEL(key) {
			maxDeepLength++
		}
		return nil
	}

}

/*
	虽然AbstractPropertyResolver已经实现了ContainsProperty，这里仍然可以重新实现，类似于java的重写
*/
// ContainsProperty impl IPropertyResolver.ContainsProperty
func (pspr *PropertySourcesPropertyResolver) ContainsProperty(key string) bool {
	if pspr.PropertySources != nil {
		pslist := pspr.PropertySources.GetPropertySources()
		for _, ps := range pslist {
			if ps.ContainsProperty(key) {
				return true
			}
		}
	}
	return false
}
