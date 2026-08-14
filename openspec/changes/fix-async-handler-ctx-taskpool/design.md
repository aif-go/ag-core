## Context

异步日志 handler（`ag/ag_log/async`）中，`AsyncHandler.Handle` 从 `taskPool` 取 `logTask`，填入 `ctx`、克隆后的 `record` 与 `original` handler，再 `Submit` 到 `WorkerGroup` 的带缓冲 `logQueue`（容量 `config.Queue`），由 worker goroutine 异步消费。

当前缺陷：
1. `Submit` 的丢弃路径（`drop_new` 满、`drop_old` 弹出旧任务）不归还 task，导致 `sync.Pool` 复用失效、被丢弃 task 持有的 `ctx`/`record` 引用延迟释放。
2. `drop_old` 用阻塞式 `wg.logQueue <- task`，并发下弹出制造的空位可能被其他 producer 抢占，导致意外阻塞，违背非阻塞语义。
3. `drop_old` 漏记 `Queued`/`Dropped` 计数。
4. `Handle` 把调用方 `ctx` 原样交给异步 worker，ctx 可能已被取消，底层 handler 误判取消状态。

约束：Go 1.24（支持 `context.WithoutCancel`）；`slog.Handler` 接口不变；`FullStrategy` 对外语义不变。

## Goals / Non-Goals

**Goals:**
- taskPool 借/还封装，所有丢弃路径归还，`Get`/`Put` 严格配对。
- `drop_old` 全程非阻塞。
- 计数准确，可对账（`Queued + Dropped == 提交数`、`Processed == Queued`）。
- `AsyncHandler` 切断 ctx 取消传播、保留 ctx 值读取。

**Non-Goals:**
- 不修复 `Close()`/`Ref`/`Unref`/`ReleaseWorkerGroup` 引用计数问题（留待独立 change）。
- 不彻底释放 ctx value 链引用（受 `AttrFromContext` 动态提取限制，无法在 `Handle` 层枚举 key）。
- 不新增外部依赖，不改变 `FullStrategy` 语义。

## Decisions

1. **taskPool 封装为 `logTaskPool` 类型 + `Borrow`/`Return`**
   - 理由：`Reset` 内聚到 `Return`，所有归还点不会遗漏；职责纯生命周期，不含业务统计（统计留在 `WorkerGroup`）。
   - 备选：保持裸 `sync.Pool` + 手写 `Reset`/`Put`——调用点分散、易漏。

2. **丢弃路径归还 + 计数补全**
   - `drop_new` 满：`Dropped+1` 后 `Return(task)`。
   - `drop_old` 弹出旧：`Dropped+1`、`Return(old)`，再入队新任务。
   - `drop_old` 内层弹不出（竞态兜底）：`Dropped+1`、`Return(task)`。
   - 补 `drop_old` 入队时的 `Queued+1`。

3. **`drop_old` 入队改非阻塞 `select`**
   - 理由：消除“弹出空位被其他 producer 抢占导致阻塞”的竞态。
   - 备选：保留裸阻塞发送——串行下正确，但并发有阻塞窗口。
   - 代价：极端竞态下可能“丢一旧 + 丢一新”，换取绝不阻塞。

4. **ctx 用 `context.WithoutCancel`**
   - 理由：Go 1.21+，剥离取消/超时/deadline、保留值链，改动最小，解决取消语义丢失这一核心风险。
   - 备选：同步阶段提取 `AttrFromContext` 属性后异步传 `context.Background()`——因 `SlogAttrFromContext` 是用户注入的任意函数、无法枚举 key，不可行。

## Risks / Trade-offs

- [drop_old 极端竞态丢一旧+丢一新] → 接受，换取非阻塞语义；单测不覆盖该竞态分支。
- [WithoutCancel 不释放 value 链引用，ctx 对象仍被 task 持有到 worker 消费] → 记为已知限制，属 Non-Goal。
- [sync.Pool 复用断言依赖单 goroutine、无 GC 干扰] → 测试用指针复用断言，测试内避免并发提交与强制 GC。
- [WorkerGroup 引用计数未修，worker 永不 Stop] → 明确 out of scope。

## Files & Function Signatures

**修改 `ag/ag_log/async/async_worker_group.go`:**
- 新增 `type logTaskPool struct { pool sync.Pool }`
- 新增 `func (p *logTaskPool) Borrow() *logTask`
- 新增 `func (p *logTaskPool) Return(t *logTask)`
- 修改 `var taskPool = &logTaskPool{...}`
- 修改 `func (wg *WorkerGroup) Submit(task *logTask) error`
- 修改 `func (w *worker) processTask(task *logTask)`

**修改 `ag/ag_log/async/async_handler.go`:**
- 修改 `func (h *AsyncHandler) Handle(ctx context.Context, r slog.Record) error`

## Testing Strategy

新增 `ag/ag_log/async/async_worker_group_test.go`（unit，`package async`）：

| 行为 | 测试函数 |
|---|---|
| Return 重置字段 | `TestTaskPoolReturnResets` |
| Return(nil) 安全 | `TestTaskPoolReturnNil` |
| drop_new 满丢弃并归还 | `TestSubmitDropNewDropsAndReturns` |
| drop_old 丢旧放新归还 | `TestSubmitDropOldReturnsOldQueuesNew` |
| 未知策略 | `TestSubmitUnknownStrategy` |
| 计数对账 | `TestSubmitCounterReconciliation` |
| ctx 切断取消 | `TestHandleCutsCancellation` |
| ctx 保留值 | `TestHandlePreservesCtxValues` |

隔离策略：直接 `NewWorkerGroup(config)` 构造独立 group（绕开 `globalManager` 单例），ctx 测试直接构造 `&AsyncHandler{workerGroup: wg, original: mock}`，每个测试 `defer wg.Stop()`。

## Migration Plan

无外部接口变更，仅内部实现。直接合入，无需数据迁移；回滚即还原两个源文件。

## Open Questions

无（`Close`/引用计数问题已明确排除，待独立 change 处理）。
