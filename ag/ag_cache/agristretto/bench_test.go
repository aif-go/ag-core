package agristretto_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aif-go/ag-core/ag/ag_cache"
	"github.com/aif-go/ag-core/ag/ag_cache/agristretto"
	"github.com/dgraph-io/ristretto/v2"
)

// 纯 Ristretto 异步写（不 Wait）—— 引擎理论性能上限
func BenchmarkRistrettoAsyncSet(b *testing.B) {
	cache, _ := ristretto.NewCache[string, []byte](&ristretto.Config[string, []byte]{
		NumCounters: 1 << 20,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		cache.SetWithTTL(key, []byte("value"), 1, time.Minute)
	}
}

// engine.Set（独立实例，无 Wait，无索引）—— 当前实现的写性能
func BenchmarkEngineSet(b *testing.B) {
	e, _ := agristretto.NewRistrettoEngine(agristretto.RistrettoConfig{MaxCost: 1 << 30})
	defer e.Close()
	ctx := b.Context()
	cache := ag_cache.NewWithEngine[string](e)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, fmt.Sprintf("key-%d", i), "value")
	}
}

// GetOrElse miss → loader → Set（写路径完整链路）
func BenchmarkGetOrElse_Miss(b *testing.B) {
	e, _ := agristretto.NewRistrettoEngine(agristretto.RistrettoConfig{MaxCost: 1 << 30})
	defer e.Close()
	ctx := b.Context()
	cache := ag_cache.NewWithEngine[string](e)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("miss-%d", i)
		cache.GetOrElse(ctx, key, func(ctx context.Context, key string) (string, error) {
			return "loaded", nil
		})
	}
}

// GetOrElse 命中路径
func BenchmarkGetOrElse_Hit(b *testing.B) {
	e, _ := agristretto.NewRistrettoEngine(agristretto.RistrettoConfig{MaxCost: 1 << 30})
	defer e.Close()
	ctx := b.Context()
	cache := ag_cache.NewWithEngine[string](e)
	cache.Set(ctx, "hot-key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetOrElse(ctx, "hot-key", func(ctx context.Context, key string) (string, error) {
			return "loaded", nil
		})
	}
}
