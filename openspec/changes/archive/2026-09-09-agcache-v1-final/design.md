# AgCache v1.0 定版 API 实施设计

## Context

- **现状**：AgCache 已完成 v2（`agcache-v2`）与接口定稿 v3（`agcache-interface-refine`，commit `8450509`）。经 SPI 审计与多轮设计讨论（TTL 归属、fx 多引擎组合），定版方向在 `2026-08-26-ag-cache-v1-final-api.md` 落档。
- **目标**：把定版 API 落地：入口重构（`GetCacheWithLoader`/`GetCache`/`DefaultManager`）、TTL 语义（`SetWithTTL`/`TTLSetter`/`WithDefaultTTL` 优先级链）、引擎模型（`EngineFactory{Name;Create(name)}` fx group + config 选默认）、key 前缀方案 A、`Engine.Set` 无 ttl + `Engine.Clear(ctx, prefix)`，使 ag_cache 达 v1.0 稳定可用。

## Goals / Non-Goals

**Goals:**
- 入口简单易用：`GetCacheWithLoader`/`GetCache`（显式 Manager）+ `DefaultManager()`；业务构造绑定一次、方法内复用
- TTL 优先级链：`SetWithTTL` > `WithDefaultTTL`（经引擎 TTLSetter）> `engine.Set`（引擎内部默认）
- 引擎模型：Manager 收集多种 `EngineFactory`（fx group）+ config 选默认 + fail-fast；删全局注册表/WithEngine
- key 前缀方案 A：typedCache 拼 `agcache::<name>::<key>`，Engine 零 name
- `Engine.Set` 无 ttl；`Engine.Clear(ctx, prefix)`；agristretto/mock 适配

**Non-Goals:**
- `Cache[T]` 嵌入类型（后置）
- opts 冲突校验（文档规范）
- 切面 AOP / 无感（Go 无运行时 AOP，后置）
- Redis 引擎实现（方向 3）
- per-name 多引擎（users local/params redis 同时，后置；当前所有 name 用默认引擎）
- 遗留边界 P4/P5/异步 Set（另开）

## Decisions

### D1. 入口重构（简单 + 易用）
```go
func GetCacheWithLoader[T any](m *Manager, name string, loader LoaderFunc[T], opts ...Option[T]) *LoaderCache[T]
func GetCache[T any](m *Manager, name string) ICache[T]
func DefaultManager() *Manager
```
- 移除 v3 包级 `New`/`Get`（走 default）→ 显式 `m` 参数
- **命名**：`GetCacheWithLoader`（读穿透）+ `GetCache`（纯读）同级对称（都获取缓存实例，差异在 loader 绑定），底层对应 getOrCreate
- 新增 `DefaultManager()` getter；**不提供 `GetCacheWithLoaderDefault`**（运行时 `DefaultManager()` 获取自用）
- 业务构造注入 `*Manager`，`GetCacheWithLoader(m, ...)` 绑定一次到字段，方法内复用
- 时序：业务构造依赖 `*Manager` → fx 依赖拓扑保证 Manager 先建
- 保留 `SetDefault`/`CloseAll`/`WithLoader`/`NewWithEngine`

### D2. TTL 语义（优先级链）
```
SetWithTTL(ctx, key, value, ttl)     —— 单条显式（最优先，ttl=0 永不过期）
> WithDefaultTTL                     —— 业务 per-cache 默认（经引擎 TTLSetter）
> engine.Set                        —— 引擎内部默认（业务未配时，引擎配置/实现决定）
```
```go
// ICache 新增
SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) error

// 可选接口（Engine SPI 之外）
type TTLSetter interface { SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error }
```
- `Engine.Set(ctx, key, value []byte)` **无 ttl 参数**——默认 TTL 由引擎内部决定
- typedCache 持可选 `defaultTTL`/`ttlSet`（`WithDefaultTTL` 配置时）：
  - `Set`：`ttlSet` → 探测 `TTLSetter` 用默认；未配 → `engine.Set`
  - `SetWithTTL`：探测 `TTLSetter` 传显式 ttl；不实现 → `engine.Set`（忽略 ttl）
