package agslog

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

var (
	_ slog.Handler  = (*NamedHandler)(nil)
	_ INamedHandler = (*NamedHandler)(nil)
)

type INamedHandler interface {
	slog.Handler

	Name() string
	// 获取原始handler
	Original() slog.Handler
}

// NamedHandler 命名handler
// 为handler添加一个名称，方便在多handler场景下，根据名称获取对应的handler
type NamedHandler struct {
	name    string
	Handler slog.Handler
}

// NewNamedHandler 创建一个命名handler
func NewNamedHandler(name string, h slog.Handler) slog.Handler {
	return &NamedHandler{
		Handler: h,
		name:    name,
	}
}

// Name 获取handler名称
func (n *NamedHandler) Name() string {
	return n.name
}

// Original 获取原始handler
func (n *NamedHandler) Original() slog.Handler {
	return n.Handler
}

// Enabled 判断是否启用
func (n *NamedHandler) Enabled(ctx context.Context, l slog.Level) bool {
	startName := ctx.Value(HandlerStartCtxKey{})
	if startName == nil {
		startName = n.name
		ctx = context.WithValue(ctx, HandlerStartCtxKey{}, startName)
		// r.AddAttrs(slog.Attr{Key: HandlerStartKey, Value: slog.StringValue(startName.(string))})
	}
	return n.Handler.Enabled(ctx, l)
}

// Handle 处理日志
func (n *NamedHandler) Handle(ctx context.Context, r slog.Record) error {
	// handlerName := ctx.Value(HandlerNameCtxKey{})
	// if handlerName == nil {
	// 	handlerName = n.name
	// } else {
	// 	handlerName = fmt.Sprintf("%s.%s", handlerName, n.name)
	// }

	// ctx = context.WithValue(ctx, HandlerNameCtxKey{}, handlerName)
	// r.AddAttrs(slog.Attr{Key: HandlerNameKey, Value: slog.StringValue(handlerName.(string))})

	startName := ctx.Value(HandlerStartCtxKey{})
	if startName == nil {
		startName = n.name
		ctx = context.WithValue(ctx, HandlerStartCtxKey{}, startName)
		// r.AddAttrs(slog.Attr{Key: HandlerStartKey, Value: slog.StringValue(startName.(string))})
	}
	// ctx = context.WithValue(ctx, HandlerEndCtxKey{}, n.name)
	// r.AddAttrs(slog.Attr{Key: HandlerEndKey, Value: slog.StringValue(n.name)})

	return n.Handler.Handle(ctx, r)
}

// WithAttrs 添加属性
func (n *NamedHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &NamedHandler{
		name:    n.name,
		Handler: n.Handler.WithAttrs(attrs),
	}
}

// WithGroup 添加分组
func (n *NamedHandler) WithGroup(name string) slog.Handler {
	return &NamedHandler{
		name:    n.name,
		Handler: n.Handler.WithGroup(name),
	}
}

type HandlerFactory struct {
	Name     string
	instance slog.Handler
	// DoGetHandler func([]INamedHandler, []HandlerFactory) (slog.Handler, error)
	DoGetHandler func(func(handlerName string) (slog.Handler, error)) (slog.Handler, error)

	mu sync.Mutex
}

func NewHandlerFactory(name string, doGetHandler func(func(handlerName string) (slog.Handler, error)) (slog.Handler, error)) *HandlerFactory {
	return &HandlerFactory{
		Name:         name,
		DoGetHandler: doGetHandler,
	}
}

// GetHandler 获取handler，调用子实现DoGetHandler
func (f *HandlerFactory) GetHandler(resolveHandler func(handlerName string) (slog.Handler, error)) (slog.Handler, error) {

	ok := f.mu.TryLock() // FIXME 此锁是控制递归循环调用，避免循环调用导致死循环
	if !ok {
		// 此处极有可能存在循环调用handler的情况
		return nil, fmt.Errorf("handler factory %s get handler failed, lock failed, maybe circular call", f.Name)
	}
	defer f.mu.Unlock()

	if f.instance != nil {
		return f.instance, nil
	}

	if f.DoGetHandler == nil {
		return nil, fmt.Errorf("handler factory [%s] do build handler is nil", f.Name)
	}

	handler, err := f.DoGetHandler(resolveHandler)
	if err != nil {
		return nil, fmt.Errorf("handler factory [%s] get handler failed, err:\n >>> %w", f.Name, err)
	}

	if handler == nil {
		return nil, fmt.Errorf("handler factory [%s] get handler failed, handler is nil", f.Name)
	}

	f.instance = handler

	return handler, nil
}
