package erm

import (
	"strconv"
	"strings"

	"ag-core/tool/cmd/gen-go-db/gendb/common"
	"ag-core/tool/cmd/gen-go-db/gendb/render"
)

var ColIdReflectMap map[string]*NormalColumn
type ErmDataConvert struct {
}

// colRefInexMap 按照erm的格式定义列和索引的关系 key为表名+列名
var colRefInexMap map[string][]*render.ColumnRefIndex
// wordToMap 将erm中的word标签部分转换为map,方便后续根据word_id获取word对象信息用
var ermColMap map[string]*NormalColumn
var wordMap map[string]*Word
// Convert 将源文件数据转换为模板文件数据
func (convert *ErmDataConvert) Convert(diagram *Diagram) []*render.TableData {

	colRefInexMap=make(map[string][]*render.ColumnRefIndex)
	wordToMap(diagram)
	list:=[]*render.TableData{}
	normalColumnToMap(diagram.Contents.Tables)

	for _, table := range diagram.Contents.Tables {
		tableData := &render.TableData {
			DbType: diagram.DbType,
			ObjectName:common.TitleCase(common.ToCamelCase(table.TableName)),
			TableName:table.TableName,
			ColumnDataMap:make(map[string]*render.ColumnData,10),
			Encode:table.TableProperties.EnCode,
			Engine:table.TableProperties.Engine,
			Sort:table.TableProperties.Sort,
			// DaoImportsFilterMap: map[string]string{},
			// EntityImportsFilterMap: map[string]string{},
			NoNameUiqueList: []string{},
		}		
		// 索引的处理
		indexProcess(table, tableData)
        collist:=[]*render.ColumnData{}
		primarylist:=[]string{}
		colcount:=len(table.Columns.NormalColumns)
		lastColEndSymbol:=false
		// 开始按照表处理数据
		for i, column := range table.Columns.NormalColumns {
			// 先判断是否引用列
			columnData := &render.ColumnData{}
			word := wordMap[column.WordId]
			// 需要进一步按照列名处理
			if word.Description == "///@create" {
				columnData.AutoCreate = true
			}
			if word.Description == "///@update" {
				columnData.AutoUpdate = true
			}
			columnData.Comment = word.LogicalName
			columnData.Description = word.Description
			columnData.DbColName = word.PhysicalName
			columnData.DefaultVal = column.DefaultValue
			key:=tableData.TableName+":"+word.PhysicalName
			columnData.ColumnRefIndexList = colRefInexMap[key]
			columnData.GoColName = common.ToCamelCase(word.PhysicalName)
			columnData.GoType = render.DbTypeConvertGoType(word.Type, word.Decimal)
			columnData.DecimalLength = word.Decimal
			columnData.Length = word.Length
			columnData.DbType = word.Type
			columnData.DbType = render.FormtDbTypeToDDL(columnData)
			columnData.NotNullFlag = column.NotNull
			columnData.PrimaryKey = column.PrimaryKey
			if columnData.PrimaryKey{
				lastColEndSymbol = true
				primarylist=append(primarylist, columnData.DbColName)
			}
			columnData.EndSymbol=","
			// 如果表包含主键那么最后一列的标识符号就要设置为""
			if i==colcount-1{
			    // 存在主键或者存在无名唯一索引的情况
				if lastColEndSymbol || len(tableData.NoNameUiqueList)!=0 {
					columnData.EndSymbol=","
				}else{
					columnData.EndSymbol=""
				}
			}
			// 构建struct 模板数据
			render.CreateTableModel(tableData,columnData)
			tableData.ColumnDataMap[columnData.DbColName]=columnData
			// 主键的处理
			primaryKeyProcess(columnData, tableData)
			columnData.AutoIncrement=column.AutoIncrement
			collist=append(collist, columnData)
		}
		// 构建主键模板数据,避免html模板中做逻辑判断和拼接数据,使模板简单化
		tableData.PrimaryKeyList = strings.Join(primarylist,",")
		// if len(tableData.NoNameUiqueList)!=0{
		// 	tableData.PrimaryKeyList = tableData.PrimaryKeyList+ ","
		// }
		// 对于erm中设置的唯一索引无名字的情况
		// processNoNameUniqueEndSymbol(tableData)
		tableData.ColumnList = collist
		list=append(list, tableData)
	}
	return list
}

// func processNoNameUniqueEndSymbol(tableData *render.TableData) {
// 	 list:=tableData.NoNameUiqueList
// 	 length:=len(list)
// 	 if length<=1{
// 		return
// 	 }
// 	 for index,noname:=range list {
// 		if index != length -1{
// 			tableData.NoNameUiqueList[index]=noname+" ,"
// 		}
// 	 }

