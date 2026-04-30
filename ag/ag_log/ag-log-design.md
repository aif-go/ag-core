# ag_log 系统架构设计文档

**文档版本**: 1.0  
**创建日期**: 2026-04-16  
**作者**: ag-core team  
**状态**: 已发布

---

## 1. 文档概述

### 1.1 文档目的

本文档全面描述 ag_log 模块的系统架构、设计理念、核心组件、配置系统和使用方式。旨在为开发者提供深入理解 ag_log 系统的技术文档，同时为系统维护和扩展提供参考依据。

### 1.2 文档范围

本文档涵盖：
- ag_log 的整体架构设计
- 核心组件的技术实现细节
- 配置系统和使用方式
- 扩展机制和最佳实践
- 性能优化建议和已知问题

### 1.3 目标读者

- **系统架构师**：了解整体架构设计理念和技术选型
- **后端开发者**：学习和使用 ag_log 进行日志集成
- **技术维护者**：理解系统实现细节，进行问题诊断和优化
- **新成员**：快速了解 ag_log 系统的设计和使用

### 1.4 术语和缩写

| 术语 | 全称 | 说明 |
|------|------|------|
| slog | Structured Log | Go 标准库的结构化日志 API |
| FX | Fx Framework | Go 的依赖注入框架 |
| Handler | Log Handler | 日志处理器，负责日志的输出和处理 |
| Builder | Builder Pattern | 构建器模式，用于复杂对象的构建 |
| Factory | Factory Pattern | 工厂模式，用于对象的创建 |
| NamedHandler | Named Handler | 带有名称标识的日志处理器 |
| Fanout | Fanout Pattern | 扇出模式，将日志分发到多个目标 |

---

## 2. 系统概述

### 2.1 ag_log 的定位和作用

ag_log 是 ag-core 项目中的日志模块，是对 Go 标准 `slog` 的封装，为整个项目提供统一的日志输出 API。通过 slog 作为抽象层，实现对第三方日志框架的插件式封装。

**核心作用**：
- 提供统一的日志接口，屏蔽底层日志框架差异
- 支持多种日志框架的灵活切换和组合
- 实现灵活的配置管理和动态配置
- 支持多种日志输出模式（fanout、failover、pool、route）
- 提供完善的日志追踪和调试能力

### 2.2 核心设计理念

#### 2.2.1 抽象层优先

以 Go 标准库 `slog` 作为抽象层，所有日志框架都实现为 slog.Handler，确保接口统一。

#### 2.2.2 插件式架构

支持通过插件方式集成第三方日志框架，无需修改核心代码即可扩展新的日志框架支持。

#### 2.2.3 配置驱动

所有日志配置通过配置文件管理，支持动态配置和热重载。

#### 2.2.4 命名和层级

每个 Handler 都有唯一名称，支持层级命名（如 `root.handler1.handler2`），便于管理和调试。

#### 2.2.5 工厂和构建器模式

使用工厂模式创建 Handler，使用构建器模式构建 Logger，提供灵活的配置方式。

### 2.3 与 Go slog 和第三方日志框架的关系

```
┌─────────────────────────────────────────┐
│         Application Code               │
└────────────────┬──────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│         slog.Logger                    │
│         (Go 标准库抽象层)              │
└────────────────┬──────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│         ag_log (封装层)                │
│         - NamedHandler                 │
│         - Builder                      │
│         - Factory                      │
└────────────────┬──────────────────────┘
                 │
    ┌────────────┼────────────┐
    ▼            ▼            ▼
┌─────────┐  ┌─────────┐  ┌──────────┐
│  zap    │  │logrus   │  │ custom   │
│(已实现) │  │(待扩展) │  │ (自定义) │
└─────────┘  └─────────┘  └──────────┘
```

---

## 3. 架构设计

### 3.1 三层架构

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

#### 3.1.1 NamedHandler 层

**职责**：
- 为每个 Handler 添加唯一名称标识
- 支持 Context 传递层级关系
- 提供统一的接口和生命周期管理

