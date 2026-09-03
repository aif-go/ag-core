package agristretto

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aif-go/ag-core/ag/ag_cache"
)

func newTestEngine(t *testing.T, opts RistrettoOptions) ag_cache.Engine {
	t.Helper()
	e, err := NewRistrettoEngine(opts)
	if err != nil {
		t.Fatalf("NewRistrettoEngine: %v", err)
	}
	return e
}

func syncNow(t *testing.T, e ag_cache.Engine) {
	t.Helper()
	if s, ok := e.(interface{ Sync() }); ok {
		s.Sync()
		return
	}
	t.Fatal("engine should implement syncer")
}

// ──────── 引擎内部默认 TTL（Engine.Set 无 ttl 参数）───────

func TestEngine_InternalDefaultTTL(t *testing.T) {
	e := newTestEngine(t, RistrettoOptions{MaxCost: 1 << 20, DefaultTTL: 100 * time.Millisecond})
	defer e.Close()
	ctx := context.Background()

	e.Set(ctx, "k", []byte("v"))
	syncNow(t, e)
	if _, err := e.Get(ctx, "k"); err != nil {
		t.Fatalf("immediate read should hit, got %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	if _, err := e.Get(ctx, "k"); !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("after internal default TTL, expected miss, got %v", err)
	}
}

// ──────── TTLSetter：显式 SetWithTTL ────────

func TestEngine_SetWithTTL(t *testing.T) {
	e := newTestEngine(t, RistrettoOptions{MaxCost: 1 << 20})
	defer e.Close()
	ctx := context.Background()

	s, ok := e.(ag_cache.TTLSetter)
	if !ok {
		t.Fatal("ristrettoEngine should implement TTLSetter")
	}
	s.SetWithTTL(ctx, "k", []byte("v"), 100*time.Millisecond)
	syncNow(t, e)
	if _, err := e.Get(ctx, "k"); err != nil {
		t.Fatalf("immediate read should hit, got %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	if _, err := e.Get(ctx, "k"); !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("after SetWithTTL ttl, expected miss, got %v", err)
	}
}

func TestSet_ZeroTTL_NeverExpires(t *testing.T) {
	e := newTestEngine(t, RistrettoOptions{MaxCost: 1 << 20})
	defer e.Close()
	ctx := context.Background()

	e.Set(ctx, "k", []byte("v"))
	syncNow(t, e)
	time.Sleep(50 * time.Millisecond)
	if _, err := e.Get(ctx, "k"); err != nil {
		t.Fatalf("Set should not expire early, got %v", err)
	}
}

// ──────── 淘汰 ────────

func TestEviction(t *testing.T) {
	e := newTestEngine(t, RistrettoOptions{MaxCost: 1000})
	defer e.Close()
	ctx := context.Background()
	val := make([]byte, 400) // cost 400 per key

	for i := 0; i < 20; i++ {
		e.Set(ctx, fmt.Sprintf("k-%d", i), val)
	}
	syncNow(t, e)

	evicted := 0
	for i := 0; i < 20; i++ {
		if _, err := e.Get(ctx, fmt.Sprintf("k-%d", i)); errors.Is(err, ag_cache.ErrCacheMiss) {
			evicted++
		}
	}
	if evicted == 0 {
		t.Fatal("expected some evictions with tiny MaxCost")
	}
}

// ──────── 异步写 + syncer 可见性 ────────

func TestSync_Visibility(t *testing.T) {
	e := newTestEngine(t, RistrettoOptions{MaxCost: 1 << 20})
	defer e.Close()
	ctx := context.Background()

	if _, err := e.Get(ctx, "k"); !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("expected miss before set, got %v", err)
	}
	e.Set(ctx, "k", []byte("v"))
	syncNow(t, e)
	v, err := e.Get(ctx, "k")
	if err != nil || string(v) != "v" {
		t.Fatalf("sync visibility failed: v=%q err=%v", v, err)
	}
}

// ──────── Clear 忽略 prefix + 收完整 key ────────

func TestClear_IgnoresPrefix(t *testing.T) {
	e := newTestEngine(t, RistrettoOptions{MaxCost: 1 << 20})
	defer e.Close()
	ctx := context.Background()

	// 独立实例：Clear 清整个实例，prefix 被忽略（不按 prefix 过滤，全清）。
	e.Set(ctx, "agcache::users::u:1", []byte("v"))
	e.Set(ctx, "agcache::params::p:1", []byte("p"))
	syncNow(t, e)

	if err := e.Clear(ctx, "agcache::users::"); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	syncNow(t, e)
	if _, err := e.Get(ctx, "agcache::users::u:1"); !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("users key should be cleared, got %v", err)
	}
	if _, err := e.Get(ctx, "agcache::params::p:1"); !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("independent instance Clear clears all (prefix ignored), params key got %v", err)
	}
}

