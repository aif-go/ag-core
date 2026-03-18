package conditonwhere

// Operator 定义支持的SQL操作符
// type Operator string

// const (
// 	OpEq        Operator = "="
// 	OpNeq       Operator = "!="
// 	OpGt        Operator = ">"
// 	OpLt        Operator = "<"
// 	OpGte       Operator = ">="
// 	OpLte       Operator = "<="
// 	OpLike      Operator = "like"
// 	OpIn        Operator = "in"
// 	OpNotIn     Operator = "not in"
// 	OpBetween   Operator = "between"
// )

// LogicOperator 定义逻辑连接符
// type LogicOperator string

// const (
// 	LogicAnd LogicOperator = "AND"
// 	LogicOr  LogicOperator = "OR"
// )

// WhereClause 定义用户输入的 WHERE 条件结构
// 支持嵌套条件和多种操作符
// type WhereClause struct {
// 	Field    string        // 列名
// 	Operator Operator      // 操作符: =, !=, >, <, >=, <=, like, in, not in, between
// 	Value    interface{}   // 值（对于 between 和 in 可能是数组）
// 	Logic    LogicOperator // 逻辑连接符: AND, OR
// 	Children []*WhereClause // 子条件（用于嵌套，如 (a=1 OR b=2)）
// }

// WhereBuilder 高性能的 WHERE 条件构建器
// type WhereBuilder struct {
// 	root *WhereClause
// }

// // NewWhereBuilder 创建新的 WhereBuilder
// func NewWhereBuilder() *WhereBuilder {
// 	return &WhereBuilder{}
// }

// AddCondition 添加一个条件
// func (b *WhereBuilder) AddCondition(clause *WhereClause) *WhereBuilder {
// 	if b.root == nil {
// 		b.root = clause
// 	} else {
// 		// 将新条件添加到根的子条件中
// 		b.root.Children = append(b.root.Children, clause)
// 	}
// 	return b
// }

// // AddConditions 批量添加条件
// func (b *WhereBuilder) AddConditions(clauses ...*WhereClause) *WhereBuilder {
// 	for _, clause := range clauses {
// 		b.AddCondition(clause)
// 	}
// 	return b
// }

// SetRoot 设置根条件
// func (b *WhereBuilder) SetRoot(clause *WhereClause) *WhereBuilder {
// 	b.root = clause
// 	return b
// }

// // Build 构建最终的 WHERE SQL 和参数
// // 返回: (whereSQL, args, error)
// func (b *WhereBuilder) Build() (string, []interface{}, error) {
// 	if b.root == nil {
// 		return "", nil, nil
// 	}

// 	sql, args, err := buildClause(b.root)
// 	if err != nil {
// 		return "", nil, err
// 	}

// 	if sql == "" {
// 		return "", nil, nil
// 	}

// 	return "WHERE " + sql, args, nil
// }

// buildClause 递归构建条件
// func buildClause(clause *WhereClause) (string, []interface{}, error) {
// 	if clause == nil {
// 		return "", nil, nil
// 	}

// 	var sqlParts []string
// 	var args []interface{}

// 	// 构建当前条件
// 	if clause.Field != "" {
// 		conditionSQL, conditionArgs, err := buildCondition(clause)
// 		if err != nil {
// 			return "", nil, err
// 		}
// 		if conditionSQL != "" {
// 			sqlParts = append(sqlParts, conditionSQL)
// 			args = append(args, conditionArgs...)
// 		}
// 	}

// 	// 构建子条件（嵌套条件）
// 	if len(clause.Children) > 0 {
// 		for i, child := range clause.Children {
// 			childSQL, childArgs, err := buildClause(child)
// 			if err != nil {
// 				return "", nil, err
// 			}

// 			if childSQL != "" {
// 				// 添加逻辑连接符
// 				if len(sqlParts) > 0 {
// 					logic := string(child.Logic)
// 					if logic == "" {
// 						logic = string(LogicAnd) // 默认使用 AND
// 					}
// 					sqlParts = append(sqlParts, logic)
// 				}
// 				sqlParts = append(sqlParts, childSQL)
// 				args = append(args, childArgs...)
// 			}

