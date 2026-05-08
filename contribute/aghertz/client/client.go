package client

import (
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
)

// HertzClientParams is the parameters for creating a hertz client.
type HertzClientParams struct {
	ClientProperties *HertzClientProperties

	ClientOptions    []*config.ClientOption
	ClientMiddleware []client.Middleware

	// PrioritizedClientMiddleware 带优先级的客户端中间件，用于fx注入时保证顺序
	PrioritizedClientMiddleware      []PrioritizedClientMiddleware
	PrioritizedClientMiddlewareSuite PrioritizedClientMiddlewareSuite

	Resolver discovery.Resolver
}

// NewHertzClient creates a new hertz client.
func NewHertzClient(params *HertzClientParams) (*client.Client, error) {
	// 应用客户端选项
	suite := &SimpleClientSuite{
		opts: params.ClientOptions,
	}

	c, err := client.NewClient(WithClientSuite(suite))
	if err != nil {
		slog.Error("failed to create hertz client", "err", err)
		return nil, err
	}

	// 应用客户端中间件
	var cliMws []PrioritizedClientMiddleware

	// 服务发现中间件（如果有）
	if params.Resolver != nil {
		cliMws = append(cliMws, &SimplePrioritizedClientMiddleware{
			Order:      ClientMiddlewarePriorityNormal,
			Middleware: sd.Discovery(params.Resolver),
		})
	}

	// 优先使用带优先级的中间件（保证顺序）
	if len(params.PrioritizedClientMiddleware) > 0 {
		cliMws = append(cliMws, params.PrioritizedClientMiddleware...)
		// SortAndApplyMiddleware(c, params.PrioritizedClientMiddleware)
	}

	// 普通中间件（不保证顺序），默认配置在Normal级别
	if len(params.ClientMiddleware) > 0 {
		// 兼容旧版本：使用普通中间件（不保证顺序）
		// c.Use(params.ClientMiddleware...)
		for _, mw := range params.ClientMiddleware {
			cliMws = append(cliMws, &SimplePrioritizedClientMiddleware{
				Order:      ClientMiddlewarePriorityNormal,
				Middleware: mw,
			})
		}
	}

	if params.PrioritizedClientMiddlewareSuite != nil {
		cliMws = append(cliMws, params.PrioritizedClientMiddlewareSuite.GetMiddlewares()...)
	}

	if len(cliMws) > 0 {
		SortAndApplyMiddleware(c, cliMws)
	}

	return c, nil
}
