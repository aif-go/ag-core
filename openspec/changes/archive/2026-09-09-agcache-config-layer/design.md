# agristretto 配置层重构设计

## Context

- **现状**：agristretto 引擎当前有三层散落配置：`RistrettoConfig`（ristretto.go，运行时，MaxCost/NumCounters）、`RistrettoConfigProperties`（config_ristretto.go，YAML 绑定，含 DefaultTTL string）、`agristrettoFactory{cfg, ttl}`（工厂拆两个字段）、`ProvideAgristrettoFactory`（zfx，揉合绑定+转换+构造）。
- **问题**：字段重复、TTL 游离、BufferItems 硬编码、构造器职责混杂。
- **目标**：分层配置模型（叶子绑定层 + 容器层 per-name + 运行时层）+ 构造器职责分离 + 校验覆盖所有入口（含启动期预解析）+ 默认值显式化。兼容 core `ag_cache` 包（Engine SPI 不变）。

## Goals / Non-Goals

**Goals:**
- `RistrettoConfig`（绑定层叶子）：YAML 模型，显式默认值，`ToOptions()`（校验+转换+默认填充）
- `RistrettoConfigs`（容器层）：`Default` + `Namespaces map[string]RistrettoConfig` per-name 覆盖（非零覆盖继承）
- `RistrettoOptions`（运行时层）：完整运行时参数（含 `DefaultTTL time.Duration`），`Validate()` 导出，`NewRistrettoEngine(opts)` 单参
- 构造器分离：`NewAgristrettoFactory(cfg)` + fx group 自动装配（删 `ProvideAgristrettoFactory`）
- `agristrettoFactory{cfg, opts map}`：保留原始 + **启动期预解析全部配置**（Default+所有 Namespaces → map[string]RistrettoOptions）
- 校验覆盖：绑定装配（ToOptions + Validate 双保险）与手动场景（NewRistrettoEngine 内 Validate）
- per-name 配置：`Create(name)` 查表（未命中用 default），配置错误启动即 fail-fast

**Non-Goals:**
- 暴露 Ristretto `Metrics`/`TtlTickerDurationInSec`（本轮保持内部默认）
- Ristretto 底层缓存淘汰策略调优

## Decisions

### D1. 分层配置模型（RistrettoConfig / RistrettoConfigs / RistrettoOptions）
```go
// config_ristretto.go —— 绑定层叶子（单 name 配置）
type RistrettoConfig struct {
    MaxCost     int64  // 内存预算（字节），0=默认 100MB
    NumCounters int64  // 频率计数器数量，0=默认 131072（2^17，100K 档）
    BufferItems int64  // 读缓冲环大小，0=默认 64
    DefaultTTL  string // ""或"0"=永不过期；"60s"=60秒；非法→ToOptions 报错
}

// config_ristretto.go —— 容器层（全局默认 + per-name 覆盖）
type RistrettoConfigs struct {
    Default    RistrettoConfig            // 全局限量默认（YAML 缺省 → DefaultRistrettoConfig() 兜底）
    Namespaces map[string]RistrettoConfig // 按缓存名覆盖，键如 "users"
}
// 注意：叶子与容器分离，避免 map[string]RistrettoConfig 递归嵌套。

// ristretto.go —— 运行时层
type RistrettoOptions struct {
    MaxCost     int64
    NumCounters int64
    BufferItems int64
    DefaultTTL  time.Duration
}
```
- 删除 `RistrettoConfigProperties`/`DefaultRistrettoConfigProperties`/`BindRistrettoConfigProperties`
- `NumCounters` 是 TinyLFU 频率 sketch 大小（影响淘汰精度、非条数上限）；条数上限由 MaxCost 决定

### D2. 默认值与默认填充
```go
const (
    defaultMaxCost     = 100_000_000 // 100MB 内容预算
    defaultNumCounters = 131_072     // 2^17 = 100K 档
    defaultBufferItems = 64
)

func DefaultRistrettoConfig() RistrettoConfig {
    return RistrettoConfig{MaxCost: defaultMaxCost, NumCounters: defaultNumCounters, BufferItems: defaultBufferItems, DefaultTTL: "0"}
}
```
- `DefaultRistrettoConfig()` 显式给出全部四字段默认（`DefaultTTL="0"` 显式"永不过期"，YAML 缺省起点）
- **默认 NumCounters=2^17（131072，100K 档）**：sketch 预分配 ~0.7MB/实例（10M 档 48MB），多 name 线性叠加可控。NumCounters 只影响淘汰精度（缓存满时 TinyLFU 判断），不影响容量（MaxCost 定）与正确性；多数缓存 name（参数/配置/枚举）key 有限，永不淘汰，小 NumCounters 足够
- **不做 MaxCost×10% 推导**：原推导 100MB→10M→next2Power 取 2^24（33.5MB），跳档浪费且默认过大。改为固定默认 131072；特殊大缓存（接口热缓存，10 万+ key）后续 per-name 配置覆盖

