// origin: github.com/aif-go/ag-core/contribute/agdb/conditonwhere/fieldmask.go
// 从 contribute/agdb 裁剪，仅保留 gen-go-db 编译依赖的部分
// 审计时 diff 对应文件即可
package conditonwhere

import (
	"errors"
	"strings"
)

// WhereCondition where条件，支持嵌套结构
type MaskWhereCondition struct {
	Operator   string           `json:"operator"`
	Conditions []MaskWhereCondition `json:"conditions"`
	Expr       string           `json:"expr"`
}


// GenerateWhereSQL 生成where条件SQL语句
func GenerateWhereSQL(condition *MaskWhereCondition) (string, error) {
	if condition.Expr != "" {
		return condition.Expr, nil
	}

	if len(condition.Conditions) == 0 {
		return "", errors.New("invalid where condition: no expression or sub-conditions")
	}

	conditions := make([]string, 0, len(condition.Conditions))
	for _, cond := range condition.Conditions {
		where,err:=GenerateWhereSQL(&cond)
		if err != nil {
			return "", err
		}
		conditions = append(conditions, where)
	}

	operator := condition.Operator
	if operator == "" {
		operator = "AND"
	}

	return "(" + strings.Join(conditions, " "+operator+" ") + ")",nil
}



// 解析where条件
func ParseWhereCondition(whereData map[interface{}]interface{}) *MaskWhereCondition {
	condition := &MaskWhereCondition{}

	// 解析operator
	if operator, ok := whereData["operator"].(string); ok {
		condition.Operator = operator
	} else {
		condition.Operator = "AND" // 默认使用AND
	}

	// 解析conditions
	if conditionsData, ok := whereData["conditions"].([]interface{}); ok {
		condition.Conditions = make([]MaskWhereCondition, 0, len(conditionsData))
		for _, condData := range conditionsData {
			if condMap, ok := condData.(map[interface{}]interface{}); ok {
				// 检查是嵌套条件还是表达式
				if _, hasExpr := condMap["expr"]; hasExpr {
					// 是表达式
					subCond := MaskWhereCondition{
						Expr: condMap["expr"].(string),
					}
					condition.Conditions = append(condition.Conditions, subCond)
				} else {
					// 是嵌套条件
					subCond := ParseWhereCondition(condMap)
					condition.Conditions = append(condition.Conditions, *subCond)
				}
			}
		}
	}

	return condition
}


