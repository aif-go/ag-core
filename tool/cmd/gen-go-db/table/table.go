package table

import "github.com/aif-go/ag-core/contribute/agdb/conditonwhere"

// TableData 存储解析后的表数据，用于模板渲染
type TableData struct {
	ModuleName        string
	TableName         string
	StructName        string
	Columns           []ColumnData
	PrimaryKeys       []string   // 主键列表
	Indexes           []IndexData
	SelfQueries       []QueryData
	ModelTemplateData *ModelTemplateData
	HasPage           bool
	HasSelfQuery      bool // 是否有自定义查询
	AllowUpdateCols   []string // 支持更新的列名切片
}

type ModelTemplateData struct {
	ImportPackages []string
}

// ColumnData 列数据
type ColumnData struct {
	Name            string
	Type            string
	GoType          string
	GormTag         string
	JsonTag         string
	IsPrimaryKey    bool
	IsAutoCreate    bool
	IsAutoUpdate    bool
	IsJavaVersion   bool
	SupportUpdate   bool           // 是否支持更新
	IndexPriorities map[string]int // 索引优先级映射
}

// IndexData 索引数据
type IndexData struct {
	Name     string
	Columns  []string
	Priority int
	IsUnique bool // 是否为唯一索引
}

// WhereCondition where条件，支持嵌套结构
// type WhereCondition struct {
// 	Operator   string           `json:"operator"`
// 	Conditions []WhereCondition `json:"conditions"`
// 	Expr       string           `json:"expr"`
// }

// WhereColField where条件列字段信息
type WhereColField struct {
	ColName   string // 列名
	FieldName string // 字段名
	IsSlice   bool   // 是否为切片类型
	Operator  string // 操作符
	GoType string // 类型
}

// QueryData 查询数据
type QueryData struct {
	Name           string          `json:"name"`
	SelectFields   string          `json:"selectFields"`
	Fields         []string        `json:"fields"`
	HasPage        bool            `json:"hasPage"`
	WhereFields    []string        `json:"whereFields"`
	WhereColFields []WhereColField `json:"whereColFields"`
	Where          *conditonwhere.MaskWhereCondition `json:"where"`
	WhereDataYaml string          // 原始where条件的YAML字符串表示
	SqlTemplate   string `json:"sqlTemplate"`
	DynamicSql    bool // 是否动态sql
	Sort string // 排序内容

}