### D3. ToOptions（转换 + 非负校验 + parseTTL + 默认填充）
```go
func (c RistrettoConfig) ToOptions() (RistrettoOptions, error) {
    if c.MaxCost < 0 || c.NumCounters < 0 || c.BufferItems < 0 {
        return RistrettoOptions{}, fmt.Errorf("agcache: negative value in RistrettoConfig")
    }
    ttl, err := parseTTL(c.DefaultTTL)   // string→duration，纯数字报错（防纳秒陷阱）
    if err != nil { return RistrettoOptions{}, err }
    opts := RistrettoOptions{DefaultTTL: ttl}
    if c.MaxCost <= 0 { opts.MaxCost = defaultMaxCost } else { opts.MaxCost = c.MaxCost }
    if c.NumCounters <= 0 { opts.NumCounters = defaultNumCounters } else { opts.NumCounters = c.NumCounters }
    if c.BufferItems <= 0 { opts.BufferItems = 64 } else { opts.BufferItems = c.BufferItems }
    return opts, nil
}
```

### D4. RistrettoOptions.Validate + NewRistrettoEngine
```go
func (o RistrettoOptions) Validate() error  // 非负校验，导出供手动场景

func NewRistrettoEngine(opts RistrettoOptions) (ag_cache.Engine, error) {
    if err := opts.Validate(); err != nil { return nil, err }
    // 零值兜底（0→默认/推导），然后 ristretto.NewCache
}
```
- 覆盖手动构造 Options 场景（测试/嵌入直接 NewRistrettoEngine）

### D5. 构造器职责分离（Bind + Factory）+ 启动期预解析
```go
// 职责1：绑定容器（Default + Namespaces）
func BindRistrettoConfig(binder ag_conf.IBinder) (*RistrettoConfigs, error) {
    cfg := RistrettoConfigs{Default: DefaultRistrettoConfig(), Namespaces: map[string]RistrettoConfig{}}
    if err := binder.Bind(&cfg, RistrettoConfPrefix); err != nil {
        return nil, err
    }
    return &cfg, nil
}

// 职责2：构造工厂——启动期预解析全部配置并校验
func NewAgristrettoFactory(cfg *RistrettoConfigs) (ag_cache.EngineFactory, error) {
    opts := make(map[string]RistrettoOptions, 1+len(cfg.Namespaces))
    def, err := cfg.Default.ToOptions()
    if err != nil { return nil, err }
    if err := def.Validate(); err != nil { return nil, err }
    opts[""] = def // 空 name = 全局默认

    for name, nc := range cfg.Namespaces {
        merged := mergeConfig(cfg.Default, nc)   // 非零覆盖继承
        o, err := merged.ToOptions()
        if err != nil { return nil, fmt.Errorf("agcache: namespace %q: %w", name, err) }
        if err := o.Validate(); err != nil { return nil, err }
        opts[name] = o
    }
    return agristrettoFactory{cfg: *cfg, opts: opts}, nil
}

// 非零覆盖继承：per-name 指定字段（非零/非空）覆盖 Default，未指定继承 Default
func mergeConfig(def, nc RistrettoConfig) RistrettoConfig {
    if nc.MaxCost != 0 { def.MaxCost = nc.MaxCost }
    if nc.NumCounters != 0 { def.NumCounters = nc.NumCounters }
    if nc.BufferItems != 0 { def.BufferItems = nc.BufferItems }
    if nc.DefaultTTL != "" { def.DefaultTTL = nc.DefaultTTL }
    return def
}

type agristrettoFactory struct {
    cfg  RistrettoConfigs
    opts map[string]RistrettoOptions // "" = default + 每 per-name
}
func (f agristrettoFactory) Name() string { return "ristretto" }
func (f agristrettoFactory) Create(name string) (ag_cache.Engine, error) {
    o, ok := f.opts[name]
    if !ok { o = f.opts[""] } // 未命中用全局默认
    return NewRistrettoEngine(o)
}
```
- **启动期预解析（方案 B）**：`NewAgristrettoFactory` 装配期把 Default + 所有 Namespaces 逐一 `ToOptions`+`Validate`，非法即返回 error → fx 启动失败。`Create(name)` 纯查表，运行期零解析零报错
- YAML 缺省 `default:` 键 → `BindRistrettoConfig` 已用 `DefaultRistrettoConfig()` 兜底
```go
// zfx_agristretto.go —— 删 ProvideAgristrettoFactory，fx 依赖图组合
var FxAgCacheRistrettoMode = fx.Module("ag_cache.agristretto",
    fx.Provide(
        BindRistrettoConfig,                                                          // 职责1：绑定
        fx.Annotate(NewAgristrettoFactory, fx.ResultTags(`group:"agcache.engine"`)), // 职责2：构造+预解析
    ),
)
```

