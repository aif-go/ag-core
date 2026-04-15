# ag_log

## 概述

ag_log 是对 Go 标准 `slog` 的封装,为 ag-core 提供统一的日志输出 API。通过 slog 作为抽象层，实现对第三方日志框架的插件式封装。

**已支持日志框架**：
- zap（通过 slogzap 封装）
- slog 多模式输出（fanout、failover、pool、route）

## 核心设计理念

### 1. 架构模式

ag_log 采用 **NamedHandler + Builder + Factory** 三层架构：

```
                    ┌─────────────────┐
                    │   slog.Logger   │
                    └────────┬────────┘
                             │
                ┌────────────┼────────────┐
                │            │            │
       ┌────────▼──────┐ ┌──▼─────────┐ ┌─▼──────────────┐
       │ NamedHandler  │ │  Factory   │ │  Builder       │
       └───────────────┘ └────────────┘ └────────────────┘
                │            │            │
        ┌───────┴────────────┴───────┐    │
        │                            │    │
   ┌────▼─────┐              ┌──────▼─────┐
   │ zap handler │              │ fanout handler │
   └───────────┘              └──────────────┘
```

### 2. NamedHandler

为 slog.Handler 添加名称标识，支持通过 Context 传递层级关系（如 `root.handler1.handler2`）。

**作用**：
- 唯一标识每个 handler
- 支持通过 Context 传递 handler 名称
- 便于调试和日志追踪

**示例**：
```go
handler := slog.New(slog.HandlerOptions{
    Level: slog.LevelInfo,
}).WithAttrs([]slog.Attr{
    slog.String("handler_name", "zap1"),
})
```

### 3. Builder 模式

使用构建器模式创建 slog.Logger，支持链式调用。

**核心方法**：
- `AddHandlerFactorys()` - 添加多个 HandlerFactory
- `AddHandlers()` - 直接添加 slog.Handler
- `AddMiddlewares()` - 添加中间件
- `Build()` - 构建最终的 slog.Logger

### 4. HandlerFactory

工厂模式创建 handler，内部递归处理层级关系。使用互斥锁避免递归死锁。

**特点**：
- 支持层级递归构建
- 线程安全（互斥锁）
- 自动解析 handler 名称

### 5. 顶层 Fanout 模式

默认使用 `slogmulti.Fanout` 作为顶层模式，提供多模式日志输出能力：
- `fanout` - 多目标分发
- `failover` - 降级策略
- `pool` - 连接池
- `route` - 路由分发

## 核心组件

### 1. agslog

slog 的核心封装模块。

**文件**：
- `slog_handler.go` - NamedHandler 定义和 HandlerFactory 实现
- `slog_wrap_config.go` - Builder 模式实现
- `slog_common.go` - 类型定义
- `zfx_agslog.go` - FX 依赖注入

**核心类型**：

#### SlogAttrFromContext
从 context 中提取 slog 属性的函数类型。

```go
type SlogAttrFromContext func(ctx context.Context) []slog.Attr
```

#### HandlerInitFunc
Handler 初始化函数类型。

```go
type HandlerInitFunc func(handlers []slog.Handler) error
```

### 2. slogzap

zap 日志框架的封装模块，将 zap.Logger 转换为 slog.Handler。

**文件**：
- `slog_zap.go` - zap 到 slog 的转换实现
- `zfx_slogzap.go` - FX 依赖注入

**核心功能**：
- 将多个 zap.Logger 实例转换为 slog.Handler
- 支持动态配置

### 3. fanout

多模式日志输出封装模块。

**文件**：
- `slog_fanout.go` - Fanout 模式实现
- `zfx_aglog_fanout.go` - FX 依赖注入

**核心功能**：
- 支持多个 fanout 组
- 每个组内组合多个 handler
- 支持级联和链式配置

### 4. logzap

原生 zap.Logger 的配置实现，支持日志滚动、编码格式等。

**文件**：
- `zaplog.go` - zap.Logger 创建和配置

**配置项**：
- `LogFileName` - 日志文件路径
- `LogLevel` - 日志级别（debug/info/warn/error）
- `MaxSize` - 单个日志文件最大大小（MB）
- `MaxBackups` - 保留的备份文件数
- `Compress` - 是否压缩备份文件
- `MaxAge` - 保留日志文件的最大天数
- `Console` - 是否使用控制台编码
- `Prod` - 是否使用生产模式

## 配置说明

### 配置层次结构

```
aglog
├── topHandler           # 顶层 handler
│   ├── zap1             # zap 实例1
│   ├── zap2             # zap 实例2
│   └── f1               # fanout 组1
├── fanout                # fanout 配置
│   └── logs              # fanout 组列表
│       ├── f1
│       │   ├── zap1
│       │   └── zap2
│       └── f2
│           └── f1       # 级联
└── zap                   # zap 配置
    └── logs              # zap 实例列表
        ├── zap1
        └── zap2
```

### 配置示例 1：基础配置

