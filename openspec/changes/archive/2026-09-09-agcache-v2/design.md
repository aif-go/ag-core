# AgCache v2 实现设计

## Context

- **现状**：ag-core 无统一进程内缓存抽象。POC `agcache-poc` v3.3 已验证"CacheManager + 独立命名空间实例"模型（27 测试全绿）。v1 落地（`agcache-implement`）在评审后暴露架构问题，已整体回退。
- **目标框架**：`github.com/aif-go/ag-core`。core 纯 SPI 入 `ag/ag_cache`，Ristretto 引擎入 `ag/ag_cache/agristretto`，配置接入 `ag_conf` Binder，Fx 经 group 动态注入。
- **设计来源**：`AgCache设计讨论/10-正式实现详细设计.md`（v1 设计）+ v1 评审教训 + 本设计（v2 演进）。
- **约束**：`ag/ag_cache` 顶层包零引擎依赖（Ristretto 只出现在 agristretto 子包）；引擎子包单向依赖 core（core 不得 import 引擎，循环依赖禁止）；业务 API 形态不变。

## Goals / Non-Goals

**Goals:**
- core 纯 SPI：`Engine`/`EngineFactory{Create() 无参}`/`DefaultTTLProvider`/注册表/包级实例管理/`typedCache`
- 引擎经 fx group（`group:"agcache.engine"`）动态注入 core，core 统一幂等注册，引擎不直接调 RegisterEngine
- TTL 归引擎（`DefaultTTLProvider` 自声明，core 兜底 5min，业务 `WithDefaultTTL` 覆盖）
- core 配置仅 `AgCacheProperties{DefaultEngine}`（只选引擎），引擎配置自绑 `agcache.ristretto.*`
- 迁移全部 POC 回归测试（core mock / agristretto 真实 Ristretto 拆分）

**Non-Goals:**
- Redis 引擎（未来独立里程碑）
- 空值缓存 / 布隆过滤器 / 随机 TTL jitter（穿透/雪崩不内置，负缓存由业务承担）
- Prometheus 指标注册（后置）
- v1 review 遗留 P1 项：`GetOrElse` Set 失败错误包装、`Peek` 统计语义、loader panic 标注（另开任务）

## Decisions

### D1. 引擎经 fx group 动态注入（替代引擎直接 RegisterEngine）
v1 由引擎子包 `init()` 自注册或 Fx Invoke 直接调 `RegisterEngine`。v2 改为：
```go
// core：收集 group，统一注册（幂等）
type EngineFactoryParams struct {
    fx.In
    Factories []EngineFactory `group:"agcache.engine"`
}
func NewAgCacheManager(p EngineFactoryParams, props *AgCacheProperties) (*Manager, error) {
    for _, f := range p.Factories {
        if !EngineRegistered(f.Name()) { RegisterEngine(f) }
    }
    return NewManager(props)
}
// agristretto：只 Provide 到 group
fx.Provide(fx.Annotate(ProvideAgristrettoFactory, fx.ResultTags(`group:"agcache.engine"`)))
```
**理由**：local→core 单向依赖合法（引擎只 import core 类型）；core 零引擎认知（只认 EngineFactory 接口）；多引擎可扩展（redis 加一个模块 Provide 到同一 group）；注册顺序/幂等/冲突由 core 统一控制。对标 agslog 的 `group:"agslog.handler"` 模式。
**替代**：引擎直接 `RegisterEngine`——core 失去注册控制，多引擎冲突靠 init 副作用；core import 引擎子包——循环依赖不可行。

