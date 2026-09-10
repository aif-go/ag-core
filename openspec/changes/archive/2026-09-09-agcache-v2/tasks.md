# AgCache v2 实现任务（TDD）

> 验证命令：`go test -race ./ag/ag_cache/...`；`go vet ./ag/ag_cache/...`；`go build ./...`
> 迁移源：`/home/houzw/rundata/DATA_hermes/hermes/references/agcache-poc/`（27 测试）
> 设计方案：`/home/houzw/document/hzw-obsidian/ag-core/ag_cache/2026-08-21-ag-cache-v2-design.md`

## 1. core — 接口与错误哨兵（cache.go）

- [x] RED: 定义业务接口与哨兵错误（编译断言）— `ag/ag_cache/cache.go` 创建
    Assertion: `ICache[T]`/`AdminCache[T]`/`Stats`/`LoaderFunc[T]` 存在；`ErrCacheMiss`/`ErrBackend` 可 `errors.Is`
    Expected failure: 包与类型不存在，编译失败
- [x] GREEN: 创建 `ag/ag_cache/cache.go` — 迁移 POC cache.go（无 `ErrNotSupported`）
    References RED test: 本阶段编译断言
    Verification: `go build ./ag/ag_cache/`

## 2. core — Engine SPI 与注册表（engine.go）

- [x] RED: 定义 SPI/注册表/可选能力（编译断言）— `ag/ag_cache/engine.go` 创建
    Assertion: `Engine`/`EngineFactory{Name, Create() 无参}`/`DefaultTTLProvider`/`RegisterEngine`/`EngineRegistered`/`getFactory`/`syncer`/`errBackend` 存在；`errBackend(ErrCacheMiss)` 原样、其他包装 `ErrBackend`
    Expected failure: 函数/类型不存在，编译失败
- [x] GREEN: 创建 `ag/ag_cache/engine.go` — SPI + 注册表（Create 无参）+ `DefaultTTLProvider` + `errBackend`
    References RED test: 本阶段编译断言
    Verification: `go build ./ag/ag_cache/`

## 3. core — Mock 工具（mock.go）

- [x] RED: MockCache/MockEngine 行为测试 — `ag/ag_cache/mock.go` 创建
    Assertion: `NewMock[T]()`/`NewMockEngine()` 可用；`PanicNext` 触发 panic、`Err` 注入后端错误；`MockCache.SetError` 生效；二者实现 `AdminCache`/`Engine`
    Expected failure: mock 类型不存在
- [x] GREEN: 创建 `ag/ag_cache/mock.go` — 迁移 POC mock.go（MockCache）+ engine.go 的 MockEngine
    References RED test: 本阶段编译断言
    Verification: `go build ./ag/ag_cache/`

## 4. core — 序列化（serializer.go）

- [x] RED: 序列化往返测试 — `ag/ag_cache/cache_test.go` → `TestSerialization_StructType`
    Assertion: 结构体经 `Set`+`Peek` 字段往返一致
    Expected failure: `Serializer`/`DefaultSerializer` 不存在
- [x] GREEN: 创建 `ag/ag_cache/serializer.go` — 迁移 POC serializer.go
    References RED test: TestSerialization_StructType
    Verification: `go test ./ag/ag_cache/ -run TestSerialization_StructType -count=1`

## 5. core — typedCache（typed.go）

- [x] RED: 错误语义/singleflight/panic 恢复/Option 测试 — `ag/ag_cache/cache_test.go` → `TestBackendError_NotTreatedAsMiss` / `TestSingleflight_LoaderCalledOnce` / `TestErrBackend_PanicRecovery` / `TestWithDefaultTTL_Overrides` / `TestWithEngine_SelectsEngine`
    Assertion: 后端故障返回 `ErrBackend` 不调 loader；并发 miss 只调一次 loader；引擎 panic 转 `ErrBackend`；`WithDefaultTTL` 覆盖默认；`WithEngine` 选择指定引擎
    Expected failure: `typedCache`/`Option`/`NewWithEngine` 不存在
