# 基于 Builder 模式的动态 WHERE 条件方案设计（优化版）

## 一、问题分析

### 1.1 零值判断的局限性

**问题场景**：

```go
// 场景 1: 用户想查询 ID=0 的记录
arg := &TblPurchaseResponseFindAaaArg{Id: 0}
// 零值判断：Id == 0，认为条件应该被剔除 ❌ 错误！

// 场景 2: 用户想查询 Name="" 的记录
arg := &TblPurchaseResponseFindAaaArg{Name: ""}
// 零值判断：Name == ""，认为条件应该被剔除 ❌ 错误！

// 场景 3: 用户想查询 Age=0 的记录
arg := &TblPurchaseResponseFindAaaArg{Age: 0}
// 零值判断：Age == 0，认为条件应该被剔除 ❌ 错误！
```

**根本原因**：零值判断无法区分"用户未设置"和"用户设置了零值"。

### 1.2 Builder 模式的优势

| 对比项 | 零值判断 | Builder 模式 |
|--------|---------|------------|
| 明确性 | ❌ 无法区分未设置和零值 | ✅ 明确标记哪些字段被设置 |
| 灵活性 | ❌ 无法查询零值 | ✅ 可以查询零值 |
| 语义清晰 | ❌ 依赖隐式规则 | ✅ 显式标记，语义清晰 |
| 类型安全 | ⚠️ 需要反射判断 | ✅ 编译期检查 |

---

## 二、设计方案（优化版）

### 2.1 核心思路

**为每个查询参数结构体生成一个 Builder**：

1. **Builder 包含**：
   - 原始参数结构体
   - 字段标记（FieldMask，标记哪些字段被设置了）

2. **Builder 提供**：
   - 链式设置方法（`WithId()`, `WithName()` 等）
   - `Build()` 方法生成参数对象（包含 FieldMask）

3. **参数结构体包含**：
   - 原始字段
   - `FieldMask` 字段（标记哪些字段被设置了）

4. **运行时**：
   - DAO 层直接从 args 中获取 FieldMask
   - 根据 FieldMask 判断是否保留条件

### 2.2 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    代码生成器                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ 1. 生成参数结构体（Arg，包含 FieldMask）              │  │
│  │ 2. 生成 Builder 结构体                                 │  │
│  │ 3. 生成 Builder 的 WithXxx() 方法                      │  │
│  │ 4. 生成 WhereClause 变量                              │  │
│  │ 5. 生成 DAO 方法                                      │  │
│  └───────────────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                    Model 层                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ // TblPurchaseResponseFindAaaArg 参数结构体           │  │
│  │ type TblPurchaseResponseFindAaaArg struct {            │  │
│  │   db.Page                                            │  │
│  │   Id int64                                           │  │
│  │   Name string                                         │  │
│  │   Age int64                                          │  │
│  │   FieldMask *conditonwhere.FieldMask // 新增！        │  │
│  │ }                                                     │  │
│  │                                                       │  │
│  │ // TblPurchaseResponseFindAaaArgBuilder Builder         │  │
│  │ type TblPurchaseResponseFindAaaArgBuilder struct {     │  │
│  │   arg *TblPurchaseResponseFindAaaArg                  │  │
│  │ }                                                     │  │
│  │                                                       │  │
│  │ func (b *Builder) WithId(id int64) *Builder { ... }   │  │
│  │ func (b *Builder) WithName(name string) *Builder { ... }│  │
│  │ func (b *Builder) Build() *Arg { ... }               │  │
│  └───────────────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                    DAO 层                                   │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ func (dao *TblPurchaseResponseDao) doFindAaa(...) { │  │
│  │   // 1. 类型断言，获取参数                           │  │
│  │   queryArgs, ok := args.(*model.TblPurchaseResponseFindAaaArg)│  │
│  │                                                      │  │
│  │   // 2. 获取基础 WHERE 子句                            │  │
│  │   baseClause := model.FindAaaWhereClause              │  │
│  │                                                      │  │
│  │   // 3. 根据 FieldMask 过滤条件（直接从 args 获取）  │  │
│  │   filteredClause := FilterWhereClauseByFieldMask(      │  │
│  │       baseClause,                                      │  │
│  │       queryArgs.FieldMask,  // 直接使用！              │  │
│  │   )                                                   │  │
│  │                                                      │  │
│  │   // 4. 转换为 SQL WHERE 条件                         │  │
│  │   whereSQL := BuildWhereSQL(filteredClause)           │  │
│  │                                                      │  │
│  │   // 5. 拼接 SQL 并执行                              │  │
│  │   sql := fmt.Sprintf("SELECT %s FROM %s %s", ...)    │  │
│  │ }                                                     │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、详细设计