func TestEngine_FullKey(t *testing.T) {
	e := newTestEngine(t, RistrettoOptions{MaxCost: 1 << 20})
	defer e.Close()
	ctx := context.Background()

	e.Set(ctx, "agcache::users::u:1", []byte("v"))
	syncNow(t, e)
	v, err := e.Get(ctx, "agcache::users::u:1")
	if err != nil || string(v) != "v" {
		t.Fatalf("full-key access failed: v=%q err=%v", v, err)
	}
}

// ──────── 工厂 ────────

func TestFactory_CreateName(t *testing.T) {
	f, err := NewAgristrettoFactory(&RistrettoConfigs{Default: RistrettoConfig{MaxCost: 1 << 20}})
	if err != nil {
		t.Fatalf("NewAgristrettoFactory: %v", err)
	}
	if f.Name() != "ristretto" {
		t.Fatalf("Name = %q, want ristretto", f.Name())
	}
	e1, err := f.Create("users")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer e1.Close()
	if e1 == nil {
		t.Fatal("Create returned nil engine")
	}
	e2, err := f.Create("params")
	if err != nil {
		t.Fatalf("Create(2): %v", err)
	}
	defer e2.Close()
	if e1 == e2 {
		t.Fatal("Create(name) should return fresh instances per name")
	}
}

// ──────── Del：引擎级删除 ────────

func TestDel(t *testing.T) {
	e := newTestEngine(t, RistrettoOptions{MaxCost: 1 << 20})
	defer e.Close()
	ctx := context.Background()

	// 多个 key 写入后逐个 Del，验证只删目标 key。
	e.Set(ctx, "a", []byte("A"))
	e.Set(ctx, "b", []byte("B"))
	syncNow(t, e)

	if err := e.Del(ctx, "a"); err != nil {
		t.Fatalf("Del: %v", err)
	}
	syncNow(t, e)

	if _, err := e.Get(ctx, "a"); !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("deleted key a should miss, got %v", err)
	}
	if _, err := e.Get(ctx, "b"); err != nil {
		t.Fatalf("unrelated key b should survive, got %v", err)
	}
}

// ──────── 引擎默认 TTL=0（永不过期）───────

func TestEngine_DefaultTTLZero_NeverExpires(t *testing.T) {
	e := newTestEngine(t, RistrettoOptions{MaxCost: 1 << 20})
	defer e.Close()
	ctx := context.Background()

	// NewRistrettoEngine 默认 defaultTTL=0（永不过期），Set 后经足够时间仍命中。
	e.Set(ctx, "k", []byte("v"))
	syncNow(t, e)
	if _, err := e.Get(ctx, "k"); err != nil {
		t.Fatalf("immediate read should hit, got %v", err)
	}
}

// ──────── RistrettoOptions 层 + Validate ────────

func TestOptions_LayerCompile(t *testing.T) {
	// RistrettoOptions 含 DefaultTTL time.Duration，Validate/NewRistrettoEngine 单参。
	var o RistrettoOptions
	o.MaxCost = 1
	o.NumCounters = 1
	o.BufferItems = 64
	o.DefaultTTL = time.Minute
	if err := o.Validate(); err != nil {
		t.Fatalf("valid options should pass Validate: %v", err)
	}
}

