package gormdb

import "strings"

// 分页结构体
type Page struct {
	PageSize int64
	PageNum  int64
}

// 分页结果结构体
type PageResult struct {
	CurrentPage int64
	TotalCount  int64
	TotalPage   int64
	PageSize    int64
}

// 定义Order结构体，用于分页查询时指定排序列和排序方式
type OrderSort string
const (
	ASC  OrderSort = "ASC"
	DESC OrderSort = "DESC"
)

type Order struct {
	ColName string
	Sort   OrderSort
}

func (o Order) String() string {
	return o.ColName + " " + string(o.Sort)
}

func ToSqlOrder(orderSlice []Order) string {
	var orderStrings []string
	for _, order := range orderSlice {
		orderStrings = append(orderStrings, order.String())
	}
	return strings.Join(orderStrings, ", ")
}

// ==================== OrderBuilder 链式调用 ====================

// OrderBuilder ORDER BY 构建器
type OrderBuilder struct {
	orders []Order
}

// NewOrderBuilder 创建新的 OrderBuilder
func NewOrderBuilder() *OrderBuilder {
	return &OrderBuilder{
		orders: make([]Order, 0),
	}
}

// Asc 添加升序排序
func (b *OrderBuilder) Asc(colName string) *OrderBuilder {
	b.orders = append(b.orders, Order{
		ColName: colName,
		Sort:   ASC,
	})
	return b
}

// Desc 添加降序排序
func (b *OrderBuilder) Desc(colName string) *OrderBuilder {
	b.orders = append(b.orders, Order{
		ColName: colName,
		Sort:   DESC,
	})
	return b
}

// Order 添加自定义排序
func (b *OrderBuilder) Order(colName string, sort OrderSort) *OrderBuilder {
	b.orders = append(b.orders, Order{
		ColName: colName,
		Sort:   sort,
	})
	return b
}

// Orders 批量添加排序
func (b *OrderBuilder) Orders(orders ...Order) *OrderBuilder {
	b.orders = append(b.orders, orders...)
	return b
}

// Build 构建完整的 ORDER BY SQL 语句
func (b *OrderBuilder) Build() string {
	if len(b.orders) == 0 {
		return ""
	}
	return ToSqlOrder(b.orders)
}

// BuildWithoutKeyword 构建排序 SQL 语句（不包含 ORDER BY 关键字）
func (b *OrderBuilder) BuildWithoutKeyword() string {
	if len(b.orders) == 0 {
		return ""
	}
	return ToSqlOrder(b.orders)
}

// ToOrders 转换为 Order 切片
func (b *OrderBuilder) ToOrders() []Order {
	return b.orders
}

// Clear 清空所有排序条件
func (b *OrderBuilder) Clear() *OrderBuilder {
	b.orders = make([]Order, 0)
	return b
}