### 3.1 FieldMask 设计

```go
package conditonwhere

// FieldMask 字段标记，用于标记哪些字段被设置了
type FieldMask struct {
    fields map[string]bool
}

// NewFieldMask 创建 FieldMask
func NewFieldMask() *FieldMask {
    return &FieldMask{
        fields: make(map[string]bool),
    }
}

// Set 设置字段标记
func (m *FieldMask) Set(field string) {
    m.fields[field] = true
}

// IsSet 检查字段是否被设置
func (m *FieldMask) IsSet(field string) bool {
    return m.fields[field]
}

// GetFields 获取所有被设置的字段
func (m *FieldMask) GetFields() []string {
    fields := make([]string, 0, len(m.fields))
    for field := range m.fields {
        fields = append(fields, field)
    }
    return fields
}

// Merge 合并另一个 FieldMask
func (m *FieldMask) Merge(other *FieldMask) {
    if other == nil {
        return
    }
    for field := range other.fields {
        m.fields[field] = true
    }
}
```

### 3.2 参数结构体设计（包含 FieldMask）

```go
// TblPurchaseResponseFindAaaArg FindAaa 查询参数
type TblPurchaseResponseFindAaaArg struct {
    db.Page
    Id       int64
    Name      string
    Age       int64
    FieldMask *conditonwhere.FieldMask // 新增：字段标记
}

// ConvertToMap 将参数转换为map
func (tblpurchaseresponseFindAaaArg *TblPurchaseResponseFindAaaArg) ConvertToMap() map[string]interface{} {
    if tblpurchaseresponseFindAaaArg == nil {
        return nil
    }
    return map[string]interface{}{
        "Page": tblpurchaseresponseFindAaaArg.Page,
        "Id":   tblpurchaseresponseFindAaaArg.Id,
        "Name":  tblpurchaseresponseFindAaaArg.Name,
        "Age":   tblpurchaseresponseFindAaaArg.Age,
    }
}

// TblPurchaseResponseFindDuplicatePurchaseArg FindDuplicatePurchase 查询参数
type TblPurchaseResponseFindDuplicatePurchaseArg struct {
    OrderId        string
    MerchantId     string
    TransactionType string
    FieldMask      *conditonwhere.FieldMask // 新增：字段标记
}

// ConvertToMap 将参数转换为map
func (tblpurchaseresponseFindDuplicatePurchaseArg *TblPurchaseResponseFindDuplicatePurchaseArg) ConvertToMap() map[string]interface{} {
    if tblpurchaseresponseFindDuplicatePurchaseArg == nil {
        return nil
    }
    return map[string]interface{}{
        "OrderId":        tblpurchaseresponseFindDuplicatePurchaseArg.OrderId,
        "MerchantId":     tblpurchaseresponseFindDuplicatePurchaseArg.MerchantId,
        "TransactionType": tblpurchaseresponseFindDuplicatePurchaseArg.TransactionType,
    }
}
```

### 3.3 Builder 设计

