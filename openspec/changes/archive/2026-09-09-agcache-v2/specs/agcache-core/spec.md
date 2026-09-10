## ADDED Requirements

### Requirement: ICache 业务接口
`ICache[T]` SHALL 提供 `Get(ctx, key)`（纯读，miss 返回 `ErrCacheMiss`）、`GetOrElse(ctx, key, loader)`（读穿透）、`Set(ctx, key, value, ttl...)`（ttl 省略用默认、0=永不过期、>0=显式）、`Del(ctx, keys...)`、`Clear(ctx)`（清独立实例）。
`AdminCache[T]` SHALL 扩展 `ICache[T]`，额外提供 `Peek(ctx, key)` 与 `Stats()`。
`typedCache[T]` 实例 SHALL 同时实现 `ICache[T]` 与 `AdminCache[T]`（同一实例不同视角）。

#### Scenario: Get miss 返回 ErrCacheMiss
- **WHEN** 对未缓存 key 调用 `Get`
- **THEN** 返回 `(zero, ErrCacheMiss)` 且不调用任何 loader

#### Scenario: GetOrElse 读穿透
- **WHEN** 对未缓存 key 调用 `GetOrElse` 且 loader 返回 `(v, nil)`
- **THEN** 返回 v 且写入缓存；再次 `Get` 命中缓存不再调用 loader

#### Scenario: Clear 只清独立实例
- **WHEN** 对 params 实例调用 `Clear`
- **THEN** params 条目清空且 users 实例条目不受影响

### Requirement: Engine SPI 与工厂
`Engine` SHALL 提供 `Get`/`Set`/`Del`/`Clear`/`Stats`/`Close`，全方法带 ctx；错误契约：`Get` 返回 `ErrCacheMiss` 为未命中，其他 error 为后端故障（调用方不得视为 miss）。
`EngineFactory` SHALL 提供 `Name()` 与 `Create()`（**无参**，引擎配置由工厂构造时自持）。
`RegisterEngine` SHALL 按 Name 注册，重复注册 panic；`EngineRegistered(name)` SHALL 返回是否已注册（幂等查询）。

#### Scenario: 工厂注册与幂等查询
- **WHEN** 注册引擎后调用 `EngineRegistered(name)`
- **THEN** 返回 true；未注册的名称返回 false

#### Scenario: Create 无参
- **WHEN** 调用已注册工厂的 `Create()`
- **THEN** 返回引擎实例且不接收外部配置（引擎配置已由工厂自持）

### Requirement: 包级实例管理
`New[T](name, loader, opts...)` SHALL 返回绑定 loader 的 `*LoaderCache[T]`，按 name 懒创建 typedCache 并复用；同名不同类型 SHALL panic。
`Get[T](name)` / `GetAdmin[T](name)` SHALL 复用已创建实例；`CloseAll` SHALL 幂等关闭全部实例。
引擎选择：默认使用 core 配置 `DefaultEngine`，`WithEngine` Option SHALL 可覆盖指定引擎实现名；同名复用首次创建的引擎/TTL，后续忽略选择差异。
`NewManager` SHALL 校验 defaultEngine 已注册（启动期快速失败）。

#### Scenario: 懒创建与复用
- **WHEN** 同一 name 两次 `New`/`Get` 且首次写入后再次读取
- **THEN** 返回同一实例，第二次读命中缓存、loader 不被调用

#### Scenario: 同名不同类型 panic
- **WHEN** 对同一 name 先后用不同类型 T 获取
- **THEN** panic

#### Scenario: 未知引擎快速失败
- **WHEN** `NewManager` 的 defaultEngine 未注册
- **THEN** 返回错误，错误信息包含引擎名

#### Scenario: WithEngine 选择指定引擎
- **WHEN** `New` 传入 `WithEngine("other")` 且该引擎已注册
- **THEN** 实例使用指定引擎创建（与默认引擎隔离）

