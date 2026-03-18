package conditonwhere

import (
	"fmt"
	"strings"
)

// ============ V2 版本 - 灵活的 WHERE 条件构建器 ============

// SQLOperator 定义支持的SQL操作符
type SQLOperator string

const (
	SQLOpEq        SQLOperator = "="
	SQLOpNeq       SQLOperator = "!="
	SQLOpGt        SQLOperator = ">"
	SQLOpLt        SQLOperator = "<"
	SQLOpGte       SQLOperator = ">="
	SQLOpLte       SQLOperator = "<="
	// SQLOpLike      SQLOperator = "like"
	SQLOpIn        SQLOperator = "in"
	SQLOpNotIn     SQLOperator = "not in"
	SQLOpBetween   SQLOperator = "between"
)

// SQLLogicOperator 定义逻辑连接符
type SQLLogicOperator string

const (
	SQLLogicAnd SQLLogicOperator = "AND"
	SQLLogicOr  SQLLogicOperator = "OR"
)

// WhereCondition 定义用户输入的 WHERE 条件结构
// 支持嵌套条件和多种操作符
type WhereCondition struct {
	Field    string            // 列名
	Operator SQLOperator       // 操作符: =, !=, >, <, >=, <=, like, in, not in, between
	Value    interface{}       // 值（对于 between 和 in 可能是数组）
	Logic    SQLLogicOperator  // 逻辑连接符: AND, OR
	Children []*WhereCondition // 子条件（用于嵌套，如 (a=1 OR b=2)）
}

// WhereClauseBuilder 高性能的 WHERE 条件构建器
type WhereClauseBuilder struct {
	root *WhereCondition
}

// NewWhereClauseBuilder 创建新的 WhereClauseBuilder
func NewWhereClauseBuilder() *WhereClauseBuilder {
	return &WhereClauseBuilder{}
}

// AddCondition 添加一个条件
func (b *WhereClauseBuilder) AddCondition(cond *WhereCondition) *WhereClauseBuilder {
	if b.root == nil {
		b.root = cond
	} else {
		// 将新条件添加到根的子条件中
		b.root.Children = append(b.root.Children, cond)
	}
	return b
}

// AddConditions 批量添加条件
func (b *WhereClauseBuilder) AddConditions(conds ...*WhereCondition) *WhereClauseBuilder {
	for _, cond := range conds {
		b.AddCondition(cond)
	}
	return b
}

// SetRoot 设置根条件
func (b *WhereClauseBuilder) SetRoot(cond *WhereCondition) *WhereClauseBuilder {
	b.root = cond
	return b
}

// ==================== 链式调用方法 ====================

// Eq 添加等于条件
func (b *WhereClauseBuilder) Eq(field string, value interface{}) *WhereClauseBuilder {
	return b.AddCondition(&WhereCondition{
		Field:    field,
		Operator: SQLOpEq,
		Value:    value,
		Logic:    SQLLogicAnd,
	})
}

// Neq 添加不等于条件
func (b *WhereClauseBuilder) Neq(field string, value interface{}) *WhereClauseBuilder {
	return b.AddCondition(&WhereCondition{
		Field:    field,
		Operator: SQLOpNeq,
		Value:    value,
		Logic:    SQLLogicAnd,
	})
}

// Gt 添加大于条件
func (b *WhereClauseBuilder) Gt(field string, value interface{}) *WhereClauseBuilder {
	return b.AddCondition(&WhereCondition{
		Field:    field,
		Operator: SQLOpGt,
		Value:    value,
		Logic:    SQLLogicAnd,
	})
}

// Lt 添加小于条件
func (b *WhereClauseBuilder) Lt(field string, value interface{}) *WhereClauseBuilder {
	return b.AddCondition(&WhereCondition{
		Field:    field,
		Operator: SQLOpLt,
		Value:    value,
		Logic:    SQLLogicAnd,
	})
}

