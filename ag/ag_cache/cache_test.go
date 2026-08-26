package ag_cache_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aif-go/ag-core/ag/ag_cache"
)

// ──────── helpers ────────

// mockEngineFactory registers a "mock" engine for core-layer tests (no Ristretto).
type mockEngineFactory struct{}

func (mockEngineFactory) Name() string { return "mock" }
func (mockEngineFactory) Create() (ag_cache.Engine, error) {
	return ag_cache.NewMockEngine(), nil
}

// countingFactory registers a "counting" engine and counts Create calls.
type countingFactory struct {
	creates atomic.Int32
}

func (f *countingFactory) Name() string { return "counting" }
func (f *countingFactory) Create() (ag_cache.Engine, error) {
	f.creates.Add(1)
	return ag_cache.NewMockEngine(), nil
}

var registerMockOnce sync.Once

func registerMockEngine() {
	registerMockOnce.Do(func() {
		ag_cache.RegisterEngine(mockEngineFactory{})
	})
}

// setupManager sets the default manager backed by mock engines.
func setupManager(t *testing.T) {
	t.Helper()
	registerMockEngine()
	props := ag_cache.DefaultAgCacheProperties()
	props.DefaultEngine = "mock"
	m, err := ag_cache.NewManager(props)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	ag_cache.SetDefault(m)
	t.Cleanup(ag_cache.CloseAll)
}

func strLoader(v string) ag_cache.LoaderFunc[string] {
	return func(ctx context.Context, key string) (string, error) { return v, nil }
}

// ──────── POC 7: 基础语义回归 ────────

func TestGetOrElse_Basic(t *testing.T) {
	setupManager(t)
	cache := ag_cache.New[string]("users", strLoader("loaded"))
	ctx := context.Background()

	callCount := 0
	loader := func(ctx context.Context, key string) (string, error) {
		callCount++
		return "loaded-" + key, nil
	}
	_ = loader

	v, err := cache.GetOrElse(ctx, "key1", func(ctx context.Context, key string) (string, error) {
		callCount++
		return "loaded-" + key, nil
	})
	if err != nil || v != "loaded-key1" {
		t.Fatalf("first call: v=%q err=%v", v, err)
	}
	if callCount != 1 {
		t.Fatalf("loader called %d times, expected 1", callCount)
	}

	v, err = cache.Get(ctx, "key1")
	if err != nil || v != "loaded-key1" {
		t.Fatalf("hit read: v=%q err=%v", v, err)
	}
	if callCount != 1 {
		t.Fatalf("loader called %d times after hit, expected 1", callCount)
	}
}

func TestGet_PureRead(t *testing.T) {
	setupManager(t)
	cache := ag_cache.Get[string]("users")
	ctx := context.Background()

	_, err := cache.Get(ctx, "missing")
	if !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}
}

func TestSingleflight_LoaderCalledOnce(t *testing.T) {
	setupManager(t)
	cache := ag_cache.New[string]("users", strLoader("loaded"))
	ctx := context.Background()

	var mu sync.Mutex
	callCount := 0
	loader := func(ctx context.Context, key string) (string, error) {
		mu.Lock()
		callCount++
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
		return "loaded", nil
	}

	var wg sync.WaitGroup
	errs := make([]error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = cache.GetOrElse(ctx, "sf-key", loader)
		}(i)
	}
	wg.Wait()

	for i, e := range errs {
		if e != nil {
			t.Fatalf("goroutine %d: %v", i, e)
		}
	}
	if callCount != 1 {
		t.Fatalf("loader called %d times, expected 1", callCount)
	}
}

func TestSerialization_StructType(t *testing.T) {
	setupManager(t)
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	cache := ag_cache.New[User]("users", func(ctx context.Context, key string) (User, error) {
		return User{Name: "Alice", Age: 30}, nil
	})
	ctx := context.Background()

	user := User{Name: "Alice", Age: 30}
	if _, err := cache.GetOrElse(ctx, "u1", func(ctx context.Context, key string) (User, error) {
		return user, nil
	}); err != nil {
		t.Fatalf("GetOrElse: %v", err)
	}

	v, ok, err := cache.TryGet(ctx, "u1")
	if err != nil || !ok || v != user {
		t.Fatalf("tryget: v=%+v ok=%v err=%v", v, ok, err)
	}
}

