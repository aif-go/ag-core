# 使用 decimal.Decimal 与数据库交互

## 重要说明

虽然我们将数据库类型映射改为了 `*decimal.Decimal`，但数据库驱动本身并不直接支持这个类型。不过，`shopspring/decimal` 包已经实现了必要的接口来处理数据库交互：

- `driver.Valuer` 接口：用于将 Go 类型转换为数据库值
- `sql.Scanner` 接口：用于将数据库值扫描到 Go 类型

## GORM 支持

GORM 可以很好地处理 `decimal.Decimal` 类型，因为：

1. `shopspring/decimal` 包已经实现了 `driver.Valuer` 和 `sql.Scanner` 接口
2. GORM 会自动使用这些接口进行类型转换

## 使用示例

### 结构体定义
```go
import (
    "github.com/shopspring/decimal"
    "time"
)

type Product struct {
    ID       uint             `gorm:"primaryKey"`
    Name     string           `gorm:"column:name"`
    Price    decimal.Decimal  `gorm:"column:price;type:DECIMAL(10,2)"`
    Created  time.Time        `gorm:"column:created_at"`
}
```

### 查询操作
```go
// 查询 - GORM 会自动将数据库的 DECIMAL 值转换为 decimal.Decimal
var product Product
db.First(&product, 1)
fmt.Printf("Price: %s\n", product.Price.String()) // 输出: Price: 99.99

// 插入/更新 - GORM 会自动将 decimal.Decimal 转换为数据库格式
newProduct := Product{
    Name:  "Test Product",
    Price: decimal.NewFromFloat(123.45),
}
db.Create(&newProduct)
```

### 直接 SQL 查询
```go
import (
    "database/sql"
    "github.com/shopspring/decimal"
)

// 使用 sql.Scanner 接口直接查询
var price decimal.Decimal
err := db.Raw("SELECT price FROM products WHERE id = ?", id).Scan(&price).Error
if err != nil {
    // 处理错误
}
```

## 注意事项

1. **依赖导入**：确保在使用代码中导入了 `github.com/shopspring/decimal`
2. **数据库迁移**：在数据库中使用适当的 DECIMAL/NUMERIC 类型定义
3. **精度设置**：根据业务需求设置合适的精度，如 `DECIMAL(10,2)` 用于货币
4. **空值处理**：如果字段可能为空，使用 `*decimal.Decimal` 并处理 nil 情况

## GORM 模式生成

当使用我们的代码生成工具时，生成的结构体将包含 `*decimal.Decimal` 字段，GORM 会自动处理与数据库的转换。