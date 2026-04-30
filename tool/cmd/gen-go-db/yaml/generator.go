package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ag-core/tool/cmd/gen-go-db/excel"
	"ag-core/tool/cmd/gen-go-db/utils"

	"gopkg.in/yaml.v2"
)

// printCondition 递归打印条件内容
func printCondition(cond *excel.Condition, depth int) {
	// 生成缩进
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	// 打印条件信息
	if cond.Expr != "" {
		fmt.Printf("%sExpr: %s\n", indent, cond.Expr)
	}
	if cond.Operator != "" {
		fmt.Printf("%sOperator: %s\n", indent, cond.Operator)
		if len(cond.Conditions) > 0 {
			fmt.Printf("%sConditions:\n", indent)
			for i, subCond := range cond.Conditions {
				fmt.Printf("%s  Condition %d:\n", indent, i+1)
				printCondition(subCond, depth+2)
			}
		}
	}
}

// generateWhereCondition 递归生成where条件
func generateWhereCondition(cond *excel.Condition) yaml.MapSlice {
	condData := yaml.MapSlice{}
	if cond.Expr != "" {
		condData = append(condData, yaml.MapItem{Key: "expr", Value: cond.Expr})
	}
	if cond.Operator != "" {
		condData = append(condData, yaml.MapItem{Key: "operator", Value: cond.Operator})
		if len(cond.Conditions) > 0 {
			subConditions := []yaml.MapSlice{}
			for _, subCond := range cond.Conditions {
				subConditions = append(subConditions, generateWhereCondition(subCond))
			}
			condData = append(condData, yaml.MapItem{Key: "conditions", Value: subConditions})
		}
	}
	return condData
}

// GenerateYAMLFromExcel 从Excel文件生成YAML文件
func GenerateYAMLFromExcel(inputFile string, outputDir string, testMode bool, tableName string) error {
	// 确保输出目录存在，在用户输入的基础上拼接repository/yaml
	yamlOutputDir := outputDir
	if yamlOutputDir != "" {
		// 判断用户输入的是否以/结尾
		if !strings.HasSuffix(yamlOutputDir, "/") {
			yamlOutputDir += "/"
		}
		yamlOutputDir += "repository/yaml"
	} else {
		// 如果用户没有指定输出目录，使用默认的repository/yaml
		yamlOutputDir = "repository/yaml"
	}
	if err := os.MkdirAll(yamlOutputDir, 0755); err != nil {
		return err
	}

	// 处理Excel文件或生成测试数据
	var tables map[string]*excel.ExcelInfo
	var err error

	if testMode {
		fmt.Println("测试模式：生成示例YAML文件")
		tables = generateTestData()
	} else {
		// 检查输入文件是否存在
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return fmt.Errorf("输入文件不存在: %s", inputFile)
		}

		fmt.Printf("正在处理Excel文件: %s\n", inputFile)
		tables, err = excel.ParseExcel(inputFile)
		if err != nil {
			return fmt.Errorf("解析Excel文件失败: %v", err)
		}
	}

	// 生成YAML文件
	fmt.Printf("正在生成YAML文件到: %s\n", yamlOutputDir)

	// 解析表名列表，支持多个表名以逗号分隔
	tableNames := utils.ParseCommaSeparatedList(tableName)

	for sheetName, table := range tables {
		// 如果指定了表名，则只生成指定表的yaml文件
		if len(tableNames) > 0 {
			// 检查当前表名是否在指定的表名列表中
			if !utils.ContainsIgnoreCase(tableNames, sheetName) {
				continue
			}
		}
		yamlPath := filepath.Join(yamlOutputDir, sheetName+".yaml")
		if err := GenerateYAML(table, yamlPath); err != nil {
			return fmt.Errorf("生成YAML文件失败: %v", err)
		}
		fmt.Printf("生成成功: %s\n", yamlPath)
	}

	return nil
}

