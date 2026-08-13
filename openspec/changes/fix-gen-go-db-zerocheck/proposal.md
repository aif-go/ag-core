# Proposal

## Problem

代码审查发现 `tool/cmd/gen-go-db/model/template.go` 存在两个编译错误：

1. `getZeroCheck` 函数中 `bool` 类型列的 GoType 走 `default` 分支，生成 `Field == 0` 代码。Go 不允许 `bool` 与 `int` 比较，编译报错。
2. `generateListZeroValueColsMethod` 无条件声明 `var generalColZeroVal bool = false`，但表中无普通列（全部是主键/索引/特殊列）时该变量未被赋值使用，Go 编译报错 "declared and not used"。

## Features

### Feature: 修复 bool 类型的零值检查
- Scenario: 表中包含 boolean 列时 WHEN 生成的代码如下 THEN bool 类型列使用 `!字段名` 检查而非 `字段名 == 0`
- Scenario: 表中包含 string/time.Time 列时 WHEN 生成的零值检查 THEN 保持原有行为不变

### Feature: 修复无条件声明的 generalColZeroVal 变量
- Scenario: 表中存在普通列时 WHEN 生成 ListZeroValueCols 方法 THEN generalColZeroVal 被正常声明和赋值
- Scenario: 表中无普通列（全部是主键/索引/特殊列）时 WHEN 生成 ListZeroValueCols 方法 THEN generalColZeroVal 不被声明

## Out of Scope
- 不改变 gen-go-db 工具的功能行为
- 不涉及其他 GoType 类型的零值检查逻辑

## Constraints
- 必须向后兼容已有生成代码
- Go 语法合法，必须通过 `go vet` 和 `go build`

## Assumptions
- `gen-go-db` 工具生成的代码会由 ag-core 用户编译使用，必须确保语法正确
- MySQL `BOOLEAN`/`TINYINT(1)` 类型会被映射为 GoType `bool`

## Acceptance Criteria
- `go build ./tool/cmd/gen-go-db/...` 通过
- 生成的代码中 bool 列检查为 `!字段名`
- 无普通列的表不生成 `generalColZeroVal` 变量
