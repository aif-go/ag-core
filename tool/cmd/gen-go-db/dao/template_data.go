package dao

import (
	"fmt"
	"strings"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/conditonwhere"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/table"
)

// isPointerGoType 判断 Go 类型是否为指针类型（如 *time.Time）。
func isPointerGoType(goType string) bool {
	return strings.HasPrefix(goType, "*")
}

// generateZeroValueCheck 生成零值判断代码（变量名固定为 entity）。
func generateZeroValueCheck(columns []table.ColumnData) string {
	return generateZeroValueCheckFor("entity", columns)
}

// generateZeroValueCheckFor 生成指定变量名的零值判断代码。
func generateZeroValueCheckFor(owner string, columns []table.ColumnData) string {
	var checkCode string
	for i, col := range columns {
		if i > 0 {
			checkCode += " || ("
		} else {
			checkCode += "("
		}
		// 根据字段类型生成不同的零值判断条件（唯一事实源：table.ZeroCheckExpr）
		checkCode += table.ZeroCheckExpr(owner, col, false)
		checkCode += ")"
	}
	return checkCode
}

// generatePrimaryKeyWhere 生成主键查询条件
func generatePrimaryKeyWhere(tableData *table.TableData) string {
	var whereConditions []string
	for i, pk := range tableData.PrimaryKeys {
		// 找到对应的主键列
		for _, col := range tableData.Columns {
			if col.Name == pk {
				if i > 0 {
					whereConditions = append(whereConditions, col.Name+" = ?")
				} else {
					whereConditions = append(whereConditions, col.Name+" = ?")
				}
				break
			}
		}
	}
	return strings.Join(whereConditions, " AND ")
}

// generatePrimaryKeyArgs 生成主键查询参数
func generatePrimaryKeyArgs(tableData *table.TableData) string {
	var args []string
	for _, pk := range tableData.PrimaryKeys {
		// 找到对应的主键列
		for _, col := range tableData.Columns {
			if col.Name == pk {
				args = append(args, "primaryKey."+col.JsonTag)
				break
			}
		}
	}
	return strings.Join(args, ", ")
}

// generateFindByPrimaryKeyInterface 生成 FindByPrimaryKey 接口定义
func generateFindByPrimaryKeyInterface(tableData *table.TableData) string {
	structName := tableData.StructName
	if len(tableData.PrimaryKeys) == 1 {
		// 单主键，使用类型别名
		return "\tFindByPrimaryKey(ctx context.Context, id model." + structName + "PrimaryKey) (*model." + structName + ", error)"
	}
	// 多主键，使用结构体（历史拼写 Primarkey 以别名兼容，见 model 产物）
	return "\tFindByPrimaryKey(ctx context.Context, primaryKey model." + structName + "PrimaryKey) (*model." + structName + ", error)"
}

// PKParamDecl FindByPrimaryKey 的形参声明（id / primaryKey + 类型别名）。
func (d *DaoTemplateData) PKParamDecl() string {
	if len(d.TableData.PrimaryKeys) == 1 {
		return "id model." + d.TableData.StructName + "PrimaryKey"
	}
	return "primaryKey model." + d.TableData.StructName + "PrimaryKey"
}

// PKQueryZeroCond 零值主键前置校验表达式（变量名随分支：id / primaryKey）。
func (d *DaoTemplateData) PKQueryZeroCond() string {
	var primaryKeyColumns []table.ColumnData
	for _, column := range d.TableData.Columns {
		if column.IsPrimaryKey {
			primaryKeyColumns = append(primaryKeyColumns, column)
		}
	}
	owner := "id"
	if len(d.TableData.PrimaryKeys) != 1 {
		owner = "primaryKey"
	}
	return generateZeroValueCheckForPK(owner, primaryKeyColumns)
}

// PKWhereClause 主键查询条件串。
func (d *DaoTemplateData) PKWhereClause() string {
	return generatePrimaryKeyWhere(d.TableData)
}

// PKQueryArgs 主键查询绑定参数（id / primaryKey.<Field>…）。
func (d *DaoTemplateData) PKQueryArgs() string {
	if len(d.TableData.PrimaryKeys) == 1 {
		return "id"
	}
	return generatePrimaryKeyArgs(d.TableData)
}

