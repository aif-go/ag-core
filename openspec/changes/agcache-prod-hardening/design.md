# ag_cache 生产加固设计

## Context

- **现状**：`getOrCreate` 引擎创建失败 → panic（请求路径可能崩进程）；Ristretto `setBuf` 满被误报 ErrBackend（背压≠故障）；`GetOrElse` loader 成功后缓存写失败返回 error（丢弃新鲜数据）。
- **目标**：错误用 error 显式上报（替代 panic）；背压 drop 视为成功（计数可观测）；读穿透以数据为准。仓库无 ag_cache 外部业务调用（仅 test/usage），签名改动处于窗口期。

## Goals / Non-Goals

**Goals:**
- 入口 `GetCache`/`GetCacheWithLoader` 返回 `(T, error)`；`getOrCreate` 两处 panic → error
- Ristretto drop（setBuf 满）→ nil + `dropped` 计数；真后端故障仍 error
- `GetOrElse` loader 成功但缓存写失败 → 返回 `(v, nil)`
- 示例/文档同步（app.yml schema + README + 定版 API）

**Non-Goals:**
- GOMEMLIMIT/进程内存兜底（属应用/部署层职责，非 ag_cache）
- 缓存写失败重试/补偿（读穿透写是尽力而为，数据源为准）
- failedEngine 降级类型（返回 error 直接根治，无需注入坏引擎）
- 失败后自动重建引擎（失败是永久的，业务重建 Manager 或重启）

## Decisions

### D1. 入口返回 error（替代 panic）
```go
// manager.go
func GetCache[T any](m *Manager, name string) (ICache[T], error)
func GetCacheWithLoader[T any](m *Manager, name string, loader LoaderFunc[T], opts ...Option[T]) (*LoaderCache[T], error)

func getOrCreate[T any](m *Manager, name string, opts ...Option[T]) (*typedCache[T], error) {
    // 已有实例 → 返回
    // 引擎工厂缺失 → return nil, fmt.Errorf("agcache: engine %q not registered", ...)
    // Create(name) 失败 → return nil, fmt.Errorf("agcache: create engine for %q: %w", name, err)
}
```
- nil Manager 参数仍 panic（programming error，非运行期失败）
- 已有缓存实例的二次调用：直接返回 nil error（懒创建只首次失败）

### D2. Ristretto 背压 drop 视为成功
```go
// ristretto.go
type ristrettoEngine struct {
    cache      *ristretto.Cache[string, []byte]
    defaultTTL time.Duration
    dropped    atomic.Uint64 // 背压丢弃计数（setBuf 满）
}

func (e *ristrettoEngine) setWithTTL(key string, value []byte, ttl time.Duration) error {
    cost := int64(len(value)); if cost < 1 { cost = 1 }
    if !e.cache.SetWithTTL(key, value, cost, ttl) {
        e.dropped.Add(1)   // 背压 drop → 视为成功（等同容量淘汰），计数可观测
        return nil
    }
    return nil
}

func (e *ristrettoEngine) Dropped() uint64 { return e.dropped.Load() }
```
- Ristretto 语义：drop 仅发生在新 key 且 setBuf 满（cache.go:354-367）；已存在 key 的 update 即使 setBuf 满也返回 true
- 真后端故障（如 mock 引擎 Err 注入）走 Engine 接口原路径，不经 setWithTTL，不被吞

### D3. GetOrElse 写失败不影响返回值
```go
// typed.go —— GetOrElse 的 singleflight 写路径
v, lerr := loader(...)          // loader 成功
data, serr := serializer.Marshal(v)
if serr != nil { return ... }   // 序列化失败仍报错（数据不可用）
var setErr error
if c.ttlSet { setErr = c.setWithTTL(...) } else { setErr = c.engine.Set(...) }
if setErr != nil {
    // 写缓存失败：尽力而为，不丢弃已加载数据
    slog.Warn("agcache: cache write failed", "key", key, "err", setErr)
    return result{v, nil}, nil    // 返回数据，nil error
}
if s, ok := c.engine.(syncer); ok { s.Sync() }
return result{v, nil}, nil
```
- 读路径真故障（engine.Get 非 miss）仍 ErrBackend 不调 loader——现有逻辑不变
- 序列化失败保留 error（值无法存也没法返回一致对象？实际 unmarshal 才是读；Marshal 失败是数据 bug，仍报）

