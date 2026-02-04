package gormdb

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
