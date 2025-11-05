package client

import (
	"sort"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
)

// 定义客户端中间件优先级常量
const (
	ClientMiddlewarePriorityHighest = 0
	ClientMiddlewarePriorityHigh    = 1000
	ClientMiddlewarePriorityNormal  = 2000
	ClientMiddlewarePriorityLow     = 3000
	ClientMiddlewarePriorityLowest  = 4000
)

// PrioritizedClientMiddleware 带优先级的客户端中间件接口
type PrioritizedClientMiddleware interface {
	// GetOrder 优先级，数值越小优先级越高
	GetOrder() int

	// GetMiddleware 获取实际的中间件函数
	GetMiddleware() endpoint.Middleware
}

// SimplePrioritizedClientMiddleware 简单的优先级客户端中间件实现
type SimplePrioritizedClientMiddleware struct {
	Order      int
	Middleware endpoint.Middleware
}

// PrioritizedClientMiddlewareSuite 带优先级的客户端中间件套件接口
type PrioritizedClientMiddlewareSuite interface {
	// GetMiddlewares 获取所有中间件
	GetMiddlewares() []PrioritizedClientMiddleware
}

// SimplePrioritizedClientMiddlewareSuite 简单的优先级客户端中间件套件实现
type SimplePrioritizedClientMiddlewareSuite struct {
	Middlewares []PrioritizedClientMiddleware
}

// ByClientPriority 实现 sort.Interface 用于排序
type ByClientPriority []PrioritizedClientMiddleware

func (p ByClientPriority) Len() int           { return len(p) }
func (p ByClientPriority) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ByClientPriority) Less(i, j int) bool { return p[i].GetOrder() < p[j].GetOrder() }

// NewSimplePrioritizedClientMiddleware 创建一个简单的优先级客户端中间件实例
func NewSimplePrioritizedClientMiddleware(order int, mw endpoint.Middleware) *SimplePrioritizedClientMiddleware {
	return &SimplePrioritizedClientMiddleware{
		Order:      order,
		Middleware: mw,
	}
}

func (s *SimplePrioritizedClientMiddleware) GetOrder() int {
	return s.Order
}

func (s *SimplePrioritizedClientMiddleware) GetMiddleware() endpoint.Middleware {
	return s.Middleware
}

func (s *SimplePrioritizedClientMiddleware) SetOrder(order int) {
	s.Order = order
}

func (s *SimplePrioritizedClientMiddleware) SetMiddleware(mw endpoint.Middleware) {
	s.Middleware = mw
}

func (s *SimplePrioritizedClientMiddlewareSuite) GetMiddlewares() []PrioritizedClientMiddleware {
	return s.Middlewares
}

func (s *SimplePrioritizedClientMiddlewareSuite) AddPrioritizedMiddleware(mw PrioritizedClientMiddleware) {
	s.Middlewares = append(s.Middlewares, mw)
}

func (s *SimplePrioritizedClientMiddlewareSuite) AddMiddleware(mw endpoint.Middleware) {
	s.AddPrioritizedMiddleware(NewSimplePrioritizedClientMiddleware(ClientMiddlewarePriorityNormal, mw))
}

// BuildMiddlewareOptions 构建中间件选项
func BuildMiddlewareOptions(prioritizedMws []PrioritizedClientMiddleware) []client.Option {
	if len(prioritizedMws) == 0 {
		return nil
	}

	// 1. 按优先级排序（优先级值小的先执行）
	sort.Sort(ByClientPriority(prioritizedMws))

	// 2. 提取中间件函数并按顺序构建选项
	options := make([]client.Option, 0, len(prioritizedMws))
	for _, pmw := range prioritizedMws {
		options = append(options, client.WithMiddleware(pmw.GetMiddleware()))
	}

	return options
}
