package conditonwhere

import (
	"strings"
)

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

func (m *FieldMask) NhWhere(where string) string{
	// where = "a = @a and (b = @b or c = @c) or d = @d"
	// 提取参数名 a,b,c,d
	// 判断 a,b,c,d 是否在 m.fields 中,m.fields是一个map结构，如果在则保留条件，否则删除条件
	// 如果 后面的值是常量值 比如 d = 1 则直接保留条件
	// 最终返回过滤后的 where 条件
	return ""
	// 解析 where 条件，提取参数名
	// 判断参数名是否在 m.fields 中，如果在则保留条件，否则删除条件
	// 构建新的 where 条件并返回
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
func FilterWhereConditions(conditions []string, fieldMask *FieldMask) []string {
	if fieldMask == nil || len(fieldMask.fields) == 0 {
		// 如果没有设置任何字段，返回所有条件
		return conditions
	}

	filtered := make([]string, 0, len(conditions))
	for _, cond := range conditions {
		// 从表达式中提取参数名
		paramName := extractParamName(cond)
		if paramName == "" {
			// 没有参数，是字面量条件，无条件保留
			filtered = append(filtered, cond)
			continue
		}
		// 检查字段是否被设置
		if fieldMask.IsSet(paramName) {
			filtered = append(filtered, cond)
		}
	}
	return filtered
}

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
		if c == ' ' || c == ')' {
			return paramName[:i]
		}
	}
	return paramName
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