**核心接口**：
```go
type INamedHandler interface {
    slog.Handler
    Name() string
    Original() slog.Handler
}
```

#### 3.1.2 Factory 层

**职责**：
- 负责创建和管理 Handler 实例
- 支持递归解析和层级构建
- 提供线程安全的 Handler 获取

**核心类型**：
```go
type HandlerFactory struct {
    Name         string
    instance     INamedHandler
    DoGetHandler func(func(handlerName string) (slog.Handler, error)) (slog.Handler, error)
    mu           sync.Mutex
}
```

#### 3.1.3 Builder 层

**职责**：
- 构建最终的 slog.Logger 实例
- 管理 Handler 的注册和缓存
- 支持中间件的添加和管道处理

**核心方法**：
- `AddHandlerFactorys()` - 添加多个 HandlerFactory
- `AddHandlers()` - 直接添加 slog.Handler
- `AddMiddlewares()` - 添加中间件
- `Build()` - 构建最终的 slog.Logger

### 3.2 组件交互图

```
┌─────────────────────────────────────────────────────────────┐
│                       Builder                              │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐      │
│  │ Handler     │  │ Middleware   │  │ Properties   │      │
│  │ Factories   │  │     Pipe     │  │              │      │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘      │
└─────────┼──────────────────┼───────────┼────────────────┘
          │                  │           │
          │                  │           │
          ▼                  │           │
┌───────────────────────────┼───────────┼────────────────┐
│        resolveHandler()     │           │                │
│  ┌─────────────────────────┼───────────┴───────┐      │
│  │                           resolveTopHandlers()│      │
│  │  ┌────────────────────────┼──────────────────┼───┐  │
│  │  │      HandlerFactory.GetHandler()           │   │  │
│  │  │  ┌────────────────────────┼──────────────┼──┼──┼┐ │
│  │  │  │   DoGetHandler(resolveHandler)         │  │  ││ │
│  │  │  │         ▼                              │  │  ││ │
│  │  │  │  ┌─────────────┐                        │  │  ││ │
│  │  │  │  │NamedHandler │                        │  │  ││ │
│  │  │  │  └──────┬──────┘                        │  │  ││ │
│  │  │  │         │                               │  │  ││ │
│  │  │  │         ▼                               │  │  ││ │
│  │  │  │  ┌─────────────┐                        │  │  ││ │
│  │  │  │  │slog.Handler │                        │  │  ││ │
│  │  │  │  └──────┬──────┘                        │  │  ││ │
│  │  │  │         │                               │  │  ││ │
│  │  │  └─────────┼───────────────────────────────┘  │  ││ │
│  │  └─────────────┼──────────────────────────────────┘  ││ │
│  └─────────────────┼─────────────────────────────────────┼│ │
└────────────────────┼─────────────────────────────────────┼┼┘
                   │                                     ││
                   ▼                                     ││
            ┌─────────────┐                               ││
            │ slog.Logger │◄──────────────────────────────┼┘
            └─────────────┘                               │
                          Build()                         │
```

### 3.3 设计模式应用

| 设计模式 | 应用位置 | 作用 |
|----------|----------|------|
| Builder Pattern | Builder | 提供灵活的 Logger 构建方式 |
| Factory Pattern | HandlerFactory | 负责创建和管理 Handler 实例 |
| Named Pattern | NamedHandler | 为 Handler 添加名称标识 |
| Fanout Pattern | slogmulti.Fanout | 将日志分发到多个目标 |
| Pipe Pattern | slogmulti.Pipe | 中间件管道处理 |
| Strategy Pattern | slog.Handler | 不同日志框架实现相同接口 |
| Singleton Pattern | TopLogger | 全局唯一的顶层 Logger |

---

## 4. 核心组件详解

### 4.1 agslog（核心封装层）

**位置**: `ag/ag_log/agslog/`

