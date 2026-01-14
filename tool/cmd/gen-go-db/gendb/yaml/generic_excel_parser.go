package yaml

import (
	"ag-core/tool/cmd/gen-go-db/gendb/render"
	"fmt"
	"os"
	"strings"

	"github.com/iancoleman/orderedmap"
	"github.com/tealeg/xlsx"
	"gopkg.in/yaml.v3"
)

// GenerateYamlFromExcel 将excel数据转换为yaml文件
func GenerateYamlFromExcel(config *render.AGInfraStructrueConfig) error {
	xlFile, err := xlsx.OpenFile(config.DbTemplatePath)
	if err != nil {
		return err
	}
	supportTablesLen := len(config.SupportTables)
	for _, sheet := range xlFile.Sheets {
		// 如果配置了表名但是当前处理的表名又不在里面，则放弃处理该表
		if _, ok := config.SupportTables[strings.ToLower(sheet.Name)]; !ok && supportTablesLen != 0 {
			continue
		}
		columns, rules, primaryKeys, constraints, indexes, err := ParseGenericExcelSheet(sheet)

		if err != nil {
			fmt.Println("sheetname:", sheet.Name, "解析失败,错误信息:", err.Error())
			continue
		}

		yamlDataConfig, err := ConvertGenericToYAML(sheet.Name, columns, rules, primaryKeys, constraints, indexes)
		if err != nil {
			fmt.Println("sheetname:", sheet.Name, "转换为yaml数据失败,错误信息:", err.Error())
			continue
		}

		// 序列化为YAML格式
		yamlBytes, err := yaml.Marshal(yamlDataConfig)
		if err != nil {
			fmt.Println("sheetname:", sheet.Name, "yaml对象转换为字节流失败,错误信息:", err.Error())
			continue
		}

		// 写入到generic_tm_media_act.yaml文件
		filename := strings.ToLower(sheet.Name) + ".yaml"
		outputPath := config.OutputPath + filename
		err = os.WriteFile(outputPath, yamlBytes, 0644)
		if err != nil {
			fmt.Println("sheetname:", sheet.Name, "最后阶段生成yaml文件失败,错误信息:", err.Error())
		}
	}

	return nil
}

