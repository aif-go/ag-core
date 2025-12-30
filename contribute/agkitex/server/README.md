# AgKitex 服务器设计文档

## 概述

AgKitex 服务器模块是一个基于 CloudWeGo Kitex 框架的服务器封装，提供了配置驱动、依赖注入、服务注册等高级特性，使 Kitex 服务器开发更加简单和标准化。

## 设计理念

1. **配置驱动**：基于 ag-core 配置系统，支持属性绑定和环境变量
2. **依赖注入**：使用 fx 框架实现松耦合的组件管理
3. **模块化设计**：清晰的模块划分，便于扩展和维护
4. **原生兼容**：完全兼容 Kitex 原生接口和特性

## 核心组件

### 1. 配置管理 (`config.go`)

#### 配置结构
```go
type KitexServerProperties struct {
    Host          string `value:"${:}"`           // 服务器主机地址
    Port          int    `value:"${:7000}"`       // 服务器端口
    AdaptivePort  bool   `value:"${:false}"`      // 是否启用自适应端口
    ServiceName   string                          // 服务名称
    EnableIPRange string `value:"${:}"`           // IP 范围限制

    Grpc Grpc                                     // gRPC 配置
}

type Grpc struct {
    Enable            bool `value:"${:true}"`     // 是否启用 gRPC
    MaxConnectionIdle int  `value:"${:0}"`        // 最大空闲连接时间
}
```

#### 配置示例
```yaml
kitex:
  server:
    host: "0.0.0.0"
    port: 8080
    adaptivePort: true
    serviceName: "user-service"
    grpc:
      enable: true
      maxConnectionIdle: 300
```

### 2. 中间件系统 (`middleware.go`)

#### 优先级中间件
```go
// 优先级常量
const (
    ServerMiddlewarePriorityHighest = 0      // 最高优先级
    ServerMiddlewarePriorityHigh    = 1000   // 高优先级
    ServerMiddlewarePriorityNormal  = 2000   // 普通优先级
    ServerMiddlewarePriorityLow     = 3000   // 低优先级
    ServerMiddlewarePriorityLowest  = 4000   // 最低优先级
)

// 中间件接口
type PrioritizedServerMiddleware interface {
    GetOrder() int                    // 获取优先级
    GetMiddleware() endpoint.Middleware // 获取中间件函数
}
```

#### 使用示例
```go
// 创建优先级中间件
authMiddleware := NewSimplePrioritizedServerMiddleware(
    ServerMiddlewarePriorityHighest,
    func(next endpoint.Endpoint) endpoint.Endpoint {
        return func(ctx context.Context, req, resp interface{}) (err error) {
            // 认证逻辑
            return next(ctx, req, resp)
        }
    },
)

// 构建中间件选项
options := BuildServerMiddlewareOptions([]PrioritizedServerMiddleware{
    authMiddleware,
})
```

### 3. 服务器套件 (`server.go`)

#### 核心结构
```go
type KitexServerSuite struct {
    opts []server.Option
}

type KitexServerSuiteBuilder struct {
    ServerOptions          []server.Option
    Properties             *KitexServerProperties
    Registry               registry.Registry
    Middlewares            []endpoint.Middleware
    PrioritizedMiddlewares []PrioritizedServerMiddleware
}

type AgKitexServer struct {
    KitexServer server.Server
}
```

#### 构建流程
1. **配置解析**：从配置属性构建服务器选项
2. **中间件构建**：按优先级排序中间件
3. **服务注册**：配置服务注册中心
4. **套件组装**：将所有选项组装成 Suite

### 4. 服务注册 (`service_registry.go`)

#### 服务注册器
```go
type AgKitexServiceRegistry struct {
    ServiceInfo *serviceinfo.ServiceInfo
    Handler     interface{}
    Opts        []server.RegisterOption
}

type AgKitexServiceRegistryHolder struct {
    Registries []*AgKitexServiceRegistry
}
```

#### 注册示例
```go
// 创建服务注册器
registry := NewAgKitexServiceRegistry(
    serviceInfo,
    handler,
)

// 注册服务
registry.Register(server)
```

### 5. 依赖注入 (`zfx_kitex_server.go`)

