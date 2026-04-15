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
	name string
	// handler slog.Handler
	// handler atomic.Value
	handler atomic.Pointer[slog.Handler]
	// handler slog.Handler
	// mu      sync.RWMutex
}

func NewReplaceableHandler(name string, handler slog.Handler) *ReplaceableHandler {
	h := &ReplaceableHandler{
		name: name,
	}
	// h.handler.Store(handler)
	h.handler.Store(&handler)
	return h
}

func (rh *ReplaceableHandler) Name() string {
	return rh.name
}

// Original 获取原始handler
func (rh *ReplaceableHandler) Original() slog.Handler {
	// return rh.handler.Load().(slog.Handler)
	return *rh.handler.Load()
}

// ReplaceHandler 替换handler
func (rh *ReplaceableHandler) ReplaceHandler(handler slog.Handler) {
	// rh.handler.Store(handler)
	rh.handler.Store(&handler)
	// rh.handler.Swap(handler)
}

// IsMatchesName 是否handler是否符合name
func (rh *ReplaceableHandler) IsMatchesName() bool {
	// if handler, ok := rh.handler.Load().(INamedHandler); ok {
	if handler, ok := (*rh.handler.Load()).(INamedHandler); ok {
		return handler.Name() == rh.name
	}
	return false
}

/* 实现slog.Handler接口 */
func (rh *ReplaceableHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// return rh.handler.Load().(slog.Handler).Enabled(ctx, level)
	return (*rh.handler.Load()).Enabled(ctx, level)
}

func (rh *ReplaceableHandler) Handle(ctx context.Context, r slog.Record) error {
	// return rh.handler.Load().(slog.Handler).Handle(ctx, r)
	return (*rh.handler.Load()).Handle(ctx, r)
}

func (rh *ReplaceableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := &ReplaceableHandler{
		name: rh.name,
	}
	// 新handler包装了WithAttrs后的handler
	// current := rh.handler.Load().(slog.Handler)
	current := (*rh.handler.Load()).(slog.Handler)
	// newHandler.handler.Store(current.WithAttrs(attrs))
	ah := current.WithAttrs(attrs)
	newHandler.handler.Store(&ah)
	return newHandler
	// return rh.handler.Load().(slog.Handler).WithAttrs(attrs)
}

func (rh *ReplaceableHandler) WithGroup(name string) slog.Handler {
	newHandler := &ReplaceableHandler{
		name: rh.name,
	}
	// current := rh.handler.Load().(slog.Handler)
	current := (*rh.handler.Load()).(slog.Handler)
	// newHandler.handler.Store(current.WithGroup(name))
	gh := current.WithGroup(name)
	newHandler.handler.Store(&gh)
	return newHandler
	// return rh.handler.Load().(slog.Handler).WithGroup(name)
}
