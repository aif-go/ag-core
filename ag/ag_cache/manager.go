package ag_cache

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// Manager 懒创建并按名复用缓存实例。
// 它收集引擎工厂（fx group 注入）并经 config 选择默认引擎；
// 实例的引擎选择由配置驱动。
type Manager struct {
	mu              sync.Mutex
	defaultEngine   string
	engineFactories map[string]EngineFactory // name → factory
	caches          map[string]any           // name → *typedCache[T]
}

// NewManager 从 core 属性创建 Manager。
func NewManager(props *AgCacheProperties) (*Manager, error) {
	if props == nil {
		return nil, errors.New("agcache: nil AgCacheProperties")
	}
	engine := props.DefaultEngine
	if engine == "" {
		engine = "ristretto"
	}
	return &Manager{
		defaultEngine:   engine,
		engineFactories: make(map[string]EngineFactory),
		caches:          make(map[string]any),
	}, nil
}

// SetEngineFactory 按 Name 注册引擎工厂（fx group 消费）。
func (m *Manager) SetEngineFactory(name string, f EngineFactory) {
	if m == nil {
		panic("agcache: nil Manager")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.engineFactories[name] = f
}

// EngineFactory 返回 name 下注册的引擎工厂（不存在返回 nil）。
func (m *Manager) EngineFactory(name string) EngineFactory {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.engineFactories[name]
}

// DefaultEngine 返回配置的默认引擎名。
func (m *Manager) DefaultEngine() string {
	if m == nil {
		return ""
	}
	return m.defaultEngine
}

// Close 关闭所有懒创建的缓存实例。幂等。
// 注意：Close 后 Manager 不应复用；请新建一个。
func (m *Manager) Close() error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.caches {
		c.(engineCloser).closeEngine()
	}
	m.caches = make(map[string]any)
	return nil
}

// engineCloser 使 Close 能关闭任意 T 的 typedCache 实例。
type engineCloser interface{ closeEngine() }

// ──────── 可替换默认实例（sql-package 风格）───────

var defaultManager atomic.Pointer[Manager]

// SetDefault 替换 DefaultManager() 使用的默认 Manager。
// 通常在 Fx Invoke（或测试 setup）中调用。线程安全。
func SetDefault(m *Manager) {
	defaultManager.Store(m)
}

// GetCacheWithLoader 返回绑定 loader 的 LoaderCache，底层是显式 Manager 的
// 具名缓存实例。opts 可覆盖默认 TTL（WithDefaultTTL）或序列化器（WithSerializer）。
// 缓存实例首次使用时懒创建，并按名复用。
func GetCacheWithLoader[T any](m *Manager, name string, loader LoaderFunc[T], opts ...Option[T]) *LoaderCache[T] {
	if m == nil {
		panic("agcache: nil Manager")
	}
	return &LoaderCache[T]{inner: getOrCreate[T](m, name, opts...), loader: loader}
}

// GetCache 从显式 Manager 返回具名缓存（纯读，无 loader）。
func GetCache[T any](m *Manager, name string) ICache[T] {
	if m == nil {
		panic("agcache: nil Manager")
	}
	return getOrCreate[T](m, name)
}

// DefaultManager 返回 SetDefault 设置的当前默认 Manager。
// 未调用 SetDefault 时返回 nil。
func DefaultManager() *Manager {
	return defaultManager.Load()
}

// CloseAll 关闭默认 Manager 并清空。幂等。
func CloseAll() {
	m := defaultManager.Load()
	if m == nil {
		return
	}
	_ = m.Close()
	defaultManager.Store(nil)
}

// getOrCreate 懒创建指定 name 的 typed cache。
// 引擎选择：config 默认引擎工厂 Create(name)。
// TTL 优先级：WithDefaultTTL > 引擎内部默认。
// 引擎创建发生在锁外；重取锁后双重检查。
func getOrCreate[T any](m *Manager, name string, opts ...Option[T]) *typedCache[T] {
	m.mu.Lock()
	if c, ok := m.caches[name]; ok {
		m.mu.Unlock()
		return c.(*typedCache[T])
	}
	m.mu.Unlock()

	// 先应用 opts 以确定任何显式 TTL。
	c := &typedCache[T]{
		name:       name,
		prefix:     cachePrefix(name),
		serializer: DefaultSerializer[T](),
		defaultTTL: 0,
	}
	for _, o := range opts {
		o(c)
	}

	f := m.EngineFactory(m.defaultEngine)
	if f == nil {
		panic(fmt.Sprintf("agcache: engine %q not registered", m.defaultEngine))
	}

	engine, err := f.Create(name)
	if err != nil {
		panic(fmt.Sprintf("agcache: create engine for %q: %v", name, err))
	}
	c.engine = engine

	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.caches[name]; ok { // 另一 goroutine 赢得了竞争
		_ = engine.Close()
		return existing.(*typedCache[T])
	}
	m.caches[name] = c
	return c
}
