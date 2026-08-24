# AgCache v2 正式实现

## Why

业务缺少统一缓存抽象，POC（`agcache-poc` v3.3，27 测试全绿）已验证"CacheManager + 独立命名空间实例"模型。v1 落地（`agcache-implement` change）在实现与评审后暴露架构问题：core 绑定引擎配置（`EngineConfig any`）、default/overrides 过度设计、TTL 归错位、`Create(cfg)` 弱类型、且 core 无法 import 引擎子包（循环依赖）。v2 重构为 **core 纯 SPI + 引擎经 fx group 动态注入**，让 core 只做"按缓存名选引擎"，引擎完全自配自注册。

## What Changes

- **新增 `ag/ag_cache` 核心包（SPI 纯净，零引擎依赖）**：
  - `Engine` SPI（全方法带 ctx + 错误契约）；`EngineFactory{Name, Create()}`（**Create 无参**，引擎配置由工厂自持）
  - 可选能力 `DefaultTTLProvider`：引擎自声明默认 TTL，core 兜底 5min
  - 引擎注册表：`RegisterEngine` / `EngineRegistered`
  - 包级实例管理：`New[T](name, loader, opts...)` / `Get[T]` / `GetAdmin[T]` / `CloseAll`（缓存名 → typedCache 懒创建复用）
  - `Option`：`WithEngine`（可选指定引擎实现名）/ `WithDefaultTTL` / `WithSerializer`
  - `typedCache`（序列化 + singleflight + 错误语义 + panic recovery）/ `Serializer` / `LoaderCache` / `MockCache` / `MockEngine`
  - 核心配置仅 `AgCacheProperties{DefaultEngine}`（core 自绑 `agcache.defaultEngine`，只选引擎）
  - `FxAgCacheMode`：经 `group:"agcache.engine"` 动态注入所有引擎工厂 → 幂等注册 → Manager → 生命周期
- **新增 `ag/ag_cache/agristretto` 引擎子包**：
  - Ristretto 引擎实现（`ristrettoEngine`）+ `agristrettoFactory{cfg,ttl}`（`Name()="ristretto"`，实现 `DefaultTTLProvider`）
  - 自绑配置 `agcache.ristretto.*`（`maxCost`/`numCounters`/`defaultTtl`）
  - `FxAgCacheRistrettoMode`：`fx.Provide` 工厂到 `group:"agcache.engine"`（**不直接调用 RegisterEngine**）
- **错误语义**：`ErrCacheMiss`（miss，唯一触发 loader）与 `ErrBackend`（后端故障，绝不调 loader）`errors.Is` 可判
- **主 `go.mod` 新增依赖** `github.com/dgraph-io/ristretto/v2 v2.4.2`
- **对比 v1**：删除 `Config{Default, Overrides}`、`NamespaceConfig`、`EngineConfig any`、`Create(cfg)`、core 端 TTL 配置/绑定模型；引擎命名 `local`→`agristretto`/`ristretto`
- 业务 API 形态保持 POC/v1 一致（`New[T](name, loader)` 3 行体验），新增 `WithEngine`/`WithDefaultTTL` Option

## Capabilities

### New Capabilities
- `agcache-core`: 核心 SPI 抽象——`Engine`/`EngineFactory{Create 无参}`/`DefaultTTLProvider`/注册表/包级实例管理（`New`/`Get`/`GetAdmin`/`CloseAll`）/`typedCache`（序列化+singleflight+错误语义）/`LoaderCache`/`Option`（WithEngine/WithDefaultTTL/WithSerializer）/Mock 工具，含错误契约与核心配置 `AgCacheProperties{DefaultEngine}`
- `agcache-ristretto-engine`: Ristretto 本地引擎实现——`Engine` SPI 实现、`agristrettoFactory`（Name="ristretto"、Create 无参、DefaultTTLProvider）、TTL/淘汰/Stats/异步写+syncer、自绑配置 `agcache.ristretto.*`
- `agcache-fx-integration`: Fx group 集成——`FxAgCacheMode`（`group:"agcache.engine"` 收集→幂等注册→Manager→生命周期）、`FxAgCacheRistrettoMode`（Provide 工厂到 group）、core 绑 `agcache.defaultEngine`

### Modified Capabilities
<!-- 无：openspec/specs/ 尚无现有 spec -->

## Impact

- **新增目录/文件**：
  - `ag/ag_cache/`：cache.go / engine.go / config.go / manager.go / typed.go / loader_cache.go / serializer.go / mock.go / zfx_ag_cache.go / cache_test.go / manager_test.go / config_test.go / zfx_ag_cache_test.go
  - `ag/ag_cache/agristretto/`：ristretto.go / config_ristretto.go / zfx_agristretto.go / ristretto_test.go / config_ristretto_test.go / bench_test.go
- **依赖**：主 `go.mod` 新增 `github.com/dgraph-io/ristretto/v2 v2.4.2`；`golang.org/x/sync` 沿用现有 `v0.17.0`；`go.sum` 更新
- **无破坏性变更**：`ag_log`、`contribute/*`、`fxs/*`、`go.work`、`release_aif.sh` 均不改动
- **迁移来源**：`agcache-poc`（27 测试按"core mock / agristretto 真实 Ristretto"拆分迁移）
- **API 契约**：业务接口形态不变（`New[T](name, loader)` / `Get` / `Clear` / `Set(ttl...)` / `Peek`）；引擎实现者遵循 `Engine` SPI + `Create()` 无参 + 可选 `DefaultTTLProvider`
