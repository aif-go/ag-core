# WhereClauseBuilder 动态链式使用指南

## 快速开始

```go
builder := conditonwhere.NewWhereClauseBuilder()
sql, args, _ := builder.Eq("name", "test").Build()
// sql: "name = ?", args: ["test"]
```

## 基础方法

### 条件方法

| 方法 | SQL 操作符 | 示例 |
|------|-----------|------|
| `Eq(field, value)` | `=` | `Eq("name", "tom")` → `name = ?` |
| `Neq(field, value)` | `!=` | `Neq("status", "deleted")` → `status != ?` |
| `Gt(field, value)` | `>` | `Gt("age", 18)` → `age > ?` |
| `Lt(field, value)` | `<` | `Lt("price", 100)` → `price < ?` |
| `Gte(field, value)` | `>=` | `Gte("score", 60)` → `score >= ?` |
| `Lte(field, value)` | `<=` | `Lte("count", 10)` → `count <= ?` |
| `In(field, values...)` | `IN` | `In("status", "a", "b")` → `status IN (?, ?)` |
| `NotIn(field, values...)` | `NOT IN` | `NotIn("type", "x", "y")` → `type NOT IN (?, ?)` |
| `Between(field, min, max)` | `BETWEEN` | `Between("age", 18, 65)` → `age BETWEEN ? AND ?` |

### 逻辑控制

| 方法 | 说明 | 示例 |
|------|------|------|
| `Or()` | 设置下一个条件为 OR | `Eq("a",1).Or().Eq("b",2)` → `a = ? OR b = ?` |
| `And()` | 设置下一个条件为 AND（默认） | `Eq("a",1).And().Eq("b",2)` → `a = ? AND b = ?` |

### 嵌套控制

| 方法 | 说明 | 示例 |
|------|------|------|
| `BeginGroup()` | 开始嵌套组 | 见下文 |
| `EndGroup()` | 结束嵌套组 | 见下文 |

---

## 链式调用示例

### 1. 简单 AND 条件

```go
builder := NewWhereClauseBuilder().
    Eq("status", "active").
    Gt("age", 18).
    Eq("type", "user")

sql, args, _ := builder.Build()
// sql: "status = ? AND age > ? AND type = ?"
// args: ["active", 18, "user"]
```

### 2. 使用 Or()

```go
builder := NewWhereClauseBuilder().
    Eq("status", "active").
    Or().
    Eq("status", "pending")

sql, args, _ := builder.Build()
// sql: "status = ? OR status = ?"
// args: ["active", "pending"]
```

### 3. 混合 AND/OR

```go
builder := NewWhereClauseBuilder().
    Eq("a", 1).
    Or().
    Eq("b", 2).
    And().
    Eq("c", 3)

sql, args, _ := builder.Build()
// sql: "(a = ? OR b = ? AND c = ?)"
// args: [1, 2, 3]
```

### 4. 嵌套组 BeginGroup/EndGroup

```go
builder := NewWhereClauseBuilder().
    Eq("a", 1).
    BeginGroup().
    Eq("b", 2).
    Or().
    Eq("c", 3).
    EndGroup()

sql, args, _ := builder.Build()
// sql: "a = ? AND (b = ? OR c = ?)"
// args: [1, 2, 3]
```

### 5. 多层嵌套

```go
builder := NewWhereClauseBuilder().
    Eq("a", 1).
    BeginGroup().
    Eq("b", 2).
    BeginGroup().
    Eq("c", 3).
    Or().
    Eq("d", 4).
    EndGroup().
    EndGroup()

sql, args, _ := builder.Build()
// sql: "a = ? AND (b = ? AND (c = ? OR d = ?))"
// args: [1, 2, 3, 4]
```

### 6. OR 后接嵌套组

```go
builder := NewWhereClauseBuilder().
    Eq("a", 1).
    Or().
    BeginGroup().
    Eq("b", 2).
    And().
    Eq("c", 3).
    EndGroup()

sql, args, _ := builder.Build()
// sql: "a = ? OR (b = ? AND c = ?)"
// args: [1, 2, 3]
```

### 7. 连续嵌套组

```go
builder := NewWhereClauseBuilder().
    Eq("a", 1).
    BeginGroup().
    Eq("b", 2).
    Or().
    Eq("c", 3).
    EndGroup().
    BeginGroup().
    Eq("d", 4).
    And().
    Eq("e", 5).
    EndGroup()

sql, args, _ := builder.Build()
// sql: "a = ? AND (b = ? OR c = ?) AND (d = ? AND e = ?)"
// args: [1, 2, 3, 4, 5]
```

### 8. IN 查询

```go
builder := NewWhereClauseBuilder().
    In("status", "active", "pending", "draft").
    Gt("create_time", "2024-01-01")

sql, args, _ := builder.Build()
// sql: "status IN (?, ?, ?) AND create_time > ?"
// args: ["active", "pending", "draft", "2024-01-01"]
```

### 9. BETWEEN 查询

```go
builder := NewWhereClauseBuilder().
    Between("age", 18, 65).
    Eq("status", "active")

sql, args, _ := builder.Build()
// sql: "age BETWEEN ? AND ? AND status = ?"
// args: [18, 65, "active"]
```

---

## 进阶用法

### 索引检查 BuildWithIndexCheck

```go
builder := NewWhereClauseBuilder().Eq("id", 100)

sql, args, index, err := builder.BuildWithIndexCheck(
    []string{"id"},                    // 主键列
    [][]string{{"user_id", "name"}}, // 索引列
)
// sql: "WHERE id = ?"
// index: "PRIMARY"
```

### 错误情况

```go
builder := NewWhereClauseBuilder().Eq("name", "test")

sql, args, index, err := builder.BuildWithIndexCheck(
    []string{"id"},
    [][]string{{"user_id", "name"}},
)
// err: "query does not use any index or primary key. used fields: [name]"
```

---

## 方法对照表

| 需求 | 推荐方法 |
|------|----------|
| 简单条件 | `Eq`, `Neq`, `Gt`, `Lt` 等 |
| OR 逻辑 | `Or()` |
| 单层嵌套 | `BeginGroup().EndGroup()` |
| 多层嵌套 | 多次 `BeginGroup().EndGroup()` |
| IN 查询 | `In()` |
| 范围查询 | `Between()` |
| 批量条件 | `AddConditions()` |
| 预定义条件 | `ConditionEq()` 等 + `Group()` |

---

## 完整示例

```go
package main

import (
    "fmt"
    "github.com/yourpkg/conditonwhere"
)

func main() {
    // 示例：用户查询
    builder := conditonwhere.NewWhereClauseBuilder().
        Eq("status", "active").
        Gte("age", 18).
        In("role", "admin", "user").
        BeginGroup().
        Eq("city", "beijing").
        Or().
        Eq("city", "shanghai").
        EndGroup().
        Between("create_time", "2024-01-01", "2024-12-31")

    sql, args, err := builder.Build()
    if err != nil {
        panic(err)
    }

    fmt.Println("SQL:", sql)
    fmt.Println("Args:", args)
}
```

输出：
```
SQL: status = ? AND age >= ? AND role IN (?, ?, ?) AND (city = ? OR city = ?) AND create_time BETWEEN ? AND ?
Args: [active 18 admin user beijing shanghai 2024-01-01 2024-12-31]
```