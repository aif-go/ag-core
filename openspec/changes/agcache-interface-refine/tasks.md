# AgCache 接口定稿升级任务（TDD）

> 验证命令：`go test -race ./ag/ag_cache/...`；`go vet ./ag/ag_cache/...`；`go build ./...`
> 参考：SPI 审计接口定稿章节 + issues 文档（`/home/houzw/document/hzw-obsidian/ag-core/ag_cache/`）

## 1. ICache 接口定稿（cache.go）

- [x] RED: 编译断言 ICache 新形态 — `ag/ag_cache/cache_test.go` → `TestICache_InterfaceCompile`
    Assertion: `ICache[T]` 含 `Get/TryGet/GetOrElse/Set(ctx,key,value)/Del(keys...)/Clear`；`AdminCache`/`Peek` 不存在（编译失败即证明移除）
    Expected failure: 旧 ICache 带 `Set(ttl...)` 变参、`AdminCache` 仍存在
- [x] GREEN: 更新 `ag/ag_cache/cache.go` — ICache 去 ttl 变参 + 新增 `TryGet` 签名；删除 `AdminCache[T]`
    References RED test: TestICache_InterfaceCompile
    Verification: `go build ./ag/ag_cache/`

## 2. Engine SPI 定稿（engine.go）

- [x] RED: 编译断言 Engine 新形态 + BulkDelEngine — `ag/ag_cache/cache_test.go` → `TestEngine_InterfaceCompile`
    Assertion: `Engine` 含 `Get/Set(ctx,key,value,ttl)/Del/Clear/Close`，无 `Stats`；`BulkDelEngine` 存在
    Expected failure: 旧 Engine 含 `Stats()`；`BulkDelEngine` 不存在
- [x] GREEN: 更新 `ag/ag_cache/engine.go` — Engine 删 `Stats`；新增 `type BulkDelEngine interface{ DelMany(ctx, keys...) error }`；保留 DefaultTTLProvider/syncer
    References RED test: TestEngine_InterfaceCompile
    Verification: `go build ./ag/ag_cache/`

## 3. typedCache 适配（typed.go）

- [x] RED: TryGet 行为 + Set 无变参 + Del 批量探测 — `ag/ag_cache/cache_test.go` → `TestTryGet_Miss` / `TestTryGet_Hit` / `TestSet_NoTTLVararg` / `TestDel_BulkDelEngine`
    Assertion: `TryGet` miss→`(zero,false,nil)`、hit→`(v,true,nil)`；`Set(ctx,k,v)` 编译且写默认 TTL；引擎实现 `BulkDelEngine` 时 `Del` 走 `DelMany`
    Expected failure: `TryGet` 不存在；`Set` 仍带 ttl 变参；`Del` 未探测 BulkDelEngine
- [x] GREEN: 更新 `ag/ag_cache/typed.go` — `Set` 去 ttl 变参（内部 `engine.Set(ctx,key,data,c.defaultTTL)`）；新增 `TryGet`；`Del` 探测 `BulkDelEngine`（实现则 `DelMany`，否则循环）；删除 `Peek`/`Stats`/`stats()`
    References RED test: TestTryGet_Miss / TestTryGet_Hit / TestSet_NoTTLVararg / TestDel_BulkDelEngine
    Verification: `go test ./ag/ag_cache/ -run 'TestTryGet_|TestSet_NoTTL|TestDel_BulkDelEngine' -count=1`

## 4. Manager 调整（manager.go）

- [x] RED: 移除 GetAdmin/Visit + TTL 优先级 + 负 TTL 防御 — `ag/ag_cache/manager_test.go` → `TestTTL_Priority` / `TestWithDefaultTTL_Negative_Errors`
    Assertion: `WithDefaultTTL(30s)` > 引擎 DefaultTTLProvider(60s) > 5min；`WithDefaultTTL(-1)` 返回错误
    Expected failure: `GetAdmin`/`Visit` 仍存在；TTL 优先级未实现；负 TTL 未防御
- [x] GREEN: 更新 `ag/ag_cache/manager.go` — 删 `GetAdmin`/`Visit`；`getOrCreate` TTL 计算（Option>工厂>5min）且 `WithDefaultTTL` 校验 `ttl<0`；保留 New/Get/SetDefault/CloseAll
    References RED test: TestTTL_Priority / TestWithDefaultTTL_Negative_Errors
    Verification: `go test ./ag/ag_cache/ -run 'TestTTL_Priority|TestWithDefaultTTL_Negative' -count=1`

## 5. Fx 生命周期调整（zfx_ag_cache.go）

- [x] RED: OnStop 只 Close（无 LogStats）— `ag/ag_cache/zfx_ag_cache_test.go` → `TestFxAgCacheMode_OnStopClose`
    Assertion: `fxtest.New(...)` 联合装配 RequireStart/RequireStop 通过；`LogStats` 不再被引用
    Expected failure: `LogStats` 仍存在；OnStop 引用统计
