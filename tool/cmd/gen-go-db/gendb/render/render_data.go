package render

import (
	"sync"
)

var General string = "General"
var Unique string = "Unqiue"

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
	GoType    string `yaml:"GoType"`
	GoColName string `yaml:"GoColName"`
	DbType    string `yaml:"DbType"`
	// DDLType       string
	DbColName     string `yaml:"DbColName"`
	PrimaryKey    bool `yaml:"PrimaryKey"`
	NotNullFlag   bool `yaml:"NotNullFlag"`
	Length        string `yaml:"Length"`
	DecimalLength string `yaml:"-"`
	Comment       string  `yaml:"Comment"`
	Description   string  `yaml:"Description"`
	DefaultVal    string  `yaml:"DefaultVal"`
	// GeneralIndexName   string
	// UniqueIndexName    string
	ColumnRefIndexList []*ColumnRefIndex `yaml:"-"`
	Priority           string   `yaml:"Priority"`
	AutoUpdate         bool   `yaml:"AutoUpdate"`
	AutoCreate         bool   `yaml:"AutoCreate"`
	AutoIncrement      bool   `yaml:"AutoIncrement"`
	EndSymbol          string `yaml:"-"`
}

type NamingSqlData struct {
	MethodName       string `yaml:"MethodName"`
	ParamColNameList []SqlParameter  `yaml:"ParamColNameList"`
	// 自定sql中涉及到的参数列，重复的可省略
	// BindParam []string
	NamingSql string  `yaml:"NamingSql"`
	DbType string `yaml:"DbType"`
	// 对于自定义sql中涉及到自定义列的部分
	SelectColumns []*SelectColumn `yaml:"-"`
	// 分页查询时，获取总页数的sql
	PageCountSql string `yaml:"-"`
	Page bool `yaml:"-"`
}

type SelectColumn struct{
	ColumnName string // 根据它获取GoType值
	Alias string // 定义的别名
	GoType string
	Tag string
}

type SqlParameter struct{
	ColName string
	ParameterName string
	IsSlice bool
}

type NamingSqlParameter struct{
	DbName string
	ParameterName string
	NotSlice bool
}

type TableData struct {
	DbType string
	SchemaName  string
	ObjectName  string
	ModelName   string
	TableName   string
	Imports     []string
	DaoImports  []string
	PackageName string
	// 列数据
	TableModelList []*TableModel
	// 索引数据
	GeneralIndexList []*IndexData
	// 唯一索引数据
	UniqueIndexList []*IndexData
	// 主键数据
	// PrimryIndexList []*IndexData
	PrimryRIndex *IndexData
	PrimryUIndex *IndexData
	PrimryDIndex *IndexData

	RNamingSqlList   []*NamingSqlTemplate
	CUDNamingSqlList []*NamingSqlTemplate
	NamingSqlMap map[string]*NamingSqlData

	// 数据转换用
	ColumnDataMap map[string]*ColumnData
	// 构建ddl使用
	ColumnList []*ColumnData
	// 构建ddl使用
	Engine                 string
	Encode                 string
	Sort                   string
	PrimaryKeyList         string
	DaoImportsFilterMap    sync.Map
	EntityImportsFilterMap sync.Map
	UniqueIndexSort        int
	NoNameUiqueList        []string
	NamingSqlMapEnable   bool
}

type IndexData struct {
	IndexName string
	// 方法参数列表
	BindParamList []*BindParam
	// 方法参数列表
	HashParamters string `yaml:"-"`
	MethodName    string `yaml:"-"`
	IndexType     string
	IndexColList  string `yaml:"-"` 
	Imports       []string `yaml:"-"`
	Use           bool `yaml:"-"`
}

type NamingSqlTemplate struct {
	MethodName string
	BindParam  []*BindParam `yaml:"-"`
	NamingSql  string
	SelectColumns []*SelectColumn `yaml:"-"`
	MethodResName string `yaml:"-"`
	GenerateBol  bool `yaml:"-"`
	// DbType string
}

type BindParam struct {
	GoType    string `yaml:"-"`
	GoColName string `yaml:"-"`
	DbColName string
}

type TableModel struct {
	GoType    string
	GoColName string
	DbColName string
	GoTag     string
}

// ColumnRefIndex 列和普通索引的映射关系 比如对应的索引名+在索引中的排序
type ColumnRefIndex struct {
	IndexName string
	Priority  string
	IndexType string // 索引类型 普通索引还是唯一索引
}

// 构建DDL模板数据
type TableDDL struct {
	TableName        string
	GeneralIndexList []*IndexData
	UniqueIndexList  []*IndexData
	PrimaryKey       string
	ColumnList       []*ColumnData
}


type YamlData struct{
	ModelName  string `yaml:"ModelName"`
	// 表名
	TableName string `yaml:"TableName"`
	// 表元素的列数据集合
	ColumnList []*ColumnData `yaml:"ColumnList"`
	// 普通索引集合
	GeneralIndexList []*IndexData `yaml:"GeneralIndexList"`
	// 约束索引集合
	UniqueIndexList []*IndexData `yaml:"UniqueIndexList"`
	// 主键
	PrimaryKeyList []string `yaml:"PrimaryKeyList"`
	// 自定义sql集合 key为后续要生成的方法名
	RNamingSqlList   []*NamingSqlTemplate  `yaml:"RNamingSqlList"`
	CUDNamingSqlList []*NamingSqlTemplate `yaml:"CUDNamingSqlList"`
	Encode string `yaml:"Encode"` // 数据库编码
	Engine string `yaml:"Engine"` // 数据库存储引擎
	Sort  string `yaml:"Sort"` // 数据库排序规则
}


// 这一套东西要定义到model中
type WhereCondition[T any] struct{

	Condition  Condition[any] // 包含使用的列名和列的值

	Operate  Operate

}

type Condition[T any] struct{
	Equal T
	NotEqual T
	In T
}

type Operate struct{
	Or string
	And string
}