```go
// TblPurchaseResponseFindAaaArgBuilder FindAaa 查询参数 Builder
type TblPurchaseResponseFindAaaArgBuilder struct {
    arg *TblPurchaseResponseFindAaaArg
}

// NewTblPurchaseResponseFindAaaArgBuilder 创建 Builder
func NewTblPurchaseResponseFindAaaArgBuilder() *TblPurchaseResponseFindAaaArgBuilder {
    return &TblPurchaseResponseFindAaaArgBuilder{
        arg: &TblPurchaseResponseFindAaaArg{
            Page:      db.Page{PageNum: 1, PageSize: 10}, // 默认分页参数
            FieldMask: conditonwhere.NewFieldMask(),
        },
    }
}

// WithId 设置 Id
func (b *TblPurchaseResponseFindAaaArgBuilder) WithId(id int64) *TblPurchaseResponseFindAaaArgBuilder {
    b.arg.Id = id
    b.arg.FieldMask.Set("Id")
    return b
}

// WithName 设置 Name
func (b *TblPurchaseResponseFindAaaArgBuilder) WithName(name string) *TblPurchaseResponseFindAaaArgBuilder {
    b.arg.Name = name
    b.arg.FieldMask.Set("Name")
    return b
}

// WithAge 设置 Age
func (b *TblPurchaseResponseFindAaaArgBuilder) WithAge(age int64) *TblPurchaseResponseFindAaaArgBuilder {
    b.arg.Age = age
    b.arg.FieldMask.Set("Age")
    return b
}

// WithPage 设置分页参数
func (b *TblPurchaseResponseFindAaaArgBuilder) WithPage(pageNum, pageSize int64) *TblPurchaseResponseFindAaaArgBuilder {
    b.arg.PageNum = pageNum
    b.arg.PageSize = pageSize
    b.arg.FieldMask.Set("Page")
    return b
}

// Build 构建参数对象
func (b *TblPurchaseResponseFindAaaArgBuilder) Build() *TblPurchaseResponseFindAaaArg {
    return b.arg
}

// TblPurchaseResponseFindDuplicatePurchaseArgBuilder FindDuplicatePurchase 查询参数 Builder
type TblPurchaseResponseFindDuplicatePurchaseArgBuilder struct {
    arg *TblPurchaseResponseFindDuplicatePurchaseArg
}

// NewTblPurchaseResponseFindDuplicatePurchaseArgBuilder 创建 Builder
func NewTblPurchaseResponseFindDuplicatePurchaseArgBuilder() *TblPurchaseResponseFindDuplicatePurchaseArgBuilder {
    return &TblPurchaseResponseFindDuplicatePurchaseArgBuilder{
        arg: &TblPurchaseResponseFindDuplicatePurchaseArg{
            FieldMask: conditonwhere.NewFieldMask(),
        },
    }
}

// WithOrderId 设置 OrderId
func (b *TblPurchaseResponseFindDuplicatePurchaseArgBuilder) WithOrderId(orderId string) *TblPurchaseResponseFindDuplicatePurchaseArgBuilder {
    b.arg.OrderId = orderId
    b.arg.FieldMask.Set("OrderId")
    return b
}

// WithMerchantId 设置 MerchantId
func (b *TblPurchaseResponseFindDuplicatePurchaseArgBuilder) WithMerchantId(merchantId string) *TblPurchaseResponseFindDuplicatePurchaseArgBuilder {
    b.arg.MerchantId = merchantId
    b.arg.FieldMask.Set("MerchantId")
    return b
}

// WithTransactionType 设置 TransactionType
func (b *TblPurchaseResponseFindDuplicatePurchaseArgBuilder) WithTransactionType(transactionType string) *TblPurchaseResponseFindDuplicatePurchaseArgBuilder {
    b.arg.TransactionType = transactionType
    b.arg.FieldMask.Set("TransactionType")
    return b
}

// Build 构建参数对象
func (b *TblPurchaseResponseFindDuplicatePurchaseArgBuilder) Build() *TblPurchaseResponseFindDuplicatePurchaseArg {
    return b.arg
}
```

### 3.4 FilterWhereClauseByFieldMask 实现

```go
package conditonwhere

// FilterWhereClauseByFieldMask 根据 FieldMask 过滤 WHERE 条件
// clause: 原始 WHERE 子句
// fieldMask: 字段标记
// 返回: 过滤后的 WHERE 子句
func FilterWhereClauseByFieldMask(clause *WhereClause, fieldMask *FieldMask) *WhereClause {
    if clause == nil {
        return nil
    }

    filtered := make([]*Condition, 0, len(clause.Conditions))
    for _, cond := range clause.Conditions {
        // 字面量条件始终保留
        if cond.IsLiteral {
            filtered = append(filtered, cond)
            continue
        }

        // 检查字段是否被设置
        if fieldMask != nil && fieldMask.IsSet(cond.ParamName) {
            filtered = append(filtered, cond)
        }
    }

    return &WhereClause{
        Operator:   clause.Operator,
        Conditions: filtered,
    }
}
```

### 3.5 BuildWhereSQL 实现

