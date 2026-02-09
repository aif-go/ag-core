package dao

import (
	"fmt"
	"strings"

	"ag-core/tool/cmd/new-gen-db/table"
)

// GetDaoTemplate 获取DAO模板代码
func GetDaoTemplate(tableData *table.TableData) string {
	moduleName := tableData.ModuleName
	structName := tableData.StructName
	tableName := tableData.TableName

	// 构建AllowUpdateCols的字符串表示
	allowUpdateColsStr := "[]string{"
	for i, col := range tableData.AllowUpdateCols {
		if i > 0 {
			allowUpdateColsStr += ", "
		}
		allowUpdateColsStr += "\"" + col + "\""
	}
	allowUpdateColsStr += "}"

	// 获取主键列
	var primaryKeys []string
	for _, col := range tableData.Columns {
		if col.IsPrimaryKey {
			primaryKeys = append(primaryKeys, col.Name)
		}
	}

	// 构建主键查询代码
	var primaryKeyCheck string
	if len(primaryKeys) > 0 {
		primaryKeyCheck = "\t// 检查主键是否为空\n"
		for _, col := range tableData.Columns {
			if col.IsPrimaryKey {
				// 根据字段类型生成不同的空值判断条件
				var nullCheck string
				switch col.GoType {
				case "string":
					nullCheck = "entity." + col.JsonTag + " != \"\""
				case "time.Time":
					nullCheck = "!entity." + col.JsonTag + ".IsZero()"
				default:
					// 数值类型
					nullCheck = "entity." + col.JsonTag + " != 0"
				}
				primaryKeyCheck += "\tif " + nullCheck + " {\n"
				primaryKeyCheck += "\t\tdb = db.Where(\"" + col.JsonTag + " = ?\", entity." + col.JsonTag + ")\n"
				primaryKeyCheck += "\t\tresult := db.Find(&list)\n"
				primaryKeyCheck += "\t\treturn list, result.Error\n"
				primaryKeyCheck += "\t}\n"
			}
		}
	}

	// 构建索引查询代码
	var indexCheck string
	if len(tableData.Indexes) > 0 {
		indexCheck = "\t// 检查索引列，确保使用了索引\n"
		indexCheck += "\tindexUsed := false\n"

		// 过滤掉空索引
		var validIndexes []table.IndexData
		for _, index := range tableData.Indexes {
			if len(index.Columns) > 0 {
				validIndexes = append(validIndexes, index)
			}
		}

		// 按优先级排序索引（如果有）
		// 简单实现：假设索引已经按优先级排序
		for _, index := range validIndexes {
			indexCheck += "\t// 检查索引 " + index.Name + "\n"

			// 生成索引列检查条件
			colName := index.Columns[0]
			for _, col := range tableData.Columns {
				if col.Name == colName {
					// 根据字段类型生成不同的空值判断条件
					var nullCheck string
					switch col.GoType {
					case "string":
						nullCheck = "entity." + col.JsonTag + " != \"\""
					case "time.Time":
						nullCheck = "!entity." + col.JsonTag + ".IsZero()"
					default:
						// 数值类型
						nullCheck = "entity." + col.JsonTag + " != 0"
					}

					// 每个索引都是独立的if检查
					indexCheck += "\tif " + nullCheck + " {\n"
					indexCheck += "\t\tdb = db.Where(\"" + col.JsonTag + " = ?\", entity." + col.JsonTag + ")\n"

					// 其他列作为次要条件
					for j := 1; j < len(index.Columns); j++ {
						secondaryColName := index.Columns[j]
						for _, secondaryCol := range tableData.Columns {
							if secondaryCol.Name == secondaryColName {
								// 根据字段类型生成不同的空值判断条件
								var secondaryNullCheck string
								switch secondaryCol.GoType {
								case "string":
									secondaryNullCheck = "entity." + secondaryCol.JsonTag + " != \"\""
								case "time.Time":
									secondaryNullCheck = "!entity." + secondaryCol.JsonTag + ".IsZero()"
								default:
									// 数值类型
									secondaryNullCheck = "entity." + secondaryCol.JsonTag + " != 0"
								}
								indexCheck += "\t\tif " + secondaryNullCheck + " {\n"
								indexCheck += "\t\t\tdb = db.Where(\"" + secondaryCol.JsonTag + " = ?\", entity." + secondaryCol.JsonTag + ")\n"
								indexCheck += "\t\t}\n"
								break
							}
						}
					}

					indexCheck += "\t\tindexUsed = true\n"
					indexCheck += "\t}\n"
					break
				}
			}
		}

		// 添加索引使用检查
		indexCheck += "\n\t// 如果没有使用任何索引，返回错误\n"
		indexCheck += "\tif !indexUsed {\n"
		indexCheck += "\t\treturn nil, errors.New(\"query not use any index\")\n"
		indexCheck += "\t}\n"
	}

	// 生成UpdateByPrimaryKey方法的switch case代码
	var updateSwitchCase string
	for _, col := range tableData.AllowUpdateCols {
		for _, column := range tableData.Columns {
			if col == column.Name {
				updateSwitchCase += "\t\tcase \"" + col + "\":\n"
				// 根据字段类型生成不同的零值判断条件
				switch column.GoType {
				case "string":
					updateSwitchCase += "\t\t\tif entity." + column.JsonTag + " != \"\" {\n"
				case "time.Time":
					updateSwitchCase += "\t\t\tif !entity." + column.JsonTag + ".IsZero() {\n"
				// case "bool":
				// 	// 布尔类型特殊处理，因为false也是零值，但可能需要更新为false
				// 	updateSwitchCase += "\t\t\tif true {\n"
				default:
					// 数值类型
					updateSwitchCase += "\t\t\tif entity." + column.JsonTag + " != 0 {\n"
				}
				updateSwitchCase += "\t\t\t\tallowedUpdateCols = append(allowedUpdateCols, \"" + col + "\")\n"
				updateSwitchCase += "\t\t\t}\n"
				break
			}
		}
	}

	// 生成主键更新条件代码
	var primaryKeyUpdate string
	for _, column := range tableData.Columns {
		if column.IsPrimaryKey {
			primaryKeyUpdate += "\twhere[\"" + column.Name + "\"] = entity." + column.JsonTag + "\n"
		}
	}

	var reflectImpor string
	if len(tableData.SelfQueries) > 0 {
		reflectImpor = "reflect"
	}

	// 生成自定义规则查询的switch语句和do方法
	switchCases := generateCustomerRuleSwitch(tableData)
	doMethods := generateDoMethods(tableData)

	// 构建完整的模板字符串
	return `package dao

import (
	db "ag-core/contribute/agdb/gormdb"
	"` + moduleName + `/repository/model"
	"context"
	"` + reflectImpor + `"
	"errors"

	agdao "ag-core/contribute/agdb/agdao"
	"strings"

	"gorm.io/gorm"
)

// ` + structName + `Dao ` + tableName + ` DAO
// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT
type ` + structName + `Dao struct {
	*db.Repository
	info    agdao.TableInfo
	baseDao agdao.BaseDao
}

// I` + structName + `Dao ` + structName + ` DAO接口
type I` + structName + `Dao interface {
	InsertOne(ctx context.Context, entity *model.` + structName + `) (int64, error)
	InsertOneIgnoreZeroValCols(ctx context.Context, entity *model.` + structName + `) (int64, error)
	UpdateByPrimaryKey(ctx context.Context, entity *model.` + structName + `) (int64, error)
	UpdaeByPrimaryKeyIngoreZeroValCols(ctx context.Context, entity *model.` + structName + `) (int64, error)
	FindByStruct(ctx context.Context, entity *model.` + structName + `) ([]*model.` + structName + `, error)
	FindByCustomerRule(ctx context.Context, namingInfo *db.NameingSqlArgInfo, args any) (any, error)
}

// New` + structName + `Dao get dao instance
func New` + structName + `Dao(repository *db.Repository, baseDao agdao.BaseDao) I` + structName + `Dao {
	Init` + structName + `NamingSql()
	return &` + structName + `Dao{
		Repository: repository,
		baseDao:    baseDao,
		info: agdao.TableInfo{
			TableName: "` + tableName + `",
		},
	}
}

// insertOne 插入一条数据库数据
func (dao *` + structName + `Dao) InsertOne(ctx context.Context, entity *model.` + structName + `) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	result := db.Create(entity)
	return result.RowsAffected, result.Error
}

// InsertOneIgnorenNullCols 插入数据时，自动剔除零值的列
func (dao *` + structName + `Dao) InsertOneIgnoreZeroValCols(ctx context.Context, entity *model.` + structName + `) (int64, error) {
	insertIgnoreZeroValSlice := db.CollectZeroValWithOmitEmpty(entity, exclude` + structName + `ZeroColNames)
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	result := db.Omit(insertIgnoreZeroValSlice...).Create(entity)
	return result.RowsAffected, result.Error
}

// UpdateByPrimaryKey 根据主键更新，该操作只适合从数据库查询原实体修改值之后使用
func (dao *` + structName + `Dao) UpdateByPrimaryKey(ctx context.Context, entity *model.` + structName + `) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	// ` + structName + `AllowUpdateCols
	// 1. 定义一个切片，存储非零值的列名
	allowedUpdateCols := []string{}
	
	// 2. 只将非零值的列名添加到切片中
	for _, col := range model.` + structName + `AllowUpdateCols {
		switch col {
` + updateSwitchCase + `		}
	}
	
	
	// 4. 更新条件（主键）
	where := make(map[string]any)
` + primaryKeyUpdate + `
	// 5. 使用支持更新的列
	result := db.Model(&model.` + structName + `{}).Where(where).Select(allowedUpdateCols).Updates(entity)
	return result.RowsAffected, result.Error
}

// UpdateByPriIngoreNullCols 根据主键更新，自动剔除参数中的零值列
func (dao *` + structName + `Dao) UpdaeByPrimaryKeyIngoreZeroValCols(ctx context.Context, entity *model.` + structName + `) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	// ` + structName + `AllowUpdateCols
	// 1. 定义一个切片，存储非零值的列名
	allowedUpdateCols := []string{}
	
	// 2. 只将非零值的列名添加到切片中
	for _, col := range model.` + structName + `AllowUpdateCols {
		switch col {
` + updateSwitchCase + `		}
	}
	
	
	// 4. 更新条件（主键）
	where := make(map[string]any)
` + primaryKeyUpdate + `
	// 使用支持更新的列
	result := db.Model(&model.` + structName + `{}).Where(where).Select(allowedUpdateCols).Updates(entity)
	return result.RowsAffected, result.Error
}

// FindByStruct 根据实体查询
func (dao *` + structName + `Dao) FindByStruct(ctx context.Context, entity *model.` + structName + `) ([]*model.` + structName + `, error) {
	var list []*model.` + structName + `
	db, err := dao.newDB(ctx)
	if err != nil {
		return nil, err
	}

` + primaryKeyCheck + `

` + indexCheck + `

	// 执行查询
	result := db.Find(&list)
	return list, result.Error
}

// FindByCustomerRule 根据自定义规则查询
func (dao *` + structName + `Dao) FindByCustomerRule(ctx context.Context, namingInfo *db.NameingSqlArgInfo, args any) (any, error) {

	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	if namingInfo == nil {
		return nil, errors.New("namingInfo is nil")
	}

	if namingInfo.SqlName == "" {
		return nil, errors.New("namingInfo.SqlName is empty")
	}

	// 判断请求参数类型和实际类型是否一致
	reqType := reflect.TypeOf(namingInfo.ReqType)
	reqValue := reflect.ValueOf(args)
	if reqType != reqValue.Type() {
		return nil, errors.New("req type not match")
	}
	switch namingInfo.SqlName {
` + switchCases + `	default:
		return nil, errors.New("not found naming sql")
	}
}
` + doMethods + `

// getInfo 获取表信息
func (dao *` + structName + `Dao) getInfo() agdao.TableInfo {
	return dao.info
}

// getApplyInfo 获取应用表信息
func (dao *` + structName + `Dao) getApplyInfo(ctx context.Context) agdao.TableInfo {
	info := dao.getInfo()
	dao.baseDao.ApplyTbInfoOpts(ctx, &info)
	return info
}

// newDB 创建一个新的DB实例
func (dao *` + structName + `Dao) newDB(ctx context.Context) (*gorm.DB, error) {
	db := dao.DB(ctx)
	info := dao.getApplyInfo(ctx)
	tbname := info.TableName
	if tbname == "" {
		return nil, errors.New("表名不能为空")
	}

	db = db.Table(tbname)
	return db, nil
}




`
}

