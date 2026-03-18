package conditonwhere

import "fmt"

// Condition 接口定义了查询条件的通用行为
type Condition interface {
	// 为了实现Build方法使用
	Where
}

type IndexField string
// EqCondition 表示等于条件
type EqCondition struct {
	Field IndexField
	Value interface{}
}

// Build 构建等于条件的 SQL 片段
func (c *EqCondition) Build() (string, []interface{}) {
	return fmt.Sprintf("%s = ?", c.Field), []interface{}{c.Value}
}

// NeqCondition 表示不等于条件
type NeqCondition struct {
	Field IndexField
	Value interface{}
}

// Build 构建不等于条件的 SQL 片段
func (c *NeqCondition) Build() (string, []interface{}) {
	return fmt.Sprintf("%s != ?", c.Field), []interface{}{c.Value}
}

// GtCondition 表示大于条件
type GtCondition struct {
	Field IndexField
	Value interface{}
}

// Build 构建大于条件的 SQL 片段
func (c *GtCondition) Build() (string, []interface{}) {
	return fmt.Sprintf("%s > ?", c.Field), []interface{}{c.Value}
}

// LtCondition 表示小于条件
type LtCondition struct {
	Field IndexField
	Value interface{}
}

// Build 构建小于条件的 SQL 片段
func (c *LtCondition) Build() (string, []interface{}) {
	return fmt.Sprintf("%s < ?", c.Field), []interface{}{c.Value}
}

// LikeCondition 表示模糊匹配条件
type LikeCondition struct {
	Field IndexField
	Value string
}

// Build 构建模糊匹配条件的 SQL 片段
func (c *LikeCondition) Build() (string, []interface{}) {
	return fmt.Sprintf("%s LIKE ?", c.Field), []interface{}{"%" + c.Value + "%"}
}

// InCondition 表示 IN 条件
type InCondition struct {
	Field  IndexField
	Values []interface{}
}

// Build 构建 IN 条件的 SQL 片段
func (c *InCondition) Build() (string, []interface{}) {
	placeholders := ""
	for i := 0; i < len(c.Values); i++ {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
	}
	return fmt.Sprintf("%s IN (%s)", c.Field, placeholders), c.Values
}

// 便捷函数用于创建各种条件
// func Eq(field IndexField, value interface{}) Condition {
// 	return &EqCondition{Field: field, Value: value}
// }

// func Neq(field IndexField, value interface{}) Condition {
// 	return &NeqCondition{Field: field, Value: value}
// }

// func Gt(field IndexField, value interface{}) Condition {
// 	return &GtCondition{Field: field, Value: value}
// }

// func Lt(field IndexField, value interface{}) Condition {
// 	return &LtCondition{Field: field, Value: value}
// }

// func Like(field IndexField, value string) Condition {
// 	return &LikeCondition{Field: field, Value: value}
// }

// func In(field IndexField, values ...interface{}) Condition {
// 	return &InCondition{Field: field, Values: values}
// }