- **删除** `DefaultTTLProvider`（TTL 归属引擎内部，core 不再管默认 TTL）
- 保留 `WithDefaultTTL`（业务 per-cache 默认，经 TTLSetter）

### D3. 引擎模型（fx 多引擎组合）
```go
type EngineFactory interface {
    Name() string
    Create(name string) (Engine, error)
}
```
- `Manager` 收集**多种 `EngineFactory`**（fx `group:"agcache.engine"` 注入，各引擎模块 Provide 一个工厂），config `defaultEngine`（`agcache.defaultEngine`，默认 `ristretto`）选默认引擎
- fx 友好：Manager 收 `[]EngineFactory`（group），多引擎模块 Provide 不冲突（单类型 + 多提供者致 fx "duplicate provider" 报错）
- 工厂无状态：每次 `Create(name)` 新实例；实例复用与生命周期归 Manager（`getOrCreate` 懒建复用，`Manager.Close` 统一关）
- **fail-fast**：config 指定默认引擎未注册 → 装配即报错
- **删除**全局注册表 `RegisterEngine`/`EngineRegistered`/`getFactory`；**删除** `WithEngine` option（默认引擎由 config 选）
- 保留 `NewWithEngine`（底层直用/测试）

### D4. key 前缀方案 A（对齐 Spring）
```go
type typedCache[T any] struct {
    ...
    name   string   // 缓存名
    prefix string   // "agcache::users::"
}
```
- typedCache 统一拼 `prefix + key` 传 Engine；Engine 零 name（对齐 Spring RedisCacheWriter）
- 前缀格式 `agcache::<name>::<key>`：`::` 双冒号避免业务 key 单冒号歧义；`agcache` 框架前缀未来共享 Redis 隔离
- singleflight key、DelMany 也拼前缀

### D5. Engine.Clear 带 prefix
```go
Clear(ctx context.Context, prefix string) error
```
- 唯一需 namespace 范围的引擎操作；local 忽略、Redis 按前缀 SCAN+DEL

## 文件与函数签名清单

### `ag/ag_cache/manager.go`
- `NewManager(props *AgCacheProperties) (*Manager, error)`：`defaultEngine = props.DefaultEngine`，初始化空 engineFactories map
- 新增 `SetEngineFactory(name string, f EngineFactory)`（fx 装配填 map）
- 新增 `DefaultEngine() string`（fail-fast 校验用）
- `New` → `GetCacheWithLoader[T](m *Manager, name, loader, opts...) *LoaderCache[T]`
- `Get` → `GetCache[T](m *Manager, name) ICache[T]`
- `getOrCreate`：`m.engineFactories[m.defaultEngine].Create(name)`，懒建复用，`Manager.Close` 统一关
- 新增 `DefaultManager() *Manager`（`defaultManager.Load()`）
- 保留 `SetDefault`/`CloseAll`/`Close`

### `ag/ag_cache/engine.go`
- `Engine.Set(ctx context.Context, key string, value []byte) error`（无 ttl）
- `Engine.Clear(ctx context.Context, prefix string) error`
- 新增可选 `TTLSetter`
- **删除** `DefaultTTLProvider`、`RegisterEngine`/`EngineRegistered`/`getFactory`/`EngineFactory.Name()` 注册表逻辑

### `ag/ag_cache/typed.go`
- `typedCache` 加 `name string` / `prefix string`；保留可选 `defaultTTL`/`ttlSet`（`WithDefaultTTL`）
- **删除** `engineName` 字段、`WithEngine` option
- `Set`：`ttlSet` → 探测 `TTLSetter` 用默认；未配 → `engine.Set`
- `SetWithTTL`：探测 `TTLSetter` 传显式 ttl；不实现 → `engine.Set`
- Get/TryGet/GetOrElse/Set/SetWithTTL/Del：engine 调用点拼 `prefix+key`
- singleflight key 拼前缀；DelMany 拼前缀
- `Clear`：`engine.Clear(ctx, c.prefix)`

