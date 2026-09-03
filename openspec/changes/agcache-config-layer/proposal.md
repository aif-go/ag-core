# agristretto 配置层重构：两层模型 + 构造器职责分离

## Why

agristretto 引擎当前的配置设计存在三层散落 + TTL 游离 + 职责混杂问题：
- `RistrettoConfig` 与 `RistrettoConfigProperties` 字段重复（MaxCost/NumCounters 双份维护）；
- `DefaultTTL` 以 string 存在于 Properties、经 `parseTTL` 拆到 factory 的 `ttl` 字段，游离于运行时模型之外；
- `BufferItems`/`Metrics` 硬编码在 `newRistrettoEngine`，用户无法配置；
- `ProvideAgristrettoFactory` 揉合"绑定 + 转换 + 构造"三个职责，违背单一职责。

目标：按"绑定层 + 运行时层"两层模型重构，让配置清晰、可校验、可手动使用，同时让默认值显式可读。

## What Changes

- **两层配置模型**：
  - `RistrettoConfig`（绑定层）：YAML 绑定模型，含 `MaxCost`/`NumCounters`/`BufferItems`/`DefaultTTL string`；`DefaultRistrettoConfig()` 显式给出全部默认值；`BindRistrettoConfig(binder)` 对接 IBinder。
  - `RistrettoOptions`（运行时层）：实际创建 Ristretto 实例的参数，含 `DefaultTTL time.Duration`；`NewRistrettoEngine(opts)` 单参创建引擎。
  - **BREAKING**：删除 `RistrettoConfigProperties`（含 `DefaultRistrettoConfigProperties`/`BindRistrettoConfigProperties`）；原 `RistrettoConfig` 语义迁移至 `RistrettoOptions`。
- **构造器职责分离**：
  - `RistrettoConfig.ToOptions()`：转换 + 非负校验 + `parseTTL` + 默认填充。
  - `NewAgristrettoFactory(cfg)`：由绑定配置构造工厂（内部调 `ToOptions` + `opts.Validate()` 双保险）。
  - `agristrettoFactory{cfg, opts}`：同时保留原始绑定配置 + 解析后的运行时配置。
  - `RistrettoOptions.Validate()`：导出校验（非负），覆盖手动构造 Options 的场景。
  - `FxAgCacheRistrettoMode`：删 `ProvideAgristrettoFactory`，改 `fx.Provide(BindRistrettoConfig, NewAgristrettoFactory→group)`，由 fx 依赖图自动装配。
- **非法数据 fail-fast**：`ToOptions`/`Validate` 对负值（负 MaxCost/NumCounters/BufferItems）报错；`parseTTL` 对非法 TTL 字符串报错（装配期即暴露）。
- **per-name 配置（Default + Namespaces）**：新增容器层 `RistrettoConfigs{Default RistrettoConfig; Namespaces map[string]RistrettoConfig}`——`Default` 是全局限量默认，`Namespaces` 按缓存名覆盖。YAML：`agcache.ristretto.namespaces.<name>.<field>`。合并语义为**非零覆盖继承**：per-name 指定字段（非零/非空）覆盖 Default，未指定字段继承 Default（Default 再兜底硬默认）。
- **启动期预解析校验（方案 B）**：`NewAgristrettoFactory(cfg)` 在装配期把 Default + **所有** Namespaces 条目逐一 `ToOptions`+`Validate` 预解析，结果缓存进 `map[string]RistrettoOptions`；任一 per-name 非法（TTL 错/负值）→ 构造返回 error → fx 启动失败。`Create(name)` 运行期直接查表（未命中用 default），零解析零报错——**配置错误启动即死，非运行时 panic**。
- **默认值显式化**：`DefaultRistrettoConfig()` 显式给出全部四字段默认——`MaxCost=100MB`/`NumCounters=131072`（2^17，100K 档）/`BufferItems=64`/`DefaultTTL="0"`（显式"永不过期"）。默认取小是框架取舍：sketch 预分配内存随 NumCounters 线性（10M≈48MB/实例 vs 100K≈0.7MB/实例），多数缓存 name（参数/配置/枚举，key 有限）小 NumCounters 足够；NumCounters 只影响淘汰精度、不影响容量与正确性，特殊大缓存（接口热缓存）经 `Namespaces` per-name 覆盖。`ToOptions` 对 `<=0` 字段填默认（MaxCost→100MB、NumCounters→131072、BufferItems→64），**不做 MaxCost×10% 推导**（避免 next2Power 跳档浪费，如 10M→2^24 白送内存）。

## Capabilities

### New Capabilities
- `agcache-engine-config`: agristretto 引擎配置分层——`RistrettoConfig`（绑定层叶子，含默认值/ToOptions）+ `RistrettoConfigs`（容器层 Default+Namespaces per-name）+ `RistrettoOptions`（运行时层，含 Validate/NewRistrettoEngine）+ 构造器职责分离（BindRistrettoConfig/NewAgristrettoFactory 启动期预解析校验/fx group 装配）

### Modified Capabilities
<!-- 无：openspec/specs/ 尚无现有 spec -->

## Impact

- **修改文件**：
  - `ag/ag_cache/agristretto/config_ristretto.go`：`RistrettoConfigProperties`→`RistrettoConfig`；`DefaultRistrettoConfig()` 全默认值；`BindRistrettoConfig`；`ToOptions()`；`parseTTL` 保留
  - `ag/ag_cache/agristretto/ristretto.go`：`RistrettoConfig`→`RistrettoOptions`；`Validate()` 导出；`NewRistrettoEngine(opts)` 单参；`NewAgristrettoFactory(cfg)`；`agristrettoFactory{cfg,opts}`；删双参 `newRistrettoEngine`
  - `ag/ag_cache/agristretto/zfx_agristretto.go`：删 `ProvideAgristrettoFactory`；改 `fx.Provide(BindRistrettoConfig, NewAgristrettoFactory)`
- **测试迁移**：`config_ristretto_test.go`（重命名 + ToOptions/Validate 用例）、`ristretto_test.go`/`bench_test.go`/`test/benchmark_test.go`（`RistrettoConfig{...}`→`RistrettoOptions{...}`，工厂用 `NewAgristrettoFactory`）
- **文档**：README 配置结构描述同步（绑定层 + 运行时层）
- **无破坏性变更**（对 core `ag_cache` 包、`fxs/`、`go.work`、`release_aif.sh`）
- **破坏性影响（相对当前 hzw_cache_dev）**：删除 `RistrettoConfigProperties` 类型族；`NewRistrettoEngine` 签名变化
