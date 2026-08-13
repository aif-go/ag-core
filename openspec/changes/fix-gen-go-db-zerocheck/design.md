# Design

## 变更点
### 行为变更
1. `getZeroCheck` switch 增加 `case "bool"` 分支，返回 `!字段名` 作为 bool 类型列的零值检查表达式
2. `generateListZeroValueColsMethod` 在声明 `generalColZeroVal` 前先遍历列，仅当存在普通列（非主键、非索引、非特殊）时才声明

### 修改文件
- `tool/cmd/gen-go-db/model/template.go`
  - `getZeroCheck` 函数（第 176 行后插入 bool 分支）
  - `generateListZeroValueColsMethod` 函数（第 199 行前增加普通列预检查逻辑）

## File Structure
| 文件 | 类型 | 说明 |
|------|------|------|
| `tool/cmd/gen-go-db/model/template.go` | 修改 | 修复两个零值检查编译错误 |

## Test Strategy
无新增测试文件。通过 `go build ./tool/cmd/gen-go-db/...` 验证编译通过。
