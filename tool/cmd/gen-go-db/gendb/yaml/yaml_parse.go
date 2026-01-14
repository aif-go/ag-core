package yaml

import (
	"ag-core/tool/cmd/gen-go-db/gendb/render"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ---------------------- 2. 核心工具函数 ----------------------
// 拆分模块：解析表结构
func ParseTableStruct(table DatabaseTable) {
	fmt.Println("===== 【1. 表结构模块】 =====")
	fmt.Printf("表名：%s\n", table.TableName)
	if table.TableComment != "" {
		fmt.Printf("表注释：%s\n", table.TableComment)
	}
	fmt.Println("列定义：")
	// 使用orderedmap的Keys()方法获取列名列表，以保持顺序
	columnKeys := table.Columns.Keys()
	for _, colName := range columnKeys {
		// 从orderedmap中获取列值
		columnValue, _ := table.Columns.Get(colName)
		col := columnValue.(Column) // 类型断言为Column类型

		fmt.Printf("  %s:\n", colName)
		fmt.Printf("    数据库列名：%s\n", col.DbColumn)
		fmt.Printf("    Go类型：%s\n", col.GoType)
		fmt.Printf("    注释：%s\n", col.Comment)
		fmt.Printf("    主键：%t\n", col.PrimaryKey)
		fmt.Printf("    非空：%t\n", col.NotNull)
		if col.Length != "" {
			fmt.Printf("    长度：%s\n", col.Length)
		}
		fmt.Println("---")
	}
}

// 拆分模块：解析索引和主键（适配扁平化索引）
func ParseIndexAndPK(table DatabaseTable) {
	fmt.Println("\n===== 【2. 索引/主键模块】 =====")
	// 主键
	fmt.Println("主键列表：")
	for _, pk := range table.PrimaryKeys {
		fmt.Printf("  - %s\n", pk.Column)
	}
	// 普通索引（适配Columns字段）
	fmt.Println("普通索引：")
	for _, idx := range table.Indexes.General {
		fmt.Printf("  索引名：%s，绑定列：%s\n", idx.IndexName, strings.Join(idx.Columns, ","))
	}
	// 唯一索引（适配Columns字段）
	fmt.Println("唯一索引：")
	for _, idx := range table.Indexes.Unique {
		fmt.Printf("  索引名：%s，绑定列：%s\n", idx.IndexName, strings.Join(idx.Columns, ","))
	}
}

// 递归构建WHERE条件SQL
func buildWhereSQL(node WhereNode) string {
	if node.Expr != "" {
		return node.Expr
	}

	// 检查Conditions是否为nil
	if node.Conditions == nil {
		return ""
	}

	// 如果只有一个条件，直接返回该条件的SQL
	if len(*node.Conditions) == 1 {
		return buildWhereSQL((*node.Conditions)[0])
	}

	var parts []string
	for _, child := range *node.Conditions {
		parts = append(parts, buildWhereSQL(child))
	}

	// 根据测试要求，在条件两侧添加括号
	if len(parts) > 1 {
		joined := strings.Join(parts, " "+node.Operator+" ")
		// 只有当父节点有操作符时才添加外层括号
		if node.Operator != "" {
			return "(" + joined + ")"
		}
		return joined
	}

	return strings.Join(parts, " "+node.Operator+" ")
}

// 从WhereNode中提取参数
func extractParamsFromWhereNode(node WhereNode) []render.SqlParameter {
	seenParams := make(map[string]bool) // 在整个WHERE节点范围内跟踪已见参数
	return extractParamsFromWhereNodeWithSeenParams(node, seenParams)
}

// 从表达式中提取参数
func extractParamsFromExpression(expr string, seenParams map[string]bool) []render.SqlParameter {
	if seenParams == nil {
		seenParams = make(map[string]bool) // 用于跟踪已添加的参数名，避免重复
	}
	var params []render.SqlParameter

	// 使用正则表达式匹配 "列名 操作符 @参数名" 格式，例如 "name = @nameParam"
	re := regexp.MustCompile(`(\w+)\s*([=!<>]+|LIKE|like|Like|IN|in|In)\s*@(\w+)`)
	matches := re.FindAllStringSubmatch(expr, -1)

	for _, match := range matches {
		if len(match) == 4 {
			parameterName := "@" + match[3]
			// 检查参数是否已经存在
			if !seenParams[parameterName] {
				param := render.SqlParameter{
					ColName:       match[1],                             // 列名
					ParameterName: parameterName,                        // 完整的@参数名
					IsSlice:       strings.HasSuffix(match[3], "Slice"), // 检查参数名是否以Slice结尾
				}
				params = append(params, param)
				seenParams[parameterName] = true
			}
		}
	}

	return params
}

// 从表达式中提取列名
func extractColumnsFromExpression(expr string) []string {
	var columns []string

	// 使用正则表达式匹配 "列名 操作符 @参数名" 格式，例如 "name = @nameParam"
	re := regexp.MustCompile(`(\w+)\s*([=!<>]+|LIKE|like|Like|IN|in|In)\s*@(\w+)`)
	matches := re.FindAllStringSubmatch(expr, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			columnName := match[1]
			columns = append(columns, columnName)
		}
	}

	// Also handle other common patterns like "column = value" or "column IN (...)"
	re2 := regexp.MustCompile(`(\w+)\s*([=!<>]+|LIKE|like|Like|IN|in|In)\s*(\w+|\(.+?\)|'.*?'|".*?")`)
	matches2 := re2.FindAllStringSubmatch(expr, -1)

	for _, match := range matches2 {
		if len(match) >= 2 {
			columnName := match[1]
			// Check if it's already added
			found := false
			for _, col := range columns {
				if col == columnName {
					found = true
					break
				}
			}
			if !found {
				columns = append(columns, columnName)
			}
		}
	}

	return columns
}