**核心文件**：
- `slog_handler.go` - NamedHandler 定义和 HandlerFactory 实现
- `slog_wrap_config.go` - Builder 模式实现
- `slog_common.go` - 类型定义
- `slog_context.go` - Context 上下文处理
- `slog_handler_replaceable.go` - 可替换的 Handler
- `zfx_agslog.go` - FX 依赖注入

#### 4.1.1 NamedHandler

**功能**：为 slog.Handler 添加名称标识，支持通过 Context 传递层级关系。

**核心方法**：
- `Name()` - 获取 Handler 名称
- `Original()` - 获取原始 Handler
- `Handle()` - 处理日志记录，在 Context 中传递层级信息

**Context 传递机制**：
```
HandlerNameCtxKey: root.handler1.handler2
HandlerStartCtxKey: root
HandlerEndCtxKey: handler2
```

#### 4.1.2 HandlerFactory

**功能**：工厂模式创建 Handler，内部递归处理层级关系，使用互斥锁避免递归死锁。

**关键特性**：
- 支持层级递归构建
- 线程安全（互斥锁）
- 自动解析 Handler 名称
- 循环调用检测

**递归解析流程**：
1. 检查是否已缓存实例
2. 尝试获取锁（避免循环调用）
3. 调用 DoGetHandler 函数
4. 解析 Handler 名称
5. 封装为 NamedHandler
6. 缓存实例

#### 4.1.3 Builder

**功能**：构建最终的 slog.Logger，管理 Handler 的注册和缓存。

**核心数据结构**：
```go
type Builder struct {
    props               *AgSlogProperties
    custTopHandlers     []slog.Handler
    handlersCaches      sync.Map      // 缓存每个原子 handler
    handlers            sync.Map      // 缓存每个层级 handler
    namedLogger         sync.Map      // 缓存命名 logger
    replaceableHandlers sync.Map      // 可替换的 handler
    factories           []*HandlerFactory
    middlewares         []slogmulti.Middleware
    logMu               sync.Mutex
}
```

**构建流程**：
1. 解析顶层 Handler（resolveTopHandlers）
2. 通过 Fanout 组合多个 Handler
3. 应用中间件管道
4. 替换全局默认 Logger
5. 缓存并返回结果

### 4.2 slogzap（Zap 封装层）

**位置**: `ag/ag_log/slogzap/`

**核心文件**：
- `slog_zap.go` - zap 到 slog 的转换实现
- `zfx_slogzap.go` - FX 依赖注入

**功能**：将 zap.Logger 转换为 slog.Handler，支持动态配置。

**配置结构**：
```go
type SlogZapProperties struct {
    Logs map[string]logzap.ZlogProperties
}
```

**创建流程**：
1. 遍历配置中的所有 zap 实例
2. 为每个实例创建 zap.Logger
3. 使用 slogzap.Option 转换为 slog.Handler
4. 封装为 NamedHandler

### 4.3 fanout（多模式输出层）

**位置**: `ag/ag_log/fanout/`

**核心文件**：
- `slog_fanout.go` - Fanout 模式实现
- `zfx_aglog_fanout.go` - FX 依赖注入

**功能**：支持多个 fanout 组，每个组内可组合多个 handler，支持级联和链式配置。

**配置结构**：
```go
type AgSlogFanoutProperties struct {
    Logs map[string][]string  // fanout 组名 -> 子 handler 名称列表
}
```

**支持的输出模式**：
- `fanout` - 多目标分发
- `failover` - 降级策略
- `pool` - 连接池
- `route` - 路由分发

### 4.4 logzap（原生 Zap 配置层）

**位置**: `ag/ag_log/logzap/`

**核心文件**：
- `zaplog.go` - zap.Logger 创建和配置

**功能**：原生 zap.Logger 的配置实现，支持日志滚动、编码格式等。

