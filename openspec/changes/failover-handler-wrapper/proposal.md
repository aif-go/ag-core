## Why

ag-core 的 `ag_log` 已通过 `fanout` 包包装 `slogmulti.Fanout` 实现多 handler 分发，但缺少 `slogmulti.Failover` 的等价包装（按优先级依次尝试 handler，失败则切换下一个）。同时 `fanout` 包存在 nil 安全隐患：`Bind` 失败时返回 `(nil, nil)`，且工厂函数缺少 nil 保护，配置异常时可能触发 fx 启动失败或 nil 解引用。

## What Changes

- 新建 `ag/ag_log/failover` 包，镜像 `fanout` 包结构，包装 `slogmulti.Failover`。
- 修改 `fanout` 包：`BindAgSLogFanoutProperties` 出错时返回非 nil 空结构；`NewFanoutHandlerFactorys` 增加 nil 保护；修正错误日志文案。
- 修改 `ag/ag_log/zfx_aglog.go`，在 `FxAglogMode` 中注册 `failover.FxAgSlogFailoverProvide`。
- 新增 failover 单元测试与 fanout nil 安全回归测试。
- 新增 `test/agslog_failover.yaml` 集成测试配置及 fx 装配验证。

## Capabilities

### New Capabilities

- `failover-handler`: failover 日志 handler 的配置绑定、工厂构建与 fx 装配，按优先级顺序组合子 handler 并在失败时切换。
- `fanout-nil-safety`: fanout 配置绑定与工厂在配置缺失/绑定失败时的 nil 安全行为。

### Modified Capabilities

<!-- 无既有 spec 变更 -->

## Impact

- 新增文件：`ag/ag_log/failover/slog_failover.go`、`ag/ag_log/failover/zfx_aglog_failover.go`
- 修改文件：`ag/ag_log/fanout/slog_fanout.go`、`ag/ag_log/zfx_aglog.go`
- 新增测试：`ag/ag_log/failover/slog_failover_test.go`、`ag/ag_log/fanout/slog_fanout_test.go`、`ag/ag_log/test/agslog_failover.yaml`
- 无外部依赖变更（复用 `samber/slog-multi` v1.4.1 已具备的 `Failover`）
- 不改变 `Fanout` 的对外语义，仅修复其异常路径
