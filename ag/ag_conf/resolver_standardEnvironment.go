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
	eps := NewSystemEnvironmentPropertySource(SourceKeySystemEnvironment, e.GetSystemEnvironment())
	ps.AddFirst(eps)

	pps := NewPropertiesPropertySource(SourceKeySystemProperties, e.GetSystemProperties()) // -D参数优先级更高
	err := ps.AddBefore(eps.GetName(), pps)
	if err != nil {
		return err
	}

	// =解密=
	err = CreateOrUpdateDecryptForPropertySource(e, eps)
	if err != nil {
		return err
	}
	err = CreateOrUpdateDecryptForPropertySource(e, pps)
	if err != nil {
		return err
	}
	// // 做本地配置文件中相关字段的解密处理 TODO 应该在加载local配置后做一遍boot阶段的解密处理
	// err = DecryptSystemConfig(ps)
	// if err != nil {
	// 	return err
	// }
	return nil
}