// ParseGenericExcelSheet parses a generic Excel sheet
func ParseGenericExcelSheet(sheet *xlsx.Sheet) ([]GenericColumn, []GenericRule, []PrimaryKeyInfo, []ConstraintInfo, []IndexInfo, error) {
	// 解析数据
	var columns []GenericColumn
	var queryRules []GenericRule
	var primaryKeys []PrimaryKeyInfo
	var constraints []ConstraintInfo
	var indexes []IndexInfo

	// 第一行是标题行，从第二行开始读取数据
	// 在第3行之后，如果遇到空行或者特殊标记行，则停止解析列定义
	dataSectionEnded := false
	inRulesSection := false
	inPrimaryKeySection := false
	inConstraintSection := false
	inIndexSection := false
	skipNextRowAsTitle := false // 标记是否需要跳过下一行（标题行）

	for i, row := range sheet.Rows {
		// 跳过标题行
		if i == 0 {
			continue
		}

		// 检查是否需要跳过当前行（标题行）
		if skipNextRowAsTitle {
			skipNextRowAsTitle = false
			continue
		}

		// 检查是否进入不同的定义部分
		if i > 2 && len(row.Cells) > 0 {
			firstCell := strings.TrimSpace(row.Cells[0].String())

			// 检查是否进入主键定义部分
			if firstCell == "主键" {
				dataSectionEnded = true
				inPrimaryKeySection = true
				inConstraintSection = false
				inIndexSection = false
				inRulesSection = false
				continue
			}

			// 检查是否进入约束定义部分
			if firstCell == "约束" {
				dataSectionEnded = true
				inPrimaryKeySection = false
				inConstraintSection = true
				inIndexSection = false
				inRulesSection = false
				continue
			}

			// 检查是否进入索引定义部分
			if firstCell == "索引" {
				dataSectionEnded = true
				inPrimaryKeySection = false
				inConstraintSection = false
				inIndexSection = true
				inRulesSection = false
				continue
			}

			// 检查是否进入规则定义部分
			if firstCell == "自定义脚本名字" {
				dataSectionEnded = true
				inPrimaryKeySection = false
				inConstraintSection = false
				inIndexSection = false
				inRulesSection = true
				skipNextRowAsTitle = true // 标记下一行是标题行，需要跳过
				continue
			}

			// 处理主键定义部分
			if inPrimaryKeySection {
				// 解析主键定义
				if len(row.Cells) > 1 {
					keyName := strings.TrimSpace(row.Cells[0].String())
					columnName := strings.TrimSpace(row.Cells[1].String())

					// 如果主键名称是PRIMARY_KEY，则记录主键信息
					if keyName == "PRIMARY_KEY" && columnName != "" {
						primaryKeys = append(primaryKeys, PrimaryKeyInfo{Column: columnName})

						// 同时标记对应列为主键
						for j := range columns {
							if columns[j].Name == columnName {
								columns[j].IsPrimaryKey = true
								break
							}
						}
					}
				}
				continue
			}

			// 处理约束定义部分
			if inConstraintSection {
				// 解析约束定义
				if len(row.Cells) > 0 {
					constraintName := strings.TrimSpace(row.Cells[0].String())
					if constraintName != "" {
						var constraintColumns []string
						// 获取所有非空的列名
						for j := 1; j < len(row.Cells); j++ {
							columnName := strings.TrimSpace(row.Cells[j].String())
							if columnName != "" {
								constraintColumns = append(constraintColumns, columnName)
							}
						}

						if len(constraintColumns) > 0 {
							constraints = append(constraints, ConstraintInfo{
								Name:    constraintName,
								Columns: constraintColumns,
							})
						}
					}
				}
				continue
			}

			// 处理索引定义部分
			if inIndexSection {
				// 解析索引定义
				if len(row.Cells) > 0 {
					indexName := strings.TrimSpace(row.Cells[0].String())
					if indexName != "" {
						var indexColumns []string
						// 获取所有非空的列名
						for j := 1; j < len(row.Cells); j++ {
							columnName := strings.TrimSpace(row.Cells[j].String())
							if columnName != "" {
								indexColumns = append(indexColumns, columnName)
							}
						}

						if len(indexColumns) > 0 {
							indexes = append(indexes, IndexInfo{
								Name:    indexName,
								Columns: indexColumns,
							})
						}
					}
				}
				continue
			}

			// 处理规则定义部分
			if inRulesSection {
				// 解析规则定义
				rule := parseGenericRuleRow(row)
				// 检查规则是否有效（通过检查Name字段是否为空）
				if rule.Name != "" {
					queryRules = append(queryRules, rule)
				}
				continue
			}
		}

		// 如果已经到达数据定义结束点，则跳过后续行
		if dataSectionEnded && !inPrimaryKeySection && !inConstraintSection && !inIndexSection && !inRulesSection {
			continue
		}

		// 确保行有足够的单元格
		if len(row.Cells) < 7 {
			continue
		}

		// 创建列定义实例
		column := GenericColumn{}

		// 根据实际列数和结构填充column字段
		if len(row.Cells) > 0 {
			column.Name = strings.TrimSpace(row.Cells[0].String())
		}
		if len(row.Cells) > 1 {
			column.Type = strings.TrimSpace(row.Cells[1].String())
		}
		if len(row.Cells) > 2 {
			column.Length = strings.TrimSpace(row.Cells[2].String())
		}
		if len(row.Cells) > 3 {
			column.NotNull = strings.TrimSpace(row.Cells[3].String())
		}
		if len(row.Cells) > 4 {
			column.DefaultValue = strings.TrimSpace(row.Cells[4].String())
		}
		if len(row.Cells) > 5 {
			column.AutoIncrement = strings.TrimSpace(row.Cells[5].String())
		}
		if len(row.Cells) > 6 {
			column.Description = strings.TrimSpace(row.Cells[6].String())
		}
		if len(row.Cells) > 7 {
			column.Tag = strings.TrimSpace(row.Cells[7].String())
		}

		// 如果列名为空，则跳过该行
		if column.Name == "" {
			continue
		}

		// 如果类型为空或者不合法，则跳过该行
		if column.Type == "" {
			continue
		}

		columns = append(columns, column)
	}

	return columns, queryRules, primaryKeys, constraints, indexes, nil
}

