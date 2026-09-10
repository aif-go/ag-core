## ADDED Requirements

### Requirement: key 前缀方案 A（对齐 Spring）
`typedCache` SHALL 记录 `name`（缓存名/namespace）与 `prefix`（`agcache::<name>::`）。
核心（typedCache）SHALL 统一拼装完整 key（`prefix + key`）后传给 `Engine`；`Engine` SHALL 无 name 概念（收到已拼前缀的 key）。

#### Scenario: 前缀拼装
- **WHEN** name="users"、业务 key="u:1" 调用 `Get(ctx, "u:1")`
- **THEN** engine 收到 key `"agcache::users::u:1"`

#### Scenario: singleflight key 拼前缀
- **WHEN** `GetOrElse(ctx, "u:1", loader)` 并发 miss
- **THEN** singleflight 的 key 用 `"agcache::users::u:1"`（避免跨 name 同 key 竞争）

#### Scenario: DelMany 前缀
- **WHEN** `Del(ctx, "a", "b")`（引擎实现 BulkDelEngine）
- **THEN** DelMany 收到 `"agcache::users::a"`、`"agcache::users::b"`

### Requirement: Engine.Clear 带 prefix
`Engine.Clear(ctx context.Context, prefix string)` SHALL 接收 namespace 前缀，供共享后端引擎（Redis）按前缀 SCAN+DEL 清 namespace。
local 引擎（agristretto）SHALL 忽略 prefix（清独立实例）。

#### Scenario: local Clear 忽略 prefix
- **WHEN** `typedCache.Clear(ctx)` 调用 `engine.Clear(ctx, "agcache::users::")`
- **THEN** agristretto 清空其独立实例，忽略 prefix

#### Scenario: 前缀格式
- **WHEN** 拼装 `prefix + key`
- **THEN** 结果为 `agcache::<name>::<key>`（双冒号分隔，业务 key 单冒号不歧义）