### D2. EngineFactory.Create() 无参（引擎配置自持）
```go
type EngineFactory interface {
    Name() string
    Create() (Engine, error)   // 无参
}
type agristrettoFactory struct { cfg RistrettoConfig; ttl time.Duration }
func (f agristrettoFactory) Create() (ag_cache.Engine, error) { return NewRistrettoEngine(f.cfg) }
```
**理由**：v1 的 `Create(cfg any)` 让 core 透传引擎配置（`EngineConfig any`），弱类型且引擎认知泄漏进 core。无参后引擎实例化参数是引擎内部事务，工厂构造时注入（经 `ProvideAgristrettoFactory` 绑定配置）。
**替代**：`Create(cfg any)`——保留 EngineConfig 透传，SPI 不纯净。

### D3. TTL 归引擎（DefaultTTLProvider 可选接口）
```go
type DefaultTTLProvider interface { DefaultTTL() time.Duration }
```
- 引擎实现该接口 → core 用引擎声明的默认 TTL；未实现 → core 兜底 5min；业务 `WithDefaultTTL` 最高优先级。
- 不是所有引擎都有 TTL 属性，core 不做统一假设；`Engine.Set(ctx, key, value, ttl)` 的 ttl 参数仍是统一 SPI 契约（`0`=永不过期），引擎决定是否支持。
**理由**：v1 把 defaultTtl 放 core 配置或经 TTL group 传递，都是错误抽象；TTL 是引擎特性，由引擎自声明最自然。
**替代**：TTL 放 core 配置——对所有引擎做统一 TTL 假设；TTL 经独立 fx group 传递——为单一标量引入 group 机制，过度设计。

### D4. core 配置仅 DefaultEngine（只选引擎）
```go
type AgCacheProperties struct { DefaultEngine string }
```
core 自绑 `agcache.defaultEngine`（无 value 标签，按字段名，binder 大小写不敏感匹配）。引擎参数与 TTL 全部在 `agcache.ristretto.*`（agristretto 自绑）。
**理由**：core 职责 = "用配置按 key 选引擎"，仅此而已；v1 的 default/overrides/EngineConfig 继承模板是过度设计。

### D5. 命名体现引擎实现
包 `ag/ag_cache/agristretto`、配置段 `agcache.ristretto`、注册名 `"ristretto"`、Fx 模块 `FxAgCacheRistrettoMode`。
对标 `ag_log/logzap`（zap 引擎子包避同名）；"local" 过于含糊。

### D6. 包级实例管理（去 Manager 装配）
`New[T](name, loader, opts...)` / `Get[T]` / `GetAdmin[T]` / `CloseAll`，内部 `Manager{defaultEngine, caches map[string]any}` 懒创建复用。
引擎选择：`WithEngine(opts)` 或 defaultEngine；TTL 优先级 = `WithDefaultTTL` > 工厂 `DefaultTTLProvider` > 5min；同名复用首次创建的引擎/TTL。
**理由**：core 无装配配置后，v1 的 `Config{Default, Overrides}`/`NamespaceConfig` 无意义，退化为包级注册表 + 生命周期。

### D7. 核心约束与循环依赖
`ag_cache` 顶层不得 import `agristretto`（`agristretto` 依赖 core，反向即环）。"core 模块默认带引擎模块"字面不可行，单模块效果由引擎模块反向 include core 模块实现（`agristretto.FxAgCacheRistrettoMode` 内含 `ag_cache.FxAgCacheMode`）。

## 文件与函数签名清单

