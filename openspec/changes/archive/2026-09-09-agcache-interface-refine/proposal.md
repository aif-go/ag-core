# AgCache 接口定稿升级（v3）

## Why

AgCache v2 已实现并提交（`agcache-v2`，core SPI + agristretto 引擎 + fx group 注入）。经集成测试（`ag/ag_cache/test/`）与以 Redis 为评估场景的 SPI 审计，发现接口存在两类问题：① **抽象 API 非通用**——`AdminCache`/`GetAdmin`/`Peek`/`Stats` 均按 Ristretto 本地缓存特化，Stats 是 Metrics 直接映射、Peek 更新统计、Engine 无批量删；② **ICache 不够基础**——`Set(ttl...)` 变参让开发人员承担 TTL 心智。经讨论确认采用 Spring Cache 同构的分层 TTL 设计，需将接口定稿（v3）落地实施。

## What Changes

- **`ICache` 业务接口更基础**：
  - `Set(ctx, key, value)` **去掉 ttl 变参**（TTL 归 per-namespace 配置 + 执行层）
  - **新增 `TryGet(ctx, key) (T, bool, error)`** 宽松读（miss → `(zero, false, nil)`），满足"不返回异常的存在性检查"，替代 Peek 用途
  - 移除 `AdminCache` / `GetAdmin` / `Peek` / `SetWithTTL`（监控用 `TryGet` 预期 Hit）
- **`Engine` SPI 定稿（执行层）**：
  - `Set(ctx, key, value, ttl)` **保留 ttl 参数**（执行层传递，无 TTL 引擎忽略）
  - **移除 `Stats()`**（Stats 整体后置，`StatsProvider` 暂不保留）
  - **新增可选接口 `BulkDelEngine{ DelMany(ctx, keys...) }`**（批量删，Redis 用 DEL 一把梭）
  - 保留可选 `DefaultTTLProvider` / `syncer`
- **TTL 三层设计（Spring Cache 同构）**：业务接口无 ttl / per-namespace 配置（`WithDefaultTTL` + `DefaultTTLProvider`）/ 执行层带 ttl
- **移除统计链路**：`LogStats` / `Manager.Visit`（OnStop 只 Close）
- **包级 API**：`New[T](name, loader, opts...)` / `Get[T](name)` / `SetDefault` / `CloseAll`；移除 `GetAdmin`
- **agristretto 适配**：`Set` 保留 ttl（`SetWithTTL`）；不实现 `BulkDelEngine`（循环删）；可关 Ristretto Metrics（省开销）
- **健壮性问题修复（方向 2，并入本 change）**：
  - 2.1 `GetOrElse` double-check 用 `context.WithoutCancel`（首调用者取消不连累等待者，Redis 防击穿健壮性）
  - 2.2 `GetOrElse` 缓存写入失败包装 `ErrBackend`（`errors.Is` 降级判断生效）
  - 2.3 loader panic 与引擎 panic **区分标注**（`loader panic` / `engine panic`，排查方向正确）
  - 2.5 异步 `Set` 可见性：**文档化"Set 异步"语义** + 保留 `syncer`（`GetOrElse` 写后同步）；`WithDefaultTTL` 对 `ttl<0` 校验防御（2.4/P6）

## Capabilities

### New Capabilities
- `agcache-icache-api`: ICache 业务接口定稿——`TryGet` 新增、`Set` 去 ttl 变参、移除 `AdminCache`/`Peek`/`SetWithTTL`/`GetAdmin`，监控用 TryGet 预期 Hit；**含 GetOrElse 健壮性**（2.1 WithoutCancel、2.2 写失败 ErrBackend、2.3 panic 区分标注、2.5 异步 Set 文档化）
- `agcache-engine-spi`: Engine SPI 定稿——`Set` 保留 ttl（执行层）、移除 `Stats`（后置）、可选 `BulkDelEngine`（批量删）、保留 `DefaultTTLProvider`/`syncer`
- `agcache-ttl-policy`: TTL 三层设计——业务接口无 ttl、per-namespace 配置（`WithDefaultTTL`+`DefaultTTLProvider`）、执行层带 ttl；与 Spring Cache 同构

### Modified Capabilities
<!-- 无：openspec/specs/ 尚无现有 spec -->

## Impact

- **修改文件**：
  - `ag/ag_cache/cache.go`：ICache（去 ttl 变参、新增 TryGet）；删除 AdminCache
  - `ag/ag_cache/engine.go`：Engine（删 Stats）；新增 BulkDelEngine
  - `ag/ag_cache/typed.go`：Set 无变参、新增 TryGet、Del 探测 BulkDelEngine、recoverPanic 区分 loader/engine
  - `ag/ag_cache/manager.go`：删 GetAdmin/Visit；SetDefault 等保留
  - `ag/ag_cache/zfx_ag_cache.go`：删 LogStats（OnStop 只 Close）
  - `ag/ag_cache/agristretto/ristretto.go`：Set 保留 ttl；可关 Metrics；不实现 BulkDelEngine
- **删除**：`AdminCache`、`GetAdmin`、`Peek`、`SetWithTTL`、`Stats`、`LogStats`、`Manager.Visit`、`StatsProvider`（未实现）
- **测试**：`ag/ag_cache/test/` 适配新接口（`Set` 无变参、`TryGet` 替代 Peek 探测、移除 GetAdmin 用例）
- **无破坏性变更**（对 `ag_log`/`contribute/*`/`fxs/*`/`go.work`/`release_aif.sh` 零改动）
- **破坏性影响（相对 agcache-v2 已实现）**：`ICache.Set(ttl...)` 变参移除、`AdminCache`/`GetAdmin`/`Peek`/`Stats` 移除——影响现有调用方，需同步开发文档 11