// generateZeroValueCheckForPK 生成主键零值判断表达式。
// 单主键时 owner 为类型别名（如 id），直接比较零值；
// 多主键时 owner 为结构体（如 primaryKey），访问其字段。
func generateZeroValueCheckForPK(owner string, columns []table.ColumnData) string {
	if len(columns) == 1 {
		col := columns[0]
		// 单主键类型别名直接比较零值，不通过 .JsonTag 访问字段
		switch col.GoType {
		case "string":
			return "(" + owner + ` == ""` + ")"
		case "bool":
			return "(!" + owner + ")"
		default:
			// 数值类型
			return "(" + owner + " == 0)"
		}
	}
	return generateZeroValueCheckFor(owner, columns)
}

// DaoTemplateData DAO 模板渲染数据（§7.2）。
type DaoTemplateData struct {
	*table.TableData
}

// HasSelfQueries 是否存在自定义查询（决定 init 调用与 strings import）。
func (d *DaoTemplateData) HasSelfQueries() bool {
	return len(d.TableData.SelfQueries) > 0
}

// StringsImport 生成 `+"`"+`"strings"`+"`"+` import 行或空行。
func (d *DaoTemplateData) StringsImport() string {
	if len(d.TableData.SelfQueries) > 0 {
		return "\"strings\""
	}
	return ""
}

// InitMethodCall New 方法内的 init 调用行（有 SelfQueries 时生成）。
func (d *DaoTemplateData) InitMethodCall() string {
	if len(d.TableData.SelfQueries) > 0 {
		return "Init" + d.TableData.StructName + "NamingSql()"
	}
	return ""
}

// HasLockCol 表是否存在乐观锁列（模板分支数据）。
func (d *DaoTemplateData) HasLockCol() bool {
	return findLockCol(d.TableData) != nil
}

// LockColJsonTag 乐观锁列实体字段名（仅 HasLockCol=true 时被模板使用）。
func (d *DaoTemplateData) LockColJsonTag() string {
	if col := findLockCol(d.TableData); col != nil {
		return col.JsonTag
	}
	return ""
}

// HasPK 是否存在主键（决定更新方法前置校验形态）。
func (d *DaoTemplateData) HasPK() bool {
	return len(d.TableData.PrimaryKeys) > 0
}

// PKZeroCond 主键零值校验表达式（单/复合主键通用）。
func (d *DaoTemplateData) PKZeroCond() string {
	var primaryKeyColumns []table.ColumnData
	for _, column := range d.TableData.Columns {
		if column.IsPrimaryKey {
			primaryKeyColumns = append(primaryKeyColumns, column)
		}
	}
	return generateZeroValueCheck(primaryKeyColumns)
}

// GuardPKCond FindByStruct 守卫的主键非零判断表达式。
func (d *DaoTemplateData) GuardPKCond() string {
	if len(d.TableData.PrimaryKeys) == 0 {
		return ""
	}
	var firstPkCol *table.ColumnData
	for _, col := range d.TableData.Columns {
		if col.Name == d.TableData.PrimaryKeys[0] {
			firstPkCol = &col
			break
		}
	}
	return table.ZeroCheckExpr("entity", *firstPkCol, true)
}

// GuardIndex FindByStruct 索引导引列判断项。
type GuardIndex struct {
	Name string // 索引名（注释使用）
	Cond string // 首列非零判断表达式
}

// GuardIndexes 索引导引列判断项集（有列索引才生成）。
func (d *DaoTemplateData) GuardIndexes() []GuardIndex {
	var items []GuardIndex
	for _, index := range d.TableData.Indexes {
		if len(index.Columns) == 0 {
			continue
		}
		for _, c := range d.TableData.Columns {
			if c.Name == index.Columns[0] {
				items = append(items, GuardIndex{Name: index.Name, Cond: table.ZeroCheckExpr("entity", c, true)})
				break
			}
		}
	}
	return items
}

// findLockCol 查找表中的乐观锁列（IsOptimisticLock 标志驱动，禁止硬编码字段名）。
func findLockCol(tableData *table.TableData) *table.ColumnData {
	for i, col := range tableData.Columns {
		if col.IsOptimisticLock {
			return &tableData.Columns[i]
		}
	}
	return nil
}