### `ag/ag_cache/`（核心包，主 module）
- **cache.go**：`type LoaderFunc[T] func(ctx, key) (T, error)`；`type ICache[T] interface{Get; GetOrElse; Set(ttl...); Del(keys...); Clear}`；`type AdminCache[T] interface{ICache[T]; Peek; Stats}`；`type Stats struct{Hits,Misses,Evictions,EntryCount int64}`；`var ErrCacheMiss, ErrBackend error`
- **engine.go**：`type Engine interface{Get; Set; Del; Clear; Stats; Close}`；`type EngineFactory interface{Name; Create() (Engine, error)}`；`type DefaultTTLProvider interface{DefaultTTL() time.Duration}`；`func RegisterEngine(f)`；`func EngineRegistered(name) bool`；`func getFactory(name) (EngineFactory, error)`；`type syncer interface{Sync()}`；`func errBackend(err) error`
- **config.go**：`const AgCacheConfPrefix = "agcache"`；`type AgCacheProperties struct{DefaultEngine string}`；`func DefaultAgCacheProperties() *AgCacheProperties`；`func BindAgCacheProperties(binder ag_conf.IBinder) (*AgCacheProperties, error)`
- **manager.go**：`type Manager struct{...}`；`func NewManager(props *AgCacheProperties) (*Manager, error)`；`func (m *Manager) Close() error`；`func (m *Manager) Visit(fn func(name string, s Stats))`；`func SetDefault(m *Manager)`；`func New[T any](name string, loader LoaderFunc[T], opts ...Option[T]) *LoaderCache[T]`；`func Get[T any](name string) ICache[T]`；`func GetAdmin[T any](name string) AdminCache[T]`；`func CloseAll()`；`func getOrCreate[T any](m *Manager, name string, opts ...Option[T]) *typedCache[T]`
- **typed.go**：`type typedCache[T] struct{engine Engine; engineName string; serializer Serializer[T]; defaultTTL time.Duration; sf singleflight.Group}`；`type Option[T] func(*typedCache[T])`；`func WithEngine[T any](name string) Option[T]`；`func WithDefaultTTL[T any](ttl time.Duration) Option[T]`；`func WithSerializer[T any](s Serializer[T]) Option[T]`；`func NewWithEngine[T any](engine Engine, opts ...Option[T]) ICache[T]`；`func NewAdminWithEngine[T any](...) AdminCache[T]`；方法 `Get/GetOrElse/Set/Del/Clear/Peek/Stats/stats/closeEngine/unmarshal/recoverPanic`
- **loader_cache.go**：`type LoaderCache[T] struct{inner ICache[T]; loader LoaderFunc[T]}`；`func WithLoader[T any](c ICache[T], loader LoaderFunc[T]) *LoaderCache[T]`；方法 `Get/GetOrElse/Set/Del/Clear/Peek/Stats`
- **serializer.go**：`type Serializer[T] interface{Marshal; Unmarshal}`；`func DefaultSerializer[T]() Serializer[T]`
- **mock.go**：`type MockCache[T]` + `NewMock[T]()` + `SetError`；`type MockEngine` + `NewMockEngine()`（含 PanicNext/Err 注入）
- **zfx_ag_cache.go**：`type EngineFactoryParams struct{fx.In; Factories []EngineFactory \`group:"agcache.engine"\`}`；`func NewAgCacheManager(p, props) (*Manager, error)`；`func registerHooks(lc, m)`；`func LogStats(m)`；`var FxAgCacheMode`

### `ag/ag_cache/agristretto/`
- **ristretto.go**：`type RistrettoConfig struct{MaxCost, NumCounters int64}`；`func DefaultRistrettoConfig() RistrettoConfig`；`type ristrettoEngine struct{cache *ristretto.Cache[string, []byte]}`；`func NewRistrettoEngine(cfg RistrettoConfig) (ag_cache.Engine, error)`；方法 `Get/Set/Del/Clear/Stats/Close/Sync`；`type agristrettoFactory struct{cfg RistrettoConfig; ttl time.Duration}`；`Name()="ristretto"`；`Create()` 无参；`DefaultTTL() time.Duration`
- **config_ristretto.go**：`const RistrettoConfPrefix = "agcache.ristretto"`；`type RistrettoConfigProperties struct{MaxCost, NumCounters int64; DefaultTTL string}`；`func DefaultRistrettoConfigProperties() *RistrettoConfigProperties`；`func BindRistrettoConfigProperties(binder ag_conf.IBinder) (*RistrettoConfigProperties, error)`；`func parseTTL(s string) (time.Duration, error)`
- **zfx_agristretto.go**：`func ProvideAgristrettoFactory(binder ag_conf.IBinder) (ag_cache.EngineFactory, error)`；`var FxAgCacheRistrettoMode = fx.Module("ag_cache.agristretto", fx.Provide(fx.Annotate(ProvideAgristrettoFactory, fx.ResultTags(\`group:"agcache.engine"\`))))`