// ──────── POC 2/3: 独立实例隔离与 Clear ────────

func TestIndependentInstances_Isolation(t *testing.T) {
	setupManager(t)
	ctx := context.Background()

	users := ag_cache.New[string]("users", strLoader("user-value"))
	params := ag_cache.New[string]("params", strLoader("param-value"))

	if _, err := users.GetOrElse(ctx, "shared-key", func(ctx context.Context, key string) (string, error) {
		return "user-value", nil
	}); err != nil {
		t.Fatalf("users load: %v", err)
	}
	if _, err := params.GetOrElse(ctx, "shared-key", func(ctx context.Context, key string) (string, error) {
		return "param-value", nil
	}); err != nil {
		t.Fatalf("params load: %v", err)
	}

	uv, _ := users.Get(ctx, "shared-key")
	pv, _ := params.Get(ctx, "shared-key")
	if uv != "user-value" || pv != "param-value" {
		t.Fatalf("isolation failed: users=%q params=%q", uv, pv)
	}
}

func TestClear_OnlyAffectsOwnInstance(t *testing.T) {
	setupManager(t)
	ctx := context.Background()

	users := ag_cache.Get[string]("users")
	params := ag_cache.Get[string]("params")

	users.GetOrElse(ctx, "u1", func(ctx context.Context, key string) (string, error) { return "U1", nil })
	params.GetOrElse(ctx, "p1", func(ctx context.Context, key string) (string, error) { return "P1", nil })

	params.Clear(ctx)

	_, errU := users.Get(ctx, "u1")
	_, errP := params.Get(ctx, "p1")

	if errU != nil {
		t.Fatalf("users should be unaffected by params.Clear, got err=%v", errU)
	}
	if !errors.Is(errP, ag_cache.ErrCacheMiss) {
		t.Fatalf("params should be cleared, got err=%v", errP)
	}
}

// ──────── P0-2: Engine error 通道 ────────

func TestBackendError_NotTreatedAsMiss(t *testing.T) {
	engine := ag_cache.NewMockEngine()
	engine.Err = errors.New("connection refused") // backend down
	ctx := context.Background()

	cache := ag_cache.NewWithEngine[string](engine)

	_, err := cache.Get(ctx, "k")
	if errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatal("backend error must NOT be treated as cache miss")
	}
	if !errors.Is(err, ag_cache.ErrBackend) {
		t.Fatalf("expected ErrBackend, got %v", err)
	}

	loaderCalled := false
	_, err = cache.GetOrElse(ctx, "k", func(ctx context.Context, key string) (string, error) {
		loaderCalled = true
		return "loaded", nil
	})
	if loaderCalled {
		t.Fatal("loader must NOT be called when backend is down (would storm the source)")
	}
	if !errors.Is(err, ag_cache.ErrBackend) {
		t.Fatalf("expected ErrBackend, got %v", err)
	}
}

func TestBackendError_RecoversWhenEngineHeals(t *testing.T) {
	engine := ag_cache.NewMockEngine()
	ctx := context.Background()
	cache := ag_cache.NewWithEngine[string](engine)

	engine.Err = errors.New("down")
	if _, err := cache.Get(ctx, "k"); !errors.Is(err, ag_cache.ErrBackend) {
		t.Fatalf("expected ErrBackend, got %v", err)
	}

	engine.Err = nil
	_, err := cache.Get(ctx, "k")
	if !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("after recovery, expected ErrCacheMiss, got %v", err)
	}
}

func TestErrBackend_PanicRecovery(t *testing.T) {
	engine := ag_cache.NewMockEngine()
	ctx := context.Background()
	cache := ag_cache.NewWithEngine[string](engine)

	engine.PanicNext = true
	_, err := cache.Get(ctx, "any-key")
	if !errors.Is(err, ag_cache.ErrBackend) {
		t.Fatalf("expected ErrBackend after engine panic, got %v", err)
	}

	engine.PanicNext = true
	err = cache.Set(ctx, "k", "v")
	if !errors.Is(err, ag_cache.ErrBackend) {
		t.Fatalf("Set: expected ErrBackend, got %v", err)
	}
}

