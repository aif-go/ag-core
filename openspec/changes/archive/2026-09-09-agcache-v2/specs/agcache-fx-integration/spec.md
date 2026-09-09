## ADDED Requirements

### Requirement: core 配置绑定（AgCacheProperties）
`AgCacheProperties` SHALL 通过 `ag_conf.IBinder` 以 `agcache` 前缀绑定，含 `DefaultEngine string`。
配置 key SHALL 默认按字段名（ag_conf 大小写不敏感匹配，`DefaultEngine` 匹配 YAML `defaultEngine`），不使用 value 标签。
`DefaultAgCacheProperties()` SHALL 返回带默认值的配置（`DefaultEngine="ristretto"`）。
`BindAgCacheProperties(binder)` SHALL 基于默认配置执行绑定，仅覆盖 YAML 出现的键。
核心配置 SHALL 只包含引擎选择，不包含引擎参数与 TTL。

#### Scenario: 完整配置绑定
- **WHEN** YAML 提供 `agcache.defaultEngine`
- **THEN** 绑定结果 DefaultEngine 与 YAML 一致

#### Scenario: 未配置字段保持默认
- **WHEN** YAML 缺失 `defaultEngine`
- **THEN** `DefaultEngine` 保持默认 "ristretto"

### Requirement: FxAgCacheMode 模块
`FxAgCacheMode` SHALL 为 `fx.Module("ag_cache")`，Provide 链为 `BindAgCacheProperties` → `NewAgCacheManager`。
`NewAgCacheManager` SHALL 经 `fx.In` 收集 `[]EngineFactory`（`group:"agcache.engine"`），对每个工厂 `EngineRegistered` 幂等判断后 `RegisterEngine`，再以 core 配置创建 Manager（defaultEngine 未注册快速失败）。
模块 SHALL 在 Invoke 阶段 `SetDefault(m)`（包级 `New/Get` 可用）并注册 OnStop 钩子：先 `LogStats` 输出各 namespace 统计，再 `Manager.Close()`。

#### Scenario: Fx 模块可解析 Manager
- **WHEN** `fxtest.New(t, fx.Provide(binder), FxAgCacheMode, FxAgCacheRistrettoMode)` 并注入使用
- **THEN** `*ag_cache.Manager` 可注入，`New`/`Get` 可用，RequireStart/RequireStop 通过且 OnStop 正确 Close

#### Scenario: 幂等注册
- **WHEN** 两个引擎模块提供同名工厂（如重复装配）
- **THEN** 不重复注册、不 panic（core 经 EngineRegistered 判断）

### Requirement: FxAgCacheRistrettoMode 模块
`FxAgCacheRistrettoMode` SHALL 为 `fx.Module("ag_cache.agristretto")`，经 `fx.Provide` + `fx.ResultTags` 将 `ProvideAgristrettoFactory` 的产物注入 `group:"agcache.engine"`。
`ProvideAgristrettoFactory(binder)` SHALL 绑定 `agcache.ristretto.*` 配置，构造 `agristrettoFactory{cfg, ttl}` 返回 `ag_cache.EngineFactory`。
引擎模块 SHALL 不直接调用 `RegisterEngine`（注册由 core 经 group 统一完成）。

#### Scenario: 工厂注入 group
- **WHEN** 装配 `FxAgCacheRistrettoMode`
- **THEN** `group:"agcache.engine"` 收到一个 Name="ristretto" 的工厂，且引擎注册表中有该工厂

#### Scenario: 业务装配端到端
- **WHEN** 应用装配 `FxAgConfModule` + `FxAgCacheMode` + `FxAgCacheRistrettoMode` 且 YAML 提供 `agcache.defaultEngine: ristretto` 与 `agcache.ristretto.*`
- **THEN** `ag_cache.New[string]("users", loader).Get(...)` 可读写缓存，默认 TTL 来自引擎配置

### Requirement: 生命周期统计日志
`LogStats(m *Manager)` SHALL 遍历 Manager 已创建 namespace，用 slog 输出 `hits/misses/evictions/entries` 统计（namespace 维度 label）。
OnStop SHALL 先输出统计再关闭缓存。

#### Scenario: OnStop 输出统计并关闭
- **WHEN** Fx 应用停止
- **THEN** 日志包含各 namespace 统计且 Manager 已 Close（幂等）
