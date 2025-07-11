package ag_conf

// IPropertySources 属性源集合接口
type IPropertySources interface {
	// Get 获取指定名称的属性源，不存在时返回nil
	Get(name string) IPropertySource
	// Contains 判断是否存在指定名称的属性源
	Contains(name string) bool
	// GetPropertySources 获取属性源集合
	GetPropertySources() []IPropertySource
	// RangePropertySourceHandler 遍历处理属性源集合 end:是否结束遍历, err:错误
	RangePropertySourceHandler(func(ps IPropertySource) (end bool, err error)) error
	RangePropertySourceHandlerReverse(func(ps IPropertySource) (end bool, err error)) error
}
