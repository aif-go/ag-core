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

// generateZeroValueCheck 生成零值判断代码
func generateZeroValueCheck(columns []table.ColumnData) string {
	var checkCode string
	checkCode += "("
	for i, col := range columns {
		if i > 0 {
			checkCode += " || ("
		} else {
			checkCode += "("
		}
		// 根据字段类型生成不同的零值判断条件
		if isPointerGoType(col.GoType) {
			checkCode += "entity." + col.JsonTag + " == nil"
		} else {
			switch col.GoType {
			case "string":
				checkCode += "entity." + col.JsonTag + " == \"\""
			case "time.Time", "decimal.Decimal":
				checkCode += "entity." + col.JsonTag + ".IsZero()"
			case "bool":
				checkCode += "!entity." + col.JsonTag
			default:
				// 数值类型
				checkCode += "entity." + col.JsonTag + " == 0"
			}
		}
		checkCode += ")"
	}
	checkCode += ")"
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
	// 多主键，使用结构体
	return "\tFindByPrimaryKey(ctx context.Context, primaryKey model." + structName + "Primarkey) (*model." + structName + ", error)"
}

// generateFindByPrimaryKeyMethod 生成 FindByPrimaryKey 方法实现
func generateFindByPrimaryKeyMethod(tableData *table.TableData) string {
	structName := tableData.StructName

	// 生成查询条件
	whereClause := generatePrimaryKeyWhere(tableData)

	if len(tableData.PrimaryKeys) == 1 {
		// 单主键，使用类型别名
		return `// FindByPrimaryKey 根据主键查询
func (dao *` + structName + `Dao) FindByPrimaryKey(ctx context.Context, id model.` + structName + `PrimaryKey) (*model.` + structName + `, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return nil, err
	}
	
	var entity model.` + structName + `
	result := db.Where("` + whereClause + `", id).First(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &entity, result.Error
}`
	}

	// 多主键，使用结构体
	argsClause := generatePrimaryKeyArgs(tableData)
	return `// FindByPrimaryKey 根据主键查询
func (dao *` + structName + `Dao) FindByPrimaryKey(ctx context.Context, primaryKey model.` + structName + `Primarkey) (*model.` + structName + `, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return nil, err
	}
	
	var entity model.` + structName + `
	result := db.Where("` + whereClause + `", ` + argsClause + `).First(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &entity, result.Error
}`
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

// GuardCheck 守卫块代码（主键/索引引导列非零判断，含恒报错兜底）。
func (d *DaoTemplateData) GuardCheck() string {
	tableData := d.TableData
	var guardCheck string
	guardCheck += "\t// 检查是否使用了主键或索引，避免全表扫描\n"
	guardCheck += "\tkeyUsed := false\n"

	// 主键检查（主键即索引；仅当主键存在时生成）
	if len(tableData.PrimaryKeys) > 0 {
		var firstPkCol *table.ColumnData
		for _, col := range tableData.Columns {
			if col.Name == tableData.PrimaryKeys[0] {
				firstPkCol = &col
				break
			}
		}
		if firstPkCol != nil {
			var nullCheck string
			if isPointerGoType(firstPkCol.GoType) {
				nullCheck = "entity." + firstPkCol.JsonTag + " != nil"
			} else {
				switch firstPkCol.GoType {
				case "string":
					nullCheck = "entity." + firstPkCol.JsonTag + " != \"\""
				case "time.Time", "decimal.Decimal":
					nullCheck = "!entity." + firstPkCol.JsonTag + ".IsZero()"
				case "bool":
					nullCheck = "entity." + firstPkCol.JsonTag
				default:
					nullCheck = "entity." + firstPkCol.JsonTag + " != 0"
				}
			}
			guardCheck += "\t// 检查主键\n"
			guardCheck += "\tif " + nullCheck + " {\n"
			guardCheck += "\t\tkeyUsed = true\n"
			guardCheck += "\t}\n"
		}
	}

	// 索引引导列检查（仅当索引存在时生成，最左前缀即可命中索引）
	if len(tableData.Indexes) > 0 {
		var validIndexes []table.IndexData
		for _, index := range tableData.Indexes {
			if len(index.Columns) > 0 {
				validIndexes = append(validIndexes, index)
			}
		}

		// 按优先级排序索引（如果有）
		// 简单实现：假设索引已经按优先级排序
		for _, index := range validIndexes {
			guardCheck += "\t// 检查索引 " + index.Name + "\n"

			colName := index.Columns[0]
			for _, col := range tableData.Columns {
				if col.Name == colName {
					var nullCheck string
					if isPointerGoType(col.GoType) {
						nullCheck = "entity." + col.JsonTag + " != nil"
					} else {
						switch col.GoType {
						case "string":
							nullCheck = "entity." + col.JsonTag + " != \"\""
						case "time.Time", "decimal.Decimal":
							nullCheck = "!entity." + col.JsonTag + ".IsZero()"
						case "bool":
							nullCheck = "entity." + col.JsonTag
						default:
							nullCheck = "entity." + col.JsonTag + " != 0"
						}
					}

					guardCheck += "\tif " + nullCheck + " {\n"
					guardCheck += "\t\tkeyUsed = true\n"
					guardCheck += "\t}\n"
					break
				}
			}
		}
	}

	// 最终守卫判断（无条件生成；无主键无索引表 keyUsed 恒为 false，恒报错）
	guardCheck += "\tif !keyUsed {\n"
	guardCheck += "\t\treturn nil, errors.New(\"query not use any index\")\n"
	guardCheck += "\t}\n"
	return guardCheck
}

// PrimaryKeyUpdate 生成主键更新条件代码。
func (d *DaoTemplateData) PrimaryKeyUpdate() string {
	tableData := d.TableData
	var primaryKeyUpdate string

	var primaryKeyColumns []table.ColumnData
	var uniqueKeyColumns []table.ColumnData

	for _, column := range tableData.Columns {
		if column.IsPrimaryKey {
			primaryKeyColumns = append(primaryKeyColumns, column)
		}
	}

	for _, index := range tableData.Indexes {
		if index.IsUnique {
			for _, colName := range index.Columns {
				for _, column := range tableData.Columns {
					if column.Name == colName {
						exists := false
						for _, existingCol := range uniqueKeyColumns {
							if existingCol.Name == column.Name {
								exists = true
								break
							}
						}
						if !exists {
							uniqueKeyColumns = append(uniqueKeyColumns, column)
						}
						break
					}
				}
			}
		}
	}

	if len(primaryKeyColumns) > 0 {
		primaryKeyUpdate = "\t// 检查主键是否为空，如果为空继续检查唯一键\n"
		primaryKeyUpdate += "\tif " + generateZeroValueCheck(primaryKeyColumns) + " {\n"
		primaryKeyUpdate += "\t\treturn 0, errors.New(\"when update,primary key or unique key is required\")\n"
		primaryKeyUpdate += "\t} else {\n"
		for _, pk := range primaryKeyColumns {
			primaryKeyUpdate += "\t\twhere[\"" + pk.Name + "\"] = entity." + pk.JsonTag + "\n"
		}
		primaryKeyUpdate += "\t}\n"
	}
	return primaryKeyUpdate
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

// PKMethodBody FindByPrimaryKey 方法实现。
func (d *DaoTemplateData) PKMethodBody() string {
	return generateFindByPrimaryKeyMethod(d.TableData)
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
		if col.IsJavaVersion || col.IsAutoCreate || col.IsAutoUpdate {
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
