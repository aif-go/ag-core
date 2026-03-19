package model

import (
	"ag-core/contribute/agdb/conditonwhere"
	"ag-core/tool/cmd/new-gen-db/table"
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

// ParseYAML 解析YAML文件并生成TableData
func ParseYAML(yamlPath string, moduleName string) (*table.TableData, error) {
	// 读取YAML文件
	data, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("读取YAML文件失败: %v", err)
	}

	// 解析YAML到map
	var yamlData map[interface{}]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("解析YAML文件失败: %v", err)
	}

	// 提取表信息
	tableName, ok := yamlData["table_name"].(string)
	if !ok {
		return nil, fmt.Errorf("YAML文件中缺少table_name字段")
	}

	// 转换表名为大驼峰命名
	structName := toCamelCase(tableName)

	// 提取列信息
	columns := []table.ColumnData{}
	primaryKeys := []string{}  // 新增：主键列表
	importPackages := []string{"fmt"}

	// 处理列数据
	if columnsData, ok := yamlData["columns"].([]interface{}); ok {
		for _, colData := range columnsData {
			if colMap, ok := colData.(map[interface{}]interface{}); ok {
				col := table.ColumnData{
					IndexPriorities: make(map[string]int),
				}

				// 提取基本信息
				if name, ok := colMap["name"].(string); ok {
					col.Name = name
					col.JsonTag = toCamelCase(name)
				}

				if colType, ok := colMap["type"].(string); ok {
					col.Type = colType
					col.GoType = getGoType(colType)

					// 检查是否需要导入额外的包
					if col.GoType == "time.Time" {
						importPackages = append(importPackages, "time")
					}
				}

				// 处理标签
				if tag, ok := colMap["tag"].(string); ok {
					if strings.Contains(tag, "///@create") {
						col.IsAutoCreate = true
					}
					if strings.Contains(tag, "///@update") {
						col.IsAutoUpdate = true
					}
					if strings.Contains(tag, "///@javaVersion") {
						col.IsJavaVersion = true
					}
				}

				// 处理support_update字段
				if supportUpdate, ok := colMap["support_update"].(bool); ok {
					col.SupportUpdate = supportUpdate
				}

				columns = append(columns, col)
			}
		}
	}

	// 处理索引信息
	indexes := []table.IndexData{}
	if indexesData, ok := yamlData["indexes"].([]interface{}); ok {
		for _, idxData := range indexesData {
			if idxMap, ok := idxData.(map[interface{}]interface{}); ok {
				idx := table.IndexData{}

				if name, ok := idxMap["name"].(string); ok {
					idx.Name = name
				}

				if cols, ok := idxMap["columns"].([]interface{}); ok {
					for i, col := range cols {
						if colName, ok := col.(string); ok {
							idx.Columns = append(idx.Columns, colName)
							// 更新列的索引优先级
							for j := range columns {
								if columns[j].Name == colName {
									columns[j].IndexPriorities[idx.Name] = i + 1
									break
								}
							}
						}
					}
				}
				indexes = append(indexes, idx)
			}
		}
	}

	// 处理约束信息（唯一索引等）
	if constraintsData, ok := yamlData["constraints"].([]interface{}); ok {
		for _, consData := range constraintsData {
			if consMap, ok := consData.(map[interface{}]interface{}); ok {
				consName, ok := consMap["name"].(string)
				if !ok {
					continue
				}

				consColumns := []string{}
				if cols, ok := consMap["columns"].([]interface{}); ok {
					for i, col := range cols {
						if colName, ok := col.(string); ok {
							consColumns = append(consColumns, colName)
							// 将约束视为索引处理
							for j := range columns {
								if columns[j].Name == colName {
									columns[j].IndexPriorities[consName] = i + 1
									break
								}
							}
						}
					}
				}

				// 将约束作为索引添加到indexes切片中
				if len(consColumns) > 0 {
					indexes = append(indexes, table.IndexData{
						Name:     consName,
						Columns:  consColumns,
						IsUnique: true, // 唯一索引
					})
				}
			}
		}
	}

	// 处理主键信息
	if primaryKey, ok := yamlData["primary_key"].([]interface{}); ok {
		for _, pk := range primaryKey {
			if pkName, ok := pk.(string); ok {
				// 添加到主键列表
				primaryKeys = append(primaryKeys, pkName)
				
				// 设置 IsPrimaryKey 标志（保留原有逻辑）
				for i := range columns {
					if columns[i].Name == pkName {
						columns[i].IsPrimaryKey = true
						break
					}
				}
			}
		}
	}

	// 生成GORM标签
	for i := range columns {
		columns[i].GormTag = generateGormTag(&columns[i], indexes)
	}

	// 处理自查询信息
	selfQueries := []table.QueryData{}
	globalHasPage := false
	globalHasSelfQuery := false
	if selfQueryData, ok := yamlData["self_query_rules"].(map[interface{}]interface{}); ok {
		globalHasSelfQuery = true
		for name, query := range selfQueryData {
			if queryMap, ok := query.(map[interface{}]interface{}); ok {
				q := table.QueryData{
					Name: name.(string),
				}

				if selectFields, ok := queryMap["select_fields"].(string); ok {
					q.SelectFields = selectFields
					if selectFields == "*" {
						// 全表查询
						for _, col := range columns {
							q.Fields = append(q.Fields, col.Name)
						}
					} else {
						// 指定字段查询
						fields := strings.Split(selectFields, ",")
						for _, field := range fields {
							q.Fields = append(q.Fields, strings.TrimSpace(field))
						}
					}
				}

				if page, ok := queryMap["page"].(bool); ok {
					// 区分每个自定已规则是否存在分页查询
					q.HasPage = page
					// 标记存在分页查询, import时引入page相关的内容使用
					if page {
						globalHasPage = true
					}
				}

				// 提取WHERE条件
				if where, ok := queryMap["where"].(map[interface{}]interface{}); ok {
					q.Where = conditonwhere.ParseWhereCondition(where)
					var whereYamlStr string
					whereYamlStr, err = whereDataToYAML(where)
					if err != nil {
						return nil, fmt.Errorf("解析where条件失败: %v", err)
					}
					q.WhereDataYaml = whereYamlStr
					// 提取所有字段信息
					extractWhereFields(q.Where, &q.WhereFields, &q.WhereColFields)
				}

				selfQueries = append(selfQueries, q)
			}
		}
	}

	// 去重导入包
	importPackages = unique(importPackages)

	// 填充AllowUpdateCols
	allowUpdateCols := []string{}
	for _, col := range columns {
		if col.SupportUpdate {
			allowUpdateCols = append(allowUpdateCols, col.Name)
		}
	}

	return &table.TableData{
		ModuleName:  moduleName,
		TableName:   tableName,
		StructName:  structName,
		Columns:     columns,
		PrimaryKeys: primaryKeys,  // 新增：主键列表
		Indexes:     indexes,
		SelfQueries: selfQueries,
		ModelTemplateData: &table.ModelTemplateData{
			ImportPackages: importPackages,
		},
		HasPage:         globalHasPage,
		HasSelfQuery:    globalHasSelfQuery,
		AllowUpdateCols: allowUpdateCols,
	}, nil
}

