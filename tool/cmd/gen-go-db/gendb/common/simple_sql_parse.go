package common

import (
	"fmt"
	"regexp"
	"strings"

	// "ag-core/contribute/agdb/gormdb"
	"ag-core/tool/cmd/gen-go-db/gendb/render"
	"ag-core/tool/cmd/gen-go-db/gendb/util"

	"github.com/blastrain/vitess-sqlparser/sqlparser"
)

// 解析WHERE条件并提取参数
func ParseWhereConditions(sql string) ([]render.SqlParameter, error) {

	_, paramters := util.ReplaceNamedParamsWithIn(sql)
	sqlParamMap := map[string]string{}
	for _, paramName := range paramters {
		sqlParamMap[strings.ToLower(paramName)] = paramName
	}
	// 1. 提取WHERE子句
	sql = strings.TrimSpace(sql)
	// 先把所有的sql转小写
	whereIndex := strings.Index(strings.ToLower(sql), "where")
	if whereIndex == -1 {
		return []render.SqlParameter{}, nil // 无WHERE条件
	}
	whereClause := strings.TrimSpace(sql[whereIndex+5:])
	if whereClause == "" {
		return []render.SqlParameter{}, nil
	}

	// 2. 清理括号（简化处理）
	whereClause = strings.ReplaceAll(whereClause, "(", " ")
	whereClause = strings.ReplaceAll(whereClause, ")", " ")

	// 3. 拆分条件（按AND/OR拆分）不区分and 或者 or的大小写，通用匹配
	condRegex := regexp.MustCompile(`(?i)\s+and\s+|\s+or\s+`)
	conditions := condRegex.Split(strings.ToLower(whereClause), -1)

	var params []render.SqlParameter

	// 4. 正则匹配列名、变量和运算符
	// 匹配模式：列名 运算符 变量（支持=、IN、NOT IN）
	pattern := `(?i)(\w+)\s+(=|in|not\s+in)\s+(@\w+)`
	re := regexp.MustCompile(pattern)

	for _, cond := range conditions {
		cond = strings.TrimSpace(cond)
		if cond == "" {
			continue
		}

		// 匹配条件
		matches := re.FindStringSubmatch(cond)
		if len(matches) != 4 {
			continue // 不匹配的条件跳过
		}

		colName := matches[1]                     // 列名
		operator := strings.TrimSpace(matches[2]) // 运算符
		varName := matches[3]                     // 变量名

		// 判断是否为IN/NOT IN（需要切片）
		isSlice := operator == "in" || operator == "not in"

		params = append(params, render.SqlParameter{
			ColName:       strings.ToUpper(colName),
			ParameterName: sqlParamMap[strings.Replace(varName, "@", "", -1)],
			IsSlice:       isSlice,
		})
	}

	return params, nil
}

// Column 存储解析出的列信息：原始表达式和别名
// type Column struct {
// 	Original string // 原始列表达式（如 "app_id"、"count(*)、"t.name"）
// 	Alias    string // 别名（AS 后的名称，无别名则为空）
// }

// ParseSqlSelect 解析查询sql的select 部分，从而支持自定义列的查询返回
func ParseSqlSelect(sql string) ([]*render.SelectColumn, error) {

	// select * |SELECT * 就用数据库实体接收返回的参数，其余根据客户定义的查询的列构建返回实体
	if strings.HasPrefix(sql, "select *") || strings.HasPrefix(sql, "SELECT *") {
		return nil, nil
	}
	// 1. 解析 SQL 语句
	stmt, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("SQL 解析失败: %v", err)
	}

	// 2. 确认是 SELECT 语句
	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		return nil, nil
	}

	// 3. 提取 SELECT 子句中的列信息
	var columns []*render.SelectColumn
	for _, expr := range selectStmt.SelectExprs {
		var original, alias string
		switch e := expr.(type) {
		case *sqlparser.AliasedExpr: // 聚合的动作必须包含as
			// 带别名的列（如 "app_id AS a" 或 "app_id a"）
			original = sqlparser.String(e.Expr) // 原始表达式（如 "app_id"、"count(*)"）
			alias = sqlparser.String(e.As)      // 别名（AS 后的名称）
			if alias == "" {
				alias = ToCamelCase(original)
			}
		default:
			// 无别名的列（如 "cardno"、"t.age"）
			original = sqlparser.String(expr)
			alias = ToCamelCase(original)
		}
		// if isAggregateFunction(original){
		// 	inferReturnType(original)
		// }
		columns = append(columns, &render.SelectColumn{
			ColumnName: original,
			Alias:      alias,
		})
	}

	return columns, nil
}

// 判断表达式是否为聚合函数（简单匹配函数名前缀）
// func isAggregateFunction(expr string) bool {
// 	aggFunctions := []string{"count(", "sum(", "avg(", "max(", "min("}
// 	for _, f := range aggFunctions {
// 		if strings.HasPrefix(strings.ToLower(expr), f) {
// 			return true
// 		}
// 	}
// 	return false
// }

// inferReturnType 推断聚合函数的返回类型
// func inferReturnType(funcName string, expr string) string {
// 	switch funcName {
// 	case "count":
// 		return "int64" // count 始终返回整数
// 	case "sum":
// 		// 简化：默认返回 int64（实际可根据参数类型细化，如参数是 float 则返回 float64）
// 		return "int64"
// 	case "avg":
// 		return "float64" // avg 通常返回浮点数
// 	case "max", "min":
// 		// 简化：无法确定时返回 interface{}，实际需结合参数类型（如日期、字符串、数值）
// 		return "interface{}"
// 	default:
// 		return "interface{}"
// 	}
// }