**配置项**：
```go
type ZlogProperties struct {
    LogFileName string   // 日志文件路径
    LogLevel    string   // 日志级别
    MaxSize     int      // 单个日志文件最大大小（MB）
    MaxBackups  int      // 保留的备份文件数
    Compress    bool     // 是否压缩备份文件
    MaxAge      int      // 保留日志文件的最大天数
    Console     bool     // 是否使用控制台编码
    Prod        bool     // 是否使用生产模式
    Stdout      bool     // 是否打印到控制台
}
```

**日志滚动**：使用 lumberjack 库实现自动日志文件管理。

### 4.5 依赖注入（FX 集成）

**FX Provider 层次**：

```go
// 1. 配置绑定
ag_conf.FxAgConfProvide

// 2. ag_log 核心组件
agslog.FxAgSlogProvide

// 3. zap 日志框架
slogzap.FxSlogZapProvide

// 4. fanout 多模式输出
fanout.FxAglogFanoutProvide
```

**依赖关系**：
```
ag_conf (配置中心)
    ↓
slogzap (zap 封装)
    ↓
fanout (fanout 封装)
    ↓
agslog (核心构建)
    ↓
slog.Logger (最终产物)
```

---

## 5. 配置系统

### 5.1 配置层次结构

```
aglog
├── topHandler           # 顶层 handler 配置
│   ├── zap1
│   ├── zap2
│   └── f1
├── fanout                # fanout 配置
│   └── logs
│       ├── f1
│       │   ├── zap1
│       │   └── zap2
│       └── f2
│           └── f1       # 级联
└── zap                   # zap 配置
    └── logs
        ├── zap1
        └── zap2
```

### 5.2 配置示例

#### 5.2.1 基础配置

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

#### 5.2.2 fanout 多模式

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

#### 5.2.3 级联和链式

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

#### 5.2.4 日志文件滚动

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

#### 5.2.5 自定义 Handler

```yaml
aglog:
  topHandler:
    - customHandler

aglog.handler.customHandler:
  name: customHandler
  log_level: info
```

### 5.3 配置验证和错误处理

**错误处理策略**：
1. **配置加载失败不阻断应用启动** - slogzap 和 fanout 的配置加载失败时返回 nil 而非错误
2. **Handler 解析失败跳过** - resolveHandler 失败时打印错误但继续处理其他 Handler
3. **配置默认值** - 提供合理的配置默认值

**验证机制**：
- Handler 名称唯一性检查
- 日志级别有效性验证
- 文件路径可写性检查（future）

---

## 6. 使用方式

### 6.1 FX 依赖注入方式（推荐）

```go
package main

import (
    "go.uber.org/fx"
    "ag-core/ag/ag_conf"
    "ag-core/ag/ag_log"
)

func main() {
    // 启动 FX
    app := fx.New(
        ag_log.FxAgSlogProvide,  // ag_log fx provider
        ag_conf.FxAgConfProvide, // 配置 fx provider
        // 其他模块...
    )
    app.Run()

    // 获取 logger
    logger := agslog.MustGetLogger()
    logger.Info("hello world")
}
```

**优点**：
- 自动依赖注入
- 配置自动绑定
- 生命周期管理

### 6.2 直接构建方式

```go
package main

import (
    "log/slog"
    "ag-core/ag/ag_log/agslog
)

func main() {
    // 创建 builder
    builder := agslog.NewBuilder()

    // 添加 handler
    builder.AddHandlers(handlers...)

    // 添加中间件
    builder.AddMiddlewares(middlewares...)

    // 构建 logger
    logger, err := builder.Build()
    if err != nil {
        panic(err)
    }

    // 使用
    logger.Info("hello world")
}
```

**适用场景**：
- 需要精细控制构建过程
- 不使用 FX 框架

### 6.3 Context 传递方式

```go
// 传递 handler 名称
ctx := context.WithValue(
    context.Background(),
    slog.HandlerKey,
    slog.HandlerOptions{
        Level: slog.LevelInfo,
    },
).WithAttrs([]slog.Attr{
    slog.String("handler_name", "zap1"),
})

logger.Info("message", slog.Handler(ctx))
```

