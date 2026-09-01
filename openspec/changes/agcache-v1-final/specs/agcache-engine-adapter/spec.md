## ADDED Requirements

### Requirement: Engine SPI（Set 无 ttl + Clear 带 prefix）
`Engine` SHALL 提供 `Get(ctx,key)`/`Set(ctx,key,value)`/`Del(ctx,key)`/`Clear(ctx,prefix)`/`Close()`。
`Engine.Set(ctx, key, value []byte)` SHALL **无 ttl 参数**——默认 TTL 由引擎内部决定（自身配置 `defaultTtl` / 自然行为）。
`Engine.Clear(ctx context.Context, prefix string)` SHALL 接收 namespace 前缀，供共享后端引擎（Redis）按前缀 SCAN+DEL 清 namespace。
local 引擎（agristretto）SHALL 忽略 prefix（清独立实例）。
agristretto（local）SHALL 接收已拼前缀的完整 key（无需 name 概念）。
MockEngine SHALL 适配 `Set` 无 ttl 与 `Clear(ctx, prefix)` 签名。
MockCache 等测试替身 SHALL 保持 `ICache` 契约（key 为业务 key，前缀由 core 拼装，mock 不感知前缀）。

#### Scenario: Engine.Set 无 ttl
- **WHEN** 业务 `Set(ctx, key, value)`（未配 WithDefaultTTL）
- **THEN** 引擎按其内部默认 TTL 写入（core 不传 ttl）

#### Scenario: agristretto 收完整 key
- **WHEN** engine 收到 `"agcache::users::u:1"` 并 `Set`/`Get`
- **THEN** 按完整 key 存取（与业务 key 无关）

#### Scenario: local Clear 忽略 prefix
- **WHEN** `typedCache.Clear(ctx)` 调用 `engine.Clear(ctx, "agcache::users::")`
- **THEN** agristretto 清空其独立实例，忽略 prefix

#### Scenario: MockEngine.Clear 适配
- **WHEN** `MockEngine.Clear(ctx, "agcache::users::")` 调用
- **THEN** 清空 mock 数据，忽略 prefix

### Requirement: TTLSetter（可选外部 TTL 接口）
可选接口 `TTLSetter{ SetWithTTL(ctx, key, value []byte, ttl time.Duration) error }` SHALL 支持外部指定 TTL（业务经 `ICache.SetWithTTL`/`WithDefaultTTL` 显式指定）。
引擎不实现 `TTLSetter` SHALL 将外部 TTL 忽略，等同 `Set`（无 TTL 引擎）。
`DefaultTTLProvider` SHALL 被删除（TTL 归属引擎内部，core 不再管默认 TTL）。

#### Scenario: 引擎实现 TTLSetter
- **WHEN** 业务 `SetWithTTL(ctx, key, v, 60s)` 且引擎实现 `TTLSetter`
- **THEN** 按显式 ttl 写入（60s）

#### Scenario: 引擎不实现 TTLSetter
- **WHEN** 业务 `SetWithTTL(ctx, key, v, 60s)` 且引擎未实现 `TTLSetter`
- **THEN** 等同 `Set`，用引擎内部默认 TTL（显式 ttl 被忽略，不报错）

### Requirement: EngineFactory 与 fx group 注入
`EngineFactory` SHALL 提供 `Name() string`（map 键）与 `Create(name string) (Engine, error)`（name 为命名空间上下文，每次新创建实例）。
`Manager` SHALL 收集多种 `EngineFactory`（fx `group:"agcache.engine"` 注入，各引擎模块 Provide 一个工厂）。
config `defaultEngine`（`agcache.defaultEngine`，默认 `ristretto`）SHALL 选择默认引擎工厂。
`getOrCreate(name)` SHALL 用默认引擎工厂 `Create(name)` 创建（实例复用与生命周期归 Manager）。
**删除**全局注册表 `RegisterEngine`/`EngineRegistered`/`getFactory`；**删除** `WithEngine` option（默认引擎由 config 选）。
`NewWithEngine[T](engine Engine, opts ...Option[T]) ICache[T]` SHALL 保留（底层直用/测试，不经 Manager）。

#### Scenario: fx group 多引擎注入
- **WHEN** fx 中 `agristretto.FxAgCacheRistrettoMode` Provide 工厂进 `group:"agcache.engine"`
- **THEN** Manager 收集到 agristrettoFactory，config `defaultEngine: ristretto` 选其为默认

#### Scenario: config 选默认引擎
- **WHEN** config `agcache.defaultEngine: redis` 且 redis 模块 Provide 工厂进 group
- **THEN** 默认引擎为 redis；`getOrCreate(name)` 用 redisFactory.Create(name)

#### Scenario: fail-fast 默认引擎未注册
- **WHEN** config `agcache.defaultEngine: redis` 但 group 中无 redis 工厂
- **THEN** Manager 装配即报错（不静默回退）

#### Scenario: 工厂无状态 + Create(name)
- **WHEN** `getOrCreate("users")` 与 `getOrCreate("params")` 各自调用 `factory.Create(name)`
- **THEN** 每次 Create 新实例，Manager 管理复用与 Close 生命周期

### Requirement: 测试替身 key 语义
MockCache（实现 `ICache`）SHALL 使用业务 key（`u:1`），不感知前缀（前缀由 typedCache 拼装后传引擎；mock 作为 ICache 契约测试替身，业务 key 层）。

#### Scenario: MockCache 用业务 key
- **WHEN** `MockCache.Set(ctx, "u:1", v)` 后 `Get(ctx, "u:1")`
- **THEN** 命中（mock 不涉及前缀）
