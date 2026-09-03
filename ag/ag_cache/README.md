# AgCache

AgCache 是 ag-core 的通用缓存抽象库，提供类型安全、读穿透（read-through）、引擎无关的缓存 API。业务代码只依赖 `ICache[T]`，底层引擎（本地内存 Ristretto、未来 Redis 等）通过 fx group 注入，引擎实现零侵入替换。

> 版本：v1.0（定版 API）

---

## 一、设计理念

1. **简单 + 易用**：砍掉累积机制，回归最小可用。业务代码零框架概念，只面向 `ICache[T]`。
2. **类型安全**：泛型 `ICache[T]` + 序列化层，无需手动 `[]byte` 转换。
3. **引擎无关**：`Engine` SPI 收 `[]byte`，`EngineFactory` 经 fx group 注入；换引擎只改配置，业务零改动。
4. **读穿透**：`GetCacheWithLoader` 绑定 loader，miss 自动加载 + 写缓存（singleflight 去重，防缓存击穿）。
5. **TTL 归引擎**：`Engine.Set` 无 ttl 参数，默认 TTL 由引擎内部决定；业务可显式覆盖（优先级链）。

## 二、架构分层

```
逻辑抽象层（框架，引擎无关）
  ├─ ICache[T]         业务接口（Get/TryGet/GetOrElse/Set/SetWithTTL/Del/Clear）
  ├─ typedCache        内部实现（序列化/singleflight/错误语义/key 前缀拼装）
  ├─ Manager           实例注册表 + 生命周期（收集引擎工厂、config 选默认引擎）
  ├─ LoaderCache       读穿透句柄（绑 loader，薄包装不管理生命周期）
  └─ 前缀 agcache::<name>::、Clear 语义、错误契约
        │  桥：Engine 接口 + 已拼前缀的完整 key
        ▼
引擎实现层（引擎特定）
  ├─ Engine SPI        Get/Set(无ttl)/Del/Clear(ctx,prefix)/Close
  ├─ EngineFactory     Name() + Create(name)——fx group 注入 Manager
  └─ agristretto       每 name 独立 Ristretto 实例（物理自决）
```

### 核心类型

| 类型 | 角色 |
|------|------|
| `ICache[T]` | 业务接口契约（泛型，类型安全） |
| `LoaderCache[T]` | 读穿透句柄：`Get` = miss→loader→缓存→返回 |
| `typedCache[T]` | 内部实现（序列化/singleflight/前缀/错误语义） |
| `Manager` | 实例注册表 + 生命周期，`Close` 统一关引擎 |
| `Engine` | 引擎 SPI（收 `[]byte`，引擎无关） |
| `EngineFactory` | 引擎工厂（`Name()` + `Create(name)`），fx group 注入 |

## 三、TTL 优先级链

```
SetWithTTL(ctx, key, value, ttl)     —— 单条显式（最优先，ttl=0 永不过期）
> WithDefaultTTL                     —— 业务 per-cache 默认（经引擎 TTLSetter）
> engine.Set                        —— 引擎内部默认（默认 0=永不过期，defaultTtl 可配置）
```

- `Set`：配了 `WithDefaultTTL` → 经引擎 `TTLSetter` 用默认；未配 → `engine.Set`（引擎内部默认）
- `SetWithTTL`：单条显式覆盖；引擎无 `TTLSetter` → 等同 `Set`（忽略 ttl）
- 引擎内部默认：**默认 0（永不过期）**，agristretto 配置 `defaultTtl` 可显式设过期

## 四、快速开始

### 1. Fx 装配

```go
import (
    ag_cache "github.com/aif-go/ag-core/ag/ag_cache"
    "github.com/aif-go/ag-core/ag/ag_cache/agristretto"
    "github.com/aif-go/ag-core/fxs"
)

app := fx.New(fx.Module("app",
    fxs.FxAgConfModule,                 // 配置绑定（IBinder）
    ag_cache.FxAgCacheMode,             // core：收集引擎 group → 建 Manager → 生命周期
    agristretto.FxAgCacheRistrettoMode, // 引擎：Provide ristretto 工厂到 group
    // 业务模块...
))
```

### 2. 配置（app.yml）

```yaml
agcache:                    # core：选默认引擎（config 选默认）
  defaultEngine: ristretto
  ristretto:                # 引擎配置：default（全局限量）+ namespaces（per-name 覆盖）
    default:
      maxCost: 104857600    # 内容预算（字节），缺省 100MB；空缓存不占，多 name 按写入叠加
      numCounters: 131072   # TinyLFU 频率 sketch（淘汰精度，非容量上限）；缺省 100K 档
      bufferItems: 64
      defaultTtl: 60s       # 引擎默认 TTL：缺省/0=永不过期；"60s"=60 秒；非法字符串装配期报错
    namespaces:             # per-name 覆盖（非零覆盖继承 default）——特殊大缓存用
      users:
        maxCost: 1073741824      # 大热缓存：只覆盖需要的字段，其余继承 default
        numCounters: 8388608     # 10 万+ key 场景显式调大（配 2 的幂）
      params:
        defaultTtl: 30s
```

