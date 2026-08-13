## ADDED Requirements

### Requirement: bool 类型列零值检查 {#req-001}
gen-go-db SHALL 为 GoType 为 bool 的列生成 `!字段名` 的零值检查表达式。

#### Scenario: bool 列生成正确检查语句
- GIVEN 表包含一个 GoType 为 bool 的列 is_active
- WHEN 调用 getZeroCheck
- THEN 返回的表达式为 `!user.IsActive`

#### Scenario: 非 bool 列行为不变
- GIVEN 表包含 string 类型列 name
- WHEN 调用 getZeroCheck
- THEN 返回的表达式为 `user.Name == ""`

### Requirement: 条件声明 generalColZeroVal 变量 {#req-002}
gen-go-db SHALL 仅在表中存在普通列时声明 `generalColZeroVal` 变量。

#### Scenario: 有普通列时声明变量
- GIVEN 表包含至少一个非主键、非索引、非特殊列
- WHEN 生成 ListZeroValueCols 方法
- THEN 代码中包含 `var generalColZeroVal bool = false` 声明

#### Scenario: 无普通列时不声明变量
- GIVEN 表全部列都是主键、索引或特殊列
- WHEN 生成 ListZeroValueCols 方法
- THEN 代码中不包含 `generalColZeroVal` 变量声明
