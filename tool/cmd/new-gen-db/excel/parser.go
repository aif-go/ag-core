package excel

import (
	"ag-core/tool/cmd/new-gen-db/utils"
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
	for index, row := range rows {
		if index == 0 {
			continue // 跳过表头
		}

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
	}

	return nil
}
