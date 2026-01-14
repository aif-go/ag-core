package excel

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"ag-core/tool/cmd/gen-go-db/gendb/common"
	"ag-core/tool/cmd/gen-go-db/gendb/render"

	"github.com/tealeg/xlsx"
)

type ExcelGenerate struct {
}

type IndexType int

const (
	General IndexType = iota
	Unique
)

// ParseTemplateFile 解析模板文件
func (generate *ExcelGenerate) ParseTemplateFile(config *render.AGInfraStructrueConfig) (*ExcelAllData, error) {

	excelFile, error := xlsx.OpenFile(config.DbTemplatePath)
	if error != nil {
		return nil, fmt.Errorf("打开资源文件失败: %w", error)
	}

	sheets := excelFile.Sheets
	wait := &sync.WaitGroup{}
	wait.Add(len(sheets))

	list := []*ExcelData{}
	for _, sheet := range sheets {
		exelData := &ExcelData{
			DbType: strings.ToUpper(config.DbType),
		}
		list = append(list, exelData)
		// 多协程并行处理文件中的sheet数据
		go processSheet(sheet, wait, exelData)
	}
	wait.Wait()
	// close(yamlCh)
	log.Println("all yaml file has create over!!!")
	return &ExcelAllData{
		ExcelDataList: list,
	}, nil
}

