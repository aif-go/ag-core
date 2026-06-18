## Why

Agsarama 的 `PartitionerType` 枚举中 `Manual` 的值 `"Manual"` 首字母大写，与其他所有枚举的全小写风格不一致，导致两个问题：YAML 中 `partitioner: Manual` 显得突兀且开发者直觉写 `manual`（静默降级为 hash），以及当新增分区器类型时风格约束不明确。

## What Changes

- 改 `PartitionerTypeManual` 常量值从 `"Manual"` 为 `"manual"`，统一全小写风格
- `ToSarama()` 改用 `strings.ToLower(string(p))` 归一化后匹配，兼容存量配置中的大小写差异
- 新增 `PartitionerTypeRandom` / `PartitionerTypeRoundRobin` 两种分区器类型
- `default` 分支从静默降级（`slog.Warn` + 返回 nil）改为返回 `fmt.Errorf`，无效值立即报错

## Capabilities

### New Capabilities
- `partitioner-type`: 定义 `PartitionerType` 枚举常量及其到 Sarama 分区器的映射，包括大小写不敏感的匹配规则

### Modified Capabilities
- 无（无现有 specs）

## Impact

- **仅 `contribute/agsarama/config.go` 一个文件**（约 40 行改动）
- 常量值变更对外部调用者有理论影响，但包内无直接引用，搜索确认项目内也无外部等值比较
