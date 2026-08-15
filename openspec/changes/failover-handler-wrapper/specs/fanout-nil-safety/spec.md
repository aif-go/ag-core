## ADDED Requirements

### Requirement: fanout 绑定失败返回空结构 {#req-001}
`BindAgSLogFanoutProperties` SHALL 在绑定出错时返回非 nil 的空结构，而非 nil。

#### Scenario: 绑定失败返回非 nil
- **WHEN** `binder.Bind` 返回错误
- **THEN** 返回的 `props` 非 nil（空结构）且 error 为 nil

### Requirement: fanout 工厂 nil 安全 {#req-002}
`NewFanoutHandlerFactorys` SHALL 在 `props` 为 nil 或 `Logs` 为 nil 时返回空列表且不 panic。

#### Scenario: nil props 安全
- **WHEN** 传入 nil `props`
- **THEN** 返回空 factory 列表且不 panic

#### Scenario: nil Logs 安全
- **WHEN** 传入 `Logs` 为 nil 的 `props`
- **THEN** 返回空 factory 列表且不 panic
