package client

import (
	"sort"

	"github.com/cloudwego/hertz/pkg/app/client"
)

// 定义客户端中间件优先级常量
const (
	ClientMiddlewarePriorityHighest = 0
	ClientMiddlewarePriorityHigh    = 1000
	ClientMiddlewarePriorityNormal  = 2000
	ClientMiddlewarePriorityLow     = 3000
	ClientMiddlewarePriorityLowest  = 4000
)

type (
	// ByClientPriority 实现 sort.Interface 用于排序
	ByClientPriority []PrioritizedClientMiddleware

	// PrioritizedClientMiddleware 带优先级的客户端中间件接口
	PrioritizedClientMiddleware interface {
		// GetOrder 优先级，数值越小优先级越高
		GetOrder() int

		// GetMiddleware 获取实际的中间件函数
		GetMiddleware() client.Middleware
	}

	// SimplePrioritizedClientMiddleware 简单的优先级客户端中间件实现
	SimplePrioritizedClientMiddleware struct {
		Order      int
		Middleware client.Middleware
	}

	// PrioritizedClientMiddlewareSuite 带优先级的客户端中间件套件接口
	PrioritizedClientMiddlewareSuite interface {
		// GetMiddleware 获取所有中间件
		GetMiddlewares() []PrioritizedClientMiddleware
	}

	// SimplePrioritizedClientMiddlewareSuite 简单的优先级客户端中间件套件实现
	SimplePrioritizedClientMiddlewareSuite struct {
		Middlewares []PrioritizedClientMiddleware
	}
)

func (p ByClientPriority) Len() int           { return len(p) }
func (p ByClientPriority) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ByClientPriority) Less(i, j int) bool { return p[i].GetOrder() < p[j].GetOrder() }

// NewSimplePrioritizedClientMiddleware 创建一个简单的优先级客户端中间件实例
func NewSimplePrioritizedClientMiddleware(order int, mw client.Middleware) *SimplePrioritizedClientMiddleware {
	return &SimplePrioritizedClientMiddleware{
		Order:      order,
		Middleware: mw,
	}
}

func (s *SimplePrioritizedClientMiddleware) GetOrder() int {
	return s.Order
}
func (s *SimplePrioritizedClientMiddleware) GetMiddleware() client.Middleware {
	return s.Middleware
}
func (s *SimplePrioritizedClientMiddleware) SetOrder(order int) {
	s.Order = order
}
func (s *SimplePrioritizedClientMiddleware) SetMiddleware(mw client.Middleware) {
	s.Middleware = mw
}
func (s *SimplePrioritizedClientMiddlewareSuite) GetMiddlewares() []PrioritizedClientMiddleware {
	return s.Middlewares
}
func (s *SimplePrioritizedClientMiddlewareSuite) AddPrioritizedMiddleware(mw PrioritizedClientMiddleware) {
	s.Middlewares = append(s.Middlewares, mw)
}
func (s *SimplePrioritizedClientMiddlewareSuite) AddMiddleware(mw client.Middleware) {
	s.AddPrioritizedMiddleware(NewSimplePrioritizedClientMiddleware(ClientMiddlewarePriorityNormal, mw))
}

// SortAndApplyMiddleware 对中间件进行排序并应用到客户端
func SortAndApplyMiddleware(c *client.Client, prioritizedMws []PrioritizedClientMiddleware) {
	if len(prioritizedMws) == 0 {
		return
	}

	// 1. 按优先级排序（优先级值小的先执行）
	sort.Sort(ByClientPriority(prioritizedMws))

	// 2. 提取中间件函数并按顺序应用
	for _, pmw := range prioritizedMws {
		c.Use(pmw.GetMiddleware())
	}
}
