# agristretto 配置层重构任务（TDD）

> 验证命令：`go build ./...`；`go vet ./ag/ag_cache/...`；`go test -race ./ag/ag_cache/...`

## 1. 绑定层重构（config_ristretto.go）

- [x] RED: RistrettoConfig 类型编译断言 — `ag/ag_cache/agristretto/config_ristretto_test.go` → `TestConfig_LayerCompile`
    Assertion: `RistrettoConfig{MaxCost,NumCounters,BufferItems,DefaultTTL string}` 存在；`RistrettoConfigProperties`/`DefaultRistrettoConfigProperties`/`BindRistrettoConfigProperties` 不存在
    Expected failure: 旧 Properties 类型仍存在、新 Config 缺失
- [x] RED: DefaultRistrettoConfig 全默认值 — `ag/ag_cache/agristretto/config_ristretto_test.go` → `TestDefaultRistrettoConfig`
    Assertion: `DefaultRistrettoConfig()` 返回 MaxCost=100_000_000、NumCounters=131_072（2^17，100K 档）、BufferItems=64、DefaultTTL="0"（四字段显式，DefaultTTL 显式永不过期）
    Expected failure: 当前 Properties 默认只设 MaxCost
- [x] RED: BindRistrettoConfig 绑定 — `ag/ag_cache/agristretto/config_ristretto_test.go` → `TestBindRistrettoConfig`
    Assertion: `BindRistrettoConfig(binder)` 绑定 `agcache.ristretto.maxcost/numcounters/bufferitems/defaultttl`；缺省保持默认值；failBinder 返回 error
    Expected failure: 绑定函数名为 BindRistrettoConfigProperties
- [x] GREEN: 重构 `ag/ag_cache/agristretto/config_ristretto.go` — `RistrettoConfigProperties`→`RistrettoConfig`（含 BufferItems）；`DefaultRistrettoConfig()` 全默认值；`BindRistrettoConfig`；`parseTTL` 保留
    References RED test: TestConfig_LayerCompile / TestDefaultRistrettoConfig / TestBindRistrettoConfig
    Verification: `go test ./ag/ag_cache/agristretto/ -run 'TestConfig_LayerCompile|TestDefaultRistrettoConfig|TestBindRistrettoConfig' -count=1`

## 2. ToOptions 转换与校验

- [x] RED: ToOptions 转换/校验测试 — `ag/ag_cache/agristretto/config_ristretto_test.go` → `TestToOptions`
    Assertion: `RistrettoConfig{DefaultTTL:"60s"}.ToOptions()` → DefaultTTL=60s；`MaxCost<=0`→默认 100MB；`NumCounters<=0`→默认 131072（固定，非 MaxCost×10% 推导）；`BufferItems<=0`→64；负值报错；`DefaultTTL:"abc"`/`"60"`（纯数字）报错
    Expected failure: ToOptions 未定义
- [x] GREEN: 新增 `ag/ag_cache/agristretto/config_ristretto.go` → `func (c RistrettoConfig) ToOptions() (RistrettoOptions, error)`
    References RED test: TestToOptions
    Verification: `go test ./ag/ag_cache/agristretto/ -run TestToOptions -count=1`

## 3. 运行时层重构（ristretto.go）

- [x] RED: RistrettoOptions + Validate 编译断言 — `ag/ag_cache/agristretto/ristretto_test.go` → `TestOptions_LayerCompile`
    Assertion: `RistrettoOptions{...DefaultTTL time.Duration}` 存在；`Validate()` 存在；`NewRistrettoEngine(opts RistrettoOptions)` 单参存在；原 `RistrettoConfig`（运行时）不再有 DefaultTTL 语义
    Expected failure: 旧 RistrettoConfig 运行时模型/双参 newRistrettoEngine 仍存在
- [x] RED: Validate 负值报错 + 零值兜底 — `ag/ag_cache/agristretto/ristretto_test.go` → `TestOptions_Validate`
    Assertion: `RistrettoOptions{MaxCost:-1}.Validate()` 报错；`NewRistrettoEngine(RistrettoOptions{MaxCost:1<<20})`（NumCounters/BufferItems=0）成功且推导默认
    Expected failure: 负值被 `<=0→默认` 静默吞掉、Validate 未定义
- [x] GREEN: 重构 `ag/ag_cache/agristretto/ristretto.go` — 原 `RistrettoConfig`→`RistrettoOptions`（加 DefaultTTL time.Duration）；`Validate()` 导出；`NewRistrettoEngine(opts)` 单参（先 Validate 再零值兜底）；删双参 `newRistrettoEngine`
    References RED test: TestOptions_LayerCompile / TestOptions_Validate
    Verification: `go test ./ag/ag_cache/agristretto/ -run 'TestOptions_LayerCompile|TestOptions_Validate' -count=1`

## 4. 容器层 + 工厂构造器分离（config_ristretto.go + ristretto.go + zfx_agristretto.go）

