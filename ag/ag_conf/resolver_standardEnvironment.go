package ag_conf

const ()

// StandardEnvironment is a standard implementation of Environment.
type StandardEnvironment struct {
	AbstractEnvironment
}

// NewStandardEnvironment creates a new StandardEnvironment instance.
func NewStandardEnvironment() (*StandardEnvironment, error) {
	e := &StandardEnvironment{}
	// 初始化MutablePropertySources
	e.PropertySources = NewMutablePropertySources()

	// 初始化PropertyResolver,传入PropertySources
	e.PropertyResolver = NewPropertySourcesPropertyResolver(e.PropertySources)

	// customizePropertySources 环境变量和-D属性添加到配置源中
	err := e.customizePropertySources(e.PropertySources)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// customizePropertySources 环境变量和-D属性添加到配置源中
func (e *StandardEnvironment) customizePropertySources(ps *MutablePropertySources) error {
	ps.AddLast(NewPropertiesPropertySource(SourceKeySystemProperties, e.GetSystemProperties()))
	ps.AddLast(NewSystemEnvironmentPropertySource(SourceKeySystemEnvironment, e.GetSystemEnvironment()))
	// 做本地配置文件中相关字段的解密处理 TODO 应该在加载local配置后做一遍boot阶段的解密处理
	return DecryptSystemConfig(ps)
}
