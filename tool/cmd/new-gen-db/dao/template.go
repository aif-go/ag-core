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
	UpdateByPrimarykey(ctx context.Context, entity *model.` + structName + `) (int64, error)
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
func (dao *` + structName + `Dao) UpdateByPrimarykey(ctx context.Context, entity *model.` + structName + `) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	// 使用支持更新的列
	db.UpdateColumns(model.` + structName + `AllowUpdateCols)
	result := db.Save(entity)
	return result.RowsAffected, result.Error
}

// UpdateByPriIngoreNullCols 根据主键更新，自动剔除参数中的零值列
func (dao *` + structName + `Dao) UpdaeByPrimaryKeyIngoreZeroValCols(ctx context.Context, entity *model.` + structName + `) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	// 使用支持更新的列
	result := db.Model(entity).UpdateColumns(model.` + structName + `AllowUpdateCols).Updates(entity)
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
func GetNamingSqlTemplate(tableData *table.TableData) string {
	structName := tableData.StructName
	return `package dao

// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT

func Init` + structName + `NamingSql() {
	// 执行一次初始化操作
	Init` + structName + `MYSQL()
	// 执行一次初始化操作
	Init` + structName + `DB2()
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
		doMethods += fmt.Sprintf(`// do%s 执行%s查询
func (dao *%sDao) do%s(ctx context.Context, namingInfo *db.NameingSqlArgInfo, args any) ([]*model.%s%sRes, error) {

	queryArgs, ok := args.(model.%s%sArg)
	if !ok {
		return nil, errors.New("do%s args type not match")
	}

	execSql := %sNamingSqlMap[namingInfo.SqlName]
	if execSql == "" {
		return nil, errors.New("not found naming sql")
	}
	newTableName := dao.getApplyInfo(ctx).TableName
	if newTableName != "" {
		enity := &model.%s{}
		execSql = strings.ReplaceAll(execSql, "FROM "+enity.TableName()+" WHERE", "FROM "+newTableName+" WHERE")
	}

	argsMap := queryArgs.ConvertToMap()
	var list []*model.%s%sRes
	result := dao.DB(ctx).Raw(execSql, argsMap).Find(&list)
	if result.Error != nil {
		return nil, result.Error
	}
	return list, nil
}

`, query.Name, query.Name, structName, query.Name, structName, query.Name, structName, query.Name, query.Name, structName, structName, structName, query.Name)
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
	constants += fmt.Sprintf(`// exclude%sZeroColNames 插入忽略空值时标记哪些字段需要排除在外
var exclude%sZeroColNames = map[string]int{"CreatedTime": 0, "LastModifiedTime": 0}

`, structName, structName)
	
	// 只有当有自定义查询时才生成命名SQL参数信息
	if len(tableData.SelfQueries) > 0 {
		for _, query := range tableData.SelfQueries {
			constants += fmt.Sprintf(`var %sNamingInfo = &db.NameingSqlArgInfo{
	SqlName:  "%s",
	ReqType:  (*model.%s%sArg)(nil),
	RespType: ([]*model.%s%sRes)(nil),
}

`, query.Name, query.Name, structName, query.Name, structName, query.Name)
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
	// moduleName := tableData.ModuleName

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

		// 组合SQL语句
		sql := selectClause + " FROM " + tableName + " " + whereClause
		sqlExample := fmt.Sprintf("const %s_%s_%s = \"%s\"", dbType, structName, query.Name, sql)
		sqlExamples = append(sqlExamples, sqlExample)
	}

	// 如果没有自定义查询，生成一个默认查询
	if len(sqlExamples) == 0 {
		sqlExamples = append(sqlExamples, fmt.Sprintf("const %s_%s_Default = \"SELECT * FROM %s WHERE 1=1\"", dbType, structName, tableName))
	}

	// 生成初始化函数
	initFunc := fmt.Sprintf("func Init%s%s() {\n", structName, dbType)
	for _, query := range tableData.SelfQueries {
		initFunc += fmt.Sprintf("\t%sNamingSqlMap[\"%s_%s_%s\"] = %s_%s_%s\n", structName, dbType, structName, query.Name, dbType, structName, query.Name)
		// 设置返回对象类型
		resultType := structName
		if query.SelectFields != "" && query.SelectFields != "*" {
			resultType += query.Name + "Res"
		}
		// initFunc += fmt.Sprintf("\t%sNamingSqlMethodMap[\"%s_%s_%s\"] = &db.NamingSqlMethod{DbResultObjName: []interface{}{&model.%s{}}}\n", structName, dbType, structName, query.Name, resultType)
	}
	// 如果没有自定义查询，添加默认查询
	if len(tableData.SelfQueries) == 0 {
		initFunc += fmt.Sprintf("\t%sNamingSqlMap[\"%s_%s_Default\"] = %s_%s_Default\n", structName, dbType, structName, dbType, structName)
		// initFunc += fmt.Sprintf("\t%sNamingSqlMethodMap[\"%s_%s_Default\"] = &db.NamingSqlMethod{DbResultObjName: []interface{}{&model.%s{}}}\n", structName, dbType, structName, structName)
	}
	initFunc += "}\n"

	return `package dao

// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT

` + strings.Join(sqlExamples, "\n\n") + "\n\n" + initFunc
}