// }

// indexProcess 处理erm每个表的索引数据
func indexProcess(table *Table, tableData *render.TableData) {
	// 索引数据
	generalIndexList := []*render.IndexData{}
	// 唯一索引数据
	uniqueIndexList := []*render.IndexData{}
	if tableData.DaoImports == nil{
		tableData.DaoImports=[]string{}
	}
	for _, index := range table.Indexs.Indexs {
		// 规避配置了索引名但是未配置索引对应的列的内容
		if index.IndexColumns.IndexColumn == nil {
			continue
		}
		indexData := &render.IndexData{}
		indexData.IndexName = index.Name
		// 区分普通索引和唯一索引
		if index.NonUnique {
			generalIndexList = append(generalIndexList, indexData)
			indexColumnProcess(render.General,tableData,indexData,index.IndexColumns.IndexColumn)
		
			// 如果设置的索引的为组合索引，此时就要拆分索引组装查询
			// if len(indexData.BindParamList)>=2 {
			// 	indexData := &render.IndexData{}
			// 	generalIndexList = append(generalIndexList, indexData)
			// }
		} else {

			uniqueIndexList = append(uniqueIndexList, indexData)
			indexColumnProcess(render.Unique,tableData,indexData,index.IndexColumns.IndexColumn)
		}
		indexData.IndexType=index.Type
	    // 标记该索引为真实索引，需要后续建表使用
		indexData.Use = true
		tableData.DaoImports=append(tableData.DaoImports, indexData.Imports...)
	}

	// erm处理复合主键用
	for _,unique:=range table.UniqueIndexes.UniqueIndexes{
	// 规避配置了索引名但是未配置索引对应的列的内容
		if unique.IndexColumns.IndexColumn == nil {
			continue
		}
		indexData := &render.IndexData{}
		indexData.IndexName = unique.Name
		// TODO 对于该类的是自动加名字还是按照UNIQUE的方式处理???
		if indexData.IndexName == ""{
			sort:=tableData.UniqueIndexSort+1
			indexData.IndexName= table.TableName+"_UNIUQE_"+strconv.Itoa(sort)
		}
		// 区分普通索引和唯一索引
		uniqueIndexList = append(uniqueIndexList, indexData)
		indexColumnProcess(render.Unique, tableData, indexData, unique.IndexColumns.IndexColumn)
	    nonameuniquecol:=[]string{}
		for _,bind:= range indexData.BindParamList{
			nonameuniquecol=append(nonameuniquecol, bind.DbColName)
		}
		tableData.NoNameUiqueList = append(tableData.NoNameUiqueList, strings.Join(nonameuniquecol,","))
		tableData.DaoImports=append(tableData.DaoImports, indexData.Imports...)
	}
	tableData.GeneralIndexList = generalIndexList
	tableData.UniqueIndexList = uniqueIndexList
}


// indexColumnProcess 索引对应的列处理
func indexColumnProcess(indexType string, tableData *render.TableData, indexData *render.IndexData, indexColList []*IndexColumn) {
	var bindParamList []*render.BindParam
	// var columnData *ColumnData = &ColumnData{}
	priority := 0
	tableIndexColNameList:=[]string{}
	methodParamterList:=[]string{}
	methodName:=&strings.Builder{}
	for _, indexColumn := range indexColList {
		priority++
		word := wordMap[ermColMap[indexColumn.ID].WordId]
		bindparam := &render.BindParam{}
		bindparam.DbColName = word.PhysicalName
		bindparam.GoColName = common.ToCamelCase(word.PhysicalName)
		// 此处要单独处理
		bindparam.GoType = render.DbTypeConvertGoType(word.Type,word.Decimal)
		// 此处需要处理Dao的Imports
		goType,pck:=render.Imports(bindparam.GoType,"")
		// bindparam.GoType = goType
		if pck!=""{
			if _,ok:=tableData.DaoImportsFilterMap.LoadOrStore(pck,pck);!ok{
				indexData.Imports=append(indexData.Imports,pck)
				// tableData.DaoImportsFilterMap[pck]=pck
			}
			bindparam.GoType = goType
		}

		bindParamList = append(bindParamList, bindparam)
		methodParamterList=append(methodParamterList, bindparam.GoColName+" "+bindparam.GoType)
		methodName.WriteString(bindparam.GoColName)
		tableIndexColNameList=append(tableIndexColNameList, word.PhysicalName)
		indexRef:=&render.ColumnRefIndex{
			IndexName: indexData.IndexName,
			IndexType: indexType,
			Priority: strconv.Itoa(priority),
		}
		key:=tableData.TableName+":"+word.PhysicalName
		// 如果map中没有对应的值，则新建切片放入数组中，key的维度是啥?
		if colRefInexMap[key]==nil{
			colRefInexMap[key]=make([]*render.ColumnRefIndex,3)
		}
		colRefInexMap[key]=append(colRefInexMap[key],indexRef)
	}
	indexData.BindParamList = bindParamList
	// 构建方法参数列表
	indexData.HashParamters=strings.Join(methodParamterList,",")
	// TODO方法名需要商议
	indexData.MethodName=methodName.String()
	indexData.IndexColList=strings.Join(tableIndexColNameList,",")
	// 对于组合索引(A,B,C)，此时要组装方法(A,B,C)（A，B） (A) 
}

