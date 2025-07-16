package ag_conf

import (
	"fmt"
	"log/slog"
)

// PropertySourcesPropertyResolver 属性解析器
type PropertySourcesPropertyResolver struct {
	// 嵌入抽象实现
	AbstractPropertyResolver

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
	apr.AbstractPropertyResolver.GetPropertyAsRawString = apr.GetPropertyAsRawString

	apr.PropertySources = propertySources

	return apr
}

// GetProperty impl IPropertyResolver.GetProperty
// 实例化时要重写赋值给AbstractPropertyResolver.GetProperty
// TODO 需要对外抛出error
func (pspr *PropertySourcesPropertyResolver) GetProperty(key string) string {
	v, err := pspr.getProperty(key, true)
	if err != nil {
		slog.Error("Error resolving property for key '"+key+"'", "err", err)
	}
	return v
}

// GetRequiredProperty impl IPropertyResolver.GetRequiredProperty
func (pspr *PropertySourcesPropertyResolver) GetRequiredProperty(key string) (string, error) {
	v, err := pspr.getProperty(key, true)
	return v, err
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

/*
	#### 自定义实现 ####
*/

/**
* getProperty 自定义实现获取属性值
*
* @param key 属性的键名
* @param resolveNestedPlaceholders 是否解析嵌套占位符
* @return string 返回对应的属性值，找不到时返回空字符串
**/
func (pspr *PropertySourcesPropertyResolver) getProperty(key string, resolveNestedPlaceholders bool) (string, error) {
	if pspr.PropertySources != nil {
		pslist := pspr.PropertySources.GetPropertySources()
		for _, ps := range pslist {
			if slog.Default().Enabled(nil, slog.LevelDebug) {
				slog.Debug("Searching for key '" + key + "' in PropertySource '" + ps.GetName() + "'")
			}

			v := ps.GetProperty(key) // TODO v是否存在非string场景
			if v != nil {
				v2, ok := v.(string) // TODO  v 目前都是string
				if ok && resolveNestedPlaceholders {
					// v2 = pspr.ResolvePlaceholders(v2)
					v2, err := pspr.ResolveNestedPlaceholders(v2)
					if err != nil {
						// slog.Error("Error resolving nested placeholders for key '"+key+"' in PropertySource '"+ps.GetName()+"'", err)
						slog.Error(fmt.Sprintf("Error resolving nested placeholders for key '%s' in PropertySource '%s'", key, ps.GetName()), "err", err)
						return "", err // TODO 异常处理
					}
					v = v2
				}

				logKeyFound(key, ps, v)

				return v.(string), nil
			}
		}
	}

	if slog.Default().Enabled(nil, slog.LevelDebug) {
		slog.Debug("Could not find key '" + key + "' in any property source")
	}

	return "", nil
}

func logKeyFound(key string, ps IPropertySource, value any) {
	// slog.Debug("Found key '" + key + "' in PropertySource '" + ps.GetName())
	if slog.Default().Enabled(nil, slog.LevelDebug) {
		slog.Debug(fmt.Sprintf("Found key '%s' in PropertySource '%s'", key, ps.GetName()))
	}
}

func (pspr *PropertySourcesPropertyResolver) GetPropertyAsRawString(key string) string {
	v, _ := pspr.getProperty(key, false)
	return v
}
