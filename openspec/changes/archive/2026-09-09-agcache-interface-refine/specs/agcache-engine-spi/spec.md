## ADDED Requirements

### Requirement: Engine SPI 定稿
`Engine` SHALL 提供 `Get(ctx, key)`、`Set(ctx, key, value, ttl)`（**保留 ttl 参数**，执行层）、`Del(ctx, key)`、`Clear(ctx)`、`Close()`。
`Engine` SHALL **不包含** `Stats()`（统计后置，`StatsProvider` 暂不保留）。

#### Scenario: Engine.Set 带 ttl
- **WHEN** 调用 `engine.Set(ctx, key, value, ttl)`
- **THEN** 支持 TTL 的引擎做 `SetWithTTL`/`SETEX`；无 TTL 引擎忽略（= 永不过期）

#### Scenario: Engine 无 Stats
- **WHEN** 编译引用 `engine.Stats()`
- **THEN** 编译失败（Stats 已从基础接口移除）

### Requirement: BulkDelEngine 可选接口
`BulkDelEngine` SHALL 为可选接口：`DelMany(ctx, keys ...string) error`。
`typedCache.Del(ctx, keys...)` SHALL 探测 `BulkDelEngine`，实现则用 `DelMany`（一次批量），否则循环 `Del`。
`Engine` 基础接口的 `Del` SHALL 保持单 key（不变）。

#### Scenario: 引擎实现 BulkDelEngine 走批量
- **WHEN** 引擎实现 `BulkDelEngine` 且调用 `Del(ctx, "a","b","c")`
- **THEN** 调用 `DelMany(ctx, "a","b","c")`（一次）

#### Scenario: 引擎未实现 BulkDelEngine 走循环
- **WHEN** 引擎不实现 `BulkDelEngine` 且调用 `Del(ctx, "a","b","c")`
- **THEN** 循环调用 `Del` 三次（行为正确，仅非批量）

### Requirement: 保留可选接口
`DefaultTTLProvider`（`DefaultTTL() time.Duration`）SHALL 保留，供引擎声明默认 TTL；无 TTL 引擎不实现。
`syncer`（`Sync()`）SHALL 保留，供异步写引擎在 `GetOrElse` miss-load 后同步可见。

#### Scenario: 引擎声明默认 TTL
- **WHEN** 引擎实现 `DefaultTTLProvider` 且业务未传 `WithDefaultTTL`
- **THEN** typedCache 默认 TTL 为引擎声明的值

#### Scenario: 无 TTL 引擎
- **WHEN** 引擎不实现 `DefaultTTLProvider`
- **THEN** typedCache 兜底 5min；引擎实际忽略 TTL（永不过期，靠 Del/Clear）