// GetNamingSqlTemplate 获取命名SQL模板代码
func GetNamingSqlTemplate(tableData *table.TableData, dbType string) string {
	structName := tableData.StructName
	var initCalls string

	// 将dbType转换为大写，确保大小写不敏感
	upperDbType := strings.ToUpper(dbType)

	// 根据dbType参数决定生成哪些初始化函数调用
	if dbType == "" {
		// 未指定dbType，生成所有数据库类型的初始化函数调用
		initCalls = "\t// 执行一次初始化操作\n\tInit" + structName + "MYSQL()\n\t// 执行一次初始化操作\n\tInit" + structName + "DB2()"
	} else if upperDbType == "MYSQL" {
		// 只生成MySQL的初始化函数调用
		initCalls = "\t// 执行一次初始化操作\n\tInit" + structName + "MYSQL()"
	} else if upperDbType == "DB2" {
		// 只生成DB2的初始化函数调用
		initCalls = "\t// 执行一次初始化操作\n\tInit" + structName + "DB2()"
	}

	return `package dao

// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT

func Init` + structName + `NamingSql() {
` + initCalls + `
}
`
}

// generateWhereSQL 生成where条件SQL语句
func generateWhereSQL(condition *table.WhereCondition) string {
	if condition.Expr != "" {
		return condition.Expr
	}

	if len(condition.Conditions) == 0 {
		return "1=1"
	}

	conditions := make([]string, 0, len(condition.Conditions))
	for _, cond := range condition.Conditions {
		conditions = append(conditions, generateWhereSQL(&cond))
	}

	operator := condition.Operator
	if operator == "" {
		operator = "AND"
	}

	return "(" + strings.Join(conditions, " "+operator+" ") + ")"
}