// 			// 如果不是最后一个子条件，添加逻辑连接符
// 			if i < len(clause.Children)-1 {
// 				nextChild := clause.Children[i+1]
// 				if nextChild.Logic != "" {
// 					sqlParts = append(sqlParts, string(nextChild.Logic))
// 				} else {
// 					sqlParts = append(sqlParts, string(LogicAnd))
// 				}
// 			}
// 		}
// 	}

// 	if len(sqlParts) == 0 {
// 		return "", nil, nil
// 	}

// 	// 如果有多个部分，需要用括号包裹
// 	sql := strings.Join(sqlParts, " ")
// 	if strings.Count(sql, " ") > 1 {
// 		sql = "(" + sql + ")"
// 	}

// 	return sql, args, nil
// }

// // buildCondition 构建单个条件的 SQL
// func buildCondition(clause *WhereClause) (string, []interface{}, error) {
// 	if clause.Field == "" {
// 		return "", nil, nil
// 	}

// 	switch clause.Operator {
// 	case OpEq:
// 		return fmt.Sprintf("%s = ?", clause.Field), []interface{}{clause.Value}, nil
// 	case OpNeq:
// 		return fmt.Sprintf("%s != ?", clause.Field), []interface{}{clause.Value}, nil
// 	case OpGt:
// 		return fmt.Sprintf("%s > ?", clause.Field), []interface{}{clause.Value}, nil
// 	case OpLt:
// 		return fmt.Sprintf("%s < ?", clause.Field), []interface{}{clause.Value}, nil
// 	case OpGte:
// 		return fmt.Sprintf("%s >= ?", clause.Field), []interface{}{clause.Value}, nil
// 	case OpLte:
// 		return fmt.Sprintf("%s <= ?", clause.Field), []interface{}{clause.Value}, nil
// 	case OpLike:
// 		return fmt.Sprintf("%s LIKE ?", clause.Field), []interface{}{"%" + fmt.Sprintf("%v", clause.Value) + "%"}, nil
// 	case OpIn:
// 		values, ok := clause.Value.([]interface{})
// 		if !ok {
// 			return "", nil, fmt.Errorf("IN operator requires []interface{} value")
// 		}
// 		if len(values) == 0 {
// 			return "", nil, nil
// 		}
// 		placeholders := make([]string, len(values))
// 		for i := range placeholders {
// 			placeholders[i] = "?"
// 		}
// 		return fmt.Sprintf("%s IN (%s)", clause.Field, strings.Join(placeholders, ", ")), values, nil
// 	case OpNotIn:
// 		values, ok := clause.Value.([]interface{})
// 		if !ok {
// 			return "", nil, fmt.Errorf("NOT IN operator requires []interface{} value")
// 		}
// 		if len(values) == 0 {
// 			return "", nil, nil
// 		}
// 		placeholders := make([]string, len(values))
// 		for i := range placeholders {
// 			placeholders[i] = "?"
// 		}
// 		return fmt.Sprintf("%s NOT IN (%s)", clause.Field, strings.Join(placeholders, ", ")), values, nil
// 	case OpBetween:
// 		values, ok := clause.Value.([]interface{})
// 		if !ok || len(values) != 2 {
// 			return "", nil, fmt.Errorf("BETWEEN operator requires []interface{} with 2 values")
// 		}
// 		return fmt.Sprintf("%s BETWEEN ? AND ?", clause.Field), values, nil
// 	default:
// 		return "", nil, fmt.Errorf("unsupported operator: %s", clause.Operator)
// 	}
// }

// // ==================== 便捷方法 ====================

// // Eq 创建等于条件
// func Eq(field string, value interface{}) *WhereClause {
// 	return &WhereClause{
// 		Field:    field,
// 		Operator: OpEq,
// 		Value:    value,
// 		Logic:    LogicAnd,
// 	}
// }

