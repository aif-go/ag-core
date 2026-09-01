# AgCache v1.0 定版 API 实施

## Why

AgCache 已完成 v2（`agcache-v2`）与接口定稿 v3（`agcache-interface-refine`）实现。经多轮设计讨论（SPI 审计、Spring Cache 对齐、简单+易用收敛、TTL 归属、fx 多引擎组合），定版 API 方向已在 `2026-08-26-ag-cache-v1-final-api.md` 落档。本 change 将该**定版 API** 落地：入口重构（`GetCache`/`GetCacheWithLoader`/`DefaultManager`）、TTL 语义（`SetWithTTL`/`TTLSetter`/`WithDefaultTTL` 优先级链）、引擎模型（`EngineFactory{Name;Create(name)}` fx group 注入 + config 选默认）、key 前缀方案 A、`Engine.Set` 无 ttl + `Engine.Clear(ctx, prefix)`，使 ag_cache 达到 v1.0 稳定可用。

## What Changes

- **入口重构（简单 + 易用）**：
  - `NewCache` → **`GetCacheWithLoader[T](m *Manager, name, loader, opts...)`**（显式 Manager，读穿透，构造绑定一次）
  - `GetCache[T](m *Manager, name)`（纯读）
  - 新增 `DefaultManager() *Manager`（运行时获取 default，替代不提供的 `GetCacheWithLoaderDefault`）
  - 保留 `SetDefault` / `CloseAll` / `WithLoader` / `NewWithEngine`（底层直用）
- **TTL 语义（优先级链）**：
  - `ICache` 新增 `SetWithTTL(ctx, key, value, ttl)`（单条显式，`ttl=0` 永不过期）
  - `Engine.Set(ctx, key, value)` **无 ttl 参数**（引擎内部默认 TTL，自身配置/实现决定）
  - 可选接口 `TTLSetter{ SetWithTTL(ctx,key,value,ttl) }`：业务经 `ICache.SetWithTTL`/`WithDefaultTTL` 显式指定 TTL；引擎不实现 → 同 `Set`
  - 保留 `WithDefaultTTL`（业务 per-cache 默认，经引擎 `TTLSetter` 实现）
  - **删除** `DefaultTTLProvider`（TTL 归属引擎内部，core 不再管默认 TTL）
- **引擎模型（fx 多引擎组合）**：
  - `EngineFactory` 改 `{ Name() string; Create(name string) (Engine, error) }`
  - `Manager` 收集多种 `EngineFactory`（fx `group:"agcache.engine"` 注入），config `defaultEngine`（`agcache.defaultEngine`，默认 `ristretto`）选**默认引擎**
  - **删除**全局注册表 `RegisterEngine`/`EngineRegistered`/`getFactory`（fx group app 作用域替代）；**删除** `WithEngine` option（默认引擎由 config 选）
  - **fail-fast**：config 指定默认引擎未注册 → Manager 装配即报错
- **key 前缀方案 A（对齐 Spring）**：
  - `typedCache` 新增 `name` / `prefix` 字段，核心统一拼 `agcache::<name>::<key>` 完整 key 传 `Engine`
  - `Engine` 零 name 概念（对齐 Spring `RedisCacheWriter` 收完整 key）
  - `Engine.Clear(ctx)` → `Clear(ctx, prefix string)`（唯一需 namespace 范围的引擎操作）
- **命名确认**：`GetCacheWithLoader`（读穿透）+ `GetCache`（纯读）同级对称，砍掉 `Cache[T]` 嵌入、`GetCacheWithLoaderDefault`、opts 冲突校验、切面 AOP（均后置/文档化）

## Capabilities

### New Capabilities
- `agcache-v1-entry`: 定版入口 + TTL 业务 API——`GetCacheWithLoader`/`GetCache`（显式 Manager）+ `DefaultManager()` getter；`ICache.SetWithTTL`（单条显式 TTL）；`WithDefaultTTL`（per-cache 默认）；业务构造注入 Manager 绑定一次、方法内复用；时序由 fx 依赖图保证
- `agcache-key-prefix`: key 前缀方案 A——`typedCache` 持 name/prefix、统一拼 `agcache::<name>::<key>`；`Engine` 零 name；`Engine.Clear(ctx, prefix)`
- `agcache-engine-adapter`: 引擎适配——`Engine.Set` 无 ttl（内部默认）、可选 `TTLSetter`；`EngineFactory{Name;Create(name)}` fx group 注入 + config 选默认 + fail-fast；agristretto 收完整 key、Clear 忽略 prefix、内部默认 TTL；Mock 引擎签名适配；删注册表/WithEngine

### Modified Capabilities
<!-- 无：openspec/specs/ 尚无现有 spec -->

## Impact

- **修改文件**：
  - `ag/ag_cache/manager.go`：`NewManager(props)`（defaultEngine=config）、`SetEngineFactory(name,f)`、`GetCacheWithLoader(m,...)`、`GetCache(m,...)`、新增 `DefaultManager()`、getOrCreate 用默认引擎工厂 `Create(name)`、fail-fast 校验
  - `ag/ag_cache/typed.go`：typedCache 加 `name`/`prefix`；`Set` 配了 `WithDefaultTTL` → 经 `TTLSetter`；未配 → `engine.Set`；`SetWithTTL` → 探测 `TTLSetter`；删 `engineName`/`WithEngine`；各 engine 调用点拼前缀；singleflight key 拼前缀
  - `ag/ag_cache/cache.go`：`ICache` 加 `SetWithTTL`
  - `ag/ag_cache/engine.go`：`Engine.Set` 无 ttl；`Clear(ctx, prefix string)`；加 `TTLSetter`；删 `DefaultTTLProvider`/`RegisterEngine`/`EngineRegistered`/`getFactory`/`Name`（注册表）
  - `ag/ag_cache/agristretto/ristretto.go`：`ristrettoEngine` 内部 `defaultTTL`（config 喂入）；`SetWithTTL`；`Clear` 忽略 prefix；`agristrettoFactory{Name(); Create(name)}`；删 `DefaultTTL()`
  - `ag/ag_cache/agristretto/config_ristretto.go`：`defaultTtl` 保留（喂 ristrettoEngine 内部默认）
  - `ag/ag_cache/zfx_ag_cache.go`：`NewAgCacheManager` 消费 group（`[]EngineFactory`）填 map + config 选默认 + fail-fast；保留 `SetDefault`
  - `ag/ag_cache/mock.go`：`MockEngine` 适配 `Set` 无 ttl、`Clear(ctx,prefix)`、可选 `SetWithTTL`
  - `ag/ag_cache/config.go`：`AgCacheProperties.DefaultEngine`（config 选默认引擎）
- **测试迁移**：`cache_test.go`/`manager_test.go`/`zfx_ag_cache_test.go`/`agristretto/ristretto_test.go`/`test/*`/`test/usage/*` 改用新入口/新 SPI
- **文档**：`2026-08-26-ag-cache-v3-usage.md` 更新 API 名（`GetCacheWithLoader`/`GetCache`/`SetWithTTL`）
- **无破坏性变更**（对 `ag_log`/`contribute/*`/`fxs/*`/`go.work`/`release_aif.sh`）
- **破坏性影响（相对 v3 8450509）**：`New`/`Get` 包级 default 版移除、`Engine.Set` 签名变化、`Engine.Clear` 签名变化、注册表 `RegisterEngine`/`EngineRegistered` 移除、`WithEngine` option 移除