// generateTestData 生成测试数据
func generateTestData() map[string]*excel.ExcelInfo {
	tables := make(map[string]*excel.ExcelInfo)

	// 创建测试表
	table := &excel.ExcelInfo{
		Name:        "TM_TEST",
		Columns:     []*excel.ColumnInfo{},
		PrimaryKey:  []string{"SEQ"},
		Constraints: []*excel.ConstraintInfo{},
		Indexes:     []*excel.IndexInfo{},
		SelfQueries: make(map[string]*excel.SelfQueryInfo),
	}

	// 添加列信息
	columns := []*excel.ColumnInfo{
		{
			Name:          "SEQ",
			Type:          "int64",
			Length:        "",
			NotNull:       true,
			Default:       "",
			AutoIncrement: true,
			SupportUpdate: false,
			Description:   "",
			Tag:           "",
		},
		{
			Name:          "SEX",
			Type:          "int",
			Length:        "",
			NotNull:       false,
			Default:       "1",
			AutoIncrement: false,
			SupportUpdate: true,
			Description:   "1-男 2-女",
			Tag:           "///@omitempty",
		},
		{
			Name:          "NAME",
			Type:          "string",
			Length:        "20",
			NotNull:       true,
			Default:       "",
			AutoIncrement: false,
			SupportUpdate: true,
			Description:   "卡号",
			Tag:           "",
		},
		{
			Name:          "PHONE",
			Type:          "string",
			Length:        "11",
			NotNull:       false,
			Default:       "",
			AutoIncrement: false,
			SupportUpdate: true,
			Description:   "",
			Tag:           "",
		},
		{
			Name:          "ADDRESS",
			Type:          "string",
			Length:        "100",
			NotNull:       true,
			Default:       "",
			AutoIncrement: false,
			SupportUpdate: true,
			Description:   "应用编号",
			Tag:           "",
		},
		{
			Name:          "JPA_VERSION",
			Type:          "int",
			Length:        "",
			NotNull:       true,
			Default:       "",
			AutoIncrement: false,
			SupportUpdate: false,
			Description:   "JPA_VERSION",
			Tag:           "",
		},
		{
			Name:          "CREATED_TIME",
			Type:          "time",
			Length:        "",
			NotNull:       false,
			Default:       "",
			AutoIncrement: false,
			SupportUpdate: false,
			Description:   "创建时间",
			Tag:           "///@create",
		},
		{
			Name:          "LAST_MODIFIED_TIME",
			Type:          "time",
			Length:        "",
			NotNull:       false,
			Default:       "",
			AutoIncrement: false,
			SupportUpdate: false,
			Description:   "最后更新时间",
			Tag:           "///@update",
		},
	}
	table.Columns = columns

	// 添加约束
	constraints := []*excel.ConstraintInfo{
		{
			Name:    "TM_TEST_UNIUQE_1",
			Columns: []string{"PHONE"},
		},
	}
	table.Constraints = constraints

	// 添加索引
	indexes := []*excel.IndexInfo{
		{
			Name:    "INDEX1_TM_TEST",
			Columns: []string{"NAME", "ADDRESS"},
		},
	}
	table.Indexes = indexes

	// 添加自定义查询
	selfQueries := map[string]*excel.SelfQueryInfo{
		"Xxxxx": {
			SelectFields: "SEQ,SEX,NAME,PHONE,ADDRESS",
			Where: &excel.WhereClause{
				Operator: "AND",
				Conditions: []*excel.Condition{
					{
						Expr: "PHONE = @Phone",
					},
				},
			},
			Page: false,
		},
		"FindAllColsByPage": {
			SelectFields: "*",
			Where: &excel.WhereClause{
				Operator: "AND",
				Conditions: []*excel.Condition{
					{
						Expr: "ADDRESS = @Address",
					},
				},
			},
			Page: true,
		},
		"FindAaByPage": {
			SelectFields: "SEQ,SEX,NAME,PHONE,ADDRESS",
			Where: &excel.WhereClause{
				Operator: "AND",
				Conditions: []*excel.Condition{
					{
						Expr: "ADDRESS = @Address",
					},
				},
			},
			Page: true,
		},
		"FindAllCols": {
			SelectFields: "*",
			Where: &excel.WhereClause{
				Operator: "AND",
				Conditions: []*excel.Condition{
					{
						Expr: "ADDRESS = @Address",
					},
				},
			},
			Page: false,
		},
	}
	table.SelfQueries = selfQueries

	tables[table.Name] = table
	return tables
}