// primaryKeyProcess 主键类索引处理
// 非主键列直接跳出处理
func primaryKeyProcess(col *render.ColumnData, tableData *render.TableData) {

	if !col.PrimaryKey {
		return
	}
	var primryDIndex *render.IndexData
	var primryRIndex *render.IndexData
	var primryUIndex *render.IndexData
	// var list []*render.BindParam
	if tableData.DaoImports == nil{
		tableData.DaoImports=[]string{}
	}
	if tableData.PrimryDIndex == nil {
		list := []*render.BindParam{}
		primryDIndex = &render.IndexData{
			IndexName:     "DeleteByPrimaryKey",
			BindParamList: list,
		}
		primryRIndex = &render.IndexData{IndexName: "FindByPrimaryKey",
			BindParamList: list,
		}
		primryUIndex = &render.IndexData{IndexName: "UpdateByPrimaryKey",
			BindParamList: list,
		}
		tableData.PrimryDIndex = primryDIndex
		tableData.PrimryRIndex = primryRIndex
		tableData.PrimryUIndex = primryUIndex
	}

	bindparam:=&render.BindParam{
		GoType:    col.GoType,
		GoColName: col.GoColName,
		DbColName: col.DbColName,
	}
	// 准备生成dao的时候引入第三方的的pkg
	// 此处需要处理Dao的Imports
	goType,pck:=render.Imports(col.GoType,"")
	if pck!=""{
		if _,ok:=tableData.DaoImportsFilterMap.LoadOrStore(pck,pck);!ok{
			tableData.DaoImports=append(tableData.DaoImports,pck)
			// tableData.DaoImportsFilterMap[pck]=pck
		}
		bindparam.GoType = goType
	}

	tableData.PrimryDIndex.BindParamList = append(tableData.PrimryDIndex.BindParamList, bindparam)
	if tableData.PrimryDIndex.HashParamters == "" {
		tableData.PrimryDIndex.HashParamters=col.GoColName+" "+col.GoType
	}else{
		tableData.PrimryDIndex.HashParamters=tableData.PrimryDIndex.HashParamters+","+col.GoColName+" "+col.GoType
	}
	tableData.PrimryRIndex.BindParamList = tableData.PrimryDIndex.BindParamList
	tableData.PrimryRIndex.HashParamters = tableData.PrimryDIndex.HashParamters
	tableData.PrimryUIndex.BindParamList = tableData.PrimryDIndex.BindParamList
	tableData.PrimryUIndex.HashParamters = tableData.PrimryDIndex.HashParamters
}


func wordToMap(diagram *Diagram) {

	wordMap = make(map[string]*Word,20)
	for _, word := range diagram.Dictionary.Words {
		wordMap[word.ID] = word
	}

	ermColMap=make(map[string]*NormalColumn,500)
	for _, ermTable:= range diagram.Contents.Tables{
		for _,ermCol:= range ermTable.Columns.NormalColumns{
			ermColMap[ermCol.ID]=ermCol
		}
	}

	// 处理列和列之间相互引用的问题
	for _,value:=range ermColMap{
		if value.ReferencedColumn!="" && value.WordId==""{
			value.WordId=ermColMap[value.ReferencedColumn].WordId
			value.Description = wordMap[value.WordId].Description
		}
	}
}

// normalColumnToMap 由于erm的同名列之间存在引用的关系
// 可以将erm中的列先做map维护，为后面根据id查找使用方便
func normalColumnToMap(tableSlice []*Table){
	ColIdReflectMap = make(map[string]*NormalColumn,20)
	// 每个表都有列，要过滤每个表的映射ID和列的关系
	for _,table:=range tableSlice{
		for _,normalCol:= range table.Columns.NormalColumns{
			if normalCol.WordId!=""{
				ColIdReflectMap[normalCol.ID] = normalCol
			}
		}
	}
}





