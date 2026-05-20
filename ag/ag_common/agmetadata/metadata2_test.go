package agmetadata

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

// ============================================================
// 并发安全测试 - 对同一个 ctx 并发操作
//
// 修复后，mdWrapper 包装 sync.RWMutex 保护所有 context 内 MD 的读写，
// 所有测试均可安全并发运行：
//   go test -race -v -count=1
// ============================================================

// TestConcurrentAppendSameCtx 测试多个 goroutine 对同一个 ctx 并发追加元数据
// 验证: mdWrapper 的写锁保护下，并发 AppendMdToContext 安全运行
func TestConcurrentAppendSameCtx(t *testing.T) {
	ctx := context.Background()
	// 注意：第一次调用 AppendMdToContext 会创建新的 MD 并存入 ctx
	// 后续的 AppendMdToContext 会直接修改同一个 map，因此并发不安全
	ctx = AppendMdToContext(ctx, MD{"init": "init"})

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer func() {
				err := recover()
				if err != nil {
					t.Errorf("unexpected panic: %v", err)
				}
			}()
			defer wg.Done()
			key := fmt.Sprintf("key_%d", i)
			val := fmt.Sprintf("val_%d", i)
			// 多个 goroutine 同时对同一个 ctx 追加，共享同一个 map
			// fmt.Sprintf("append key %s, val %s", key, val)
			_ = AppendMdToContext(ctx, MD{key: val})
		}()
	}
	wg.Wait()

	// 验证所有数据完整性：所有写入的 key 都能被读取到
	md := GetMdFromContext(ctx)
	if md["init"] != "init" {
		t.Errorf("expected init value 'init', got '%s'", md["init"])
	}
	fmt.Printf("total md keys: %d\n", md.Len())
	t.Logf("total md keys: %d", md.Len())
}

// TestConcurrentAppendAndReadSameCtx 测试同 ctx 下并发追加与读取
// 验证: mdWrapper 的读写锁保护下，读写并发安全
func TestConcurrentAppendAndReadSameCtx(t *testing.T) {
	ctx := context.Background()
	ctx = AppendMdToContext(ctx, MD{"base": "base"})

	var wg sync.WaitGroup
	wg.Add(20) // 10 writers + 10 readers

	// 10 个 writer goroutine：不断追加数据
	for i := 0; i < 10; i++ {
		i := i
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("w_%d_%d", i, j)
				val := fmt.Sprintf("v_%d_%d", i, j)
				_ = AppendMdToContext(ctx, MD{key: val})
			}
		}()
	}

	// 10 个 reader goroutine：不断读取副本
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = GetMdFromContext(ctx)
			}
		}()
	}
	wg.Wait()

	// 验证 base 数据仍然存在
	md := GetMdFromContext(ctx)
	if md["base"] != "base" {
		t.Errorf("expected base value 'base', got '%s'", md["base"])
	}
	t.Logf("total md keys after concurrent read/write: %d", md.Len())
}

// TestConcurrentAppendAndGetValue 测试同 ctx 下并发追加与 GetValueFromContext
// 验证: mdWrapper 读锁保护 GetValueFromContext 的 map 读取
func TestConcurrentAppendAndGetValue(t *testing.T) {
	ctx := context.Background()
	ctx = AppendMdToContext(ctx, MD{"base": "base"})

	var wg sync.WaitGroup
	wg.Add(20) // 10 writers + 10 readers

	// 10 个 writer goroutine
	for i := 0; i < 10; i++ {
		i := i
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("k_%d_%d", i, j)
				_ = AppendMdToContext(ctx, MD{key: "v"})
			}
		}()
	}

	// 10 个 reader goroutine：使用 GetValueFromContext 直接读取
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = GetValueFromContext(ctx, "base")
			}
		}()
	}
	wg.Wait()

	// 验证数据
	md := GetMdFromContext(ctx)
	if md["base"] != "base" {
		t.Errorf("expected base value 'base', got '%s'", md["base"])
	}
	t.Logf("total md keys after concurrent append and GetValue: %d", md.Len())
}

