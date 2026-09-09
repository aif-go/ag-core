# ag_cache 生产加固：入口 error 化 + 背压语义 + 读穿透写失败修正

## Why

生产就绪评估发现 3 个语义问题：
1. `getOrCreate` 引擎创建失败 → `panic`：请求路径懒建缓存时，引擎 Create 失败（如未来 Redis 连接失败）会 panic，若无框架 recover 则崩进程。`GetCache`/`GetCacheWithLoader` 无 error 返回，失败无上报通道。
2. Ristretto `setBuf` 满（写背压）被误报 `ErrBackend`：写密集批量缓存时业务 `errors.Is(ErrBackend)` 误降级/误告警，实为临时背压非故障。
3. `GetOrElse` loader 成功但缓存写失败 → 返回 error：业务丢弃已加载的新鲜数据走降级，读穿透以数据为准的原则被破坏。

## What Changes

- **入口错误语义（错误上报，替代 panic）**：
  - `GetCache[T](m, name) ICache[T]` → **`(ICache[T], error)`**
  - `GetCacheWithLoader[T](m, name, loader, opts...) *LoaderCache[T]` → **`(*LoaderCache[T], error)`**
  - `getOrCreate` 返回 `(*typedCache[T], error)`；引擎未注册与 `Create(name)` 失败两处 **panic → error**
  - **BREAKING**：公开 API 签名变化（当前仓库无 ag_cache 外部业务调用，窗口期改）
- **Ristretto 背压 drop 视为成功（消除 ErrBackend 误报）**：
  - `ristrettoEngine.setWithTTL`：`setBuf` 满（`SetWithTTL` 返回 false）→ 返回 nil（等同容量淘汰的背压），不再 `ErrBackend`
  - 引擎维护 `dropped` 计数（`atomic.Uint64`）+ 导出 `Dropped()` 供可观测
  - 真后端故障（engine.Err 等）仍 `ErrBackend`，语义不变
- **GetOrElse 写失败不影响返回值**：
  - loader 成功但缓存写失败 → 返回 `(v, nil)`（读穿透以数据为准，缓存写尽力而为），写失败仅日志/计数
  - 真后端读故障（engine.Get 非 miss）仍 `ErrBackend` 不调 loader（防击穿不变）
- **示例/文档同步**：
  - `test/usage/app.yml` 改 `default`/`namespaces` schema（当前旧扁平 schema 实际被绑定忽略）
  - README + 定版 API 文档更新新签名、drop 背压语义、GetOrElse 写失败语义

## Capabilities

### New Capabilities
- `agcache-prod-error`: 生产错误语义——入口 `(T, error)` 替代 panic；Ristretto 背压 drop 视为成功（dropped 计数）；GetOrElse 写失败返回已加载值

### Modified Capabilities
<!-- 无：openspec/specs/ 尚无现有 spec -->

## Impact

- **修改文件**：
  - `ag/ag_cache/manager.go`：`GetCache`/`GetCacheWithLoader`/`getOrCreate` 返回 error；删两处 panic
  - `ag/ag_cache/agristretto/ristretto.go`：`setWithTTL` drop → nil；`dropped` 计数 + `Dropped()`；真故障仍 ErrBackend
  - `ag/ag_cache/typed.go`：`GetOrElse` 写失败 → 返回 `(v, nil)`（记日志/计数）
- **测试迁移**：`manager_test.go`/`cache_test.go`/`zfx_ag_cache_test.go`（47 处调用点适配 error）、`test/usage/service.go`（构造器返回 error）、`test/*`、`benchmark_test.go`
- **测试新增**：drop→nil+dropped 计数、GetOrElse 写失败返回 v、getOrCreate 失败 error（不 panic）、引擎未注册 error
- **文档**：README、obsidian `2026-08-26-ag-cache-v1-final-api.md`/`2026-08-26-ag-cache-v3-usage.md` 同步签名/语义
- **无破坏性变更**（对 `fxs/`、`go.work`、`release_aif.sh`、其他 ag_* 包）
- **破坏性影响（相对当前 hzw_cache_dev）**：`GetCache`/`GetCacheWithLoader` 签名加 error；Ristretto drop 行为从 ErrBackend 变 nil