- [x] GREEN: 创建 `ag/ag_cache/typed.go` — typedCache（`GetOrElse` 用 `WithoutCancel` + `sf.Do`，`recoverPanic` 转 `ErrBackend`）+ `Option{WithEngine, WithDefaultTTL, WithSerializer}` + `NewWithEngine`/`NewAdminWithEngine` + 非泛型 `stats()`
    References RED test: TestBackendError_NotTreatedAsMiss / TestSingleflight_LoaderCalledOnce / TestErrBackend_PanicRecovery / TestWithDefaultTTL_Overrides / TestWithEngine_SelectsEngine
    Verification: `go test ./ag/ag_cache/ -run 'TestBackendError_NotTreatedAsMiss|TestSingleflight_LoaderCalledOnce|TestErrBackend_PanicRecovery|TestWithDefaultTTL_|TestWithEngine_' -count=1`

## 6. core — LoaderCache（loader_cache.go）

- [x] RED: LoaderCache 全组测试 — `ag/ag_cache/cache_test.go` → `TestLoaderCache_Get_ReadThrough` / `TestLoaderCache_GetOrElse_CustomLoader` / `TestLoaderCache_Peek_NoLoader` / `TestLoaderCache_WithLoader`
    Assertion: 绑定 loader 后 `Get` 读穿透、多 key 复用、`GetOrElse` 临时 loader、`Peek` 不触发 loader、`WithLoader` 包装
    Expected failure: `LoaderCache`/`WithLoader` 不存在
- [x] GREEN: 创建 `ag/ag_cache/loader_cache.go` — LoaderCache + `WithLoader`（`New` 放 manager.go 用 `getOrCreate`）
    References RED test: TestLoaderCache_Get_ReadThrough 等 4 项
    Verification: `go test ./ag/ag_cache/ -run 'TestLoaderCache_' -count=1`

## 7. core — 配置（config.go）

- [x] RED: core 配置绑定测试 — `ag/ag_cache/config_test.go` → `TestBindAgCacheProperties` / `TestDefaultAgCacheProperties`
    Assertion: YAML 提供 `agcache.defaultEngine` 绑定正确；缺失时保持默认 "ristretto"；配置不含引擎参数/TTL
    Expected failure: `AgCacheProperties`/`BindAgCacheProperties` 不存在
- [x] GREEN: 创建 `ag/ag_cache/config.go` — `AgCacheProperties{DefaultEngine}` + `DefaultAgCacheProperties()` + `BindAgCacheProperties(binder)`（绑 `agcache`，无 value 标签按字段名）
    References RED test: TestBindAgCacheProperties / TestDefaultAgCacheProperties
    Verification: `go test ./ag/ag_cache/ -run 'TestBindAgCacheProperties|TestDefaultAgCacheProperties' -count=1`

## 8. core — 包级实例管理（manager.go）

- [x] RED: 实例管理测试 — `ag/ag_cache/manager_test.go` → `TestNew_LazyCreateAndReuse` / `TestManager_UnknownEngine_FailFast` / `TestManager_Visit` / `TestManager_SameNameDiffType_Panics` / `TestGetAdmin`
    Assertion: 懒创建复用；未知引擎 `NewManager` 报错；`Visit` 遍历已建 namespace；同名不同类型 panic；`GetAdmin` 返回 Admin 视角
    Expected failure: `Manager`/`NewManager`/`SetDefault`/`New`/`Get`/`GetAdmin`/`CloseAll` 不存在
- [x] GREEN: 创建 `ag/ag_cache/manager.go` — `Manager{defaultEngine, caches}` + `NewManager`（fail-fast）+ `Close`/`Visit`/`SetDefault` + `New`/`Get`/`GetAdmin`/`CloseAll` + `getOrCreate`（引擎名=WithEngine 或 defaultEngine；TTL=Option>工厂 DefaultTTLProvider>5min；同名复用首次）
    References RED test: TestNew_LazyCreateAndReuse / TestManager_UnknownEngine_FailFast / TestManager_Visit / TestManager_SameNameDiffType_Panics / TestGetAdmin
    Verification: `go test ./ag/ag_cache/ -run 'TestNew_|TestManager_|TestGetAdmin' -count=1`
- [x] RED: 默认 TTL 三级优先级测试 — `ag/ag_cache/manager_test.go` → `TestDefaultTTL_Priority`
    Assertion: 未传 Option 且引擎实现 `DefaultTTLProvider` → 引擎 TTL；引擎未实现 → 5min；`WithDefaultTTL` → 覆盖
    Expected failure: TTL 优先级逻辑未实现