> **默认值设计**：`numCounters` 默认 100K 档（sketch ~0.26MB/实例）——多数缓存 name（参数/配置/枚举，key 有限）永不淘汰，小 sketch 足够且省预分配内存；NumCounters 只影响淘汰精度、不影响容量（MaxCost 定）与正确性。特殊大缓存（接口热缓存，10 万+ key）用 `namespaces` per-name 覆盖。
>
> **启动校验**：配置在装配期预解析校验——`default` 与**所有** `namespaces` 条目逐一校验（非法 TTL/负值 → fx 启动失败，含 name 定位），配置错误启动即死，非运行时 panic。
>
> 多引擎组合：加 `redis.FxAgCacheRedisMode` → `defaultEngine: redis` 即切，Manager 零改动。默认引擎未注册 → 装配 fail-fast。

### 3. 业务代码（构造注入 Manager，绑定一次）

```go
type UserService struct {
    users *ag_cache.LoaderCache[*User]
    repo  *UserRepo
}

// 构造时绑定一次（fx 注入 *Manager → 依赖拓扑保证 Manager 先建）
func NewUserService(m *ag_cache.Manager, repo *UserRepo) *UserService {
    return &UserService{repo: repo, users: ag_cache.GetCacheWithLoader(m, "users", repo.GetUser)}
}

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    return s.users.Get(ctx, id)   // 读穿透：miss → repo.GetUser → 缓存
}
func (s *UserService) RefreshUser(ctx context.Context, id string) error {
    return s.users.Del(ctx, id)
}
```

### 4. 非 Fx / 运行时获取

```go
ag_cache.SetDefault(mgr)                // 装配时设置默认
...
m := ag_cache.DefaultManager()          // 运行时获取
users := ag_cache.GetCacheWithLoader[*User](m, "users", repo.GetUser)
c := ag_cache.GetCache[string](m, "users")
```

## 五、ICache 方法

| 方法 | 语义 |
|------|------|
| `Get(ctx, key) (T, error)` | 纯读，miss → `ErrCacheMiss`（不调 loader） |
| `TryGet(ctx, key) (T, bool, error)` | 宽松读，miss → `(zero, false, nil)`；后端故障 → `(zero, false, err)` |
| `GetOrElse(ctx, key, loader) (T, error)` | 读穿透：miss → loader → 写缓存 → 返回 |
| `Set(ctx, key, value)` | 写，TTL 用 namespace 默认（`WithDefaultTTL` 或引擎内部默认） |
| `SetWithTTL(ctx, key, value, ttl)` | 单条显式 TTL（最高优先级；`ttl=0` 永不过期；引擎无 TTLSetter 等同 Set） |
| `Del(ctx, keys...)` | 批量删（引擎支持则 `DelMany`） |
| `Clear(ctx)` | 清整个实例（只影响本 name） |

**错误契约**：
- `errors.Is(err, ag_cache.ErrCacheMiss)` — miss，**唯一**触发 loader
- `errors.Is(err, ag_cache.ErrBackend)` — 后端故障，**绝不**触发 loader（防缓存击穿）

> ⚠️ **Set 异步可见**（Ristretto 异步写）：`Set` 后立即 `Get` 同 key 可能 miss（微秒级窗口）。需要"写后立即可读"用 `GetOrElse`（内部 sync）或接受短窗口。

## 六、常用写法

### 读缓存（推荐绑定 loader）
```go
users := ag_cache.GetCacheWithLoader(m, "users", userRepo.GetUser)
u, err := users.Get(ctx, "u:1")           // 读穿透
```

### 纯读 / 存在性检查
```go
c := ag_cache.GetCache[*User](m, "users") // 不绑 loader
u, err := c.Get(ctx, "u:1")               // miss → ErrCacheMiss
v, ok, err := c.TryGet(ctx, "u:1")         // miss → ok=false, 无 error
```

### 写 / 删 / 清
```go
users.Set(ctx, "u:1", user)                          // 用默认 TTL
users.SetWithTTL(ctx, "u:1", user, time.Minute)      // 单条显式 TTL
users.Del(ctx, "u:1", "u:2")                         // 批量删
params.Clear(ctx)                                    // 参数变更 → 清空 params（不影响 users）
```

### TTL 覆盖（构造期）
```go
params := ag_cache.GetCacheWithLoader(m, "params", paramCenter.Get,
    ag_cache.WithDefaultTTL(30*time.Second))   // 业务 per-cache 默认，经引擎 TTLSetter
```

