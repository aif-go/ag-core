package server

import (
	"sort"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/server"
)

// 定义服务器中间件优先级常量
const (
	ServerMiddlewarePriorityHighest = 0
	ServerMiddlewarePriorityHigh    = 1000
	ServerMiddlewarePriorityNormal  = 2000
	ServerMiddlewarePriorityLow     = 3000
	ServerMiddlewarePriorityLowest  = 4000
)

// PrioritizedServerMiddleware 带优先级的服务器中间件接口
type PrioritizedServerMiddleware interface {
	// GetOrder 优先级，数值越小优先级越高
	GetOrder() int

	// GetMiddleware 获取实际的中间件函数
	GetMiddleware() endpoint.Middleware
}

// SimplePrioritizedServerMiddleware 简单的优先级服务器中间件实现
type SimplePrioritizedServerMiddleware struct {
	Order      int
	Middleware endpoint.Middleware
}

// PrioritizedServerMiddlewareSuite 带优先级的服务器中间件套件接口
type PrioritizedServerMiddlewareSuite interface {
	// GetMiddlewares 获取所有中间件
	GetMiddlewares() []PrioritizedServerMiddleware
}

// SimplePrioritizedServerMiddlewareSuite 简单的优先级服务器中间件套件实现
type SimplePrioritizedServerMiddlewareSuite struct {
	Middlewares []PrioritizedServerMiddleware
}

// ByServerPriority 实现 sort.Interface 用于排序
type ByServerPriority []PrioritizedServerMiddleware

func (p ByServerPriority) Len() int           { return len(p) }
func (p ByServerPriority) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ByServerPriority) Less(i, j int) bool { return p[i].GetOrder() < p[j].GetOrder() }

// NewSimplePrioritizedServerMiddleware 创建一个简单的优先级服务器中间件实例
func NewSimplePrioritizedServerMiddleware(order int, mw endpoint.Middleware) *SimplePrioritizedServerMiddleware {
	return &SimplePrioritizedServerMiddleware{
		Order:      order,
		Middleware: mw,
	}
}

func (s *SimplePrioritizedServerMiddleware) GetOrder() int {
	return s.Order
}

func (s *SimplePrioritizedServerMiddleware) GetMiddleware() endpoint.Middleware {
	return s.Middleware
}

func (s *SimplePrioritizedServerMiddleware) SetOrder(order int) {
	s.Order = order
}

func (s *SimplePrioritizedServerMiddleware) SetMiddleware(mw endpoint.Middleware) {
	s.Middleware = mw
}

func (s *SimplePrioritizedServerMiddlewareSuite) GetMiddlewares() []PrioritizedServerMiddleware {
	return s.Middlewares
}

func (s *SimplePrioritizedServerMiddlewareSuite) AddPrioritizedMiddleware(mw PrioritizedServerMiddleware) {
	s.Middlewares = append(s.Middlewares, mw)
}

func (s *SimplePrioritizedServerMiddlewareSuite) AddMiddleware(mw endpoint.Middleware) {
	s.AddPrioritizedMiddleware(NewSimplePrioritizedServerMiddleware(ServerMiddlewarePriorityNormal, mw))
}

// BuildServerMiddlewareOptions 构建中间件选项
func BuildServerMiddlewareOptions(prioritizedMws []PrioritizedServerMiddleware) []server.Option {
	if len(prioritizedMws) == 0 {
		return nil
	}

	// 1. 按优先级排序（优先级值小的先执行）
	sort.Sort(ByServerPriority(prioritizedMws))

	// 2. 提取中间件函数并按顺序构建选项
	options := make([]server.Option, 0, len(prioritizedMws))
	for _, pmw := range prioritizedMws {
		options = append(options, server.WithMiddleware(pmw.GetMiddleware()))
	}

	return options
}
