# agcache-prod-error Specification

## Purpose
TBD - created by archiving change agcache-prod-hardening. Update Purpose after archive.

## Requirements

### Requirement: 入口返回 error（替代 panic）
`GetCache[T](m, name) (ICache[T], error)` SHALL 返回缓存实例与创建错误——引擎未注册或 `EngineFactory.Create(name)` 失败 SHALL 返回 error 而非 panic。
`GetCacheWithLoader[T](m, name, loader, opts...) (*LoaderCache[T], error)` SHALL 同理返回 error。
`getOrCreate` SHALL 内部无 panic：引擎工厂缺失与 Create 失败均经 error 返回。

#### Scenario: Create 失败返回 error 不 panic
- **WHEN** mock 工厂 `Create(name)` 返回错误，调用 `GetCache(m, "users")`
- **THEN** 返回 `(nil, error)`，不 panic

#### Scenario: 引擎未注册返回 error
- **WHEN** Manager 无默认引擎工厂（`defaultEngine` 未注册），调用 `GetCache(m, "users")`
- **THEN** 返回 `(nil, error)`（错误含引擎名），不 panic

#### Scenario: 正常获取无 error
- **WHEN** 引擎已注册且 Create 成功，调用 `GetCache(m, "users")`
- **THEN** 返回 `(缓存实例, nil)`；同 name 二次调用复用实例（err 恒 nil）

#### Scenario: GetCacheWithLoader 构造期 error
- **WHEN** fx 构造 `GetCacheWithLoader(m, "users", loader)` 返回 error
- **THEN** 构造失败传播至 fx（`fx.Provide` 返回 error → 装配失败，启动 fail-fast）

### Requirement: Ristretto 背压 drop 视为成功
`ristrettoEngine` 写路径（`Set`/`SetWithTTL`）SHALL 在 Ristretto `setBuf` 满（`SetWithTTL` 返回 false，新 key 背压丢弃）时返回 nil——视为可接受背压，等同容量淘汰，不包装 `ErrBackend`。
引擎 SHALL 维护 `dropped` 计数（原子），并导出 `Dropped() uint64` 供可观测。
真后端故障（引擎内部错误注入等）SHALL 仍返回 error（不吞）。

#### Scenario: buffer full 不误报 ErrBackend
- **WHEN** 高并发写触发 Ristretto `setBuf` 满，`Set` 或 `SetWithTTL` 被 drop
- **THEN** 引擎返回 nil（非 ErrBackend）；`Dropped()` 计数递增

#### Scenario: 真后端故障仍报错
- **WHEN** 引擎被注入真实后端错误（如 mock 引擎 `Err` 字段）
- **THEN** 返回 error（不被 drop-as-success 吞掉）

### Requirement: GetOrElse 写失败不影响返回值
`GetOrElse` SHALL 在 loader 成功加载值后，若缓存写失败（背压或后端写错误）仍返回 `(value, nil)`——读穿透以数据为准，缓存写是尽力而为；写失败仅记录日志/计数。
读路径真故障（engine.Get 返回非 ErrCacheMiss）SHALL 仍返回 ErrBackend 且不调 loader（防击穿语义不变）。

#### Scenario: loader 成功但缓存写失败返回数据
- **WHEN** loader 成功返回 v，但引擎 `Set` 返回错误（如写失败/背压）
- **THEN** `GetOrElse` 返回 `(v, nil)`（不丢弃已加载数据）

#### Scenario: 读故障仍不调 loader
- **WHEN** engine.Get 返回非 ErrCacheMiss 的后端故障
- **THEN** `GetOrElse` 返回 ErrBackend，不调 loader（防击穿不变）
