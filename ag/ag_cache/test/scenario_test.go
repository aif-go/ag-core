package ag_cache_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/aif-go/ag-core/ag/ag_cache"
)

// ---------------------------------------------------------------------------
// 业务场景集成测试（真实 agristretto 引擎 + fx 装配全链路）
// ---------------------------------------------------------------------------

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Param struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// 用户缓存 Cache-Aside：miss→loader→缓存→命中→TTL 过期重载
// TTL 用 WithDefaultTTL(1s) 显式指定（不依赖引擎 defaultttl，规避全局引擎注册表被
// 首次装配固化的 ISSUE-P4，保证测试间 TTL 独立）。
func TestScenario_UserCache_CacheAside(t *testing.T) {
	stop := startFx(t, nil)
	defer stop()
	ctx := context.Background()

	db := map[string]*User{"u:1": {ID: "u:1", Name: "Alice"}}
	var dbCalls int
	var mu sync.Mutex
	users := ag_cache.New[*User]("users", func(ctx context.Context, key string) (*User, error) {
		mu.Lock()
		dbCalls++
		mu.Unlock()
		u, ok := db[key]
		if !ok {
			return nil, ag_cache.ErrCacheMiss
		}
		return u, nil
	}, ag_cache.WithDefaultTTL[*User](1*time.Second))

	// miss → loader
	u, err := users.Get(ctx, "u:1")
	if err != nil || u.ID != "u:1" {
		t.Fatalf("first Get: u=%+v err=%v", u, err)
	}
	// hit → 不调 loader
	u, err = users.Get(ctx, "u:1")
	if err != nil || u.Name != "Alice" {
		t.Fatalf("hit Get: u=%+v err=%v", u, err)
	}
	if dbCalls != 1 {
		t.Fatalf("dbCalls = %d, want 1 (hit should not hit db)", dbCalls)
	}

	// TTL 过期 → 重新加载（异步 eviction 需要时间）
	time.Sleep(1500 * time.Millisecond)
	u, err = users.Get(ctx, "u:1")
	if err != nil || u.ID != "u:1" {
		t.Fatalf("reload after TTL: u=%+v err=%v", u, err)
	}
	mu.Lock()
	c := dbCalls
	mu.Unlock()
	if c != 2 {
		t.Fatalf("dbCalls = %d, want 2 (TTL expiry should reload)", c)
	}
}

// 参数缓存：Clear 只清自己，不影响其他 namespace；清后重载新值
func TestScenario_ParamCache_ClearReload(t *testing.T) {
	stop := startFx(t, map[string]any{
		"agcache.defaultengine":     "ristretto",
		"agcache.ristretto.maxcost": "104857600",
	})
	defer stop()
	ctx := context.Background()

	paramStore := map[string]string{"host:port": "10.0.0.1:8080"}
	users := ag_cache.New[*User]("users", func(ctx context.Context, key string) (*User, error) {
		return &User{ID: key, Name: "U"}, nil
	})
	params := ag_cache.New[*Param]("params", func(ctx context.Context, key string) (*Param, error) {
		v, ok := paramStore[key]
		if !ok {
			return nil, ag_cache.ErrCacheMiss
		}
		return &Param{Key: key, Value: v}, nil
	})

	users.Get(ctx, "u:1") // warm users
	p, err := params.Get(ctx, "host:port")
	if err != nil || p.Value != "10.0.0.1:8080" {
		t.Fatalf("param first Get: p=%+v err=%v", p, err)
	}

	// 参数更新（数据源变更）
	paramStore["host:port"] = "10.0.0.2:8080"
	params.Clear(ctx)
	p, err = params.Get(ctx, "host:port")
	if err != nil || p.Value != "10.0.0.2:8080" {
		t.Fatalf("param after clear+broadcast: p=%+v err=%v", p, err)
	}

	// users 不被 params.Clear 影响
	if u, err := users.Get(ctx, "u:1"); err != nil || u.Name != "U" {
		t.Fatalf("users affected by params.Clear: u=%+v err=%v", u, err)
	}
}