### `ag/ag_cache/cache.go`
- `ICache[T]` 加 `SetWithTTL(ctx context.Context, key string, value T, ttl time.Duration) error`

### `ag/ag_cache/agristretto/ristretto.go`
- `ristrettoEngine` 加 `defaultTTL time.Duration` 字段（构造时 config `defaultTtl` 喂入，引擎内部默认）
- `Set`：用内部 `defaultTTL`；`SetWithTTL`：Ristretto 原生显式 ttl
- `Clear(ctx, prefix)` 忽略 prefix 清实例
- `agristrettoFactory{Name() string; Create(name string) (Engine, error)}`（每次新实例）
- **删除** `DefaultTTL()`（TTL 归 ristrettoEngine 内部字段）

### `ag/ag_cache/agristretto/config_ristretto.go`
- `defaultTtl` 保留（喂 ristrettoEngine 内部默认）；`parseTTL` 保留

### `ag/ag_cache/zfx_ag_cache.go`
- `NewAgCacheManager(p EngineFactoryParams, props)`：消费 group（`[]EngineFactory`）填 map + config 选默认 + fail-fast；保留 `SetDefault`
- OnStop 只 Close

### `ag/ag_cache/mock.go`
- `MockEngine` 适配 `Set(ctx,k,v)` 无 ttl、`Clear(ctx,prefix)`；可选实现 `SetWithTTL`

## 测试策略

| 层 | 文件 | 覆盖 |
|----|------|------|
| core 单元 | `cache_test.go`/`manager_test.go` | GetCacheWithLoader/GetCache 读穿透/纯读、DefaultManager、构造绑定一次、SetWithTTL/TTL 优先级链 |
| TTL | `manager_test.go`/`cache_test.go` | SetWithTTL 显式 ttl、WithDefaultTTL 经 TTLSetter、ttl=0、引擎不实现 TTLSetter → 同 Set |
| 引擎模型 | `manager_test.go`/`zfx_ag_cache_test.go` | fx group 注入、config 选默认、fail-fast、Create(name) |
| key 前缀 | `manager_test.go`/新测试 | 前缀拼装（recording 引擎验证收到 `agcache::users::u:1`）、singleflight 前缀、DelMany 前缀 |
| 引擎 | `agristretto/ristretto_test.go` | Set 用内部默认、SetWithTTL、Clear 忽略 prefix、收完整 key、Create(name) |
| 集成 | `test/*`、`test/usage/*` | 迁移新入口/新 SPI，端到端 |

验证：`go test -race ./ag/ag_cache/...`、`go vet ./ag/ag_cache/...`、`go build ./...`。

## Risks / Trade-offs

- [R1] **破坏性变更**：v3 的 `New`/`Get` 移除、`Engine.Set`/`Engine.Clear` 签名变化、注册表/WithEngine 移除——需迁移测试/usage。
- [R2] **前缀冗余**：local 独立实例下前缀冗余（key 变长），换取架构统一（Engine 零 name、Redis 实现简单）。
- [R3] **DefaultManager nil**：未 `SetDefault` 时 `DefaultManager()` 返回 nil，业务需判空。
- [R4] **fail-fast**：config 默认引擎未注册 → 装配报错（比静默回退清晰，但要求模块已 Provide）。

## Migration Plan

- 按「十一、迁移清单」实施：manager/typed/cache/engine/agristretto/config/zfx/mock + 测试迁移 + usage 适配 + 文档同步（v3-usage 更新 API 名）
- 回滚：未合入前丢弃

## Open Questions

- 无（定版方向已在 `2026-08-26-ag-cache-v1-final-api.md` 落档）
