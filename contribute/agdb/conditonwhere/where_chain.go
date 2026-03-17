package conditonwhere

import "strings"

// ChainBuilder 链式条件构建器
type ChainBuilder struct {
	currentGroup Where
}

// NewChainBuilder 创建新的链式构建器
func NewChainBuilder() *ChainBuilder {
	return &ChainBuilder{}
}

// AND 开始一个AND条件组
func AND(conditions ...Where) *ChainBuilder {
	builder := &ChainBuilder{}
	if len(conditions) > 0 {
		builder.currentGroup = And(conditions...)
	}
	return builder
}

// OR 开始一个OR条件组
func OR(conditions ...Where) *ChainBuilder {
	builder := &ChainBuilder{}
	if len(conditions) > 0 {
		builder.currentGroup = Or(conditions...)
	}
	return builder
}

// AND 添加AND条件到当前组
func (b *ChainBuilder) AND(conditions ...Where) *ChainBuilder {
	if b.currentGroup == nil {
		return AND(conditions...)
	}
	
	// 如果当前组已经是AND，直接添加条件
	if andGroup, ok := b.currentGroup.(*AndWhere); ok {
		andGroup.conditions = append(andGroup.conditions, conditions...)
	} else {
		// 否则创建新的AND组
		allConditions := []Where{b.currentGroup}
		allConditions = append(allConditions, conditions...)
		b.currentGroup = And(allConditions...)
	}
	return b
}

// OR 添加OR条件到当前组
func (b *ChainBuilder) OR(conditions ...Where) *ChainBuilder {
	if b.currentGroup == nil {
		return OR(conditions...)
	}
	
	// 如果当前组已经是OR，直接添加条件
	if orGroup, ok := b.currentGroup.(*OrWhere); ok {
		orGroup.conditions = append(orGroup.conditions, conditions...)
	} else {
		// 否则创建新的OR组
		allConditions := []Where{b.currentGroup}
		allConditions = append(allConditions, conditions...)
		b.currentGroup = Or(allConditions...)
	}
	return b
}

// Build 构建最终的WHERE条件
func (b *ChainBuilder) Build() (string, []interface{}) {
	if b.currentGroup == nil {
		return "", nil
	}
	wheresql,args:=b.currentGroup.Build()
	wheresql=strings.TrimSuffix(strings.TrimPrefix(wheresql,"("),")")
	return wheresql,args
}

// 便捷方法用于链式调用
// func (b *ChainBuilder) EQ(field IndexField, value interface{}) *ChainBuilder {
// 	return b.AND(Eq(field, value))
// }

// func (b *ChainBuilder) NEQ(field IndexField, value interface{}) *ChainBuilder {
// 	return b.AND(Neq(field, value))
// }

// func (b *ChainBuilder) GT(field IndexField, value interface{}) *ChainBuilder {
// 	return b.AND(Gt(field, value))
// }

// func (b *ChainBuilder) LT(field IndexField, value interface{}) *ChainBuilder {
// 	return b.AND(Lt(field, value))
// }

// func (b *ChainBuilder) LIKE(field IndexField, value string) *ChainBuilder {
// 	return b.AND(Like(field, value))
// }

// func (b *ChainBuilder) IN(field IndexField, values ...interface{}) *ChainBuilder {
// 	return b.AND(In(field, values...))
// }