// parseGenericRuleRow parses a row as a query rule
func parseGenericRuleRow(row *xlsx.Row) GenericRule {
	// 确保行有足够的单元格 (方法名字, 查询列, 聚合函数, 聚合类型, 检索条件, 排序条件, 分组条件, 分页,数据库类型)
	// 至少需要8列
	if len(row.Cells) < 9 {
		return GenericRule{}
	}
	// 解析规则字段
	name := strings.TrimSpace(row.Cells[0].String()) // 方法名字
	if name == "" {
		return GenericRule{}
	}
	// 解析选择字段
	selectFields := strings.TrimSpace(row.Cells[1].String()) // 查询列
	// 解析聚合函数
	aggregation := strings.TrimSpace(row.Cells[2].String()) // 聚合函数
	//  聚合函数类型
	aggregationType := strings.TrimSpace(row.Cells[3].String()) // 聚合函数
	// 解析条件表达式
	conditionStr := strings.TrimSpace(row.Cells[4].String()) // 检索条件
	// 解析排序条件
	ordering := strings.TrimSpace(row.Cells[5].String()) // 排序条件 (目前未使用)
	// 解析分组条件
	grouping := strings.TrimSpace(row.Cells[6].String())    // 分组条件 (目前未使用)
	page := strings.TrimSpace(row.Cells[7].String()) == "Y" // 分页
	// 解析数据库类型
	dbTypes := strings.TrimSpace(row.Cells[8].String()) // 数据库类型
	// 创建条件列表
	// var conditions []RuleCondition
	// if conditionStr != "" {
	// 解析复杂的WHERE条件表达式
	whereNode := parseConditionExpression(conditionStr)

	rule := GenericRule{
		Name:            name,
		SelectFields:    selectFields,
		Conditions:      whereNode,
		Aggregation:     aggregation,
		AggregationType: aggregationType,
		OrderBy:         ordering,
		GroupBy:         grouping,
		Page:            page,
		DBTypes:         dbTypes,
	}

	return rule
}

// parseConditionExpression 解析条件表达式为嵌套的where结构
func parseConditionExpression(conditionStr string) *WhereNode {
	// 如果条件字符串为空，直接返回空切片
	if conditionStr == "" {
		return &WhereNode{}
	}

	// 解析复杂的WHERE条件表达式
	whereNode := parseExpressionToWhereNode(conditionStr)
	if whereNode == nil {
		return &WhereNode{}
	}

	// 将解析后的WhereNode结构转换为RuleCondition列表
	// var conditions []RuleCondition
	// flattenWhereNode(whereNode, &conditions)

	return whereNode
}

// parseExpressionToWhereNode 解析单个表达式为嵌套的WhereNode结构
func parseExpressionToWhereNode(expr string) *WhereNode {
	return parseExpressionToWhereNodeWithDepth(expr, 0)
}