- [x] RED: 容器层类型编译断言 — `ag/ag_cache/agristretto/config_ristretto_test.go` → `TestConfigs_LayerCompile`
    Assertion: `RistrettoConfigs{Default RistrettoConfig; Namespaces map[string]RistrettoConfig}` 存在；`mergeConfig` 存在；`BindRistrettoConfig(binder) (*RistrettoConfigs, error)` 返回容器
    Expected failure: 容器层/mergeConfig 未定义、Bind 仍返回叶子
- [x] GREEN: 新增 `ag/ag_cache/agristretto/config_ristretto.go` — `RistrettoConfigs` 容器；`mergeConfig(def, nc)` 非零覆盖继承；`BindRistrettoConfig` 绑容器（Default 起点 + Namespaces 空 map）
    References RED test: TestConfigs_LayerCompile
    Verification: `go build ./ag/ag_cache/agristretto/`
- [x] RED: NewAgristrettoFactory 预解析 + per-name 测试 — `ag/ag_cache/agristretto/ristretto_test.go` → `TestFactory_NewAgristrettoFactory` / `TestFactory_Namespaces`
    Assertion: `NewAgristrettoFactory(&RistrettoConfigs{Default:{DefaultTTL:"60s"}, Namespaces:{"users":{MaxCost:500_000_000}}})` 成功且 Name()="ristretto"；`Create("users")` 用 MaxCost=500MB 建引擎；`Create("other")` 未命中用 Default；per-name 非法 TTL/负值 → 构造报错（含 "users" 定位）；未使用 name 非法也报错
    Expected failure: NewAgristrettoFactory 未定义、仍持单 opts 无 map、ProvideAgristrettoFactory 仍揉合
- [x] GREEN: 重构 `ag/ag_cache/agristretto/ristretto.go` — `func NewAgristrettoFactory(cfg *RistrettoConfigs) (ag_cache.EngineFactory, error)`：启动期预解析 Default+全部 Namespaces → `opts map[string]RistrettoOptions`（""=default）；`agristrettoFactory{cfg, opts map}`（删 ttl 字段）；`Create(name)` 查表（未命中用 ""）
    References RED test: TestFactory_NewAgristrettoFactory / TestFactory_Namespaces
    Verification: `go test ./ag/ag_cache/agristretto/ -run 'TestFactory_NewAgristrettoFactory|TestFactory_Namespaces' -count=1`
- [x] GREEN: 重构 `ag/ag_cache/agristretto/zfx_agristretto.go` — 删 `ProvideAgristrettoFactory`；`fx.Provide(BindRistrettoConfig, fx.Annotate(NewAgristrettoFactory, fx.ResultTags(group:"agcache.engine")))`
    References RED test: 装配测试（见下）
    Verification: `go build ./ag/ag_cache/agristretto/`

## 5. 装配与迁移验证

- [x] RED: fx 装配回归 — `ag/ag_cache/agristretto/zfx_agristretto.go` + `test/setup_test.go`（startFx）自动组合验证
    Assertion: fx 依赖图 binder→Configs（Default+Namespaces）→Factory→Manager 装配成功（`FxAgCacheMode` + `FxAgCacheRistrettoMode` + IBinder 即可用 ristretto）
    Expected failure: fx 依赖缺失（BindRistrettoConfig/NewAgristrettoFactory 未被 Provide）
- [x] RED: 迁移旧测试到新模型 — 替换 `RistrettoConfig{MaxCost:...}` 构造为 `RistrettoOptions{...}`；工厂构造改 `NewAgristrettoFactory(&RistrettoConfigs{...})`
    Assertion: `ristretto_test.go`/`bench_test.go`/`test/benchmark_test.go` 用 `RistrettoOptions` 编译通过
    Expected failure: 引用已删的 RistrettoConfig（运行时）/双参 newRistrettoEngine
- [x] GREEN: 更新测试 — `config_ristretto_test.go`（删旧 Properties 测试、加 ToOptions/Validate/RistrettoConfigs/mergeConfig 用例）、`ristretto_test.go`（工厂测试改容器 + 补 per-name 覆盖/继承/非法用例）、`bench_test.go`/`benchmark_test.go`（RistrettoOptions 替换）
    References RED test: 全部迁移测试
    Verification: `go test ./ag/ag_cache/... -count=1`

## 6. 收尾验证

- [x] `go build ./...`
- [x] `go vet ./ag/ag_cache/...`
- [x] `goimports -l ag/ag_cache/` 为空（新增/改动源码须经 goimports 格式化）
- [x] `go test -race ./ag/ag_cache/...` 全绿
- [x] 同步 README 配置结构描述（RistrettoConfig 叶子 + RistrettoConfigs 容器 Default/Namespaces per-name + RistrettoOptions 运行时 + 启动预解析校验）
- [x] 确认 core `ag_cache` 包、`fxs/`、`go.work`、`release_aif.sh` 零改动