**作用**：
- 在特定上下文中使用指定 Handler
- 支持日志追踪和层级关系

### 6.4 动态配置

```go
// 获取 builder
builder := agslog.MustGetBuilder()

// 重新加载配置
builder.Reload()

// 获取 logger
logger := agslog.MustGetLogger()
logger.Info("message")
```

**注意**：Reload 功能需要在 Builder 中实现（当前版本未完全实现）。

---

## 7. 扩展机制

### 7.1 添加新的日志框架

**步骤**：

1. **创建封装目录**
```bash
mkdir -p ag/ag_log/sloglogrus
```

2. **实现 Handler 创建**
```go
package sloglogrus

import (
    "log/slog"
    "ag-core/ag/ag_log/agslog"
)

func NewSlogHandler4LogrusProps(props *SlogLogrusProperties) ([]slog.Handler, error) {
    var handlers []slog.Handler
    
    for k, v := range props.Logs {
        // 创建 logrus.Logger
        logger := createLogrusLogger(&v)
        
        // 转换为 slog.Handler
        handler := logrus2slog.NewLogrusHandler(logger)
        
        // 封装为 NamedHandler
        nhandler := agslog.NewNamedHandler(k, handler)
        handlers = append(handlers, nhandler)
    }
    
    return handlers, nil
}
```

3. **创建 FX Provider**
```go
package sloglogrus

import (
    "go.uber.org/fx"
    "ag-core/ag/ag_conf"
)

var FxSlogLogrusProvide = fx.Provide(
    BindSlogLogrusProperties,
    fx.Annotate(
        NewSlogHandler4LogrusProps,
        fx.ResultTags(`group:"agslog.handlers"`),
    ),
)

func BindSlogLogrusProperties(binder ag_conf.IBinder) (*SlogLogrusProperties, error) {
    prop := &SlogLogrusProperties{}
    err := binder.Bind(prop, "aglog.logrus")
    if err != nil {
        return nil, err
    }
    return prop, nil
}
```

4. **更新 FX 配置**
```go
fx.New(
    ag_log.FxAgSlogProvide,
    sloglogrus.FxSlogLogrusProvide,  // 添加新的 provider
)
```

5. **添加配置**
```yaml
aglog:
  topHandler:
    - logrus1

aglog.logrus.logs:
  logrus1:
    log_level: info
    format: json
```

### 7.2 自定义 Handler

**接口定义**：
```go
type ISlogHandler interface {
    NewHandler(props SlogHandlerProperties) slog.Handler
    GetName() string
}
```

**实现示例**：
```go
type CustomHandler struct {
    name  string
    props SlogHandlerProperties
}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
    return level >= slog.LevelInfo
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
    // 自定义处理逻辑
    return nil
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    // 返回带有新属性的新 handler
    return h
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
    // 返回带有分组的新 handler
    return h
}
```

**注册自定义 Handler**：
```go
customHandler := &CustomHandler{
    name:  "customHandler",
    props: props,
}

builder.AddHandler(customHandler)
```

### 7.3 中间件开发

**中间件类型**：
```go
type Middleware func(next slog.Handler) slog.Handler
```

**示例中间件**：
```go
// 添加时间戳中间件
func AddTimestampMiddleware(next slog.Handler) slog.Handler {
    return &timestampHandler{next: next}
}

type timestampHandler struct {
    next slog.Handler
}

func (h *timestampHandler) Enabled(ctx context.Context, level slog.Level) bool {
    return h.next.Enabled(ctx, level)
}

func (h *timestampHandler) Handle(ctx context.Context, r slog.Record) error {
    r.AddAttrs(slog.Time("timestamp", time.Now()))
    return h.next.Handle(ctx, r)
}

// 添加中间件
builder.AddMiddlewares(AddTimestampMiddleware)
```

---

## 8. 技术实现细节

### 8.1 NamedHandler 的工作原理

**Context 传递流程**：

