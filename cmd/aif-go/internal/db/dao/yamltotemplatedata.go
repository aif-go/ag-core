package dao

import (
	"log"
	"strconv"
	"strings"
	"sync"
)

type IndexType int

const (
	General IndexType = iota
	Unique
)

func YamlDataToTemplate(yamlData *YamlData) *TableData {
	tableData := &TableData{}
	tableData.SchemaName = yamlData.SchemaName
	tableData.TableName = yamlData.TableName
	tableData.ModuleName = yamlData.ModuleName
	tableData.ObjectName = ToCamelCase(tableData.TableName)
	colMap := make(map[string]*ColumnData, 20)
	convertToMap(colMap, yamlData)
	tableData.ColumnDataMap = colMap
	// TableData.PackageName = "model"
	var waitprocess = &sync.WaitGroup{}
	waitprocess.Add(4)
	// 处理索引的数据
	go createIndexData(yamlData.GeneralIndexList, tableData, General, waitprocess)
	go createIndexData(yamlData.UniqueIndexList, tableData, Unique, waitprocess)
	// 构建主键数据
	go createPrimaryData(yamlData.PrimaryKeyList, tableData, waitprocess)
	go createNamingSqlData(yamlData, tableData, waitprocess)
	waitprocess.Wait()
	log.Println("create model template data")
	// 处理model的数据
	createTableModel(yamlData, tableData)
	return tableData
}

func convertToMap(colMap map[string]*ColumnData, yamlData *YamlData) {
	for _, coldata := range yamlData.ColumnList {
		colMap[coldata.DbColName] = coldata
		// 转驼峰处理
		coldata.GoColName = ToCamelCase(coldata.DbColName)
	}
}

// 处理自定义sql
func createNamingSqlData(yamlData *YamlData, tableData *TableData, wait *sync.WaitGroup) {

	defer wait.Done()
	rnaingsqls := []*NamingSqlTemplate{}
	cudnaingsqls := []*NamingSqlTemplate{}

	for _, sqlData := range yamlData.NamingSqlList {
		template := &NamingSqlTemplate{}
		template.MethodName = sqlData.MethodName
		template.NamingSql = sqlData.NamingSql
		list := []*BindParam{}
		for _, colname := range sqlData.ParamColNameList {
			bindParam := &BindParam{}
			bindParam.GoColName = tableData.ColumnDataMap[colname].GoColName
			bindParam.GoType = tableData.ColumnDataMap[colname].GoType
			list = append(list, bindParam)
		}
		template.BindParam = list

		sqllower := strings.ToLower(template.NamingSql)
		// 区分查询sql和更新sql
		if strings.HasPrefix(sqllower, "select") {
			rnaingsqls = append(rnaingsqls, template)
		} else {
			cudnaingsqls = append(cudnaingsqls, template)
		}
	}
	tableData.CUDNamingSqlList = cudnaingsqls
	tableData.RNamingSqlList = rnaingsqls
}

// 构建按照主键操作相关的数据
func createPrimaryData(primarykeys []string, tableData *TableData, wait *sync.WaitGroup) {
	defer wait.Done()
	tempArr := []string{}
	list := []*BindParam{}
	// primaryIndexList:=[]*IndexData{}
	indexData := &IndexData{IndexName: "FindByPrimaryKey"}
	for _, colname := range primarykeys {
		coldata := tableData.ColumnDataMap[colname]
		tempArr = append(tempArr, coldata.GoColName+" "+coldata.GoType)
		bindParam := &BindParam{}
		bindParam.DbColName = coldata.DbColName
		bindParam.GoColName = coldata.GoColName
		list = append(list, bindParam)
	}
	indexData.BindParamList = list
	indexData.HashParameters = strings.Join(tempArr, ",")
	tableData.PrimaryRIndex = indexData
	// 按照主键删除记录
	deleteIndexData := &IndexData{IndexName: "DeleteByPrimaryKey"}
	deleteIndexData.BindParamList = indexData.BindParamList
	deleteIndexData.HashParameters = indexData.HashParameters
	// primaryIndexList=append(primaryIndexList, deleteIndexData)
	tableData.PrimaryDIndex = deleteIndexData
}