// generateCustomerRuleSwitch 生成自定义规则查询的switch语句
func generateCustomerRuleSwitch(tableData *table.TableData) string {
	var switchCases string
	for _, query := range tableData.SelfQueries {
		switchCases += fmt.Sprintf("\tcase \"%s\":\n\t\treturn dao.do%s(ctx, namingInfo, args)\n", query.Name, query.Name)
	}
	return switchCases
}

// generateDoMethods 生成do方法
func generateDoMethods(tableData *table.TableData) string {
	var doMethods string
	structName := tableData.StructName
	for _, query := range tableData.SelfQueries {
		// 根据selectFields决定返回类型
		var resultType string
		if query.SelectFields == "*" {
			resultType = structName
		} else {
			resultType = structName + query.Name + "Res"
		}

		if query.HasPage {
			// 生成分页的do方法
			doMethods += `// do` + query.Name + ` 执行` + query.Name + `查询（分页）
func (dao *` + structName + `Dao) do` + query.Name + `(ctx context.Context, namingInfo *db.NameingSqlArgInfo, args any) (*model.` + structName + query.Name + `PageRes, error) {

	queryArgs, ok := args.(*model.` + structName + query.Name + `Arg)
	if !ok {
		return nil, errors.New("do` + query.Name + ` args type not match")
	}

	sqlName := dao.DbType + "_" + "` + structName + `" + "_" + namingInfo.SqlName
	execSql := ` + structName + `NamingSqlMap[sqlName]
	if execSql == "" {
		return nil, errors.New("not found naming sql")
	}
	execCountSql := ` + structName + `NamingSqlMap[sqlName+"_Count"]
	if execCountSql == "" {
		return nil, errors.New("not found naming sql count")
	}

	newTableName := dao.getApplyInfo(ctx).TableName
	if newTableName != "" {
		enity := &model.` + structName + `{}
		execSql = strings.ReplaceAll(execSql, "FROM "+enity.TableName()+" WHERE", "FROM "+newTableName+" WHERE")
		execCountSql = strings.ReplaceAll(execCountSql, "FROM "+enity.TableName()+" WHERE", "FROM "+newTableName+" WHERE")
	}

	argsMap := queryArgs.ConvertToMap()
	var totalCount int64
	result := dao.DB(ctx).Raw(execCountSql, argsMap).Scan(&totalCount)
	if result.Error != nil {
		return nil, result.Error
	}
	startRecord, endRecord, totalPage := db.CalcPageStartRecord(queryArgs.PageNum, queryArgs.PageSize, totalCount, dao.DbType)
	argsMap["Start"] = startRecord
	argsMap["End"] = endRecord
	var list []*model.` + resultType + `
	resultlist := dao.DB(ctx).Raw(execSql, argsMap).Find(&list)
	if resultlist.Error != nil {
		return nil, resultlist.Error
	}

	return &model.` + structName + query.Name + `PageRes{
		PageResult: db.PageResult{
			CurrentPage: queryArgs.PageNum,
			PageSize:    queryArgs.PageSize,
			TotalCount:  totalCount,
			TotalPage:   totalPage,
		},
		ResultList: list,
	}, nil
}

`
		} else {
			// 生成非分页的do方法
			doMethods += `// do` + query.Name + ` 执行` + query.Name + `查询（非分页）
func (dao *` + structName + `Dao) do` + query.Name + `(ctx context.Context, namingInfo *db.NameingSqlArgInfo, args any) ([]*model.` + resultType + `, error) {

	queryArgs, ok := args.(*model.` + structName + query.Name + `Arg)
	if !ok {
		return nil, errors.New("do` + query.Name + ` args type not match")
	}

	sqlName := dao.DbType + "_" + "` + structName + `" + "_" + namingInfo.SqlName
	execSql := ` + structName + `NamingSqlMap[sqlName]
	if execSql == "" {
		return nil, errors.New("not found naming sql")
	}
	newTableName := dao.getApplyInfo(ctx).TableName
	if newTableName != "" {
		enity := &model.` + structName + `{}
		execSql = strings.ReplaceAll(execSql, "FROM "+enity.TableName()+" WHERE", "FROM "+newTableName+" WHERE")
	}

	argsMap := queryArgs.ConvertToMap()
	var list []*model.` + resultType + `
	result := dao.DB(ctx).Raw(execSql, argsMap).Find(&list)
	if result.Error != nil {
		return nil, result.Error
	}
	return list, nil
}

`
		}
	}
	return doMethods
}