```go
package conditonwhere

import (
    "fmt"
    "strings"
)

// BuildWhereSQL 构建 SQL WHERE 子句
// clause: WHERE 子句结构
// 返回: SQL WHERE 子句字符串
func BuildWhereSQL(clause *WhereClause) string {
    if clause == nil || len(clause.Conditions) == 0 {
        return "WHERE 1=1"
    }

    if len(clause.Conditions) == 1 {
        return fmt.Sprintf("WHERE (%s)", clause.Conditions[0].Expr)
    }

    var parts []string
    for _, cond := range clause.Conditions {
        parts = append(parts, cond.Expr)
    }

    return fmt.Sprintf("WHERE (%s)", strings.Join(parts, " "+clause.Operator+" "))
}
```

### 3.6 WhereClause 变量

```go
// FindAaaWhereClause FindAaa 的 WHERE 子句定义
var FindAaaWhereClause = conditonwhere.NewWhereClause(
    "AND",
    conditonwhere.NewCondition("ID = @Id", "Id"),
    conditonwhere.NewCondition("Name = @Name", "Name"),
    conditonwhere.NewCondition("Age = @Age", "Age"),
)

// FindDuplicatePurchaseWhereClause FindDuplicatePurchase 的 WHERE 子句定义
var FindDuplicatePurchaseWhereClause = conditonwhere.NewWhereClause(
    "AND",
    conditonwhere.NewCondition("ORDER_ID = @OrderId", "OrderId"),
    conditonwhere.NewCondition("MERCHANT_ID = @MerchantId", "MerchantId"),
    conditonwhere.NewCondition("TRANSACTION_TYPE = @TransactionType", "TransactionType"),
    conditonwhere.NewLiteralCondition("RESPONSE_CODE = '00'"), // 字面量条件
)
```

---

## 四、DAO 方法实现

### 4.1 doFindAaa 实现

```go
// doFindAaa 执行FindAaa查询（分页）
func (dao *TblPurchaseResponseDao) doFindAaa(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) (*model.TblPurchaseResponseFindAaaPageRes, error) {
    // 1. 类型断言，获取参数
    queryArgs, ok := args.(*model.TblPurchaseResponseFindAaaArg)
    if !ok {
        return nil, errors.New("doFindAaa args type not match")
    }

    // 2. 获取基础 WHERE 子句
    baseClause := model.FindAaaWhereClause

    // 3. 根据 FieldMask 过滤条件（直接从 args 获取）
    filteredClause := conditonwhere.FilterWhereClauseByFieldMask(baseClause, queryArgs.FieldMask)

    // 4. 构建 SQL WHERE 条件
    whereSQL := conditonwhere.BuildWhereSQL(filteredClause)

    // 5. 构建 SELECT 语句
    selectFields := "NETWORK_ID,RESULT_CODE"
    tableName := dao.getApplyInfo(ctx).TableName
    if tableName == "" {
        tableName = (&model.TblPurchaseResponse{}).TableName()
    }

    // 6. 构建 COUNT SQL
    countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", tableName, whereSQL)

    // 7. 构建 DATA SQL
    dataSQL := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY ID LIMIT ?, ?", selectFields, tableName, whereSQL)

    // 8. 执行 COUNT
    argsMap := queryArgs.ConvertToMap()
    var totalCount int64
    result := dao.DB(ctx).Raw(countSQL, argsMap).Scan(&totalCount)
    if result.Error != nil {
        return nil, result.Error
    }

    // 9. 计算分页
    startRecord, endRecord, totalPage := gormdb.CalcPageStartRecord(queryArgs.PageNum, queryArgs.PageSize, totalCount, dao.DbType)

    // 10. 执行 DATA 查询
    var list []*model.TblPurchaseResponseFindAaaRes
    resultlist := dao.DB(ctx).Raw(dataSQL, append([]interface{}{startRecord, endRecord}, getValuesByFieldMask(argsMap, queryArgs.FieldMask)...)).Find(&list)
    if resultlist.Error != nil {
        return nil, resultlist.Error
    }

    return &model.TblPurchaseResponseFindAaaPageRes{
        PageResult: gormdb.PageResult{
            CurrentPage: queryArgs.PageNum,
            PageSize:    queryArgs.PageSize,
            TotalCount:  totalCount,
            TotalPage:   totalPage,
        },
        ResultList: list,
    }, nil
}

// getValuesByFieldMask 根据 FieldMask 获取参数值
func getValuesByFieldMask(argsMap map[string]interface{}, fieldMask *conditonwhere.FieldMask) []interface{} {
    if fieldMask == nil {
        return nil
    }
    
    var values []interface{}
    for _, field := range fieldMask.GetFields() {
        if value, ok := argsMap[field]; ok {
            values = append(values, value)
        }
    }
    return values
}
```