// SwitchCases 生成自定义规则查询的 switch case 分支。
func (d *DaoTemplateData) SwitchCases() string {
	var switchCases string
	for _, query := range d.TableData.SelfQueries {
		switchCases += fmt.Sprintf("\tcase \"%s\":\n\t\treturn dao.do%s(ctx, namingInfo, args)\n", query.Name, query.Name)
	}
	return switchCases
}

// DoMethod 单个自定义查询 do 方法数据。
type DoMethod struct {
	Name       string
	ResultType string
	HasPage    bool
	DynamicSql bool
}

// DoMethods 生成 do 方法数据切片。
func (d *DaoTemplateData) DoMethods() []DoMethod {
	var methods []DoMethod
	structName := d.TableData.StructName
	for _, query := range d.TableData.SelfQueries {
		var resultType string
		if query.SelectFields == "*" {
			resultType = structName
		} else {
			resultType = structName + query.Name + "Res"
		}
		methods = append(methods, DoMethod{
			Name:       query.Name,
			ResultType: resultType,
			HasPage:    query.HasPage,
			DynamicSql: query.DynamicSql,
		})
	}
	return methods
}

// PKInterfaceMethod FindByPrimaryKey 接口定义行。
func (d *DaoTemplateData) PKInterfaceMethod() string {
	return generateFindByPrimaryKeyInterface(d.TableData)
}

// NamingInfo 命名SQL参数信息。
type NamingInfo struct {
	Name     string
	RespType string
}

// ConstantTemplateData constant.go 渲染数据。
type ConstantTemplateData struct {
	StructName  string
	ModuleName  string
	ExcludeMap  string
	NamingInfos []NamingInfo
}

// BuildConstantData 构建常量模板数据（原 GetConstantTemplate 逻辑）。
func BuildConstantData(tableData *table.TableData) *ConstantTemplateData {
	data := &ConstantTemplateData{
		StructName: tableData.StructName,
		ModuleName: tableData.ModuleName,
	}

	excludeCols := []string{}
	for _, col := range tableData.Columns {
		if col.IsJavaVersion || col.IsAutoCreate || col.IsAutoUpdate || col.IsOptimisticLock {
			excludeCols = append(excludeCols, col.JsonTag)
			continue
		}
	}

	excludeMap := ""
	for i, col := range excludeCols {
		if i > 0 {
			excludeMap += ", "
		}
		excludeMap += fmt.Sprintf("\"%s\": 0", col)
	}
	data.ExcludeMap = "{" + excludeMap + "}"

	if len(tableData.SelfQueries) > 0 {
		for _, query := range tableData.SelfQueries {
			var resultType string
			var respTypeFormat string
			if query.HasPage {
				resultType = tableData.StructName + query.Name + "PageRes"
				respTypeFormat = "(*model.%s)(nil)"
			} else {
				if query.SelectFields == "*" {
					resultType = tableData.StructName
				} else {
					resultType = tableData.StructName + query.Name + "Res"
				}
				respTypeFormat = "([]*model.%s)(nil)"
			}
			data.NamingInfos = append(data.NamingInfos, NamingInfo{
				Name:     query.Name,
				RespType: fmt.Sprintf(respTypeFormat, resultType),
			})
		}
	}
	return data
}

// NamingSqlData namingsql.go 渲染数据。
type NamingSqlData struct {
	StructName string
	InitCalls  string
}

// BuildNamingSqlData 构建命名SQL模板数据（原 GetNamingSqlTemplate 逻辑）。
func BuildNamingSqlData(tableData *table.TableData, dbType string) *NamingSqlData {
	structName := tableData.StructName
	var initCalls string

	upperDbType := strings.ToUpper(dbType)

	if dbType == "" {
		initCalls = "\t// 执行一次初始化操作\n\tInit" + structName + "MYSQL()\n\t// 执行一次初始化操作\n\tInit" + structName + "DB2()"
	} else if upperDbType == "MYSQL" {
		initCalls = "\t// 执行一次初始化操作\n\tInit" + structName + "MYSQL()"
	} else if upperDbType == "DB2" {
		initCalls = "\t// 执行一次初始化操作\n\tInit" + structName + "DB2()"
	}

	return &NamingSqlData{
		StructName: structName,
		InitCalls:  initCalls,
	}
}

// DBTypeNamingSqlData dbtype_namingsql.go 渲染数据。
type DBTypeNamingSqlData struct {
	SqlExamplesBlock string
	InitFunc         string
}

