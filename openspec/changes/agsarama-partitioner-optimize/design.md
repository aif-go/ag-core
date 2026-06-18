## Context

`contribute/agsarama/config.go` 中 `PartitionerType` 枚举的 `Manual` 常量值为 `"Manual"`（首字母大写），而 `PartitionerTypeHash` 及其余所有枚举均为全小写。`ToSarama()` 方法做精确匹配，当用户传入 `"manual"`（全小写）时落入 `default` 分支，仅打 `slog.Warn` 后返回 `nil`，Sarama 使用默认 `hash` 分区器——无声丢失语义。

## Goals / Non-Goals

**Goals:**
- 统一 `PartitionerType` 枚举风格为全小写
- 存量配置 (`partitioner: Manual`) 依然兼容
- 新增 `random` / `roundrobin` 两种分区器
- 无效枚举值改为显式 error，不再静默降级

**Non-Goals:**
- 不修改 Sarama 客户端行为
- 不改动 `Config` 结构体字段名或序列化格式

## Decisions

| 决策 | 选择 | 替代方案 | 理由 |
|------|------|---------|------|
| 大小写处理 | `strings.ToLower` 归一化后 switch | `strings.EqualFold` | ToLower 配合 hard-coded case 字符串，性能等同、更易读 |
| 新增分区器 | 直接新增常量 + case | 通过反射、动态注册 | 类型安全，编译检查 |
| 无效值响应 | 返回 `fmt.Errorf` | 返回 error + caller 处理 | 调用方 `ToSaramaConfig()` 已有 `if err != nil` 逻辑 |
| 常量名 | 保留 `PartitionerTypeManual` 符号名 | 新增别名如 `PartitionerTypeManualLower` | 零外部引用风险 |

## Risks / Trade-offs

- **常量值变更风险**：`PartitionerTypeManual == "Manual"` 的外部比较会失效 → 项目内搜索已确认无此用法；极端情况可添加 `// Deprecated: use strings.EqualFold for comparison` 注释
- **强制要求 `strings.ToLower`**：若未来添加大小写敏感的分区器类型，需单独处理 → 当前无此需求，且约定为新常量也统一为全小写
- **测试覆盖**：`config_test.go` 当前无 partitioner 专项测试 → 需补充 `TestPartitionerType_ToSarama` 表驱动测试覆盖所有枚举值 + 大小写变体 + 无效值

### 测试策略

```go
// contribute/agsarama/config_test.go
func TestPartitionerType_ToSarama(t *testing.T) {
    tests := []struct {
        name    string
        input   PartitionerType
        wantNil bool   // nil 表示期望 error
    }{
        {"hash", PartitionerTypeHash, false},
        {"manual lowercase", PartitionerTypeManual, false},
        {"Manual uppercase", "Manual", false},    // 兼容旧值
        {"MANUAL uppercase", "MANUAL", false},    // 大小写不敏感
        {"random", PartitionerTypeRandom, false},
        {"roundrobin", PartitionerTypeRoundRobin, false},
        {"invalid", "unknown", true},
    }
    // ...
}
```
