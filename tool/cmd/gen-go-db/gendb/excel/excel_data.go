package excel

import "ag-core/tool/cmd/gen-go-db/gendb/render"

type ExcelData struct {
	DbType string
	// dbname 用来做dao模块的model
	// SchemaName string
	ModelName string
	// 表名
	TableName string
	// 表元素的列数据集合
	ColumnList []*render.ColumnData
	// 普通索引集合
	GeneralIndexList []*render.IndexData
	// 约束索引集合
	UniqueIndexList []*render.IndexData
	// 主键
	PrimaryKeyList []string
	// 自定义sql集合 key为后续要生成的方法名
	NamingSqlList []*render.NamingSqlData
	// 用来支持多数据类型的场景
	NamingsqlMap map[string]*render.NamingSqlData
	Encode       string
	Engine       string
	Sort         string
}

type ExcelAllData struct {
	ExcelDataList []*ExcelData
}
