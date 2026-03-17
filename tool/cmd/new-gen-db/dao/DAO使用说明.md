# DAO 使用说明文档

本文档描述了自动生成的 DAO 接口中各方法的使用方法和示例代码。

---

## 目录

- [DAO 使用说明文档](#dao-使用说明文档)
  - [目录](#目录)
  - [InsertOne - 插入单条数据](#insertone---插入单条数据)
    - [方法签名](#方法签名)
    - [说明](#说明)
    - [参数](#参数)
    - [返回值](#返回值)
    - [示例代码](#示例代码)
  - [InsertOneIgnoreZeroValCols - 插入单条数据（忽略零值列）](#insertoneignorezerovalcols---插入单条数据忽略零值列)
    - [方法签名](#方法签名-1)
    - [说明](#说明-1)
    - [参数](#参数-1)
    - [返回值](#返回值-1)
    - [示例代码](#示例代码-1)
  - [UpdateByPrimaryKey - 根据主键更新](#updatebyprimarykey---根据主键更新)
    - [方法签名](#方法签名-2)
    - [说明](#说明-2)
    - [参数](#参数-2)
    - [返回值](#返回值-2)
    - [示例代码](#示例代码-2)
  - [UpdateByPrimaryKeyIngoreZeroValCols - 根据主键更新（忽略零值列）](#updatebyprimarykeyingorezerovalcols---根据主键更新忽略零值列)
    - [方法签名](#方法签名-3)
    - [说明](#说明-3)
    - [参数](#参数-3)
    - [返回值](#返回值-3)
    - [示例代码](#示例代码-3)
  - [UpdateDynamic - 动态列更新](#updatedynamic---动态列更新)
    - [方法签名](#方法签名-4)
    - [说明](#说明-4)
    - [参数](#参数-4)
    - [返回值](#返回值-4)
    - [示例代码](#示例代码-4)
  - [FindByPrimaryKey - 根据主键查询](#findbyprimarykey---根据主键查询)
    - [方法签名](#方法签名-5)
    - [说明](#说明-5)
    - [参数](#参数-5)
    - [返回值](#返回值-5)
    - [示例代码](#示例代码-5)
  - [FindByStruct - 根据实体查询](#findbystruct---根据实体查询)
    - [方法签名](#方法签名-6)
    - [说明](#说明-6)
    - [参数](#参数-6)
    - [返回值](#返回值-6)
    - [示例代码](#示例代码-6)
    - [索引使用说明](#索引使用说明)
  - [FindByCustomerRule - 根据自定义规则查询](#findbycustomerrule---根据自定义规则查询)
    - [方法签名](#方法签名-7)
    - [说明](#说明-7)
    - [参数](#参数-7)
    - [返回值](#返回值-7)
    - [示例代码](#示例代码-7)
  - [FindByPage - 分页查询](#findbypage---分页查询)
    - [方法签名](#方法签名-8)
    - [说明](#说明-8)
    - [参数](#参数-8)
    - [返回值](#返回值-8)
    - [示例代码](#示例代码-8)
  - [FindByWhere - 根据WHERE条件查询](#findbywhere---根据where条件查询)
    - [方法签名](#方法签名-9)
    - [说明](#说明-9)
    - [参数](#参数-9)
    - [返回值](#返回值-9)
    - [示例代码](#示例代码-9)
    - [复杂条件示例](#复杂条件示例)
    - [索引验证](#索引验证)
  - [注意事项](#注意事项)
    - [1. 索引使用](#1-索引使用)
    - [2. 事务支持](#2-事务支持)
    - [3. 表名动态替换](#3-表名动态替换)
    - [4. 错误处理](#4-错误处理)
  - [附录：WhereClauseBuilder 使用指南](#附录whereclausebuilder-使用指南)
    - [基本条件](#基本条件)
    - [逻辑组合](#逻辑组合)
    - [链式调用](#链式调用)

---

## InsertOne - 插入单条数据

### 方法签名

```go
InsertOne(ctx context.Context, entity *model.{TableName}) (int64, error)
```

### 说明

插入一条完整的数据库记录，包含所有字段（包括零值）。

### 参数

- `ctx`: 上下文
- `entity`: 要插入的实体对象

### 返回值

- `int64`: 受影响的行数
- `error`: 错误信息

### 示例代码

```go
package main

import (
    "context"
    "fmt"
    "ag-core/tool/cmd/new-gen-db/repository/dao"
    "ag-core/tool/cmd/new-gen-db/repository/model"
)

func main() {
    // 创建DAO实例
    tblPurchaseResponseDao := dao.NewTblPurchaseResponseDao(repository, baseDao)

    // 创建实体
    entity := &model.TblPurchaseResponse{
        OrderId:              "ORD123456",
        MerchantId:             "MERCHANT001",
        TransactionType:        "02",
        SettlementAmount:       10000,
        ResponseCode:          "00",
    }

    // 插入数据
    ctx := context.Background()
    rowsAffected, err := tblPurchaseResponseDao.InsertOne(ctx, entity)
    if err != nil {
        fmt.Printf("插入失败: %v\n", err)
        return
    }

    fmt.Printf("插入成功，影响行数: %d\n", rowsAffected)
}
```

---

## InsertOneIgnoreZeroValCols - 插入单条数据（忽略零值列）

### 方法签名

```go
InsertOneIgnoreZeroValCols(ctx context.Context, entity *model.{TableName}) (int64, error)
```

### 说明

插入一条数据库记录，自动剔除零值的列。适用于部分字段插入的场景。

### 参数

- `ctx`: 上下文
- `entity`: 要插入的实体对象

### 返回值

- `int64`: 受影响的行数
- `error`: 错误信息

### 示例代码

```go
package main

import (
    "context"
    "fmt"
    "ag-core/tool/cmd/new-gen-db/repository/dao"
    "ag-core/tool/cmd/new-gen-db/repository/model"
)

func main() {
    // 创建DAO实例
    tblPurchaseResponseDao := dao.NewTblPurchaseResponseDao(repository, baseDao)

    // 创建实体（只设置部分字段）
    entity := &model.TblPurchaseResponse{
        OrderId:              "ORD123456",
        MerchantId:             "MERCHANT001",
        TransactionType:        "02",
        // 其他字段保持零值，将不会被插入
    }

    // 插入数据（自动忽略零值列）
    ctx := context.Background()
    rowsAffected, err := tblPurchaseResponseDao.InsertOneIgnoreZeroValCols(ctx, entity)
    if err != nil {
        fmt.Printf("插入失败: %v\n", err)
        return
    }

    fmt.Printf("插入成功，影响行数: %d\n", rowsAffected)
}
```

---

## UpdateByPrimaryKey - 根据主键更新

### 方法签名

```go
UpdateByPrimaryKey(ctx context.Context, entity *model.{TableName}) (int64, error)
```

### 说明

根据主键更新数据库记录，更新所有字段（包括零值）。**注意：该方法只适合从数据库查询原实体修改值之后使用。**

### 参数

- `ctx`: 上下文
- `entity`: 要更新的实体对象（必须包含主键值）

### 返回值

- `int64`: 受影响的行数
- `error`: 错误信息

### 示例代码

```go
package main

import (
    "context"
    "fmt"
    "ag-core/tool/cmd/new-gen-db/repository/dao"
    "ag-core/tool/cmd/new-gen-db/repository/model"
)

func main() {
    // 创建DAO实例
    tblPurchaseResponseDao := dao.NewTblPurchaseResponseDao(repository, baseDao)

    // 先查询原数据
    ctx := context.Background()
    entity, err := tblPurchaseResponseDao.FindByPrimaryKey(ctx, 123)
    if err != nil {
        fmt.Printf("查询失败: %v\n", err)
        return
    }
    if entity == nil {
        fmt.Println("记录不存在")
        return
    }

    // 修改数据
    entity.ResponseCode = "99"
    entity.ResultMessage = "更新后的消息"

    // 更新数据
    rowsAffected, err := tblPurchaseResponseDao.UpdateByPrimaryKey(ctx, entity)
    if err != nil {
        fmt.Printf("更新失败: %v\n", err)
        return
    }

    fmt.Printf("更新成功，影响行数: %d\n", rowsAffected)
}
```

---

## UpdateByPrimaryKeyIngoreZeroValCols - 根据主键更新（忽略零值列）

### 方法签名

```go
UpdateByPrimaryKeyIngoreZeroValCols(ctx context.Context, entity *model.{TableName}) (int64, error)
```

### 说明

根据主键更新数据库记录，自动剔除零值的列。适用于部分字段更新的场景。

### 参数

- `ctx`: 上下文
- `entity`: 要更新的实体对象（必须包含主键值）

### 返回值

- `int64`: 受影响的行数
- `error`: 错误信息

### 示例代码

```go
package main

import (
    "context"
    "fmt"
    "ag-core/tool/cmd/new-gen-db/repository/dao"
    "ag-core/tool/cmd/new-gen-db/repository/model"
)

func main() {
    // 创建DAO实例
    tblPurchaseResponseDao := dao.NewTblPurchaseResponseDao(repository, baseDao)

    // 创建更新实体（只设置需要更新的字段）
    entity := &model.TblPurchaseResponse{
        Id:            123,  // 必须设置主键
        ResponseCode:   "99",  // 只更新这个字段
    }

    // 更新数据（自动忽略零值列）
    ctx := context.Background()
    rowsAffected, err := tblPurchaseResponseDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, entity)
    if err != nil {
        fmt.Printf("更新失败: %v\n", err)
        return
    }

    fmt.Printf("更新成功，影响行数: %d\n", rowsAffected)
}
```

---

## UpdateDynamic - 动态列更新

### 方法签名

```go
UpdateDynamic(ctx context.Context, entity *model.{TableName}, cols []string) (int64, error)
```

### 说明

根据主键动态更新指定的列。可以精确控制更新哪些字段。

### 参数

- `ctx`: 上下文
- `entity`: 要更新的实体对象（必须包含主键值和要更新的字段值）
- `cols`: 要更新的列名列表

### 返回值

- `int64`: 受影响的行数
- `error`: 错误信息

### 示例代码

```go
package main

import (
    "context"
    "fmt"
    "ag-core/tool/cmd/new-gen-db/repository/dao"
    "ag-core/tool/cmd/new-gen-db/repository/model"
)

func main() {
    // 创建DAO实例
    tblPurchaseResponseDao := dao.NewTblPurchaseResponseDao(repository, baseDao)

    // 创建更新实体
    entity := &model.TblPurchaseResponse{
        Id:            123,
        ResponseCode:   "99",
        ResultMessage:  "更新后的消息",
        RiskLevel:     "H",  // 这个字段不会被更新，因为不在cols中
    }

    // 指定只更新这两个列
    cols := []string{"ResponseCode", "ResultMessage"}

    // 动态更新
    ctx := context.Background()
    rowsAffected, err := tblPurchaseResponseDao.UpdateDynamic(ctx, entity, cols)
    if err != nil {
        fmt.Printf("更新失败: %v\n", err)
        return
    }

    fmt.Printf("更新成功，影响行数: %d\n", rowsAffected)
}
```

---

## FindByPrimaryKey - 根据主键查询

### 方法签名

```go
FindByPrimaryKey(ctx context.Context, id model.{TableName}PrimaryKey) (*model.{TableName}, error)
```

### 说明

根据主键查询单条记录。如果记录不存在，返回 nil 和 nil 错误。

### 参数

- `ctx`: 上下文
- `id`: 主键值

### 返回值

- `*model.{TableName}`: 查询到的实体对象，不存在时为 nil
- `error`: 错误信息

### 示例代码

```go
package main

import (
    "context"
    "fmt"
    "ag-core/tool/cmd/new-gen-db/repository/dao"
    "ag-core/tool/cmd/new-gen-db/repository/model"
)

func main() {
    // 创建DAO实例
    tblPurchaseResponseDao := dao.NewTblPurchaseResponseDao(repository, baseDao)

    // 根据主键查询
    ctx := context.Background()
    entity, err := tblPurchaseResponseDao.FindByPrimaryKey(ctx, 123)
    if err != nil {
        fmt.Printf("查询失败: %v\n", err)
        return
    }

    if entity == nil {
        fmt.Println("记录不存在")
        return
    }

    fmt.Printf("查询成功: %+v\n", entity)
}
```

---

## FindByStruct - 根据实体查询

### 方法签名

```go
FindByStruct(ctx context.Context, entity *model.{TableName}) ([]*model.{TableName}, error)
```

### 说明

根据实体中的非零值字段进行查询。**必须使用索引（主键或索引的引导列）**，否则返回错误。

### 参数

- `ctx`: 上下文
- `entity`: 查询条件实体（非零值字段作为查询条件）

### 返回值

- `[]*model.{TableName}`: 查询结果列表
- `error`: 错误信息

### 示例代码

```go
package main

import (
    "context"
    "fmt"
    "ag-core/tool/cmd/new-gen-db/repository/dao"
    "ag-core/tool/cmd/new-gen-db/repository/model"
)

func main() {
    // 创建DAO实例
    tblPurchaseResponseDao := dao.NewTblPurchaseResponseDao(repository, baseDao)

    // 使用索引 IDX_ORDER_ID 查询（OrderId 是引导列）
    entity := &model.TblPurchaseResponse{
        OrderId:       "ORD123456",
        MerchantId:      "MERCHANT001",
        TransactionType: "02",
    }

    // 查询数据
    ctx := context.Background()
    list, err := tblPurchaseResponseDao.FindByStruct(ctx, entity)
    if err != nil {
        fmt.Printf("查询失败: %v\n", err)
        return
    }

    fmt.Printf("查询成功，共 %d 条记录\n", len(list))
    for _, item := range list {
        fmt.Printf("- ID: %d, OrderId: %s\n", item.Id, item.OrderId)
    }
}
```

### 索引使用说明

`FindByStruct` 方法会自动检查是否使用了索引。查询条件必须满足以下条件之一：

1. **使用主键**：主键的引导列有值
2. **使用索引**：任意索引的引导列有值

**支持的索引示例**（以 TblPurchaseResponse 表为例）：

| 索引名称 | 引导列 | 其他列 |
|-----------|---------|--------|
| PRIMARY | ID | - |
| IDX_ORDER_ID | OrderId | MerchantId, TransactionType |
| IDX_STAN | TransmissionDatetime | Stan, TransactionType |
| IDX_RRN | RetrievalReferenceNumber | MerchantId, TransactionType |
| IDX_INSERT_TIMESTATMP | InsertTimestamp | - |

---

## FindByCustomerRule - 根据自定义规则查询

### 方法签名

```go
FindByCustomerRule(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) (any, error)
```

### 说明

根据 YAML 配置文件中定义的自定义查询规则执行查询。支持分页查询。

### 参数

- `ctx`: 上下文
- `namingInfo`: 命名SQL参数信息（包含 SqlName 和 ReqType）
- `args`: 查询参数对象

### 返回值

- `any`: 查询结果（分页查询返回 PageResult，普通查询返回列表）
- `error`: 错误信息

### 示例代码

```go
package main

import (
    "context"
    "fmt"
    "ag-core/tool/cmd/new-gen-db/repository/dao"
    "ag-core/tool/cmd/new-gen-db/repository/model"
    "ag-core/contribute/agdb/gormdb"
)

func main() {
    // 创建DAO实例
    tblPurchaseResponseDao := dao.NewTblPurchaseResponseDao(repository, baseDao)

    // 创建命名SQL参数
    namingInfo := &gormdb.NameingSqlArgInfo{
        SqlName: "FindDuplicatePurchase",
        ReqType: model.TblPurchaseResponseFindDuplicatePurchaseArg{},
    }

    // 创建查询参数
    args := &model.TblPurchaseResponseFindDuplicatePurchaseArg{
        OrderId:              "ORD123456",
        MerchantId:             "MERCHANT001",
        TransactionType:        "02",
    }

    // 执行查询
    ctx := context.Background()
    result, err := tblPurchaseResponseDao.FindByCustomerRule(ctx, namingInfo, args)
    if err != nil {
        fmt.Printf("查询失败: %v\n", err)
        return
    }

    // 处理结果
    if pageRes, ok := result.(*model.TblPurchaseResponseFindDuplicatePurchasePageRes); ok {
        fmt.Printf("查询成功，共 %d 条记录\n", pageRes.TotalCount)
        for _, item := range pageRes.ResultList {
            fmt.Printf("- ID: %d, OrderId: %s\n", item.Id, item.OrderId)
        }
    }
}
```

---

## FindByPage - 分页查询

### 方法签名

```go
FindByPage(ctx context.Context, entity *model.{TableName}, page gormdb.Page, orders []gormdb.Order) ([]*model.{TableName}, *gormdb.PageResult, error)
```

### 说明

根据实体条件进行分页查询，支持排序。**必须使用索引（主键或索引的引导列）**，否则返回错误。

### 参数

- `ctx`: 上下文
- `entity`: 查询条件实体（非零值字段作为查询条件）
- `page`: 分页参数
- `orders`: 排序条件列表

### 返回值

- `[]*model.{TableName}`: 查询结果列表
- `*gormdb.PageResult`: 分页结果信息
- `error`: 错误信息

### 示例代码

```go
package main

import (
    "context"
    "fmt"
    "ag-core/tool/cmd/new-gen-db/repository/dao"
    "ag-core/tool/cmd/new-gen-db/repository/model"
    "ag-core/contribute/agdb/gormdb"
)

func main() {
    // 创建DAO实例
    tblPurchaseResponseDao := dao.NewTblPurchaseResponseDao(repository, baseDao)

    // 使用索引 IDX_ORDER_ID 查询
    entity := &model.TblPurchaseResponse{
        OrderId:       "ORD123456",
        MerchantId:      "MERCHANT001",
        TransactionType: "02",
    }

    // 设置分页参数
    page := gormdb.Page{
        PageNum:  1,
        PageSize: 10,
    }

    // 设置排序
    orders := []gormdb.Order{
        {Column: "InsertTimestamp", Direction: "DESC"},
    }

    // 分页查询
    ctx := context.Background()
    list, pageResult, err := tblPurchaseResponseDao.FindByPage(ctx, entity, page, orders)
    if err != nil {
        fmt.Printf("查询失败: %v\n", err)
        return
    }

    fmt.Printf("查询成功，共 %d 条记录，当前页 %d/%d\n", 
        pageResult.TotalCount, pageResult.CurrentPage, pageResult.TotalPage)
    for _, item := range list {
        fmt.Printf("- ID: %d, OrderId: %s\n", item.Id, item.OrderId)
    }
}
```

---

## FindByWhere - 根据WHERE条件查询

### 方法签名

```go
FindByWhere(ctx context.Context, orders []gormdb.Order, whereClauseBuilder *conditonwhere.WhereClauseBuilder) ([]*model.{TableName}, error)
```

### 说明

使用 `WhereClauseBuilder` 构建复杂的WHERE条件进行查询，支持排序。**必须使用索引（主键或索引的引导列）**，否则返回错误。

### 参数

- `ctx`: 上下文
- `orders`: 排序条件列表
- `whereClauseBuilder`: WHERE条件构建器

### 返回值

- `[]*model.{TableName}`: 查询结果列表
- `error`: 错误信息

### 示例代码

```go
package main

import (
    "context"
    "fmt"
    "ag-core/tool/cmd/new-gen-db/repository/dao"
    "ag-core/tool/cmd/new-gen-db/repository/model"
    "ag-core/contribute/agdb/gormdb"
    conditonwhere "ag-core/contribute/agdb/conditonwhere"
)

func main() {
    // 创建DAO实例
    tblPurchaseResponseDao := dao.NewTblPurchaseResponseDao(repository, baseDao)

    // 构建WHERE条件
    builder := conditonwhere.NewWhereClauseBuilder()
    builder.AddCondition(conditonwhere.ConditionEq("OrderId", "ORD123456"))
    builder.AddCondition(conditonwhere.ConditionEq("MerchantId", "MERCHANT001"))
    builder.AddCondition(conditonwhere.ConditionEq("TransactionType", "02"))
    builder.AddCondition(conditonwhere.ConditionEq("ResponseCode", "00"))

    // 设置排序
    orders := []gormdb.Order{
        {Column: "InsertTimestamp", Direction: "DESC"},
    }

    // 查询数据
    ctx := context.Background()
    list, err := tblPurchaseResponseDao.FindByWhere(ctx, orders, builder)
    if err != nil {
        fmt.Printf("查询失败: %v\n", err)
        return
    }

    fmt.Printf("查询成功，共 %d 条记录\n", len(list))
    for _, item := range list {
        fmt.Printf("- ID: %d, OrderId: %s, ResponseCode: %s\n", 
            item.Id, item.OrderId, item.ResponseCode)
    }
}
```

### 复杂条件示例

```go
// 示例1: 使用 OR 条件
builder := conditonwhere.NewWhereClauseBuilder()
cond1 := conditonwhere.ConditionEq("OrderId", "ORD123456").Or()
cond2 := conditonwhere.ConditionEq("OrderId", "ORD789012").Or()
builder.AddCondition(cond1)
builder.AddCondition(cond2)

// 示例2: 使用嵌套条件
builder := conditonwhere.NewWhereClauseBuilder()
orGroup := conditonwhere.ConditionOrGroup(
    conditonwhere.ConditionEq("ResponseCode", "00"),
    conditonwhere.ConditionEq("ResponseCode", "01"),
)
builder.AddCondition(conditonwhere.ConditionEq("OrderId", "ORD123456"))
builder.AddCondition(orGroup)

// 示例3: 使用 IN 条件
builder := conditonwhere.NewWhereClauseBuilder()
builder.AddCondition(conditonwhere.ConditionEq("OrderId", "ORD123456"))
builder.AddCondition(conditonwhere.ConditionIn("TransactionType", "02", "03", "20"))

// 示例4: 使用 BETWEEN 条件
builder := conditonwhere.NewWhereClauseBuilder()
builder.AddCondition(conditonwhere.ConditionEq("OrderId", "ORD123456"))
builder.AddCondition(conditonwhere.ConditionBetween("InsertTimestamp", "2024-01-01", "2024-12-31"))
```

### 索引验证

`FindByWhere` 方法内部会自动验证索引使用情况：

- **PRIMARY**: 使用了主键的引导列
- **INDEX_0**: 使用了第一个索引的引导列
- **INDEX_1**: 使用了第二个索引的引导列
- 以此类推...

如果查询条件没有使用任何索引的引导列，方法会返回错误，避免全表扫描。

---

## 注意事项

### 1. 索引使用

`FindByStruct`、`FindByPage` 和 `FindByWhere` 方法都要求必须使用索引。这是为了避免全表扫描，提高查询性能。

### 2. 事务支持

所有方法都支持事务，只需在 context 中绑定事务即可：

```go
// 开启事务
err := repository.Transaction(ctx, func(ctx context.Context) error {
    // 在事务中执行操作
    rowsAffected, err := tblPurchaseResponseDao.InsertOne(ctx, entity)
    if err != nil {
        return err
    }
    return nil
})
```

### 3. 表名动态替换

支持通过 `TableInfo` 动态替换表名，适用于分表场景。

### 4. 错误处理

建议对所有方法返回的错误进行检查和处理：

```go
rowsAffected, err := tblPurchaseResponseDao.InsertOne(ctx, entity)
if err != nil {
    // 记录日志
    log.Errorf("操作失败: %v", err)
    // 返回错误给上层
    return err
}
```

---

## 附录：WhereClauseBuilder 使用指南

### 基本条件

```go
// 等于
conditonwhere.ConditionEq("field", value)

// 不等于
conditonwhere.ConditionNeq("field", value)

// 大于
conditonwhere.ConditionGt("field", value)

// 小于
conditonwhere.ConditionLt("field", value)

// 大于等于
conditonwhere.ConditionGte("field", value)

// 小于等于
conditonwhere.ConditionLte("field", value)

// 模糊匹配
conditonwhere.ConditionLike("field", "value")

// IN 条件
conditonwhere.ConditionIn("field", value1, value2, value3)

// NOT IN 条件
conditonwhere.ConditionNotIn("field", value1, value2)

// BETWEEN 条件
conditonwhere.ConditionBetween("field", min, max)
```

### 逻辑组合

```go
// OR 条件
cond := conditonwhere.ConditionEq("field1", value1).Or()

// AND 条件（默认）
cond := conditonwhere.ConditionEq("field1", value1).And()

// 条件组
andGroup := conditonwhere.ConditionAndGroup(
    conditonwhere.ConditionEq("field1", value1),
    conditonwhere.ConditionEq("field2", value2),
)

orGroup := conditonwhere.ConditionOrGroup(
    conditonwhere.ConditionEq("field1", value1),
    conditonwhere.ConditionEq("field2", value2),
)

// 嵌套条件
parentCond := conditonwhere.ConditionEq("field1", value1)
parentCond.AddChild(conditonwhere.ConditionEq("field2", value2))
```

### 链式调用

```go
builder := conditonwhere.NewWhereClauseBuilder()
builder.
    AddCondition(conditonwhere.ConditionEq("field1", value1)).
    AddCondition(conditonwhere.ConditionEq("field2", value2)).
    AddCondition(conditonwhere.ConditionEq("field3", value3))
```

---

**文档版本**: 1.0  
**最后更新**: 2024-03-15
