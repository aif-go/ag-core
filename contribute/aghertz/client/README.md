# Hertz 客户端中间件顺序保证方案

## 问题背景

在使用 fx 依赖注入框架时，通过 `group` 标签注入的中间件切片无法保证顺序，这会导致中间件执行顺序不确定，影响业务逻辑。

## 解决方案

我们设计了一个带优先级的中间件系统，通过 `PrioritizedClientMiddleware` 接口和排序机制来保证中间件执行顺序。

### 核心组件

1. **PrioritizedClientMiddleware 接口**
   - `GetOrder() int` - 返回优先级（数值越小优先级越高）
   - `GetMiddleware() client.Middleware` - 返回实际的中间件函数

2. **排序机制**
   - 使用 `sort.Sort(ByClientPriority)` 对中间件按优先级排序
   - 优先级常量定义：
     - `ClientMiddlewarePriorityHighest` = 0
     - `ClientMiddlewarePriorityHigh` = 1000
     - `ClientMiddlewarePriorityNormal` = 2000
     - `ClientMiddlewarePriorityLow` = 3000
     - `ClientMiddlewarePriorityLowest` = 4000

3. **Fx 集成**
   - 通过 `group:"aghertz_client_middleware"` 标签注入带优先级的中间件
   - 使用 `NewFxClientMiddlewareProvider` 包装中间件提供者

### 使用方法

#### 1. 定义带优先级的中间件

```go
type AuthClientMiddleware struct{}

func (a AuthClientMiddleware) GetOrder() int {
    return ClientMiddlewarePriorityHigh  // 优先级10
}

func (a AuthClientMiddleware) GetMiddleware() client.Middleware {
    return func(next client.Endpoint) client.Endpoint {
        return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
            // 中间件逻辑
            req.Header.Set("Authorization", "Bearer token")
            return next(ctx, req, resp)
        }
    }
}
```

#### 2. 使用 Fx 模块注入

```go
var FxClientMiddlewareModule = fx.Module("client_middleware",
    fx.Provide(
        NewFxClientMiddlewareProvider(func() PrioritizedClientMiddleware {
            return &AuthClientMiddleware{}
        }),
        // 其他中间件...
    ),
)
```

#### 3. 在应用中使用

```go
type AppParams struct {
    fx.In
    ClientOptions    []*config.ClientOption
    PrioritizedClientMiddleware []PrioritizedClientMiddleware `group:"aghertz_client_middleware"`
}

func NewHertzClientWithOrderedMiddleware(params AppParams) (*client.Client, error) {
    clientParams := &HertzClientParams{
        ClientOptions: params.ClientOptions,
        PrioritizedClientMiddleware: params.PrioritizedClientMiddleware,
    }
    return NewHertzClient(clientParams)
}
```

### 执行顺序保证

中间件将按照优先级从高到低（数值从小到大）的顺序执行：

1. 优先级 0 (Highest) - 最先执行
2. 优先级 10 (High)
3. 优先级 20 (Normal) 
4. 优先级 30 (Low)
5. 优先级 40 (Lowest) - 最后执行

### 向后兼容

- 原有的 `ClientMiddleware []client.Middleware` 字段仍然可用
- 如果同时提供了 `PrioritizedClientMiddleware` 和 `ClientMiddleware`，所有中间件都会合并并按优先级排序
- 普通中间件会被自动赋予 `ClientMiddlewarePriorityNormal` (20) 的默认优先级
- 执行顺序完全可控，所有中间件都参与排序

### 混合使用示例

```go
// 带优先级的中间件
type AuthClientMiddleware struct{}  // 优先级10
type LoggingClientMiddleware struct{} // 优先级0

// 普通中间件（自动赋予Normal优先级20）
var normalMiddleware = func(next client.Endpoint) client.Endpoint {
    return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
        // 普通中间件逻辑
        return next(ctx, req, resp)
    }
}

// 创建客户端
params := &HertzClientParams{
    PrioritizedClientMiddleware: []PrioritizedClientMiddleware{
        &AuthClientMiddleware{},    // 优先级10
        &LoggingClientMiddleware{}, // 优先级0
    },
    ClientMiddleware: []client.Middleware{
        normalMiddleware, // 自动赋予优先级20
    },
}
```

执行顺序：Logging (0) → Auth (10) → normalMiddleware (20)
