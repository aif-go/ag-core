package ag_cache_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/aif-go/ag-core/ag/ag_cache"
)

// ──────── recordingEngine: 观察 Set 收到的 ttl ────────

type recordingEngine struct {
	*ag_cache.MockEngine
	mu      sync.Mutex
	lastTTL time.Duration
}

func (e *recordingEngine) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	e.mu.Lock()
	e.lastTTL = ttl
	e.mu.Unlock()
	return e.MockEngine.Set(ctx, key, value, ttl)
}

func (e *recordingEngine) LastTTL() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.lastTTL
}

// plainTrackingFactory: 无 DefaultTTLProvider（core 兜底 5min）
type plainTrackingFactory struct {
	name string
	last *recordingEngine
}

func (f *plainTrackingFactory) Name() string { return f.name }
func (f *plainTrackingFactory) Create() (ag_cache.Engine, error) {
	e := &recordingEngine{MockEngine: ag_cache.NewMockEngine()}
	f.last = e
	return e, nil
}

// ttlTrackingFactory: 实现 DefaultTTLProvider（引擎自声明默认 TTL）
type ttlTrackingFactory struct {
	plainTrackingFactory
	ttl time.Duration
}

func (f *ttlTrackingFactory) DefaultTTL() time.Duration { return f.ttl }

// ──────── Manager 实例管理 ────────

func TestNew_LazyCreateAndReuse(t *testing.T) {
	registerMockEngine()
	props := ag_cache.DefaultAgCacheProperties()
	props.DefaultEngine = "mock"
	m, err := ag_cache.NewManager(props)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	ag_cache.SetDefault(m)
	defer ag_cache.CloseAll()
	ctx := context.Background()

	c1 := ag_cache.New[string]("users", func(ctx context.Context, key string) (string, error) { return "v", nil })
	c1.GetOrElse(ctx, "k", func(ctx context.Context, key string) (string, error) { return "v", nil })

	// Same name reuses the same instance.
	loaded := false
	v, err := ag_cache.New[string]("users", func(ctx context.Context, key string) (string, error) {
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

func TestGetAdmin(t *testing.T) {
	registerMockEngine()
	props := ag_cache.DefaultAgCacheProperties()
	props.DefaultEngine = "mock"
	m, err := ag_cache.NewManager(props)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	ag_cache.SetDefault(m)
	defer ag_cache.CloseAll()
	ctx := context.Background()

	admin := ag_cache.GetAdmin[string]("users")
	admin.GetOrElse(ctx, "k", func(ctx context.Context, key string) (string, error) { return "v", nil })
	if s := admin.Stats(); s.EntryCount < 1 {
		t.Fatalf("expected entries>=1, got %d", s.EntryCount)
	}
}

func TestManager_UnknownEngine_FailFast(t *testing.T) {
	props := ag_cache.DefaultAgCacheProperties()
	props.DefaultEngine = "no-such-engine"
	if _, err := ag_cache.NewManager(props); err == nil {
		t.Fatal("NewManager should fail fast for unknown default engine")
	}
}

func TestManager_Visit(t *testing.T) {
	registerMockEngine()
	props := ag_cache.DefaultAgCacheProperties()
	props.DefaultEngine = "mock"
	m, err := ag_cache.NewManager(props)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	ag_cache.SetDefault(m)
	defer ag_cache.CloseAll()
	ctx := context.Background()

	ag_cache.New[string]("users", strLoader("U1")).GetOrElse(ctx, "u1", strLoader("U1"))
	ag_cache.Get[string]("params").Get(ctx, "missing")

	visited := map[string]ag_cache.Stats{}
	m.Visit(func(name string, s ag_cache.Stats) {
		visited[name] = s
	})
	if len(visited) != 2 {
		t.Fatalf("Visit should report 2 namespaces, got %d: %v", len(visited), visited)
	}
	if _, ok := visited["users"]; !ok {
		t.Fatalf("users not visited: %v", visited)
	}
	if _, ok := visited["params"]; !ok {
		t.Fatalf("params not visited: %v", visited)
	}
}

func TestManager_SameNameDiffType_Panics(t *testing.T) {
	registerMockEngine()
	props := ag_cache.DefaultAgCacheProperties()
	props.DefaultEngine = "mock"
	m, err := ag_cache.NewManager(props)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	ag_cache.SetDefault(m)
	defer ag_cache.CloseAll()
	ctx := context.Background()

	ag_cache.New[string]("users", strLoader("v")).Get(ctx, "k")
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on same name with different type")
		}
	}()
	ag_cache.New[int]("users", func(ctx context.Context, key string) (int, error) { return 1, nil }).Get(ctx, "k")
}

// ──────── 默认 TTL 三级优先级 ────────

func TestDefaultTTL_Priority(t *testing.T) {
	registerMockEngine()
	plain := &plainTrackingFactory{name: "plain"} // 无 DefaultTTLProvider
	ttl30 := &ttlTrackingFactory{ttl: 30 * time.Second}
	ttl30.plainTrackingFactory.name = "ttl30" // 嵌入工厂的名字
	ctx := context.Background()

	ag_cache.RegisterEngine(plain)
	ag_cache.RegisterEngine(ttl30)
	defer ag_cache.CloseAll()

	// 1) 无 DefaultTTLProvider → core 兜底 5min
	m1, err := ag_cache.NewManager(&ag_cache.AgCacheProperties{DefaultEngine: "plain"})
	if err != nil {
		t.Fatalf("m1: %v", err)
	}
	ag_cache.SetDefault(m1)
	ag_cache.New[string]("a", strLoader("v")).Set(ctx, "k", "v")
	if got := plain.last.LastTTL(); got != 5*time.Minute {
		t.Fatalf("plain engine default TTL = %v, want 5min", got)
	}

	// 2) 引擎声明 DefaultTTLProvider = 30s
	m2, err := ag_cache.NewManager(&ag_cache.AgCacheProperties{DefaultEngine: "ttl30"})
	if err != nil {
		t.Fatalf("m2: %v", err)
	}
	ag_cache.SetDefault(m2)
	ag_cache.New[string]("b", strLoader("v")).Set(ctx, "k", "v")
	if got := ttl30.last.LastTTL(); got != 30*time.Second {
		t.Fatalf("ttl30 engine default TTL = %v, want 30s", got)
	}

	// 3) WithDefaultTTL 覆盖引擎默认
	ag_cache.New[string]("c", strLoader("v"), ag_cache.WithDefaultTTL[string](60*time.Second)).Set(ctx, "k", "v")
	if got := ttl30.last.LastTTL(); got != 60*time.Second {
		t.Fatalf("WithDefaultTTL override = %v, want 60s", got)
	}
}
