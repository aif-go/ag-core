package ag_cache_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/aif-go/ag-core/ag/ag_cache"
	"github.com/aif-go/ag-core/ag/ag_cache/agristretto"
)

// 构造真实引擎 + core 缓存（bench 关注核心读/写路径）
func benchCache(b *testing.B) ag_cache.ICache[string] {
	b.Helper()
	e, err := agristretto.NewRistrettoEngine(agristretto.RistrettoOptions{MaxCost: 1 << 30})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = e.Close() })
	return ag_cache.NewWithEngine[string](e)
}

// ---- 读路径 ----

func Benchmark_GetOrElse_Hit(b *testing.B) {
	c := benchCache(b)
	ctx := b.Context()
	// 预热：用 GetOrElse 触发 loader+内部 Sync，确保 "hot" 已缓存可见（避免异步 Set 未完成）
	if _, err := c.GetOrElse(ctx, "hot", func(ctx context.Context, key string) (string, error) {
		return "v", nil
	}); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := c.Get(ctx, "hot"); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_GetOrElse_Miss(b *testing.B) {
	c := benchCache(b)
	ctx := b.Context()
	// 写缓存路径：miss → loader → set
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			key := fmt.Sprintf("miss-%d", i&0xffff)
			if _, err := c.GetOrElse(ctx, key, func(ctx context.Context, key string) (string, error) {
				return "loaded", nil
			}); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_Get_PureRead(b *testing.B) {
	c := benchCache(b)
	ctx := b.Context()
	for i := 0; i < 1024; i++ {
		c.Set(ctx, fmt.Sprintf("k-%d", i), "v")
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			if _, err := c.Get(ctx, fmt.Sprintf("k-%d", i&1023)); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---- 写路径 ----

func Benchmark_Set(b *testing.B) {
	c := benchCache(b)
	ctx := b.Context()
	// 并发高写可能触发 Ristretto setBuf 满（BufferItems=64）→ SetWithTTL 返回 false，
	// 引擎将其作为 ErrBackend 返回（真实行为，见 issue_probe TestProbe_SetDroppedBufferFull）。
	// 此处统计 dropped 次数而非 Fatal，衡量尽力而为写入吞吐。
	dropped := int64(0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			if err := c.Set(ctx, fmt.Sprintf("k-%d", i&0xffff), "value"); err != nil {
				atomic.AddInt64(&dropped, 1)
			}
		}
	})
	if dropped > 0 {
		b.Logf("Set dropped (buffer full): %d 次（高并发写触发 Ristretto setBuf 满）", dropped)
	}
}

func Benchmark_Del(b *testing.B) {
	c := benchCache(b)
	ctx := b.Context()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			_ = c.Del(ctx, fmt.Sprintf("k-%d", i&0xffff))
		}
	})
}

func Benchmark_Clear(b *testing.B) {
	c := benchCache(b)
	ctx := b.Context()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := c.Clear(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// ---- 序列化：基础类型 fast path vs JSON 回退 ----

func Benchmark_Serializer_String(b *testing.B) {
	s := ag_cache.DefaultSerializer[string]()
	for i := 0; i < b.N; i++ {
		data, _ := s.Marshal("hello")
		_, _ = s.Unmarshal(data)
	}
}

func Benchmark_Serializer_Int64(b *testing.B) {
	s := ag_cache.DefaultSerializer[int64]()
	for i := 0; i < b.N; i++ {
		data, _ := s.Marshal(42)
		_, _ = s.Unmarshal(data)
	}
}

func Benchmark_Serializer_Bytes(b *testing.B) {
	s := ag_cache.DefaultSerializer[[]byte]()
	payload := []byte("payload-payload-payload")
	for i := 0; i < b.N; i++ {
		data, _ := s.Marshal(payload)
		_, _ = s.Unmarshal(data)
	}
}

func Benchmark_Serializer_Struct(b *testing.B) {
	s := ag_cache.DefaultSerializer[User]()
	u := User{ID: "u:1", Name: "Alice"}
	for i := 0; i < b.N; i++ {
		data, _ := s.Marshal(u)
		_, _ = s.Unmarshal(data)
	}
}
