package ag_ext

import (
	"context"
	"log/slog"
	"sort"
	"time"
)

type HandlerFunc func(ctx context.Context, req interface{}) (interface{}, error)

type PrioritizedMiddleware interface {
	// GetOrder 优先级，数值越小优先级越高
	GetOrder() int

	GetMiddleware() Middleware
}

// ByPriority 实现 sort.Interface 用于排序
type ByPriority []PrioritizedMiddleware

func (p ByPriority) Len() int           { return len(p) }
func (p ByPriority) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ByPriority) Less(i, j int) bool { return p[i].GetOrder() < p[j].GetOrder() }

// 定义中间件优先级常量
const (
	MiddlewarePriorityHighest = 0
	MiddlewarePriorityHigh    = 10
	MiddlewarePriorityNormal  = 20
	MiddlewarePriorityLow     = 30
	MiddlewarePriorityLowest  = 40
)

// Middleware 中间件类型
type Middleware func(
	method string,
	ctx context.Context,
	req interface{},
	next func(context.Context, interface{}) (interface{}, error),
) (interface{}, error)

// RegisterHandler 注册单个方法的处理链
func RegisterHandler(methodName string, prioritizedMws []PrioritizedMiddleware, handler HandlerFunc) HandlerFunc {
	// 1. 按优先级排序（优先级值小的先执行）
	sort.Sort(ByPriority(prioritizedMws))

	// 2. 提取中间件函数
	middlewares := make([]Middleware, 0, len(prioritizedMws))
	for _, pmw := range prioritizedMws {
		middlewares = append(middlewares, pmw.GetMiddleware())
	}

	// 3.从后向前包装中间件
	wrappedHandler := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		mw := middlewares[i]
		next := wrappedHandler
		wrappedHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
			return mw(methodName, ctx, req, next)
		}
	}

	return wrappedHandler
}

// LoggingMiddleware 示例：日志中间件
type LoggingMiddleware struct{}

func (l LoggingMiddleware) GetOrder() int {
	return MiddlewarePriorityHighest
}

func (l LoggingMiddleware) GetMiddleware() Middleware {
	return loggingMiddleware
}

func loggingMiddleware(
	method string,
	ctx context.Context,
	req interface{},
	next func(context.Context, interface{}) (interface{}, error),
) (interface{}, error) {
	start := time.Now()
	slog.Info("request received.", "method", method)

	res, err := next(ctx, req)

	if err != nil {
		slog.Error("failed to handle request!", "method", method, "error", err, "duration", time.Since(start))
	} else {
		slog.Info("request handled completed!", "method", method, "duration", time.Since(start))
	}
	return res, err
}