- [x] GREEN: 在 `manager.go` 的 `getOrCreate` 实现 TTL 优先级 — `ag/ag_cache/manager.go` → `getOrCreate`
    References RED test: TestDefaultTTL_Priority
    Verification: `go test ./ag/ag_cache/ -run TestDefaultTTL_Priority -count=1`

## 9. core — 回归测试迁移（mock 层）

- [x] RED: 迁移剩余逻辑回归测试（隔离/Clear/懒创建/WithoutCancel/多引擎/纯读）— `ag/ag_cache/cache_test.go` → `TestIndependentInstances_Isolation` / `TestClear_OnlyAffectsOwnInstance` / `TestGetOrElse_Basic` / `TestGet_PureRead` / `TestLoader_NotCancelledByFirstCallerCtx` / `TestMultiEngine_RegisteredFactory`
    Assertion: 全部通过（mock 引擎工厂 + 注册表；多引擎注册 "mock"/"mem"）
    Expected failure: 注册表 mock 工厂未注册或断言失败
- [x] GREEN: 在 `ag/ag_cache/cache_test.go` 增加测试用 mock 引擎工厂并补齐引用（`registerMockEngine` 幂等）
    References RED test: TestIndependentInstances_Isolation 等 6 项
    Verification: `go test ./ag/ag_cache/ -run 'TestIndependentInstances_|TestClear_|TestGetOrElse_Basic|TestGet_PureRead|TestLoader_|TestMultiEngine_' -count=1`

## 10. 依赖 — 主 go.mod 引入 ristretto

- [x] 主 `go.mod` 添加 `github.com/dgraph-io/ristretto/v2 v2.4.2`（`go mod edit -require`）
- [x] `go mod tidy` 更新 go.sum（M4 统一收敛）

## 11. agristretto — 引擎（ristretto.go）

- [x] RED: 引擎集成测试（先失败）— `ag/ag_cache/agristretto/ristretto_test.go` → `TestTTL_Expiry` / `TestSet_ZeroTTL_NeverExpires` / `TestEngine_Stats` / `TestEviction` / `TestFactory_Create` / `TestFactory_DefaultTTL`
    Assertion: 正 ttl 到期 miss；ttl=0 永不过期；namespace 级 Stats 独立；小 MaxCost 淘汰；`agristrettoFactory.Create()` 无参成功且 `DefaultTTL()` 返回构造值
    Expected failure: `agristretto` 包与 `NewRistrettoEngine`/`agristrettoFactory` 不存在
- [x] GREEN: 创建 `ag/ag_cache/agristretto/ristretto.go` — `RistrettoConfig`/`DefaultRistrettoConfig`/`ristrettoEngine`（Metrics 开/cost 内部化/Sync）+ `agristrettoFactory{cfg, ttl}`（`Name()="ristretto"`、`Create()` 无参、`DefaultTTL()`）
    References RED test: TestTTL_Expiry / TestSet_ZeroTTL_NeverExpires / TestEngine_Stats / TestEviction / TestFactory_Create / TestFactory_DefaultTTL
    Verification: `go test ./ag/ag_cache/agristretto/ -run 'TestTTL_|TestSet_ZeroTTL_|TestEngine_|TestEviction|TestFactory_' -count=1`
- [x] RED: 异步写 + Sync 可见性 — `ag/ag_cache/agristretto/ristretto_test.go` → `TestSync_Visibility`
    Assertion: `GetOrElse` 返回后立即 `Get` 命中且值一致
    Expected failure: `syncer` 未实现或 `Sync` 缺失
- [x] GREEN: `ristrettoEngine` 实现 `Sync()`（`e.cache.Wait()`）— `ag/ag_cache/agristretto/ristretto.go` → `ristrettoEngine.Sync`
    References RED test: TestSync_Visibility
    Verification: `go test ./ag/ag_cache/agristretto/ -run TestSync_Visibility -count=1`

## 12. agristretto — 配置（config_ristretto.go）