// 构建 table model
func createTableModel(yamlData *YamlData, tableData *TableData) {

	importMap := make(map[string]string, 5)
	imports := []string{}
	tableModelList := []*TableModel{}
	tableModelMap := make(map[string]*TableModel, 20)
	goTypeMap := make(map[string]string, 10)
	for _, columnData := range yamlData.ColumnList {
		tableModel := &TableModel{}
		tableModel.DbColName = columnData.DbColName
		tableModel.GoColName = columnData.GoColName
		tableModel.GoType = columnData.GoType
		if _, ok := importMap[tableModel.GoType]; !ok {
			// 目前只要time需要额外的添加
			if strings.Contains(tableModel.GoType, "time") {
				// map中不存在的场景才需要放入
				if _, ok := goTypeMap["time"]; !ok {
					// 标记用
					goTypeMap["time"] = "time"
					imports = append(imports, "time")
				}
			}
		}
		tableModel.GoTag = createTag(columnData)
		tableModelList = append(tableModelList, tableModel)
		tableModelMap[tableModel.DbColName] = tableModel
	}
	tableData.TableModelList = tableModelList
	tableData.Imports = imports
}

func createIndexData(indexDataList []*IndexData, tableData *TableData, indexType IndexType, wait *sync.WaitGroup) {
	defer wait.Done()
	for _, indexData := range indexDataList {
		tempArr := []string{}
		builder := &strings.Builder{}
		for index, bindParam := range indexData.BindParamList {
			coldata := tableData.ColumnDataMap[bindParam.DbColName]
			coldata.Priority = strconv.Itoa(index + 1)
			switch indexType {
			case General:
				coldata.GeneralIndexName = indexData.IndexName
			case Unique:
				coldata.UniqueIndexName = indexData.IndexName
			default:
			}
			builder.WriteString(coldata.GoColName)
			tempArr = append(tempArr, coldata.GoColName+" "+coldata.GoType)
			bindParam.GoColName = coldata.GoColName
		}
		// 构建索引列表参数
		indexData.HashParameters = strings.Join(tempArr, ",")
		indexData.MethodName = builder.String()
		// 构建 gorm 自动化拼接sql的参数
	}
	switch indexType {
	case General:
		tableData.GeneralIndexList = indexDataList
	case Unique:
		tableData.UniqueIndexList = indexDataList
	default:
	}
}

// 构建model的tag数据
func createTag(columnData *ColumnData) string {

	builder := strings.Builder{}

	builder.WriteString(`gorm:"`)
	if columnData.AutoCreate {
		builder.WriteString("AUTOCREATETIME;")
	}
	if columnData.AutoUpdate {
		builder.WriteString("AUTOUPDATETIME;")
	}
	builder.WriteString(`column:`)

	builder.WriteString(columnData.DbColName)
	if columnData.DefaultVal != "" {
		builder.WriteString(";default:")
		builder.WriteString(columnData.DefaultVal)
	}
	if columnData.PrimaryKey {
		builder.WriteString(";primaryKey")
	}

	if columnData.NotNullFlag {
		builder.WriteString(";not null")
	}

	if columnData.UniqueIndexName != "" {
		builder.WriteString(";uniqueIndex:")
		builder.WriteString(columnData.UniqueIndexName)
		if columnData.Priority != "" {
			builder.WriteString(",priority:")
			builder.WriteString(columnData.Priority)
		}
	}

	if columnData.GeneralIndexName != "" {
		builder.WriteString(";index:")
		builder.WriteString(columnData.GeneralIndexName)
		if columnData.Priority != "" {
			builder.WriteString(",priority:")
			builder.WriteString(columnData.Priority)
		}
	}

	builder.WriteString(`"`)
	builder.WriteString(` json:"`)
	builder.WriteString(columnData.GoColName)
	builder.WriteString(`"`)

	return builder.String()
}
