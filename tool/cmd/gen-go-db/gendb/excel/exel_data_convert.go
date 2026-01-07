package excel

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"ag-core/tool/cmd/gen-go-db/gendb/common"
	"ag-core/tool/cmd/gen-go-db/gendb/render"
)

type ExcelDataConvert struct {
}

// Convert 将源文件数据转换为模板文件数据
func (convert *ExcelDataConvert) Convert(excelDataList [] *ExcelData) []*render.TableData {

	list := []*render.TableData{}
	for _,excelData:= range excelDataList{
		tableData := &render.TableData{
			TableName: excelData.TableName,
			ObjectName: common.ToCamelCase(excelData.TableName),		
			// DaoImportsFilterMap:map[string]string{},
			// EntityImportsFilterMap:map[string]string{},
			Encode: excelData.Encode,
			Engine:excelData.Engine,
			Sort: excelData.Sort,
			DbType: excelData.DbType,
		}

		colMap:= make(map[string]*render.ColumnData,20)
		convertToMap(colMap,excelData)
		tableData.ColumnDataMap=colMap
		// TableData.PackageName = "model"
		var waitprocess=&sync.WaitGroup{}
		waitprocess.Add(4)
		// 处理索引的数据
		createIndexData(excelData.GeneralIndexList,tableData,General,waitprocess)
		createIndexData(excelData.UniqueIndexList,tableData,Unique,waitprocess)
		// 构建主键数据
		createPrimaryData(excelData.PrimaryKeyList,tableData,waitprocess)
		createNamingSqlData(excelData,tableData,waitprocess)
		waitprocess.Wait()
		log.Println("create model template data")
		collist:=[]*render.ColumnData{}
		// 处理model的数据
		for _,coldata:=range excelData.ColumnList{
			render.CreateTableModel(tableData,coldata)
			collist=append(collist, coldata)
		}
		// createTableModel(yamlData,tableData)
		tableData.ColumnList=collist
		list=append(list, tableData)
	}
	
	return list
}

// convertToMap 将表的列转换为Map
func convertToMap(colMap map[string]*render.ColumnData, excelData *ExcelData){
	for _, coldata:=range excelData.ColumnList{
		colMap[coldata.DbColName]=coldata
		// 转驼峰处理
		coldata.GoColName=common.ToCamelCase(coldata.DbColName)
	}
}

// createNamingSqlData 处理自定义sql
func createNamingSqlData(excelData *ExcelData, tableData *render.TableData, wait *sync.WaitGroup){

	defer wait.Done()
	rnaingsqls:=[]*render.NamingSqlTemplate{}
	cudnaingsqls:=[]*render.NamingSqlTemplate{}

	keyPrefix:=strings.ToUpper(tableData.DbType)
	for _,sqlData:=range excelData.NamingSqlList {
		// 非默认或者匹配的db类型不处理
		if sqlData.DbType!="" && sqlData.DbType != tableData.DbType{
			continue
		}
		template:=&render.NamingSqlTemplate{}
		// 去判定sqlData是否符合要求
		// ToUpper(dbtype)+@+"methodName"
		key:=keyPrefix+"@"+sqlData.MethodName
		if value,ok:=excelData.NamingsqlMap[key]; ok{
			sqlData = value
		}
		template.MethodName=sqlData.MethodName
		template.NamingSql=sqlData.NamingSql
		list:=[]*render.BindParam{};
		for _,sqlParameter:=range sqlData.ParamColNameList {
				bindParam:=&render.BindParam{}
				// 应该根据列名去找对应的变量对应的类型
				if colData,ok:=tableData.ColumnDataMap[sqlParameter.ColName];!ok{
					// bindParam.GoType=colname
					log.Printf("警告: 自定义sql:%s的条件列:%s在表中不存在，跳过该SQL模板", template.MethodName, sqlParameter.ColName)
					continue
				} else {
					if sqlParameter.IsSlice {
						// 对于Slice的主动将参数转换为切片模式
						// 是否添加tag 标签?
						bindParam.GoType="[]"+colData.GoType
					} else {
						bindParam.GoType=colData.GoType
						// bindParam.GoColName=sqlParameter.ParameterName
						// tableData.ColumnDataMap[sqlParameter.ColName].GoColName
					}
					bindParam.GoColName=sqlParameter.ParameterName
					list=append(list, bindParam)
				}
			}
		template.BindParam=list
		// *** 对于自定义查询列的处理 start  *** 
		template.SelectColumns = sqlData.SelectColumns
		for _,selectCol:= range template.SelectColumns{
			// 对于自定义的查询的列，如果能在db的原先列中找到对应的列，则复用对应列的go类型,否则默认为interface{}
			key:=strings.ToUpper(selectCol.ColumnName)
			if val,ok:=tableData.ColumnDataMap[key]; ok {
				selectCol.GoType = val.GoType
				selectCol.Tag = "gorm:\"column:"+selectCol.ColumnName+"\""
			} else {
				//TODO  gorm 无法将数据库类型的数据转换到interface{}类型的接口上,暂时先拿string接收
				selectCol.GoType = "string"
			}
		}
		if template.SelectColumns == nil {
			template.MethodResName = tableData.ObjectName
		}else{
			template.MethodResName = template.MethodName+"Res"
			template.GenerateBol = true
		}
		//  *** 对于自定义查询列的处理 end  *** 
		sqllower:=strings.ToLower(template.NamingSql)
		// 区分查询sql和更新sql
		if strings.HasPrefix(sqllower,"select") {
			rnaingsqls=append(rnaingsqls, template)
		}else{
			cudnaingsqls=append(cudnaingsqls, template)
		}
	}
	tableData.CUDNamingSqlList=cudnaingsqls
	tableData.RNamingSqlList=rnaingsqls
	tableData.NamingSqlMap = excelData.NamingsqlMap
	if len(cudnaingsqls)!=0 || len(rnaingsqls) !=0{
		tableData.DaoImports = append(tableData.DaoImports, "errors")
	}
}

