## ADDED Requirements

### Requirement: 定版入口 GetCacheWithLoader / GetCache
`GetCacheWithLoader[T](m *Manager, name string, loader LoaderFunc[T], opts ...Option[T]) *LoaderCache[T]` SHALL 从显式 Manager 获取（不存在则创建）缓存实例并绑定 loader（读穿透）。
`GetCache[T](m *Manager, name string) ICache[T]` SHALL 从显式 Manager 获取缓存实例（纯读，无 loader）。
包级 `New` / `Get`（走 defaultManager 的 v3 版本）SHALL 被移除，改为显式 `m` 参数。
`DefaultManager() *Manager` SHALL 返回当前 defaultManager（`SetDefault` 设置的），未设置时返回 nil。
`SetDefault` / `CloseAll` / `WithLoader` / `NewWithEngine` SHALL 保留。

#### Scenario: GetCacheWithLoader 读穿透
- **WHEN** `GetCacheWithLoader(m, "users", repo.GetUser)` 后 `Get(ctx, "u:1")` 且缓存 miss
- **THEN** 调用 loader（repo.GetUser）并写入缓存，返回 `*User`

#### Scenario: GetCache 纯读
- **WHEN** `GetCache(m, "users")` 后 `Get(ctx, "u:1")` 且缓存 miss
- **THEN** 返回 `ErrCacheMiss`（不调 loader）

#### Scenario: DefaultManager 获取
- **WHEN** `SetDefault(mgr)` 后调用 `DefaultManager()`
- **THEN** 返回 mgr；未 `SetDefault` 时返回 nil

#### Scenario: 构造时绑定一次
- **WHEN** 业务构造 `GetCacheWithLoader(m, "users", loader)` 存入字段，方法内复用 `s.users.Get`
- **THEN** 不每次方法调用重新创建（同 name 复用底层实例）

### Requirement: fx 时序保证
业务构造依赖 `*Manager`（fx 注入）SHALL 保证 Manager 先于业务构造创建（fx 依赖拓扑）。
`GetCacheWithLoaderDefault` 便捷版 SHALL 不提供；运行时自定义场景用 `DefaultManager()` 获取 m 再调 `GetCacheWithLoader`。

#### Scenario: fx 装配时序
- **WHEN** fx 中业务 Provide 构造依赖 `*Manager`
- **THEN** fx 保证 Manager 先创建，`GetCacheWithLoader(m, ...)` 安全调用

#### Scenario: 运行时 default 使用
- **WHEN** 非 Fx 场景 `SetDefault(mgr)` 后运行时 `DefaultManager()` 获取 m
- **THEN** 可 `GetCacheWithLoader(m, "users", loader)` 使用

### Requirement: ICache.SetWithTTL（单条显式 TTL）
`ICache[T]` SHALL 提供 `SetWithTTL(ctx, key, value, ttl)`——单条显式 TTL 写入，优先级最高。
`Set(ctx, key, value)` SHALL 用 namespace 默认 TTL（`WithDefaultTTL` > 引擎内部默认）。
`ttl=0` SHALL 表示永不过期。

#### Scenario: SetWithTTL 显式 TTL
- **WHEN** `SetWithTTL(ctx, "u:1", v, 60*time.Second)` 后 `Get(ctx, "u:1")`
- **THEN** 命中，条目 TTL 为显式指定的 60s（非默认 TTL）

#### Scenario: Set 用默认 TTL
- **WHEN** `Set(ctx, "u:1", v)`（未配 WithDefaultTTL）
- **THEN** 用引擎内部默认 TTL（引擎配置/实现决定）

#### Scenario: ttl=0 永不过期
- **WHEN** `SetWithTTL(ctx, "u:1", v, 0)` 后等待超过默认 TTL 时间
- **THEN** 条目仍命中（永不过期）

### Requirement: WithDefaultTTL（业务 per-cache 默认 TTL）
`WithDefaultTTL[T](ttl) Option[T]` SHALL 设置 namespace 默认 TTL（per-cache 业务默认）。
配了 `WithDefaultTTL` 时 `Set` SHALL 经引擎 `TTLSetter` 用默认 TTL；未配时 `Set` SHALL 走 `engine.Set`（引擎内部默认）。
`ttl=0` SHALL 表示永不过期。

#### Scenario: WithDefaultTTL 生效
- **WHEN** `GetCacheWithLoader(m, "users", loader, WithDefaultTTL(30*time.Second))` 后 `Set(ctx, "u:1", v)`
- **THEN** 条目 TTL 为 30s（业务默认，经引擎 TTLSetter）

#### Scenario: 未配 WithDefaultTTL
- **WHEN** `GetCacheWithLoader(m, "users", loader)` 后 `Set(ctx, "u:1", v)`
- **THEN** 用引擎内部默认 TTL（`engine.Set`，业务未配置时引擎决定）

#### Scenario: TTL 优先级链
- **WHEN** 同时存在 `WithDefaultTTL(30s)` 与单条 `SetWithTTL(ctx, "u:1", v, 60s)`
- **THEN** 该条目 TTL 为 60s（`SetWithTTL` > `WithDefaultTTL` > 引擎内部默认）
