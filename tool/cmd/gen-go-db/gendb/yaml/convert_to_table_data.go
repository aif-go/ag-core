package yaml

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"ag-core/tool/cmd/gen-go-db/gendb/render"

	"gopkg.in/yaml.v3"
)

// ConvertToTableData 将解析后的YAML数据转换为render包中的TableData
func ConvertToTableData(dbTable DatabaseTable, selfQueryRules *yaml.Node) *render.TableData {
	tableData := &render.TableData{
		TableName:      dbTable.TableName,
		ObjectName:      ToCamelCase(dbTable.TableName),
		// PackageName:    "model",
		ColumnDataMap:  make(map[string]*render.ColumnData),
		ColumnList:     make([]*render.ColumnData, 0),
		TableModelList: make([]*render.TableModel, 0),
		GeneralIndexList: make([]*render.IndexData, 0),
		UniqueIndexList:  make([]*render.IndexData, 0),
		NamingSqlMap:   make(map[string]*render.NamingSqlData),
		NamingSqlMapEnable: true,
		PrimaryKeyList: "",
		// Engine:         "InnoDB",
		// Encode:         "utf8mb4",
		// Sort:           "utf8mb4_unicode_ci",
		DaoImportsFilterMap: sync.Map{},      // 添加 DaoImportsFilterMap
		EntityImportsFilterMap: sync.Map{},   // 添加 EntityImportsFilterMap
		DbType: dbTable.DbType,
	}

	// 转换列数据
	tableData.ColumnDataMap = make(map[string]*render.ColumnData) // 确保map已初始化
	// 处理列数据
	processCol(dbTable,tableData)
	// 处理主键数据
	processPrimaryKeys(dbTable,tableData)
	processIndex(dbTable.Indexes.General,tableData,false)
	processIndex(dbTable.Indexes.Unique,tableData,true)
	// 现在所有列的索引引用信息都已添加，创建TableModel
	for _, colName := range  dbTable.Columns.Keys(){
		colData := tableData.ColumnDataMap[strings.ToUpper(colName)]
		if colData == nil{
			fmt.Println("未找到指定的列信息:",colName)
		}
		// 使用render.CreateTableModel方法创建TableModel
		render.CreateTableModel(tableData, colData)
	}

	// 处理自定义SQL
	var waitprocess = &sync.WaitGroup{}
	waitprocess.Add(1)
	createNamingSqlData(&dbTable, tableData, waitprocess, selfQueryRules)
	waitprocess.Wait()

	return tableData
}

