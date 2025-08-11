package agslog

import (
	"context"
	"log/slog"
)

// SlogAttrFromContext 从context中获取slog属性
type SlogAttrFromContext func(ctx context.Context) []slog.Attr

// HandlerInitFunc handler初始化函数
type HandlerInitFunc func(handlers []slog.Handler) error

// HandlerDefiniton handler注册描述对象
// type HandlerDefinition struct {
// 	Handler slog.Handler    // 处理器
// 	Init    HandlerInitFunc // 处理器初始化函数
// }