### Requirement: 后端故障 ≠ miss
`GetOrElse` SHALL 仅在引擎返回 `ErrCacheMiss` 时调用 loader；其他引擎错误 SHALL 包装为 `ErrBackend` 直接返回，绝不调用 loader（防打爆源）。
`Get`/`Set`/`Del`/`Clear` SHALL 将非 `ErrCacheMiss` 的引擎错误包装为 `ErrBackend`（`errors.Is` 可判）。
引擎 panic SHALL 被 `typedCache` 捕获并转换为 `ErrBackend`（保留 panic 上下文）；引擎恢复后自动回到 miss 语义。

#### Scenario: 后端故障不调 loader
- **WHEN** 引擎注入后端故障且调用 `GetOrElse`
- **THEN** 返回 `ErrBackend`，loader 不被调用

#### Scenario: 引擎恢复后回到 miss
- **WHEN** 引擎故障清除后对未缓存 key 调用 `Get`
- **THEN** 返回 `ErrCacheMiss`

#### Scenario: 引擎 panic 被恢复
- **WHEN** 引擎在 `Get`/`Set` 中 panic
- **THEN** 返回 `ErrBackend`（不向上传播 panic）

### Requirement: 防击穿 singleflight
`GetOrElse` SHALL 对同一 key 的并发 miss 合并为一次 loader 调用（singleflight）；loader SHALL 使用 `context.WithoutCancel(ctx)`，首个调用者 ctx 取消不影响其余等待者。

#### Scenario: 并发 miss 只调用一次 loader
- **WHEN** 10 个 goroutine 并发对同一 key `GetOrElse`
- **THEN** loader 恰好被调用 1 次，所有 goroutine 拿到相同结果

#### Scenario: 首个调用者 ctx 取消不影响 loader
- **WHEN** 首个调用者使用已取消的 ctx，其余调用者使用正常 ctx
- **THEN** loader 仍执行一次并成功，所有调用者不报错

### Requirement: 默认 TTL 归属（DefaultTTLProvider）
`DefaultTTLProvider` SHALL 为可选接口：引擎实现 `DefaultTTL() time.Duration` 时，core 在创建 typedCache 时用它作为默认 TTL；未实现时 core 兜底 5min。
业务 `WithDefaultTTL` Option SHALL 优先级最高，覆盖引擎默认与 core 兜底。

#### Scenario: 引擎声明默认 TTL
- **WHEN** 引擎实现 `DefaultTTLProvider` 且业务未传 `WithDefaultTTL`
- **THEN** 缓存默认 TTL 为引擎声明的值

#### Scenario: 引擎未声明时 core 兜底
- **WHEN** 引擎不实现 `DefaultTTLProvider`
- **THEN** 缓存默认 TTL 为 5min

#### Scenario: 业务 Option 覆盖
- **WHEN** 业务传入 `WithDefaultTTL(30s)`
- **THEN** 缓存默认 TTL 为 30s（覆盖引擎默认）

### Requirement: 序列化
`Serializer[T]` SHALL 提供 `Marshal(v T) ([]byte, error)` 与 `Unmarshal(data []byte) (*T, error)`，实现 SHALL 线程安全。
`DefaultSerializer[T]` SHALL 对 string/[]byte 直接字节、int/int64/float64/bool 用 fast path、其他类型 JSON 回退。

#### Scenario: 结构体序列化往返
- **WHEN** 缓存写入结构体后 `Peek`
- **THEN** 读回的结构体字段值与写入一致

### Requirement: LoaderCache 语法糖
`LoaderCache[T]` SHALL 绑定默认 loader，使 `Get` 成为读穿透（miss → loader → 缓存 → 返回）。
`New[T](name, loader, opts...)` SHALL 基于默认 Manager 创建并绑定 loader；`WithLoader[T](c, loader)` SHALL 包装显式缓存。
`LoaderCache.GetOrElse(ctx, key, customLoader)` SHALL 允许临时覆盖 loader；`Peek` SHALL 保持纯读（不触发 loader）。

#### Scenario: Get 读穿透且多 key 复用 loader
- **WHEN** 对绑定 loader 的缓存先后 `Get` 两个不同 key
- **THEN** 首次 miss 均触发 loader 并缓存，再次 Get 命中不触发

#### Scenario: Peek 不触发绑定 loader
- **WHEN** 对未缓存 key 调用 `Peek`
- **THEN** 返回 miss 且 loader 未被调用
