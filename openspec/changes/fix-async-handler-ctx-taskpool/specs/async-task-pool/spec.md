## ADDED Requirements

### Requirement: taskPool 借还封装 {#req-001}
taskPool SHALL 提供 `Borrow()` 与 `Return(t)` 方法，`Return` SHALL 在归还前重置 task 的 `ctx`、`handler` 为 nil、`record` 为零值。

#### Scenario: Return 重置 task 字段
- **WHEN** 借用 task 并设置其 `ctx`/`handler`/`record` 后调用 `Return`
- **THEN** 再次 `Borrow` 得到的 task 的 `ctx` 与 `handler` 为 nil、`record` 为零值

#### Scenario: Return nil 不 panic
- **WHEN** 调用 `Return(nil)`
- **THEN** 不触发 panic

### Requirement: drop_new 满时丢弃并归还 {#req-002}
当 `FullStrategy` 为 `drop_new` 且队列满时，`Submit` SHALL 丢弃新任务、递增 `Dropped` 计数，并归还 task 到池。

#### Scenario: drop_new 队列满丢弃新任务并归还
- **WHEN** 队列满时提交新任务
- **THEN** `Dropped` 递增 1、`Queued` 不变、task 归还到池（随后 `Borrow` 得到同一对象）

### Requirement: drop_old 满时丢最旧放新且非阻塞 {#req-003}
当 `FullStrategy` 为 `drop_old` 且队列满时，`Submit` SHALL 弹出最旧任务并归还、入队新任务，全程非阻塞，并正确递增计数。

#### Scenario: drop_old 弹出旧任务并归还新任务入队
- **WHEN** 队列满时提交新任务
- **THEN** `Dropped` 递增 1、被弹出的旧任务归还到池、新任务入队（`Queued` 递增 1）、新任务最终被 worker 处理（`Processed` 递增）

#### Scenario: 未知策略不 panic
- **WHEN** `FullStrategy` 为未知值
- **THEN** `Submit` 返回 nil 且不 panic

### Requirement: Submit 计数对账一致 {#req-004}
对任意一批提交，SHALL 满足 `Queued + Dropped` 等于提交总数，且 worker 排空后 `Processed` 等于 `Queued`。

#### Scenario: 混合入队与丢弃的计数对账
- **WHEN** 提交 N 个任务到小容量队列并等待 worker 排空
- **THEN** `Queued + Dropped == N` 且 `Processed == Queued`