// 主动刷新：Set 覆盖 + Del 后 miss→loader
// 重要：engin.Set 为异步写；主动 Set 后需等可见。验证用 Admin 纯读视角，避免
// LoaderCache.Get 读穿透在异步窗口调 loader 覆盖 Set 的值。
func TestScenario_ForceRefresh_SetAndDel(t *testing.T) {
	stop := startFx(t, nil)
	defer stop()
	ctx := context.Background()

	var loaded int
	c := ag_cache.New[string]("cfg", func(ctx context.Context, key string) (string, error) {
		loaded++
		return "from-db", nil
	})
	reader := ag_cache.Get[string]("cfg") // 纯读视角（同实例，TryGet 不调 loader）

	c.Set(ctx, "k", "v1")
	if v := waitVisible(t, reader, "k"); v != "v1" {
		t.Fatalf("after Set: got %q, want v1", v)
	}
	c.Set(ctx, "k", "v2") // overwrite
	if v := waitVisible(t, reader, "k"); v != "v2" {
		t.Fatalf("after overwrite Set: got %q, want v2", v)
	}
	c.Del(ctx, "k")
	waitGone(t, reader, "k")
	if v, err := c.Get(ctx, "k"); err != nil || v != "from-db" {
		t.Fatalf("after Del reload: v=%q err=%v", v, err)
	}
	if loaded != 1 {
		t.Fatalf("loaded = %d, want 1", loaded)
	}
}

