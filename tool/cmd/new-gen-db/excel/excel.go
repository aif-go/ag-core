package excel

// TableInfo 表结构信息
type ExcelInfo struct {
	Name        string
	Columns     []*ColumnInfo
	PrimaryKey  []string
	Constraints []*ConstraintInfo
	Indexes     []*IndexInfo
	SelfQueries map[string]*SelfQueryInfo
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name          string
	Type          string
	Length        string
	NotNull       bool
	Default       string
	AutoIncrement bool
	SupportUpdate bool
	Description   string
	Tag           string
}

// ConstraintInfo 约束信息
type ConstraintInfo struct {
	Name    string
	Columns []string
}

// IndexInfo 索引信息
type IndexInfo struct {
	Name    string
	Columns []string
}

// SelfQueryInfo 自定义查询信息
type SelfQueryInfo struct {
	SelectFields string
	Where        *WhereClause
	Page         bool
}

// WhereClause WHERE子句信息
type WhereClause struct {
	Operator   string
	Conditions []*Condition
}

// Condition 条件信息
type Condition struct {
	Operator   string
	Conditions []*Condition
	Expr       string
}
