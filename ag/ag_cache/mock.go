package ag_cache

import (
	"context"
	"sync"
	"time"
)

// MockCache is an in-memory cache for testing. No TTL, no eviction.
// Use in unit tests to avoid depending on a real engine.
type MockCache[T any] struct {
	mu    sync.RWMutex
	data  map[string]T
	stats Stats
	err   error // if set, all GetOrElse calls return this error (simulates cache failure)
}

// NewMock creates a MockCache for testing.
func NewMock[T any]() *MockCache[T] {
	return &MockCache[T]{data: make(map[string]T)}
}

// SetError injects an error that GetOrElse will return, simulating cache unavailability.
func (m *MockCache[T]) SetError(err error) { m.err = err }

// Get implements ICache — pure read.
func (m *MockCache[T]) Get(ctx context.Context, key string) (T, error) {
	m.mu.RLock()
	v, ok := m.data[key]
	m.mu.RUnlock()
	if ok {
		m.mu.Lock()
		m.stats.Hits++
		m.mu.Unlock()
		return v, nil
	}
	m.mu.Lock()
	m.stats.Misses++
	m.mu.Unlock()
	var zero T
	return zero, ErrCacheMiss
}

// GetOrElse implements ICache.
func (m *MockCache[T]) GetOrElse(ctx context.Context, key string, loader LoaderFunc[T]) (T, error) {
	if m.err != nil {
		var zero T
		return zero, m.err
	}
	m.mu.RLock()
	v, ok := m.data[key]
	m.mu.RUnlock()
	if ok {
		m.mu.Lock()
		m.stats.Hits++
		m.mu.Unlock()
		return v, nil
	}

	// Call loader.
	v, err := loader(ctx, key)
	if err != nil {
		m.mu.Lock()
		m.stats.Misses++
		m.mu.Unlock()
		return v, err
	}

	m.mu.Lock()
	m.data[key] = v
	m.stats.Misses++ // Count as miss even though we loaded (matches Ristretto semantics)
	m.stats.EntryCount = int64(len(m.data))
	m.mu.Unlock()
	return v, nil
}

// Set implements ICache.
func (m *MockCache[T]) Set(ctx context.Context, key string, value T, ttl ...time.Duration) error {
	m.mu.Lock()
	m.data[key] = value
	m.stats.EntryCount = int64(len(m.data))
	m.mu.Unlock()
	return nil
}

// Del implements ICache.
func (m *MockCache[T]) Del(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	for _, key := range keys {
		delete(m.data, key)
	}
	m.stats.EntryCount = int64(len(m.data))
	m.mu.Unlock()
	return nil
}

// Clear implements ICache.
func (m *MockCache[T]) Clear(ctx context.Context) error {
	m.mu.Lock()
	m.data = make(map[string]T)
	m.stats.EntryCount = 0
	m.mu.Unlock()
	return nil
}

// Peek implements AdminCache — pure read, no stats update.
func (m *MockCache[T]) Peek(ctx context.Context, key string) (T, bool, error) {
	m.mu.RLock()
	v, ok := m.data[key]
	m.mu.RUnlock()
	return v, ok, nil
}

// Stats implements AdminCache.
func (m *MockCache[T]) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

var (
	_ ICache[string]     = (*MockCache[string])(nil)
	_ AdminCache[string] = (*MockCache[string])(nil)
)

// ──────── MockEngine ────────

// MockEngine is an Engine test double for core-layer tests.
type MockEngine struct {
	mu        sync.Mutex
	data      map[string][]byte
	stats     Stats
	PanicNext bool
	Err       error // backend error injection
}

// NewMockEngine creates a MockEngine.
func NewMockEngine() *MockEngine {
	return &MockEngine{data: make(map[string][]byte)}
}

// Get implements Engine.
func (e *MockEngine) Get(ctx context.Context, key string) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.PanicNext {
		e.PanicNext = false
		panic("mock engine failure")
	}
	if e.Err != nil {
		return nil, e.Err
	}
	v, ok := e.data[key]
	if ok {
		e.stats.Hits++
		return v, nil
	}
	e.stats.Misses++
	return nil, ErrCacheMiss
}

// Set implements Engine.
func (e *MockEngine) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.PanicNext {
		e.PanicNext = false
		panic("mock engine failure")
	}
	if e.Err != nil {
		return e.Err
	}
	e.data[key] = value
	e.stats.EntryCount = int64(len(e.data))
	return nil
}

// Del implements Engine.
func (e *MockEngine) Del(ctx context.Context, key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.data, key)
	e.stats.EntryCount = int64(len(e.data))
	return nil
}

// Clear implements Engine.
func (e *MockEngine) Clear(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.data = make(map[string][]byte)
	e.stats.EntryCount = 0
	return nil
}

// Stats implements Engine.
func (e *MockEngine) Stats() Stats {
	e.mu.Lock()
	defer e.mu.Unlock()
	s := e.stats
	s.EntryCount = int64(len(e.data))
	return s
}

// Close implements Engine.
func (e *MockEngine) Close() error { return nil }

var _ Engine = (*MockEngine)(nil)