// parseExpressionToWhereNodeWithDepth 带深度控制的解析函数，防止无限递归
func parseExpressionToWhereNodeWithDepth(expr string, depth int) *WhereNode {
	// 防止无限递归
	if depth > 100 {
		return &WhereNode{
			Expr: expr,
		}
	}

	// 去除首尾空格
	expr = strings.TrimSpace(expr)

	// 如果表达式为空，返回nil
	if expr == "" {
		return nil
	}

	// 先尝试解析括号表达式中的操作符
	// 检查是否是括号包围的表达式
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		// 计算括号层数，确保是最外层括号
		if isOutermostParentheses(expr) {
			// 去除最外层括号
			innerExpr := strings.TrimSpace(expr[1 : len(expr)-1])

			// 尝试先解析内部表达式的操作符
			// 检查是否包含OR操作符（不区分大小写），优先处理OR
			parts := splitByOperator(innerExpr, " OR ")
			if len(parts) > 1 {
				// 创建OR节点
				orNode := &WhereNode{
					Operator:   "OR",
					Conditions: &[]WhereNode{},
				}

				// 为每个OR部分创建子节点
				for _, part := range parts {
					// 解析每个部分
					subNode := parseExpressionToWhereNodeWithDepth(part, depth+1)
					if subNode != nil {
						*orNode.Conditions = append(*orNode.Conditions, *subNode)
					}
				}

				// 如果有子节点，返回OR节点
				if len(*orNode.Conditions) > 0 {
					return orNode
				}
			} else {
				// 检查是否包含AND操作符（不区分大小写）
				parts := splitByOperator(innerExpr, " AND ")
				if len(parts) > 1 {
					// 创建AND节点
					andNode := &WhereNode{
						Operator:   "AND",
						Conditions: &[]WhereNode{},
					}

					// 为每个AND部分创建子节点
					for _, part := range parts {
						// 解析每个部分
						subNode := parseExpressionToWhereNodeWithDepth(part, depth+1)
						if subNode != nil {
							*andNode.Conditions = append(*andNode.Conditions, *subNode)
						}
					}

					// 如果有子节点，返回AND节点
					if len(*andNode.Conditions) > 0 {
						return andNode
					}
				}
			}

			// 如果去除括号后没有操作符，则检查是否还有嵌套的括号表达式需要处理
			if !containsOperator(innerExpr) {
				// 简单条件，直接作为表达式
				return &WhereNode{
					Expr: expr, // 保留原始带括号的表达式
				}
			}

			// 如果有操作符，则递归解析内部表达式
			innerNode := parseExpressionToWhereNodeWithDepth(innerExpr, depth+1)
			if innerNode != nil {
				// 如果内部解析结果是一个操作符节点，则保留外层括号信息
				if innerNode.Operator != "" {
					return &WhereNode{
						Expr: expr, // 保留原始带括号的表达式
					}
				}
				// 否则返回内部解析结果
				return innerNode
			}
		}
	}

	// 检查是否包含OR操作符（不区分大小写），优先处理OR
	parts := splitByOperator(expr, " OR ")
	if len(parts) > 1 {
		// 创建OR节点
		orNode := &WhereNode{
			Operator:   "OR",
			Conditions: &[]WhereNode{},
		}

		// 为每个OR部分创建子节点
		for _, part := range parts {
			// 解析每个部分
			subNode := parseExpressionToWhereNodeWithDepth(part, depth+1)
			if subNode != nil {
				*orNode.Conditions = append(*orNode.Conditions, *subNode)
			}
		}

		// 如果有子节点，返回OR节点
		if len(*orNode.Conditions) > 0 {
			return orNode
		}
	} else {
		// 检查是否包含AND操作符（不区分大小写）
		parts := splitByOperator(expr, " AND ")
		if len(parts) > 1 {
			// 创建AND节点
			andNode := &WhereNode{
				Operator:   "AND",
				Conditions: &[]WhereNode{},
			}

			// 为每个AND部分创建子节点
			for _, part := range parts {
				// 解析每个部分
				subNode := parseExpressionToWhereNodeWithDepth(part, depth+1)
				if subNode != nil {
					*andNode.Conditions = append(*andNode.Conditions, *subNode)
				}
			}

			// 如果有子节点，返回AND节点
			if len(*andNode.Conditions) > 0 {
				return andNode
			}
		} else {
			// 简单条件，直接作为表达式
			return &WhereNode{
				Expr: expr,
			}
		}
	}

	return nil
}

// containsOperator 检查表达式是否包含OR或AND操作符
func containsOperator(expr string) bool {
	upperExpr := strings.ToUpper(expr)
	return strings.Contains(upperExpr, " OR ") || strings.Contains(upperExpr, " AND ")
}

// isOutermostParentheses 检查字符串是否由最外层括号包围
func isOutermostParentheses(expr string) bool {
	if len(expr) < 2 || !strings.HasPrefix(expr, "(") || !strings.HasSuffix(expr, ")") {
		return false
	}

	// 计算括号层数
	count := 0
	for i := 0; i < len(expr); i++ {
		if expr[i] == '(' {
			count++
		} else if expr[i] == ')' {
			count--
			// 如果在最后一个字符之前计数变为0，则不是最外层括号
			if count == 0 && i < len(expr)-1 {
				return false
			}
		}
	}

	return count == 0
}

