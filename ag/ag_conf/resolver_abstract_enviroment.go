package ag_conf

import (
	"os"
	"strings"
)

// IAbstractEnvironment Dep 抽象环境接口 AbstractEnvironment无法实现customizePropertySources方法, 所以需要自定义接口描述
// Deprecated: xxx
type IAbstractEnvironment interface {
	IConfigurableEnvironment
	customizePropertySources(ps *MutablePropertySources) error
}

// AbstractEnvironment 实现ConfigurableEnvironment接口
type AbstractEnvironment struct {
	PropertySources  *MutablePropertySources       // 属性源
	PropertyResolver IConfigurablePeopertyResolver // 属性解析器，默认实现为PropertySourcesPropertyResolver
}

// ---------------------------------------------------------------------
// Implementation of IPropertyResolver interface
// ---------------------------------------------------------------------

// ContainsProperty impl IPropertyResolver.ContainsProperty
func (e *AbstractEnvironment) ContainsProperty(key string) bool {
	return e.PropertyResolver.ContainsProperty(key)
}

// GetProperty impl IPropertyResolver.GetProperty
func (e *AbstractEnvironment) GetProperty(key string) string {
	return e.PropertyResolver.GetProperty(key)
}

// GetPropertyDefault impl IPropertyResolver.GetPropertyDefault
func (e *AbstractEnvironment) GetPropertyDefault(key string, defaultValue string) string {
	return e.PropertyResolver.GetPropertyDefault(key, defaultValue)
}

// GetRequiredProperty impl IPropertyResolver.GetRequiredProperty
func (e *AbstractEnvironment) GetRequiredProperty(key string) (string, error) {
	return e.PropertyResolver.GetRequiredProperty(key)
}

// ResolvePlaceholders impl IPropertyResolver.ResolvePlaceholders
func (e *AbstractEnvironment) ResolvePlaceholders(text string) string {
	return e.PropertyResolver.ResolvePlaceholders(text)
}

// ResolveRequiredPlaceholders impl IPropertyResolver.ResolveRequiredPlaceholders
func (e *AbstractEnvironment) ResolveRequiredPlaceholders(text string) (string, error) {
	return e.PropertyResolver.ResolveRequiredPlaceholders(text)
}

// ---------------------------------------------------------------------
// Implementation of IEnvironment interface
// ---------------------------------------------------------------------
/*
// GetActiveProfiles impl IEnvironment.GetActiveProfiles
func (e *AbstractEnvironment) GetDefaultProfiles() []string {
	// TODO 暂无使用 暂时不实现
	return nil
}
*/

// ---------------------------------------------------------------------
// Implementation of IConfigurablePeopertyResolver interface
// ---------------------------------------------------------------------

// SetPlaceholdPrefix impl IConfigurablePeopertyResolver.SetPlaceholdPrefix
func (e *AbstractEnvironment) SetPlaceholdPrefix(placeholderPrefic string) {
	e.PropertyResolver.SetPlaceholdPrefix(placeholderPrefic)
}

// SetPlaceholdSuffix impl IConfigurablePeopertyResolver.SetPlaceholdSuffix
func (e *AbstractEnvironment) SetPlaceholdSuffix(placeholderSuffix string) {
	e.PropertyResolver.SetPlaceholdSuffix(placeholderSuffix)
}

// SetValueSeparator impl IConfigurablePeopertyResolver.SetValueSeparator
func (e *AbstractEnvironment) SetValueSeparator(valueSeparator string) {
	e.PropertyResolver.SetValueSeparator(valueSeparator)
}

// SetIgnoreUnresolvableNestedPlaceholders impl IConfigurablePeopertyResolver.SetIgnoreUnresolvableNestedPlaceholders
func (e *AbstractEnvironment) SetIgnoreUnresolvableNestedPlaceholders(ignoreUnresolvableNestedPlaceholders bool) {
	e.PropertyResolver.SetIgnoreUnresolvableNestedPlaceholders(ignoreUnresolvableNestedPlaceholders)
}

// SetRequiredProperties impl IConfigurablePeopertyResolver.SetRequiredProperties
func (e *AbstractEnvironment) SetRequiredProperties(requiredProperties ...string) {
	e.PropertyResolver.SetRequiredProperties(requiredProperties...)
}

// ValidateRequiredProperties impl IConfigurablePeopertyResolver.ValidateRequiredProperties
func (e *AbstractEnvironment) ValidateRequiredProperties() error {
	return e.PropertyResolver.ValidateRequiredProperties()
}

// ---------------------------------------------------------------------
// Implementation of IConfigurableEnvironment interface
// ---------------------------------------------------------------------

// GetPropertySources impl IConfigurableEnvironment.GetPropertySources
func (e *AbstractEnvironment) GetPropertySources() *MutablePropertySources {
	// func (e *AbstractEnvironment) GetPropertySources() IPropertySources {
	return e.PropertySources
}

// Merge impl IConfigurableEnvironment.Merge
func (e *AbstractEnvironment) Merge(parent IConfigurableEnvironment) {
	// TODO 抽象实现
}

// GetSystemProperties 解析命令行参数，返回map[string]string，格式：`-Dkey[=value/true]`
func (e *AbstractEnvironment) GetSystemProperties() map[string]any {
	if len(os.Args) == 0 {
		return nil
	}

	option := "-D"
	if s := strings.TrimSpace(os.Getenv(CommandArgsPrefix)); s != "" {
		option = s
	}

	props := make(map[string]any)
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, option) {
			kv := strings.TrimPrefix(arg, option)
			ss := strings.SplitN(kv, "=", 2)
			k, v := ss[0], "true"
			if len(ss) > 1 {
				v = ss[1]
			}
			propKey := strings.ToLower(replaceKey(k))
			props[propKey] = v
		}
	}
	return props
}

// GetSystemEnvironment impl IConfigurableEnvironment.GetSystemEnvironment
func (e *AbstractEnvironment) GetSystemEnvironment() map[string]any {
	environ := os.Environ()
	// envmap := make(map[string]any, len(environ))
	envmap := make(map[string]any)
	for _, env := range environ {
		ss := strings.SplitN(env, "=", 2)
		k, v := ss[0], ""
		if len(ss) > 1 {
			v = ss[1]
		}
		propKey := k
		propKey = strings.ToLower(replaceKey(propKey))
		envmap[propKey] = v
	}
	return envmap
}

// GetNacosProperty 获取nacos配置的资源
// func GetNacosProperty(context string) map[string]any {

// 	// 提供nacos config client相关的配置
// 	// 获取本地文件中的对于nacos的配置
// 	// 找到DataId 遍历,然后拉取文件,增加对应的监听器

// 	return nil
// }

// replaceKey replace '_' with '.'
func replaceKey(s string) string {
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] == '_' {
			b[i] = '.'
		}
	}
	return string(b)
}
