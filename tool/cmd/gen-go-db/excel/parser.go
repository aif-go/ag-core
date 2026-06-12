package excel

import (
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/utils"
	"errors"
	"regexp"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// ParseExcel 解析Excel文件，返回表结构信息
func ParseExcel(filePath string) (map[string]*ExcelInfo, error) {
	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}

	// 存储所有表信息
	tables := make(map[string]*ExcelInfo)

	
	// 遍历所有工作表
	for _, sheetName := range f.GetSheetMap() {
		// 跳过名为 "自定义脚本名字" 的工作表
		// if strings.TrimSpace(sheetName) == "自定义脚本名字" {
		// 	continue
		// }
		if strings.HasSuffix(sheetName,utils.CUSTOM_RULE_SUFFIX){
			// 如果sheet名字以 "@custom" 结尾，说明是分割出来的sheet，也跳过
			continue			
		}

		table := &ExcelInfo{
			// Name:        sheetName,
			Columns:     []*ColumnInfo{},
			PrimaryKey:  []string{},
			Constraints: []*ConstraintInfo{},
			Indexes:     []*IndexInfo{},
			SelfQueries: make(map[string]*SelfQueryInfo),
		}

		// 解析工作表内容
		rows := f.GetRows(sheetName)
		if len(rows) == 0 {
			continue
		}

		// 解析列信息
		inColumns := false
		inPrimaryKey := false
		inConstraints := false
		inIndexes := false
		// inSelfQueries := false

		for _, row := range rows {
			// 跳过空行
			if len(row) == 0 || strings.TrimSpace(row[0]) == "" {
				continue
			}

			// 处理首行数据
			if strings.TrimSpace(row[0]) == "表名"{
				table.Name = row[1]
				continue;
			}
			// 检查当前区域
			if strings.TrimSpace(row[0]) == "列名" {
				inColumns = true
				inPrimaryKey = false
				inConstraints = false
				inIndexes = false
				// inSelfQueries = false
				continue
			} else if strings.TrimSpace(row[0]) == "主键" {
				inColumns = false
				inPrimaryKey = true
				inConstraints = false
				inIndexes = false
				// inSelfQueries = false
				continue
			} else if strings.TrimSpace(row[0]) == "约束" {
				inColumns = false
				inPrimaryKey = false
				inConstraints = true
				inIndexes = false
				// inSelfQueries = false
				continue
			} else if strings.TrimSpace(row[0]) == "索引" {
				inColumns = false
				inPrimaryKey = false
				inConstraints = false
				inIndexes = true
				// inSelfQueries = false
				continue
			} else if strings.TrimSpace(row[0]) == "方法名字" {
				inColumns = false
				inPrimaryKey = false
				inConstraints = false
				inIndexes = false
				// inSelfQueries = true
				continue
			}

			// 解析列信息
			if inColumns {
				if len(row) >= 9 {
					column := &ColumnInfo{
						Name:          strings.TrimSpace(row[0]),
						Type:          strings.TrimSpace(row[1]),
						Length:        strings.TrimSpace(row[2]),
						NotNull:       strings.TrimSpace(row[3]) == "Y",
						Default:       strings.TrimSpace(row[4]),
						AutoIncrement: strings.TrimSpace(row[5]) == "Y",
						SupportUpdate: strings.TrimSpace(row[6]) == "Y",
						Description:   strings.TrimSpace(row[7]),
						Tag:           strings.TrimSpace(row[8]),
					}
					table.Columns = append(table.Columns, column)
				}
			}

			// 解析主键
			if inPrimaryKey {
				if len(row) >= 2 && strings.TrimSpace(row[0]) == "PRIMARY_KEY" {
					// 遍历所有列，获取所有主键
					for j := 1; j < len(row); j++ {
						if strings.TrimSpace(row[j]) != "" {
							table.PrimaryKey = append(table.PrimaryKey, strings.TrimSpace(row[j]))
						}
					}
				}
			}

			// 解析约束
			if inConstraints {
				if len(row) >= 2 && strings.TrimSpace(row[0]) != "" {
					constraint := &ConstraintInfo{
						Name:    strings.TrimSpace(row[0]),
						Columns: []string{},
					}
					for j := 1; j < len(row); j++ {
						if strings.TrimSpace(row[j]) != "" {
							constraint.Columns = append(constraint.Columns, strings.TrimSpace(row[j]))
						}
					}
					table.Constraints = append(table.Constraints, constraint)
				}
			}

			// 解析索引
			if inIndexes {
				if len(row) >= 2 && strings.TrimSpace(row[0]) != "" {
					indexName := strings.TrimSpace(row[0])
					// 跳过名为 "自定义脚本名字" 的索引
					if indexName == "自定义脚本名字" {
						continue
					}
					index := &IndexInfo{
						Name:    indexName,
						Columns: []string{},
					}
					for j := 1; j < len(row); j++ {
						if strings.TrimSpace(row[j]) != "" {
							index.Columns = append(index.Columns, strings.TrimSpace(row[j]))
						}
					}
					table.Indexes = append(table.Indexes, index)
				}
			}
		}
		// 处理自定义脚本的工作表
		processCustomScriptSheet(f, sheetName, table)
		tables[sheetName] = table
	}

	return tables, nil
}

