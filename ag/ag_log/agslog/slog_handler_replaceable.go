package agslog

import (
	"context"
	"log/slog"
	"sync/atomic"
)

var (
	_ slog.Handler  = (*ReplaceableHandler)(nil)
	_ INamedHandler = (*ReplaceableHandler)(nil)
)

type ReplaceableHandler struct {
	name     string
	handler  atomic.Pointer[slog.Handler]
	replaced atomic.Bool
}

func NewReplaceableHandler(name string, handler slog.Handler) *ReplaceableHandler {
	h := &ReplaceableHandler{
		name: name,
	}
	h.handler.Store(&handler)
	return h
}

func (rh *ReplaceableHandler) Name() string {
	return rh.name
}

// Original 获取原始handler
func (rh *ReplaceableHandler) Original() slog.Handler {
	hl := rh.handler.Load()
	if hl == nil {
		return nil
	}
	return *hl
}

// ReplaceHandler 替换handler
// 注意：替换后，旧的handler会被释放 TODO 原WithAttrs/WithGroup会被丢失
func (rh *ReplaceableHandler) ReplaceHandler(handler slog.Handler) {
	if rh.replaced.Load() {
		return
	}
	rh.handler.Store(&handler)
	rh.replaced.Store(true)
}

func (rh *ReplaceableHandler) IsReplaced() bool {
	return rh.replaced.Load()
}

/* 实现slog.Handler接口 */
func (rh *ReplaceableHandler) Enabled(ctx context.Context, level slog.Level) bool {
	hl := rh.handler.Load()
	if hl == nil {
		return false
	}
	return (*hl).Enabled(ctx, level)
}

func (rh *ReplaceableHandler) Handle(ctx context.Context, r slog.Record) error {
	hl := rh.handler.Load()
	if hl == nil {
		return nil
	}
	return (*hl).Handle(ctx, r)
}

func (rh *ReplaceableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hload := rh.handler.Load()
	if hload == nil {
		return slog.DiscardHandler
	}
	// 新handler包装了WithAttrs后的handler
	current := *hload
	ah := current.WithAttrs(attrs)
	return ah
}

func (rh *ReplaceableHandler) WithGroup(name string) slog.Handler {
	hload := rh.handler.Load()
	if hload == nil {
		return slog.DiscardHandler
	}
	current := *hload
	gh := current.WithGroup(name)
	return gh
}
