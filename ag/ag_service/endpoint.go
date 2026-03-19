package ag_service

import (
	"context"
)

// 定义中间件优先级常量
const (
	ServiceInfoMiddlewarePriorityHighest = 0
	ServiceInfoMiddlewarePriorityHigh    = 1000
	ServiceInfoMiddlewarePriorityNormal  = 2000
	ServiceInfoMiddlewarePriorityLow     = 3000
	ServiceInfoMiddlewarePriorityLowest  = 4000
)

type Endpoint func(ctx context.Context, req interface{}) (interface{}, error)

type MiddlewareFunc func(next Endpoint) Endpoint

// PrioritizedMiddlewareProvider 服务中间件提供者接口
type PrioritizedMiddlewareProvider interface {
	// GetOrder 优先级，数值越小优先级越高
	GetOrder() int

	// Middleware 获取实际的中间件函数
	Middleware() MiddlewareFunc
}

type MiddleWareCondition interface {
	// Condition 判断是否满足条件
	Condition(callInfo *CallInfo) bool
}

// middlewareProviderSorter 实现 sort.Interface 用于排序
type middlewareProviderSorter []PrioritizedMiddlewareProvider

func (p middlewareProviderSorter) Len() int           { return len(p) }
func (p middlewareProviderSorter) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p middlewareProviderSorter) Less(i, j int) bool { return p[i].GetOrder() < p[j].GetOrder() }

// MiddlewareProvider 服务中间件提供者接口
type MiddlewareProvider interface {
	// Middleware 获取中间件函数
	Middleware() MiddlewareFunc
}

type SimpleMiddleware struct {
	Mw MiddlewareFunc
}

func (p *SimpleMiddleware) Middleware() MiddlewareFunc {
	return p.Mw
}

type SimplePrioritizedMiddleware struct {
	Order int
	Mw    MiddlewareFunc
}

func (p *SimplePrioritizedMiddleware) GetOrder() int {
	return p.Order
}

func (p *SimplePrioritizedMiddleware) Middleware() MiddlewareFunc {
	return p.Mw
}
