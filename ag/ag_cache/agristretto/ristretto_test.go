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

func syncNow(t *testing.T, e ag_cache.Engine) {
	t.Helper()
	if s, ok := e.(interface{ Sync() }); ok {
		s.Sync()
		return
	}
	t.Fatal("engine should implement syncer")
}

// ──────── TTL 语义 ────────

func TestTTL_Expiry(t *testing.T) {
	e := newTestEngine(t, RistrettoConfig{MaxCost: 1 << 20})
	defer e.Close()
	ctx := context.Background()

	e.Set(ctx, "k", []byte("v"), 100*time.Millisecond)
	syncNow(t, e)
	if _, err := e.Get(ctx, "k"); err != nil {
		t.Fatalf("immediate read should hit, got %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	if _, err := e.Get(ctx, "k"); !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("after TTL, expected miss, got %v", err)
	}
}

func TestSet_ZeroTTL_NeverExpires(t *testing.T) {
	e := newTestEngine(t, RistrettoConfig{MaxCost: 1 << 20})
	defer e.Close()
	ctx := context.Background()

	e.Set(ctx, "k", []byte("v"), 0)
	syncNow(t, e)
	time.Sleep(50 * time.Millisecond)
	if _, err := e.Get(ctx, "k"); err != nil {
		t.Fatalf("ttl=0 should never expire, got %v", err)
	}
}

// ──────── 淘汰 ────────

func TestEviction(t *testing.T) {
	e := newTestEngine(t, RistrettoConfig{MaxCost: 1000})
	defer e.Close()
	ctx := context.Background()
	val := make([]byte, 400) // cost 400 per key

	for i := 0; i < 20; i++ {
		e.Set(ctx, fmt.Sprintf("k-%d", i), val, 0)
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
	e.Set(ctx, "k", []byte("v"), 0)
	syncNow(t, e)
	v, err := e.Get(ctx, "k")
	if err != nil || string(v) != "v" {
		t.Fatalf("sync visibility failed: v=%q err=%v", v, err)
	}
}

// ──────── 工厂 ────────

func TestFactory_Create(t *testing.T) {
	f := agristrettoFactory{
		cfg: RistrettoConfig{MaxCost: 1 << 20},
		ttl: 30 * time.Second,
	}
	if f.Name() != "ristretto" {
		t.Fatalf("Name = %q, want ristretto", f.Name())
	}
	e, err := f.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer e.Close()
	if e == nil {
		t.Fatal("Create returned nil engine")
	}
}

func TestFactory_DefaultTTL(t *testing.T) {
	f := agristrettoFactory{ttl: 30 * time.Second}
	if p, ok := any(f).(ag_cache.DefaultTTLProvider); ok {
		if got := p.DefaultTTL(); got != 30*time.Second {
			t.Fatalf("DefaultTTL = %v, want 30s", got)
		}
	} else {
		t.Fatal("agristrettoFactory should implement DefaultTTLProvider")
	}
}
