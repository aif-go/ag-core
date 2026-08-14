package failover

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
)

// failHandler 总是返回 error，用于模拟首选 handler 失败
type failHandler struct{}

func (failHandler) Enabled(ctx context.Context, level slog.Level) bool { return true }
func (failHandler) Handle(ctx context.Context, r slog.Record) error    { return errors.New("fail") }
func (failHandler) WithAttrs(attrs []slog.Attr) slog.Handler           { return failHandler{} }
func (failHandler) WithGroup(name string) slog.Handler                 { return failHandler{} }

// countingHandler 记录 Handle 调用次数并返回 nil
type countingHandler struct {
	count atomic.Int64
}

func (c *countingHandler) Enabled(ctx context.Context, level slog.Level) bool { return true }
func (c *countingHandler) Handle(ctx context.Context, r slog.Record) error {
	c.count.Add(1)
	return nil
}
func (c *countingHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return c }
func (c *countingHandler) WithGroup(name string) slog.Handler       { return c }

// mockGetHandler 根据 name 返回预置 handler
func mockGetHandler(m map[string]slog.Handler) func(string) (slog.Handler, error) {
	return func(name string) (slog.Handler, error) {
		h, ok := m[name]
		if !ok {
			return nil, errors.New("handler not found: " + name)
		}
		return h, nil
	}
}

func TestFailoverFallsBackOnError(t *testing.T) {
	ok := &countingHandler{}
	get := mockGetHandler(map[string]slog.Handler{"fail": failHandler{}, "ok": ok})

	f := getDoGetHandlerFunc([]string{"fail", "ok"})
	h, err := f(get)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := h.Handle(context.Background(), slog.Record{}); err != nil {
		t.Fatalf("expected nil error after fallback, got %v", err)
	}
	if ok.count.Load() != 1 {
		t.Fatalf("expected fallback handler called once, got %d", ok.count.Load())
	}
}

func TestFailoverStopsOnFirstSuccess(t *testing.T) {
	first := &countingHandler{}
	second := &countingHandler{}
	get := mockGetHandler(map[string]slog.Handler{"first": first, "second": second})

	f := getDoGetHandlerFunc([]string{"first", "second"})
	h, err := f(get)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := h.Handle(context.Background(), slog.Record{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first.count.Load() != 1 {
		t.Fatalf("expected first handler called once, got %d", first.count.Load())
	}
	if second.count.Load() != 0 {
		t.Fatalf("expected second handler not called, got %d", second.count.Load())
	}
}

func TestFailoverAllFailReturnsError(t *testing.T) {
	get := mockGetHandler(map[string]slog.Handler{"f1": failHandler{}, "f2": failHandler{}})

	f := getDoGetHandlerFunc([]string{"f1", "f2"})
	h, err := f(get)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := h.Handle(context.Background(), slog.Record{}); err == nil {
		t.Fatalf("expected error when all handlers fail")
	}
}

func TestNewFailoverHandlerFactorysNilSafe(t *testing.T) {
	factories, err := NewFailoverHandlerFactorys(nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(factories) != 0 {
		t.Fatalf("expected empty factories for nil props, got %d", len(factories))
	}

	factories, err = NewFailoverHandlerFactorys(&AgSlogFailoverProperties{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(factories) != 0 {
		t.Fatalf("expected empty factories for nil Logs, got %d", len(factories))
	}
}