// GenerateYAML 生成YAML文件
func GenerateYAML(table *excel.ExcelInfo, outputPath string) error {
	// 使用yaml.MapSlice来保持字段的固定顺序
	data := yaml.MapSlice{}

	// 在最上方添加table_name属性
	data = append(data, yaml.MapItem{Key: "table_name", Value: table.Name})

	// 表基本信息 - 使用固定顺序
	// tableData := yaml.MapSlice{}
	// tableData = append(tableData, yaml.MapItem{Key: "name", Value: table.Name})
	// data = append(data, yaml.MapItem{Key: "table", Value: tableData})

	// 列信息
	columns := []yaml.MapSlice{}
	for _, col := range table.Columns {
		// 为列属性定义固定顺序
		columnData := yaml.MapSlice{}

		// 按照固定顺序添加非空值
		columnData = append(columnData, yaml.MapItem{Key: "name", Value: col.Name})
		columnData = append(columnData, yaml.MapItem{Key: "type", Value: col.Type})
		if col.Length != "" {
			columnData = append(columnData, yaml.MapItem{Key: "length", Value: col.Length})
		}
		columnData = append(columnData, yaml.MapItem{Key: "not_null", Value: col.NotNull})
		if col.Default != "" {
			columnData = append(columnData, yaml.MapItem{Key: "default", Value: col.Default})
		}
		columnData = append(columnData, yaml.MapItem{Key: "auto_increment", Value: col.AutoIncrement})
		columnData = append(columnData, yaml.MapItem{Key: "support_update", Value: col.SupportUpdate})
		if col.Description != "" {
			columnData = append(columnData, yaml.MapItem{Key: "description", Value: col.Description})
		}
		if col.Tag != "" {
			columnData = append(columnData, yaml.MapItem{Key: "tag", Value: col.Tag})
		}

		columns = append(columns, columnData)
	}
	data = append(data, yaml.MapItem{Key: "columns", Value: columns})

	// 主键信息
	if len(table.PrimaryKey) > 0 {
		data = append(data, yaml.MapItem{Key: "primary_key", Value: table.PrimaryKey})
	}

	// 约束信息
	if len(table.Constraints) > 0 {
		constraints := []yaml.MapSlice{}
		for _, cons := range table.Constraints {
			// 为约束属性定义固定顺序
			constraintData := yaml.MapSlice{}
			constraintData = append(constraintData, yaml.MapItem{Key: "name", Value: cons.Name})
			constraintData = append(constraintData, yaml.MapItem{Key: "columns", Value: cons.Columns})
			constraints = append(constraints, constraintData)
		}
		data = append(data, yaml.MapItem{Key: "constraints", Value: constraints})
	}

	// 索引信息
	if len(table.Indexes) > 0 {
		indexes := []yaml.MapSlice{}
		for _, idx := range table.Indexes {
			// 为索引属性定义固定顺序
			indexData := yaml.MapSlice{}
			indexData = append(indexData, yaml.MapItem{Key: "name", Value: idx.Name})
			indexData = append(indexData, yaml.MapItem{Key: "columns", Value: idx.Columns})
			indexes = append(indexes, indexData)
		}
		data = append(data, yaml.MapItem{Key: "indexes", Value: indexes})
	}

	// 自定义查询规则
	if len(table.SelfQueries) > 0 {
		// 获取所有查询名称并排序，保证生成顺序固定
		queryNames := make([]string, 0, len(table.SelfQueries))
		for name := range table.SelfQueries {
			queryNames = append(queryNames, name)
		}
		sort.Strings(queryNames)

		// 按照排序后的顺序生成YAML
		selfQueries := yaml.MapSlice{}
		for _, queryName := range queryNames {
			query := table.SelfQueries[queryName]
			// 为查询属性定义固定顺序
			queryData := yaml.MapSlice{}
			
			if query.SelectFields != "" {
				queryData = append(queryData, yaml.MapItem{Key: "select_fields", Value: query.SelectFields})
			}
			queryData = append(queryData, yaml.MapItem{Key: "page", Value: query.Page})
			// 是否动态模版
			queryData = append(queryData, yaml.MapItem{Key: "dynamic_sql", Value: query.DynamicTemplate})
			queryData = append(queryData, yaml.MapItem{Key: "sql_template", Value: query.SqlTemplate})
			if query.WhereParams !=nil{
				queryData = append(queryData, yaml.MapItem{Key: "Where_params", Value: query.WhereParams})
			}
			// 添加WHERE子句
			if query.Where != nil {
				// 打印where条件的内容，用于调试
				fmt.Printf("=== 打印where条件内容 ===\n")
				fmt.Printf("Query name: %s\n", queryName)
				fmt.Printf("Where operator: %s\n", query.Where.Operator)
				fmt.Printf("Where conditions count: %d\n", len(query.Where.Conditions))
				for _, cond := range query.Where.Conditions {
					printCondition(cond, 1)
				}
				fmt.Printf("=========================\n\n")

				whereData := yaml.MapSlice{}

				if query.Where.Operator != "" {
					whereData = append(whereData, yaml.MapItem{Key: "operator", Value: query.Where.Operator})
				}

				if len(query.Where.Conditions) > 0 {
					conditions := []yaml.MapSlice{}
					for _, cond := range query.Where.Conditions {
						conditions = append(conditions, generateWhereCondition(cond))
					}
					whereData = append(whereData, yaml.MapItem{Key: "conditions", Value: conditions})
				}

				if len(whereData) > 0 {
					queryData = append(queryData, yaml.MapItem{Key: "where", Value: whereData})
				}
			}

			if len(queryData) > 0 {
				selfQueries = append(selfQueries, yaml.MapItem{Key: queryName, Value: queryData})
			}
		}

		if len(selfQueries) > 0 {
			data = append(data, yaml.MapItem{Key: "self_query_rules", Value: selfQueries})
		}
	}

	// 转换为YAML格式
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(outputPath, yamlData, 0644)
}
