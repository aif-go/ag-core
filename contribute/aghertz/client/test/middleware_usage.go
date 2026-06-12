package test

import (
	"context"
	"log/slog"
	"time"

	aghertzclient "github.com/aif-go/ag-core/contribute/aghertz/client"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"go.uber.org/fx"
)

// 示例：认证中间件
type AuthClientMiddleware struct{}

func (a AuthClientMiddleware) GetOrder() int {
	return aghertzclient.ClientMiddlewarePriorityHigh
}

func (a AuthClientMiddleware) GetMiddleware() client.Middleware {
	return func(next client.Endpoint) client.Endpoint {
		return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
			// 添加认证头
			req.Header.Set("Authorization", "Bearer example-token")
			slog.Debug("auth middleware applied")
			return next(ctx, req, resp)
		}
	}
}

// 示例：重试中间件
type RetryClientMiddleware struct{}

func (r RetryClientMiddleware) GetOrder() int {
	return aghertzclient.ClientMiddlewarePriorityNormal
}

func (r RetryClientMiddleware) GetMiddleware() client.Middleware {
	return func(next client.Endpoint) client.Endpoint {
		return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
			var err error
			// 重试逻辑
			for i := 0; i < 3; i++ {
				err = next(ctx, req, resp)
				if err == nil {
					break
				}
				slog.Warn("request failed, retrying", "attempt", i+1, "error", err)
			}
			return err
		}
	}
}

// 示例：监控中间件
type MetricsClientMiddleware struct{}

func (m MetricsClientMiddleware) GetOrder() int {
	return aghertzclient.ClientMiddlewarePriorityLow
}

func (m MetricsClientMiddleware) GetMiddleware() client.Middleware {
	return func(next client.Endpoint) client.Endpoint {
		return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
			slog.Debug("metrics middleware: request started")
			err := next(ctx, req, resp)
			slog.Debug("metrics middleware: request completed")
			return err
		}
	}
}

// ExamplePrioritizedClientMiddleware 示例：日志中间件
type ExampleLoggingClientMiddleware struct{}

func (l ExampleLoggingClientMiddleware) GetOrder() int {
	return aghertzclient.ClientMiddlewarePriorityHighest
}

func (l ExampleLoggingClientMiddleware) GetMiddleware() client.Middleware {
	return func(next client.Endpoint) client.Endpoint {
		return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
			// 记录请求开始时间
			start := time.Now()
			slog.Info("client request started", "url", string(req.URI().FullURI()))

			// 调用下一个中间件或处理函数
			err := next(ctx, req, resp)

			// 记录请求完成
			if err != nil {
				slog.Error("client request failed", "url", string(req.URI().FullURI()), "error", err, "duration", time.Since(start))
			} else {
				slog.Info("client request completed", "url", string(req.URI().FullURI()), "duration", time.Since(start))
			}

			return err
		}
	}
}

// 使用fx模块的示例
var FxClientMiddlewareModule = fx.Module("client_middleware",
	fx.Provide(
		// 提供带优先级的中间件
		aghertzclient.NewFxClientMiddlewareProvider(func() aghertzclient.PrioritizedClientMiddleware {
			return &AuthClientMiddleware{}
		}),
		aghertzclient.NewFxClientMiddlewareProvider(func() aghertzclient.PrioritizedClientMiddleware {
			return &RetryClientMiddleware{}
		}),
		aghertzclient.NewFxClientMiddlewareProvider(func() aghertzclient.PrioritizedClientMiddleware {
			return &MetricsClientMiddleware{}
		}),
	),
)

// 创建带中间件的客户端示例
func CreateClientWithOrderedMiddleware() (*client.Client, error) {
	params := &aghertzclient.HertzClientParams{
		PrioritizedClientMiddleware: []aghertzclient.PrioritizedClientMiddleware{
			&AuthClientMiddleware{},    // 优先级1000
			&RetryClientMiddleware{},   // 优先级2000
			&MetricsClientMiddleware{}, // 优先级3000
		},
	}

	return aghertzclient.NewHertzClient(params)
}
