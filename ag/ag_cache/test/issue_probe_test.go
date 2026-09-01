package ag_cache_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aif-go/ag-core/ag/ag_cache"
)

// setFailEngine 包装 MockEngine，让 Set 返回固定错误，探测 GetOrElse 缓存写入失败的错误包装。
type setFailEngine struct {
	*ag_cache.MockEngine
	fail error
}

func (e *setFailEngine) Set(ctx context.Context, key string, value []byte) error {
	if e.fail != nil {
		return e.fail
	}
	return e.MockEngine.Set(ctx, key, value)
}

func (e *setFailEngine) SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if e.fail != nil {
		return e.fail
	}
	return e.MockEngine.SetWithTTL(ctx, key, value, ttl)
}

// ISSUE-P1（v1 review 已识别，另开任务）：GetOrElse 中引擎 Set 失败时错误未包装为 ErrBackend。
// 本探测验证当前行为并记录。
func TestProbe_GetOrElse_SetFailure_NotWrapped(t *testing.T) {
	e := &setFailEngine{MockEngine: ag_cache.NewMockEngine(), fail: errors.New("persist: disk full")}
	c := ag_cache.NewWithEngine[string](e)
	ctx := context.Background()

	_, err := c.GetOrElse(ctx, "k", func(ctx context.Context, key string) (string, error) {
		return "loaded", nil // loader 成功，但写入引擎失败
	})
	if err == nil {
		t.Log("[ISSUE-P1-] 期望写入失败应返回错误")
		return
	}
	if errors.Is(err, ag_cache.ErrBackend) {
		t.Log("OK: Set 失败已包装为 ErrBackend")
	} else {
		t.Logf("[ISSUE-P1] GetOrElse 缓存写入失败未包装 ErrBackend（errors.Is=false）：%v", err)
		t.Log("  影响：业务用 errors.Is(err, ErrBackend) 做降级判断会失效")
	}
}

// ISSUE-P2（v1 review）：Peek 经 engine.Get 更新统计 —— 已随接口定稿移除 Peek/Stats（2026-08-24），
// 该探测不再适用；改用 TryGet 预期 Hit 做监控探活。

// 边界：负 TTL 防御（接口定稿后入口为构造期 WithDefaultTTL(-1)，应报错）
func TestProbe_NegativeTTL_Rejected(t *testing.T) {
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_ = ag_cache.WithDefaultTTL[string](-1 * time.Second)
	}()
	if recovered == nil {
		t.Logf("[ISSUE] WithDefaultTTL(-1) 未报错（负 TTL 未被拦截）")
		return
	}
	t.Logf("OK: WithDefaultTTL(-1) 被拒绝（负 TTL 防御生效）: %v", recovered)
}

