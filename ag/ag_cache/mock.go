package ag_cache

import (
	"context"
	"sync"
	"time"
)

// MockCache 是用于测试的内存缓存。无 TTL、无淘汰。
// 在单元测试中使用，避免依赖真实引擎。
type MockCache[T any] struct {
	mu    sync.RWMutex
	data  map[string]T
	stats Stats
	err   error // 若设置，所有 GetOrElse 返回该错误（模拟缓存故障）
}

// NewMock 创建一个 MockCache 供测试使用。
func NewMock[T any]() *MockCache[T] {
	return &MockCache[T]{data: make(map[string]T)}
}

// SetError 注入一个 GetOrElse 会返回的错误，模拟缓存不可用。
func (m *MockCache[T]) SetError(err error) { m.err = err }

// Get 实现 ICache——纯读。
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

// TryGet 实现 ICache——无未命中错误的宽松读。
func (m *MockCache[T]) TryGet(ctx context.Context, key string) (T, bool, error) {
	m.mu.RLock()
	v, ok := m.data[key]
	m.mu.RUnlock()
	if !ok {
		var zero T
		return zero, false, nil
	}
	return v, true, nil
}

// GetOrElse 实现 ICache。
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

	// 调用 loader。
	v, err := loader(ctx, key)
	if err != nil {
		m.mu.Lock()
		m.stats.Misses++
		m.mu.Unlock()
		return v, err
	}

	m.mu.Lock()
	m.data[key] = v
	m.stats.Misses++ // 计入未命中（与 Ristretto 语义一致）
	m.stats.EntryCount = int64(len(m.data))
	m.mu.Unlock()
	return v, nil
}

// Set 实现 ICache。
func (m *MockCache[T]) Set(ctx context.Context, key string, value T) error {
	m.mu.Lock()
	m.data[key] = value
	m.stats.EntryCount = int64(len(m.data))
	m.mu.Unlock()
	return nil
}

// SetWithTTL 实现 ICache——等同 Set（mock 无 TTL）。
func (m *MockCache[T]) SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) error {
	return m.Set(ctx, key, value)
}

// Del 实现 ICache。
func (m *MockCache[T]) Del(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	for _, key := range keys {
		delete(m.data, key)
	}
	m.stats.EntryCount = int64(len(m.data))
	m.mu.Unlock()
	return nil
}

// Clear 实现 ICache。
func (m *MockCache[T]) Clear(ctx context.Context) error {
	m.mu.Lock()
	m.data = make(map[string]T)
	m.stats.EntryCount = 0
	m.mu.Unlock()
	return nil
}

var _ ICache[string] = (*MockCache[string])(nil)

// ──────── MockEngine ────────

// MockEngine 是 core 层测试的 Engine 测试替身。
type MockEngine struct {
	mu        sync.Mutex
	data      map[string][]byte
	stats     Stats
	PanicNext bool
	Err       error // 后端错误注入
}

// NewMockEngine 创建一个 MockEngine。
func NewMockEngine() *MockEngine {
	return &MockEngine{data: make(map[string][]byte)}
}

// Get 实现 Engine。
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

// Set 实现 Engine。
func (e *MockEngine) Set(ctx context.Context, key string, value []byte) error {
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

// SetWithTTL 实现 TTLSetter——记录 ttl 同时存值。
func (e *MockEngine) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.data[key] = value
	e.stats.EntryCount = int64(len(e.data))
	return nil
}

// Del 实现 Engine。
func (e *MockEngine) Del(ctx context.Context, key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.data, key)
	e.stats.EntryCount = int64(len(e.data))
	return nil
}

// Clear 实现 Engine——清空本独立实例，忽略 prefix。
func (e *MockEngine) Clear(ctx context.Context, prefix string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.data = make(map[string][]byte)
	e.stats.EntryCount = 0
	return nil
}

// Stats 实现 Engine。
func (e *MockEngine) Stats() Stats {
	e.mu.Lock()
	defer e.mu.Unlock()
	s := e.stats
	s.EntryCount = int64(len(e.data))
	return s
}

// Close 实现 Engine。
func (e *MockEngine) Close() error { return nil }

var _ Engine = (*MockEngine)(nil)
