## ADDED Requirements

### Requirement: failover 配置绑定 {#req-001}
系统 SHALL 通过 `aglog.failover` 前缀绑定 failover 配置，其中 `Logs` 为 `map[string][]string`，key 为 failover handler 名，value 为按优先级排序的子 handler 名列表。

#### Scenario: 绑定 failover 配置
- **WHEN** 配置中存在 `aglog.failover.logs` 段
- **THEN** 绑定得到 `AgSlogFailoverProperties`，其 `Logs` 包含对应 key 及有序的 handler 名列表

### Requirement: failover 工厂构建 {#req-002}
系统 SHALL 为每个 failover 名创建 `HandlerFactory`，其构建逻辑按 value 顺序解析子 handler，并用 `slogmulti.Failover()` 组合。

#### Scenario: 解析并按序组合子 handler
- **WHEN** 通过 `getHandler` 按名称解析子 handler
- **THEN** 子 handler 按配置顺序组合为 failover handler

#### Scenario: 子 handler 缺失时跳过
- **WHEN** 某个子 handler 名称无法解析
- **THEN** 跳过该名称并继续解析，剩余子 handler 仍正常组合

#### Scenario: 全部子 handler 缺失返回错误
- **WHEN** 所有子 handler 名称均无法解析
- **THEN** 返回错误且不生成 handler

### Requirement: failover 故障转移语义 {#req-003}
组合得到的 handler SHALL 按顺序尝试子 handler：第一个 `Enabled` 且 `Handle` 返回 nil 的即成功，全部失败时返回 error。

#### Scenario: 首选失败切换次选
- **WHEN** 首选 handler 的 `Handle` 返回 error
- **THEN** 记录转发给次选 handler 且返回 nil

#### Scenario: 首选成功不切换
- **WHEN** 首选 handler 的 `Handle` 返回 nil
- **THEN** 不调用后续 handler

#### Scenario: 全部失败返回错误
- **WHEN** 所有 handler 的 `Handle` 均返回 error
- **THEN** 返回非 nil error

### Requirement: failover 工厂 nil 安全 {#req-004}
`NewFailoverHandlerFactorys` SHALL 在 `props` 为 nil 或 `Logs` 为 nil 时返回空列表且不 panic。

#### Scenario: nil props 安全
- **WHEN** 传入 nil `props`
- **THEN** 返回空 factory 列表且不 panic

#### Scenario: nil Logs 安全
- **WHEN** 传入 `Logs` 为 nil 的 `props`
- **THEN** 返回空 factory 列表且不 panic
