# AgCache 接口定稿升级设计

## Context

- **现状**：AgCache v2 已实现并提交（`agcache-v2` change，`0ba6a87`）。集成测试（`ag/ag_cache/test/`）与 SPI 审计（`/home/houzw/document/hzw-obsidian/ag-core/ag_cache/2026-08-24-ag-cache-spi-audit.md`）暴露接口问题：`AdminCache`/`GetAdmin`/`Peek`/`Stats` 按 Ristretto 特化；`ICache.Set(ttl...)` 变参不"基础"。
- **目标**：按讨论定稿（v3，Spring Cache 同构）实施接口升级。
- **参考**：SPI 审计文档接口定稿章节、issues 文档、Spring Cache TTL 三层抽象。

## Goals / Non-Goals

**Goals:**
- ICache 更基础：`Set` 去 ttl 变参、新增 `TryGet`、移除 `AdminCache`/`GetAdmin`/`Peek`/`SetWithTTL`
- Engine SPI 定稿：`Set` 保留 ttl（执行层）、移除 `Stats`、新增可选 `BulkDelEngine`
- TTL 三层设计落地（Spring Cache 同构）
- agristretto 适配、测试同步、文档 11 更新

**Non-Goals:**
- Redis 引擎（方向 3，另行设计）
- Stats/Prometheus 指标（后置）
- 2.4 负 TTL 防御（已并入 TTL 设计，见 D5）

## Decisions

### D1. ICache 更基础（业务接口）
```go
type ICache[T any] interface {
    Get(ctx context.Context, key string) (T, error)
    TryGet(ctx context.Context, key string) (T, bool, error)
    GetOrElse(ctx context.Context, key string, loader LoaderFunc[T]) (T, error)
    Set(ctx context.Context, key string, value T) error
    Del(ctx context.Context, keys ...string) error
    Clear(ctx context.Context) error
}
```
- `Set` 去 ttl 变参（业务不操心 TTL）
- `TryGet` 宽松读（miss→`(zero,false,nil)`），替代 `Peek` 用途，监控用 TryGet 预期 Hit
- 移除 `AdminCache`/`GetAdmin`/`Peek`/`SetWithTTL`（YAGNI）
**理由**：业务接口最小化 + Spring `Cache.put` 无 ttl 同构。

### D2. Engine 执行层保留 ttl
```go
type Engine interface {
    Get(ctx, key) ([]byte, error)
    Set(ctx, key, value []byte, ttl time.Duration) error   // 执行层带 ttl
    Del(ctx, key) error
    Clear(ctx) error
    Close() error
}
```
- ttl 是通用传递参数（`0`=永不过期基线），支持 TTL 引擎执行、无 TTL 引擎忽略
- **理由**：Spring `RedisCache` 内部 SETEX 带 expire 同构；保证 `WithDefaultTTL` 经 typedCache 传递生效（避免"Engine.Set 去 ttl 切断 TTL 传递"）

### D3. 移除 Stats（后置）
- `Engine` 移除 `Stats()`；`StatsProvider`/`LogStats`/`Manager.Visit` 一并移除
- OnStop 只 `Close`；agristretto 可关 Ristretto Metrics（`Config.Metrics: false`，省开销）
- **理由**：Stats 是 Ristretto Metrics 特化，非通用契约；监控用 TryGet 预期 Hit，无运行期统计消费点

### D4. BulkDelEngine 可选接口
```go
type BulkDelEngine interface { DelMany(ctx context.Context, keys ...string) error }
// typedCache.Del: if b, ok := engine.(BulkDelEngine); ok { b.DelMany(...) } else { 循环 Del }
```
- 不改基础 `Engine.Del`；Redis 引擎实现 DelMany 用 DEL 一把梭；agristretto 不实现走循环
- **理由**：1.1 结论，避免破坏性 SPI 变更

### D5. TTL 三层 + 优先级
- 业务接口无 ttl / per-namespace 配置（`WithDefaultTTL` + `DefaultTTLProvider`）/ 执行层带 ttl
- 优先级：`WithDefaultTTL` > 引擎 `DefaultTTLProvider` > 兜底 5min
- `WithDefaultTTL` 对 `ttl<0` 校验返回错误（ISSUE-P6 入口防护）

### D6. 健壮性问题修复（方向 2 并入本 change）
用户要求把简单健壮性问题直接在本 change 修复，不另开任务：
- **2.1 double-check 用 `WithoutCancel`**：`GetOrElse` 的 singleflight 内 `engine.Get` 改用 `context.WithoutCancel(ctx)`（与 loader 一致）——首调用者 ctx 取消不连累等待者，Redis 引擎防击穿健壮性
- **2.2 写失败包 `ErrBackend`**：`GetOrElse` 中 `engine.Set` 失败 `return result{v, serr}, errBackend(serr)`（而非原始错误）
- **2.3 panic 来源区分**：`recoverPanic` 区分 loader panic（`loader panic`）与引擎 panic（`engine panic`）
- **2.5 异步 Set 文档化**：`Set` 为异步可见语义（不保证 Set 后立即可读），`syncer` 仅 `GetOrElse` 写后同步；候选方案选 **C（文档化 + 引导用 GetOrElse）**，暂不加显式同步 API（YAGNI，未来需要再加）
- **2.4/P6 负 TTL 防御**：`WithDefaultTTL` 对 `ttl<0` 返回错误（见 D5）
**理由**：均为引擎无关的小改动，随接口定稿一并交付，避免另开 change 的成本。

## 文件与函数签名清单

