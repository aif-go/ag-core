package conditonwhere

// import "strings"

// // Where 接口定义了查询条件的构建行为
// type Where interface {
// 	Build() (string, []interface{})
// }


// // AndWhere 表示 AND 逻辑组合
// type AndWhere struct {
// 	conditions []Where
// }

// // Build 构建 AND 条件的 SQL 片段
// func (w *AndWhere) Build() (string, []interface{}) {
// 	if len(w.conditions) == 0 {
// 		return "", nil
// 	}
	
// 	var sqlParts []string
// 	var args []interface{}
	
// 	for _, cond := range w.conditions {
// 		sql, condArgs := cond.Build()
// 		if sql != "" {
// 			sqlParts = append(sqlParts, sql)
// 			args = append(args, condArgs...)
// 		}
// 	}
	
// 	if len(sqlParts) == 0 {
// 		return "", nil
// 	}
	
// 	if len(sqlParts) == 1 {
// 		return sqlParts[0], args
// 	}
	
// 	return "(" + strings.Join(sqlParts, " AND ") + ")", args
// }


// // OrWhere 表示 OR 逻辑组合
// type OrWhere struct {
// 	conditions []Where
// }

// // Build 构建 OR 条件的 SQL 片段
// func (w *OrWhere) Build() (string, []interface{}) {
// 	if len(w.conditions) == 0 {
// 		return "", nil
// 	}
	
// 	var sqlParts []string
// 	var args []interface{}
	
// 	for _, cond := range w.conditions {
// 		sql, condArgs := cond.Build()
// 		if sql != "" {
// 			sqlParts = append(sqlParts, sql)
// 			args = append(args, condArgs...)
// 		}
// 	}
	
// 	if len(sqlParts) == 0 {
// 		return "", nil
// 	}
	
// 	if len(sqlParts) == 1 {
// 		return sqlParts[0], args
// 	}
	
// 	return "(" + strings.Join(sqlParts, " OR ") + ")", args
// }



// // And 组合多个条件为 AND 逻辑
// func And(conditions ...Where) Where {
// 	return &AndWhere{
// 		conditions: conditions,
// 	}
// }

// // Or 组合多个条件为 OR 逻辑
// func Or(conditions ...Where) Where {
// 	return &OrWhere{
// 		conditions: conditions,
// 	}
// }

// // WhereBuilder 用于构建复杂的查询条件
// type WhereBuilder struct {
// 	conditions []Where
// }


// // NewWhereBuilder 创建新的 WhereBuilder
// func NewWhereBuilder() *WhereBuilder {
// 	return &WhereBuilder{}
// }



// // ComplexWhereBuilder 用于构建更复杂的嵌套条件
// type ComplexWhereBuilder struct {
// 	groups []Where
// }

// // NewComplexWhereBuilder 创建新的复杂条件构建器
// func NewComplexWhereBuilder() *ComplexWhereBuilder {
// 	return &ComplexWhereBuilder{}
// }

// // AddGroup 添加一个条件组
// func (b *ComplexWhereBuilder) AddGroup(group Where) *ComplexWhereBuilder {
// 	b.groups = append(b.groups, group)
// 	return b
// }

// // Build 构建最终的复杂条件
// func (b *ComplexWhereBuilder) Build() Where {
// 	if len(b.groups) == 0 {
// 		return nil
// 	}
	
// 	if len(b.groups) == 1 {
// 		return b.groups[0]
// 	}
	
// 	return Or(b.groups...)
// }


// // And 添加 AND 条件
// func (b *WhereBuilder) And(cond Where) *WhereBuilder {
// 	b.conditions = append(b.conditions, cond)
// 	return b
// }

// // Or 添加 OR 条件
// func (b *WhereBuilder) Or(cond Where) *WhereBuilder {
// 	b.conditions = append(b.conditions, cond)
// 	return b
// }



// // Build 构建最终的 WHERE 条件
// func (b *WhereBuilder) Build() Where {
// 	if len(b.conditions) == 0 {
// 		return nil
// 	}
	
// 	if len(b.conditions) == 1 {
// 		return b.conditions[0]
// 	}
	
// 	return And(b.conditions...)
// }