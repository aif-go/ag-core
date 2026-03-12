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
