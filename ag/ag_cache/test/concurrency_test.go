package ag_cache_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aif-go/ag-core/ag/ag_cache"
)

// 并发 miss 合并：10 goroutine 并发加载同一 key，loader 只调用 1 次
func TestConcurrency_Singleflight(t *testing.T) {
	stop := startFx(t, nil)
	defer stop()
	ctx := context.Background()

	var loads atomic.Int32
	c := ag_cache.GetCacheWithLoader[string](dflt(), "svc", func(ctx context.Context, key string) (string, error) {
		loads.Add(1)
		time.Sleep(80 * time.Millisecond)
		return "loaded:" + key, nil
	})

	const n = 10
	var wg sync.WaitGroup
	errs := make([]error, n)
	vals := make([]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v, err := c.Get(ctx, "hot-key")
			vals[i], errs[i] = v, err
		}(i)
	}
	wg.Wait()

	for i := range errs {
		if errs[i] != nil {
			t.Fatalf("goroutine %d: %v", i, errs[i])
		}
		if vals[i] != "loaded:hot-key" {
			t.Fatalf("goroutine %d got %q", i, vals[i])
		}
	}
	if got := loads.Load(); got != 1 {
		t.Fatalf("loader called %d times, want 1 (singleflight)", got)
	}
}

// 高并发混合读写稳定性（配合 -race 检查数据竞争）
// 同一缓存同时被多个 reader/writer 访问，不得 panic 或返回不一致数据。
func TestConcurrency_HighVolumeStability(t *testing.T) {
	stop := startFx(t, nil)
	defer stop()
	ctx := context.Background()

	c := ag_cache.GetCache[string](dflt(), "load")
	for i := 0; i < 50; i++ { // warm
		c.Set(ctx, fmt.Sprintf("k-%d", i), fmt.Sprintf("v-%d", i))
	}

	var wg sync.WaitGroup
	var failures atomic.Int32

	// 20 读、5 写
	for w := 0; w < 20; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				k := fmt.Sprintf("k-%d", (w*7+i)%50)
				if _, err := c.Get(ctx, k); err != nil && !errors.Is(err, ag_cache.ErrCacheMiss) {
					failures.Add(1)
					return
				}
			}
		}(w)
	}
	for w := 0; w < 5; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				k := fmt.Sprintf("k-%d", (w*13+i)%50)
				c.Set(ctx, k, fmt.Sprintf("new-%d", i))
				if i%20 == 0 {
					c.Del(ctx, k)
				}
			}
		}(w)
	}
	wg.Wait()

	if failures.Load() != 0 {
		t.Fatalf("%d reads failed during concurrent read/write", failures.Load())
	}
}

// 并发重建：并发 Get 一个 miss 的 key，全部应通过 singleflight 合并为一次 loader。
// 不依赖"Del 后立即 miss"的异步时序（Ristretto Del 为异步，race 下可能未生效），
// 直接对未缓存 key 并发加载，验证 singleflight 合并。
func TestConcurrency_DeleteThenConcurrentReload(t *testing.T) {
	stop := startFx(t, nil)
	defer stop()
	ctx := context.Background()

	var loads atomic.Int32
	c := ag_cache.GetCacheWithLoader[string](dflt(), "svc", func(ctx context.Context, key string) (string, error) {
		loads.Add(1)
		return "reloaded", nil
	})

	// 从未写入 → 天然 miss；并发 Get 触发 singleflight 合并
	const n = 8
	var wg sync.WaitGroup
	errs := make([]error, n)
	vals := make([]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			vals[i], errs[i] = c.Get(ctx, "k")
		}(i)
	}
	wg.Wait()
	for i := range errs {
		if errs[i] != nil {
			t.Fatalf("goroutine %d rebuild failed: %v", i, errs[i])
		}
		if vals[i] != "reloaded" {
			t.Fatalf("goroutine %d got %q, want reloaded", i, vals[i])
		}
	}
	if loads.Load() != 1 {
		t.Fatalf("concurrent miss should be singleflight, loads=%d", loads.Load())
	}
}
