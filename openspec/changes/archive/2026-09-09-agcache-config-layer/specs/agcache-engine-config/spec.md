## ADDED Requirements

### Requirement: 两层配置模型（RistrettoConfig 绑定层 + RistrettoOptions 运行时层）
agristretto 引擎 SHALL 提供两层配置：
- `RistrettoConfig`（绑定层）：YAML 绑定模型，含 `MaxCost`/`NumCounters`/`BufferItems`/`DefaultTTL string`。
- `RistrettoOptions`（运行时层）：实际创建 Ristretto 实例的参数，含 `MaxCost`/`NumCounters`/`BufferItems`/`DefaultTTL time.Duration`。
`RistrettoConfigProperties` SHALL 被删除。

#### Scenario: 绑定层与运行时层字段对应
- **WHEN** `RistrettoConfig{MaxCost:100, NumCounters:10, BufferItems:64, DefaultTTL:"60s"}` 经 `ToOptions()`
- **THEN** 得到 `RistrettoOptions{MaxCost:100, NumCounters:10, BufferItems:64, DefaultTTL:60*time.Second}`

### Requirement: DefaultRistrettoConfig 显式默认值
`DefaultRistrettoConfig()` SHALL 显式返回全部字段的默认值：`MaxCost=100_000_000`（100MB）、`NumCounters=131_072`（2^17，100K 档）、`BufferItems=64`、`DefaultTTL="0"`（显式永不过期）。
默认 NumCounters 取小（100K 档，sketch ~0.7MB/实例）——多数缓存 name（参数/配置/枚举）key 有限永不淘汰，小 NumCounters 足够且省预分配内存；NumCounters 只影响淘汰精度、不影响容量与正确性。

#### Scenario: 默认值完整可读
- **WHEN** 调用 `DefaultRistrettoConfig()`
- **THEN** 四字段均有显式默认值（MaxCost=100MB, NumCounters=131072, BufferItems=64, DefaultTTL="0"），且 YAML 缺省绑定时以此为起点

### Requirement: RistrettoConfig.ToOptions 转换与校验
`RistrettoConfig.ToOptions() (RistrettoOptions, error)` SHALL：
- 解析 `DefaultTTL string` 为 `time.Duration`（`parseTTL`：`""`/`"0"`→0 永不过期、`"60s"`→60s、非法字符串→error）。
- 对负值（MaxCost/NumCounters/BufferItems < 0）返回 error（fail-fast）。
- 默认填充（`<=0` → 固定默认，不做 MaxCost×10% 推导）：`MaxCost<=0`→100MB；`NumCounters<=0`→131072；`BufferItems<=0`→64。

#### Scenario: 非法 TTL 字符串报错
- **WHEN** `RistrettoConfig{DefaultTTL:"abc"}.ToOptions()`
- **THEN** 返回 error（非"60s"单位格式不被接受）

#### Scenario: 纯数字 TTL 报错（防纳秒陷阱）
- **WHEN** `RistrettoConfig{DefaultTTL:"60"}.ToOptions()`
- **THEN** 返回 error（强制带单位，避免被误解析为纳秒）

#### Scenario: 负值字段报错
- **WHEN** `RistrettoConfig{MaxCost:-1}.ToOptions()`
- **THEN** 返回 error

#### Scenario: 零值默认填充
- **WHEN** `RistrettoConfig{MaxCost:200_000_000}.ToOptions()`（NumCounters/BufferItems 为零）
- **THEN** MaxCost=200MB 保留，NumCounters=131072（固定默认，不随 MaxCost 推导），BufferItems=64

### Requirement: RistrettoOptions.Validate 导出校验
`RistrettoOptions.Validate() error` SHALL 校验非负性（MaxCost/NumCounters/BufferItems >= 0），供手动构造 Options 的场景显式校验。
`NewRistrettoEngine(opts RistrettoOptions)` SHALL 内部先调 `Validate()`（负值报错），再做零值默认兜底（0→默认/推导），创建引擎。

#### Scenario: 手动 Options 负值报错
- **WHEN** `NewRistrettoEngine(RistrettoOptions{MaxCost:-1})`
- **THEN** 返回 error（不经 Config 绑定层的场景也覆盖）

