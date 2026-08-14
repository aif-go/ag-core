## Context

`ag_log` 已通过 `fanout` 包包装 `samber/slog-multi` 的 `Fanout`，实现多 handler 分发：配置 `aglog.fanout.logs` 为 `map[string][]string`，经 `NewFanoutHandlerFactorys` 生成 `HandlerFactory`，通过 fx `group:"agslog.factorys"` 注入 `agslog.Builder`，最终由 `resolveHandler` 按名递归解析子 handler 并组合。

`Fanout` 与 `Failover` 的接入点完全一致，仅组合 API 不同：`slogmulti.Fanout(sub...)` vs `slogmulti.Failover()(sub...)`。因此 failover 的包装可镜像 fanout 包实现。

同时 `fanout` 包存在 nil 安全隐患：`BindAgSLogFanoutProperties` 在 `binder.Bind` 出错时 `return nil, nil`，会导致 fx 报 nil 返回值而启动失败（违背“配置异常不中断”的意图，与同目录 `slogzap` 的 `return prop, nil` 风格不一致）；且 `NewFanoutHandlerFactorys` 无 nil 保护。

## Goals / Non-Goals

**Goals:**
- 新增 `failover` 包，镜像 fanout，包装 `slogmulti.Failover`，支持按优先级顺序的故障转移。
- 修复 fanout 的 nil 安全（Bind 失败返回非 nil 空结构 + 工厂 nil 保护 + 文案 typo）。
- 提供 failover 单元测试、fanout nil 回归测试，以及 `test/agslog_failover.yaml` 集成测试。

**Non-Goals:**
- 不实现 round-robin 等其它路由策略（`slogmulti` 该特性 `@TODO` 未实现）。
- 不做 failover 与 `AsyncHandler` 组合的运行时保护（仅文档约定）。
- 不改变 `async` 包 `return nil, err` 的 Bind 语义。

## Decisions

1. **新建独立 `failover` 包**，与 fanout 平级对称（而非并入 fanout），职责清晰。
2. **配置结构 `AgSlogFailoverProperties{ Logs map[string][]string }`**，前缀 `aglog.failover`；value 的 slice 顺序即 failover 优先级。
3. **`BindAgSLogFailoverProperties` 出错返回 `return prop, nil`**（沿用 slogzap 风格），配合工厂 nil 检查，实现“配置异常不中断”。
4. **工厂逻辑镜像 fanout**：`getDoGetHandlerFunc` 按序解析子 handler，缺失跳过并告警，全部缺失返回 error，最终 `slogmulti.Failover()(subHandlers...)`。
5. **fanout 修复**：`BindAgSLogFanoutProperties` 改 `return prop, nil`，`NewFanoutHandlerFactorys` 加 `props == nil || props.Logs == nil` 保护，顺带修正 `BindSlogZapProperties` 文案 typo。

## Risks / Trade-offs

- [Failover 切换静默] → `slog.Logger` 忽略 `Handle` 的 error，切换无日志；在 proposal/spec 中已明确为语义边界。
- [Failover 与 AsyncHandler 不兼容] → 异步 handler 入队即返回 nil，吞掉 error，failover 永不切换；仅文档约定。
- [子 handler 解析失败被静默跳过] → 可能掩盖配置错误；沿用 fanout 现有行为（仅打印告警）。

## Files & Function Signatures

**新增 `ag/ag_log/failover/slog_failover.go`:**
- `const AgSlogFailoverPropertiesKeyPrefix = "aglog.failover"`
- `type AgSlogFailoverProperties struct { Logs map[string][]string }`
- `func BindAgSLogFailoverProperties(binder ag_conf.IBinder) (*AgSlogFailoverProperties, error)`
- `func NewFailoverHandlerFactorys(props *AgSlogFailoverProperties) ([]*agslog.HandlerFactory, error)`
- `func getDoGetHandlerFunc(names []string) func(getHandler func(string)(slog.Handler,error))(slog.Handler,error)`

**新增 `ag/ag_log/failover/zfx_aglog_failover.go`:**
- `var FxAgSlogFailoverProvide = fx.Provide(...)`

**修改 `ag/ag_log/fanout/slog_fanout.go`:**
- `func BindAgSLogFanoutProperties(...)` — 改 `return prop, nil`
- `func NewFanoutHandlerFactorys(...)` — 增加 nil 保护

**修改 `ag/ag_log/zfx_aglog.go`:**
- `FxAglogMode` 增加 `failover.FxAgSlogFailoverProvide`

## Testing Strategy

| 行为 | 测试文件 | 策略 | 测试函数 |
|---|---|---|---|
| failover 故障转移（失败切换/成功不切/全失败） | `ag/ag_log/failover/slog_failover_test.go` | unit | `TestFailoverFallsBackOnError`、`TestFailoverStopsOnFirstSuccess`、`TestFailoverAllFailReturnsError` |
| failover 工厂 nil 安全 | 同上 | unit | `TestNewFailoverHandlerFactorysNilSafe` |
| failover 子 handler 解析跳过/全缺 | 同上 | unit | `TestGetDoGetHandlerFuncSkipMissing`（可选） |
| fanout Bind 失败返回非 nil | `ag/ag_log/fanout/slog_fanout_test.go` | unit | `TestBindAgSLogFanoutPropertiesBindError` |
| fanout 工厂 nil 安全 | 同上 | unit | `TestNewFanoutHandlerFactorysNilSafe` |
| failover fx 装配 | `ag/ag_log/test/`（`agslog_failover.yaml`） | integration | 复用 `FxAglogMode` 装配，验证 `GetSlogByName` 解析 failover handler |

隔离策略：unit 测试直接调用 `getDoGetHandlerFunc` + mock `getHandler`，不经 `globalManager`/fx；集成测试用独立 `agslog_failover.yaml` 配置。

## Migration Plan

纯增量 + 内部修复，无外部接口破坏。回滚即还原相关文件。

## Open Questions

无。
