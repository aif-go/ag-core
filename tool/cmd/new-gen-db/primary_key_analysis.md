# 主键列表逻辑分析报告

## 问题分析

当前获取主键列表的逻辑存在以下问题：

1. **当前实现**：在 `dao/template.go` 中的 `getPrimaryKey` 函数只返回第一个主键列名：
   ```go
   func getPrimaryKey(tableData *table.TableData) string {
       for _, col := range tableData.Columns {
           if col.IsPrimaryKey {
               return col.Name  // 只返回第一个主键
           }
       }
       return ""
   }
   ```

2. **数据源问题**：YAML文件中确实有 `primary_key` 列表（如 TM_MEDIA_ACT.yaml 中的 `primary_key: - SEQ`），但在解析后，主键信息只存储在 `ColumnData.IsPrimaryKey` 布尔字段中，没有保留原始的主键列表。

## 可行性分析

### 方案1：将yaml解析时的primary_key列表放到tableData中

**优点**：
1. 保持原始数据完整性
2. 支持多主键场景
3. 修改最小，只需在 `TableData` 结构体中添加一个字段

**实现步骤**：
1. 在 `table.TableData` 结构体中添加 `PrimaryKeys []string` 字段
2. 在 `model/parser.go` 的 `ParseYAML` 函数中，解析 `primary_key` 时同时保存到 `PrimaryKeys` 字段
3. 修改 `getPrimaryKey` 函数或创建新函数 `getPrimaryKeys` 返回主键列表

**代码修改点**：
```go
// table/table.go 中修改 TableData 结构体
type TableData struct {
    // 现有字段...
    PrimaryKeys []string // 新增：主键列表
}

// model/parser.go 中修改 ParseYAML 函数
// 处理主键信息
if primaryKey, ok := yamlData["primary_key"].([]interface{}); ok {
    for _, pk := range primaryKey {
        if pkName, ok := pk.(string); ok {
            // 添加到 PrimaryKeys 列表
            tableData.PrimaryKeys = append(tableData.PrimaryKeys, pkName)
            
            // 设置 IsPrimaryKey 标志（保留原有逻辑）
            for i := range columns {
                if columns[i].Name == pkName {
                    columns[i].IsPrimaryKey = true
                    break
                }
            }
        }
    }
}
```

### 方案2：修改现有getPrimaryKey函数返回多个主键

**优点**：
1. 不需要修改数据结构
2. 向后兼容性好

**缺点**：
1. 需要修改所有调用 `getPrimaryKey` 的地方
2. 函数签名变化可能影响其他代码

## 现有调用点分析

`getPrimaryKey` 函数目前只在以下位置被调用：

1. **dao/template.go 第744行**：在生成示例SQL时使用
   ```go
   // 获取主键列名
   primaryKey := getPrimaryKey(tableData)
   
   // 构建排序语句
   sortClause := ""
   if primaryKey != "" {
       sortClause = " ORDER BY " + primaryKey
   }
   ```

   用途：为查询结果添加排序，使用主键作为排序字段。

## 推荐方案

**推荐使用方案1**，原因如下：

1. **数据完整性**：保留原始的主键列表信息，避免信息丢失
2. **扩展性**：支持未来可能的多主键需求
3. **兼容性**：保留现有的 `IsPrimaryKey` 标志，确保向后兼容
4. **最小修改**：只需添加一个字段和一个函数，对现有代码影响最小

## 实现建议

1. 在 `TableData` 中添加 `PrimaryKeys []string` 字段
2. 在 `ParseYAML` 函数中同时填充 `PrimaryKeys` 和 `IsPrimaryKey`
3. 添加 `getPrimaryKeys` 函数返回主键列表
4. 保留现有 `getPrimaryKey` 函数以确保向后兼容

### 兼容性处理

对于现有的 `getPrimaryKey` 调用点，可以有以下两种处理方式：

1. **保持原样**：现有调用点只需要一个主键用于排序，可以继续使用第一个主键
2. **修改为使用主键列表**：将排序改为使用所有主键，如 `ORDER BY pk1, pk2`

推荐采用方式1，因为：
- 现有调用点只有一个（排序场景）
- 使用第一个主键作为排序字段是合理的
- 避免不必要的代码修改

### 修改后的getPrimaryKey函数

```go
// getPrimaryKey 获取主键列名（返回第一个主键，保持向后兼容）
func getPrimaryKey(tableData *table.TableData) string {
    // 优先使用PrimaryKeys字段
    if len(tableData.PrimaryKeys) > 0 {
        return tableData.PrimaryKeys[0]
    }
    
    // 兼容旧逻辑：从Columns中查找
    for _, col := range tableData.Columns {
        if col.IsPrimaryKey {
            return col.Name
        }
    }
    return ""
}

// getPrimaryKeys 获取所有主键列名
func getPrimaryKeys(tableData *table.TableData) []string {
    if len(tableData.PrimaryKeys) > 0 {
        return tableData.PrimaryKeys
    }
    
    // 兼容旧逻辑：从Columns中收集主键
    var primaryKeys []string
    for _, col := range tableData.Columns {
        if col.IsPrimaryKey {
            primaryKeys = append(primaryKeys, col.Name)
        }
    }
    return primaryKeys
}
```

这样既解决了主键列表获取的问题，又保持了代码的向后兼容性。