// GetConstantTemplate 获取常量模板代码
func GetConstantTemplate(tableData *table.TableData) string {
	structName := tableData.StructName
	moduleName := tableData.ModuleName

	// 生成常量定义
	var constants string

	// 添加命名SQL映射
	constants += fmt.Sprintf(`// %sNamingSqlMap 命名SQL映射
var %sNamingSqlMap = map[string]string{}

`, structName, structName)

	// 添加排除空值字段映射
	excludeCols := []string{}
	for _, col := range tableData.Columns {
		// 检查列的标签是否包含需要排除的标记
		if col.IsJavaVersion || col.IsAutoCreate || col.IsAutoUpdate {
			excludeCols = append(excludeCols, col.JsonTag)
			continue
		}
	}

	// 生成排除空值字段映射
	excludeMap := ""
	for i, col := range excludeCols {
		if i > 0 {
			excludeMap += ", "
		}
		excludeMap += fmt.Sprintf("\"%s\": 0", col)
	}

	constants += fmt.Sprintf(`// exclude%sZeroColNames 插入忽略空值时标记哪些字段需要排除在外
var exclude%sZeroColNames = map[string]int{%s}

`, structName, structName, excludeMap)

	// 只有当有自定义查询时才生成命名SQL参数信息
	if len(tableData.SelfQueries) > 0 {
		for _, query := range tableData.SelfQueries {
			// 根据selectFields决定返回类型
			var resultType string
			if query.SelectFields == "*" {
				resultType = structName
			} else {
				resultType = structName + query.Name + "Res"
			}
			constants += fmt.Sprintf(`
var %sNamingInfo = &db.NameingSqlArgInfo{
	SqlName:  "%s",
	ReqType:  (*model.%s%sArg)(nil),
	RespType: ([]*model.%s)(nil),
}
`, query.Name, query.Name, structName, query.Name, resultType)
		}
	}

	return `package dao

import (
	db "ag-core/contribute/agdb/gormdb"
	"` + moduleName + `/repository/model"
)

` + constants
}