#### fx 模块
```go
var FxKitexServerBaseModule = fx.Module("fx_kitex_server_base",
    // 注册中心模块
    agkitexReg.FxKitexRegistyModule,
    
    // 提供者
    fx.Provide(
        NewKitexServerProperties,           // 配置
        FxNewKitexServerSuiteBuilder,       // 套件构建器
        FxBuilderKitexServerSuite,          // 套件
        NewKitexServerWithSuite,            // Kitex 服务器
        NewAgKitexServer,                   // AgKitex 服务器
        FxNewKitexServiceRegistryHolder,    // 服务注册器持有者
    ),
    
    // 调用者
    fx.Invoke(
        FxInvokerKitexServiceRegistryHolder, // 服务注册
    ),
)
```

#### 注入参数
```go
type FxInKitexServerParams struct {
    fx.In
    ServerOptions          []server.Option                    `group:"kitex_server_options",optional:"true"`
    Config                 *KitexServerProperties
    Registry               registry.Registry
    Middlewares            []endpoint.Middleware              `group:"kitex_server_middlewares",optional:"true"`
    PrioritizedMiddlewares []PrioritizedServerMiddleware      `group:"kitex_server_prioritized_middlewares",optional:"true"`
}
```

## 使用方式

### 1. 基础使用
```go
// 创建配置
config := &KitexServerProperties{
    Host:        "0.0.0.0",
    Port:        8080,
    ServiceName: "my-service",
}

// 构建服务器套件
builder := &KitexServerSuiteBuilder{
    Properties: config,
}
suite, _ := builder.BuildServerSuite()

// 创建服务器
server, _ := NewKitexServerWithSuite(suite)
agServer := NewAgKitexServer(server)

// 启动服务器
agServer.Start(context.Background())
```

### 2. fx 集成使用
```go
// 创建应用
app := fx.New(
    FxKitexServerBaseModule,
    // 添加自定义中间件
    fx.Provide(
        NewFxServerMiddlewareProvider(func() endpoint.Middleware {
            return func(next endpoint.Endpoint) endpoint.Endpoint {
                return func(ctx context.Context, req, resp interface{}) error {
                    // 自定义中间件逻辑
                    return next(ctx, req, resp)
                }
            }
        }),
    ),
    // 添加服务注册器
    fx.Provide(
        NewFxAgKitexServiceRegistry(func() *AgKitexServiceRegistry {
            return NewAgKitexServiceRegistry(serviceInfo, handler)
        }),
    ),
)

// 启动应用
app.Run()
```

### 3. 配置中间件
```go
// 普通中间件
middleware := func(next endpoint.Endpoint) endpoint.Endpoint {
    return func(ctx context.Context, req, resp interface{}) error {
        // 中间件逻辑
        return next(ctx, req, resp)
    }
}

// 优先级中间件
prioritizedMiddleware := NewSimplePrioritizedServerMiddleware(
    ServerMiddlewarePriorityHigh,
    middleware,
)
```

## 高级特性

### 1. 自适应端口
当 `AdaptivePort` 为 `true` 时，系统会自动寻找可用端口：
```yaml
kitex:
  server:
    adaptivePort: true
    port: 7000  # 起始端口
```

### 2. IP 范围限制
支持 IP 范围限制，自动选择范围内的 IP 地址，注册服务时的IP：
```yaml
kitex:
  server:
    enableIPRange: "192.168.1.100:192.168.1.200,127.1:127.2"
```

### 3. gRPC 配置
完整的 gRPC 服务器配置支持：
```yaml
kitex:
  server:
    grpc:
      enable: true
      maxConnectionIdle: 300  # 秒
```

## 扩展指南

### 1. 添加新的中间件类型
```go
// 实现 PrioritizedServerMiddleware 接口
type CustomMiddleware struct {
    Order int
    Func  endpoint.Middleware
}

func (c *CustomMiddleware) GetOrder() int {
    return c.Order
}

func (c *CustomMiddleware) GetMiddleware() endpoint.Middleware {
    return c.Func
}
```

### 2. 添加新的注册中心
在 `registry` 目录下创建新的注册中心实现，并在 `zfx_registy.go` 中注册。

### 3. 自定义配置属性
扩展 `KitexServerProperties` 结构体，添加新的配置字段。

## 最佳实践

1. **配置管理**：使用 ag-core 配置系统管理所有配置
2. **中间件顺序**：合理设置中间件优先级，确保执行顺序正确
3. **服务注册**：使用服务注册器统一管理所有服务
4. **依赖注入**：充分利用 fx 框架的依赖注入能力
5. **错误处理**：在关键位置添加适当的错误处理和日志记录

## 总结

AgKitex 服务器模块提供了一个强大而灵活的服务器框架，通过配置驱动、依赖注入和模块化设计，大大简化了 Kitex 服务器的开发流程，同时保持了与原生 Kitex 的完全兼容性。