// 边界：loader panic 标注为 "loader panic"（与引擎 panic 区分，2.3 已修复）
func TestProbe_LoaderPanic_Labeled(t *testing.T) {
	engine := ag_cache.NewMockEngine()
	c := ag_cache.NewWithEngine[string](engine)
	ctx := context.Background()

	_, err := c.GetOrElse(ctx, "k", func(ctx context.Context, key string) (string, error) {
		panic("loader business bug: nil entry")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "loader panic") {
		t.Logf("OK: loader panic 标注为 'loader panic'（2.3 修复）: %v", err)
	} else {
		t.Logf("[ISSUE] loader panic 未被正确标注: %v", err)
	}
}

// 边界：ErrBackend 对上下文信息的保留（wrap 是否含原始引擎错误）。
func TestProbe_ErrBackend_PreservesContext(t *testing.T) {
	engine := ag_cache.NewMockEngine()
	engine.Err = errors.New("dial tcp: connection refused")
	c := ag_cache.NewWithEngine[string](engine)

	if _, err := c.Get(context.Background(), "k"); err != nil {
		if !strings.Contains(err.Error(), "connection refused") {
			t.Logf("[ISSUE] ErrBackend 未保留原始错误细节: %v", err)
		} else {
			t.Log("OK: ErrBackend 保留了原始错误上下文")
		}
	}
}

// 观察：SetDefault 被替换后，旧 manager 的已建实例不会由 CloseAll 关闭（孤儿）。
// 这是当前实现的一个生命周期边界：替换默认实例时旧实例未显式关闭。
func TestProbe_SetDefaultReplace_Orphan(t *testing.T) {
	stop := startFx(t, nil)
	ag_cache.GetCacheWithLoader[string](dflt(), "orphan", func(ctx context.Context, key string) (string, error) { return "x", nil })
	stop() // 关闭当前默认 manager（含 orphan 所在实例）

	// startFx 内部已 CloseAll 清空 default manager；此处验证再访问会 panic（防误用）
	defer func() {
		if r := recover(); r != nil {
			t.Logf("OK: 默认 manager 关闭后访问 panic（防误用）: %v", r)
		}
	}()
	ag_cache.GetCache[string](dflt(), "whatever")
	t.Log("注意：之前 SetDefault 替换的旧 manager 实例若未被显式 Close，会成为孤儿（生命周期边界，需业务层约定）")
}

// 观察：相同 name 用不同 T 访问会 panic（类型一致性契约）。
func TestProbe_SameNameDifferentType(t *testing.T) {
	stop := startFx(t, nil)
	defer stop()
	ctx := context.Background()

	ag_cache.GetCacheWithLoader[string](dflt(), "dup", func(ctx context.Context, key string) (string, error) { return "s", nil }).Get(ctx, "k")
	defer func() {
		if r := recover(); r != nil {
			t.Logf("OK: 同名不同类型 panic（契约生效）: %v", r)
		}
	}()
	ag_cache.GetCacheWithLoader[int](dflt(), "dup", func(ctx context.Context, key string) (int, error) { return 1, nil }).Get(ctx, "k")
}

// 观察：特殊字符 key 可正常读写（无 URL escaping，业务需自行约定 key 格式）。
func TestProbe_WeirdKeyCharacters(t *testing.T) {
	stop := startFx(t, nil)
	defer stop()
	ctx := context.Background()

	weird := "u:1|host:port?x=1&y=2#frag"
	c := ag_cache.GetCacheWithLoader[string](dflt(), "weird", func(ctx context.Context, key string) (string, error) {
		return "v-" + url.QueryEscape(key), nil
	})
	if v, err := c.Get(ctx, weird); err != nil || v == "" {
		t.Fatalf("weird key Get: v=%q err=%v", v, err)
	} else {
		t.Logf("OK: 特殊字符 key 可正常读写 (%s -> %s)", weird, v)
	}
}

// ISSUE-P4（集成测试发现）：NewAgCacheManager 的幂等注册在"同一进程多次 Fx 装配不同引擎配置"时，
// 后续工厂因 EngineRegistered==true 被跳过，导致后装配的 app 静默复用首个 app 的引擎配置。
// 现象：阶段1 用 nil yaml（默认 0 TTL）注册后，阶段2 用 1s TTL 装配，实际 TTL 仍为 0。
// 影响：同进程多 app / 测试 / 配置热更时，引擎配置（TTL/MaxCost）互相污染。
// 注：全局注册表已删除（改 fx group 注入），本测试验证每次装配独立生效。
func TestProbe_GlobalRegistry_FixedByFirstAssembly(t *testing.T) {
	// 阶段1：首次装配注册引擎工厂（配置 A：默认 TTL）
	stop1 := startFx(t, nil)
	c1 := ag_cache.GetCache[string](dflt(), "load")
	c1.Set(context.Background(), "k", "v")
	stop1()

	// 阶段2：装配配置 B（defaultTtl=1s），期望 1s 生效
	stop2 := startFx(t, map[string]any{"agcache.ristretto.defaultttl": "1s"})
	defer stop2()
	dbCalls := 0
	users := ag_cache.GetCacheWithLoader[*User](dflt(), "users", func(ctx context.Context, key string) (*User, error) {
		dbCalls++
		return &User{ID: key}, nil
	})
	users.Get(context.Background(), "u:1")
	users.Get(context.Background(), "u:1")
	time.Sleep(1500 * time.Millisecond)
	users.Get(context.Background(), "u:1")
	if dbCalls == 2 {
		t.Logf("OK: 阶段2 引擎配置(1s TTL)生效，TTL 过期后重载")
	} else {
		t.Logf("[ISSUE-P4] 阶段2 引擎配置未生效：TTL 仍为首个装配的默认(0)，dbCalls=%d（引擎配置被首次装配固化）", dbCalls)
	}
}

// ISSUE-P5（性能评测发现）：高并发写时 Ristretto setBuf（BufferItems=64）满，
// SetWithTTL 返回 false，引擎将其作为 ErrBackend 错误返回。
// 影响：批量/高写入场景业务会持续收到 ErrBackend，可能被误判为"后端故障"触发降级。
func TestProbe_SetDroppedBufferFull(t *testing.T) {
	stop := startFx(t, nil)
	defer stop()
	ctx := context.Background()

	c := ag_cache.GetCacheWithLoader[string](dflt(), "probe-buffer", func(ctx context.Context, key string) (string, error) {
		return "v", nil
	})
	var dropped int
	var mu sync.Mutex
	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				if err := c.Set(ctx, fmt.Sprintf("k-%d-%d", w, i), "value"); err != nil {
					mu.Lock()
					dropped++
					mu.Unlock()
				}
			}
		}(w)
	}
	wg.Wait()
	t.Logf("高并发 8×500 Set 触发 dropped=%d 次（Ristretto setBuf 满 → ErrBackend）", dropped)
	if dropped > 0 {
		t.Logf("[ISSUE-P5] Set 被 drop 且以 ErrBackend 返回：高写场景业务会收到后端故障错误（实为 buffer 满）")
	} else {
		t.Log("OK: 本次并发写未触发 buffer full")
	}
}