// getPrimaryKey 获取主键列名列表
func getPrimaryKey(tableData *table.TableData) []string {
	if len(tableData.PrimaryKeys) > 0 {
		return tableData.PrimaryKeys
	}

	var primaryKeys []string
	for _, col := range tableData.Columns {
		if col.IsPrimaryKey {
			primaryKeys = append(primaryKeys, col.Name)
		}
	}
	return primaryKeys
}

// BuildDBTypeNamingSqlData 构建数据库类型命名SQL模板数据（原 GetDBTypeNamingSqlTemplate 逻辑）。
func BuildDBTypeNamingSqlData(tableData *table.TableData, dbType string) (*DBTypeNamingSqlData, error) {
	structName := tableData.StructName
	tableName := tableData.TableName

	primaryKeys := getPrimaryKey(tableData)

	var sqlExamples []string
	for _, query := range tableData.SelfQueries {
		var baseSql string
		var sortClause = query.Sort
		var whereClause string
		if query.DynamicSql {
			baseSql = query.SqlTemplate
		} else {
			selectClause := "SELECT *"
			if query.SelectFields != "" && query.SelectFields != "*" {
				selectClause = "SELECT " + query.SelectFields
			}

			if query.Where != nil {
				sqlwhere, err := conditonwhere.GenerateWhereSQL(query.Where)
				if err != nil {
					return nil, err
				}
				whereClause = "WHERE " + sqlwhere
			}

			if sortClause != "" && len(primaryKeys) > 0 {
				sortClause = " ORDER BY " + strings.Join(primaryKeys, ", ")
			}
			baseSql = selectClause + " FROM " + tableName + " " + whereClause
		}

		if query.HasPage {
			var pageSql string
			if dbType == "MYSQL" {
				pageSql = baseSql + sortClause + " LIMIT @Start, @End"
			} else if dbType == "DB2" {
				fieldsPart := "*"
				if query.SelectFields != "" && query.SelectFields != "*" {
					fieldsPart = query.SelectFields
				}
				fromWhereStart := strings.Index(baseSql, " FROM ")
				fromWhereClause := ""
				if fromWhereStart != -1 {
					fromWhereClause = baseSql[fromWhereStart:]
				}
				pageSql = "SELECT " + fieldsPart + " FROM (SELECT " + fieldsPart + ", ROW_NUMBER() OVER(ORDER BY " + strings.Join(primaryKeys, ", ") + ") AS RN " + fromWhereClause + ") AS T WHERE RN BETWEEN @Start AND @End"
			}
			if pageSql != "" {
				pageSqlExample := fmt.Sprintf("const %s_%s_%s = \"%s\"", dbType, structName, query.Name, pageSql)
				sqlExamples = append(sqlExamples, pageSqlExample)

				countSql := "SELECT COUNT(*) FROM " + tableName + " " + whereClause
				countSqlExample := fmt.Sprintf("const %s_%s_%s_Count = \"%s\"", dbType, structName, query.Name, countSql)
				sqlExamples = append(sqlExamples, countSqlExample)
			}
		} else {
			sqlExample := fmt.Sprintf("const %s_%s_%s = \"%s%s\"", dbType, structName, query.Name, baseSql, sortClause)
			sqlExamples = append(sqlExamples, sqlExample)
		}
	}

	initFunc := fmt.Sprintf("func Init%s%s() {\n", structName, dbType)
	for _, query := range tableData.SelfQueries {
		initFunc += fmt.Sprintf("\t%sNamingSqlMap[\"%s_%s_%s\"] = %s_%s_%s\n", structName, dbType, structName, query.Name, dbType, structName, query.Name)
		if query.HasPage {
			initFunc += fmt.Sprintf("\t%sNamingSqlMap[\"%s_%s_%s_Count\"] = %s_%s_%s_Count\n", structName, dbType, structName, query.Name, dbType, structName, query.Name)
		}
	}
	if len(tableData.SelfQueries) == 0 {
		initFunc += fmt.Sprintf("\t%sNamingSqlMap[\"%s_%s_Default\"] = %s_%s_Default\n", structName, dbType, structName, dbType, structName)
	}
	initFunc += "}\n"

	return &DBTypeNamingSqlData{
		SqlExamplesBlock: strings.Join(sqlExamples, "\n\n") + "\n\n",
		InitFunc:         initFunc,
	}, nil
}