```
调用: logger.Info("message")
    ↓
slog.Logger.Handler().Handle(ctx, record)
    ↓
NamedHandler.Handle(ctx, record)
    ↓
1. 从 Context 获取当前 handler 名称: "root"
2. 构造新名称: "root.zap1"
3. 更新 Context: WithValue(HandlerNameCtxKey, "root.zap1")
4. 设置起始点: HandlerStartCtxKey = "root"
5. 设置结束点: HandlerEndCtxKey = "zap1"
    ↓
原始 Handler.Handle(newCtx, record)
```

**层级命名规则**：
- 根节点：直接使用 Handler 名称
- 子节点：父节点名称 + "." + 当前节点名称
- 示例：`root.zap1`, `root.f1.zap1`

### 8.2 HandlerFactory 的递归解析

**递归解析示例**：

假设配置：
```yaml
topHandler: [f1]
fanout.logs:
  f1: [f2, zap1]
  f2: [zap2, zap3]
```

解析流程：
```
1. resolveHandler("f1")
   ↓
2. HandlerFactory[f1].GetHandler(resolveHandler)
   ↓
3. DoGetHandler(resolveHandler) 被调用
   - 解析 "f2": resolveHandler("f2")
     - 解析 "zap2": resolveHandler("zap2") → 返回 zap2 handler
     - 解析 "zap3": resolveHandler("zap3") → 返回 zap3 handler
     - Fanout(zap2, zap3) → f2 handler
   - 解析 "zap1": resolveHandler("zap1") → 返回 zap1 handler
   - Fanout(f2, zap1) → f1 handler
   ↓
4. 返回 f1 handler
```

**循环调用检测**：
```go
func (f *HandlerFactory) GetHandler(...) (slog.Handler, error) {
    ok := f.mu.TryLock()
    if !ok {
        return nil, fmt.Errorf("circular call detected: %s", f.Name)
    }
    defer f.mu.Unlock()
    // ...
}
```

### 8.3 Builder 的构建流程

**完整构建流程**：

```
1. Build()
   ↓
2. initTopLogger()
   ↓
3. resolveTopHandlers()
   - 解析配置中的 topHandler
   - 查找每个 handler 的工厂
   - 调用 GetHandler 递归解析
   ↓
4. Fanout 组合多个 topHandlerHandler
   ↓
5. 应用中间件管道
   ↓
6. wrapNamedHandlerIfNeed - 包装为 NamedHandler
   ↓
7. tryReplaceNamedHandler - 替换可替换的 handler
   ↓
8. 替换全局默认 Logger（如果 IsDefault = true）
   ↓
9. 返回 slog.Logger
```

### 8.4 缓存机制和线程安全

**三层缓存**：

1. **handlersCaches** (sync.Map)
   - 缓存原子 handler
   - Key: handler 名称
   - Value: INamedHandler

2. **handlers** (sync.Map)
   - 缓存层级 handler（已应用中间件）
   - Key: handler 名称
   - Value: INamedHandler

3. **namedLogger** (sync.Map)
   - 缓存命名 logger
   - Key: logger 名称
   - Value: *slog.Logger

**线程安全保证**：
- 使用 `sync.Map` 保证缓存的并发安全
- HandlerFactory 使用 `sync.Mutex` 避免并发创建
- Builder 使用 `sync.Mutex` 保护 logger 创建

---

## 9. 性能优化

### 9.1 缓存策略

**Handler 缓存**：
```go
// 第一次调用
handler, err := b.resolveHandler("zap1")
// handler 被缓存在 handlersCaches

// 第二次调用
handler, err := b.resolveHandler("zap1")
// 直接从缓存返回，无需重新创建
```

**Logger 缓存**：
```go
logger := builder.GetSlogByName("zap1")
// logger 被缓存在 namedLogger

// 后续调用直接返回缓存
logger := builder.GetSlogByName("zap1")
```

### 9.2 并发安全设计

**互斥锁使用**：
- HandlerFactory.GetHandler: 防止并发创建同一 handler
- Builder.logMu: 保护 logger 创建