- [x] RED: 引擎配置绑定测试 — `ag/ag_cache/agristretto/config_ristretto_test.go` → `TestBindRistrettoConfigProperties` / `TestDefaultRistrettoConfigProperties` / `TestParseTTL`
    Assertion: YAML 提供 `agcache.ristretto.maxCost`/`defaultTtl` 绑定正确；缺失保持默认；`parseTTL`：""→0、"60s"→60s、"0"→0、非法→错误
    Expected failure: `RistrettoConfigProperties`/`BindRistrettoConfigProperties`/`parseTTL` 不存在
- [x] GREEN: 创建 `ag/ag_cache/agristretto/config_ristretto.go` — `RistrettoConfigProperties{MaxCost, NumCounters, DefaultTTL string}` + `DefaultRistrettoConfigProperties()` + `BindRistrettoConfigProperties(binder)`（绑 `agcache.ristretto`）+ `parseTTL`
    References RED test: TestBindRistrettoConfigProperties / TestDefaultRistrettoConfigProperties / TestParseTTL
    Verification: `go test ./ag/ag_cache/agristretto/ -run 'TestBindRistrettoConfigProperties|TestDefaultRistrettoConfigProperties|TestParseTTL' -count=1`

## 13. agristretto — Fx 模块（zfx_agristretto.go）

- [x] RED: Fx 装配联合测试（core + 引擎）— `ag/ag_cache/zfx_ag_cache_test.go` → `TestFxAgCacheMode`
    Assertion: `fxtest.New(t, fx.Provide(binder), FxAgCacheMode, FxAgCacheRistrettoMode)` 可注入 `*ag_cache.Manager`，`New[string]("test", loader).Get` 可用，RequireStart/RequireStop 通过，OnStop Close
    Expected failure: `FxAgCacheMode`/`FxAgCacheRistrettoMode` 未装配或 group 收集失败
- [x] GREEN: 创建 `ag/ag_cache/agristretto/zfx_agristretto.go` — `ProvideAgristrettoFactory(binder)`（绑定配置 → `agristrettoFactory{cfg, ttl}`，返回 `ag_cache.EngineFactory`）+ `FxAgCacheRistrettoMode`（`fx.Provide(fx.Annotate(..., fx.ResultTags(\`group:"agcache.engine"\`)))`）
    References RED test: TestFxAgCacheMode
    Verification: `go test ./ag/ag_cache/ -run TestFxAgCacheMode -count=1`
- [x] RED: 幂等注册 + 生命周期统计 — `ag/ag_cache/zfx_ag_cache_test.go` → `TestFxAgCacheMode_IdempotentRegister` / `TestLogStats`
    Assertion: 引擎模块重复装配不重复注册不 panic；`LogStats` 输出包含 namespace 名与计数（经 Visit 验证）
    Expected failure: core 注册未做 `EngineRegistered` 幂等；`LogStats` 不存在
- [x] GREEN: 在 core `zfx_ag_cache.go` 实现 `NewAgCacheManager`（group 收集→幂等注册）+ `registerHooks`（SetDefault + OnStop LogStats/Close）+ `LogStats`
    References RED test: TestFxAgCacheMode_IdempotentRegister / TestLogStats
    Verification: `go test ./ag/ag_cache/ -run 'TestFxAgCacheMode_IdempotentRegister|TestLogStats' -count=1`

## 14. bench 迁移

- [x] 迁移 `ag/ag_cache/agristretto/bench_test.go` — 4 个 benchmark（`BenchmarkRistrettoAsyncSet`/`BenchmarkEngineSet`/`BenchmarkGetOrElse_Miss`/`BenchmarkGetOrElse_Hit`），import 改为 `github.com/aif-go/ag-core/ag/ag_cache` 与 `ag/ag_cache/agristretto`

## 15. 收尾验证（M4）

- [x] `go mod tidy`（主 go.mod 依赖收敛，go.sum 更新）
- [x] `go build ./...`
- [x] `go vet ./ag/ag_cache/...`
- [x] `go test -race ./ag/ag_cache/...` 全绿（27 个 POC 测试迁移 + 新增配置/Option/Fx 测试）
- [x] 对照设计方案文档校对公开 API 名称与语义（`New`/`Get`/`GetAdmin`/`CloseAll`/`WithEngine`/`WithDefaultTTL`/`Clear`/`Set(ttl...)`）
- [x] 确认 `fxs/`、`go.work`、`release_aif.sh` 零改动
