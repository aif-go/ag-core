package ag_cache_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/aif-go/ag-core/ag/ag_cache"
)

// ──────── recordingEngine: 观察 SetWithTTL 收到的 ttl 与 key ────────

type recordingEngine struct {
	*ag_cache.MockEngine
	mu      sync.Mutex
	lastTTL time.Duration
	lastKey string
}

func (e *recordingEngine) Set(ctx context.Context, key string, value []byte) error {
	e.mu.Lock()
	e.lastTTL = 0
	e.lastKey = key
	e.mu.Unlock()
	return e.MockEngine.Set(ctx, key, value)
}

func (e *recordingEngine) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	e.mu.Lock()
	e.lastTTL = ttl
	e.lastKey = key
	e.mu.Unlock()
	return e.MockEngine.SetWithTTL(ctx, key, value, ttl)
}

func (e *recordingEngine) LastTTL() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.lastTTL
}

func (e *recordingEngine) LastKey() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.lastKey
}

// ──────── key 前缀方案 A ────────

func TestPrefix_KeyJoin(t *testing.T) {
	tf := &plainTrackingFactory{name: "tracking"}
	props := ag_cache.DefaultAgCacheProperties()
	props.DefaultEngine = "tracking"
	m, err := ag_cache.NewManager(props)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	m.SetEngineFactory("tracking", tf)
	ag_cache.SetDefault(m)
	defer ag_cache.CloseAll()
	ctx := context.Background()

	// 触发 "users" 实例创建（engine 来自 tracking factory）
	_ = ag_cache.GetCacheWithLoader[string](dflt(), "users", strLoader("v"))
	if tf.last == nil {
		t.Fatal("tracking engine not created")
	}
	ag_cache.GetCacheWithLoader[string](dflt(), "users", strLoader("v")).Set(ctx, "u:1", "v")
	if got := tf.last.LastKey(); got != "agcache::users::u:1" {
		t.Fatalf("engine Set key = %q, want agcache::users::u:1", got)
	}
}

// plainTrackingFactory: 引擎自身默认 TTL（经内部字段，非 DefaultTTLProvider）
type plainTrackingFactory struct {
	name string
	last *recordingEngine
}

func (f *plainTrackingFactory) Name() string { return f.name }
func (f *plainTrackingFactory) Create(name string) (ag_cache.Engine, error) {
	e := &recordingEngine{MockEngine: ag_cache.NewMockEngine()}
	f.last = e
	return e, nil
}

// ──────── Manager 实例管理 ────────

func TestNew_LazyCreateAndReuse(t *testing.T) {
	setupManager(t)
	ctx := context.Background()

	c1 := ag_cache.GetCacheWithLoader[string](dflt(), "users", func(ctx context.Context, key string) (string, error) { return "v", nil })
	c1.GetOrElse(ctx, "k", func(ctx context.Context, key string) (string, error) { return "v", nil })

	// Same name reuses the same instance.
	loaded := false
	v, err := ag_cache.GetCacheWithLoader[string](dflt(), "users", func(ctx context.Context, key string) (string, error) {
		loaded = true
		return "fresh", nil
	}).GetOrElse(ctx, "k", func(ctx context.Context, key string) (string, error) {
		loaded = true
		return "fresh", nil
	})
	if err != nil || v != "v" {
		t.Fatalf("reuse failed: v=%q err=%v", v, err)
	}
	if loaded {
		t.Fatal("loader should not be called on cache hit (instance not reused)")
	}
}

func TestManager_SameNameDiffType_Panics(t *testing.T) {
	setupManager(t)
	ctx := context.Background()

	ag_cache.GetCacheWithLoader[string](dflt(), "users", strLoader("v")).Get(ctx, "k")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on same name with different type")
		}
	}()
	ag_cache.GetCacheWithLoader[int](dflt(), "users", func(ctx context.Context, key string) (int, error) { return 1, nil }).Get(ctx, "k")
}

// ──────── TTL 优先级链 ────────

func TestTTL_PriorityChain(t *testing.T) {
	tf := &plainTrackingFactory{name: "ttl-track"}
	props := ag_cache.DefaultAgCacheProperties()
	props.DefaultEngine = "ttl-track"
	m, err := ag_cache.NewManager(props)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	m.SetEngineFactory("ttl-track", tf)
	ag_cache.SetDefault(m)
	defer ag_cache.CloseAll()
	ctx := context.Background()

	// 1) 未配 WithDefaultTTL → engine.Set（引擎内部默认，lastTTL=0 表示走 Set）
	ag_cache.GetCacheWithLoader[string](dflt(), "a", strLoader("v")).Set(ctx, "k", "v")
	if got := tf.last.LastTTL(); got != 0 {
		t.Fatalf("plain Set should use engine.Set (lastTTL=0), got %v", got)
	}

	// 2) 配了 WithDefaultTTL → Set 经 TTLSetter 用默认
	ag_cache.GetCacheWithLoader[string](dflt(), "b", strLoader("v"), ag_cache.WithDefaultTTL[string](30*time.Second)).Set(ctx, "k", "v")
	if got := tf.last.LastTTL(); got != 30*time.Second {
		t.Fatalf("WithDefaultTTL via TTLSetter = %v, want 30s", got)
	}

	// 3) SetWithTTL 单条显式覆盖（最高优先级）
	ag_cache.GetCacheWithLoader[string](dflt(), "b", strLoader("v"), ag_cache.WithDefaultTTL[string](30*time.Second)).SetWithTTL(ctx, "k", "v", 60*time.Second)
	if got := tf.last.LastTTL(); got != 60*time.Second {
		t.Fatalf("SetWithTTL explicit = %v, want 60s", got)
	}
}

// ──────── 入口编译断言 ────────

func TestEntry_InterfaceCompile(t *testing.T) {
	setupManager(t)
	m := dflt()
	ctx := context.Background()

	// GetCacheWithLoader / GetCache / DefaultManager 存在。
	l := ag_cache.GetCacheWithLoader[string](m, "users", strLoader("v"))
	c := ag_cache.GetCache[string](m, "params")
	_ = l
	_ = c
	_ = ag_cache.DefaultManager()
	_ = m.Close

	// v3 包级 New / Get（走 default）已移除。
	if _, err := c.Get(ctx, "k"); !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("pure read should miss: %v", err)
	}
}