// parseSubExpression 解析子表达式 (已弃用，使用parseExpressionToWhereNode替代)
func parseSubExpression(expr string) *WhereNode {
	return parseExpressionToWhereNodeWithDepth(expr, 0)
}

// splitByOperator 按操作符分割表达式，考虑括号
func splitByOperator(expr, operator string) []string {
	var parts []string
	start := 0
	parenCount := 0

	// 特殊处理BETWEEN AND的情况
	isBetweenAnd := strings.ToUpper(operator) == " AND " && strings.Contains(strings.ToUpper(expr), " BETWEEN ")

	for i := 0; i < len(expr); i++ {
		char := expr[i]

		if char == '(' {
			parenCount++
		} else if char == ')' {
			parenCount--
		} else if parenCount == 0 && i+len(operator) <= len(expr) {
			// 检查是否匹配操作符（忽略大小写）
			if strings.ToUpper(expr[i:i+len(operator)]) == strings.ToUpper(operator) {
				// 特殊处理：如果是BETWEEN AND表达式，则跳过中间的AND
				if isBetweenAnd {
					// 检查当前位置是否是BETWEEN AND表达式的一部分
					if isBetweenAndOperator(expr, i) {
						// 这是BETWEEN AND表达式中的AND，跳过它
						continue
					}
				}

				// 添加前面的部分
				part := strings.TrimSpace(expr[start:i])
				if part != "" {
					parts = append(parts, part)
				}
				// 更新起始位置
				start = i + len(operator)
				// 跳过操作符长度
				i += len(operator) - 1
			}
		}
	}

	// 添加最后一部分
	part := strings.TrimSpace(expr[start:])
	if part != "" {
		parts = append(parts, part)
	}

	return parts
}

// isBetweenAndOperator 检查指定位置是否是BETWEEN AND表达式中的AND操作符
func isBetweenAndOperator(expr string, andPos int) bool {
	// 确保AND前后有足够的字符
	if andPos < 1 || andPos+len(" AND ") > len(expr) {
		return false
	}

	// 获取AND前的文本
	before := strings.ToUpper(expr[:andPos])

	// 查找最后出现的"BETWEEN"位置
	lastBetweenPos := strings.LastIndex(before, "BETWEEN")
	if lastBetweenPos == -1 {
		return false
	}

	// 检查在"BETWEEN"和"AND"之间是否还有其他操作符
	// 提取BETWEEN之后到AND之前的部分
	betweenAndPart := strings.TrimSpace(before[lastBetweenPos+len("BETWEEN"):])

	// 在这部分中不应该有其他操作符（AND/OR）
	// 如果这部分只包含数字或其他非操作符内容，则这是一个有效的BETWEEN AND结构
	hasAnd := strings.Contains(betweenAndPart, " AND ")
	hasOr := strings.Contains(betweenAndPart, " OR ")

	// 如果在BETWEEN和AND之间没有其他操作符，则这是一个BETWEEN AND结构
	return !hasAnd && !hasOr
}

