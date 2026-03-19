package gormdb

import (
	"testing"
)

// TestOrderBuilder 测试 OrderBuilder 的链式调用
func TestOrderBuilder(t *testing.T) {
	
	t.Run("Example1_SimpleAsc", func(t *testing.T) {
		// 测试简单的升序排序
		sql := NewOrderBuilder().
			Asc("id").
			Build()
		
		expectedSQL := "ORDER BY id ASC"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		t.Logf("Test1 - SimpleAsc - SQL: %s\n", sql)
	})
	
	t.Run("Example2_SimpleDesc", func(t *testing.T) {
		// 测试简单的降序排序
		sql := NewOrderBuilder().
			Desc("created_at").
			Build()
		
		expectedSQL := "ORDER BY created_at DESC"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		t.Logf("Test2 - SimpleDesc - SQL: %s\n", sql)
	})
	
	t.Run("Example3_MultipleOrders", func(t *testing.T) {
		// 测试多个排序条件
		sql := NewOrderBuilder().
			Asc("status").
			Desc("created_at").
			Asc("id").
			Build()
		
		expectedSQL := "ORDER BY status ASC, created_at DESC, id ASC"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		t.Logf("Test3 - MultipleOrders - SQL: %s\n", sql)
	})
	
	t.Run("Example4_UsingOrderMethod", func(t *testing.T) {
		// 测试使用 Order 方法
		sql := NewOrderBuilder().
			Order("name", ASC).
			Order("age", DESC).
			Build()
		
		expectedSQL := "ORDER BY name ASC, age DESC"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		t.Logf("Test4 - UsingOrderMethod - SQL: %s\n", sql)
	})
	
	t.Run("Example5_UsingOrdersMethod", func(t *testing.T) {
		// 测试使用 Orders 方法批量添加
		sql := NewOrderBuilder().
			Orders(
				Order{ColName: "id", Sort: ASC},
				Order{ColName: "created_at", Sort: DESC},
			).
			Build()
		
		expectedSQL := "ORDER BY id ASC, created_at DESC"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		t.Logf("Test5 - UsingOrdersMethod - SQL: %s\n", sql)
	})
	
	t.Run("Example6_BuildWithoutKeyword", func(t *testing.T) {
		// 测试不包含 ORDER BY 关键字的构建
		sql := NewOrderBuilder().
			Asc("id").
			BuildWithoutKeyword()
		
		expectedSQL := "id ASC"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		t.Logf("Test6 - BuildWithoutKeyword - SQL: %s\n", sql)
	})
	
	t.Run("Example7_EmptyBuilder", func(t *testing.T) {
		// 测试空的构建器
		sql := NewOrderBuilder().Build()
		if sql != "" {
			t.Errorf("Expected empty SQL, got: %s", sql)
		}
		
		t.Logf("Test7 - EmptyBuilder - SQL: '%s'\n", sql)
	})
	
	t.Run("Example8_ToOrders", func(t *testing.T) {
		// 测试转换为 Order 切片
		orders := NewOrderBuilder().
			Asc("id").
			Desc("created_at").
			ToOrders()
		
		if len(orders) != 2 {
			t.Errorf("Expected 2 orders, got: %d", len(orders))
		}
		
		if orders[0].ColName != "id" || orders[0].Sort != ASC {
			t.Errorf("Expected first order: id ASC, got: %s %s", orders[0].ColName, orders[0].Sort)
		}
		
		if orders[1].ColName != "created_at" || orders[1].Sort != DESC {
			t.Errorf("Expected second order: created_at DESC, got: %s %s", orders[1].ColName, orders[1].Sort)
		}
		
		t.Logf("Test8 - ToOrders - Orders: %v\n", orders)
	})
	
	t.Run("Example9_Clear", func(t *testing.T) {
		// 测试清空构建器
		builder := NewOrderBuilder().
			Asc("id").
			Desc("created_at")
		
		// 清空后应该返回空字符串
		sql := builder.Clear().Build()
		if sql != "" {
			t.Errorf("Expected empty SQL after clear, got: %s", sql)
		}
		
		t.Logf("Test9 - Clear - SQL: '%s'\n", sql)
	})
	
	t.Run("Example10_ComplexChain", func(t *testing.T) {
		// 测试复杂的链式调用
		sql := NewOrderBuilder().
			Asc("status").
			Desc("created_at").
			Asc("priority").
			Desc("updated_at").
			Build()
		
		expectedSQL := "ORDER BY status ASC, created_at DESC, priority ASC, updated_at DESC"
		if sql != expectedSQL {
			t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
		}
		
		t.Logf("Test10 - ComplexChain - SQL: %s\n", sql)
	})
}

// BenchmarkOrderBuilder 性能测试
func BenchmarkOrderBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewOrderBuilder().
			Asc("id").
			Desc("created_at").
			Asc("status").
			Build()
	}
}

// BenchmarkOrderBuilderBuild 性能测试（仅构建）
func BenchmarkOrderBuilderBuild(b *testing.B) {
	builder := NewOrderBuilder().
		Asc("id").
		Desc("created_at").
		Asc("status")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = builder.Build()
	}
}
