package async

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

// recordingHandler 通过 channel 传递收到的 ctx，避免 data race
type recordingHandler struct {
	ctxCh chan context.Context
}

func (r *recordingHandler) Enabled(ctx context.Context, level slog.Level) bool { return true }
func (r *recordingHandler) Handle(ctx context.Context, rec slog.Record) error {
	select {
	case r.ctxCh <- ctx:
	default:
	}
	return nil
}
func (r *recordingHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return r }
func (r *recordingHandler) WithGroup(name string) slog.Handler       { return r }

// noopHandler 简单 handler，Handle 返回 nil
type noopHandler struct{}

func (n noopHandler) Enabled(ctx context.Context, level slog.Level) bool { return true }
func (n noopHandler) Handle(ctx context.Context, rec slog.Record) error  { return nil }
func (n noopHandler) WithAttrs(attrs []slog.Attr) slog.Handler           { return n }
func (n noopHandler) WithGroup(name string) slog.Handler                 { return n }

// submitOnlyGroup 构造只含 Submit 所需字段的 WorkerGroup（不启动 worker）
func submitOnlyGroup(strategy string, queue int) *WorkerGroup {
	return &WorkerGroup{
		config:   &AsyncGroupConfig{Queue: queue, FullStrategy: strategy},
		logQueue: make(chan *logTask, queue),
		stats:    &WorkerStats{},
	}
}

func newRecord(msg string) slog.Record {
	return slog.NewRecord(time.Now(), slog.LevelInfo, msg, 0)
}

func TestTaskPoolReturnResets(t *testing.T) {
	task := taskPool.Borrow()
	task.ctx = context.Background()
	task.handler = noopHandler{}
	task.record = newRecord("test")

	taskPool.Return(task)

	reused := taskPool.Borrow()
	if reused.ctx != nil {
		t.Fatalf("expected ctx nil after Reset, got %v", reused.ctx)
	}
	if reused.handler != nil {
		t.Fatalf("expected handler nil after Reset, got %v", reused.handler)
	}
	if !reused.record.Time.IsZero() {
		t.Fatalf("expected zero record after Reset, got %v", reused.record)
	}
}

func TestTaskPoolReturnNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Return(nil) panicked: %v", r)
		}
	}()
	taskPool.Return(nil)
}

func TestSubmitDropNewDropsAndReturns(t *testing.T) {
	wg := submitOnlyGroup("drop_new", 1)

	filler := taskPool.Borrow()
	wg.logQueue <- filler

	task := taskPool.Borrow()
	if err := wg.Submit(task); err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	if got := wg.stats.Dropped.Load(); got != 1 {
		t.Fatalf("expected Dropped=1, got %d", got)
	}
	if got := wg.stats.Queued.Load(); got != 0 {
		t.Fatalf("expected Queued=0, got %d", got)
	}
	if reused := taskPool.Borrow(); reused != task {
		t.Fatalf("expected dropped task returned to pool")
	}
}

func TestSubmitDropOldReturnsOldQueuesNew(t *testing.T) {
	wg := submitOnlyGroup("drop_old", 1)

	old := taskPool.Borrow()
	wg.logQueue <- old

	new := taskPool.Borrow()
	if err := wg.Submit(new); err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	if got := wg.stats.Dropped.Load(); got != 1 {
		t.Fatalf("expected Dropped=1, got %d", got)
	}
	if got := wg.stats.Queued.Load(); got != 1 {
		t.Fatalf("expected Queued=1, got %d", got)
	}
	if reused := taskPool.Borrow(); reused != old {
		t.Fatalf("expected old task returned to pool")
	}
	select {
	case got := <-wg.logQueue:
		if got != new {
			t.Fatalf("expected new task enqueued")
		}
	default:
		t.Fatalf("expected new task in queue")
	}
}

func TestSubmitUnknownStrategy(t *testing.T) {
	wg := submitOnlyGroup("bogus", 1)

	task := taskPool.Borrow()
	if err := wg.Submit(task); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	taskPool.Return(task)
}

func TestSubmitCounterReconciliation(t *testing.T) {
	wg := NewWorkerGroup(&AsyncGroupConfig{
		Worker:          2,
		Queue:           8,
		FullStrategy:    "drop_new",
		ShutdownTimeout: time.Second,
	})
	defer wg.Stop()

	handler := &AsyncHandler{workerGroup: wg, original: noopHandler{}}

	const total = 1000
	ctx := context.Background()
	for i := 0; i < total; i++ {
		if err := handler.Handle(ctx, newRecord("reconcile")); err != nil {
			t.Fatalf("Handle error: %v", err)
		}
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		stats := wg.GetStats()
		if stats.Queued.Load()+stats.Dropped.Load() == total &&
			stats.Processed.Load() == stats.Queued.Load() {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout: queued=%d dropped=%d processed=%d",
				stats.Queued.Load(), stats.Dropped.Load(), stats.Processed.Load())
		}
		time.Sleep(time.Millisecond)
	}

	stats := wg.GetStats()
	if stats.Queued.Load()+stats.Dropped.Load() != total {
		t.Fatalf("Queued+Dropped != total: %d + %d != %d",
			stats.Queued.Load(), stats.Dropped.Load(), total)
	}
	if stats.Processed.Load() != stats.Queued.Load() {
		t.Fatalf("Processed != Queued: %d != %d",
			stats.Processed.Load(), stats.Queued.Load())
	}
}

func TestHandleCutsCancellation(t *testing.T) {
	wg := NewWorkerGroup(&AsyncGroupConfig{
		Worker:          1,
		Queue:           16,
		FullStrategy:    "block_wait",
		ShutdownTimeout: time.Second,
	})
	defer wg.Stop()

	rec := &recordingHandler{ctxCh: make(chan context.Context, 1)}
	handler := &AsyncHandler{workerGroup: wg, original: rec}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := handler.Handle(ctx, newRecord("cut cancellation")); err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	select {
	case got := <-rec.ctxCh:
		if got.Done() != nil {
			t.Fatalf("expected Done()==nil (cancellation cut), got non-nil")
		}
		if got.Err() != nil {
			t.Fatalf("expected Err()==nil, got %v", got.Err())
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for handler")
	}
}

func TestHandlePreservesCtxValues(t *testing.T) {
	wg := NewWorkerGroup(&AsyncGroupConfig{
		Worker:          1,
		Queue:           16,
		FullStrategy:    "block_wait",
		ShutdownTimeout: time.Second,
	})
	defer wg.Stop()

	rec := &recordingHandler{ctxCh: make(chan context.Context, 1)}
	handler := &AsyncHandler{workerGroup: wg, original: rec}

	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "trace-123")

	if err := handler.Handle(ctx, newRecord("preserve values")); err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	select {
	case got := <-rec.ctxCh:
		if v := got.Value(ctxKey{}); v != "trace-123" {
			t.Fatalf("expected ctx value preserved, got %v", v)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for handler")
	}
}
