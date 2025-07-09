package ag_conf

import "testing"

// 用于测试类型的接口实现情况
var _IPropertyResolver IPropertyResolver
var _IEnvironment IEnvironment // Deprecated: 精简移除IEnvironment
var _IConfigurablePeopertyResolver IConfigurablePeopertyResolver
var _IConfigurableEnvironment IConfigurableEnvironment

var _IAbstractEnvironment IAbstractEnvironment // Deprecated: 精简移除IAbstractEnvironment

var _AbstractEnvironment *AbstractEnvironment
var _StandardEnvironment *StandardEnvironment
var _AbstractPropertyResolver *AbstractPropertyResolver
var _PropertySourcesPropertyResolver *PropertySourcesPropertyResolver

func TestResoverApiCheck(t *testing.T) {
	if false {
		// --- 直接组合关系(类似继承) ---
		_IPropertyResolver = _IEnvironment
		_IPropertyResolver = _IConfigurablePeopertyResolver
		// _IEnvironment = _IConfigurableEnvironment
		_IConfigurablePeopertyResolver = _IConfigurableEnvironment
		_IConfigurableEnvironment = _IAbstractEnvironment

		// --- 间接 ---
		_IPropertyResolver = _IConfigurableEnvironment
		_IPropertyResolver = _IAbstractEnvironment
		_IConfigurablePeopertyResolver = _IAbstractEnvironment
		// _IEnvironment = _IAbstractEnvironment
	}
}

func TestResolverCheck(t *testing.T) {
	if false {
		// --- 直接关系 ---
		// TODO AbstractPropertyResolver 实现 IConfigurablePeopertyResolver
		// _IConfigurablePeopertyResolver = _AbstractPropertyResolver

		// PropertySourcesPropertyResolver 组合了 AbstractPropertyResolver
		_PropertySourcesPropertyResolver.AbstractPropertyResolver = *_AbstractPropertyResolver

		// AbstractEnvironment 实现了 IConfigurableEnvironment
		_IConfigurableEnvironment = _AbstractEnvironment
		_IPropertyResolver = _AbstractEnvironment

		// StandardEnvironment 实现 IAbstractEnvironment
		_IAbstractEnvironment = _StandardEnvironment
		// StandardEnvironment 继承 AbstractEnvironment
		_StandardEnvironment.AbstractEnvironment = *_AbstractEnvironment

		// --- 间接 ---
		_IConfigurablePeopertyResolver = _PropertySourcesPropertyResolver
		_IPropertyResolver = _PropertySourcesPropertyResolver

		// Environment 的 IPropertyResolver 通过 PropertySourcesPropertyResolver 间接实现
		_AbstractEnvironment.PropertyResolver = _PropertySourcesPropertyResolver

	}
}

// sources 部分类型测试
var _IPropertySources IPropertySources
var _MutablePropertySources *MutablePropertySources

var _IPropertySource IPropertySource
var _MapPropertySource *MapPropertySource
var _PropertiesPropertySource *PropertiesPropertySource
var _SystemEnvironmentPropertySource *SystemEnvironmentPropertySource

func TestSourcesCheck(t *testing.T) {
	if false {
		// MutablePropertySources 实现 IPropertySources
		_IPropertySources = _MutablePropertySources
		// MutablePropertySources 聚合 IPropertySource
		_MutablePropertySources.propertySourceList.Add(_IPropertySource)

		// IPropertySource 的实现
		_IPropertySource = _MapPropertySource
		_IPropertySource = _PropertiesPropertySource
		_IPropertySource = _SystemEnvironmentPropertySource

		// PropertiesPropertySource 继承 MapPropertySource
		_PropertiesPropertySource.MapPropertySource = *_MapPropertySource
		// SystemEnvironmentPropertySource 继承 MapPropertySource
		_SystemEnvironmentPropertySource.MapPropertySource = *_MapPropertySource
	}
}

func TestEnvironmentSources(t *testing.T) {
	if false {
		_StandardEnvironment.customizePropertySources(_MutablePropertySources)
		// StandardEnvironment 持有 MutablePropertySources
		_StandardEnvironment.PropertySources = _MutablePropertySources
		// StandardEnvironment 的 PropertySources 是 MutablePropertySources 类型
		_MutablePropertySources = _StandardEnvironment.PropertySources
	}
}