// Generate 构建dao,entity,sql文件的方法
func (generate *ExcelGenerate) Generate(config *render.AGInfraStructrueConfig, excelAllData *ExcelAllData) error {
	dataConvert := ExcelDataConvert{}
	tableDataList := dataConvert.Convert(excelAllData.ExcelDataList)
	supportTableSize := len(config.SupportTables)

	// 过滤需要生成的表
	filteredTables := make([]*render.TableData, 0, len(tableDataList))
	for _, tableData := range tableDataList {
		if _, ok := config.SupportTables[tableData.TableName]; !ok && supportTableSize != 0 {
			continue
		}
		tableData.ModelName = config.PackageNamePrefix
		tableData.DbType = strings.ToUpper(config.DbType)
		if len(tableData.NamingSqlMap) != 0 {
			tableData.NamingSqlMapEnable = true
		}
		filteredTables = append(filteredTables, tableData)
	}

	// 使用并发方式生成文件
	var wg sync.WaitGroup
	errCh := make(chan error, len(filteredTables))

	for _, tableData := range filteredTables {
		wg.Add(1)
		go func(data *render.TableData) {
			defer wg.Done()

			// 构建yaml文件
			// if err := render.RenderYaml(config, data); err != nil {
			// 	errCh <- err
			// 	return
			// }

			// 构建struct.go文件
			if config.Entityable {
				if err := render.RenderEnityTemplate(config.EntityPath, data); err != nil {
					errCh <- err
					return
				}
			}

			// 构建dao.go文件
			if config.Daoable {
				if err := render.RenderDaoTemplate(config.DaoPath, data); err != nil {
					errCh <- err
					return
				}
				if data.NamingSqlMapEnable {
					if err := render.RenderNamingSqlConstant(config, data); err != nil {
						errCh <- err
						return
					}
					if err := render.RenderTableNamingSql(config, data); err != nil {
						errCh <- err
						return
					}
				}
			}

			// 构建sql文件
			if config.Sqlable {
				if err := render.RenerDDLTemplate(config, data); err != nil {
					errCh <- err
					return
				}
			}
		}(tableData)
	}

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// 检查是否有错误发生
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// processSheet 按sheet处理excel,类似表级操作
func processSheet(sheet *xlsx.Sheet, wait *sync.WaitGroup, excelData *ExcelData) {
	defer wait.Done()
	// schemaName := conf.SchemaName
	sheetName := sheet.Name
	// 目标文件名不为空,做过滤文件处理
	log.Println("start process sheetName:", sheetName)
	// schemaData := ExcelData{SchemaName: schemaName}
	excelData.TableName = sheetName
	// schemaData.ModelName = conf.ModelName
	rows := sheet.Rows
	var elementMap map[string]int = make(map[string]int, 6)
	for rownum, row := range rows {
		// log.Println("当前行数:",rownum)
		for _, cell := range row.Cells {
			value := cell.Value
			// 处理表级别的定义数据，格式为在第一行，不能调整
			if rownum == 1 {
				excelData.Encode = row.Cells[2].Value
				excelData.Engine = row.Cells[0].Value
				excelData.Sort = row.Cells[1].Value
				continue
			}
			if value == "主键" {
				elementMap["primary"] = rownum
				break
			}
			if value == "索引" {
				elementMap["index"] = rownum
				break
			}
			if value == "约束" {
				elementMap["constraint"] = rownum
				break
			}
			if value == "自定义脚本名字" {
				elementMap["namingsql"] = rownum
				goto process
			}
			break
		}
	}
process:
	colInfoRowsArr := rows[3:elementMap["primary"]]
	primarkeyRowsArr := rows[elementMap["primary"]+1 : elementMap["constraint"]]
	constraintRowsArr := rows[elementMap["constraint"]+1 : elementMap["index"]]
	indexRowsArr := rows[elementMap["index"]+1 : elementMap["namingsql"]]
	namingsqlRowsArr := rows[elementMap["namingsql"]+1:]
	// 开始按照拆分后的数据row处理对应的数据
	processColRow(excelData, colInfoRowsArr)
	processPrimaryRow(excelData, primarkeyRowsArr)
	processIndexRow(excelData, constraintRowsArr, Unique)
	processIndexRow(excelData, indexRowsArr, General)
	processNamingSqlRow(excelData, namingsqlRowsArr)
	//   CreateYaml(&schemaData,conf.OutputPath)
}

// func CreateYaml(yamlTemplate *YamlTemplate, outpath string) {
// 	// key 为表名  value为yaml的数据
// 	key := yamlTemplate.TableName
// 	yamldata, err := yaml.Marshal(yamlTemplate)
// 	if err != nil {
// 		log.Println(key, "将excel数据转换为yaml格式数据")
// 	}
// 	os.WriteFile(outpath+key+".yaml", yamldata, 0755)
// }

// processColRow 处理列数据
func processColRow(excelData *ExcelData, rows []*xlsx.Row) {
	list := []*render.ColumnData{}
	length := len(rows)
	primary := false
	for rowi, row := range rows {
		if len(row.Cells) == 0 {
			break
		}
		col := &render.ColumnData{}
		for index, cel := range row.Cells {
			value := cel.Value
			// 第一单个格为空的时候,跳出对当前行的处理
			if index == 0 && value == "" {
				break
			}
			switch index {
			case 0:
				col.DbColName = strings.ToUpper(value) // 所有的列名必须大写
			case 1:
				// DB的类型
				// goType = value
				col.GoType = value
			case 2:
				col.Length = value
			case 3:
				if strings.EqualFold(value, "Y") {
					col.NotNullFlag = true
				}
			case 4:
				col.DefaultVal = value
			case 5:
				if strings.EqualFold(value, "Y") {
					col.PrimaryKey = true
					primary = true
				}
			case 6:
				// 自动增长
				if value == "Y" {
					col.AutoIncrement = true
				}
			case 7:
				// 列中文描述
				col.Comment = value
			case 8:
				// 自定义类型处理
				col.Description = value
				if strings.HasPrefix(value, "///@create") {
					col.AutoCreate = true
				}
				if strings.HasPrefix(value, "///@update") {
					col.AutoUpdate = true
				}
			default:
				log.Println("非目标列,不做任何处理")
			}

		}
		col.EndSymbol = ","
		// -2的原因是excel模板中每个模块最后必须保留一个空白行
		if rowi == length-2 && !primary {
			col.EndSymbol = ""
		}
		// 设置db类型
		col.DbType = render.ConvertGoTypeToDbType(col.GoType, col.Length, excelData.DbType)
		// if col.Length!=""{
		// 	col.DbType = dbType+"("+col.Length+")"
		// }
		// col.GoType = render.ConvertDbTypeToGoType(dbType, col.Length)
		if col.DbColName != "" {
			list = append(list, col)
		}
	}
	excelData.ColumnList = list
}

// processPrimaryRow 处理主键行数据
func processPrimaryRow(table *ExcelData, rows []*xlsx.Row) {
	primarkeys := []string{}
	for _, row := range rows {
		if len(row.Cells) == 0 {
			break
		}
		for index, cel := range row.Cells {
			value := cel.Value
			// log.Println("当前主键的值",value)
			if index == 0 {
				continue
			}
			if value == "" {
				break
			}
			primarkeys = append(primarkeys, value)
		}
	}
	table.PrimaryKeyList = primarkeys
}

func processIndexRow(table *ExcelData, rows []*xlsx.Row, indexType IndexType) {
	list := []*render.IndexData{}
	for _, row := range rows {
		if len(row.Cells) == 0 {
			break
		}
		var indexData *render.IndexData
		var bindParam []*render.BindParam
		// indexData.BindParamList=bindParam
		var value string
		for index, cel := range row.Cells {
			value = cel.Value

			if index == 0 && value == "" {
				break
			}
			if index == 0 {
				indexData = &render.IndexData{IndexName: value}
				bindParam = []*render.BindParam{}
				continue
			}

			// 处理索引类型
			if index == 1 && indexType == General {
				indexData.IndexType = value
				continue
			}
			// log.Println("索引的名字:",indexData.IndexName)
			bindParam = append(bindParam, &render.BindParam{
				DbColName: value,
			})
		}
		if indexData != nil {
			indexData.BindParamList = bindParam
			list = append(list, indexData)
		}
	}
	switch indexType {
	case General:
		table.GeneralIndexList = list
	case Unique:
		table.UniqueIndexList = list
	default:
	}
}

func processNamingSqlRow(table *ExcelData, rows []*xlsx.Row) {

	//map中取，没有的化就用默认的
	namingsqlArr := []*render.NamingSqlData{}
	// key 是db类型+方法名
	namingsqlMap := map[string]*render.NamingSqlData{}
	for _, row := range rows {
		// 空白行不处理
		if len(row.Cells) == 0 {
			break
		}
		var namingsql *render.NamingSqlData
		add := true
		// 先将每行数据转换为对应的实体
		for index, cel := range row.Cells {
			value := cel.Value
			switch index {
			case 0: // methodName
				// 如果自定义命名的方法名为空,直接丢弃该条数据
				if value == "" {
					add = false
					continue
				}
				namingsql = &render.NamingSqlData{
					MethodName: value,
				}
			case 1: // sql对应的db
				if value == "" {
					add = false
					continue
				}
				namingsql.NamingSql = value
				sqlParameterList, err := common.ParseWhereConditions(namingsql.NamingSql)
				if err != nil {
					panic(err.Error())
				}
				if len(sqlParameterList) != 0 {
					namingsql.ParamColNameList = append(namingsql.ParamColNameList, sqlParameterList...)
				}
				selectColList, err := common.ParseSqlSelect(namingsql.NamingSql)
				if err != nil {
					panic(err.Error())
				}
				if selectColList != nil {
					namingsql.SelectColumns = selectColList
				}
			case 2:
				// 只有前两项不为空的情况下才可以添加到map中
				if add {
					keyPrefix := strings.ToUpper(value)
					namingsql.DbType = keyPrefix
					if keyPrefix != "" {
						keyPrefix = keyPrefix + "@"
					}
					namingsqlMap[keyPrefix+namingsql.MethodName] = namingsql
				}
			default:
			}
		}
		if add && namingsql != nil {
			namingsqlArr = append(namingsqlArr, namingsql)
		}
	}
	table.NamingSqlList = namingsqlArr
	// 支持多个db类型的场景使用
	table.NamingsqlMap = namingsqlMap
}