// Gte 添加大于等于条件
func (b *WhereClauseBuilder) Gte(field string, value interface{}) *WhereClauseBuilder {
	return b.AddCondition(&WhereCondition{
		Field:    field,
		Operator: SQLOpGte,
		Value:    value,
		Logic:    SQLLogicAnd,
	})
}

// Lte 添加小于等于条件
func (b *WhereClauseBuilder) Lte(field string, value interface{}) *WhereClauseBuilder {
	return b.AddCondition(&WhereCondition{
		Field:    field,
		Operator: SQLOpLte,
		Value:    value,
		Logic:    SQLLogicAnd,
	})
}

// In 添加 IN 条件
func (b *WhereClauseBuilder) In(field string, values ...interface{}) *WhereClauseBuilder {
	return b.AddCondition(&WhereCondition{
		Field:    field,
		Operator: SQLOpIn,
		Value:    values,
		Logic:    SQLLogicAnd,
	})
}

// NotIn 添加 NOT IN 条件
func (b *WhereClauseBuilder) NotIn(field string, values ...interface{}) *WhereClauseBuilder {
	return b.AddCondition(&WhereCondition{
		Field:    field,
		Operator: SQLOpNotIn,
		Value:    values,
		Logic:    SQLLogicAnd,
	})
}

// Between 添加 BETWEEN 条件
func (b *WhereClauseBuilder) Between(field string, min, max interface{}) *WhereClauseBuilder {
	return b.AddCondition(&WhereCondition{
		Field:    field,
		Operator: SQLOpBetween,
		Value:    []interface{}{min, max},
		Logic:    SQLLogicAnd,
	})
}

// Or 设置下一个条件的逻辑为 OR
func (b *WhereClauseBuilder) Or() *WhereClauseBuilder {
	if b.root != nil && len(b.root.Children) > 0 {
		lastChild := b.root.Children[len(b.root.Children)-1]
		lastChild.Logic = SQLLogicOr
	}
	return b
}

// And 设置下一个条件的逻辑为 AND
func (b *WhereClauseBuilder) And() *WhereClauseBuilder {
	if b.root != nil && len(b.root.Children) > 0 {
		lastChild := b.root.Children[len(b.root.Children)-1]
		lastChild.Logic = SQLLogicAnd
	}
	return b
}

// Group 添加一个条件组（用于嵌套）
func (b *WhereClauseBuilder) Group(conditions ...*WhereCondition) *WhereClauseBuilder {
	if len(conditions) == 0 {
		return b
	}
	return b.AddCondition(&WhereCondition{
		Children: conditions,
	})
}

// AndGroup 添加一个 AND 连接的条件组
func (b *WhereClauseBuilder) AndGroup(conditions ...*WhereCondition) *WhereClauseBuilder {
	if len(conditions) == 0 {
		return b
	}
	for _, c := range conditions {
		c.Logic = SQLLogicAnd
	}
	return b.Group(conditions...)
}

// OrGroup 添加一个 OR 连接的条件组
func (b *WhereClauseBuilder) OrGroup(conditions ...*WhereCondition) *WhereClauseBuilder {
	if len(conditions) == 0 {
		return b
	}
	for _, c := range conditions {
		c.Logic = SQLLogicOr
	}
	return b.Group(conditions...)
}

// Build 构建最终的 WHERE SQL 和参数
// 返回: (whereSQL, args, error)
func (b *WhereClauseBuilder) Build() (string, []interface{}, error) {
	if b.root == nil {
		return "", nil, nil
	}
	
	sql, args, err := buildWhereCondition(b.root)
	if err != nil {
		return "", nil, err
	}
	
	if sql == "" {
		return "", nil, nil
	}
	// 最外层不添加 WHERE 关键字
	return sql, args, nil
}