### `ag/ag_cache/cache.go`
- `ICache[T]`：`Get`/`TryGet`/`GetOrElse`/`Set(ctx,key,value)`/`Del(keys...)`/`Clear`
- 删除 `AdminCache[T]`
- 错误哨兵 `ErrCacheMiss`/`ErrBackend` 不变

### `ag/ag_cache/engine.go`
- `Engine`：`Get`/`Set(ctx,key,value,ttl)`/`Del`/`Clear`/`Close`（删 `Stats`）
- 新增 `type BulkDelEngine interface { DelMany(ctx, keys ...string) error }`
- 保留 `DefaultTTLProvider` / `syncer` / `RegisterEngine` / `EngineRegistered` / `getFactory` / `errBackend`

### `ag/ag_cache/typed.go`
- `Set`：去 ttl 变参 → `func (c *typedCache[T]) Set(ctx, key, value T) error`，内部 `engine.Set(ctx, key, data, c.defaultTTL)`
- 新增 `TryGet(ctx, key) (T, bool, error)`（调 `engine.Get`，miss→`(zero,false,nil)`，后端错→`(zero,false,errBackend)`）
- `Del`：探测 `BulkDelEngine`
- 删除 `Peek`/`Stats`/`stats()`；保留 `closeEngine`/`unmarshal`/`recoverPanic`
- `recoverPanic` 区分 loader/engine panic 来源（2.3）：loader panic → `loader panic: ...`，引擎 panic → `engine panic: ...`
- `GetOrElse`：double-check 用 `context.WithoutCancel(ctx)`（2.1）；`engine.Set` 失败 `errBackend(serr)`（2.2）

### `ag/ag_cache/manager.go`
- 删除 `GetAdmin[T]` / `Visit`
- 保留 `NewManager`/`Close`/`SetDefault`/`New`/`Get`/`CloseAll`/`getOrCreate`
- `getOrCreate` TTL 计算：`WithDefaultTTL`（Option）> 工厂 `DefaultTTLProvider` > 5min；`WithDefaultTTL` 校验 ttl<0

### `ag/ag_cache/zfx_ag_cache.go`
- 删除 `LogStats`；`registerHooks` OnStop 只 `m.Close()`
- 保留 `EngineFactoryParams`（group）/`NewAgCacheManager`/`registerHooks`/`FxAgCacheMode`

### `ag/ag_cache/agristretto/ristretto.go`
- `Set` 保留 ttl → `e.cache.SetWithTTL(key, value, cost, ttl)`
- `Config.Metrics: false`（无 Stats 消费）
- 不实现 `BulkDelEngine`（循环删）；实现 `DefaultTTLProvider`（`DefaultTTL()` 返回工厂 ttl）

### 测试同步
- `ag/ag_cache/cache_test.go`/`manager_test.go`/`zfx_ag_cache_test.go`：`Set` 去变参、`TryGet` 用例、移除 `GetAdmin`/`Stats` 相关
- `ag/ag_cache/test/`：`waitAdmin`/`GetAdmin` 替换为 `TryGet`；移除 Peek/Stats 探测
- 新增 `TryGet` 行为测试、`BulkDelEngine` 探测测试、负 TTL 防御测试

## 测试策略

| 层 | 文件 | 覆盖 |
|----|------|------|
| core 单元 | `cache_test.go` | TryGet 宽松读/命中/后端错、Set 无变参、Del 批量探测（mock BulkDelEngine） |
| core Manager | `manager_test.go` | 移除 GetAdmin/Visit 后 New/Get/CloseAll；TTL 优先级（WithDefaultTTL>引擎>5min） |
| core 配置/Fx | `zfx_ag_cache_test.go` | OnStop 只 Close（无 LogStats）；group 装配不变 |
| agristretto | `ristretto_test.go` | SetWithTTL 保留 ttl；Metrics 关；DefaultTTLProvider |
| 集成 | `test/` | TryGet 替代 Peek；Set 无变参；监控用 TryGet 预期 Hit |

验证：`go test -race ./ag/ag_cache/...`、`go vet ./ag/ag_cache/...`、`go build ./...`。

## Risks / Trade-offs

- [R1] **破坏性变更**：`Set(ttl...)` 变参移除、`AdminCache`/`GetAdmin`/`Peek`/`Stats` 移除——现有调用方与开发文档 11 需同步更新。
- [R2] **TryGet 语义**：宽松读不区分"miss"与"后端故障"的 ok 位（后端故障也返回 ok=false + err），业务需检查 err。
- [R3] **无 TTL 引擎**：`WithDefaultTTL` 对其静默不生效（引擎忽略），文档化引擎能力差异。
- [R4] **2.1 double-check ctx**：改 `WithoutCancel` 后，double-check 不再感知首调用者取消——正确（防击穿），但 Redis 引擎的 double-check 查询不再被首调用者 ctx 约束超时（可接受，loader 本就用 WithoutCancel）。
- [R5] **2.5 异步 Set**：选方案 C（文档化），"Set 后立即可读"对异步写引擎不保证——若未来出现高频"Set 后即读"业务，需再评估显式同步 API（方案 B）。

## Migration Plan

- 接口变更同步开发文档 11（`Set` 无 ttl、`TryGet` 用法、移除 `GetAdmin`/`Peek`）
- `ag/ag_cache/test/` 与核心测试同步适配
- 回滚：未合入前丢弃；合入后为破坏性 API 变更，需一次性迁移调用方

## Open Questions

- 无（方向 2 健壮性问题已并入本 change；方向 3 Redis 引擎另行设计）