// 从WhereNode中提取列名（内部递归函数）
func extractColumnsFromWhereNode(node WhereNode) []string {
	var columns []string

	if node.Expr != "" {
		// 从表达式中提取列名
		exprColumns := extractColumnsFromExpression(node.Expr)
		columns = append(columns, exprColumns...)
		return columns
	}

	// 检查Conditions是否为nil
	if node.Conditions == nil {
		return columns
	}

	// 递归处理子条件
	for _, child := range *node.Conditions {
		childColumns := extractColumnsFromWhereNode(child)
		columns = append(columns, childColumns...)
	}

	return columns
}

// 从Indexes结构中收集所有索引的列名
func getAllIndexedColumns(tableData *render.TableData) map[string]bool {
	indexedColumns := make(map[string]bool)

	// 收集普通索引的列
	for _, index := range tableData.GeneralIndexList {
		for _, bindParam := range index.BindParamList {
			indexedColumns[strings.ToUpper(bindParam.DbColName)] = true
		}
	}
	// 收集唯一索引的列
	for _, index := range tableData.UniqueIndexList {
		for _, bindParam := range index.BindParamList {
			indexedColumns[strings.ToUpper(bindParam.DbColName)] = true
		}
	}
	for _, bindParam := range tableData.PrimryRIndex.BindParamList {
		indexedColumns[strings.ToUpper(bindParam.DbColName)] = true
	}
	return indexedColumns
}

// 从WhereNode中提取参数（内部递归函数，使用共享的seenParams）
func extractParamsFromWhereNodeWithSeenParams(node WhereNode, seenParams map[string]bool) []render.SqlParameter {
	var params []render.SqlParameter

	if node.Expr != "" {
		// 从表达式中提取参数
		exprParams := extractParamsFromExpression(node.Expr, seenParams)
		params = append(params, exprParams...)
		return params
	}

	// 检查Conditions是否为nil
	if node.Conditions == nil {
		return params
	}

	// 递归处理子条件
	for _, child := range *node.Conditions {
		childParams := extractParamsFromWhereNodeWithSeenParams(child, seenParams)
		params = append(params, childParams...)
	}

	return params
}

