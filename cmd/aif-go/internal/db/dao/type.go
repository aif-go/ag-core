package dao

type SchemaConfig struct {
	ConfigPath     string
	PackageName    string
	OutputPath     string
	SchemaName     string
	FileType       string
	TargetFileName string
	ModuleName     string
	DiffPath       string
}

type SchemaData struct {
	SchemaName string
	ObjectName string
	TableName  string
	// Columns     []*ColumnData
	ColMap      map[string]*ColumnData
	Imports     map[string]string
	PackageName string
	// table primary key
	PrimaryKeys []string
	// table general index
	GeneralIndexMap map[string][]string
	// table unique index
	UniqueIndexMap map[string][]string
	MethodNameMap  map[string]string
	NameMap        map[string]string

	NamingSqlDatas []*NamingSqlData
	// 查询的自定义sql
	// RNamingSqlDatas []*NamingSqlData
	// 增,删,改的自定义sql
	// CUDNamingSqlDatas []*NamingSqlData
}

type ColumnData struct {
	GoType           string
	GoColName        string
	DbColName        string
	PrimaryKey       bool
	NotNullFlag      bool
	Length           string
	Comment          string
	DefaultVal       string
	GeneralIndexName string
	UniqueIndexName  string
	Priority         string
	AutoUpdate       bool
	AutoCreate       bool
}

type NamingSqlData struct {
	MethodName       string
	ParamColNameList []string
	// 自定sql中涉及到的参数列，重复的可省略
	// BindParam []string
	NamingSql string
}

type TableData struct {
	SchemaName  string
	ObjectName  string
	ModuleName  string
	TableName   string
	Imports     []string
	PackageName string
	// 列数据
	TableModelList []*TableModel
	// 索引数据
	GeneralIndexList []*IndexData
	// 唯一索引数据
	UniqueIndexList []*IndexData
	// 主键数据
	// PrimaryIndexList []*IndexData
	PrimaryRIndex *IndexData
	PrimaryUIndex *IndexData
	PrimaryDIndex *IndexData

	RNamingSqlList   []*NamingSqlTemplate
	CUDNamingSqlList []*NamingSqlTemplate

	// 数据转换用
	ColumnDataMap map[string]*ColumnData
}

type IndexData struct {
	IndexName string
	// 方法参数列表
	BindParamList []*BindParam
	// 方法参数列表
	HashParameters string
	MethodName     string
}

type NamingSqlTemplate struct {
	MethodName string
	BindParam  []*BindParam
	NamingSql  string
}

type BindParam struct {
	GoType    string
	GoColName string
	DbColName string
}

type TableModel struct {
	GoType    string
	GoColName string
	DbColName string
	GoTag     string
}

type YamlData struct {
	// dbname 用来做dao模块的model
	SchemaName string
	ModuleName string
	// 表名
	TableName string
	// 表元素的列数据集合
	ColumnList []*ColumnData
	// 普通索引集合
	GeneralIndexList []*IndexData
	// 约束索引集合
	UniqueIndexList []*IndexData
	// 主键
	PrimaryKeyList []string
	// 自定义sql集合 key为后续要生成的方法名
	NamingSqlList []*NamingSqlData
}