// 处理客户自定义规则的工作表
func processCustomScriptSheet(f *excelize.File,sheetName string, table *ExcelInfo) error {
	// 处理 "自定义脚本名字" 工作表的内容
	rows := f.GetRows(sheetName + utils.CUSTOM_RULE_SUFFIX)
	if len(rows) == 0 {
		return nil
	}
	dynamicTemplate:=false
	for _, row := range rows {
		// if index == 0 {
		// 	continue // 跳过表头
		// }
		if strings.EqualFold(row[0],"动态模版"){
			dynamicTemplate=true
			continue
		}
		// 空白行  动态模版 方法名字都跳过改行不处理
		if strings.EqualFold(row[0],"") || strings.EqualFold(row[0],"方法名字"){
			continue
		}
		// 处理原先的自定义脚本
		if !dynamicTemplate{
		if len(row) >= 8 && strings.TrimSpace(row[0]) != "" {
			query := &SelfQueryInfo{
				SelectFields: strings.TrimSpace(row[1]),
				Page:         strings.TrimSpace(row[7]) == "Y",
			}
			// 解析WHERE子句
			whereExpr := strings.TrimSpace(row[4])
			if whereExpr != "" {
				query.Where = ParseWhereCondition(whereExpr)
			}
			table.SelfQueries[strings.TrimSpace(row[0])] = query
		}
		}else{
			// 处理自定义模版
			err:=processDynamicTemplate(row,table)
			if err!=nil{
				return err
			}
		}
	}

	return nil
}



func processDynamicTemplate(rows []string, table *ExcelInfo) error{
		if len(rows) >= 5 {
			query := &SelfQueryInfo{
				SelectFields: strings.TrimSpace(rows[1]),
				SqlTemplate:  strings.TrimSpace(rows[2]),
				Page:         strings.TrimSpace(rows[5]) == "Y",
				Order: 		  strings.TrimSpace(rows[4]),
				DynamicTemplate: true,
			}
			result, err := processWhereParam(query.SqlTemplate, strings.TrimSpace(rows[3]),table)
			if err!=nil{
				return err
			}
			query.WhereParams=result
			// 解析参数对象，如果未指定的则去table对象中查找 用一个正则表达式找到模版中所有 @ 开头的变量，然后获取@后面的属性名字
			table.SelfQueries[strings.TrimSpace(rows[0])] = query
			return nil
		}
		return errors.New("动态模版部分格式不正确,请检查excel")
}


func processWhereParam(sqlTmpl string, whereparam string, table *ExcelInfo) ([]*WhereParam,error) {
	// 提取出sqltmpl变量中所有 " @xxx " 规则的内容-->whereparam中paramName
	// 处理whereparam 格式如下"属性名","类型","Y/N"--whereparam中的paramName,Type,Slice
	
	if strings.Contains(sqlTmpl, "1 = 1"){
		return nil,errors.New("动态模版中不允许出现1=1的条件")
	}
	// 使用正则表达式提取所有 @xxx 格式的参数名
	re := regexp.MustCompile(`@(\w+)`)
	matches := re.FindAllStringSubmatch(sqlTmpl, -1)

	// 将结果放到[]*whereparam中
	var result []*WhereParam
	// 创建参数名到 whereparam 配置的映射
	// whereparam 格式: "属性名","类型","Y/N"
	// 解析为 map[属性名]map[string]string{Type: 类型, Slice: Y/N}
	paramConfig := make(map[string]bool)

		if whereparam != "" {
		// 按逗号分割，处理格式 "属性名","类型","Y/N"
		params := strings.Split(whereparam, ";")
		for _,param:=range params{
			parts:=strings.Split(param,",")
			// for i := 0; i < len(parts); i +=1 {
				paramName := strings.TrimSpace(parts[0])
				paramType := strings.TrimSpace(parts[1])
				sliceFlag := strings.TrimSpace(parts[2])=="Y"
				colName:=""
				if len(parts)>=4{
					colName=strings.TrimSpace(parts[3])
				}
				whereparam:=&WhereParam{
					ParaName: paramName,
					Type: paramType,
					Slice: sliceFlag,
					// 这个地方是否一定要有值???
					ColName: colName,
				}
				result=append(result, whereparam)
				paramConfig[paramName] = true
			// }
		}
	}

	for _, match := range matches {
		if len(match) > 1 {
			paramName := match[1]
			if _,ok:=paramConfig[paramName];ok {
				continue
			}
			paramConfig[paramName] = true
			
			wp := &WhereParam{
				ParaName: paramName,
			}
			
			for _, col := range table.Columns {
				// 这里应该有个驼峰策略
				if utils.ToCamelCase(col.Name) == paramName {
					wp.Type = col.Type
					wp.ColName = col.Name
					break
				}
			}
			
			result = append(result, wp)
		}
	}	
	return result, nil
}
