package conditonwhere

import (
	"fmt"
)

// ============ WHERE 条件构建器使用示例 ============
//
// 使用说明：
// 1. 需要先手动删除以下冲突文件：
//    - where_clause.go
//    - where_clause_builder.go
// 2. 然后使用 where_builder_v2.go 中的 WhereCondition 和 WhereClauseBuilder
//
// ================================================

// Example1 简单的 AND 条件: WHERE A = 1 AND B = 2
func Example1() {
	builder := NewWhereClauseBuilder()
	builder.AddCondition(ConditionEq("A", 1))
	builder.AddCondition(ConditionEq("B", 2))
	
	sql, args, _ := builder.Build()
	fmt.Println(sql)  // WHERE A = ? AND B = ?
	fmt.Println(args) // [1, 2]
}

// Example2 简单的 OR 条件: WHERE A = 1 OR B = 2
func Example2() {
	cond1 := ConditionEq("A", 1).Or()  // 设置为 OR
	cond2 := ConditionEq("B", 2).Or()  // 设置为 OR
	
	builder := NewWhereClauseBuilder()
	builder.AddCondition(cond1)
	builder.AddCondition(cond2)
	
	sql, args, _ := builder.Build()
	fmt.Println(sql)  // WHERE A = ? OR B = ?
	fmt.Println(args) // [1, 2]
}

// Example3 嵌套条件: WHERE A = 1 AND (B = 2 OR C = 3)
func Example3() {
	// 创建嵌套的 OR 条件组
	orGroup := ConditionOrGroup(
		ConditionEq("B", 2),
		ConditionEq("C", 3),
	)
	
	// 将 OR 组作为 A 条件的子条件
	condA := ConditionEq("A", 1)
	condA.AddChild(orGroup)
	
	builder := NewWhereClauseBuilder()
	builder.AddCondition(condA)
	
	sql, args, _ := builder.Build()
	fmt.Println(sql)  // WHERE A = ? AND (B = ? OR C = ?)
	fmt.Println(args) // [1, 2, 3]
}

// Example4 使用所有操作符
func Example4() {
	builder := NewWhereClauseBuilder()
	builder.AddConditions(
		ConditionEq("id", 1),
		ConditionNeq("status", "deleted"),
		ConditionGt("age", 18),
		ConditionLt("age", 60),
		ConditionGte("score", 80),
		ConditionLte("price", 100),
		// ConditionLike("name", "john"),
		ConditionIn("type", 1, 2, 3),
		ConditionNotIn("exclude", 4, 5),
		ConditionBetween("created_at", "2024-01-01", "2024-12-31"),
	)
	
	sql, args, _ := builder.Build()
	fmt.Println(sql)
	fmt.Println(args)
}

// Example5 复杂嵌套: WHERE (A = 1 OR B = 2) AND (C = 3 OR (D = 4 AND E = 5))
func Example5() {
	// 第一个 OR 组
	group1 := ConditionOrGroup(
		ConditionEq("A", 1),
		ConditionEq("B", 2),
	)
	
	// 第二个 OR 组，包含嵌套的 AND
	group2 := ConditionOrGroup(
		ConditionEq("C", 3),
		ConditionAndGroup(
			ConditionEq("D", 4),
			ConditionEq("E", 5),
		),
	)
	
	// 将两个组通过 AND 连接
	root := ConditionAndGroup(group1, group2)
	
	builder := NewWhereClauseBuilder()
	builder.SetRoot(root)
	
	sql, args, _ := builder.Build()
	fmt.Println(sql)  // WHERE (A = ? OR B = ?) AND (C = ? OR (D = ? AND E = ?))
	fmt.Println(args) // [1, 2, 3, 4, 5]
}

// Example6 链式调用风格
func Example6() {
	builder := NewWhereClauseBuilder()
	
	// 链式添加条件
	builder.
		AddCondition(ConditionEq("user_id", 100)).
		// AddCondition(ConditionLike("username", "admin")).
		AddCondition(ConditionIn("role", "admin", "superadmin"))
	
	sql, args, _ := builder.Build()
	fmt.Println(sql)
	fmt.Println(args)
}

// Example7 动态构建条件（适用于根据用户输入动态构建）
func Example7(params map[string]interface{}) {
	builder := NewWhereClauseBuilder()
	
	// 根据参数动态添加条件
	// if name, ok := params["name"]; ok {
	// 	builder.AddCondition(ConditionLike("name", name.(string)))
	// }
	
	if age, ok := params["age"]; ok {
		builder.AddCondition(ConditionEq("age", age))
	}
	
	if status, ok := params["status"]; ok {
		builder.AddCondition(ConditionIn("status", status.([]interface{})...))
	}
	
	if minPrice, ok := params["min_price"]; ok {
		if maxPrice, ok := params["max_price"]; ok {
			builder.AddCondition(ConditionBetween("price", minPrice, maxPrice))
		}
	}
	
	sql, args, _ := builder.Build()
	fmt.Println(sql)
	fmt.Println(args)
}

// Example8 性能优化 - 预分配容量
func Example8() {
	// 对于已知数量的条件，可以一次性添加
	conditions := []*WhereCondition{
		ConditionEq("id", 1),
		ConditionEq("status", "active"),
		ConditionGt("created_at", "2024-01-01"),
	}
	
	builder := NewWhereClauseBuilder()
	builder.AddConditions(conditions...)
	
	sql, args, _ := builder.Build()
	fmt.Println(sql)
	fmt.Println(args)
}

// Example9 错误处理
func Example9() {
	builder := NewWhereClauseBuilder()
	
	// BETWEEN 操作符需要两个值
	cond := ConditionBetween("age", 18,19) // 错误：只有一个值
	builder.AddCondition(cond)
	
	sql, args, err := builder.Build()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(sql)
		fmt.Println(args)
	}
}