// createPrimaryData 构建按照主键操作相关的数据
func createPrimaryData(primarykeys []string, tableData *render.TableData, wait *sync.WaitGroup){
	defer wait.Done()
	tempArr:=[]string{}
	list:=[]*render.BindParam{}
	allPrimaryKey:=[]string{}
	// primaryIndexList:=[]*IndexData{}
	indexData:=&render.IndexData{IndexName: "FindByPrimaryKey"}
	for _,colname:=range primarykeys{
			coldata:=tableData.ColumnDataMap[colname]
			tempArr=append(tempArr,coldata.GoColName+" "+coldata.GoType)
            bindParam :=&render.BindParam{}
			bindParam.DbColName=coldata.DbColName
			bindParam.GoColName=coldata.GoColName
			list=append(list, bindParam)
			allPrimaryKey=append(allPrimaryKey, colname)
			goType,pck:=render.Imports(coldata.GoType,coldata.Description)
			if pck!=""{
				if _,ok:=tableData.DaoImportsFilterMap.LoadOrStore(pck,pck);!ok{
					tableData.DaoImports=append(tableData.DaoImports, pck)
				}
				bindParam.GoType = goType
				coldata.GoType = goType
			}
	}
	indexData.BindParamList=list
	indexData.HashParamters=strings.Join(tempArr,", ")
	// 构建DDL语句用
	tableData.PrimaryKeyList = strings.Join(allPrimaryKey,",")
	tableData.PrimryRIndex=indexData
	// 按照主键删除记录
	deleteIndexData:=&render.IndexData{IndexName: "DeleteByPrimaryKey"}
	deleteIndexData.BindParamList=indexData.BindParamList
	deleteIndexData.HashParamters=indexData.HashParamters
	// primaryIndexList=append(primaryIndexList, deleteIndexData)
	tableData.PrimryDIndex=deleteIndexData
}

// createIndexData 构建索引数据
func createIndexData(indexDataList []*render.IndexData, tableData *render.TableData, indexType IndexType, wait *sync.WaitGroup){
	defer wait.Done()
	for _, indexData:=range indexDataList{
		tempArr:=[]string{}
		builder:=&strings.Builder{}
		indexAllColName:=[]string{}
		for index, bindParam:= range indexData.BindParamList{
			coldata:=tableData.ColumnDataMap[bindParam.DbColName]
			bindParam.GoType = coldata.GoType
			goType,pck:=render.Imports(coldata.GoType,coldata.Description)
			if pck!=""{
				if _,ok:=tableData.DaoImportsFilterMap.LoadOrStore(pck,pck);!ok{
					tableData.DaoImports=append(tableData.DaoImports, pck)
					bindParam.GoType = goType
					coldata.GoType = goType
				}
			}
			indexAllColName=append(indexAllColName, coldata.DbColName)
			// coldata.Priority=strconv.Itoa(index+1)
			var indexRefList []*render.ColumnRefIndex
			if indexRefList=coldata.ColumnRefIndexList; indexRefList==nil{
				indexRefList = make([]*render.ColumnRefIndex, 2)
			}
			switch indexType {
			case General:
				indexRef:=&render.ColumnRefIndex{
					IndexName: indexData.IndexName,
					IndexType: render.General,
					Priority: strconv.Itoa(index+1),
				}
				indexRefList= append(indexRefList, indexRef)
				// coldata.GeneralIndexName=indexData.IndexName
			case Unique:
				indexRef:=&render.ColumnRefIndex{
					IndexName: indexData.IndexName,
					IndexType: render.Unique,
					Priority: strconv.Itoa(index+1),
				}
				indexRefList= append(indexRefList, indexRef)
				// coldata.UniqueIndexName=indexData.IndexName
			default:
			}
			coldata.ColumnRefIndexList = indexRefList
			builder.WriteString(coldata.GoColName)
			tempArr=append(tempArr,coldata.GoColName+" "+coldata.GoType)
			bindParam.GoColName=coldata.GoColName
		}
		// 构建索引列表参数
		indexData.HashParamters=strings.Join(tempArr,", ")
		indexData.MethodName=builder.String()
		indexData.IndexColList=strings.Join(indexAllColName,",")
		// 构建 gorm 自动化拼接sql的参数
	}
	switch indexType {
		case General:
			tableData.GeneralIndexList=indexDataList
    	case Unique:
			tableData.UniqueIndexList=indexDataList
		default:
	}
}