// ToCamelCase 将下划线命名转换为驼峰命名
func ToCamelCase(name string) string {
	// 如果没有下划线，直接将首字母大写
	if !strings.Contains(name, "_") {
		if len(name) > 0 {
			return strings.ToUpper(name[:1]) + name[1:]
		}
		return name
	}
	
	// 如果有下划线，按原来的方式处理
	parts := strings.Split(name, "_")
	for i := 0; i < len(parts); i++ {  // 修改这里，从i:=0开始而不是i:=1
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// getDbTypeFromGoType 根据Go类型推断数据库类型
func getDbTypeFromGoType(goType string) string {
	switch goType {
	case "string":
		return "varchar"
	case "int", "int32":
		return "int"
	case "int64":
		return "bigint"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "bool":
		return "tinyint"
	case "time.Time":
		return "datetime"
	default:
		return "varchar"
	}
}

// getColumnEndSymbol 获取列的结束符号
func getColumnEndSymbol(column Column) string {
	if column.NotNull {
		return "NOT NULL"
	}
	return "NULL"
}

// generateGoTag 生成Go结构体标签
func generateGoTag(colName string, column Column, colData *render.ColumnData) string {
	tags := []string{}
	
	// JSON标签，使用GoColName而不是原始的colName
	jsonName := ToCamelCase(colName)
	tags = append(tags, `json:"`+jsonName+`"`)
	
	// GORM标签
	gormTags := []string{}
	gormTags = append(gormTags, "column:"+column.DbColumn)
	
	if column.PrimaryKey {
		gormTags = append(gormTags, "primaryKey")
	}
	
	if column.NotNull {
		gormTags = append(gormTags, "not null")
	}
	
	if column.AutoIncrement {
		gormTags = append(gormTags, "autoIncrement")
	}
	
	if column.Length != "" {
		gormTags = append(gormTags, "size:"+column.Length)
	}
	
	if column.DefaultValue != "" {
		gormTags = append(gormTags, "default:"+column.DefaultValue)
	}
	
	// 处理索引信息
	if colData.ColumnRefIndexList != nil {
		for _, ref := range colData.ColumnRefIndexList {
			if ref == nil {
				continue
			}
			switch ref.IndexType {
			case render.General:
				gormTags = append(gormTags, "index:"+ref.IndexName+",priority:"+ref.Priority)
			case render.Unique:
				gormTags = append(gormTags, "uniqueIndex:"+ref.IndexName+",priority:"+ref.Priority)
			}
		}
	}
	
	tags = append(tags, `gorm:"`+strings.Join(gormTags, ";")+`"`)
	
	return strings.Join(tags, " ")
}

// createNamingSqlData 处理自定义sql
func createNamingSqlData(dbTable *DatabaseTable, tableData *render.TableData, wait *sync.WaitGroup, selfQueryRules *yaml.Node) {
	defer wait.Done()
	rnaingsqls := []*render.NamingSqlTemplate{}
	cudnaingsqls := []*render.NamingSqlTemplate{}

	// 初始化空的 NamingSqlList
	var namingSqlList []*render.NamingSqlData
	
	// 如果提供了 SelfQueryRules，则处理它们并添加到 NamingSqlList
	if selfQueryRules != nil && selfQueryRules.Kind != 0 {
		// 解析有序查询规则
		orderedRules, err := parseOrderedQueryRules(selfQueryRules)
		if err == nil {
			// 转换为 NamingSqlData
			selfQueryNamingSqlList := ConvertSelfQueryRulesToNamingSql(dbTable.TableName, orderedRules, tableData)
			
			// 将转换后的数据合并到 NamingSqlList 中
			namingSqlList = append(namingSqlList, selfQueryNamingSqlList...)
			// 添加调试信息
			fmt.Printf("Converted %d self query rules to naming SQL\n", len(selfQueryNamingSqlList))
		} else {
			fmt.Printf("Error parsing ordered query rules: %v\n", err)
		}
	} else {
		fmt.Println("No selfQueryRules provided or selfQueryRules.Kind is 0")
	}
	
	// 添加调试信息
	fmt.Printf("Total namingSqlList length: %d\n", len(namingSqlList))

	// 初始化空的 NamingsqlMap
	namingsqlMap := make(map[string]*render.NamingSqlData)

	// 构建 keyPrefix
	keyPrefix := strings.ToUpper(tableData.DbType)
	if keyPrefix == "" {
		keyPrefix = "DEFAULT"
	}

	// 处理每个自定义SQL
	for _, sqlData := range namingSqlList {
		// 非默认或者匹配的db类型不处理
		if sqlData.DbType != "" && sqlData.DbType != " " && sqlData.DbType != tableData.DbType {
			continue
		}

		template := &render.NamingSqlTemplate{}
		// 去判定sqlData是否符合要求
		// ToUpper(dbtype)+"@"+"methodName"
		key := keyPrefix + "@" + sqlData.MethodName
		// 添加调试信息
		fmt.Printf("Processing sqlData: MethodName=%s, DbType=%s, key=%s\n", sqlData.MethodName, sqlData.DbType, key)
		if value, ok := namingsqlMap[key]; ok {
			// 更新sqlData为NamingsqlMap中的值
			sqlData = value
		}

		template.MethodName = sqlData.MethodName
		template.NamingSql = sqlData.NamingSql
		template.PageCountSql = sqlData.PageCountSql
		template.SelectAllCol = sqlData.SelectAllCol

		// 处理参数列表
		list := []*render.BindParam{}
		for _, sqlParameter := range sqlData.ParamColNameList {
			bindParam := &render.BindParam{}
			// 应该根据列名去找对应的变量对应的类型
			if colData, ok := tableData.ColumnDataMap[strings.ToUpper(sqlParameter.ColName)]; !ok {
				// bindParam.GoType=colname
				// log.Printf("警告: 自定义sql:%s的条件列:%s在表中不存在，跳过该SQL模板", template.MethodName, sqlParameter.ColName)
				continue
			} else {
				if sqlParameter.IsSlice {
					// 对于Slice的主动将参数转换为切片模式
					// 是否添加tag 标签?
					bindParam.GoType = "[]" + colData.GoType
				} else {
					bindParam.GoType = colData.GoType
					// bindParam.GoColName=sqlParameter.ParameterName
					// tableData.ColumnDataMap[sqlParameter.ColName].GoColName
				}
				bindParam.GoColName = strings.Trim(sqlParameter.ParameterName,"@")
				list = append(list, bindParam)
			}
		}
		template.BindParam = list

		// 处理自定义SQL的导入包
		for _, bindParam := range template.BindParam {
			goType, pkg := render.Imports(bindParam.GoType, "")
			if pkg != "" {
				// 使用 DaoImportsFilterMap 来避免重复添加导入包
				if _, loaded := tableData.DaoImportsFilterMap.LoadOrStore(pkg, pkg); !loaded {
					tableData.DaoImports = append(tableData.DaoImports, pkg)
				}
				// 更新 BindParam 的 GoType
				bindParam.GoType = goType
			}
		}

		// *** 对于自定义查询列的处理 start  ***
		// 转换 SelectColumns
		template.SelectColumns = make([]*render.SelectColumn, len(sqlData.SelectColumns))
		for i, selectCol := range sqlData.SelectColumns {
			renderSelectCol := &render.SelectColumn{
				ColumnName: selectCol.ColumnName,
				Alias:      selectCol.Alias,
			}
			// 对于自定义的查询的列，如果能在db的原先列中找到对应的列，则复用对应列的go类型,否则默认为interface{}
			key := strings.ToUpper(selectCol.ColumnName)
			if val, ok := tableData.ColumnDataMap[key]; ok {
				renderSelectCol.GoType = val.GoType
				renderSelectCol.Tag = "gorm:\"column:" + selectCol.ColumnName + "\""
				
				// 处理 SelectColumns 的导入包
				goType, pkg := render.Imports(val.GoType, "")
				if pkg != "" {
					// 使用 DaoImportsFilterMap 来避免重复添加导入包
					if _, loaded := tableData.DaoImportsFilterMap.LoadOrStore(pkg, pkg); !loaded {
						tableData.DaoImports = append(tableData.DaoImports, pkg)
					}
					// 更新 GoType
					renderSelectCol.GoType = goType
				}
			} else {
				// TODO  gorm 无法将数据库类型的数据转换到interface{}类型的接口上,暂时先拿string接收
				renderSelectCol.GoType = selectCol.GoType
				renderSelectCol.Tag = "gorm:\"column:" + selectCol.ColumnName + "\""
			}
			template.SelectColumns[i] = renderSelectCol
		}

		if template.SelectColumns == nil {
			template.MethodResName = tableData.ModelName
		} else {
			template.MethodResName = template.MethodName + "Res"
			template.GenerateBol = true
		}
		//  *** 对于自定义查询列的处理 end  ***
		// 区分查询sql和更新sql
		sqllower := strings.ToLower(template.NamingSql)
		if strings.HasPrefix(sqllower, "select") {
			rnaingsqls = append(rnaingsqls, template)
		} else {
			cudnaingsqls = append(cudnaingsqls, template)
		}
	}

	tableData.CUDNamingSqlList = cudnaingsqls
	tableData.RNamingSqlList = rnaingsqls
	tableData.NamingSqlMap = namingsqlMap

	if len(cudnaingsqls) != 0 || len(rnaingsqls) != 0 {
		tableData.DaoImports = append(tableData.DaoImports, "errors")
	}
}

// convertNamingSqlMap 转换 YAML 的 NamingSqlMap 到 render 的 NamingSqlMap
func convertNamingSqlMap(yamlMap map[string]*NamingSqlData) map[string]*render.NamingSqlData {
	if yamlMap == nil {
		return nil
	}

	result := make(map[string]*render.NamingSqlData)
	for k, v := range yamlMap {
		// 转换 ParamColNameList
		paramList := make([]render.SqlParameter, len(v.ParamColNameList))
		for i, p := range v.ParamColNameList {
			paramList[i] = render.SqlParameter{
				ColName:       p.ColName,
				ParameterName: p.ParameterName,
				IsSlice:       p.IsSlice,
			}
		}

		// 转换 SelectColumns
		selectCols := make([]*render.SelectColumn, len(v.SelectColumns))
		for i, s := range v.SelectColumns {
			selectCols[i] = &render.SelectColumn{
				ColumnName: s.ColumnName,
				Alias:      s.Alias,
			}
		}

		result[k] = &render.NamingSqlData{
			MethodName:       v.MethodName,
			NamingSql:        v.NamingSql,
			DbType:           v.DbType,
			ParamColNameList: paramList,
			SelectColumns:    selectCols,
		}
	}
	return result
}



// 处理表列的逻辑
func processCol(dbTable DatabaseTable,tableData	*render.TableData){
		// 使用orderedmap的Keys()方法获取列名列表，以保持顺序
	columnKeys := dbTable.Columns.Keys()
	colCount := len(columnKeys)
	i := 0
		// 遍历列键来访问列数据
	for _, colName := range columnKeys {
		// colName=strings.ToUpper(colName)
		// 从orderedmap中获取列值
		columnValue, exists := dbTable.Columns.Get(colName)
		if !exists {
			continue
		}
		
		column, ok := columnValue.(Column) // 类型断言为Column类型
		if !ok {
			continue
		}
		
		// 解析tag列，处理多标签的问题
		tagParts := strings.Split(column.Description, ";")
		omitempty := false
		autoCreate := false
		autoUpdate := false
		for _, tag := range tagParts {
			if strings.TrimSpace(tag) == "///@omitempty" {
				omitempty = true
				continue
			}
			if strings.TrimSpace(tag) == "///@create" {
				autoCreate = true
				continue
			}
			if strings.TrimSpace(tag) == "///@update" {
				autoUpdate = true
				continue
			}
		}

		colData := &render.ColumnData{
			GoType:        column.GoType,
			GoColName:     ToCamelCase(strings.ToLower(colName)),
			DbColName:     column.DbColumn,
			DbType:        getDbTypeFromGoType(column.GoType),
			PrimaryKey:    column.PrimaryKey,
			NotNullFlag:   column.NotNull,
			Length:        column.Length,
			Comment:       column.Comment,
			Description:   column.Description,
			DefaultVal:    column.DefaultValue,
			AutoIncrement: column.AutoIncrement,
			EndSymbol:     ",", // 默认设置为逗号
			AutoUpdate:    autoUpdate,
			AutoCreate:    autoCreate,
			Omitempty:     omitempty,
		}

		
		// 最后一列的结束符号设置为空
		if i == colCount-1 {
			colData.EndSymbol = ""
		}
		
		tableData.ColumnDataMap[strings.ToUpper(colName)] = colData
		tableData.ColumnList = append(tableData.ColumnList, colData)
		i++
	}
}



// 处理主键的数据
func processPrimaryKeys(dbTable DatabaseTable, tableData *render.TableData){
		// 处理主键
	var primaryKeys []string
	for _, pk := range dbTable.PrimaryKeys {
		primaryKeys = append(primaryKeys, pk.Column)
		if colData, exists := tableData.ColumnDataMap[strings.ToUpper(pk.Column)]; exists {
			colData.PrimaryKey = true
		}
	}
	tableData.PrimaryKeyList = strings.Join(primaryKeys, ",")

	// 处理主键相关的索引（PrimryRIndex, PrimryUIndex, PrimryDIndex）
	if len(primaryKeys) > 0 {
		// 初始化主键索引数据
		primryDIndex := &render.IndexData{
			IndexName:     "DeleteByPrimaryKey",
			BindParamList: []*render.BindParam{},
			HashParamters: "",
		}
		primryRIndex := &render.IndexData{
			IndexName:     "FindByPrimaryKey",
			BindParamList: []*render.BindParam{},
			HashParamters: "",
		}
		primryUIndex := &render.IndexData{
			IndexName:     "UpdateByPrimaryKey",
			BindParamList: []*render.BindParam{},
			HashParamters: "",
		}

		// 构建主键参数列表
		for _, pkColName := range primaryKeys {
			if colData, exists := tableData.ColumnDataMap[strings.ToUpper(pkColName)]; exists {
				// 创建绑定参数
				bindParam := &render.BindParam{
					GoType:    colData.GoType,
					GoColName: colData.GoColName,
					DbColName: colData.DbColName,
				}

				// 处理导入包
				goType, pkg := render.Imports(colData.GoType, "")
				if pkg != "" {
					// 使用 DaoImportsFilterMap 来避免重复添加导入包
					if _, loaded := tableData.DaoImportsFilterMap.LoadOrStore(pkg, pkg); !loaded {
						tableData.DaoImports = append(tableData.DaoImports, pkg)
					}
					// 更新 BindParam 的 GoType
					bindParam.GoType = goType
				}

				// 添加到各个索引的参数列表中
				primryDIndex.BindParamList = append(primryDIndex.BindParamList, bindParam)
				primryRIndex.BindParamList = append(primryRIndex.BindParamList, bindParam)
				primryUIndex.BindParamList = append(primryUIndex.BindParamList, bindParam)

				// 构建HashParamters
				hashParam := colData.GoColName + " " + bindParam.GoType
				if primryDIndex.HashParamters == "" {
					primryDIndex.HashParamters = hashParam
				} else {
					primryDIndex.HashParamters = primryDIndex.HashParamters + "," + hashParam
				}
				
			}
		}

		// 设置主键索引数据
		primryRIndex.HashParamters = primryDIndex.HashParamters
		tableData.PrimryDIndex = primryDIndex
		tableData.PrimryRIndex = primryRIndex
		tableData.PrimryUIndex = primryUIndex
	}

}



// 处理索引数据
func processIndex(indexs []Index, tableData *render.TableData, uniqueIndex bool){
	indexTypeStr := "General"
	if uniqueIndex {
		indexTypeStr = "Unique"
	}
	// 处理索引 - 先处理索引，添加索引引用信息到列数据中
	for _, idx := range indexs {
		indexData := &render.IndexData{
			IndexName: idx.IndexName,
			IndexType: indexTypeStr,
			IndexColList: strings.Join(idx.Columns, ","),
			Use: true,
		}
		
		// 处理 BindParamList 和 HashParamters
		bindParamList := []*render.BindParam{}
		var hashParamters string
		for i, colName := range idx.Columns {
			// 通过ColumnDataMap的键查找，使用原始列名（与YAML中列定义的键一致）
			colData, exists := tableData.ColumnDataMap[strings.ToUpper(colName)]
			
			if exists {
				bindParam := &render.BindParam{
					DbColName: colData.DbColName,
					GoColName: colData.GoColName,
					GoType:    colData.GoType,
				}
				bindParamList = append(bindParamList, bindParam)
				
				// 构建HashParamters
				hashParam := colData.GoColName + " " + bindParam.GoType
				if hashParamters == "" {
					hashParamters = hashParam
				} else {
					hashParamters = hashParamters + "," + hashParam
				}
				
				// 更新列的ColumnRefIndexList，添加索引引用信息
				indexRef := &render.ColumnRefIndex{
					IndexName: idx.IndexName,
					IndexType: render.General,
					Priority: strconv.Itoa(i+1),
				}
				
				// 初始化ColumnRefIndexList如果为nil
				if colData.ColumnRefIndexList == nil {
					colData.ColumnRefIndexList = []*render.ColumnRefIndex{}
				}
				colData.ColumnRefIndexList = append(colData.ColumnRefIndexList, indexRef)
			} else {
				// 列不存在于ColumnDataMap中，跳过
			}
		}
		indexData.BindParamList = bindParamList
		indexData.HashParamters = hashParamters
		
		// 生成MethodName - 与excel_data_convert.go保持一致
		var methodNameBuilder strings.Builder
		for _, colName := range idx.Columns {
			// 通过ColumnDataMap的键查找，使用原始列名（与YAML中列定义的键一致）
			colData, exists := tableData.ColumnDataMap[strings.ToUpper(colName)]
			
			if exists {
				methodNameBuilder.WriteString(colData.GoColName)
			}
		}
		indexData.MethodName = methodNameBuilder.String()
		
		// 处理 Imports
		for _, bindParam := range indexData.BindParamList {
			goType, pkg := render.Imports(bindParam.GoType, "")
			if pkg != "" {
				// 使用 DaoImportsFilterMap 来避免重复添加导入包
				if _, loaded := tableData.DaoImportsFilterMap.LoadOrStore(pkg, pkg); !loaded {
					tableData.DaoImports = append(tableData.DaoImports, pkg)
				}
				// 更新 BindParam 的 GoType
				bindParam.GoType = goType
			}
		}
		if uniqueIndex	{
			tableData.UniqueIndexList = append(tableData.UniqueIndexList, indexData)
		}else{
			tableData.GeneralIndexList = append(tableData.GeneralIndexList, indexData)
		}
	}
}