// buildWhereCondition 递归构建条件
func buildWhereCondition(cond *WhereCondition) (string, []interface{}, error) {
	if cond == nil {
		return "", nil, nil
	}
	
	var sqlParts []string
	var args []interface{}
	
	// 构建当前条件
	if cond.Field != "" {
		conditionSQL, conditionArgs, err := buildSingleCondition(cond)
		if err != nil {
			return "", nil, err
		}
		if conditionSQL != "" {
			sqlParts = append(sqlParts, conditionSQL)
			args = append(args, conditionArgs...)
		}
	}
	
	// 构建子条件（嵌套条件）
	if len(cond.Children) > 0 {
		for i, child := range cond.Children {
			childSQL, childArgs, err := buildWhereCondition(child)
			if err != nil {
				return "", nil, err
			}
			
			if childSQL != "" {
				// 如果已有条件，先添加逻辑连接符
				if len(sqlParts) > 0 {
					logic := string(child.Logic)
					if logic == "" {
						logic = string(SQLLogicAnd) // 默认使用 AND
					}
					sqlParts = append(sqlParts, logic)
				}
				sqlParts = append(sqlParts, childSQL)
				args = append(args, childArgs...)
			}
			
			// 如果不是最后一个子条件，添加逻辑连接符
			if i < len(cond.Children)-1 {
				nextChild := cond.Children[i+1]
				if nextChild.Logic != "" {
					sqlParts = append(sqlParts, string(nextChild.Logic))
				} else {
					sqlParts = append(sqlParts, string(SQLLogicAnd))
				}
			}
		}
	}
	
	if len(sqlParts) == 0 {
		return "", nil, nil
	}
	
	// 如果有多个部分，需要用括号包裹
	sql := strings.Join(sqlParts, " ")
	if strings.Count(sql, " ") > 1 {
		sql = "(" + sql + ")"
	}
	
	return sql, args, nil
}

// buildSingleCondition 构建单个条件的 SQL
func buildSingleCondition(cond *WhereCondition) (string, []interface{}, error) {
	if cond.Field == "" {
		return "", nil, nil
	}
	
	switch cond.Operator {
	case SQLOpEq:
		return fmt.Sprintf("%s = ?", cond.Field), []interface{}{cond.Value}, nil
	case SQLOpNeq:
		return fmt.Sprintf("%s != ?", cond.Field), []interface{}{cond.Value}, nil
	case SQLOpGt:
		return fmt.Sprintf("%s > ?", cond.Field), []interface{}{cond.Value}, nil
	case SQLOpLt:
		return fmt.Sprintf("%s < ?", cond.Field), []interface{}{cond.Value}, nil
	case SQLOpGte:
		return fmt.Sprintf("%s >= ?", cond.Field), []interface{}{cond.Value}, nil
	case SQLOpLte:
		return fmt.Sprintf("%s <= ?", cond.Field), []interface{}{cond.Value}, nil
	// case SQLOpLike:
	// 	return fmt.Sprintf("%s LIKE ?", cond.Field), []interface{}{"%" + fmt.Sprintf("%v", cond.Value) + "%"}, nil
	case SQLOpIn:
		values, ok := cond.Value.([]interface{})
		if !ok {
			return "", nil, fmt.Errorf("IN operator requires []interface{} value")
		}
		if len(values) == 0 {
			return "", nil, nil
		}
		placeholders := make([]string, len(values))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		return fmt.Sprintf("%s IN (%s)", cond.Field, strings.Join(placeholders, ", ")), values, nil
	case SQLOpNotIn:
		values, ok := cond.Value.([]interface{})
		if !ok {
			return "", nil, fmt.Errorf("NOT IN operator requires []interface{} value")
		}
		if len(values) == 0 {
			return "", nil, nil
		}
		placeholders := make([]string, len(values))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		return fmt.Sprintf("%s NOT IN (%s)", cond.Field, strings.Join(placeholders, ", ")), values, nil
	case SQLOpBetween:
		values, ok := cond.Value.([]interface{})
		if !ok || len(values) != 2 {
			return "", nil, fmt.Errorf("BETWEEN operator requires []interface{} with 2 values")
		}
		return fmt.Sprintf("%s BETWEEN ? AND ?", cond.Field), values, nil
	default:
		return "", nil, fmt.Errorf("unsupported operator: %s", cond.Operator)
	}
}