## 测试策略

| 层 | 文件 | 测试 | 工具 |
|----|------|------|------|
| core 单元 | `ag/ag_cache/cache_test.go` | GetOrElse 基础/Get 纯读/singleflight 一次/序列化/隔离/Clear 只影响自身/WithoutCancel/后端故障不调 loader/恢复 miss/panic 恢复/LoaderCache 全组/WithEngine/WithDefaultTTL/同名不同类型 panic | MockEngine/MockCache + mock 引擎工厂（group/注册表） |
| core Manager | `ag/ag_cache/manager_test.go` | 懒创建复用/NewManager 未知引擎 fail-fast/Visit/CloseAll 幂等/SetDefault | mock 引擎 |
| core 配置 | `ag/ag_cache/config_test.go` | `BindAgCacheProperties`（defaultEngine 默认/覆盖/缺失） | ag_conf binder |
| 引擎集成 | `agristretto/ristretto_test.go` | TTL 过期/ttl=0 永不过期/Stats 独立/淘汰/Sync 可见性/工厂 Create/DefaultTTLProvider | 真实 Ristretto |
| 引擎配置 | `agristretto/config_ristretto_test.go` | `BindRistrettoConfigProperties`（maxCost/defaultTtl/缺失/非法 TTL） | ag_conf binder |
| Fx | `ag/ag_cache/zfx_ag_cache_test.go` | `fxtest.New(t, fx.Provide(binder), FxAgCacheMode, FxAgCacheRistrettoMode)` 联合装配/生命周期/幂等注册 | fx/fxtest |
| bench | `agristretto/bench_test.go` | 4 个 benchmark（纯 Ristretto/EngineSet/GetOrElse miss/hit） | 真实 Ristretto |

验证命令：`go test -race ./ag/ag_cache/...`、`go vet ./ag/ag_cache/...`、`go build ./...`。

## Risks / Trade-offs

- [R1] **fx group 元素类型匹配**：`ProvideAgristrettoFactory` 必须声明返回 `ag_cache.EngineFactory`（接口）以匹配 group 收集 `[]EngineFactory`；若返回具体类型，core 收集失败 → 测试先行验证。
- [R2] **同名不同引擎**：`WithEngine` 与 defaultEngine 对同一 name 冲突 → 懒创建复用语义：首次创建决定，后续忽略选择差异（文档化）。
- [R3] **幂等注册跨 fxtest**：多次 `fxtest.New` 各装配引擎模块会重复 Provide → core 经 `EngineRegistered` 判断跳过，不 panic。
- [R4] **`Stats.EntryCount` 推导不精确**（Ristretto Metrics `KeysAdded-KeysEvicted-KeysUpdated`）→ 集成测试只断言存在性/非负。
- [R5] **`FxFAgCacheMode` 测试依赖 binder**：`FxAgConfModule` 在 fxtest 下需真实 app.yml（读 `os.Executable()` 目录）→ 测试用自定义 `fx.Provide(binder)` 注入。

## Migration Plan

- 无线上数据迁移；纯新增代码，不影响现有模块（`ag_log`/`contribute/*`/`fxs/*`/`go.work`/`release_aif.sh` 不改动）。
- 落地顺序：M1 core SPI+配置 → M2 agristretto 引擎 → M3 Fx group 集成 → M4 收尾验证。
- 回滚：未合入前丢弃即可；合入后为纯新增 API，无破坏面。

## Open Questions

- 示例应用（设计文档 v1 14.1"示例项目 Fx 装配版"）是否在本 change 内交付，或由业务侧另行验证（fx 集成由 fxtest 覆盖）。