**sync.Map 使用**：
- handlersCaches: 处理 handler 缓存
- handlers: 处理层级 handler 缓存
- namedLogger: 处理 logger 缓存
- replaceableHandlers: 处理可替换 handler

### 9.3 生产环境配置建议

**日志级别**：
```yaml
log_level: info  # 生产环境使用 info 或 warn
```

**日志滚动**：
```yaml
max_size: 100      # 100MB
max_backups: 10    # 保留 10 个备份
max_age: 30        # 保留 30 天
compress: true     # 压缩备份文件
```

**编码格式**：
```yaml
console: false      # 使用 JSON 编码（生产环境）
prod: true         # 启用生产模式优化
```

**性能优化配置**：
```yaml
aglog:
  is_default: true  # 替换 slog 默认 logger
  topHandler:
    - f1           # 使用 fanout 提高吞吐量
```

### 9.4 性能考虑

**避免频繁创建 Handler**：
- 使用缓存机制
- 共享 Handler 实例

**合理设置日志级别**：
- 在 Handler 层面过滤日志
- 避免不必要的日志记录

**使用 fanout 模式**：
- 并发写入多个目标
- 提高日志吞吐量

**生产模式优化**：
- `prod: true` 减少运行时开销
- JSON 编码比 console 编码更快

---

## 10. 限制和已知问题

### 10.1 当前实现的限制

1. **异步日志**：当前不支持异步日志写入
2. **日志日志服务扩展**：暂不支持分布式日志服务集成
3. **日志分类**：缺乏全局日志分类管理
4. **链路追踪**：暂不支持分布式链路追踪
5. **动态配置**：Reload 功能未完全实现

### 10.2 已知问题

1. **循环调用检测**：HandlerFactory 的 TryLock 检测需要更多测试验证
2. **配置验证**：配置验证不够完善（如文件路径可写性）
3. **错误处理**：部分错误信息不够详细
4. **日志级别配置**：部分地方硬编码日志级别

### 10.3 TODO 清单

- [ ] 异步文件日志实现
- [ ] 日志日志服务扩展支持（如 ELK、Loki）
- [ ] 日志分类（全局日志管理器）
- [ ] 链路追踪信息支持（OpenTelemetry）
- [ ] 动态配置和热重载
- [ ] 配置验证完善
- [ ] 性能监控和统计
- [ ] 详细的单元测试
- [ ] 集成测试和性能测试

---

## 11. 最佳实践

### 11.1 推荐的使用模式

#### 11.1.1 微服务架构

```yaml
# 多环境配置
aglog:
  is_default: true
  topHandler:
    - fileHandler
    - consoleHandler

aglog.zap.logs:
  fileHandler:
    log_file_name: ./logs/service.log
    log_level: info
    encoding: json
    max_size: 100
    max_backups: 10
    compress: true
    prod: true

  consoleHandler:
    log_level: debug
    encoding: console
```

#### 11.1.2 开发环境

```yaml
aglog:
  is_default: true
  topHandler:
    - consoleHandler

aglog.zap.logs:
  consoleHandler:
    log_level: debug
    encoding: console
    stdout: true
```

### 11.2 常见陷阱和解决方案

#### 11.2.1 Handler 名称冲突

**问题**：多个 Handler 使用相同名称

**解决方案**：确保每个 Handler 名称唯一
```yaml
aglog.zap.logs:
  handler1:  # 唯一名称
    ...
  handler2:  # 唯一名称
    ...
```

#### 11.2.2 循环依赖

**问题**：fanout 配置导致循环依赖
```yaml
fanout.logs:
  f1: [f2]
  f2: [f1]  # 循环依赖
```

**解决方案**：避免循环引用，使用清晰的层次结构

#### 11.2.3 日志文件权限

**问题**：日志文件无写权限导致日志写入失败

**解决方案**：
- 确保日志目录存在且有写权限
- 在应用启动前创建日志目录
- 考虑添加配置验证

#### 11.2.4 内存泄漏