// ==================== 便捷方法 ====================

// ConditionEq 创建等于条件
func ConditionEq(field string, value interface{}) *WhereCondition {
	return &WhereCondition{
		Field:    field,
		Operator: SQLOpEq,
		Value:    value,
		Logic:    SQLLogicAnd,
	}
}

// ConditionNeq 创建不等于条件
func ConditionNeq(field string, value interface{}) *WhereCondition {
	return &WhereCondition{
		Field:    field,
		Operator: SQLOpNeq,
		Value:    value,
		Logic:    SQLLogicAnd,
	}
}

// ConditionGt 创建大于条件
func ConditionGt(field string, value interface{}) *WhereCondition {
	return &WhereCondition{
		Field:    field,
		Operator: SQLOpGt,
		Value:    value,
		Logic:    SQLLogicAnd,
	}
}

// ConditionLt 创建小于条件
func ConditionLt(field string, value interface{}) *WhereCondition {
	return &WhereCondition{
		Field:    field,
		Operator: SQLOpLt,
		Value:    value,
		Logic:    SQLLogicAnd,
	}
}

// ConditionGte 创建大于等于条件
func ConditionGte(field string, value interface{}) *WhereCondition {
	return &WhereCondition{
		Field:    field,
		Operator: SQLOpGte,
		Value:    value,
		Logic:    SQLLogicAnd,
	}
}

// ConditionLte 创建小于等于条件
func ConditionLte(field string, value interface{}) *WhereCondition {
	return &WhereCondition{
		Field:    field,
		Operator: SQLOpLte,
		Value:    value,
		Logic:    SQLLogicAnd,
	}
}

// ConditionLike 创建模糊匹配条件
// func ConditionLike(field string, value string) *WhereCondition {
// 	return &WhereCondition{
// 		Field:    field,
// 		Operator: SQLOpLike,
// 		Value:    value,
// 		Logic:    SQLLogicAnd,
// 	}
// }

// ConditionIn 创建 IN 条件
func ConditionIn(field string, values ...interface{}) *WhereCondition {
	return &WhereCondition{
		Field:    field,
		Operator: SQLOpIn,
		Value:    values,
		Logic:    SQLLogicAnd,
	}
}

// ConditionNotIn 创建 NOT IN 条件
func ConditionNotIn(field string, values ...interface{}) *WhereCondition {
	return &WhereCondition{
		Field:    field,
		Operator: SQLOpNotIn,
		Value:    values,
		Logic:    SQLLogicAnd,
	}
}

// ConditionBetween 创建 BETWEEN 条件
func ConditionBetween(field string, min, max interface{}) *WhereCondition {
	return &WhereCondition{
		Field:    field,
		Operator: SQLOpBetween,
		Value:    []interface{}{min, max},
		Logic:    SQLLogicAnd,
	}
}

// Or 设置逻辑为 OR
func (c *WhereCondition) Or() *WhereCondition {
	c.Logic = SQLLogicOr
	return c
}

// And 设置逻辑为 AND
func (c *WhereCondition) And() *WhereCondition {
	c.Logic = SQLLogicAnd
	return c
}

// AddChild 添加子条件
func (c *WhereCondition) AddChild(child *WhereCondition) *WhereCondition {
	c.Children = append(c.Children, child)
	return c
}

// AddChildren 批量添加子条件
func (c *WhereCondition) AddChildren(children ...*WhereCondition) *WhereCondition {
	c.Children = append(c.Children, children...)
	return c
}

// ConditionGroup 创建一个条件组（用于嵌套）
func ConditionGroup(conditions ...*WhereCondition) *WhereCondition {
	if len(conditions) == 0 {
		return nil
	}
	return &WhereCondition{
		Children: conditions,
	}
}