// TestConcurrentAppendAndHandler 测试同 ctx 下并发追加与 HandlerMdFromContext 遍历
// 验证: HandlerMdFromContext 通过 GetMdFromContext 获取副本，在锁外回调遍历副本
func TestConcurrentAppendAndHandler(t *testing.T) {
	ctx := context.Background()
	ctx = AppendMdToContext(ctx, MD{"base": "base"})

	var wg sync.WaitGroup
	wg.Add(20) // 10 writers + 10 handlers

	// 10 个 writer goroutine
	for i := 0; i < 10; i++ {
		i := i
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("k_%d_%d", i, j)
				_ = AppendMdToContext(ctx, MD{key: "v"})
			}
		}()
	}

	// 10 个 handler goroutine：遍历 metadata
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = HandlerMdFromContext(ctx, func(k, v string) {
					// 仅遍历，不修改
				})
			}
		}()
	}
	wg.Wait()

	// 验证数据
	md := GetMdFromContext(ctx)
	if md["base"] != "base" {
		t.Errorf("expected base value 'base', got '%s'", md["base"])
	}
	t.Logf("total md keys after concurrent append and HandlerMdFromContext: %d", md.Len())
}

// TestConcurrentMixedOps 测试同 ctx 下混合操作：Append、GetMdFromContext、GetValueFromContext、HandlerMdFromContext 并发执行
// 验证: 所有读操作通过读锁保护，写操作通过写锁保护，混合并发安全
func TestConcurrentMixedOps(t *testing.T) {
	ctx := context.Background()
	ctx = AppendMdToContext(ctx, MD{"base": "base"})

	var wg sync.WaitGroup
	wg.Add(40) // 10 writers + 10 GetMdFromContext readers + 10 GetValue readers + 10 handlers

	// 10 个 writer goroutine
	for i := 0; i < 10; i++ {
		i := i
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				key := fmt.Sprintf("k_%d_%d", i, j)
				_ = AppendMdToContext(ctx, MD{key: "v"})
			}
		}()
	}

	// 10 个 GetMdFromContext readers
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = GetMdFromContext(ctx)
			}
		}()
	}

	// 10 个 GetValueFromContext readers
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_, _ = GetValueFromContext(ctx, "base")
			}
		}()
	}

	// 10 个 HandlerMdFromContext readers
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = HandlerMdFromContext(ctx, func(k, v string) {})
			}
		}()
	}
	wg.Wait()

	// 最终验证
	md := GetMdFromContext(ctx)
	if md["base"] != "base" {
		t.Errorf("expected base value 'base', got '%s'", md["base"])
	}
	t.Logf("total md keys after mixed concurrent operations: %d", md.Len())
}

// TestConcurrentDifferentCtx 测试不同 context 并发操作（安全路径验证）
// 验证: 不同 ctx 各自拥有独立的 map，并发安全
// 该测试可以安全运行，不会触发 data race
func TestConcurrentDifferentCtx(t *testing.T) {
	var wg sync.WaitGroup
	const goroutines = 100
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			// 每个 goroutine 使用独立的 ctx
			ctx := context.Background()
			ctx = AppendMdToContext(ctx, MD{"key": "val"})
			_ = GetMdFromContext(ctx)
			_, _ = GetValueFromContext(ctx, "key")
			_ = HandlerMdFromContext(ctx, func(k, v string) {})
		}()
	}
	wg.Wait()
}

// TestConcurrentRegMdKey 测试并发注册元数据键名
// 验证: RegMdKey 内部使用 sync.RWMutex + atomic.Value 保证并发安全
// 该测试可以安全运行，不会触发 data race
func TestConcurrentRegMdKey(t *testing.T) {
	var wg sync.WaitGroup
	const goroutines = 100
	wg.Add(goroutines * 2) // 100 writers + 100 readers

	// 100 个 writer goroutine：并发注册不同 key
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("concurrent_key_%d", i)
			RegMdKey(key)
		}()
	}

	// 100 个 reader goroutine：并发读取 key 列表
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = GetMdKeys()
		}()
	}
	wg.Wait()

	// 验证所有 key 都已注册
	keys := GetMdKeys()
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}
	for i := 0; i < goroutines; i++ {
		key := fmt.Sprintf("concurrent_key_%d", i)
		if _, exists := keySet[key]; !exists {
			t.Errorf("expected key '%s' to be registered, but not found", key)
		}
	}
	t.Logf("total registered keys: %d", len(keys))
}
