package async

import (
	"context"
	"log/slog"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

type benchMockHandler struct{}

func (m benchMockHandler) Enabled(ctx context.Context, level slog.Level) bool { return true }
func (m benchMockHandler) Handle(ctx context.Context, r slog.Record) error     { return nil }
func (m benchMockHandler) WithAttrs(attrs []slog.Attr) slog.Handler            { return m }
func (m benchMockHandler) WithGroup(name string) slog.Handler                  { return m }

func setupBenchAsyncHandler() slog.Handler {
	config := &AsyncGroupConfig{
		Worker:          4,
		Queue:           100000,
		FullStrategy:    "drop_new",
		ShutdownTimeout: time.Second,
	}
	return NewAsyncHandler(benchMockHandler{}, "bench", config)
}

func newBenchRecord(msg string, attrs ...slog.Attr) slog.Record {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	r := slog.NewRecord(time.Now(), slog.LevelInfo, msg, pcs[0])
	r.AddAttrs(attrs...)
	return r
}

// BenchmarkHandle 测量单次 Handle 调用的内存分配
// 预期: allocs/op >= 1 (logTask 堆分配)
func BenchmarkHandle(b *testing.B) {
	handler := setupBenchAsyncHandler()
	ctx := context.Background()
	r := newBenchRecord("benchmark message")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(ctx, r)
	}
}

// BenchmarkHandleWithAttrs 带 2 个属性的 Handle
func BenchmarkHandleWithAttrs(b *testing.B) {
	handler := setupBenchAsyncHandler()
	ctx := context.Background()
	r := newBenchRecord("benchmark with attrs",
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(ctx, r)
	}
}

// BenchmarkHandleManyAttrs 带 10 个属性的 Handle
func BenchmarkHandleManyAttrs(b *testing.B) {
	handler := setupBenchAsyncHandler()
	ctx := context.Background()
	r := newBenchRecord("benchmark with many attrs",
		slog.String("k1", "v1"),
		slog.Int("k2", 2),
		slog.Bool("k3", true),
		slog.Float64("k4", 4.0),
		slog.String("k5", "v5"),
		slog.Int("k6", 6),
		slog.String("k7", "v7"),
		slog.Float64("k8", 8.0),
		slog.Bool("k9", false),
		slog.String("k10", "v10"),
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(ctx, r)
	}
}

// BenchmarkHandleFullLifecycle 测量提交+消费的完整生命周期分配
// 每轮等待 worker 处理完成后才继续，模拟实际生产场景
func BenchmarkHandleFullLifecycle(b *testing.B) {
	ctx := context.Background()
	r := newBenchRecord("full lifecycle message", slog.String("key", "value"))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler := setupBenchAsyncHandler()
		handler.Handle(ctx, r)
		// 等待 worker 处理 (mock handler 是同步的，给一点时间)
		// 实际场景中 GC 压力来自大量未处理的 logTask 堆积
	}
}

// BenchmarkHandleHighConcurrency 高并发下的分配放大效应
func BenchmarkHandleHighConcurrency(b *testing.B) {
	handler := setupBenchAsyncHandler()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := newBenchRecord("concurrent benchmark", slog.String("g", "test"))
		for pb.Next() {
			handler.Handle(ctx, r)
		}
	})
}

// BenchmarkRecordClone 单独测量 Record.Clone() 的分配开销
// 用于评估引入 Clone 修复并发安全问题的代价
func BenchmarkRecordClone(b *testing.B) {
	r := newBenchRecord("clone test", slog.String("key", "value"))
	r2 := newBenchRecord("clone test with attrs",
		slog.String("k1", "v1"),
		slog.Int("k2", 2),
		slog.Bool("k3", true),
		slog.Float64("k4", 4.0),
		slog.String("k5", "v5"),
	)

	b.Run("noattrs", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.Clone()
		}
	})

	b.Run("manyattrs", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r2.Clone()
		}
	})
}

// BenchmarkHandleThroughLogger 通过 slog.Logger 经 AsyncHandler 写入的完整路径
// 最接近实际使用场景
func BenchmarkHandleThroughLogger(b *testing.B) {
	handler := setupBenchAsyncHandler()
	logger := slog.New(handler)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark through logger")
	}
}

// BenchmarkHandleThroughLoggerWithArgs 带参数的完整 Logger 路径
func BenchmarkHandleThroughLoggerWithArgs(b *testing.B) {
	handler := setupBenchAsyncHandler()
	logger := slog.New(handler)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark with args", "key", "value", "count", 42)
	}
}

// countingHandler 统计处理次数，用于验证异步处理的正确性
type countingHandler struct {
	processed atomic.Int64
	benched   atomic.Int64
}

func (c *countingHandler) Enabled(ctx context.Context, level slog.Level) bool { return true }
func (c *countingHandler) Handle(ctx context.Context, r slog.Record) error {
	c.processed.Add(1)
	return nil
}
func (c *countingHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return c }
func (c *countingHandler) WithGroup(name string) slog.Handler       { return c }

// BenchmarkHandleThroughput 吞吐量测试：测量 worker 处理速度和队列深度
// 验证在高频日志场景下，worker 是否能跟上提交速度
func BenchmarkHandleThroughput(b *testing.B) {
	ch := &countingHandler{}
	config := &AsyncGroupConfig{
		Worker:          4,
		Queue:           100000,
		FullStrategy:    "drop_new",
		ShutdownTimeout: time.Second,
	}
	handler := NewAsyncHandler(ch, "throughput", config)
	asyncH := handler.(*AsyncHandler)

	ctx := context.Background()
	r := newBenchRecord("throughput test")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(ctx, r)
	}

	// 等待 worker 处理完剩余任务
	b.StopTimer()
	for asyncH.workerGroup.GetStats().Queued.Load() != asyncH.workerGroup.GetStats().Processed.Load() {
		time.Sleep(time.Millisecond)
	}
	b.StartTimer()

	b.ReportMetric(float64(ch.processed.Load()), "processed")
}