// ConditionAndGroup 创建 AND 连接的条件组
func ConditionAndGroup(conditions ...*WhereCondition) *WhereCondition {
	if len(conditions) == 0 {
		return nil
	}
	for _, c := range conditions {
		c.Logic = SQLLogicAnd
	}
	return ConditionGroup(conditions...)
}

// ConditionOrGroup 创建 OR 连接的条件组
func ConditionOrGroup(conditions ...*WhereCondition) *WhereCondition {
	if len(conditions) == 0 {
		return nil
	}
	for _, c := range conditions {
		c.Logic = SQLLogicOr
	}
	return ConditionGroup(conditions...)
}

// ==================== 索引验证功能 ====================

// BuildWithIndexCheck 构建最终的 WHERE SQL 和参数，并检查索引使用情况
// primaryKeyColumns: 主键列名切片
// indexColumns: 索引列名切片的切片（每个子切片代表一个索引的列，按顺序）
// 返回: (whereSQL, args, usedIndexName, error)
// usedIndexName: 使用的索引名称（主键返回 "PRIMARY"，普通索引返回索引在 indexColumns 中的索引号，如 "INDEX_0"）
func (b *WhereClauseBuilder) BuildWithIndexCheck(
	primaryKeyColumns []string,
	indexColumns [][]string,
) (string, []interface{}, string, error) {
	if b.root == nil {
		return "", nil, "", nil
	}

	sql, args, err := buildWhereCondition(b.root)
	if err != nil {
		return "", nil, "", err
	}

	if sql == "" {
		return "", nil, "", nil
	}

	// 收集条件中使用的所有字段
	usedFields := collectConditionFields(b.root)

	// 检查是否使用了主键
	if isPrimaryKeyUsed(usedFields, primaryKeyColumns) {
		return "WHERE " + sql, args, "PRIMARY", nil
	}

	// 检查是否使用了索引
	usedIndex := findMatchingIndex(usedFields, indexColumns)
	if usedIndex != "" {
		return "WHERE " + sql, args, usedIndex, nil
	}

	// 未使用任何索引，返回错误
	return "", nil, "", fmt.Errorf("query does not use any index or primary key. used fields: %v", usedFields)
}

// collectConditionFields 收集条件中使用的所有字段
func collectConditionFields(cond *WhereCondition) []string {
	if cond == nil {
		return nil
	}

	fields := make(map[string]struct{})

	// 收集当前条件的字段
	if cond.Field != "" {
		fields[cond.Field] = struct{}{}
	}

	// 递归收集子条件的字段
	for _, child := range cond.Children {
		childFields := collectConditionFields(child)
		for _, field := range childFields {
			fields[field] = struct{}{}
		}
	}

	// 转换为切片
	result := make([]string, 0, len(fields))
	for field := range fields {
		result = append(result, field)
	}
	return result
}

// isPrimaryKeyUsed 检查是否使用了主键
func isPrimaryKeyUsed(usedFields []string, primaryKeyColumns []string) bool {
	if len(primaryKeyColumns) == 0 {
		return false
	}

	// 检查主键的引导列是否被使用
	leadingColumn := primaryKeyColumns[0]
	for _, field := range usedFields {
		if field == leadingColumn {
			return true
		}
	}
	return false
}

// findMatchingIndex 查找匹配的索引
// 返回: 索引名称（如 "INDEX_0", "INDEX_1"），如果没有匹配的索引则返回空字符串
func findMatchingIndex(usedFields []string, indexColumns [][]string) string {
	for i, index := range indexColumns {
		if len(index) == 0 {
			continue
		}

		// 检查索引的引导列是否被使用
		leadingColumn := index[0]
		for _, field := range usedFields {
			if field == leadingColumn {
				return fmt.Sprintf("INDEX_%d", i)
			}
		}
	}
	return ""
}