```yaml
aglog:
  topHandler:
    - zap1    # 使用 zap1 作为顶层 handler

aglog.zap.logs:
  zap1:
    log_level: info
    encoding: json
    console: false
```

### 配置示例 2：fanout 多模式

```yaml
aglog:
  topHandler:
    - f1

aglog.fanout.logs:
  f1:
    - zap1
    - zap2

aglog.zap.logs:
  zap1:
    log_level: info
    encoding: json
  zap2:
    log_level: debug
    encoding: json
```

### 配置示例 3：级联和链式

```yaml
aglog:
  topHandler:
    - f1

aglog.fanout.logs:
  f1:
    - f2          # 级联：f1 包含 f2
  f2:
    - zap1
    - zap2

aglog.zap.logs:
  zap1:
    log_level: info
    encoding: json
  zap2:
    log_level: debug
    encoding: console
```

### 配置示例 4：日志文件滚动

```yaml
aglog:
  topHandler:
    - zap1

aglog.zap.logs:
  zap1:
    log_file_name: ./logs/app.log
    log_level: info
    max_size: 100       # 100MB
    max_backups: 10     # 保留10个备份
    max_age: 30         # 保留30天
    compress: true
    prod: true
```

### 配置示例 5：自定义日志服务

```yaml
aglog:
  topHandler:
    - customHandler

aglog.handler.customHandler:
  name: customHandler
  log_level: info
```

需要实现 `ISlogHandler` 接口并注册为 handler。

## 使用方式

### 1. FX 依赖注入（推荐）

```go
package main

import (
    "log/slog"
    "ag-core/ag/ag_conf"
    "ag-core/ag/ag_log"
)

func main() {
    // 启动 FX
    app := fx.New(
        ag_log.FxAgSlogProvide,  // ag_log fx provider
        ag_conf.FxAgConfProvide, // 配置 fx provider
    )
    app.Run()

    // 获取 logger
    logger := agslog.MustGetLogger()
    logger.Info("hello world")
}
```

### 2. 直接构建

```go
package main

import (
    "log/slog"
    "ag-core/ag/ag_log"
)

func main() {
    // 创建 builder
    builder := agslog.NewBuilder()

    // 添加 handler
    builder.AddHandlers(handlers...)

    // 添加中间件
    builder.AddMiddlewares(middlewares...)

    // 构建 logger
    logger := builder.Build()

    // 使用
    logger.Info("hello world")
}
```

### 3. Context 使用

```go
// 传递 handler 名称
ctx := context.WithValue(context.Background(), slog.HandlerKey, slog.HandlerOptions{
    Level: slog.LevelInfo,
}).WithAttrs([]slog.Attr{
    slog.String("handler_name", "zap1"),
})

logger.Info("message", slog.Handler(ctx))
```

### 4. 动态配置

```go
// 获取 builder
builder := agslog.MustGetBuilder()

// 重新加载配置
builder.Reload()

// 获取 logger
logger := agslog.MustGetLogger()
logger.Info("message")
```

## 命名规范

### Handler 命名规则

1. **唯一性**：每个 handler 必须有唯一名称
2. **层级性**：支持 `root.handler1.handler2` 格式
3. **可读性**：使用有意义的名称（如 `zap1`, `zap2`, `f1`, `f2`）

### 命名示例

```
zap1, zap2              # zap 日志实例
f1, f2                  # fanout 组
customHandler            # 自定义 handler
root.zap1                # 顶层 handler
f1.zap1                  # fanout 组内的 zap
```

## 扩展开发

### 添加新的日志框架

1. 在 `slog_<framework>/` 目录下实现 slog.Handler
2. 创建 `zfx_slog_<framework>.go` 文件注册 FX provider
3. 在配置文件中添加框架配置

**示例**：
```go
package sloglogrus

import (
    "go.uber.org/fx"
)

var FxAgSlogLogrusProvide = fx.Provide(
    BindSlogLogrusProperties,
    fx.Annotate(
        NewSlogHandler4LogrusProps,
        fx.ResultTags(`group:"agslog.handlers"`),
    ),
)
```

### 自定义 Handler

实现 `ISlogHandler` 接口：
```go
type ISlogHandler interface {
    NewHandler(props SlogHandlerProperties) slog.Handler
    GetName() string
}
```

## 性能优化建议

1. **使用生产模式**：`prod: true` 减少运行时开销
2. **合理设置日志级别**：避免在生产环境输出过多 debug 日志
3. **使用 fanout 多模式**：支持多目标分发，避免单点性能瓶颈
4. **日志文件滚动**：使用 lumberjack 自动管理日志文件大小

## TODO

- [ ] 异步文件日志（基础的日志实现）
- [ ] 日志服务扩展支持
- [ ] 日志分类（全局日志管理器实现日志区分）
- [ ] 链路追踪信息支持

## 相关文档

- [slog 官方文档](https://pkg.go.dev/log/slog)
- [zap 官方文档](https://pkg.go.dev/go.uber.org/zap)
- [slogmulti 文档](https://pkg.go.dev/github.com/sigutils/slogmulti)