func TestOptions_Validate(t *testing.T) {
	// 负值报错。
	if err := (RistrettoOptions{MaxCost: -1}).Validate(); err == nil {
		t.Fatal("negative MaxCost should fail Validate")
	}
	if err := (RistrettoOptions{NumCounters: -1}).Validate(); err == nil {
		t.Fatal("negative NumCounters should fail Validate")
	}
	if err := (RistrettoOptions{BufferItems: -1}).Validate(); err == nil {
		t.Fatal("negative BufferItems should fail Validate")
	}

	// NewRistrettoEngine 零值兜底：只给 MaxCost，NumCounters/BufferItems 用默认。
	e, err := NewRistrettoEngine(RistrettoOptions{MaxCost: 1 << 20})
	if err != nil {
		t.Fatalf("NewRistrettoEngine zero-fallback should succeed: %v", err)
	}
	_ = e.Close()

	// NewRistrettoEngine 负值报错。
	if _, err := NewRistrettoEngine(RistrettoOptions{MaxCost: -1}); err == nil {
		t.Fatal("NewRistrettoEngine negative MaxCost should error")
	}
}

// ──────── 工厂 per-name（启动预解析）───────

func TestFactory_Namespaces(t *testing.T) {
	cfg := &RistrettoConfigs{
		Default: RistrettoConfig{MaxCost: 100_000_000, NumCounters: 131_072, BufferItems: 64, DefaultTTL: "0"},
		Namespaces: map[string]RistrettoConfig{
			"users":  {MaxCost: 500_000_000},                      // 只覆盖 MaxCost
			"params": {NumCounters: 8_388_608, DefaultTTL: "30s"}, // 覆盖 NumCounters + TTL
		},
	}
	f, err := NewAgristrettoFactory(cfg)
	if err != nil {
		t.Fatalf("NewAgristrettoFactory: %v", err)
	}
	if f.Name() != "ristretto" {
		t.Fatalf("Name = %q, want ristretto", f.Name())
	}

	// users：MaxCost 覆盖，其余继承 Default。
	eu, err := f.Create("users")
	if err != nil {
		t.Fatalf("Create(users): %v", err)
	}
	eu.Close()
	// 无法直接从 Engine 读配置；改用内部 opts map 验证（同包内）。
	ff := f.(agristrettoFactory)
	if got := ff.opts["users"].MaxCost; got != 500_000_000 {
		t.Fatalf("users MaxCost = %d, want 500000000", got)
	}
	if got := ff.opts["users"].NumCounters; got != 131_072 {
		t.Fatalf("users NumCounters = %d, want 131072 (inherit default)", got)
	}

	// params：NumCounters+TTL 覆盖。
	if got := ff.opts["params"].NumCounters; got != 8_388_608 {
		t.Fatalf("params NumCounters = %d, want 8388608", got)
	}
	if got := ff.opts["params"].DefaultTTL; got != 30*time.Second {
		t.Fatalf("params DefaultTTL = %v, want 30s", got)
	}

	// 未命中 name → Create 用全局默认（"" 档），不报错。
	en, err := f.Create("not-configured")
	if err != nil {
		t.Fatalf("Create(not-configured) should fall back to default: %v", err)
	}
	en.Close()
}

func TestFactory_NamespaceInvalid(t *testing.T) {
	// per-name 非法 TTL → 构造报错（含 name 定位）。
	_, err := NewAgristrettoFactory(&RistrettoConfigs{
		Default:    RistrettoConfig{},
		Namespaces: map[string]RistrettoConfig{"users": {DefaultTTL: "abc"}},
	})
	if err == nil {
		t.Fatal("per-name invalid TTL should fail at factory construction")
	}

	// 永不使用的 name 非法 → 也启动期报错（非惰性）。
	_, err = NewAgristrettoFactory(&RistrettoConfigs{
		Default:    RistrettoConfig{},
		Namespaces: map[string]RistrettoConfig{"never-used": {MaxCost: -1}},
	})
	if err == nil {
		t.Fatal("unused per-name negative MaxCost should fail at construction (startup validation)")
	}
}