### 多服务共享（显式 loader 包装）
```go
func NewUserService(m *ag_cache.Manager) *UserService {
    return &UserService{users: ag_cache.WithLoader(ag_cache.GetCache[*User](m, "users"), repo.GetUser)}
}
```

## 七、业务场景

### 用户缓存（Cache-Aside）
```go
users := ag_cache.GetCacheWithLoader(m, "users", userRepo.GetUser)
u, err := users.Get(ctx, "u:1")       // 缓存 60s（默认）
// 用户更新/删除时失效：
users.Del(ctx, "u:1")
```

### 参数缓存（批量失效）
```go
params := ag_cache.GetCacheWithLoader(m, "params", paramCenter.Get)
p, err := params.Get(ctx, "host:port")
// 参数系统更新 → 广播：
params.Clear(ctx)                      // 只清 params，不影响 users
```

### 负缓存（穿透防护，可选）
```go
notExist := ag_cache.GetCacheWithLoader(m, "user-notexist",
    func(ctx context.Context, key string) (bool, error) { return false, ag_cache.ErrCacheMiss })
users := ag_cache.GetCacheWithLoader(m, "users", func(ctx context.Context, key string) (*User, error) {
    if notFound(key) {
        notExist.Set(ctx, key, true)   // 记录"不存在"
        return nil, ag_cache.ErrCacheMiss
    }
    return repo.Get(ctx, key)
})
// 查询时先查负缓存：
if _, err := notExist.Get(ctx, key); err == nil {
    return nil, ag_cache.ErrCacheMiss  // 短路，不查库
}
u, err := users.Get(ctx, key)
```

## 八、引擎扩展（EngineFactory + fx group）

引擎是一个 `Engine` 实现 + 一个 `EngineFactory`。核心通过 fx group `"agcache.engine"` 收集所有工厂，config `defaultEngine` 选默认。

```go
// 引擎 SPI（收 []byte，引擎无关）
type Engine interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte) error  // 引擎内部默认 TTL
    Del(ctx context.Context, key string) error
    Clear(ctx context.Context, prefix string) error
    Close() error
}

// 可选能力
type TTLSetter interface {  // 支持显式 per-entry TTL
    SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error
}
type BulkDelEngine interface { DelMany(ctx context.Context, keys ...string) error }

// 引擎工厂（fx group 注入 Manager）
type EngineFactory interface {
    Name() string
    Create(name string) (Engine, error)  // name=缓存名命名空间上下文；每次新实例
}
```

```go
// 引擎模块 Provide 工厂进 group（参考 agristretto/zfx_agristretto.go）
var FxMyCacheEngineMode = fx.Module("ag_cache.mycache",
    fx.Provide(
        fx.Annotate(NewMyCacheFactory, fx.ResultTags(`group:"agcache.engine"`)),
    ),
)
```

- 每 `Create(name)` 返回**独立实例**（agristretto 每 name 独立 Ristretto），隔离天然
- agristretto 工厂 `Create(name)` 查 per-name 配置：`namespaces.<name>` 命中用其覆盖配置，未命中用 `default`（YAML 见上）
- 实例复用与生命周期归 `Manager`（懒建复用，`Manager.Close` 统一关）
- key 前缀 `agcache::<name>::` 由 `typedCache` 拼装，`Engine` 零 name 概念

## 九、测试

```go
// 单元测试：纯内存 mock，零依赖（实现 ICache）
cache := ag_cache.NewMock[*User]()
cache.GetOrElse(ctx, "u:1", loader)

// 显式引擎（底层/隔离测试，NewWithEngine 保留）
c := ag_cache.NewWithEngine[string](ag_cache.NewMockEngine())

// 集成（真实 Ristretto）：参考 ag/ag_cache/test/ 的 startFx 装配
```

## 十、要点提醒

- **一个缓存名对应一种类型**：`GetCacheWithLoader[*User](m, "users", ...)` 后同 name 不能用 `*Session`（panic）
- **TTL 优先级链**：`SetWithTTL(ttl)` > `WithDefaultTTL`（业务 per-cache 默认，经引擎 TTLSetter）> `engine.Set`（引擎内部默认，默认 0=永不过期，`defaultTtl` 可配置）
- **默认引擎由 config 选**：`agcache.defaultEngine`（fx group 注入多种引擎，config 选默认；默认引擎未注册 → 装配 fail-fast）
- **独立实例隔离**：`params.Clear()` 不影响 `users`；每个缓存名独立引擎实例
- **Set 异步**：写后立即可读不保证，读穿透用 `GetOrElse`
- **opts 冲突校验**：首个 `GetCacheWithLoader` 定配置（同 name 复用实例），不同数据源用不同缓存名

## 关联

- 引擎实现：`ag/ag_cache/agristretto/`（Ristretto 本地引擎）
- 使用示例：`ag/ag_cache/test/usage/`（完整业务用法集成测试）
- 单元/集成测试：`ag/ag_cache/test/`
