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
	
	t.Run("Example11_ChainEq", func(t *testing.T) {
		// 测试链式调用 Eq
		sql, args, err := NewWhereClauseBuilder().
			Eq("name", "John").
			Eq("age", 25).
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE name = ? AND age = ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test11 - ChainEq - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example12_ChainNeq", func(t *testing.T) {
		// 测试链式调用 Neq
		sql, args, err := NewWhereClauseBuilder().
			Neq("status", "deleted").
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE status != ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test12 - ChainNeq - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example13_ChainGt", func(t *testing.T) {
		// 测试链式调用 Gt
		sql, args, err := NewWhereClauseBuilder().
			Gt("age", 18).
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE age > ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test13 - ChainGt - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example14_ChainLt", func(t *testing.T) {
		// 测试链式调用 Lt
		sql, args, err := NewWhereClauseBuilder().
			Lt("age", 60).
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE age < ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test14 - ChainLt - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example15_ChainGte", func(t *testing.T) {
		// 测试链式调用 Gte
		sql, args, err := NewWhereClauseBuilder().
			Gte("score", 80).
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE score >= ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test15 - ChainGte - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example16_ChainLte", func(t *testing.T) {
		// 测试链式调用 Lte
		sql, args, err := NewWhereClauseBuilder().
			Lte("price", 100).
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE price <= ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test16 - ChainLte - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example17_ChainIn", func(t *testing.T) {
		// 测试链式调用 In
		sql, args, err := NewWhereClauseBuilder().
			In("id", 1, 2, 3, 4, 5).
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE id IN (?, ?, ?, ?, ?)"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		if len(args) != 5 {
			t.Errorf("Expected 5 args, got: %v", args)
		}
		
		fmt.Printf("Test17 - ChainIn - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example18_ChainNotIn", func(t *testing.T) {
		// 测试链式调用 NotIn
		sql, args, err := NewWhereClauseBuilder().
			NotIn("status", "deleted", "banned").
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE status NOT IN (?, ?)"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test18 - ChainNotIn - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example19_ChainBetween", func(t *testing.T) {
		// 测试链式调用 Between
		sql, args, err := NewWhereClauseBuilder().
			Between("created_at", "2024-01-01", "2024-12-31").
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE created_at BETWEEN ? AND ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test19 - ChainBetween - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example20_ChainOr", func(t *testing.T) {
		// 测试链式调用 Or
		sql, args, err := NewWhereClauseBuilder().
			Eq("status", "active").
			Or().
			Eq("status", "pending").
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE status = ? OR status = ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test20 - ChainOr - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example21_ChainAnd", func(t *testing.T) {
		// 测试链式调用 And
		sql, args, err := NewWhereClauseBuilder().
			Eq("status", "active").
			Or().
			Eq("status", "pending").
			And().
			Gte("age", 18).
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE status = ? OR status = ? AND age >= ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test21 - ChainAnd - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example22_ChainGroup", func(t *testing.T) {
		// 测试链式调用 Group
		sql, args, err := NewWhereClauseBuilder().
			Eq("status", "active").
			And().
			Group(
				ConditionEq("age", 18).Or(),
				ConditionEq("age", 19).Or(),
			).
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		fmt.Printf("Test22 - ChainGroup - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example23_ChainAndGroup", func(t *testing.T) {
		// 测试链式调用 AndGroup
		sql, args, err := NewWhereClauseBuilder().
			Eq("status", "active").
			And().
			AndGroup(
				ConditionEq("age", 18),
				ConditionGte("score", 80),
			).
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		fmt.Printf("Test23 - ChainAndGroup - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example24_ChainOrGroup", func(t *testing.T) {
		// 测试链式调用 OrGroup
		sql, args, err := NewWhereClauseBuilder().
			Eq("status", "active").
			And().
			OrGroup(
				ConditionEq("age", 18),
				ConditionEq("age", 19),
			).
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		fmt.Printf("Test24 - ChainOrGroup - SQL: %s, Args: %v\n", sql, args)
	})
	
	t.Run("Example25_ComplexChain", func(t *testing.T) {
		// 测试复杂的链式调用
		sql, args, err := NewWhereClauseBuilder().
			Eq("status", "active").
			And().
			Gte("age", 18).
			And().
			In("role", "admin", "user").
			And().
			Between("created_at", "2024-01-01", "2024-12-31").
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		fmt.Printf("Test25 - ComplexChain - SQL: %s\n", sql)
		fmt.Printf("Test25 - ComplexChain - Args: %v\n", args)
	})
	
	t.Run("Example26_ChainWithOrAndMix", func(t *testing.T) {
		// 测试 OR 和 AND 混合的链式调用
		sql, args, err := NewWhereClauseBuilder().
			Eq("type", "user").
			Or().
			Eq("type", "admin").
			And().
			Neq("status", "deleted").
			Build()
		
		if err != nil {
			t.Fatal(err)
		}
		
		expectedSQL := "WHERE type = ? OR type = ? AND status != ?"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		fmt.Printf("Test26 - ChainWithOrAndMix - SQL: %s, Args: %v\n", sql, args)
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

// BenchmarkWhereBuilderChain 链式调用的性能测试
func BenchmarkWhereBuilderChain(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = NewWhereClauseBuilder().
			Eq("id", 1).
			And().
			Gte("age", 18).
			And().
			In("status", "active", "pending").
			Build()
	}
}