// 辅助函数：转换为大驼峰命名
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	result := ""
	for _, part := range parts {
		if part != "" {
			result += strings.Title(strings.ToLower(part))
		}
	}
	return result
}

// 辅助函数：获取Go类型
func getGoType(sqlType string) string {
	switch sqlType {
	case "int", "int32":
		return "int"
	case "int64":
		return "int64"
	case "string":
		return "string"
	case "time":
		return "time.Time"
	default:
		return "string"
	}
}

// 将where条件map转换为YAML字符串
func whereDataToYAML(whereData map[interface{}]interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(whereData)
	if err != nil {
		return "", fmt.Errorf("转换为YAML失败: %v", err)
	}
	return string(yamlBytes), nil
}

// // 解析where条件
// func parseWhereCondition(whereData map[interface{}]interface{}) *conditonwhere.MaskWhereCondition {
// 	condition := &conditonwhere.MaskWhereCondition{}

// 	// 解析operator
// 	if operator, ok := whereData["operator"].(string); ok {
// 		condition.Operator = operator
// 	} else {
// 		condition.Operator = "AND" // 默认使用AND
// 	}

// 	// 解析conditions
// 	if conditionsData, ok := whereData["conditions"].([]interface{}); ok {
// 		condition.Conditions = make([]conditonwhere.MaskWhereCondition, 0, len(conditionsData))
// 		for _, condData := range conditionsData {
// 			if condMap, ok := condData.(map[interface{}]interface{}); ok {
// 				// 检查是嵌套条件还是表达式
// 				if _, hasExpr := condMap["expr"]; hasExpr {
// 					// 是表达式
// 					subCond := conditonwhere.MaskWhereCondition{
// 						Expr: condMap["expr"].(string),
// 					}
// 					condition.Conditions = append(condition.Conditions, subCond)
// 				} else {
// 					// 是嵌套条件
// 					subCond := parseWhereCondition(condMap)
// 					condition.Conditions = append(condition.Conditions, *subCond)
// 				}
// 			}
// 		}
// 	}

