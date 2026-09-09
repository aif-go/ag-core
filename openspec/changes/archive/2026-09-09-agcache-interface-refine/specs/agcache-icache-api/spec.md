## ADDED Requirements

### Requirement: ICache 业务接口定稿
`ICache[T]` SHALL 提供 `Get`（严格读，miss→`ErrCacheMiss`）、`TryGet`（宽松读，miss→`(zero, false, nil)`）、`GetOrElse`（读穿透）、`Set(ctx, key, value)`（**无 ttl 变参**，TTL 归 per-namespace 配置）、`Del(ctx, keys...)`、`Clear(ctx)`。
`AdminCache` / `GetAdmin` / `Peek` / `SetWithTTL` SHALL 被移除（监控/探活用 `TryGet` 预期 Hit）。

#### Scenario: TryGet 宽松读
- **WHEN** 对未缓存 key 调用 `TryGet`
- **THEN** 返回 `(zero, false, nil)`（非 error）

#### Scenario: TryGet 命中
- **WHEN** 对已缓存 key 调用 `TryGet`
- **THEN** 返回 `(value, true, nil)`

#### Scenario: Set 无 ttl 变参
- **WHEN** 调用 `Set(ctx, key, value)`
- **THEN** 写入缓存，使用 per-namespace 默认 TTL（业务接口不暴露 ttl）

#### Scenario: Get 严格读语义不变
- **WHEN** 对未缓存 key 调用 `Get`
- **THEN** 返回 `ErrCacheMiss`（与 TryGet 区分）

### Requirement: 移除 Peek 与 AdminCache
`AdminCache[T]` 接口、`GetAdmin[T](name)` 包级函数、`Peek` 方法 SHALL 被移除。
统计（`Stats`）SHALL 不在业务接口暴露（后置）。

#### Scenario: 无 Peek/AdminCache 可用
- **WHEN** 编译引用 `GetAdmin` 或 `Peek`
- **THEN** 编译失败（已移除）

### Requirement: GetOrElse 健壮性
`GetOrElse` SHALL 的 singleflight double-check（engine.Get）使用 `context.WithoutCancel(ctx)`，使首个调用者 ctx 取消不影响等待者（Redis 引擎防击穿健壮性）。
`GetOrElse` 中缓存写入（engine.Set）失败 SHALL 包装为 `ErrBackend`（`errors.Is` 可判），不得返回原始错误。
`typedCache` 的 panic recovery SHALL 区分 panic 来源：loader panic 标注为 `loader panic`，引擎 panic 标注为 `engine panic`（不得统一误导为 engine panic）。

#### Scenario: double-check 用 WithoutCancel
- **WHEN** 首个进入 singleflight 的调用者 ctx 已取消，其余等待者正常
- **THEN** double-check 的 engine.Get 不受取消影响，等待者仍正常加载（loader 用 WithoutCancel 已保证）

#### Scenario: 写失败包 ErrBackend
- **WHEN** loader 成功但 engine.Set 失败（返回非 ErrCacheMiss 错误）
- **THEN** `GetOrElse` 返回 `ErrBackend`（`errors.Is(err, ErrBackend)==true`）

#### Scenario: panic 来源区分
- **WHEN** loader 内 panic
- **THEN** 返回错误含 `loader panic` 标注（非 `engine panic`）

### Requirement: 异步 Set 语义文档化
`Set` 对异步写引擎 SHALL 为"异步可见"语义（`engine.Set` 入队不 Wait）；`syncer`（`Sync()`）仅由 `GetOrElse` 在 miss-load 写入后调用保证可见。
业务"Set 后立即可读"不保证（微秒级异步窗口）；需立即一致的场景 SHALL 用 `GetOrElse`（loader 路径）或业务自行保证。
`WithDefaultTTL` SHALL 对 `ttl < 0` 校验并返回错误（防御负 TTL，ISSUE-P6 入口防护）。

#### Scenario: Set 异步语义
- **WHEN** 对异步写引擎 `Set(ctx,k,v)` 后立即 `Get(ctx,k)`
- **THEN** 可能返回 miss（微秒级窗口，文档化语义，不视为 bug）

#### Scenario: GetOrElse 写后可见
- **WHEN** `GetOrElse` 加载新值（引擎实现 syncer）
- **THEN** 返回前经 `Sync()` 保证立即可见

#### Scenario: 负 TTL 防御
- **WHEN** 构造期传 `WithDefaultTTL(-1)`
- **THEN** 返回错误（不静默丢弃）
