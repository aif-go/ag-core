# 类型映射说明

## 精确数值类型处理

对于 `DECIMAL` 和 `NUMERIC` 类型，现在映射到 `*decimal.Decimal` 而不是 `float64`，以避免精度丢失问题。

### 依赖

使用此映射需要添加 `shopspring/decimal` 依赖：

```bash
go get github.com/shopspring/decimal
```

### 在生成的代码中导入

在使用生成的代码时，需要导入：

```go
import "github.com/shopspring/decimal"
```

### 使用示例

```go
// 假设生成的结构体包含 DECIMAL 字段
type MyTable struct {
    Amount *decimal.Decimal `json:"amount"`
}

// 创建新的 decimal 值
value := decimal.NewFromFloat(123.45)
myRecord := MyTable{
    Amount: &value,
}

// 获取 decimal 值
if myRecord.Amount != nil {
    fmt.Printf("Amount: %s\n", myRecord.Amount.String())
}
```

## 类型映射表

| 数据库类型 | Go 类型 |
|------------|---------|
| CHAR, VARCHAR, TEXT 等 | string |
| TINYINT | int8 |
| SMALLINT | int16 |
| INT, INTEGER | int |
| BIGINT | int64 |
| FLOAT | float32 |
| DOUBLE | float64 |
| DECIMAL, NUMERIC | *decimal.Decimal |
| DATE, DATETIME, TIMESTAMP | time.Time |
| BINARY, BLOB 等 | []byte |
| BOOLEAN, BOOL | bool |