package ag_log

import (
	"context"
	"fmt"
	"log/slog"
)

type HzwHandler struct {
}

// Enabled 检查日志级别
func (h *HzwHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

// Handle 处理日志记录
func (h *HzwHandler) Handle(ctx context.Context, r slog.Record) error {
	fmt.Printf("HzwHandler Handle\n")
	r.Attrs(func(attr slog.Attr) bool {
		fmt.Printf("k:%s, v:%s\n", attr.Key, attr.Value.String())
		return true
	})
	return nil
}

// WithAttrs 返回一个新的带有额外属性的 Handler
func (h *HzwHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fmt.Println("WithAttrs")
	for _, attr := range attrs {
		fmt.Printf("k:%s, v:%s\n", attr.Key, attr.Value.String())
	}
	return h
}

// WithGroup 返回一个新的带有分组的 Handler
func (h *HzwHandler) WithGroup(name string) slog.Handler {
	return h
}