### 4.2 doFindDuplicatePurchase 实现

```go
// doFindDuplicatePurchase 执行FindDuplicatePurchase查询（非分页）
func (dao *TblPurchaseResponseDao) doFindDuplicatePurchase(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) ([]*model.TblPurchaseResponse, error) {
    // 1. 类型断言，获取参数
    queryArgs, ok := args.(*model.TblPurchaseResponseFindDuplicatePurchaseArg)
    if !ok {
        return nil, errors.New("doFindDuplicatePurchase args type not match")
    }

    // 2. 获取基础 WHERE 子句
    baseClause := model.FindDuplicatePurchaseWhereClause

    // 3. 根据 FieldMask 过滤条件（直接从 args 获取）
    filteredClause := conditonwhere.FilterWhereClauseByFieldMask(baseClause, queryArgs.FieldMask)

    // 4. 构建 SQL WHERE 条件
    whereSQL := conditonwhere.BuildWhereSQL(filteredClause)

    // 5. 构建 SELECT 语句
    selectFields := "*"
    tableName := dao.getApplyInfo(ctx).TableName
    if tableName == "" {
        tableName = (&model.TblPurchaseResponse{}).TableName()
    }

    // 6. 构建 SQL
    sql := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY ID", selectFields, tableName, whereSQL)

    // 7. 执行查询
    argsMap := queryArgs.ConvertToMap()
    var list []*model.TblPurchaseResponse
    result := dao.DB(ctx).Raw(sql, getValuesByFieldMask(argsMap, queryArgs.FieldMask)).Find(&list)
    if result.Error != nil {
        return nil, result.Error
    }

    return list, nil
}
```

---

## 五、使用示例

### 5.1 基本使用

```go
// 场景 1: 查询 ID=1 的记录
arg := model.NewTblPurchaseResponseFindAaaArgBuilder().
    WithId(1).
    Build()
result, err := dao.FindByCustomerRule(ctx, &gormdb.NameingSqlArgInfo{SqlName: "FindAaa"}, arg)
// 生成的 SQL: WHERE (ID = @Id)

// 场景 2: 查询 ID=0 的记录（明确设置零值）
arg := model.NewTblPurchaseResponseFindAaaArgBuilder().
    WithId(0).
    Build()
result, err := dao.FindByCustomerRule(ctx, &gormdb.NameingSqlArgInfo{SqlName: "FindAaa"}, arg)
// 生成的 SQL: WHERE (ID = @Id) ✅ 正确！零值也被使用

// 场景 3: 查询 Name="" 的记录
arg := model.NewTblPurchaseResponseFindAaaArgBuilder().
    WithName("").
    Build()
result, err := dao.FindByCustomerRule(ctx, &gormdb.NameingSqlArgInfo{SqlName: "FindAaa"}, arg)
// 生成的 SQL: WHERE (Name = @Name) ✅ 正确！空字符串也被使用

// 场景 4: 查询 Age=0 的记录
arg := model.NewTblPurchaseResponseFindAaaArgBuilder().
    WithAge(0).
    Build()
result, err := dao.FindByCustomerRule(ctx, &gormdb.NameingSqlArgInfo{SqlName: "FindAaa"}, arg)
// 生成的 SQL: WHERE (Age = @Age) ✅ 正确！零值也被使用

// 场景 5: 组合条件
arg := model.NewTblPurchaseResponseFindAaaArgBuilder().
    WithId(1).
    WithName("test").
    WithPage(1, 10).
    Build()
result, err := dao.FindByCustomerRule(ctx, &gormdb.NameingSqlArgInfo{SqlName: "FindAaa"}, arg)
// 生成的 SQL: WHERE (ID = @Id AND Name = @Name)

// 场景 6: 不设置任何条件
arg := model.NewTblPurchaseResponseFindAaaArgBuilder().
    Build()
result, err := dao.FindByCustomerRule(ctx, &gormdb.NameingSqlArgInfo{SqlName: "FindAaa"}, arg)
// 生成的 SQL: WHERE 1=1
```

