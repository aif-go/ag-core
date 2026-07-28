# Tasks

## Atomic Task List

### Feature: 修复零值检查编译错误
- [x] [Task 1] Step 1: RED — 编写 getZeroCheck bool 类型测试 —— 当 GoType 为 bool 时，断言当前返回 `== 0`（编译错误），预期返回 `!字段名`
- [x] [Task 1] Step 2: GREEN — getZeroCheck 添加 bool 分支 —— switch 增加 `case "bool"` 返回 `!字段名`
- [x] [Task 2] Step 3: RED — 编写无普通列时 generalColZeroVal 测试 —— 当全部列为主键/索引/特殊列时，断言变量不产生
- [x] [Task 2] Step 4: GREEN — 条件声明 generalColZeroVal —— 遍历列，仅在存在普通列时声明变量
- [x] Task 5: 编译验证 —— `go build ./tool/cmd/gen-go-db/...` 通过