// 解析有序的self_query_rules（核心改造点）
func parseOrderedQueryRules(node *yaml.Node) ([]OrderedQueryRule, error) {
	var orderedRules []OrderedQueryRule

	// self_query_rules是Map节点，Content为[key, value, key, value...]
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("self_query_rules不是合法的Map节点")
	}

	// 遍历Map节点（步长2：key + value）
	for i := 0; i < len(node.Content); i += 2 {
		// 解析方法名（Key节点）
		keyNode := node.Content[i]
		methodName := keyNode.Value

		// 解析查询规则（Value节点）
		valNode := node.Content[i+1]
		var rule QueryRule
		if err := valNode.Decode(&rule); err != nil {
			return nil, fmt.Errorf("解析方法%s配置失败：%v", methodName, err)
		}

		// TODO 暂时不支持分页查询，待分页接口定义清晰之后公开
		rule.Page = false

		// 添加到有序列表
		orderedRules = append(orderedRules, OrderedQueryRule{
			MethodName: methodName,
			Rule:       rule,
		})
	}

	return orderedRules, nil
}

// extractAliasFromFunction 从聚合函数表达式中解析别名
// 例如："SUM(AMT) AS TOTAL_AMT" -> "TOTAL_AMT"
func extractAliasFromFunction(function string) string {
	// 将字符串转换为大写以便匹配
	upperFunction := strings.ToUpper(function)

	// 查找 "AS" 关键字的位置
	asIndex := strings.Index(upperFunction, " AS ")
	if asIndex != -1 {
		// 返回 "AS" 后面的部分作为别名
		return strings.TrimSpace(function[asIndex+4:])
	}

	// 如果没有找到 "AS"，则返回原始函数表达式
	return function
}

