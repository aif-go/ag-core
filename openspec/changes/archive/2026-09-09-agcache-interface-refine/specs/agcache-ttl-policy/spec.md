## ADDED Requirements

### Requirement: TTL 三层设计（Spring Cache 同构）
TTL SHALL 分三层：
- 业务接口层：`ICache.Set(ctx, key, value)` **无 ttl**（对应 Spring `Cache.put` 无 ttl）
- per-namespace 配置层：`WithDefaultTTL(ttl)`（构造期 Option）与引擎 `DefaultTTLProvider` 声明默认 TTL（对应 Spring `CacheManager` 的 `entryTtl`）
- 执行层：`Engine.Set(ctx, key, value, ttl)` **带 ttl**（对应 Spring `RedisCache` 内部 SETEX 带 expire）

`Engine.Set` 的 ttl SHALL 为通用传递参数（`ttl=0` 永不过期基线）；支持 TTL 的引擎执行，无 TTL 引擎忽略。

#### Scenario: 业务接口无 ttl
- **WHEN** 业务调用 `Set(ctx, key, value)`
- **THEN** 不传 ttl，使用 per-namespace 默认 TTL（WithDefaultTTL 或引擎 DefaultTTLProvider 或兜底 5min）

#### Scenario: per-namespace 覆盖
- **WHEN** 构造期传 `WithDefaultTTL(30s)`（params）且引擎支持 TTL
- **THEN** params 缓存默认 TTL 为 30s（与 users 默认不同）

#### Scenario: 执行层带 ttl
- **WHEN** typedCache 计算最终 ttl 后调用 `engine.Set(ctx, key, value, ttl)`
- **THEN** 支持 TTL 的引擎按 ttl 设置过期（SetWithTTL/SETEX）

### Requirement: TTL 优先级
最终 ttl SHALL 按 `WithDefaultTTL`（构造期）> 引擎 `DefaultTTLProvider` > 兜底 5min 计算。
`WithDefaultTTL` SHALL 对 `ttl < 0` 校验并返回错误（防御负 TTL，ISSUE-P6 入口防护）。

#### Scenario: 优先级
- **WHEN** 业务传 `WithDefaultTTL(30s)` 且引擎声明 60s
- **THEN** 最终默认 TTL 为 30s（Option 最高）

#### Scenario: 负 TTL 防御
- **WHEN** 构造期传 `WithDefaultTTL(-1)`
- **THEN** 返回错误（不静默丢弃）