### 5.2 实际应用示例

```go
// Service 层
func (s *PurchaseService) FindOrders(ctx context.Context, req *FindOrderRequest) ([]*model.TblPurchaseResponse, error) {
    // 使用 Builder 构建查询参数
    builder := model.NewTblPurchaseResponseFindDuplicatePurchaseArgBuilder()

    // 根据请求参数设置条件
    if req.OrderId != "" {
        builder.WithOrderId(req.OrderId)
    }
    if req.MerchantId != "" {
        builder.WithMerchantId(req.MerchantId)
    }
    if req.TransactionType != "" {
        builder.WithTransactionType(req.TransactionType)
    }

    // 构建参数（包含 FieldMask）
    arg := builder.Build()

    // 执行查询（DAO 层直接使用 arg.FieldMask）
    result, err := s.dao.FindByCustomerRule(ctx, &gormdb.NameingSqlArgInfo{SqlName: "FindDuplicatePurchase"}, arg)
    return result.([]*model.TblPurchaseResponse), err
}
```

---

## 六、代码生成器设计

### 6.1 生成参数结构体（包含 FieldMask）

```go
// generateArgStruct 生成参数结构体（包含 FieldMask）
func generateArgStruct(queryName string, argType string, fields []string) string {
    var builder strings.Builder

    // 结构体定义
    builder.WriteString(fmt.Sprintf("// %s %s 查询参数\n", argType, queryName))
    builder.WriteString(fmt.Sprintf("type %s struct {\n", argType))
    
    for _, field := range fields {
        builder.WriteString(fmt.Sprintf("    %s %s\n", field, getFieldType(field)))
    }
    
    // 新增：FieldMask 字段
    builder.WriteString("    FieldMask *conditonwhere.FieldMask\n")
    builder.WriteString("}\n\n")

    // ConvertToMap 方法
    builder.WriteString(fmt.Sprintf("// ConvertToMap 将参数转换为map\n"))
    builder.WriteString(fmt.Sprintf("func (%s *%s) ConvertToMap() map[string]interface{} {\n", strings.ToLower(argType), argType))
    builder.WriteString(fmt.Sprintf("    if %s == nil {\n", strings.ToLower(argType)))
    builder.WriteString("        return nil\n")
    builder.WriteString("    }\n")
    builder.WriteString("    return map[string]interface{}{\n")
    
    for _, field := range fields {
        builder.WriteString(fmt.Sprintf("        \"%s\": %s.%s,\n", field, strings.ToLower(argType), field))
    }
    
    builder.WriteString("    }\n")
    builder.WriteString("}\n")

    return builder.String()
}
```

### 6.2 生成 Builder

```go
// generateArgBuilder 生成参数 Builder
func generateArgBuilder(queryName string, argType string, fields []string) string {
    var builder strings.Builder

    // Builder 结构体
    builder.WriteString(fmt.Sprintf("// %sArgBuilder %s 查询参数 Builder\n", queryName, queryName))
    builder.WriteString(fmt.Sprintf("type %sArgBuilder struct {\n", queryName))
    builder.WriteString(fmt.Sprintf("    arg *%s\n", argType))
    builder.WriteString("}\n\n")

    // NewBuilder 方法
    builder.WriteString(fmt.Sprintf("// New%sArgBuilder 创建 Builder\n", queryName))
    builder.WriteString(fmt.Sprintf("func New%sArgBuilder() *%sArgBuilder {\n", queryName, queryName))
    builder.WriteString(fmt.Sprintf("    return &%sArgBuilder{\n", queryName))
    builder.WriteString(fmt.Sprintf("        arg: &%s{\n", argType))
    
    // 检查是否有 Page 字段
    if contains(fields, "Page") {
        builder.WriteString("            Page:      db.Page{PageNum: 1, PageSize: 10},\n")
    }
    
    builder.WriteString("            FieldMask: conditonwhere.NewFieldMask(),\n")
    builder.WriteString("        },\n")
    builder.WriteString("    }\n")
    builder.WriteString("}\n\n")

    // WithXxx 方法
    for _, field := range fields {
        // 跳过 FieldMask 字段
        if field == "FieldMask" {
            continue
        }

        // 生成 WithXxx 方法
        builder.WriteString(fmt.Sprintf("// With%s 设置 %s\n", field, field))
        builder.WriteString(fmt.Sprintf("func (b *%sArgBuilder) With%s(%s %s) *%sArgBuilder {\n", 
            queryName, field, strings.ToLower(field), getFieldType(field), queryName))
        builder.WriteString(fmt.Sprintf("    b.arg.%s = %s\n", field, strings.ToLower(field)))
        builder.WriteString(fmt.Sprintf("    b.arg.FieldMask.Set(\"%s\")\n", field))
        builder.WriteString(fmt.Sprintf("    return b\n"))
        builder.WriteString("}\n\n")
    }

    // Build 方法
    builder.WriteString("// Build 构建参数对象\n")
    builder.WriteString(fmt.Sprintf("func (b *%sArgBuilder) Build() *%s {\n", queryName, argType))
    builder.WriteString("    return b.arg\n")
    builder.WriteString("}\n")

    return builder.String()
}
```

