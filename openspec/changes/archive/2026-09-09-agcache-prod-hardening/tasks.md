# ag_cache 生产加固任务（TDD）

> 验证命令：`go build ./...`；`go vet ./ag/ag_cache/...`；`goimports -l ag/ag_cache/`；`go test -race ./ag/ag_cache/...`

## 1. 入口返回 error（manager.go）

- [x] RED: 入口签名编译断言 — `ag/ag_cache/manager_test.go` → `TestEntry_ReturnsError`
    Assertion: `GetCache(m,name) (ICache[T], error)`、`GetCacheWithLoader(m,name,loader,opts...) (*LoaderCache[T], error)` 存在
    Expected failure: 无 error 返回的旧签名仍存在
- [x] RED: getOrCreate 失败 error 不 panic — `ag/ag_cache/manager_test.go` → `TestGetOrCreate_Errors`
    Assertion: mock 工厂 `Create` 返回错误 → `GetCache` 返回 `(nil, error)` 不 panic；引擎未注册（无工厂）→ 返回 `(nil, error)` 含引擎名
    Expected failure: panic 或返回 nil cache
- [x] GREEN: 重构 `ag/ag_cache/manager.go` — `GetCache`/`GetCacheWithLoader` 加 error 返回；`getOrCreate` 返回 `(*typedCache[T], error)`；两 panic → `fmt.Errorf`；已有实例二次调用返回 nil error
    References RED test: TestEntry_ReturnsError / TestGetOrCreate_Errors
    Verification: `go test ./ag/ag_cache/ -run 'TestEntry_ReturnsError|TestGetOrCreate_Errors' -count=1`

## 2. Ristretto 背压 drop 视为成功（agristretto/ristretto.go）

- [x] RED: drop 不误报 ErrBackend — `ag/ag_cache/agristretto/ristretto_test.go` → `TestDrop_BackpressureNotError`
    Assertion: 高并发写触发 setBuf 满 → Set/SetWithTTL 返回 nil（非 ErrBackend）；`Dropped()` 计数递增
    Expected failure: drop 返回错误文本 / Dropped() 未定义
- [x] RED: 真后端故障仍报错 — `ag/ag_cache/agristretto/ristretto_test.go` → `TestDrop_RealFailureStillErrors`
    Assertion: 注入真后端错误（mock 引擎 Err）→ Set 返回 error（不被 drop-as-success 吞）
    Expected failure: 真故障被静默吞掉
- [x] GREEN: 更新 `ag/ag_cache/agristretto/ristretto.go` — `setWithTTL` `!ok` → `dropped.Add(1)` + return nil；`ristrettoEngine` 加 `dropped atomic.Uint64`；加 `Dropped() uint64`
    References RED test: TestDrop_BackpressureNotError / TestDrop_RealFailureStillErrors
    Verification: `go test ./ag/ag_cache/agristretto/ -run 'TestDrop' -count=1`

## 3. GetOrElse 写失败不影响返回值（typed.go）

- [x] RED: 写失败返回数据 — `ag/ag_cache/cache_test.go` → `TestGetOrElse_WriteFailureReturnsValue`
    Assertion: loader 成功返回 v、引擎 Set 返回错误 → `GetOrElse` 返回 `(v, nil)`（写失败仅日志）
    Expected failure: 返回 error / 丢弃 v
- [x] RED: 读故障仍不调 loader — `ag/ag_cache/cache_test.go` → `TestGetOrElse_ReadFailureNoLoader`（回归）
    Assertion: engine.Get 返回非 miss 故障 → 返回 ErrBackend 不调 loader
    Expected failure: 写失败改动误伤读故障路径
- [x] GREEN: 更新 `ag/ag_cache/typed.go` — GetOrElse 写路径：`setErr != nil` → 返回 `result{v, nil}` + Warn 日志；读故障分支不变
    References RED test: TestGetOrElse_WriteFailureReturnsValue / TestGetOrElse_ReadFailureNoLoader
    Verification: `go test ./ag/ag_cache/ -run 'TestGetOrElse_WriteFailure|TestGetOrElse_ReadFailure' -count=1`

## 4. 测试迁移（新签名适配）

- [x] RED: 迁移核心测试 — `ag/ag_cache/cache_test.go`/`manager_test.go`/`zfx_ag_cache_test.go` 调用点适配 `(X, error)`
    Assertion: 全部核心测试用新入口编译通过（err 处理或忽略）
    Expected failure: 引用旧无 error 签名
- [x] GREEN: 更新 `ag/ag_cache/cache_test.go`/`manager_test.go`/`zfx_ag_cache_test.go` — 调用点改 `c, err := GetCache(...)`；`_ = err` 或断言；factory 缺失/Create 失败用例改用 error 断言
    References RED test: 全部迁移测试
    Verification: `go test ./ag/ag_cache/ -count=1`
- [x] RED: 迁移 test/ 与 usage/ — `test/setup_test.go`/`test/*`/`test/usage/service.go` 适配新签名（构造器返回 error）
    Assertion: 集成测试用 `c, err := GetCacheWithLoader(...)` 编译通过；usage 构造器返回 `(*X, error)`
    Expected failure: 引用旧签名
- [x] GREEN: 更新 `test/setup_test.go`（startFx）、`test/*`、`test/usage/service.go`（NewUserService/NewParamService 返回 error）、`usage_test.go` — 调用点适配 + 断言 err
    References RED test: 全部集成测试
    Verification: `go test ./ag/ag_cache/test/... -count=1`

## 5. 示例/文档同步

- [x] 更新 `test/usage/app.yml` — 扁平 `maxCost` 等改 `default`/`namespaces` schema（当前旧格式被绑定忽略）
- [x] 更新 README — 新签名（GetCache/GetCacheWithLoader 返回 error）、drop 背压语义、GetOrElse 写失败语义、usage 构造器 error
- [x] 更新 obsidian `2026-08-26-ag-cache-v1-final-api.md` — 入口签名 + 遗留边界 P5（drop 视为成功已实现）+ 十二节同步
- [x] 更新 obsidian `2026-08-26-ag-cache-v3-usage.md` — 入口签名 + 写/读穿透语义

## 6. 收尾验证

- [x] `go build ./...`
- [x] `go vet ./ag/ag_cache/...`
- [x] `goimports -l ag/ag_cache/` 为空
- [x] `go test -race ./ag/ag_cache/...` 全绿
- [x] 确认 `fxs/`、`go.work`、`release_aif.sh` 零改动
