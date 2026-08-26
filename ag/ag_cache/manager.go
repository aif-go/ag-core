package ag_cache

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Manager lazily creates and reuses per-name cache instances.
// It holds no assembly config besides the default engine name; engine selection
// per instance is via WithEngine option (or the default engine).
type Manager struct {
	mu            sync.Mutex
	defaultEngine string
	caches        map[string]any // name → *typedCache[T]
}

// NewManager creates a Manager from core properties.
// Fails fast if the default engine is not registered.
func NewManager(props *AgCacheProperties) (*Manager, error) {
	if props == nil {
		return nil, errors.New("agcache: nil AgCacheProperties")
	}
	engine := props.DefaultEngine
	if engine == "" {
		engine = "ristretto"
	}
	if _, err := getFactory(engine); err != nil {
		return nil, fmt.Errorf("agcache: default engine: %w", err)
	}
	return &Manager{
		defaultEngine: engine,
		caches:        make(map[string]any),
	}, nil
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

// New returns a LoaderCache bound to a loader, backed by the named cache
// instance from the default manager. opts may select a specific engine
// (WithEngine) or override the default TTL (WithDefaultTTL).
// Panics if SetDefault has not been called.
func New[T any](name string, loader LoaderFunc[T], opts ...Option[T]) *LoaderCache[T] {
	m := defaultManager.Load()
	if m == nil {
		panic("agcache: no default manager — call SetDefault or use NewWithEngine")
	}
	return &LoaderCache[T]{inner: getOrCreate[T](m, name, opts...), loader: loader}
}

// Get returns the named cache from the default manager.
// Panics if SetDefault has not been called.
func Get[T any](name string) ICache[T] {
	m := defaultManager.Load()
	if m == nil {
		panic("agcache: no default manager — call SetDefault or use NewWithEngine")
	}
	return getOrCreate[T](m, name)
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
// Engine selection: WithEngine(opts) or the manager default engine.
// TTL priority: WithDefaultTTL > engine DefaultTTLProvider > 5min.
// Engine creation happens OUTSIDE the lock; double-check after re-acquiring.
func getOrCreate[T any](m *Manager, name string, opts ...Option[T]) *typedCache[T] {
	m.mu.Lock()
	if c, ok := m.caches[name]; ok {
		m.mu.Unlock()
		return c.(*typedCache[T])
	}
	m.mu.Unlock()

	// Apply options first to determine engine name and any explicit TTL.
	c := &typedCache[T]{
		serializer: DefaultSerializer[T](),
		defaultTTL: 0,
	}
	for _, o := range opts {
		o(c)
	}

	engineName := c.engineName
	if engineName == "" {
		engineName = m.defaultEngine
	}
	f, err := getFactory(engineName)
	if err != nil {
		panic(err.Error())
	}

	engine, err := f.Create()
	if err != nil {
		panic(fmt.Sprintf("agcache: create engine for %q: %v", name, err))
	}
	c.engine = engine

	if !c.ttlSet {
		c.defaultTTL = 5 * time.Minute
		if p, ok := f.(DefaultTTLProvider); ok {
			c.defaultTTL = p.DefaultTTL()
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.caches[name]; ok { // another goroutine won the race
		_ = engine.Close()
		return existing.(*typedCache[T])
	}
	m.caches[name] = c
	return c
}