// ConvertGenericToYAML converts generic columns and rules to YAML data structure
func ConvertGenericToYAML(tableName string, columns []GenericColumn, queryRules []GenericRule, primaryKeys []PrimaryKeyInfo, constraints []ConstraintInfo, indexes []IndexInfo) (*YamlDataConfig, error) {
	// 创建数据库表结构
	dbTable := DatabaseTable{
		TableName:   strings.ToLower(tableName),
		Columns:     orderedmap.New(),
		PrimaryKeys: make([]PrimaryKey, 0),
		Indexes:     Indexes{},
	}

	// 转换列定义
	for _, col := range columns {
		// 映射数据库类型到Go类型
		// goType := mapDBTypeToGoType(col.Type)
		goType, _ := render.Imports(col.Type, col.Type)

		// columnKey := strings.ToLower(col.Name)
		columnKey := strings.ToUpper(col.Name)
		column := Column{
			DbColumn:      col.Name,
			GoType:        goType,
			Comment:       col.Description,
			NotNull:       col.NotNull == "Y",
			Length:        col.Length,
			DefaultValue:  col.DefaultValue,
			AutoIncrement: col.AutoIncrement == "Y",
			PrimaryKey:    col.IsPrimaryKey,
			Description:   col.Tag,
		}

		// 使用orderedmap设置列，保持插入顺序
		dbTable.Columns.Set(columnKey, column)
	}

	// 处理主键信息
	for _, pk := range primaryKeys {
		dbTable.PrimaryKeys = append(dbTable.PrimaryKeys, PrimaryKey{
			Column: strings.ToLower(pk.Column),
		})
	}

	// 处理索引信息
	var generalIndexes []Index
	var uniqueIndexes []Index

	// 处理约束（作为唯一索引）
	for _, constraint := range constraints {
		var lowerColumns []string
		for _, col := range constraint.Columns {
			lowerColumns = append(lowerColumns, strings.ToLower(col))
		}

		uniqueIndexes = append(uniqueIndexes, Index{
			IndexName: constraint.Name,
			Columns:   lowerColumns,
		})
	}

	// 处理索引
	for _, idx := range indexes {
		var lowerColumns []string
		for _, col := range idx.Columns {
			lowerColumns = append(lowerColumns, strings.ToLower(col))
		}

		generalIndexes = append(generalIndexes, Index{
			IndexName: idx.Name,
			Columns:   lowerColumns,
		})
	}

	// 设置索引
	dbTable.Indexes = Indexes{
		General: generalIndexes,
		Unique:  uniqueIndexes,
	}

	// 创建查询规则
	selfQueryRules := make(map[string]QueryRule)

	// 转换查询规则
	for _, rule := range queryRules {
		if rule.Name != "" {
			queryRule := QueryRule{
				SelectFields: rule.SelectFields,
				Page:         rule.Page,
			}

			// 添加where子句
			// 使用改进的解析器处理复杂表达式
			// if len(rule.Conditions) > 0 {
			// 	// 我们假设第一个条件包含了完整的表达式（这是Excel模板的设计方式）
			// 	if len(rule.Conditions) > 0 && rule.Conditions[0].Expr != "" {
			// 		whereNode := parseExpressionToWhereNode(rule.Conditions[0].Expr)
			// 		if whereNode != nil {
			queryRule.Where = rule.Conditions
			// 		}
			// 	}
			// }

			// 添加聚合函数
			if rule.Aggregation != "" {
				queryRule.Aggregation = &Aggregation{
					Function:   rule.Aggregation,
					ResultType: rule.AggregationType, // 默认结果类型
				}
			}

			selfQueryRules[rule.Name] = queryRule
		}
	}

	// 将查询规则转换为yaml.Node以保持顺序
	var yamlNode yaml.Node
	// 先将map转换为yaml.Node
	mapNode := make(map[string]interface{})
	for k, v := range selfQueryRules {
		mapNode[k] = v
	}

	// 创建yaml.Node
	yamlNode.Kind = yaml.MappingNode
	yamlNode.Tag = "!!map"

	// 为每个键值对创建节点
	for name, rule := range selfQueryRules {
		// 键节点
		keyNode := yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: name,
		}

		// 值节点
		valueBytes, err := yaml.Marshal(rule)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal query rule %s: %v", name, err)
		}

		var valueNode yaml.Node
		if err := yaml.Unmarshal(valueBytes, &valueNode); err != nil {
			return nil, fmt.Errorf("failed to unmarshal query rule %s: %v", name, err)
		}

		// 将键值对添加到父节点
		yamlNode.Content = append(yamlNode.Content, &keyNode, valueNode.Content[0])
	}

	// 创建完整的配置
	config := &YamlDataConfig{
		DatabaseTable:  dbTable,
		SelfQueryRules: yamlNode,
	}

	return config, nil
}

// mapDBTypeToGoType maps database types to Go types
// func mapDBTypeToGoType(dbType string) string {
// 	switch strings.ToLower(dbType) {
// 	case "int64":
// 		return "int64"
// 	case "string":
// 		return "string"
// 	case "time":
// 		return "time.Time"
// 	case "date":
// 		return "time.Time"
// 	default:
// 		return "string" // 默认使用string类型
// 	}
// }
