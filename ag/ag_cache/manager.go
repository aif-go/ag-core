package ag_cache

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// Manager lazily creates and reuses per-name cache instances.
// It collects engine factories (fx group injection) and selects the default
// engine via config; engine selection per instance is config-driven.
type Manager struct {
	mu              sync.Mutex
	defaultEngine   string
	engineFactories map[string]EngineFactory // name → factory
	caches          map[string]any           // name → *typedCache[T]
}

// NewManager creates a Manager from core properties.
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

// SetEngineFactory registers an engine factory by Name (fx group consumption).
func (m *Manager) SetEngineFactory(name string, f EngineFactory) {
	if m == nil {
		panic("agcache: nil Manager")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.engineFactories[name] = f
}

// EngineFactory returns the engine factory registered under name (nil if absent).
func (m *Manager) EngineFactory(name string) EngineFactory {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.engineFactories[name]
}

// DefaultEngine returns the configured default engine name.
func (m *Manager) DefaultEngine() string {
	if m == nil {
		return ""
	}
	return m.defaultEngine
}

// Close closes all lazily-created cache instances. Idempotent.
// NOTE: after Close the Manager should not be reused; create a new one.
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

// engineCloser lets Close close typedCache instances of any T.
type engineCloser interface{ closeEngine() }

// ──────── Replaceable default instance (sql-package style) ────────

var defaultManager atomic.Pointer[Manager]

// SetDefault replaces the default manager used by package-level New/Get.
// Typically called from an Fx Invoke (or test setup). Thread-safe.
func SetDefault(m *Manager) {
	defaultManager.Store(m)
}

// GetCacheWithLoader returns a LoaderCache bound to a loader, backed by the
// named cache instance from the explicit Manager. opts may override the
// default TTL (WithDefaultTTL) or serializer (WithSerializer).
// The cache instance is lazily created on first use and reused by name.
func GetCacheWithLoader[T any](m *Manager, name string, loader LoaderFunc[T], opts ...Option[T]) *LoaderCache[T] {
	if m == nil {
		panic("agcache: nil Manager")
	}
	return &LoaderCache[T]{inner: getOrCreate[T](m, name, opts...), loader: loader}
}

// GetCache returns the named cache from the explicit Manager (pure read, no loader).
func GetCache[T any](m *Manager, name string) ICache[T] {
	if m == nil {
		panic("agcache: nil Manager")
	}
	return getOrCreate[T](m, name)
}

// DefaultManager returns the current default manager set by SetDefault.
// Returns nil if SetDefault has not been called.
func DefaultManager() *Manager {
	return defaultManager.Load()
}

// CloseAll closes the default manager and clears it. Idempotent.
func CloseAll() {
	m := defaultManager.Load()
	if m == nil {
		return
	}
	_ = m.Close()
	defaultManager.Store(nil)
}

// getOrCreate lazily creates the typed cache for a name.
// Engine selection: config default engine factory Create(name).
// TTL priority: WithDefaultTTL > engine internal default.
// Engine creation happens OUTSIDE the lock; double-check after re-acquiring.
func getOrCreate[T any](m *Manager, name string, opts ...Option[T]) *typedCache[T] {
	m.mu.Lock()
	if c, ok := m.caches[name]; ok {
		m.mu.Unlock()
		return c.(*typedCache[T])
	}
	m.mu.Unlock()

	// Apply options first to determine any explicit TTL.
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
	if existing, ok := m.caches[name]; ok { // another goroutine won the race
		_ = engine.Close()
		return existing.(*typedCache[T])
	}
	m.caches[name] = c
	return c
}