// P2-C: loader 不被第一个调用者的 ctx 取消（WithoutCancel）
func TestLoader_NotCancelledByFirstCallerCtx(t *testing.T) {
	setupManager(t)
	cache := ag_cache.New[string]("users", strLoader("loaded"))

	loaderCalled := false
	loader := func(ctx context.Context, key string) (string, error) {
		loaderCalled = true
		select {
		case <-time.After(100 * time.Millisecond):
			return "loaded", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var wg sync.WaitGroup
	errs := make([]error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx == 0 {
				_, errs[idx] = cache.GetOrElse(ctx, "k", loader)
			} else {
				_, errs[idx] = cache.GetOrElse(context.Background(), "k", loader)
			}
		}(i)
	}
	wg.Wait()

	for i, e := range errs {
		if e != nil {
			t.Fatalf("goroutine %d: loader should still run despite first caller's cancelled ctx, got %v", i, e)
		}
	}
	if !loaderCalled {
		t.Fatal("loader should have been called once")
	}
}

// ──────── LoaderCache 语法糖 ────────

func TestLoaderCache_Get_ReadThrough(t *testing.T) {
	setupManager(t)
	ctx := context.Background()

	callCount := 0
	loader := func(ctx context.Context, key string) (string, error) {
		callCount++
		return "loaded-" + key, nil
	}
	users := ag_cache.New[string]("users", loader)

	v, err := users.Get(ctx, "u:1")
	if err != nil || v != "loaded-u:1" {
		t.Fatalf("first Get: v=%q err=%v", v, err)
	}
	if callCount != 1 {
		t.Fatalf("loader called %d times, expected 1", callCount)
	}

	v, err = users.Get(ctx, "u:1")
	if err != nil || v != "loaded-u:1" {
		t.Fatalf("second Get: v=%q err=%v", v, err)
	}
	if callCount != 1 {
		t.Fatalf("loader called %d times after hit, expected 1", callCount)
	}

	v, _ = users.Get(ctx, "u:2")
	if v != "loaded-u:2" {
		t.Fatalf("different key should reuse loader: v=%q", v)
	}
	if callCount != 2 {
		t.Fatalf("loader called %d times, expected 2", callCount)
	}
}

func TestLoaderCache_GetOrElse_CustomLoader(t *testing.T) {
	setupManager(t)
	ctx := context.Background()

	users := ag_cache.New[string]("users", func(ctx context.Context, key string) (string, error) {
		return "default-loader", nil
	})

	v, err := users.GetOrElse(ctx, "k", func(ctx context.Context, key string) (string, error) {
		return "custom-loader", nil
	})
	if err != nil || v != "custom-loader" {
		t.Fatalf("custom loader: v=%q err=%v", v, err)
	}
}

func TestLoaderCache_TryGet_NoLoader(t *testing.T) {
	setupManager(t)
	ctx := context.Background()

	callCount := 0
	users := ag_cache.New[string]("users", func(ctx context.Context, key string) (string, error) {
		callCount++
		return "v", nil
	})

	_, ok, err := users.TryGet(ctx, "missing")
	if err != nil || ok {
		t.Fatalf("TryGet miss: ok=%v err=%v", ok, err)
	}
	if callCount != 0 {
		t.Fatalf("TryGet must not call loader, called %d times", callCount)
	}

	users.Get(ctx, "k")
	_, ok, _ = users.TryGet(ctx, "k")
	if !ok {
		t.Fatal("TryGet after Get should hit")
	}
}

func TestLoaderCache_WithLoader(t *testing.T) {
	setupManager(t)
	ctx := context.Background()

	users := ag_cache.WithLoader(ag_cache.Get[string]("users"), func(ctx context.Context, key string) (string, error) {
		return "from-loader", nil
	})

	v, err := users.Get(ctx, "k")
	if err != nil || v != "from-loader" {
		t.Fatalf("WithLoader: v=%q err=%v", v, err)
	}
}

// ──────── NewWithEngine 底层 ────────

func TestNewWithEngine_ExplicitIsolation(t *testing.T) {
	engineA := ag_cache.NewMockEngine()
	engineB := ag_cache.NewMockEngine()
	ctx := context.Background()

	cacheA := ag_cache.NewWithEngine[string](engineA)
	cacheB := ag_cache.NewWithEngine[string](engineB)

	cacheA.Set(ctx, "k", "value-A")
	cacheB.Set(ctx, "k", "value-B")

	vA, _ := cacheA.Get(ctx, "k")
	vB, _ := cacheB.Get(ctx, "k")
	if vA != "value-A" || vB != "value-B" {
		t.Fatalf("engine isolation failed: A=%q B=%q", vA, vB)
	}
}

// ──────── WithEngine：选择指定引擎实现名 ────────

func TestWithEngine_SelectsEngine(t *testing.T) {
	registerMockEngine()
	counting := &countingFactory{}
	ag_cache.RegisterEngine(counting) // "counting"
	defer ag_cache.CloseAll()

	props := ag_cache.DefaultAgCacheProperties()
	props.DefaultEngine = "mock"
	m, err := ag_cache.NewManager(props)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	ag_cache.SetDefault(m)
	ctx := context.Background()

	// 默认引擎：mock，counting 不被调用
	ag_cache.New[string]("a", strLoader("x")).Get(ctx, "k")
	if counting.creates.Load() != 0 {
		t.Fatalf("default engine should be mock, counting creates=%d", counting.creates.Load())
	}

	// WithEngine("counting")：切到 counting 引擎
	ag_cache.New[string]("b", strLoader("y"), ag_cache.WithEngine[string]("counting")).Get(ctx, "k")
	if counting.creates.Load() != 1 {
		t.Fatalf("WithEngine should select counting, creates=%d", counting.creates.Load())
	}
}

// ──────── MockCache 测试替身 ────────

func TestMockCache_AsTestDouble(t *testing.T) {
	svc := struct {
		cache ag_cache.ICache[int]
	}{cache: ag_cache.NewMock[int]()}

	ctx := context.Background()
	v, _ := svc.cache.GetOrElse(ctx, "max-retries", func(ctx context.Context, key string) (int, error) {
		return 3, nil
	})
	if v != 3 {
		t.Fatalf("expected 3, got %d", v)
	}
}

// ──────── TryGet 行为 ────────

func TestTryGet_Miss(t *testing.T) {
	setupManager(t)
	c := ag_cache.Get[string]("users")
	ctx := context.Background()

	v, ok, err := c.TryGet(ctx, "missing")
	if err != nil || ok || v != "" {
		t.Fatalf("TryGet miss: v=%q ok=%v err=%v", v, ok, err)
	}
}

func TestTryGet_Hit(t *testing.T) {
	setupManager(t)
	c := ag_cache.New[string]("users", strLoader("v"))
	ctx := context.Background()
	c.GetOrElse(ctx, "k", strLoader("v"))

	v, ok, err := c.TryGet(ctx, "k")
	if err != nil || !ok || v != "v" {
		t.Fatalf("TryGet hit: v=%q ok=%v err=%v", v, ok, err)
	}
}

// ──────── Del 探测 BulkDelEngine ────────

// bulkDelEngine: 实现 BulkDelEngine，记录 DelMany 调用
type bulkDelEngine struct {
	*ag_cache.MockEngine
	mu      sync.Mutex
	delMany []string
}

func (e *bulkDelEngine) DelMany(ctx context.Context, keys ...string) error {
	e.mu.Lock()
	e.delMany = append(e.delMany, keys...)
	e.mu.Unlock()
	return nil
}

func (e *bulkDelEngine) lastDelMany() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]string(nil), e.delMany...)
}