> 复核：Marshal 失败发生在 loader 成功后、写缓存前。值 v 已可用。Marshal 失败返回 error 是否也丢弃 v？为一致性，Marshal 失败也可返回 (v, 缓存侧错误)。但 Marshal 失败通常是业务数据不可序列化（配置错误），倾向仍返回 error 暴露问题。若想最大化"数据为准"，Marshal 失败也返回 v。——实施时确认，倾向 Marshal 失败仍报错（暴露配置 bug）。

### D4. 示例/文档同步
- `test/usage/app.yml`：扁平 `maxCost` 直接挂 ristretto 下 → 改 `default`/`namespaces` schema（当前被绑定器忽略，走默认）
- service.go 构造器：`NewUserService/NewParamService` 返回 `(*X, error)`（GetCacheWithLoader 现在返回 error）
- README + obsidian 文档：新签名、drop 背压、GetOrElse 写失败语义

## 函数签名清单

| 文件 | 符号 | 签名变化 |
|------|------|---------|
| `manager.go` | `GetCache[T](m, name)` | `ICache[T]` → `(ICache[T], error)` |
| `manager.go` | `GetCacheWithLoader[T](m, name, loader, opts...)` | `*LoaderCache[T]` → `(*LoaderCache[T], error)` |
| `manager.go` | `getOrCreate[T](m, name, opts...)` | `*typedCache[T]` → `(*typedCache[T], error)`；两 panic → error |
| `ristretto.go` | `ristrettoEngine` | 加 `dropped atomic.Uint64` |
| `ristretto.go` | `setWithTTL` | `!ok` → nil + `dropped.Add(1)`（不再返回 ErrBackend 文本错误） |
| `ristretto.go` | `(e *ristrettoEngine) Dropped() uint64` | 新增（可观测） |
| `typed.go` | `GetOrElse` 写路径 | `setErr != nil` → 返回 `(v, nil)` + Warn 日志（不再返回错误） |

## 测试策略

| 层 | 文件 | 覆盖 |
|----|------|------|
| Manager 单元 | `manager_test.go` | GetCache/GetCacheWithLoader 正常/未注册/Create 失败返回 error 不 panic；懒创建复用 |
| 引擎单元 | `agristretto/ristretto_test.go` | 高并发写触发 drop → 返回 nil 不 ErrBackend + Dropped() 递增；真故障（mock Err）仍报错 |
| core 单元 | `cache_test.go`/`typed.go` 相关 | GetOrElse loader 成功 + 缓存写失败 → 返回 (v, nil)；读故障仍不调 loader |
| 集成 | `test/*`、`test/usage/*` | usage 构造器返回 error；startFx 装配回归；app.yml 新 schema |
| 基准 | `benchmark_test.go`/`bench_test.go` | 适配新签名（编译） |

验证：`go build ./...`、`go vet ./ag/ag_cache/...`、`goimports -l ag/ag_cache/`、`go test -race ./ag/ag_cache/...`。

## Risks / Trade-offs

- [R1] **入口签名 breaking**：`GetCache`/`GetCacheWithLoader` 加 error——仓库无外部调用，窗口期；对外若已发布需版本迁移
- [R2] **drop 静默**：背压丢弃不报错，业务 Set 后缓存可能未写入且无提示——与容量淘汰一致（缓存写本不保证），Dropped() 计数兜底可观测
- [R3] **GetOrElse 写失败静默**：缓存写失败不再可见于返回值——Warn 日志 + 后续计数；业务强依赖"写缓存必成功"的场景需用显式 Set 检查（但框架语义是尽力而为）

## Migration Plan

按 change 分层实施：manager.go（入口 error）→ ristretto.go（drop）→ typed.go（GetOrElse）→ 测试迁移（manager/cache/zfx/usage）→ 示例/文档 → 全量验证。

## Open Questions

- D3 中序列化失败处理：返回 (v, nil)（最大化数据可用）还是保留 error（暴露数据 bug）？实施时倾向保留 error——Marshal 失败通常是不能序列化的数据（配置/类型问题），应显式暴露而非静默。
