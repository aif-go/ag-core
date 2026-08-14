## Why

异步日志 handler（`AsyncHandler`）在 `Handle` 时把调用方 `ctx` 与克隆后的 `slog.Record` 提交到 worker 队列异步处理。当前实现存在两个缺陷：(1) 队列满时被丢弃的任务未归还 `taskPool`，导致对象复用失效、且任务持有的 `ctx`/`record` 引用延迟释放；(2) 异步执行时 `ctx` 可能已被调用方取消，底层 handler 会误判取消状态。需要在高压场景下保证不泄漏、语义正确。

## What Changes

- 将裸 `sync.Pool` 封装为 `logTaskPool` 类型，提供 `Borrow()` 与 `Return(t)` 方法，`Return` 内聚 `Reset` 逻辑，纯生命周期管理、不含业务统计。
- `WorkerGroup.Submit` 的 `drop_new`/`drop_old` 三条丢弃路径统一归还 task（修复泄漏），并修正 `drop_old` 漏记的 `Queued`/`Dropped` 计数。
- `drop_old` 的阻塞式 `wg.logQueue <- task` 改为非阻塞 `select`，消除并发下空位被抢占导致的意外阻塞。
- `worker.processTask` 的归还逻辑改走 `taskPool.Return`。
- `AsyncHandler.Handle` 使用 `context.WithoutCancel(ctx)` 切断取消传播，保留 `ctx.Value` 读取。

## Capabilities

### New Capabilities

- `async-task-pool`: 异步日志任务池的借/还封装，及队列满时三种丢弃策略（drop_new/block_wait/drop_old）下任务归还与计数语义。
- `async-handler-context`: `AsyncHandler` 异步处理对 `ctx` 取消语义的隔离——保留值读取、切断取消传播。

### Modified Capabilities

<!-- 无既有 spec 变更 -->

## Impact

- 修改文件：`ag/ag_log/async/async_worker_group.go`、`ag/ag_log/async/async_handler.go`
- 新增测试：`ag/ag_log/async/async_worker_group_test.go`
- 不改变 `FullStrategy` 的对外语义、不改变 `slog.Handler` 接口、无新增外部依赖
- 依赖 Go 1.21+ 的 `context.WithoutCancel`（项目 go.mod 为 1.24，满足）