func TestDel_BulkDelEngine(t *testing.T) {
	e := &bulkDelEngine{MockEngine: ag_cache.NewMockEngine()}
	c := ag_cache.NewWithEngine[string](e)
	ctx := context.Background()

	if err := c.Del(ctx, "a", "b", "c"); err != nil {
		t.Fatalf("Del: %v", err)
	}
	if got := e.lastDelMany(); len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("DelMany not used: %v", got)
	}
}

// ──────── 健壮性 2.1: double-check 用 WithoutCancel ────────

func TestGetOrElse_DoubleCheck_WithoutCancel(t *testing.T) {
	setupManager(t)
	c := ag_cache.New[string]("users", strLoader("loaded"))
	ctx := context.Background()

	var mu sync.Mutex
	callCount := 0
	loader := func(ctx context.Context, key string) (string, error) {
		mu.Lock()
		callCount++
		mu.Unlock()
		time.Sleep(80 * time.Millisecond)
		return "loaded", nil
	}

	// 首个调用者 ctx 取消，其余等待者正常
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // 立即取消首个调用者

	var wg sync.WaitGroup
	errs := make([]error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i == 0 {
				_, errs[i] = c.GetOrElse(cancelCtx, "k", loader)
			} else {
				_, errs[i] = c.GetOrElse(context.Background(), "k", loader)
			}
		}(i)
	}
	wg.Wait()

	// 等待者应成功加载（loader 用 WithoutCancel；double-check 也用 WithoutCancel）
	for i := 1; i < 5; i++ {
		if errs[i] != nil {
			t.Fatalf("waiter %d: %v", i, errs[i])
		}
	}
	mu.Lock()
	n := callCount
	mu.Unlock()
	if n != 1 {
		t.Fatalf("loader called %d times, want 1", n)
	}
}