// waitVisible 重试 TryGet 直到命中（Ristretto 异步写窗口内轮询）。
func waitVisible(t *testing.T, c ag_cache.ICache[string], key string) string {
	t.Helper()
	for i := 0; i < 50; i++ {
		if v, ok, _ := c.TryGet(context.Background(), key); ok {
			return v
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("key %q not visible after 500ms", key)
	return ""
}

func waitGone(t *testing.T, c ag_cache.ICache[string], key string) {
	t.Helper()
	for i := 0; i < 50; i++ {
		if _, ok, _ := c.TryGet(context.Background(), key); !ok {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("key %q still visible after 500ms", key)
}

// 负缓存业务模式（穿透防护）：峰值前命中负缓存不调 loader
func TestScenario_NegativeCachePattern(t *testing.T) {
	stop := startFx(t, nil)
	defer stop()
	ctx := context.Background()

	notExist := ag_cache.New[bool]("user-notexist", func(ctx context.Context, key string) (bool, error) {
		return false, ag_cache.ErrCacheMiss // 无标记 → miss（纯读，不写缓存）
	})
	var dbCalls int
	users := ag_cache.New[*User]("users", func(ctx context.Context, key string) (*User, error) {
		dbCalls++
		if key == "u:999" {
			notExist.Set(ctx, key, true) // 记录"不存在"（默认 TTL）
			return nil, ag_cache.ErrCacheMiss
		}
		return &User{ID: key, Name: "Alice"}, nil
	})

	// 业务查询封装：负缓存标记命中（30s 内）→ 直接短路，不调主缓存
	query := func(key string) error {
		if _, err := notExist.Get(ctx, key); err == nil {
			return ag_cache.ErrCacheMiss // 负缓存命中 → 视为不存在
		}
		_, err := users.Get(ctx, key)
		return err
	}

	// 第一次：负缓存 miss → 主缓存 loader 记负缓存并返回 miss
	if err := query("u:999"); !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("first query: want ErrCacheMiss, got %v", err)
	}
	if dbCalls != 1 {
		t.Fatalf("dbCalls = %d, want 1", dbCalls)
	}

	// 等待负缓存标记异步写入可见
	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		if _, err := notExist.Get(ctx, "u:999"); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("negative cache marker not visible within 500ms")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// 第二次起：负缓存命中 → 短路，主缓存 loader 不再调用
	dbCalls = 0
	for i := 0; i < 10; i++ {
		if err := query("u:999"); !errors.Is(err, ag_cache.ErrCacheMiss) {
			t.Fatalf("shielded query %d: want miss, got %v", i, err)
		}
	}
	if dbCalls != 0 {
		t.Fatalf("negative cache should shield db, dbCalls=%d", dbCalls)
	}
}

// 基础类型序列化 + 类型安全（用 TryGet 纯读，规避 Set 异步窗口）
func TestScenario_BasicTypeSerialization(t *testing.T) {
	stop := startFx(t, nil)
	defer stop()
	ctx := context.Background()

	str := ag_cache.New[string]("s", func(ctx context.Context, key string) (string, error) { return "sv", nil })
	i64 := ag_cache.New[int64]("i64", func(ctx context.Context, key string) (int64, error) { return 42, nil })
	b := ag_cache.New[bool]("b", func(ctx context.Context, key string) (bool, error) { return true, nil })
	bytes := ag_cache.New[[]byte]("bytes", func(ctx context.Context, key string) ([]byte, error) { return []byte("bv"), nil })

	strReader := ag_cache.Get[string]("s")
	i64Reader := ag_cache.Get[int64]("i64")
	bReader := ag_cache.Get[bool]("b")
	bytesReader := ag_cache.Get[[]byte]("bytes")

	str.Set(ctx, "k", "x")
	if v := waitT(t, strReader, "k"); v != "x" {
		t.Fatalf("string roundtrip: %q", v)
	}
	i64.Set(ctx, "k", 99)
	if v := waitI64(t, i64Reader, "k"); v != 99 {
		t.Fatalf("int64 roundtrip: %d", v)
	}
	b.Set(ctx, "k", false)
	if v := waitB(t, bReader, "k"); v {
		t.Fatalf("bool roundtrip: %v", v)
	}
	bytes.Set(ctx, "k", []byte("uv"))
	if v := waitBytes(t, bytesReader, "k"); string(v) != "uv" {
		t.Fatalf("bytes roundtrip: %q", v)
	}
}

func waitT(t *testing.T, c ag_cache.ICache[string], key string) string {
	t.Helper()
	for i := 0; i < 50; i++ {
		if v, ok, _ := c.TryGet(context.Background(), key); ok {
			return v
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("string not visible")
	return ""
}

func waitI64(t *testing.T, c ag_cache.ICache[int64], key string) int64 {
	t.Helper()
	for i := 0; i < 50; i++ {
		if v, ok, _ := c.TryGet(context.Background(), key); ok {
			return v
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("int64 not visible")
	return 0
}

func waitB(t *testing.T, c ag_cache.ICache[bool], key string) bool {
	t.Helper()
	for i := 0; i < 50; i++ {
		if v, ok, _ := c.TryGet(context.Background(), key); ok {
			return v
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("bool not visible")
	return false
}

func waitBytes(t *testing.T, c ag_cache.ICache[[]byte], key string) []byte {
	t.Helper()
	for i := 0; i < 50; i++ {
		if v, ok, _ := c.TryGet(context.Background(), key); ok {
			return v
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("bytes not visible")
	return nil
}

// 监控探活：TryGet 预期 Hit（Stats 已后置移除，用 TryGet 判断命中）
func TestScenario_Probe_WithTryGet(t *testing.T) {
	stop := startFx(t, nil)
	defer stop()
	ctx := context.Background()

	c := ag_cache.New[string]("probe", func(ctx context.Context, key string) (string, error) {
		return "uv", nil
	})

	// miss → ok=false（不调 loader）
	if _, ok, err := c.TryGet(ctx, "k"); err != nil || ok {
		t.Fatalf("TryGet before load: ok=%v err=%v", ok, err)
	}
	// 加载后命中
	c.GetOrElse(ctx, "k", func(ctx context.Context, key string) (string, error) { return "uv", nil })
	if v, ok, err := c.TryGet(ctx, "k"); err != nil || !ok || v != "uv" {
		t.Fatalf("TryGet after load: v=%q ok=%v err=%v", v, ok, err)
	}
}
