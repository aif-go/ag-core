package agristretto

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aif-go/ag-core/ag/ag_cache"
)

func newTestEngine(t *testing.T, cfg RistrettoConfig) ag_cache.Engine {
	t.Helper()
	e, err := NewRistrettoEngine(cfg)
	if err != nil {
		t.Fatalf("NewRistrettoEngine: %v", err)
	}
	return e
}

func newTestEngineTTL(t *testing.T, cfg RistrettoConfig, ttl time.Duration) ag_cache.Engine {
	t.Helper()
	e, err := newRistrettoEngine(cfg, ttl)
	if err != nil {
		t.Fatalf("newRistrettoEngine: %v", err)
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
	e := newTestEngineTTL(t, RistrettoConfig{MaxCost: 1 << 20}, 100*time.Millisecond)
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
	e := newTestEngine(t, RistrettoConfig{MaxCost: 1 << 20})
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
	e := newTestEngine(t, RistrettoConfig{MaxCost: 1 << 20})
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
	e := newTestEngine(t, RistrettoConfig{MaxCost: 1000})
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
	e := newTestEngine(t, RistrettoConfig{MaxCost: 1 << 20})
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
	e := newTestEngine(t, RistrettoConfig{MaxCost: 1 << 20})
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
	e := newTestEngine(t, RistrettoConfig{MaxCost: 1 << 20})
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
	f := agristrettoFactory{
		cfg: RistrettoConfig{MaxCost: 1 << 20},
		ttl: 30 * time.Second,
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
	e := newTestEngine(t, RistrettoConfig{MaxCost: 1 << 20})
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
	e := newTestEngine(t, RistrettoConfig{MaxCost: 1 << 20})
	defer e.Close()
	ctx := context.Background()

	// NewRistrettoEngine 默认 defaultTTL=0（永不过期），Set 后经足够时间仍命中。
	e.Set(ctx, "k", []byte("v"))
	syncNow(t, e)
	if _, err := e.Get(ctx, "k"); err != nil {
		t.Fatalf("immediate read should hit, got %v", err)
	}
}