// ──────── 健壮性 2.2: GetOrElse 写失败包 ErrBackend ────────

// setFailEngine: 包装 MockEngine，Get 走 miss，Set 返回固定错误。
type setFailEngine struct {
	*ag_cache.MockEngine
	fail error
}

func (e *setFailEngine) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if e.fail != nil {
		return e.fail
	}
	return e.MockEngine.Set(ctx, key, value, ttl)
}

func TestGetOrElse_SetFailure_ErrBackend(t *testing.T) {
	e := &setFailEngine{MockEngine: ag_cache.NewMockEngine(), fail: errors.New("persist: disk full")}
	c := ag_cache.NewWithEngine[string](e)
	ctx := context.Background()

	_, err := c.GetOrElse(ctx, "k", func(ctx context.Context, key string) (string, error) {
		return "loaded", nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ag_cache.ErrBackend) {
		t.Fatalf("expected ErrBackend, got %v", err)
	}
}

// ──────── 覆盖缺口 #2: TryGet 后端故障路径 ────────

func TestTryGet_BackendFailure(t *testing.T) {
	engine := ag_cache.NewMockEngine()
	engine.Err = errors.New("connection refused")
	c := ag_cache.NewWithEngine[string](engine)
	ctx := context.Background()

	v, ok, err := c.TryGet(ctx, "k")
	if ok || v != "" {
		t.Fatalf("TryGet backend failure: v=%q ok=%v", v, ok)
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ag_cache.ErrBackend) {
		t.Fatalf("expected ErrBackend, got %v", err)
	}
}

// ──────── 覆盖缺口 #3: Del 非 bulk 引擎多 key 中间失败短路 ────────

// delFailEngine: 包装 MockEngine，删除指定 key 时返回错误。
type delFailEngine struct {
	*ag_cache.MockEngine
	failKey string
}

func (e *delFailEngine) Del(ctx context.Context, key string) error {
	if key == e.failKey {
		return errors.New("del failed")
	}
	return e.MockEngine.Del(ctx, key)
}

func TestDel_Loop_StopsOnError(t *testing.T) {
	e := &delFailEngine{MockEngine: ag_cache.NewMockEngine(), failKey: "b"}
	c := ag_cache.NewWithEngine[string](e)
	ctx := context.Background()

	err := c.Del(ctx, "a", "b", "c") // b 失败 → 短路返回 ErrBackend
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ag_cache.ErrBackend) {
		t.Fatalf("expected ErrBackend, got %v", err)
	}
}