// 	return condition
// }

// 提取where条件中的所有字段信息
func extractWhereFields(condition *conditonwhere.MaskWhereCondition, fields *[]string, whereColFields *[]table.WhereColField) {
	if condition.Expr != "" {
		// 解析表达式，提取列名、操作符和字段名
		colField := parseWhereExpr(condition.Expr)
		if colField.ColName != "" && colField.FieldName != "" {
			// 添加到WhereFields
			*fields = append(*fields, colField.ColName)
			// 添加到WhereColFields
			*whereColFields = append(*whereColFields, colField)
		}
	} else {
		for i := range condition.Conditions {
			extractWhereFields(&condition.Conditions[i], fields, whereColFields)
		}
	}
}

// parseWhereExpr 解析where表达式，提取列名、操作符和字段名
func parseWhereExpr(expr string) table.WhereColField {
	colField := table.WhereColField{}

	// 支持的操作符，按长度降序排列，确保长操作符优先匹配
	operators := []string{"!=", "not in", "in", "=", ">", "<", ">=", "<=", "between"}

	for _, op := range operators {
		if idx := strings.Index(strings.ToLower(expr), strings.ToLower(op)); idx != -1 {
			// 提取列名
			colName := strings.TrimSpace(expr[:idx])

			// 提取字段名
			fieldPart := strings.TrimSpace(expr[idx+len(op):])
			fieldName := ""
			isSlice := false

			// 处理@Field格式
			if strings.HasPrefix(fieldPart, "@") {
				fieldName = strings.TrimSpace(fieldPart[1:])
			}

			// 判断是否为切片类型
			lowerOp := strings.ToLower(op)
			if lowerOp == "in" || lowerOp == "not in" {
				isSlice = true
			}

			colField = table.WhereColField{
				ColName:   colName,
				FieldName: fieldName,
				IsSlice:   isSlice,
				Operator:  op,
			}
			break
		}
	}

	return colField
}

// 辅助函数：生成GORM标签
func generateGormTag(col *table.ColumnData, indexes []table.IndexData) string {
	tags := []string{}

	// 列名
	tags = append(tags, fmt.Sprintf("column:%s", col.Name))

	// 主键
	if col.IsPrimaryKey {
		tags = append(tags, "primaryKey")
		tags = append(tags, "not null")
	}

	// 自动创建时间
	if col.IsAutoCreate {
		tags = append(tags, "AUTOCREATETIME")
	}

	// 自动更新时间
	if col.IsAutoUpdate {
		tags = append(tags, "AUTOUPDATETIME")
	}

	// 索引
	for indexName, priority := range col.IndexPriorities {
		tags = append(tags, fmt.Sprintf("index:%s,priority:%d", indexName, priority))
	}

	return strings.Join(tags, ";")
}

// 辅助函数：生成索引列变量
func generateIndexColumns(indexes []table.IndexData) string {
	columnsMap := make(map[string]bool)
	for _, idx := range indexes {
		for _, col := range idx.Columns {
			columnsMap[col] = true
		}
	}

	columns := []string{}
	for col := range columnsMap {
		columns = append(columns, `"`+col+`"`)
	}

	return `[]string{` + strings.Join(columns, `, `) + `}`
}

// 辅助函数：去重
func unique(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}
	for _, entry := range slice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = true
			result = append(result, entry)
		}
	}
	return result
}
