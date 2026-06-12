# DAO 使用指南

本文档介绍自动生成的 DAO 的使用方法和示例。

## 目录

- [初始化 DAO](#初始化-dao)
- [插入操作](#插入操作)
- [更新操作](#更新操作)
- [查询操作](#查询操作)
- [条件构建器](#条件构建器)
- [排序构建器](#排序构建器)
- [自定义查询](#自定义查询)

---

## 初始化 DAO

```go
import (
    "context"
    "your-project/repository/dao"
    "your-project/repository/model"
    db "github.com/aif-go/ag-core/contribute/agdb/gormdb"
    agdao "github.com/aif-go/ag-core/contribute/agdb/agdao"
    "gorm.io/gorm"
)

// 创建 DAO 实例
func createDao(db *gorm.DB) dao.ITmTeacherDao {
    repository := db.NewRepository(db)
    baseDao := agdao.NewBaseDao()
    return dao.NewTmTeacherDao(repository, baseDao)
}
```

---

## 插入操作

### InsertOne - 插入一条数据

**方法说明**：向数据库表中插入一条完整的数据记录，包含所有字段（包括零值字段）。

**参数说明**：
- `ctx context.Context` - 上下文，用于控制请求超时和取消
- `entity *model.TmTeacher` - 要插入的数据实体对象

**返回数据**：
- `int64` - 受影响的行数（成功时通常为1）
- `error` - 错误信息，成功时为nil

```go
entity := &model.TmTeacher{
    Name:      "张三",
    Phone:     "13800138000",
    ClassId:   "class001",
    CardNo:    "card123",
}

rowsAffected, err := tmTeacherDao.InsertOne(ctx, entity)
if err != nil {
    // 处理错误
}
fmt.Printf("插入成功，影响行数: %d\n", rowsAffected)
```

### InsertOneIgnoreZeroValCols - 插入数据时自动剔除零值列

**方法说明**：向数据库表中插入数据，自动忽略零值字段（空字符串、0、nil等），只插入非零值字段。

**参数说明**：
- `ctx context.Context` - 上下文，用于控制请求超时和取消
- `entity *model.TmTeacher` - 要插入的数据实体对象（零值字段会被忽略）

**返回数据**：
- `int64` - 受影响的行数（成功时通常为1）
- `error` - 错误信息，成功时为nil

```go
entity := &model.TmTeacher{
    Name:      "李四",
    Phone:     "",  // 零值，会被自动忽略
    ClassId:   "class002",
}

rowsAffected, err := tmTeacherDao.InsertOneIgnoreZeroValCols(ctx, entity)
if err != nil {
    // 处理错误
}
```

---

## 更新操作

### UpdateByPrimaryKey - 根据主键更新

**方法说明**：根据主键更新数据库中的记录，该操作只适合从数据库查询原实体修改值之后使用。会更新实体的所有字段。

**参数说明**：
- `ctx context.Context` - 上下文，用于控制请求超时和取消
- `entity *model.TmTeacher` - 要更新的数据实体对象（必须包含主键值）

**返回数据**：
- `int64` - 受影响的行数（成功时通常为1）
- `error` - 错误信息，成功时为nil

```go
// 先查询数据
entity, err := tmTeacherDao.FindByPrimaryKey(ctx, 123)
if err != nil {
    // 处理错误
}

// 修改数据
entity.Name = "王五"
entity.Phone = "13900139000"

// 更新数据
rowsAffected, err := tmTeacherDao.UpdateByPrimaryKey(ctx, entity)
if err != nil {
    // 处理错误
}
fmt.Printf("更新成功，影响行数: %d\n", rowsAffected)
```

### UpdateByPrimaryKeyIngoreZeroValCols - 根据主键更新，自动剔除零值列

**方法说明**：根据主键更新数据库中的记录，自动忽略零值字段（空字符串、0、nil等），只更新非零值字段。

**参数说明**：
- `ctx context.Context` - 上下文，用于控制请求超时和取消
- `entity *model.TmTeacher` - 要更新的数据实体对象（必须包含主键值，零值字段会被忽略）

**返回数据**：
- `int64` - 受影响的行数（成功时通常为1）
- `error` - 错误信息，成功时为nil

```go
entity, err := tmTeacherDao.FindByPrimaryKey(ctx, 123)
if err != nil {
    // 处理错误
}

// 只修改部分字段
entity.Name = "赵六"
entity.Phone = ""  // 零值，更新时会被忽略

rowsAffected, err := tmTeacherDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, entity)
if err != nil {
    // 处理错误
}
```

---

## 查询操作

### FindByPrimaryKey - 根据主键查询

**方法说明**：根据主键值查询单条数据记录。

**参数说明**：
- `ctx context.Context` - 上下文，用于控制请求超时和取消
- `id model.TmTeacherPrimaryKey` - 主键值（单主键时为具体类型，多主键时为结构体）

**返回数据**：
- `*model.TmTeacher` - 查询到的数据实体，记录不存在时为nil
- `error` - 错误信息，成功时为nil

```go
entity, err := tmTeacherDao.FindByPrimaryKey(ctx, 123)
if err != nil {
    // 处理错误
}

if entity == nil {
    fmt.Println("记录不存在")
} else {
    fmt.Printf("查询结果: %+v\n", entity)
}
```

### FindByStruct - 根据实体查询

**方法说明**：根据实体对象中的非零值字段构建查询条件，查询匹配的记录列表。会自动校验是否使用了索引列。

**参数说明**：
- `ctx context.Context` - 上下文，用于控制请求超时和取消
- `entity *model.TmTeacher` - 查询条件实体对象（零值字段会被忽略）

**返回数据**：
- `[]*model.TmTeacher` - 查询到的数据实体列表
- `error` - 错误信息，成功时为nil

```go
queryEntity := &model.TmTeacher{
    Name:    "张三",
    ClassId: "class001",
}

list, err := tmTeacherDao.FindByStruct(ctx, queryEntity)
if err != nil {
    // 处理错误
}

for _, item := range list {
    fmt.Printf("姓名: %s, 电话: %s\n", item.Name, item.Phone)
}
```

### FindByCondition - 根据条件构建器查询（支持分页和排序）

**方法说明**：使用条件构建器和排序构建器进行复杂查询，支持分页功能。会自动校验是否使用了索引列，避免全表扫描。

**参数说明**：
- `ctx context.Context` - 上下文，用于控制请求超时和取消
- `condition *conditonwhere.WhereClauseBuilder` - 查询条件构建器
- `orderBuilder *gormdb.OrderBuilder` - 排序条件构建器（可为nil）
- `page *gormdb.Page` - 分页参数（可为nil，表示不分页）

**返回数据**：
- `[]*model.TmTeacher` - 查询到的数据实体列表
- `*gormdb.PageResult` - 分页结果信息（不分页时为nil），包含：
  - `CurrentPage int64` - 当前页码
  - `PageSize int64` - 每页记录数
  - `TotalCount int64` - 总记录数
  - `TotalPage int64` - 总页数
- `error` - 错误信息，成功时为nil

```go
import (
    "github.com/aif-go/ag-core/contribute/agdb/conditonwhere"
    db "github.com/aif-go/ag-core/contribute/agdb/gormdb"
)

// 构建查询条件
condition := conditonwhere.NewWhereClauseBuilder().
    Eq("status", "active").
    And().
    Gte("age", 18).
    And().
    In("class_id", "class001", "class002", "class003")

// 构建排序条件
orderBuilder := db.NewOrderBuilder().
    Desc("created_at").
    Asc("name")

// 分页参数
page := &db.Page{
    PageNum:  1,
    PageSize: 10,
}

// 执行查询
list, pageResult, err := tmTeacherDao.FindByCondition(ctx, condition, orderBuilder, page)
if err != nil {
    // 处理错误
}

fmt.Printf("总记录数: %d, 总页数: %d, 当前页: %d\n", 
    pageResult.TotalCount, pageResult.TotalPage, pageResult.CurrentPage)

for _, item := range list {
    fmt.Printf("姓名: %s, 电话: %s\n", item.Name, item.Phone)
}
```

### FindFirstOneByCondition - 根据条件查询第一条记录

**方法说明**：使用条件构建器和排序构建器查询符合条件的第一条记录。会自动校验是否使用了索引列，避免全表扫描。

**参数说明**：
- `ctx context.Context` - 上下文，用于控制请求超时和取消
- `condition *conditonwhere.WhereClauseBuilder` - 查询条件构建器
- `orderBuilder *gormdb.OrderBuilder` - 排序条件构建器（可为nil）

**返回数据**：
- `*model.TmTeacher` - 查询到的数据实体，记录不存在时为nil
- `error` - 错误信息，成功时为nil

```go
// 构建查询条件
condition := conditonwhere.NewWhereClauseBuilder().
    Eq("phone", "13800138000").
    And().
    Eq("status", "active")

// 构建排序条件（可选）
orderBuilder := db.NewOrderBuilder().
    Desc("created_at")

// 查询第一条记录
entity, err := tmTeacherDao.FindFirstOneByCondition(ctx, condition, orderBuilder)
if err != nil {
    // 处理错误
}

if entity == nil {
    fmt.Println("未找到匹配的记录")
} else {
    fmt.Printf("查询结果: %+v\n", entity)
}
```

---

## 条件构建器

### 基础比较操作符

```go
// 等于
conditonwhere.NewWhereClauseBuilder().Eq("name", "张三")

// 不等于
conditonwhere.NewWhereClauseBuilder().Neq("status", "deleted")

// 大于
conditonwhere.NewWhereClauseBuilder().Gt("age", 18)

// 小于
conditonwhere.NewWhereClauseBuilder().Lt("age", 60)

// 大于等于
conditonwhere.NewWhereClauseBuilder().Gte("score", 80)

// 小于等于
conditonwhere.NewWhereClauseBuilder().Lte("price", 100)
```

### 集合和范围操作符

```go
// IN
conditonwhere.NewWhereClauseBuilder().In("id", 1, 2, 3, 4, 5)

// NOT IN
conditonwhere.NewWhereClauseBuilder().NotIn("status", "deleted", "banned")

// BETWEEN
conditonwhere.NewWhereClauseBuilder().Between("created_at", "2024-01-01", "2024-12-31")
```

### 逻辑连接符

```go
// AND（默认）
conditonwhere.NewWhereClauseBuilder().
    Eq("status", "active").
    And().
    Gte("age", 18)

// OR
conditonwhere.NewWhereClauseBuilder().
    Eq("status", "active").
    Or().
    Eq("status", "pending")

// 混合使用
conditonwhere.NewWhereClauseBuilder().
    Eq("type", "user").
    Or().
    Eq("type", "admin").
    And().
    Neq("status", "deleted")
```

### 条件组（嵌套）

```go
// 创建条件组
group := conditonwhere.NewWhereClauseBuilder().
    Group(
        conditonwhere.ConditionEq("age", 18).Or(),
        conditonwhere.ConditionEq("age", 19).Or(),
    )

// AND 连接的条件组
andGroup := conditonwhere.NewWhereClauseBuilder().
    AndGroup(
        conditonwhere.ConditionEq("age", 18),
        conditonwhere.ConditionGte("score", 80),
    )

// OR 连接的条件组
orGroup := conditonwhere.NewWhereClauseBuilder().
    OrGroup(
        conditonwhere.ConditionEq("age", 18),
        conditonwhere.ConditionEq("age", 19),
    )
```

### 复杂条件示例

```go
// WHERE (status = 'active' OR status = 'pending') AND age >= 18
condition := conditonwhere.NewWhereClauseBuilder().
    Group(
        conditonwhere.ConditionEq("status", "active").Or(),
        conditonwhere.ConditionEq("status", "pending").Or(),
    ).
    And().
    Gte("age", 18)

// WHERE type = 'user' OR (type = 'admin' AND status != 'deleted')
condition := conditonwhere.NewWhereClauseBuilder().
    Eq("type", "user").
    Or().
    Group(
        conditonwhere.ConditionEq("type", "admin"),
        conditonwhere.ConditionNeq("status", "deleted"),
    )
```

---

## 排序构建器

### Asc - 添加升序排序

**参数说明**：
- `colName string` - 列名

**返回数据**：`*OrderBuilder` - 返回构建器本身，支持链式调用

### Desc - 添加降序排序

**参数说明**：
- `colName string` - 列名

**返回数据**：`*OrderBuilder` - 返回构建器本身，支持链式调用

### Order - 添加自定义排序

**参数说明**：
- `colName string` - 列名
- `sort OrderSort` - 排序方向（ASC 或 DESC）

**返回数据**：`*OrderBuilder` - 返回构建器本身，支持链式调用

### Orders - 批量添加排序

**参数说明**：
- `orders ...Order` - Order 对象列表（可变参数）

**返回数据**：`*OrderBuilder` - 返回构建器本身，支持链式调用

### Build - 构建完整的 ORDER BY SQL 语句

**参数说明**：无

**返回数据**：
- `string` - 完整的 ORDER BY SQL 语句（如 "ORDER BY id ASC, created_at DESC"）

### BuildWithoutKeyword - 构建排序 SQL 语句（不包含 ORDER BY 关键字）

**参数说明**：无

**返回数据**：
- `string` - 排序 SQL 语句（如 "id ASC, created_at DESC"）

### ToOrders - 转换为 Order 切片

**参数说明**：无

**返回数据**：
- `[]Order` - Order 切片

### Clear - 清空所有排序条件

**参数说明**：无

**返回数据**：`*OrderBuilder` - 返回构建器本身，支持链式调用

```go
import db "github.com/aif-go/ag-core/contribute/agdb/gormdb"

// 升序
orderBuilder := db.NewOrderBuilder().
    Asc("id")

// 降序
orderBuilder := db.NewOrderBuilder().
    Desc("created_at")

// 多字段排序
orderBuilder := db.NewOrderBuilder().
    Asc("status").
    Desc("created_at").
    Asc("id")
// 生成: ORDER BY status ASC, created_at DESC, id ASC

// 自定义排序
orderBuilder := db.NewOrderBuilder().
    Order("name", db.ASC).
    Order("age", db.DESC)

// 批量添加排序
orderBuilder := db.NewOrderBuilder().
    Orders(
        db.Order{ColName: "id", Sort: db.ASC},
        db.Order{ColName: "created_at", Sort: db.DESC},
    )

// 包含 ORDER BY 关键字
sql := orderBuilder.Build()
// 结果: ORDER BY id ASC, created_at DESC

// 不包含 ORDER BY 关键字
sql := orderBuilder.BuildWithoutKeyword()
// 结果: id ASC, created_at DESC

// 转换为 Order 切片
orders := orderBuilder.ToOrders()

// 清空所有排序条件
orderBuilder.Clear()
```

---

## 自定义查询

### FindByCustomerRule - 根据自定义规则查询

**方法说明**：使用预定义的命名 SQL 进行查询，支持自定义查询逻辑。需要预先在命名 SQL 映射中注册 SQL。

**参数说明**：
- `ctx context.Context` - 上下文，用于控制请求超时和取消
- `namingInfo *gormdb.NameingSqlArgInfo` - 命名 SQL 参数信息，包含：
  - `SqlName string` - SQL 名称
  - `ReqType any` - 请求参数类型（用于类型校验）
- `args any` - 查询参数（根据 ReqType 指定的类型）

**返回数据**：
- `any` - 查询结果，根据 SQL 配置可能是分页结果或列表
- `error` - 错误信息，成功时为nil

```go
import db "github.com/aif-go/ag-core/contribute/agdb/gormdb"

// 创建命名 SQL 参数
namingInfo := &db.NameingSqlArgInfo{
    SqlName: "FindByPhone",
    ReqType: &model.TmTeacherFindByPhoneArg{},
}

// 创建查询参数
queryArgs := &model.TmTeacherFindByPhoneArg{
    Phone: "13800138000",
}

// 执行查询
result, err := tmTeacherDao.FindByCustomerRule(ctx, namingInfo, queryArgs)
if err != nil {
    // 处理错误
}

// 根据查询类型处理结果
if pageResult, ok := result.(*model.TmTeacherFindByPhonePageRes); ok {
    // 处理分页结果
    fmt.Printf("总记录数: %d\n", pageResult.TotalCount)
    for _, item := range pageResult.ResultList {
        fmt.Printf("姓名: %s\n", item.Name)
    }
}
```

---

## 完整示例

### 示例 1：插入并查询

```go
func InsertAndQuery(ctx context.Context, tmTeacherDao dao.ITmTeacherDao) error {
    // 1. 插入数据
    entity := &model.TmTeacher{
        Name:      "张三",
        Phone:     "13800138000",
        ClassId:   "class001",
        CardNo:    "card123",
    }
    
    _, err := tmTeacherDao.InsertOne(ctx, entity)
    if err != nil {
        return fmt.Errorf("插入失败: %w", err)
    }
    
    // 2. 根据主键查询
    result, err := tmTeacherDao.FindByPrimaryKey(ctx, entity.Id)
    if err != nil {
        return fmt.Errorf("查询失败: %w", err)
    }
    
    fmt.Printf("查询结果: %+v\n", result)
    return nil
}
```

### 示例 2：条件查询和分页

```go
func QueryWithCondition(ctx context.Context, tmTeacherDao dao.ITmTeacherDao) error {
    // 构建查询条件
    condition := conditonwhere.NewWhereClauseBuilder().
        Eq("class_id", "class001").
        And().
        Gte("jpa_version", 0)
    
    // 构建排序条件
    orderBuilder := db.NewOrderBuilder().
        Desc("create_time")
    
    // 分页参数
    page := &db.Page{
        PageNum:  1,
        PageSize: 20,
    }
    
    // 执行查询
    list, pageResult, err := tmTeacherDao.FindByCondition(ctx, condition, orderBuilder, page)
    if err != nil {
        return fmt.Errorf("查询失败: %w", err)
    }
    
    fmt.Printf("总记录数: %d, 总页数: %d\n", 
        pageResult.TotalCount, pageResult.TotalPage)
    
    for _, item := range list {
        fmt.Printf("ID: %d, 姓名: %s, 电话: %s\n", 
            item.Id, item.Name, item.Phone)
    }
    
    return nil
}
```

### 示例 3：更新数据

```go
func UpdateTeacher(ctx context.Context, tmTeacherDao dao.ITmTeacherDao) error {
    // 1. 查询数据
    entity, err := tmTeacherDao.FindByPrimaryKey(ctx, 123)
    if err != nil {
        return fmt.Errorf("查询失败: %w", err)
    }
    
    if entity == nil {
        return fmt.Errorf("记录不存在")
    }
    
    // 2. 修改数据
    entity.Name = "新名字"
    entity.Phone = "13900139000"
    
    // 3. 更新数据
    rowsAffected, err := tmTeacherDao.UpdateByPrimaryKey(ctx, entity)
    if err != nil {
        return fmt.Errorf("更新失败: %w", err)
    }
    
    fmt.Printf("更新成功，影响行数: %d\n", rowsAffected)
    return nil
}
```

### 示例 4：复杂条件查询

```go
func ComplexQuery(ctx context.Context, tmTeacherDao dao.ITmTeacherDao) error {
    // 构建复杂条件：WHERE (class_id = 'class001' OR class_id = 'class002') AND jpa_version >= 0
    condition := conditonwhere.NewWhereClauseBuilder().
        OrGroup(
            conditonwhere.ConditionEq("class_id", "class001"),
            conditonwhere.ConditionEq("class_id", "class002"),
        ).
        And().
        Gte("jpa_version", 0)
    
    // 排序：按创建时间降序，按ID升序
    orderBuilder := db.NewOrderBuilder().
        Desc("create_time").
        Asc("id")
    
    // 查询第一条记录
    entity, err := tmTeacherDao.FindFirstOneByCondition(ctx, condition, orderBuilder)
    if err != nil {
        return fmt.Errorf("查询失败: %w", err)
    }
    
    if entity == nil {
        fmt.Println("未找到匹配的记录")
    } else {
        fmt.Printf("查询结果: %+v\n", entity)
    }
    
    return nil
}
```

---

## 注意事项

1. **索引校验**：自定义查询会自动校验是否使用了索引列，避免全表扫描
2. **零值处理**：`InsertOneIgnoreZeroValCols` 和 `UpdateByPrimaryKeyIngoreZeroValCols` 会自动忽略零值列
3. **分页查询**：`FindByCondition` 支持分页，返回分页结果信息
4. **条件构建**：使用 `WhereClauseBuilder` 可以灵活构建复杂的查询条件
5. **排序构建**：使用 `OrderBuilder` 可以链式构建排序条件
6. **事务处理**：DAO 方法需要配合事务使用，具体请参考事务相关文档

---

## API 参考

### ITmTeacherDao 接口

| 方法 | 说明 |
|------|------|
| `InsertOne` | 插入一条数据 |
| `InsertOneIgnoreZeroValCols` | 插入数据时自动剔除零值列 |
| `UpdateByPrimaryKey` | 根据主键更新 |
| `UpdateByPrimaryKeyIngoreZeroValCols` | 根据主键更新，自动剔除零值列 |
| `FindByPrimaryKey` | 根据主键查询 |
| `FindByStruct` | 根据实体查询 |
| `FindByCustomerRule` | 根据自定义规则查询 |
| `FindByCondition` | 根据条件构建器查询（支持分页和排序） |
| `FindFirstOneByCondition` | 根据条件查询第一条记录 |

### WhereClauseBuilder 方法

| 方法 | 说明 |
|------|------|
| `Eq(field, value)` | 添加等于条件 |
| `Neq(field, value)` | 添加不等于条件 |
| `Gt(field, value)` | 添加大于条件 |
| `Lt(field, value)` | 添加小于条件 |
| `Gte(field, value)` | 添加大于等于条件 |
| `Lte(field, value)` | 添加小于等于条件 |
| `In(field, values...)` | 添加 IN 条件 |
| `NotIn(field, values...)` | 添加 NOT IN 条件 |
| `Between(field, min, max)` | 添加 BETWEEN 条件 |
| `Or()` | 设置下一个条件的逻辑为 OR |
| `And()` | 设置下一个条件的逻辑为 AND |
| `Group(conditions...)` | 添加一个条件组（用于嵌套） |
| `AndGroup(conditions...)` | 添加一个 AND 连接的条件组 |
| `OrGroup(conditions...)` | 添加一个 OR 连接的条件组 |
| `Build()` | 构建最终的 WHERE SQL 和参数 |

### OrderBuilder 方法

| 方法 | 说明 |
|------|------|
| `Asc(colName)` | 添加升序排序 |
| `Desc(colName)` | 添加降序排序 |
| `Order(colName, sort)` | 添加自定义排序 |
| `Orders(orders...)` | 批量添加排序 |
| `Build()` | 构建完整的 ORDER BY SQL 语句 |
| `BuildWithoutKeyword()` | 构建排序 SQL 语句（不包含 ORDER BY 关键字） |
| `ToOrders()` | 转换为 Order 切片 |
| `Clear()` | 清空所有排序条件 |
