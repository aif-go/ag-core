package conditonwhere

import (
	"errors"
	"fmt"
	"strings"
	"sync"
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
		return convertLikeWildcards(condition.Expr), nil
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
// WhereConfigMap 全局存储方法名到 whereData 的映射
// Key: 方法名 (如 "FindById", "FindByOrder")
// Value: whereData (MaskWhereCondition 结构)
var WhereConfigMap = struct {
	sync.RWMutex
	data map[string]*MaskWhereCondition
}{
	data: make(map[string]*MaskWhereCondition),
}

// RegisterWhereConfig 注册 whereData 配置
// methodName: 方法名
// whereData: where 条件配置 (MaskWhereCondition 结构)
func RegisterWhereConfig(methodName string, whereData *MaskWhereCondition) {
	WhereConfigMap.Lock()
	defer WhereConfigMap.Unlock()
	WhereConfigMap.data[methodName] = whereData
}

// GetWhereConfig 获取 whereData 配置 (只读)
func GetWhereConfig(methodName string) (*MaskWhereCondition, bool) {
	WhereConfigMap.RLock()
	defer WhereConfigMap.RUnlock()
	data, ok := WhereConfigMap.data[methodName]
	return data, ok
}

// FieldMask 字段标记，用于标记哪些字段被设置了
// 用于动态 SQL 条件过滤，类似 MyBatis 的动态条件
type FieldMask struct {
	fields map[string]bool
}

// NewFieldMask 创建 FieldMask
func NewFieldMask() *FieldMask {
	return &FieldMask{
		fields: make(map[string]bool),
	}
}

// Set 设置字段标记
func (m *FieldMask) Set(field string) {
	m.fields[field] = true
}

// IsSet 检查字段是否被设置
func (m *FieldMask) IsSet(field string) bool {
	return m.fields[field]
}

// GetFields 获取所有被设置的字段
func (m *FieldMask) GetFields() []string {
	fields := make([]string, 0, len(m.fields))
	for field := range m.fields {
		fields = append(fields, field)
	}
	return fields
}

// Merge 合并另一个 FieldMask
func (m *FieldMask) Merge(other *FieldMask) {
	if other == nil {
		return
	}
	for field := range other.fields {
		m.fields[field] = true
	}
}

// Reset 重置所有字段标记
func (m *FieldMask) Reset() {
	m.fields = make(map[string]bool)
}

// Size 返回已设置字段的数量
func (m *FieldMask) Size() int {
	return len(m.fields)
}

// FilterWhereConditions 根据 FieldMask 过滤 WHERE 条件表达式列表
// conditions: 条件表达式列表，如 ["ORDER_ID = @OrderId", "MERCHANT_ID = @MerchantId"]
// fieldMask: 字段标记
// 返回: 过滤后的条件列表
// func FilterWhereConditions(conditions []string, fieldMask *FieldMask) []string {
// 	if fieldMask == nil || len(fieldMask.fields) == 0 {
// 		// 如果没有设置任何字段，返回所有条件
// 		return conditions
// 	}

// 	filtered := make([]string, 0, len(conditions))
// 	for _, cond := range conditions {
// 		// 从表达式中提取参数名
// 		paramName := extractParamName(cond)
// 		if paramName == "" {
// 			// 没有参数，是字面量条件，无条件保留
// 			filtered = append(filtered, cond)
// 			continue
// 		}
// 		// 检查字段是否被设置
// 		if fieldMask.IsSet(paramName) {
// 			filtered = append(filtered, cond)
// 		}
// 	}
// 	return filtered
// }

// extractParamName 从条件表达式中提取参数名
// 例如: "ORDER_ID = @OrderId" -> "OrderId"
func extractParamName(expr string) string {
	// 查找 @ 符号的位置
	idx := strings.Index(expr, "@")
	if idx == -1 {
		return ""
	}
	// 提取 @ 后面的部分
	paramName := expr[idx+1:]
	// 如果参数名后面还有其他字符（如空格或右括号），截断
	for i, c := range paramName {
		if c == ' ' || c == ')' || c == '%' || c == '_' {
			return paramName[:i]
		}
	}
	return paramName
}

// BuildWhereFromConfig 根据 methodName 从全局配置中获取 whereData，过滤后生成 SQL
// methodName: 方法名，用于从 WhereConfigMap 中获取对应的 whereData
// 返回: 过滤后的 WHERE SQL 子句，错误信息
func (m *FieldMask) BuildWhereFromConfig(methodName string, conditionMap map[string]*MaskWhereCondition) (string, error) {
	// 1. 从全局配置中获取 whereData
	whereData, ok := conditionMap[methodName]
	if !ok {
		return "", fmt.Errorf("未找到方法%s对应的sql条件", methodName) // 没有配置，返回空
	}

	// 2. 根据 FieldMask 过滤 MaskWhereCondition，生成新的 MaskWhereCondition
	filteredCondition := m.FilterMaskWhereCondition(whereData)
	if filteredCondition == nil {
		return "", fmt.Errorf("未设置方法%s对应的sql条件的参数值", methodName) // 过滤后没有条件，返回空
	}

	// 3. 将过滤后的 MaskWhereCondition 转换为 SQL
	sql,err := GenerateWhereSQL(filteredCondition)
	if err != nil {
		return "", err
	}

	return "WHERE " + sql, nil
}

// FilterMaskWhereCondition 根据 FieldMask 过滤 MaskWhereCondition，生成新的 MaskWhereCondition
// 返回: 过滤后的 MaskWhereCondition，如果所有条件都被过滤则返回 nil
func (m *FieldMask) FilterMaskWhereCondition(condition *MaskWhereCondition) *MaskWhereCondition {
	if m == nil {
		return nil
	}

	if condition == nil {
		return nil
	}

	// 如果有 expr，检查是否需要保留
	if condition.Expr != "" {
		paramName := extractParamName(condition.Expr)
		if paramName != "" {
			// 有参数名，检查是否在 FieldMask 中
			if !m.IsSet(paramName) {
				return nil // 参数未设置，过滤掉
			}
		}
		// 没有参数名（常量值），保留
		return &MaskWhereCondition{
			Operator: condition.Operator,
			Expr:     condition.Expr,
		}
	}

	// 处理嵌套的 conditions
	if len(condition.Conditions) == 0 {
		return nil
	}

	filteredConditions := make([]MaskWhereCondition, 0, len(condition.Conditions))
	for _, cond := range condition.Conditions {
		filtered := m.FilterMaskWhereCondition(&cond)
		if filtered != nil {
			filteredConditions = append(filteredConditions, *filtered)
		}
	}

	// 如果所有子条件都被过滤掉了，返回 nil
	if len(filteredConditions) == 0 {
		return nil
	}

	// 返回新的 MaskWhereCondition
	return &MaskWhereCondition{
		Operator:   condition.Operator,
		Conditions: filteredConditions,
	}
}

// BuildWhereSQL 根据条件表达式列表构建 WHERE 子句
// conditions: 条件表达式列表，如 ["ORDER_ID = @OrderId", "MERCHANT_ID = @MerchantId"]
// operator: 逻辑操作符 (AND/OR)
// 返回: SQL WHERE 子句字符串
func BuildWhereSQL(conditions []string, operator string) string {
	if len(conditions) == 0 {
		return "WHERE 1=1"
	}

	if len(conditions) == 1 {
		return "WHERE (" + conditions[0] + ")"
	}

	return "WHERE (" + strings.Join(conditions, " "+operator+" ") + ")"
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

// ValidateLeadingCol 验证SQL中是否包含索引前导列，避免全表扫描
func ValidateLeadingCol(sql string, leadingCols []string) bool {
	if len(leadingCols)==0 {
		return false // 没有配置索引前导列，无法验证
	}
	
	// Implementation for validating leading columns in SQL
	for _, col := range leadingCols {
		if strings.Contains(sql, col) {
			return true
		}
	}
	return false
}

// convertLikeWildcards 将表达式中 @Param%_ / %_@Param / %_@Param%_ 转为 CONCAT 格式
func convertLikeWildcards(expr string) string {
	var buf strings.Builder
	lastEnd := 0
	for {
		atIdx := strings.Index(expr[lastEnd:], "@")
		if atIdx == -1 {
			break
		}
		atIdx += lastEnd

		preStart := atIdx
		for preStart > 0 && (expr[preStart-1] == '%' || expr[preStart-1] == '_') {
			preStart--
		}
		prefixWild := expr[preStart:atIdx]

		rest := expr[atIdx+1:]
		nameEnd := 0
		for nameEnd < len(rest) {
			c := rest[nameEnd]
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
				(c >= '0' && c <= '9') || c == '_') {
				break
			}
			nameEnd++
		}
		if nameEnd == 0 {
			buf.WriteString(expr[lastEnd : atIdx+1])
			lastEnd = atIdx + 1
			continue
		}

		paramName := rest[:nameEnd]
		wcEnd := nameEnd
		for wcEnd < len(rest) && (rest[wcEnd] == '%' || rest[wcEnd] == '_') {
			wcEnd++
		}
		suffixWild := rest[nameEnd:wcEnd]

		if prefixWild == "" && suffixWild == "" {
			buf.WriteString(expr[lastEnd : atIdx+1])
			lastEnd = atIdx + 1
			continue
		}

		buf.WriteString(expr[lastEnd:preStart])
		buf.WriteString("CONCAT(")

		if prefixWild != "" {
			buf.WriteByte('\'')
			buf.WriteString(prefixWild)
			buf.WriteString("', ")
		}

		buf.WriteByte('@')
		buf.WriteString(paramName)

		if suffixWild != "" {
			buf.WriteString(", '")
			buf.WriteString(suffixWild)
			buf.WriteByte('\'')
		}

		buf.WriteByte(')')
		lastEnd = atIdx + 1 + len(paramName) + len(suffixWild)
	}
	if lastEnd == 0 {
		return expr
	}
	buf.WriteString(expr[lastEnd:])
	return buf.String()
}
