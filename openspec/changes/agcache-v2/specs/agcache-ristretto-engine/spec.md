## ADDED Requirements

### Requirement: local 引擎实现 Engine SPI
`ag/ag_cache/agristretto/ristretto.go` SHALL 实现 `Engine` SPI，底层为单个 Ristretto 实例（`ristretto.Cache[string, []byte]`），独立实例、无共享状态。
`Get` SHALL 命中返回 `(data, nil)`，未命中返回 `(nil, ErrCacheMiss)`。
`Set` SHALL 将 cost 内部化为 `len(value)`（SPI 不暴露 cost 概念），异步写入；ttl=0 永不过期，>0 过期。
Ristretto 实例 SHALL 开启 Metrics（`Config.Metrics: true`），否则 Stats 全 0。
`NumCounters` 未配置（0）SHALL 按 `MaxCost * 10%` 推导；`MaxCost` 未配置（<=0）SHALL 用默认值 100MB。

#### Scenario: local 引擎 miss 语义
- **WHEN** 对未写入 key 调用 `Get`
- **THEN** 返回 `(nil, ErrCacheMiss)`

#### Scenario: 默认配置推导
- **WHEN** 以 `RistrettoConfig{MaxCost:0, NumCounters:0}` 创建引擎
- **THEN** 使用默认 100MB，NumCounters 按 MaxCost 推导

### Requirement: agristrettoFactory
`agristrettoFactory` SHALL 实现 `ag_cache.EngineFactory`：`Name()` 返回 `"ristretto"`，`Create()` 无参返回 `ag_cache.Engine`。
工厂 SHALL 由引擎配置（`RistrettoConfig`）与默认 TTL（`time.Duration`）构造（配置自持，不接收外部参数）。
工厂 SHALL 实现 `ag_cache.DefaultTTLProvider`，`DefaultTTL()` 返回构造时携带的默认 TTL。

#### Scenario: 工厂注册与 Create
- **WHEN** 将 `agristrettoFactory` 注册后调用 `Create()`
- **THEN** 返回可用的 Ristretto 引擎，无需传入配置

#### Scenario: DefaultTTLProvider
- **WHEN** 工厂被断言为 `ag_cache.DefaultTTLProvider`
- **THEN** `DefaultTTL()` 返回构造时设置的默认 TTL

### Requirement: TTL 语义
引擎 SHALL 遵循 ttl 语义：`ttl=0` 永不过期，`ttl>0` 到期后 `Get` 返回 `ErrCacheMiss`。

#### Scenario: ttl=0 永不过期
- **WHEN** 以 `Set(..., 0)` 写入后等待并多次 `Get`
- **THEN** 始终命中

#### Scenario: 正 ttl 到期后 miss
- **WHEN** 以 `Set(..., 短 ttl)` 写入并等待 ttl 过期
- **THEN** `Get` 返回 `ErrCacheMiss`

### Requirement: 异步写与 syncer 可见性
引擎 SHALL 实现可选 `syncer` 接口（`Sync()`，调用 Ristretto `Wait()`）；`GetOrElse` 在 miss-load 写入后 SHALL 通过 `syncer` 保证已加载值对后续读可见。

#### Scenario: GetOrElse 写后立即可读
- **WHEN** `GetOrElse` 加载新值返回后立即再次 `Get` 同一 key
- **THEN** 命中且值一致（无异步写竞态）

### Requirement: 引擎 Stats
`Stats()` SHALL 返回 `Stats{Hits, Misses, Evictions, EntryCount}`，来自 Ristretto Metrics。

#### Scenario: 每 namespace 独立统计
- **WHEN** 两个独立引擎实例分别读写后各自 `Stats()`
- **THEN** 各实例计数独立，一个实例的 miss 不污染另一个的 Hits

### Requirement: 淘汰
当写入累计 cost 超过 `MaxCost` 时引擎 SHALL 按 Ristretto 策略淘汰条目，被淘汰 key `Get` 返回 `ErrCacheMiss`，`Stats.Evictions` 递增。

#### Scenario: 超过内存预算触发淘汰
- **WHEN** 以极小 `MaxCost` 写入超出预算的多个 key
- **THEN** 部分 key 被淘汰，`Get` 返回 `ErrCacheMiss`

### Requirement: 引擎配置绑定
`RistrettoConfigProperties` SHALL 为 `agcache.ristretto` 前缀的绑定模型，含 `MaxCost int64`、`NumCounters int64`、`DefaultTTL string`（""→不设默认，core 兜底 5min；"0"→永不过期；"60s"→显式）。
`DefaultRistrettoConfigProperties()` SHALL 返回带默认值的配置（`MaxCost=100MB`）。
`BindRistrettoConfigProperties(binder)` SHALL 基于默认配置执行绑定，仅覆盖 YAML 出现的键。
`parseTTL` SHALL 将 `DefaultTTL` 字符串转换为 `time.Duration`（""→0；"0"→0 永不过期由引擎语义保证；非法字符串返回错误）。

#### Scenario: 配置绑定
- **WHEN** YAML 提供 `agcache.ristretto.maxCost` 与 `agcache.ristretto.defaultTtl`
- **THEN** 绑定结果 MaxCost/DefaultTTL 与 YAML 一致

#### Scenario: 未配置字段保持默认
- **WHEN** YAML 缺失 `numCounters` 或 `defaultTtl`
- **THEN** `NumCounters` 保持 0（引擎推导），`DefaultTTL` 保持默认（""→core 兜底 5min）

#### Scenario: 非法 TTL 字符串报错
- **WHEN** `DefaultTTL` 为 `"abc"`
- **THEN** 绑定/解析返回错误
