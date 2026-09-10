# AgCache v1.0 定版 API 实施任务（TDD）

> 验证命令：`go test -race ./ag/ag_cache/...`；`go vet ./ag/ag_cache/...`；`go build ./...`
> 参考：定版 API 文档 `/home/houzw/document/hzw-obsidian/ag-core/ag_cache/2026-08-26-ag-cache-v1-final-api.md`

## 1. Engine SPI 重构（engine.go）

- [x] RED: Engine SPI 编译断言 — `ag/ag_cache/engine_test.go` → `TestEngine_SpiCompile`
    Assertion: `Engine.Set(ctx,k,v)`（无 ttl 参数）、`Engine.Clear(ctx,prefix)`、可选 `TTLSetter` 存在；`DefaultTTLProvider`/`RegisterEngine`/`EngineRegistered`/`getFactory` 不存在
    Expected failure: 旧 SPI（Set 带 ttl、注册表）仍存在
- [x] GREEN: 重构 `ag/ag_cache/engine.go` — `Set(ctx,k,v)` 无 ttl；`Clear(ctx,prefix)`；加 `TTLSetter`；删 `DefaultTTLProvider`/`RegisterEngine`/`EngineRegistered`/`getFactory`/`EngineFactory.Name()` 注册表
    References RED test: TestEngine_SpiCompile
    Verification: `go build ./ag/ag_cache/`

## 2. 入口重构（manager.go）

- [x] RED: 新入口编译断言 — `ag/ag_cache/manager_test.go` → `TestEntry_InterfaceCompile`
    Assertion: `GetCacheWithLoader[T](m,...)`/`GetCache[T](m,name)`/`DefaultManager()` 存在；v3 `New`/`Get`（包级 default）不存在
    Expected failure: 旧入口存在、新入口缺失
- [x] GREEN: 重构 `ag/ag_cache/manager.go` — `New`→`GetCacheWithLoader(m,...)`、`Get`→`GetCache(m,...)`、新增 `DefaultManager()`；getOrCreate 用默认引擎工厂 `Create(name)`
    References RED test: TestEntry_InterfaceCompile
    Verification: `go build ./ag/ag_cache/`

## 3. 引擎模型：fx group + config 选默认（manager.go + zfx_ag_cache.go）

- [x] RED: 引擎模型测试 — `ag/ag_cache/zfx_ag_cache_test.go` → `TestEngineModel_GroupDefault`
    Assertion: Manager 收集多种 `EngineFactory`（group），config `defaultEngine` 选默认；`SetEngineFactory(name,f)` 填 map；fail-fast（默认引擎未注册 → 报错）
    Expected failure: 注册表式（RegisterEngine）或单工厂注入，无 config 选默认
- [x] GREEN: 更新 `ag/ag_cache/manager.go` — `Manager{engineFactories map[string]EngineFactory; defaultEngine string}`；`SetEngineFactory`/`DefaultEngine`；`getOrCreate` 用 `m.engineFactories[m.defaultEngine].Create(name)`
    References RED test: TestEngineModel_GroupDefault
    Verification: `go test ./ag/ag_cache/ -run TestEngineModel_GroupDefault -count=1`
- [x] GREEN: 更新 `ag/ag_cache/zfx_ag_cache.go` — `NewAgCacheManager(p EngineFactoryParams, props)` 消费 group 填 map + config 选默认 + fail-fast；删注册幂等逻辑
    References RED test: TestEngineModel_GroupDefault
    Verification: `go build ./ag/ag_cache/`

## 4. TTL 语义（cache.go + typed.go）

- [x] RED: ICache.SetWithTTL 编译断言 — `ag/ag_cache/cache_test.go` → `TestICache_SetWithTTL`
    Assertion: `ICache[T].SetWithTTL(ctx,key,value,ttl)` 存在
    Expected failure: ICache 无 SetWithTTL
- [x] GREEN: 更新 `ag/ag_cache/cache.go` — `ICache[T]` 加 `SetWithTTL`
    References RED test: TestICache_SetWithTTL
    Verification: `go build ./ag/ag_cache/`
- [x] RED: TTL 优先级链测试 — `ag/ag_cache/cache_test.go` → `TestTTL_PriorityChain`
    Assertion: `SetWithTTL` 显式 ttl 最高优先级；`WithDefaultTTL` 配了 → `Set` 经 `TTLSetter` 用默认；未配 → `engine.Set`；ttl=0 永不过期；引擎不实现 `TTLSetter` → 等同 `Set`
    Expected failure: Set 传默认 ttl 给 engine.Set、无 TTLSetter 探测
- [x] GREEN: 更新 `ag/ag_cache/typed.go` — `Set`：`ttlSet` → 探测 `TTLSetter` 用默认，未配 → `engine.Set`；`SetWithTTL`：探测 `TTLSetter` 传显式 ttl，不实现 → `engine.Set`；删 `WithDefaultTTL` 直接传 engine.Set
    References RED test: TestTTL_PriorityChain
    Verification: `go test ./ag/ag_cache/ -run TestTTL_PriorityChain -count=1`

