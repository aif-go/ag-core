package future

import (
	"context"
	"testing"
)

// go test -bench=. -benchmem -run "^$" ./ 2>&1

// =============================================================================
// Benchmarks — 三种 future 方式性能对比
// =============================================================================

// --- 单任务延迟 (no-op task) ---

func BenchmarkNewFuture_NoOp(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		fut := NewFuture(func() (int, error) {
			return 42, nil
		})
		val, err := fut.Await(ctx)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		_ = val
	}
}

func BenchmarkNewFutureFunc_NoOp(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		getter := NewFutureFunc(func() (interface{}, error) {
			return 42, nil
		})
		val, err := getter()
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		_ = val
	}
}

func BenchmarkFutureCall_NoOp(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		done := make(chan struct{}, 1)
		FutureCall(func() (interface{}, error) {
			return 42, nil
		}, func(res interface{}, err error) {
			if err != nil {
				b.Errorf("unexpected error: %v", err)
			}
			done <- struct{}{}
		})
		<-done
	}
}

// --- 并发吞吐 (各自独立任务, RunParallel) ---

func BenchmarkNewFuture_Parallel(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			fut := NewFuture(func() (int, error) {
				return 1, nil
			})
			val, err := fut.Await(ctx)
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
			_ = val
		}
	})
}

func BenchmarkNewFutureFunc_Parallel(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			getter := NewFutureFunc(func() (interface{}, error) {
				return 1, nil
			})
			val, err := getter()
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
			_ = val
		}
	})
}

func BenchmarkFutureCall_Parallel(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			done := make(chan struct{}, 1)
			FutureCall(func() (interface{}, error) {
				return 1, nil
			}, func(res interface{}, err error) {
				if err != nil {
					b.Errorf("unexpected error: %v", err)
				}
				done <- struct{}{}
			})
			<-done
		}
	})
}

// --- 轻度计算任务 ---

func slowTask() (int, error) {
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += i
	}
	return sum, nil
}

func BenchmarkNewFuture_CPUWork(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		fut := NewFuture(slowTask)
		val, err := fut.Await(ctx)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		_ = val
	}
}

func BenchmarkNewFutureFunc_CPUWork(b *testing.B) {
	slowTaskIface := func() (interface{}, error) {
		return slowTask()
	}
	b.ResetTimer()
	for b.Loop() {
		getter := NewFutureFunc(slowTaskIface)
		val, err := getter()
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		_ = val
	}
}

func BenchmarkFutureCall_CPUWork(b *testing.B) {
	slowTaskIface := func() (interface{}, error) {
		return slowTask()
	}
	b.ResetTimer()
	for b.Loop() {
		done := make(chan struct{}, 1)
		FutureCall(slowTaskIface, func(res interface{}, err error) {
			if err != nil {
				b.Errorf("unexpected error: %v", err)
			}
			done <- struct{}{}
		})
		<-done
	}
}

// --- panic 恢复开销 ---

func BenchmarkNewFuture_Panic(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		fut := NewFuture(func() (int, error) {
			panic("bench boom")
		})
		_, err := fut.Await(ctx)
		if err == nil {
			b.Fatal("expected panic error")
		}
	}
}

func BenchmarkNewFutureFunc_Panic(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		getter := NewFutureFunc(func() (interface{}, error) {
			panic("bench boom")
		})
		_, err := getter()
		if err == nil {
			b.Fatal("expected panic error")
		}
	}
}

func BenchmarkFutureCall_Panic(b *testing.B) {
	b.ResetTimer()
	for b.Loop() {
		done := make(chan struct{}, 1)
		FutureCall(func() (interface{}, error) {
			panic("bench boom")
		}, func(res interface{}, err error) {
			if err == nil {
				b.Error("expected panic error")
			}
			done <- struct{}{}
		})
		<-done
	}
}

// --- 内存分配对比 ---

func BenchmarkNewFuture_Allocs(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		fut := NewFuture(func() (int, error) {
			return 0, nil
		})
		_, _ = fut.Await(ctx)
	}
}

func BenchmarkNewFutureFunc_Allocs(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		getter := NewFutureFunc(func() (interface{}, error) {
			return 0, nil
		})
		_, _ = getter()
	}
}

func BenchmarkFutureCall_Allocs(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		done := make(chan struct{}, 1)
		FutureCall(func() (interface{}, error) {
			return 0, nil
		}, func(res interface{}, err error) {
			done <- struct{}{}
		})
		<-done
	}
}

// --- NewFuture 池预热后对比 ---

func BenchmarkNewFuture_PoolWarmed(b *testing.B) {
	ctx := context.Background()
	// 预热池
	for i := 0; i < 100; i++ {
		fut := NewFuture(func() (int, error) {
			return i, nil
		})
		fut.Await(ctx)
	}
	b.ResetTimer()
	for b.Loop() {
		fut := NewFuture(func() (int, error) {
			return 0, nil
		})
		_, _ = fut.Await(ctx)
	}
}