## 函数签名清单

| 文件 | 符号 | 签名变化 |
|------|------|---------|
| `config_ristretto.go` | `RistrettoConfig` | 重命名自 `RistrettoConfigProperties`（叶子，含 DefaultTTL string） |
| `config_ristretto.go` | `RistrettoConfigs` | 新增（容器：`Default` + `Namespaces map[string]RistrettoConfig`） |
| `config_ristretto.go` | `mergeConfig(def, nc) RistrettoConfig` | 新增（非零覆盖继承） |
| `config_ristretto.go` | `DefaultRistrettoConfig()` | 替代 `DefaultRistrettoConfigProperties()`，显式全默认值 |
| `config_ristretto.go` | `BindRistrettoConfig(binder) (*RistrettoConfigs, error)` | 替代 `BindRistrettoConfigProperties`（绑容器） |
| `config_ristretto.go` | `(c RistrettoConfig) ToOptions() (RistrettoOptions, error)` | 新增 |
| `config_ristretto.go` | `parseTTL(s string) (time.Duration, error)` | 保留 |
| `ristretto.go` | `RistrettoOptions` | 原 `RistrettoConfig`（运行时），加 DefaultTTL time.Duration |
| `ristretto.go` | `(o RistrettoOptions) Validate() error` | 新增 |
| `ristretto.go` | `NewRistrettoEngine(opts RistrettoOptions) (ag_cache.Engine, error)` | 签名变化（单参，替代双参 newRistrettoEngine） |
| `ristretto.go` | `NewAgristrettoFactory(cfg *RistrettoConfigs) (ag_cache.EngineFactory, error)` | 新增（启动期预解析校验） |
| `ristretto.go` | `agristrettoFactory{cfg RistrettoConfigs; opts map[string]RistrettoOptions}` | 持原始 + 预解析 map；删 ttl 字段 |
| `zfx_agristretto.go` | `FxAgCacheRistrettoMode` | `fx.Provide(BindRistrettoConfig, NewAgristrettoFactory)` |
| `zfx_agristretto.go` | `ProvideAgristrettoFactory` | 删除 |

## 测试策略

| 层 | 文件 | 覆盖 |
|----|------|------|
| 绑定/转换单元 | `config_ristretto_test.go` | DefaultRistrettoConfig 全默认值、BindRistrettoConfig 容器缺省/覆盖、ToOptions 默认填充/负值报错/parseTTL 非法与纯数字报错 |
| 运行时单元 | `ristretto_test.go` | Options.Validate 负值报错、NewRistrettoEngine 零值兜底、引擎 CRUD/TTL/Clear/Create 行为 |
| 工厂单元 | `ristretto_test.go` | NewAgristrettoFactory 预解析（Default+Namespaces 全校验）、per-name 覆盖生效/继承、per-name 非法 TTL/负值 → 构造报错、Create(name) 查表（命中 per-name/未命中 default） |
| 装配集成 | `test/`（setup_test 的 startFx）+ `zfx_ag_cache_test.go` | fx 依赖图 binder→Config→Factory→Manager 自动装配、usage 端到端 |
| 基准 | `bench_test.go`/`benchmark_test.go` | `RistrettoOptions` 替换适配（编译） |

验证：`go build ./...`、`go vet ./ag/ag_cache/...`、`go test -race ./ag/ag_cache/...`。

## Risks / Trade-offs

- [R1] **破坏性变更**：删 `RistrettoConfigProperties` 类型族、`NewRistrettoEngine` 签名变化——需迁移测试/bench。
- [R2] **DefaultTTL string vs time.Duration**：绑定层保留 string（parseTTL 严格，纯数字报错），运行时用 duration——两层各司其职，不直接用 ag_conf 的 duration 绑定（避免纯数字纳秒陷阱）。
- [R3] **默认 100MB/131072 是每缓存名独立实例上限**：多 name 多实例按写入实际占用，非满配；sketch 预分配 ~0.7MB/实例（10M 档 48MB），多 name 线性叠加；业务大缓存（10 万+ key）需 per-name 配置或 YAML 显式覆盖（后续 change）。

## Migration Plan

按本 change 分层实施：config_ristretto.go → ristretto.go → zfx_agristretto.go → 测试迁移（config/ristretto/bench）→ README 同步 → 全量验证。

## Open Questions

- 无（设计已多轮收敛）