### 6.3 生成 DAO 方法

```go
// generateDynamicWhereMethodWithBuilder 生成使用 Builder 的动态 WHERE 方法
func generateDynamicWhereMethodWithBuilder(queryName string, queryData *QueryData) string {
    var builder strings.Builder

    // 方法签名
    builder.WriteString(fmt.Sprintf("// do%s 执行%s查询\n", queryName, queryName))
    builder.WriteString(fmt.Sprintf("func (dao *TblPurchaseResponseDao) do%s(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) (*model.%sPageRes, error) {\n", queryName, queryName))

    // 类型断言
    builder.WriteString(fmt.Sprintf("    queryArgs, ok := args.(*model.%sArg)\n", queryName))
    builder.WriteString("    if !ok {\n")
    builder.WriteString(fmt.Sprintf("        return nil, errors.New(\"do%s args type not match\")\n", queryName))
    builder.WriteString("    }\n\n")

    // 获取基础 WHERE 子句
    builder.WriteString("    // 获取基础 WHERE 子句\n")
    builder.WriteString(fmt.Sprintf("    baseClause := model.%sWhereClause\n\n", queryName))

    // 根据 FieldMask 过滤条件（直接从 args 获取）
    builder.WriteString("    // 根据 FieldMask 过滤条件\n")
    builder.WriteString("    filteredClause := conditonwhere.FilterWhereClauseByFieldMask(baseClause, queryArgs.FieldMask)\n\n")

    // 构建 SQL WHERE 条件
    builder.WriteString("    // 构建 SQL WHERE 条件\n")
    builder.WriteString("    whereSQL := conditonwhere.BuildWhereSQL(filteredClause)\n\n")

    // 构建 SELECT 语句
    builder.WriteString("    // 构建 SELECT 语句\n")
    builder.WriteString(fmt.Sprintf("    selectFields := \"%s\"\n", queryData.SelectFields))
    builder.WriteString("    tableName := dao.getApplyInfo(ctx).TableName\n")
    builder.WriteString("    if tableName == \"\" {\n")
    builder.WriteString("        tableName = (&model.TblPurchaseResponse{}).TableName()\n")
    builder.WriteString("    }\n\n")

    // 构建 COUNT SQL
    builder.WriteString("    // 构建 COUNT SQL\n")
    builder.WriteString("    countSQL := fmt.Sprintf(\"SELECT COUNT(*) FROM %s %s\", tableName, whereSQL)\n\n")

    // 构建 DATA SQL
    builder.WriteString("    // 构建 DATA SQL\n")
    builder.WriteString("    dataSQL := fmt.Sprintf(\"SELECT %s FROM %s %s ORDER BY ID LIMIT ?\", selectFields, tableName, whereSQL)\n\n")

    // 执行 COUNT
    builder.WriteString("    // 执行 COUNT\n")
    builder.WriteString("    argsMap := queryArgs.ConvertToMap()\n")
    builder.WriteString("    var totalCount int64\n")
    builder.WriteString("    result := dao.DB(ctx).Raw(countSQL, argsMap).Scan(&totalCount)\n")
    builder.WriteString("    if result.Error != nil {\n")
    builder.WriteString("        return nil, result.Error\n")
    builder.WriteString("    }\n\n")

    // 计算分页
    builder.WriteString("    // 计算分页\n")
    builder.WriteString("    startRecord, endRecord, totalPage := gormdb.CalcPageStartRecord(queryArgs.PageNum, queryArgs.PageSize, totalCount, dao.DbType)\n\n")

    // 执行 DATA 查询
    builder.WriteString("    // 执行 DATA 查询\n")
    builder.WriteString(fmt.Sprintf("    var list []*model.%sRes\n", queryName))
    builder.WriteString("    resultlist := dao.DB(ctx).Raw(dataSQL, append([]interface{}{startRecord, endRecord}, getValuesByFieldMask(argsMap, queryArgs.FieldMask)...)).Find(&list)\n")
    builder.WriteString("    if resultlist.Error != nil {\n")
    builder.WriteString("        return nil, resultlist.Error\n")
    builder.WriteString("    }\n\n")

    // 返回结果
    builder.WriteString("    return &model." + queryName + "PageRes{\n")
    builder.WriteString("        PageResult: gormdb.PageResult{\n")
    builder.WriteString("            CurrentPage: queryArgs.PageNum,\n")
    builder.WriteString("            PageSize:    queryArgs.PageSize,\n")
    builder.WriteString("            TotalCount:  totalCount,\n")
    builder.WriteString("            TotalPage:   totalPage,\n")
    builder.WriteString("        },\n")
    builder.WriteString("        ResultList: list,\n")
    builder.WriteString("    }, nil\n")
    builder.WriteString("}\n")

    return builder.String()
}
```