// GetDBTypeNamingSqlTemplate 获取数据库类型命名SQL模板代码
func GetDBTypeNamingSqlTemplate(tableData *table.TableData, dbType string) string {
	structName := tableData.StructName
	tableName := tableData.TableName

	// 获取主键列名
	primaryKey := getPrimaryKey(tableData)

	// 生成示例SQL
	var sqlExamples []string
	for _, query := range tableData.SelfQueries {
		// 构建SELECT语句
		selectClause := "SELECT *"
		if query.SelectFields != "" && query.SelectFields != "*" {
			selectClause = "SELECT " + query.SelectFields
		}

		// 构建WHERE条件
		whereClause := "WHERE 1=1"
		if query.Where != nil {
			whereClause = "WHERE " + generateWhereSQL(query.Where)
		}

		// 构建排序语句
		sortClause := ""
		if primaryKey != "" {
			sortClause = " ORDER BY " + primaryKey
		}

		// 组合基本SQL语句
		baseSql := selectClause + " FROM " + tableName + " " + whereClause

		// 根据是否需要分页生成分页SQL或基本SQL
		if query.HasPage {
			// 需要分页，只生成分页SQL，去掉_Page后缀
			var pageSql string
			if dbType == "MYSQL" {
				// MySQL分页语法
				pageSql = baseSql + sortClause + " LIMIT @Start, @End"
			} else if dbType == "DB2" {
				// DB2分页语法（简化为两层嵌套）
				// 提取SELECT子句的字段部分
				fieldsPart := "*"
				if query.SelectFields != "" && query.SelectFields != "*" {
					fieldsPart = query.SelectFields
				}
				// 提取FROM和WHERE子句（从baseSql中提取）
				fromWhereStart := strings.Index(baseSql, " FROM ")
				fromWhereClause := ""
				if fromWhereStart != -1 {
					fromWhereClause = baseSql[fromWhereStart:]
				}
				// 构建两层嵌套的DB2分页SQL
				pageSql = "SELECT " + fieldsPart + " FROM (SELECT " + fieldsPart + ", ROW_NUMBER() OVER(ORDER BY " + primaryKey + ") AS RN " + fromWhereClause + ") AS T WHERE RN BETWEEN @Start AND @End"
			}
			if pageSql != "" {
				// 分页SQL常量名去掉_Page后缀
				pageSqlExample := fmt.Sprintf("const %s_%s_%s = \"%s\"", dbType, structName, query.Name, pageSql)
				sqlExamples = append(sqlExamples, pageSqlExample)

				// 为分页查询生成Count SQL
				countSql := "SELECT COUNT(*) FROM " + tableName + " " + whereClause
				countSqlExample := fmt.Sprintf("const %s_%s_%s_Count = \"%s\"", dbType, structName, query.Name, countSql)
				sqlExamples = append(sqlExamples, countSqlExample)
			}
		} else {
			// 不需要分页，生成基本SQL
			sqlExample := fmt.Sprintf("const %s_%s_%s = \"%s%s\"", dbType, structName, query.Name, baseSql, sortClause)
			sqlExamples = append(sqlExamples, sqlExample)
			// 非分页查询不需要Count SQL
		}
	}

	// 如果没有自定义查询，生成一个默认查询
	if len(sqlExamples) == 0 {
		defaultSql := "SELECT * FROM " + tableName + " WHERE 1=1"
		if primaryKey != "" {
			defaultSql += " ORDER BY " + primaryKey
		}
		sqlExamples = append(sqlExamples, fmt.Sprintf("const %s_%s_Default = \"%s\"", dbType, structName, defaultSql))
		// 默认查询是非分页的，不需要生成Count SQL
	}

	// 生成初始化函数
	initFunc := fmt.Sprintf("func Init%s%s() {\n", structName, dbType)
	for _, query := range tableData.SelfQueries {
		initFunc += fmt.Sprintf("\t%sNamingSqlMap[\"%s_%s_%s\"] = %s_%s_%s\n", structName, dbType, structName, query.Name, dbType, structName, query.Name)
		// 只为分页查询添加Count SQL映射
		if query.HasPage {
			initFunc += fmt.Sprintf("\t%sNamingSqlMap[\"%s_%s_%s_Count\"] = %s_%s_%s_Count\n", structName, dbType, structName, query.Name, dbType, structName, query.Name)
		}
	}
	// 如果没有自定义查询，添加默认查询
	if len(tableData.SelfQueries) == 0 {
		initFunc += fmt.Sprintf("\t%sNamingSqlMap[\"%s_%s_Default\"] = %s_%s_Default\n", structName, dbType, structName, dbType, structName)
		// 默认查询是非分页的，不需要添加Count SQL映射
	}
	initFunc += "}\n"

	return `package dao

// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT

` + strings.Join(sqlExamples, "\n\n") + "\n\n" + initFunc
}

// getPrimaryKey 获取主键列名
func getPrimaryKey(tableData *table.TableData) string {
	for _, col := range tableData.Columns {
		if col.IsPrimaryKey {
			return col.Name
		}
	}
	return ""
}