**问题**：Handler 和 Logger 不释放

**解决方案**：
- 使用缓存机制避免重复创建
- 长期运行的应用注意监控内存使用
- 考虑添加资源清理机制

### 11.3 性能调优建议

1. **使用 fanout 提高吞吐量**：多目标并发写入
2. **合理设置日志级别**：避免不必要的日志记录
3. **使用生产模式**：`prod: true` 减少开销
4. **启用日志压缩**：节省磁盘空间
5. **监控日志性能**：添加日志写入耗时统计
6. **异步日志**：考虑实现异步日志写入（TODO）

---

## 12. 示例和应用场景

### 12.1 微服务架构中的日志配置

```yaml
# config/app.yaml
aglog:
  is_default: true
  topHandler:
    - fileHandler
    - consoleHandler

aglog.zap.logs:
  fileHandler:
    log_file_name: ./logs/{{.ServiceName}}.log
    log_level: info
    encoding: json
    max_size: 100
    max_backups: 10
    max_age: 30
    compress: true
    prod: true

  consoleHandler:
    log_level: debug
    encoding: console
    stdout: true
```

**使用**：
```go
logger := agslog.MustGetLogger()
logger.Info("request received",
    "service", "user-service",
    "method", "POST",
    "path", "/api/users",
)
```

### 12.2 多环境日志配置

**开发环境** (`config/dev.yaml`)：
```yaml
aglog:
  is_default: true
  topHandler:
    - consoleHandler

aglog.zap.logs:
  consoleHandler:
    log_level: debug
    encoding: console
    stdout: true
```

**生产环境** (`config/prod.yaml`)：
```yaml
aglog:
  is_default: true
  topHandler:
    - fileHandler
    - errorHandler

aglog.fanout.logs:
  errorFanout:
    - fileHandler
    - errorAlertService

aglog.zap.logs:
  fileHandler:
    log_file_name: /var/log/app/production.log
    log_level: info
    encoding: json
    max_size: 500
    max_backups: 20
    compress: true
    prod: true

  errorAlertService:
    log_level: error
    # 错误日志发送到告警服务
```

### 12.3 日志分类和路由

```yaml
aglog:
  is_default: true
  topHandler:
    - routerHandler

aglog.fanout.logs:
  routerHandler:
    - accessLog
    - businessLog
    - errorLog

aglog.zap.logs:
  accessLog:
    log_file_name: ./logs/access.log
    log_level: info
    encoding: json

  businessLog:
    log_file_name: ./logs/business.log
    log_level: info
    encoding: json

  errorLog:
    log_file_name: ./logs/error.log
    log_level: error
    encoding: json
```

**使用**：
```go
// 获取特定的 logger
accessLogger := agslog.MustGetBuilder().GetSlogByName("accessLog")
businessLogger := agslog.MustGetBuilder().GetSlogByName("businessLog")
errorLogger := agslog.MustGetBuilder().GetSlogByName("errorLog")

accessLogger.Info("api access", "method", "GET", "path", "/api/users")
businessLogger.Info("user created", "user_id", "12345")
errorLogger.Error("database error", "error", err)
```

---

## 附录

### A. 相关资源

- [Go slog 官方文档](https://pkg.go.dev/log/slog)
- [Uber Zap 官方文档](https://pkg.go.dev/go.uber.org/zap)
- [slog-multi 文档](https://pkg.go.dev/github.com/samber/slog-multi)
- [slog-zap 文档](https://pkg.go.dev/github.com/samber/slog-zap/v2)
- [FX 官方文档](https://pkg.go.dev/go.uber.org/fx)
- [Lumberjack 文档](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2)

### B. 版本历史

| 版本 | 日期 | 变更说明 |
|------|------|----------|
| 1.0 | 2026-04-16 | 初始版本 |

### C. 变更记录

**v1.0 (2026-04-16)**
- 创建初始架构设计文档
- 记录核心组件和配置系统
- 添加使用示例和最佳实践

---

**文档结束**