---

## 七、优势分析

### 7.1 与零值判断对比

| 对比项 | 零值判断 | Builder 模式（优化版）|
|--------|---------|---------------------|
| 明确性 | ❌ 无法区分未设置和零值 | ✅ 明确标记哪些字段被设置 |
| 灵活性 | ❌ 无法查询零值 | ✅ 可以查询零值 |
| 语义清晰 | ❌ 依赖隐式规则 | ✅ 显式标记，语义清晰 |
| 类型安全 | ⚠️ 需要反射判断 | ✅ 编译期检查 |
| 代码可读性 | ⚠️ 需要理解零值规则 | ✅ 链式调用，一目了然 |
| FieldMask 传递 | ❌ 需要通过 Context 传递 | ✅ 直接从 args 获取 |

### 7.2 核心优势

1. **避免零值误判**：明确区分"未设置"和"设置了零值"
2. **类型安全**：编译期检查，避免运行时错误
3. **链式调用**：Builder 模式，代码简洁易读
4. **灵活性**：可以查询任意值，包括零值
5. **可维护性**：代码结构清晰，易于维护
6. **简化传递**：FieldMask 直接在 args 中，无需通过 Context 传递

---

## 八、实现步骤

### 8.1 第一阶段：核心组件

1. 实现 `FieldMask` 结构
2. 实现 `FilterWhereClauseByFieldMask` 方法
3. 实现 `BuildWhereSQL` 方法
4. 编写单元测试

### 8.2 第二阶段：代码生成器

1. 修改 YAML 解析逻辑
2. 生成参数结构体（包含 FieldMask）
3. 生成 Builder 结构体
4. 生成 Builder 的 WithXxx() 方法
5. 修改 DAO 方法生成逻辑
6. 测试代码生成

### 8.3 第三阶段：集成和测试

1. 集成到现有 DAO
2. 编写集成测试
3. 性能测试
4. 文档编写

---

## 九、注意事项

### 9.1 FieldMask 使用

- FieldMask 必须通过 Builder 的 WithXxx() 方法设置
- 直接构造参数对象时，需要手动初始化 FieldMask

### 9.2 向后兼容

为了保持向后兼容，可以同时支持零值判断和 Builder 模式，让用户选择使用哪种方式。

### 9.3 参数传递

DAO 层直接从 args 中获取 FieldMask，无需通过 Context 传递，简化了调用链。

---

## 十、总结

本方案设计了一个基于 Builder 模式的动态 WHERE 条件机制（优化版），具有以下特点：

1. **避免零值误判**：明确区分"未设置"和"设置了零值"
2. **类型安全**：编译期检查，避免运行时错误
3. **链式调用**：Builder 模式，代码简洁易读
4. **灵活性**：可以查询任意值，包括零值
5. **可维护性**：代码结构清晰，易于维护
6. **简化传递**：FieldMask 直接在 args 中，无需通过 Context 传递

该方案能够有效解决零值误判问题，提供更友好的 API，同时保持代码的简洁性和可维护性。