// ConvertSelfQueryRulesToNamingSql 将 SelfQueryRules 转换为 NamingSqlData
func ConvertSelfQueryRulesToNamingSql(tableName string, orderedRules []OrderedQueryRule, tableData *render.TableData) []*render.NamingSqlData {
	var namingSqlList []*render.NamingSqlData
	// 获取所有索引的列
	indexedColumns := getAllIndexedColumns(tableData)
	for _, item := range orderedRules {
		methodName := item.MethodName
		rule := item.Rule
		// 构建SELECT子句
		selectAllCol := false
		selectClause := rule.SelectFields
		// 如果查询的是*，处理重置为表的所有列
		if selectClause == "*" {
			selectClause = ""
			selectAllCol = true
			allCols := []string{}
			for key, _ := range tableData.ColumnDataMap {
				allCols = append(allCols, key)
			}
			selectClause = strings.Join(allCols, ",")
		}

		if rule.Aggregation != nil {
			selectClause = strings.Join([]string{selectClause, rule.Aggregation.Function}, ",")
		}

		// 检查rule.where涉及到的列Expr必须是索引列
		if rule.Where != nil {
			// 提取WHERE子句中使用的所有列名
			whereColumns := extractColumnsFromWhereNode(*rule.Where)
			// 检查每个WHERE子句中使用的列是否都有索引
			for _, column := range whereColumns {
				if !indexedColumns[strings.ToUpper(column)] {
					// 如果列不在索引中，返回错误
					fmt.Errorf("查询方法 %s 的WHERE条件中使用了非索引列: %s。所有WHERE条件中的列都必须是索引列。", methodName, column)
					continue
				}
			}
		}

		// 构建WHERE子句
		whereClause := ""
		if rule.Where != nil {
			whereSQL := buildWhereSQL(*rule.Where)
			if whereSQL != "" {
				whereClause = "WHERE " + whereSQL
			}
		}

		// 构建ORDER BY子句
		orderByClause := ""
		if rule.OrderBy != "" {
			orderByClause = "ORDER BY " + rule.OrderBy
		} else if rule.Page {
			// 如果没有指定OrderBy，默认按主键排序
			orderByClause = "ORDER BY " + tableData.PrimaryKeyList
		}

		// 拼接完整SQL
		sqlParts := []string{
			fmt.Sprintf("SELECT %s", selectClause),
			fmt.Sprintf("FROM %s", tableName),
		}
		if whereClause != "" {
			sqlParts = append(sqlParts, whereClause)
		} else {
			sqlParts = append(sqlParts, " ")
		}
		if orderByClause != "" {
			sqlParts = append(sqlParts, orderByClause)
		} else {
			sqlParts = append(sqlParts, " ")
		}
		// TODO 此时判断是否需要按照分页处理，处理最终sql
		// finalSQL := strings.Join(sqlParts, " ")
		finalSQL := buildSql(rule.Page, sqlParts, tableData.DbType)
		// 直接从rule中提取参数，无需二次解析SQL
		var renderParams []render.SqlParameter
		if rule.Where != nil {
			// 从WHERE条件中提取参数
			renderParams = extractParamsFromWhereNode(*rule.Where)
		}

		// 直接从rule中构建selectColumns，无需二次解析SQL
		var renderSelectColumns []*render.SelectColumn
		if rule.Aggregation != nil && rule.Aggregation.Function != "" {
			// 如果有聚合函数，使用聚合函数的结果
			// 从Function属性中解析别名，例如 "SUM(AMT) AS TOTAL_AMT"
			alias := extractAliasFromFunction(rule.Aggregation.Function)
			col := &render.SelectColumn{
				ColumnName: alias,
				Alias:      ToCamelCase(strings.ToLower(alias)),
				GoType:     rule.Aggregation.ResultType,
			}
			renderSelectColumns = append(renderSelectColumns, col)
		}
		if rule.SelectFields != "" && rule.SelectFields != "*" {
			// 拆分SelectFields为多个列
			fields := strings.Split(rule.SelectFields, ",")
			for _, field := range fields {
				field = strings.TrimSpace(field)
				if field != "" {
					// 根据field作为key从tableData的colmap中获取对应的列信息，然后Alias为对应的GoName
					field = strings.ToUpper(field) // 默认使用字段名作为别名
					// 此处不需要做列不存在判断，如果通过列找不到对象的问题就是研发填写的问题
					col := &render.SelectColumn{
						ColumnName: field,
						Alias:      tableData.ColumnDataMap[field].GoColName,
						GoType:     tableData.ColumnDataMap[field].GoType,
					}

					renderSelectColumns = append(renderSelectColumns, col)
				}
			}
		}

		var pageCountSql string
		if rule.Page {
			// 分页查询，需要额外处理PageCountSql
			// 假设分页查询的SQL为：SELECT COUNT(*) FROM (SELECT %s FROM %s %s %s) t
			pageCountSql = fmt.Sprintf("SELECT COUNT(*) %s %s", sqlParts[1], sqlParts[2])
		}
		// 创建NamingSqlData
		namingSqlData := &render.NamingSqlData{
			MethodName:       methodName,
			NamingSql:        finalSQL,
			DbType:           " ", // 设置默认值为空格字符，表示适用于所有数据库类型
			ParamColNameList: renderParams,
			SelectColumns:    renderSelectColumns,
			PageCountSql:     pageCountSql,
			Page:             rule.Page,
			SelectAllCol:     selectAllCol,
		}
		namingSqlList = append(namingSqlList, namingSqlData)
	}

	return namingSqlList
}

// buildSql 构建SQL语句，根据是否分页添加分页子句
func buildSql(page bool, sqlParts []string, dbType string) string {
	if page {
		switch dbType {
		case "mysql":
			// 构建分页子句
			return fmt.Sprintf("%s %s %s %s limit %s,%s", sqlParts[0], sqlParts[1], sqlParts[2], sqlParts[3], "@StartNum", "@EndNum")
		default:
			// 构建分页子句
			return fmt.Sprintf("%s from (%s %s %s) t where t.rn between %s and %s", sqlParts[0], sqlParts[0], sqlParts[1], sqlParts[2], "@StartNum", "@EndNum")
		}
	}
	// mysql 使用limit 偏移量, 每页数量
	return strings.Join(sqlParts, " ")
}

// // ConvertToRenderTableData 将解析后的YAML数据转换为render包中的TableData
// func ConvertToRenderTableData(table DatabaseTable) *render.TableData {
// 	return ConvertToTableData(table, nil)
// }

// ConvertConfigToRenderTableData 将完整的Config解析为render包中的TableData，包括处理SelfQueryRules
func ConvertConfigToRenderTableData(config YamlDataConfig) *render.TableData {
	return ConvertToTableData(config.DatabaseTable, &config.SelfQueryRules)
}