- [x] GREEN: 更新 `ag/ag_cache/zfx_ag_cache.go` — 删 `LogStats`；`registerHooks` OnStop 只 `m.Close()`；保留 group 收集/注册
    References RED test: TestFxAgCacheMode_OnStopClose
    Verification: `go test ./ag/ag_cache/ -run 'TestFxAgCacheMode' -count=1`

## 6. agristretto 适配（ristretto.go）

- [x] RED: Set 保留 ttl + Metrics 关 + DefaultTTLProvider — `ag/ag_cache/agristretto/ristretto_test.go` → `TestSet_WithTTL` / `TestFactory_DefaultTTL`
    Assertion: `Set(ctx,k,v,ttl)` 按 ttl 过期；`agristrettoFactory` 实现 `DefaultTTLProvider`
    Expected failure: `Set` 签名不匹配；Metrics 未关（无 Stats 消费）
- [x] GREEN: 更新 `ag/ag_cache/agristretto/ristretto.go` — `Set` 保留 ttl（`SetWithTTL`）；`Config.Metrics: false`；保留 `DefaultTTLProvider`；不实现 `BulkDelEngine`
    References RED test: TestSet_WithTTL / TestFactory_DefaultTTL
    Verification: `go test ./ag/ag_cache/agristretto/ -run 'TestSet_WithTTL|TestFactory_DefaultTTL' -count=1`

## 7. 核心测试迁移

- [x] RED: 迁移旧测试 — 替换 `GetAdmin`/`Stats`/`Peek` 相关断言为 `TryGet`
    Assertion: 现有核心测试用新接口编译通过
    Expected failure: 引用已移除 API（GetAdmin/Peek/Stats/Set 变参）编译失败
- [x] GREEN: 更新 `ag/ag_cache/cache_test.go`/`manager_test.go`/`zfx_ag_cache_test.go` — 全部改用 `TryGet`/`Set(ctx,k,v)`；删除 `GetAdmin`/`Stats`/`Peek` 用例
    References RED test: 全部迁移测试
    Verification: `go test ./ag/ag_cache/ -count=1`

## 8. 集成测试适配（test/）

- [x] RED: 集成测试用新接口 — `ag/ag_cache/test/` 改用 `TryGet` 替代 `GetAdmin`/`Peek`；`Set` 无变参
    Assertion: 场景/并发/探测测试用新接口编译通过
    Expected failure: `waitAdmin`/`GetAdmin`/`Peek`/`Set(...,ttl)` 引用已移除 API
- [x] GREEN: 更新 `ag/ag_cache/test/` — `waitAdmin`→`TryGet` 轮询；`TestProbe_Peek_UpdatesStats` 移除（Peek 已消解）；`TestProbe_GlobalRegistry`/`SetDroppedBufferFull` 适配 `Set` 无变参
    References RED test: 全部集成测试
    Verification: `go test ./ag/ag_cache/test/ -count=1`

## 9. 健壮性修复（方向 2 并入）

- [x] RED: GetOrElse 健壮性测试 — `ag/ag_cache/cache_test.go` → `TestGetOrElse_DoubleCheck_WithoutCancel` / `TestGetOrElse_SetFailure_ErrBackend` / `TestLoaderPanic_Labeled`
    Assertion: 首调用者 ctx 取消时 double-check 不受影响（等待者正常）；`engine.Set` 失败返回 `ErrBackend`（`errors.Is` true）；loader panic 错误含 `loader panic` 标注
    Expected failure: double-check 用调用者 ctx；Set 失败未包 ErrBackend；panic 统一标 engine panic
- [x] GREEN: 更新 `ag/ag_cache/typed.go` — `GetOrElse` double-check 用 `context.WithoutCancel(ctx)`（2.1）；`engine.Set` 失败 `errBackend(serr)`（2.2）；`recoverPanic` 区分 loader/engine panic（2.3）
    References RED test: TestGetOrElse_DoubleCheck_WithoutCancel / TestGetOrElse_SetFailure_ErrBackend / TestLoaderPanic_Labeled
    Verification: `go test ./ag/ag_cache/ -run 'TestGetOrElse_DoubleCheck|TestGetOrElse_SetFailure|TestLoaderPanic' -count=1`
- [x] RED: 负 TTL 防御测试 — `ag/ag_cache/manager_test.go` → `TestWithDefaultTTL_Negative_Errors`
    Assertion: `WithDefaultTTL(-1)` 返回错误（2.4/P6）
    Expected failure: 负 TTL 未校验
- [x] GREEN: `WithDefaultTTL` 校验 `ttl<0` 返回错误 — `ag/ag_cache/typed.go` → `WithDefaultTTL`
    References RED test: TestWithDefaultTTL_Negative_Errors
    Verification: `go test ./ag/ag_cache/ -run 'TestWithDefaultTTL_Negative' -count=1`

## 10. 收尾验证

- [x] `go build ./...`
- [x] `go vet ./ag/ag_cache/...`
- [x] `go test -race ./ag/ag_cache/...` 全绿
- [x] 同步开发文档 11（`Set` 无 ttl、`TryGet` 用法、移除 `GetAdmin`/`Peek`/`Stats` 说明）
- [x] 确认 `fxs/`、`go.work`、`release_aif.sh` 零改动