// // Neq 创建不等于条件
// func Neq(field string, value interface{}) *WhereClause {
// 	return &WhereClause{
// 		Field:    field,
// 		Operator: OpNeq,
// 		Value:    value,
// 		Logic:    LogicAnd,
// 	}
// }

// // Gt 创建大于条件
// func Gt(field string, value interface{}) *WhereClause {
// 	return &WhereClause{
// 		Field:    field,
// 		Operator: OpGt,
// 		Value:    value,
// 		Logic:    LogicAnd,
// 	}
// }

// // Lt 创建小于条件
// func Lt(field string, value interface{}) *WhereClause {
// 	return &WhereClause{
// 		Field:    field,
// 		Operator: OpLt,
// 		Value:    value,
// 		Logic:    LogicAnd,
// 	}
// }

// // Gte 创建大于等于条件
// func Gte(field string, value interface{}) *WhereClause {
// 	return &WhereClause{
// 		Field:    field,
// 		Operator: OpGte,
// 		Value:    value,
// 		Logic:    LogicAnd,
// 	}
// }

// // Lte 创建小于等于条件
// func Lte(field string, value interface{}) *WhereClause {
// 	return &WhereClause{
// 		Field:    field,
// 		Operator: OpLte,
// 		Value:    value,
// 		Logic:    LogicAnd,
// 	}
// }

// // Like 创建模糊匹配条件
// func Like(field string, value string) *WhereClause {
// 	return &WhereClause{
// 		Field:    field,
// 		Operator: OpLike,
// 		Value:    value,
// 		Logic:    LogicAnd,
// 	}
// }

// // In 创建 IN 条件
// func In(field string, values ...interface{}) *WhereClause {
// 	return &WhereClause{
// 		Field:    field,
// 		Operator: OpIn,
// 		Value:    values,
// 		Logic:    LogicAnd,
// 	}
// }

// // NotIn 创建 NOT IN 条件
// func NotIn(field string, values ...interface{}) *WhereClause {
// 	return &WhereClause{
// 		Field:    field,
// 		Operator: OpNotIn,
// 		Value:    values,
// 		Logic:    LogicAnd,
// 	}
// }

// // Between 创建 BETWEEN 条件
// func Between(field string, min, max interface{}) *WhereClause {
// 	return &WhereClause{
// 		Field:    field,
// 		Operator: OpBetween,
// 		Value:    []interface{}{min, max},
// 		Logic:    LogicAnd,
// 	}
// }

// // Or 设置逻辑为 OR
// func (c *WhereClause) Or() *WhereClause {
// 	c.Logic = LogicOr
// 	return c
// }

// // And 设置逻辑为 AND
// func (c *WhereClause) And() *WhereClause {
// 	c.Logic = LogicAnd
// 	return c
// }

// // AddChild 添加子条件
// func (c *WhereClause) AddChild(child *WhereClause) *WhereClause {
// 	c.Children = append(c.Children, child)
// 	return c
// }

// // AddChildren 批量添加子条件
// func (c *WhereClause) AddChildren(children ...*WhereClause) *WhereClause {
// 	c.Children = append(c.Children, children...)
// 	return c
// }

// // Group 创建一个条件组（用于嵌套）
// func Group(conditions ...*WhereClause) *WhereClause {
// 	if len(conditions) == 0 {
// 		return nil
// 	}
// 	return &WhereClause{
// 		Children: conditions,
// 	}
// }

// // AndGroup 创建 AND 连接的条件组
// func AndGroup(conditions ...*WhereClause) *WhereClause {
// 	if len(conditions) == 0 {
// 		return nil
// 	}
// 	for _, c := range conditions {
// 		c.Logic = LogicAnd
// 	}
// 	return Group(conditions...)
// }

// // OrGroup 创建 OR 连接的条件组
// func OrGroup(conditions ...*WhereClause) *WhereClause {
// 	if len(conditions) == 0 {
// 		return nil
// 	}
// 	for _, c := range conditions {
// 		c.Logic = LogicOr
// 	}
// 	return Group(conditions...)
// }
