package ag_conf

import (
	"sync"
)

/* *** NamedPropertySource *** */
// NamedPropertySource 命名属性源，主要用于支持属性源根据名字比对的功能，便于PropertySources对source的维护
type NamedPropertySource struct {
	Name string
}

func (n *NamedPropertySource) GetName() string {
	return n.Name
}

func (n *NamedPropertySource) EqualsName(psname string) bool {
	return n.GetName() == psname
}

/* *** MapPropertySource *** */
// MapPropertySource 基于map的属性源，实现IPropertySource接口
type MapPropertySource struct {
	NamedPropertySource
	Source map[string]any
}

// GetSource 获取属性源的源数据， implement IPropertySource
func (p *MapPropertySource) GetSource() map[string]any {
	return p.Source
}

// GetProperty 获取属性源中的属性值， implement IPropertySource
func (p *MapPropertySource) GetProperty(key string) any {
	return p.Source[key]
}

// ContainsProperty 判断属性源中是否包含指定键 key 的属性， implement IPropertySource
func (p *MapPropertySource) ContainsProperty(key string) bool {
	_, ok := p.Source[key]
	return ok
}

// GetPropertyNames 获取属性源中的所有属性键， customize MapPropertySource
func (p *MapPropertySource) GetPropertyNames() []string {
	keys := []string{}
	for key := range p.Source {
		keys = append(keys, key)
	}
	return keys
}

/* *** PropertiesPropertySource *** */
// PropertiesPropertySource 基于Properties的属性源，实现IPropertySource接口
type PropertiesPropertySource struct {
	MapPropertySource
	lock sync.Mutex
}

// GetPropertyNames 调用MapPropertySource的GetPropertyNames方法，添加并发控制
func (p *PropertiesPropertySource) GetPropertyNames() []string {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.MapPropertySource.GetPropertyNames()
}

func NewPropertiesPropertySource(name string, source map[string]any) *PropertiesPropertySource {
	return &PropertiesPropertySource{
		MapPropertySource: MapPropertySource{
			NamedPropertySource: NamedPropertySource{
				Name: name,
			},
			Source: source,
		},
		lock: sync.Mutex{},
	}
}

/* *** SystemEnvironmentPropertySource *** */
// SystemEnvironmentPropertySource 基于系统环境变量的属性源，实现IPropertySource接口
type SystemEnvironmentPropertySource struct {
	MapPropertySource
}

func NewSystemEnvironmentPropertySource(name string, source map[string]any) *SystemEnvironmentPropertySource {
	return &SystemEnvironmentPropertySource{
		MapPropertySource: MapPropertySource{
			NamedPropertySource: NamedPropertySource{
				Name: name,
			},
			Source: source,
		},
	}
}
