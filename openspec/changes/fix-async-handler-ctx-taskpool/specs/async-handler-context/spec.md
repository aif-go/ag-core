## ADDED Requirements

### Requirement: 异步处理切断 ctx 取消传播 {#req-001}
`AsyncHandler.Handle` SHALL 在提交任务时使用 `context.WithoutCancel` 剥离取消/超时/deadline 语义，使 worker 异步执行时收到的 ctx 不处于取消状态。

#### Scenario: 已取消的 ctx 异步后不再取消
- **WHEN** 传入一个已取消的 ctx 调用 `Handle`
- **THEN** worker 执行底层 handler 时收到的 ctx 的 `Done()` 为 nil 且 `Err()` 为 nil

### Requirement: 异步处理保留 ctx 值读取 {#req-002}
`AsyncHandler.Handle` SHALL 保留 ctx 的值链，使底层 handler 异步执行时仍能读取 ctx 中的值。

#### Scenario: ctx 值在异步执行时仍可读
- **WHEN** 传入带 `context.WithValue` 的 ctx 调用 `Handle`
- **THEN** worker 执行底层 handler 时能从收到的 ctx 读取到该值