#### Scenario: 手动 Options 零值兜底
- **WHEN** `NewRistrettoEngine(RistrettoOptions{MaxCost:1<<20})`（NumCounters/BufferItems 为零）
- **THEN** 引擎创建成功，NumCounters 推导、BufferItems 用默认 64

### Requirement: 构造器职责分离 + 启动期预解析（Bind + Factory）
`BindRistrettoConfig(binder) (*RistrettoConfigs, error)` SHALL 仅负责从 binder 绑定 `agcache.ristretto.*` 容器（Default 以默认值为起点，Namespaces 缺省为空 map）。
`NewAgristrettoFactory(cfg *RistrettoConfigs) (ag_cache.EngineFactory, error)` SHALL 在装配期预解析：对 Default 与**所有** Namespaces 条目逐一 `ToOptions()` + `Validate()`，任一非法即返回 error（含具体 name 定位）；结果缓存进 `map[string]RistrettoOptions`。
`FxAgCacheRistrettoMode` SHALL 通过 `fx.Provide(BindRistrettoConfig, NewAgristrettoFactory→group:"agcache.engine")` 由依赖图自动装配；`ProvideAgristrettoFactory` SHALL 被删除。

#### Scenario: 工厂预解析 Default + Namespaces
- **WHEN** `NewAgristrettoFactory(&RistrettoConfigs{Default:{DefaultTTL:"60s"}, Namespaces:{"users":{MaxCost:500_000_000}}})` 成功
- **THEN** 工厂预解析 opts map 含 ""（Default）与 "users"（MaxCost 覆盖）两档

#### Scenario: 工厂构造时 Default 非法报错
- **WHEN** `NewAgristrettoFactory(&RistrettoConfigs{Default:{DefaultTTL:"abc"}})`
- **THEN** 返回 error（ToOptions/Validate 拦截，启动 fail-fast）

#### Scenario: 未使用 name 也启动期校验
- **WHEN** Namespaces 含永不 `Create` 的 name，其配置非法（如 `{"never-used":{DefaultTTL:"abc"}}`）
- **THEN** `NewAgristrettoFactory` 仍返回 error（全部 per-name 预解析，非惰性）

#### Scenario: fx 装配自动组合
- **WHEN** fx 中提供 IBinder + `ag_cache.FxAgCacheMode` + `FxAgCacheRistrettoMode`
- **THEN** 依赖图自动完成 binder→RistrettoConfigs→EngineFactory，Manager 拿到 ristretto 工厂（config `defaultEngine: ristretto` 命中）

### Requirement: per-name 配置（Default + Namespaces）
`RistrettoConfigs` SHALL 含 `Default RistrettoConfig`（全局限量默认）+ `Namespaces map[string]RistrettoConfig`（按缓存名覆盖）。YAML 键格式 `agcache.ristretto.namespaces.<name>.<field>`。
合并语义 SHALL 为**非零覆盖继承**：per-name 指定字段（非零/非空）覆盖 Default，未指定字段继承 Default（Default 再兜底 `DefaultRistrettoConfig()` 硬默认）。
`agristrettoFactory.Create(name)` SHALL 查预解析 map：命中返回该 name 的 Options；未命中返回 Default 的 Options。

#### Scenario: per-name 覆盖生效
- **WHEN** `Create("users")` 且 Namespaces["users"].MaxCost=500_000_000
- **THEN** 返回引擎的 MaxCost 为 500MB（覆盖 Default 的 100MB），其余字段继承 Default

#### Scenario: per-name 未命中用 Default
- **WHEN** `Create("not-configured")` 且 Namespaces 无该键
- **THEN** 返回 Default 配置的引擎（MaxCost=100MB, NumCounters=131072）

#### Scenario: per-name 非零覆盖继承
- **WHEN** Namespaces["users"] 仅指定 `NumCounters: 8_388_608`（未指定 MaxCost/BufferItems/DefaultTTL）
- **THEN** 合并结果 = Default 的 MaxCost/BufferItems/DefaultTTL + users 的 NumCounters=8M（非零覆盖，零继承）

#### Scenario: per-name 非法 TTL 启动报错（含 name 定位）
- **WHEN** Namespaces["users"].DefaultTTL="abc" 调 `NewAgristrettoFactory`
- **THEN** 返回 error，错误信息含 "users"（定位到具体 name）
