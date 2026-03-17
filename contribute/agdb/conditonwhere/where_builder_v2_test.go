package conditonwhere

import (
	"fmt"
	"testing"
)

// TestWhereBuilderV2 测试 WHERE 条件构建器的各种用法
func TestWhereBuilderV2(t *testing.T) {
	
	t.Run("Example1_SimpleAnd", func(t *testing.T) {
		// WHERE A = 1 AND B = 2
		builder := NewWhereClauseBuilder()
		builder.AddCondition(ConditionEq("A", 1))
		builder.AddCondition(ConditionEq("B", 2))
		
		sql, args, err := builder.Build()
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE A = ? AND B = ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		if len(args) != 2 || args[0] != 1 || args[1] != 2 {
			t.Errorf("Expected args [1, 2], got: %v", args)
		}
		
		fmt.Printf("Test1 - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example2_SimpleOr", func(t *testing.T) {
		// WHERE A = 1 OR B = 2
		cond1 := ConditionEq("A", 1).Or()
		cond2 := ConditionEq("B", 2).Or()
		
		builder := NewWhereClauseBuilder()
		builder.AddCondition(cond1)
		builder.AddCondition(cond2)
		
		sql, args, err := builder.Build()
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE A = ? OR B = ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test2 - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example3_NestedConditions", func(t *testing.T) {
		// WHERE A = 1 AND (B = 2 OR C = 3)
		orGroup := ConditionOrGroup(
			ConditionEq("B", 2),
			ConditionEq("C", 3),
		)
		
		condA := ConditionEq("A", 1)
		condA.AddChild(orGroup)
		
		builder := NewWhereClauseBuilder()
		builder.AddCondition(condA)
		
		sql, args, err := builder.Build()
		if err != nil {
			t.Fatal(err)
		}
		
		fmt.Printf("Test3 - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example4_AllOperators", func(t *testing.T) {
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
		
		sql, args, err := builder.Build()
		if err != nil {
			t.Fatal(err)
		}
		
		fmt.Printf("Test4 - SQL: %s\n", sql)
		fmt.Printf("Test4 - Args: %v\n", args)
	})
	
	t.Run("Example5_ComplexNested", func(t *testing.T) {
		// WHERE (A = 1 OR B = 2) AND (C = 3 OR (D = 4 AND E = 5))
		group1 := ConditionOrGroup(
			ConditionEq("A", 1),
			ConditionEq("B", 2),
		)
		
		group2 := ConditionOrGroup(
			ConditionEq("C", 3),
			ConditionAndGroup(
				ConditionEq("D", 4),
				ConditionEq("E", 5),
			),
		)
		
		root := ConditionAndGroup(group1, group2)
		
		builder := NewWhereClauseBuilder()
		builder.SetRoot(root)
		
		sql, args, err := builder.Build()
		if err != nil {
			t.Fatal(err)
		}
		
		fmt.Printf("Test5 - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example6_ChainCall", func(t *testing.T) {
		builder := NewWhereClauseBuilder()
		
		sql, args, err := builder.
			AddCondition(ConditionEq("user_id", 100)).
			// AddCondition(ConditionLike("username", "admin")).
			AddCondition(ConditionIn("role", "admin", "superadmin")).
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		fmt.Printf("Test6 - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example7_DynamicBuild", func(t *testing.T) {
		params := map[string]interface{}{
			"name":      "john",
			"age":       25,
			"status":    []interface{}{"active", "pending"},
			"min_price": 10,
			"max_price": 100,
		}
		
		builder := NewWhereClauseBuilder()
		
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
		
		sql, args, err := builder.Build()
		if err != nil {
			t.Fatal(err)
		}
		
		fmt.Printf("Test7 - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example8_ErrorHandling", func(t *testing.T) {
		builder := NewWhereClauseBuilder()
		
		// BETWEEN 操作符需要两个值，这里故意传错
		cond := &WhereCondition{
			Field:    "age",
			Operator: SQLOpBetween,
			Value:    []interface{}{18}, // 只有一个值，会报错
		}
		builder.AddCondition(cond)
		
		_, _, err := builder.Build()
		if err == nil {
			t.Error("Expected error for invalid BETWEEN condition")
		}
		
		fmt.Printf("Test8 - Expected error: %v\n", err)
	})
	
	t.Run("Example9_EmptyBuilder", func(t *testing.T) {
		builder := NewWhereClauseBuilder()
		
		sql, args, err := builder.Build()
		if err != nil {
			t.Fatal(err)
		}
		
		if sql != "" {
			t.Errorf("Expected empty SQL, got: %s", sql)
		}
		
		if args != nil {
			t.Errorf("Expected nil args, got: %v", args)
		}
		
		fmt.Printf("Test9 - Empty builder returns: SQL='%s', Args=%v\n", sql, args)
	})
	
	t.Run("Example10_InWithEmptyValues", func(t *testing.T) {
		builder := NewWhereClauseBuilder()
		builder.AddCondition(ConditionEq("id", 1))
		builder.AddCondition(ConditionIn("status")) // 空的 IN 条件
		
		sql, args, err := builder.Build()
		if err != nil {
			t.Fatal(err)
		}
		
		// 空的 IN 条件应该被忽略
		expectedSQL := "WHERE id = ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test10 - SQL: %s, Args: %v\n", sql, args)
	})
}

// BenchmarkWhereBuilder 性能测试
func BenchmarkWhereBuilder(b *testing.B) {
	builder := NewWhereClauseBuilder()
	builder.AddConditions(
		ConditionEq("id", 1),
		ConditionEq("status", "active"),
		ConditionGt("created_at", "2024-01-01"),
		// ConditionLike("name", "test"),
		ConditionIn("type", 1, 2, 3),
	)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = builder.Build()
	}
}

// BenchmarkWhereBuilderWithNesting 嵌套条件的性能测试
func BenchmarkWhereBuilderWithNesting(b *testing.B) {
	group1 := ConditionOrGroup(
		ConditionEq("A", 1),
		ConditionEq("B", 2),
		ConditionEq("C", 3),
	)
	
	group2 := ConditionOrGroup(
		ConditionEq("D", 4),
		ConditionEq("E", 5),
		ConditionEq("F", 6),
	)
	
	root := ConditionAndGroup(group1, group2)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewWhereClauseBuilder()
		builder.SetRoot(root)
		_, _, _ = builder.Build()
	}
}
