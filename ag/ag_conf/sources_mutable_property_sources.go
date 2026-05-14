package ag_conf

import (
	"ag-core/ag/ag_ext"
	"fmt"
	"sync"
)

// MutablePropertySources 可变的属性源集合实现
type MutablePropertySources struct {
	lock               sync.RWMutex // 读写锁保护并发访问
	propertySourceList *ag_ext.CopyOnWriteSlice[IPropertySource]
}

func NewMutablePropertySources() *MutablePropertySources {
	return &MutablePropertySources{
		lock:               sync.RWMutex{},
		propertySourceList: ag_ext.NewCopyOnWriteSlice[IPropertySource](),
	}
}

/* ========= 实现IPropertySources接口 ======== */

// Get 获取指定名称的属性源，不存在时返回nil
func (m *MutablePropertySources) Get(name string) IPropertySource {
	m.lock.RLock()
	defer m.lock.RUnlock()

	pslist := m.propertySourceList.Value()
	for _, ps := range pslist {
		// if ps.GetName() == name {
		if ps.EqualsName(name) {
			return ps
		}
	}
	return nil
}

func (m *MutablePropertySources) ContainsSource(ps IPropertySource) bool {
	return m.Contains(ps.GetName())
}

// Contains 判断是否存在指定名称的属性源
func (m *MutablePropertySources) Contains(name string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()

	pslist := m.propertySourceList.Value()
	for _, ps := range pslist {
		// if ps.GetName() == name {
		if ps.EqualsName(name) {
			return true
		}
	}
	return false
}

// GetPropertySources 获取属性源集合
func (m *MutablePropertySources) GetPropertySources() []IPropertySource {
	m.lock.RLock()
	defer m.lock.RUnlock()

	pslist := m.propertySourceList.Value()
	return pslist
}

// RangePropertySourceHandler 遍历处理属性源集合，由resolver遍历调，以从属性源集合中获取属性值
func (m *MutablePropertySources) RangePropertySourceHandler(handler func(ps IPropertySource) (bool, error)) error {
	// m.lock.RLock()
	// defer m.lock.RUnlock()

	// pslist := m.propertySourceList.Value()
	pslist := m.GetPropertySources()
	var handlererr error
	var end bool
	for _, ps := range pslist {
		end, handlererr = handler(ps)
		if end || handlererr != nil {
			// 若遍历结束或处理出错，则退出遍历
			break
		}
	}

	return handlererr
}

// 倒序遍历处理属性源集合，由resolver遍历调，以从属性源集合中获取属性值
func (m *MutablePropertySources) RangePropertySourceHandlerReverse(handler func(ps IPropertySource) (bool, error)) error {
	// m.lock.RLock()
	// defer m.lock.RUnlock()

	// pslist := m.propertySourceList.Value()
	pslist := m.GetPropertySources()
	var handlererr error
	var end bool
	for i := len(pslist) - 1; i >= 0; i-- {
		ps := pslist[i]
		end, handlererr = handler(ps)
		if end || handlererr != nil {
			// 若遍历结束或处理出错，则退出遍历
			break
		}
	}

	return handlererr
}

/* ========= 自实现方法 ======== */

// AddFirst 添加属性源到集合头部
func (m *MutablePropertySources) AddFirst(ps IPropertySource) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.removeIfPresent(ps.GetName())
	m.propertySourceList.AddIndex(0, ps)
}

func (m *MutablePropertySources) AddLast(ps IPropertySource) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.removeIfPresent(ps.GetName())
	m.propertySourceList.Add(ps)
}

func (m *MutablePropertySources) AddBefore(name string, ps IPropertySource) error {
	err := assertLegalRelativeAddition(name, ps)
	if err != nil {
		return err
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	index := m.indexOfName(name)
	if index == -1 {
		return fmt.Errorf("PropertySource named '%s' not found", name)
	}
	m.removeIfPresent(ps.GetName())
	m.propertySourceList.AddIndex(index, ps)
	return nil
}

func (m *MutablePropertySources) AddAfter(name string, ps IPropertySource) error {
	err := assertLegalRelativeAddition(name, ps)
	if err != nil {
		return err
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	index := m.indexOfName(name)
	if index == -1 {
		return fmt.Errorf("PropertySource named '%s' not found", name)
	}
	m.removeIfPresent(ps.GetName())
	m.propertySourceList.AddIndex(index+1, ps)
	return nil
}

func (m *MutablePropertySources) Remove(name string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	index := m.indexOfName(name)
	m.propertySourceList.DeleteIndex(index)
}

func (m *MutablePropertySources) ReplaceSource(ps IPropertySource) error {
	return m.Replace(ps.GetName(), ps)
}
func (m *MutablePropertySources) Replace(name string, ps IPropertySource) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	index := m.indexOfName(name)
	if index == -1 {
		return fmt.Errorf("PropertySource named '%s' not found", name)
	}
	m.propertySourceList.Set(index, ps)
	return nil
}
func (m *MutablePropertySources) RemoveIfPresent(toDelName string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.removeIfPresent(toDelName)
}

// Deprecated: 非并发安全的，不可外部调用
func (m *MutablePropertySources) removeIfPresent(toDelName string) {
	// m.lock.Lock()
	// defer m.lock.Unlock()
	// toDelName := ps.GetName()
	pslist := m.propertySourceList.Value()
	for _, ps := range pslist {
		if ps.GetName() == toDelName {
			m.propertySourceList.Delete(ps)
			return
		}
	}
}

// Deprecated: 非并发安全的，不可外部调用
func (m *MutablePropertySources) indexOfName(name string) int {
	pslist := m.propertySourceList.Value()
	for i, ps := range pslist {
		if ps.GetName() == name {
			return i
		}
	}
	return -1
}

// assertLegalRelativeAddition 断言相对添加是否合法，相对位置添加不可以添加到自身
func assertLegalRelativeAddition(name string, ps IPropertySource) error {
	newName := ps.GetName()
	if name == newName {
		return fmt.Errorf("PropertySource named '%s' cannot be added relative to itself", name)
	}
	return nil
}