## 5. key 前缀方案 A（typed.go + engine.go）

- [x] RED: 前缀拼装测试 — `ag/ag_cache/manager_test.go` → `TestPrefix_KeyJoin`
    Assertion: name="users"、key="u:1" 时 engine 收到 `agcache::users::u:1`；singleflight/DelMany 也拼前缀
    Expected failure: engine 收到原始 key，未拼前缀
- [x] GREEN: 更新 `ag/ag_cache/typed.go` — typedCache 加 `name`/`prefix`；Get/TryGet/GetOrElse/Set/SetWithTTL/Del 拼 `prefix+key`；singleflight key 拼前缀；`Clear` 调 `engine.Clear(ctx, prefix)`
    References RED test: TestPrefix_KeyJoin
    Verification: `go test ./ag/ag_cache/ -run TestPrefix_KeyJoin -count=1`
- [x] RED: Engine.Clear 签名编译断言 — `ag/ag_cache/engine_test.go` → `TestEngine_ClearPrefix`
    Assertion: `Engine.Clear(ctx, prefix string)` 签名存在
    Expected failure: 旧 `Clear(ctx)` 无 prefix
- [x] GREEN: 更新 `ag/ag_cache/engine.go` — `Clear(ctx, prefix string)`
    References RED test: TestEngine_ClearPrefix
    Verification: `go build ./ag/ag_cache/`

## 6. 引擎适配（agristretto + mock）

- [x] RED: 引擎内部默认 TTL + SetWithTTL + Create(name) — `ag/ag_cache/agristretto/ristretto_test.go` → `TestEngine_InternalDefaultTTL` / `TestEngine_SetWithTTL` / `TestFactory_CreateName`
    Assertion: `Set` 用内部 `defaultTTL`（config `defaultTtl` 喂入）；`SetWithTTL` 显式 ttl；`agristrettoFactory.Create("users")` 新实例；`Clear(ctx,prefix)` 忽略 prefix 清实例；Set/Get 用完整 key（`agcache::users::u:1`）正常
    Expected failure: Set 带 ttl 参数、无 SetWithTTL、Create() 无参
- [x] GREEN: 更新 `ag/ag_cache/agristretto/ristretto.go` — `ristrettoEngine` 加 `defaultTTL` 字段（构造喂入）；`Set` 用内部默认；`SetWithTTL`；`Clear(ctx,prefix)` 忽略 prefix；`agristrettoFactory{Name();Create(name)}`；删 `DefaultTTL()`
    References RED test: TestEngine_InternalDefaultTTL / TestEngine_SetWithTTL / TestFactory_CreateName
    Verification: `go test ./ag/ag_cache/agristretto/ -run 'TestEngine_InternalDefaultTTL|TestEngine_SetWithTTL|TestFactory_CreateName' -count=1`
- [x] GREEN: 更新 `ag/ag_cache/mock.go` — `MockEngine` 适配 `Set(ctx,k,v)` 无 ttl、`Clear(ctx,prefix)`；可选实现 `SetWithTTL`
    References RED test: TestEngine_SpiCompile
    Verification: `go build ./ag/ag_cache/`

## 7. 核心测试迁移

- [x] RED: 迁移旧测试到新入口/新 SPI — 替换 `New`/`Get`（包级）为 `GetCacheWithLoader(m,...)`/`GetCache(m,...)`；适配 Set 无 ttl、WithEngine 移除
    Assertion: 现有核心测试用新入口编译通过
    Expected failure: 引用已移除的包级 New/Get、WithEngine
- [x] GREEN: 更新 `ag/ag_cache/cache_test.go`/`manager_test.go`/`zfx_ag_cache_test.go` — setup 注入 Manager，用 `GetCacheWithLoader`/`GetCache`；删 WithEngine/DefaultTTLProvider 相关测试；加 SetWithTTL/TTLSetter/Create(name) 测试
    References RED test: 全部迁移测试
    Verification: `go test ./ag/ag_cache/ -count=1`

## 8. 集成/usage 测试迁移

- [x] RED: 迁移 test/ 与 usage/ — `ag/ag_cache/test/*`、`ag/ag_cache/test/usage/*` 用新入口/新 SPI
    Assertion: 集成测试用 `GetCacheWithLoader(m,...)`/`GetCache(m,...)` 编译通过
    Expected failure: 引用旧入口
- [x] GREEN: 更新 `test/setup_test.go`（startFx 暴露 Manager）、`test/*`、`test/usage/*` — 构造注入 Manager + `GetCacheWithLoader` 绑定一次
    References RED test: 全部集成测试
    Verification: `go test ./ag/ag_cache/test/... -count=1`

## 9. 收尾验证

- [x] `go build ./...`
- [x] `go vet ./ag/ag_cache/...`
- [x] `go test -race ./ag/ag_cache/...` 全绿
- [x] 同步文档 `2026-08-26-ag-cache-v3-usage.md`（API 名更新：GetCacheWithLoader/GetCache/SetWithTTL + 显式 Manager 用法）
- [x] 确认 `fxs/`、`go.work`、`release_aif.sh` 零改动
