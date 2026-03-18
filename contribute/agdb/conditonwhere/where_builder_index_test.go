package conditonwhere

import (
	"testing"
)

// 测试索引验证功能
func TestBuildWithIndexCheck(t *testing.T) {
	// 定义主键列
	primaryKeyColumns := []string{"ID"}

	// 定义索引列
	indexColumns := [][]string{
		{"OrderId", "MerchantId", "TransactionType"},      // IDX_ORDER_ID
		{"TransmissionDatetime", "Stan", "TransactionType"}, // IDX_STAN
		{"RetrievalReferenceNumber", "MerchantId", "TransactionType"}, // IDX_RRN
		{"InsertTimestamp"}, // IDX_INSERT_TIMESTATMP
	}

	t.Run("使用主键查询", func(t *testing.T) {
		builder := NewWhereClauseBuilder()
		builder.AddCondition(ConditionEq("ID", 123))

		sql, args, usedIndex, err := builder.BuildWithIndexCheck(primaryKeyColumns, indexColumns)
		if err != nil {
			t.Fatalf("BuildWithIndexCheck error: %v", err)
		}

		if usedIndex != "PRIMARY" {
			t.Errorf("expected usedIndex to be 'PRIMARY', got '%s'", usedIndex)
		}

		// 注意：buildWhereCondition会在条件周围添加括号
		if sql != "WHERE (ID = ?)" {
			t.Errorf("expected sql to be 'WHERE (ID = ?)', got '%s'", sql)
		}

		if len(args) != 1 || args[0] != 123 {
			t.Errorf("expected args to be [123], got %v", args)
		}
	})

	t.Run("使用索引IDX_ORDER_ID查询", func(t *testing.T) {
		builder := NewWhereClauseBuilder()
		builder.AddCondition(ConditionEq("OrderId", "123456"))
		builder.AddCondition(ConditionEq("MerchantId", "M001"))

		sql, args, usedIndex, err := builder.BuildWithIndexCheck(primaryKeyColumns, indexColumns)
		if err != nil {
			t.Fatalf("BuildWithIndexCheck error: %v", err)
		}

		if usedIndex != "INDEX_0" {
			t.Errorf("expected usedIndex to be 'INDEX_0', got '%s'", usedIndex)
		}

		// 注意：buildWhereCondition会在条件周围添加括号
		if sql != "WHERE (OrderId = ? AND (MerchantId = ?))" {
			t.Errorf("expected sql to be 'WHERE (OrderId = ? AND (MerchantId = ?))', got '%s'", sql)
		}

		if len(args) != 2 || args[0] != "123456" || args[1] != "M001" {
			t.Errorf("expected args to be ['123456', 'M001'], got %v", args)
		}

		// 使用sql和args避免编译器警告
		_ = sql
		_ = args
	})

	t.Run("使用索引IDX_STAN查询", func(t *testing.T) {
		builder := NewWhereClauseBuilder()
		builder.AddCondition(ConditionEq("TransmissionDatetime", "20240101"))
		builder.AddCondition(ConditionEq("Stan", "123456"))

		sql, args, usedIndex, err := builder.BuildWithIndexCheck(primaryKeyColumns, indexColumns)
		if err != nil {
			t.Fatalf("BuildWithIndexCheck error: %v", err)
		}

		if usedIndex != "INDEX_1" {
			t.Errorf("expected usedIndex to be 'INDEX_1', got '%s'", usedIndex)
		}

		// 使用sql和args避免编译器警告
		_ = sql
		_ = args
	})

	t.Run("使用索引IDX_RRN查询", func(t *testing.T) {
		builder := NewWhereClauseBuilder()
		builder.AddCondition(ConditionEq("RetrievalReferenceNumber", "123456789012"))

		sql, args, usedIndex, err := builder.BuildWithIndexCheck(primaryKeyColumns, indexColumns)
		if err != nil {
			t.Fatalf("BuildWithIndexCheck error: %v", err)
		}

		if usedIndex != "INDEX_2" {
			t.Errorf("expected usedIndex to be 'INDEX_2', got '%s'", usedIndex)
		}

		// 使用sql和args避免编译器警告
		_ = sql
		_ = args
	})

	t.Run("使用索引IDX_INSERT_TIMESTATMP查询", func(t *testing.T) {
		builder := NewWhereClauseBuilder()
		builder.AddCondition(ConditionEq("InsertTimestamp", "20240101"))

		sql, args, usedIndex, err := builder.BuildWithIndexCheck(primaryKeyColumns, indexColumns)
		if err != nil {
			t.Fatalf("BuildWithIndexCheck error: %v", err)
		}

		if usedIndex != "INDEX_3" {
			t.Errorf("expected usedIndex to be 'INDEX_3', got '%s'", usedIndex)
		}

		// 使用sql和args避免编译器警告
		_ = sql
		_ = args
	})

	t.Run("未使用任何索引应返回错误", func(t *testing.T) {
		builder := NewWhereClauseBuilder()
		builder.AddCondition(ConditionEq("ResultCode", "00"))

		_, _, _, err := builder.BuildWithIndexCheck(primaryKeyColumns, indexColumns)
		if err == nil {
			t.Error("expected error when no index is used, got nil")
		}
	})

	t.Run("空构建器应返回空结果", func(t *testing.T) {
		builder := NewWhereClauseBuilder()

		sql, args, usedIndex, err := builder.BuildWithIndexCheck(primaryKeyColumns, indexColumns)
		if err != nil {
			t.Fatalf("BuildWithIndexCheck error: %v", err)
		}

		if sql != "" {
			t.Errorf("expected empty sql, got '%s'", sql)
		}

		if len(args) != 0 {
			t.Errorf("expected empty args, got %v", args)
		}

		if usedIndex != "" {
			t.Errorf("expected empty usedIndex, got '%s'", usedIndex)
		}
	})

	t.Run("嵌套条件使用索引", func(t *testing.T) {
		builder := NewWhereClauseBuilder()
		builder.AddCondition(ConditionEq("OrderId", "123456"))
		builder.AddCondition(ConditionEq("MerchantId", "M001"))
		builder.AddCondition(ConditionEq("TransactionType", "02"))

		sql, args, usedIndex, err := builder.BuildWithIndexCheck(primaryKeyColumns, indexColumns)
		if err != nil {
			t.Fatalf("BuildWithIndexCheck error: %v", err)
		}

		if usedIndex != "INDEX_0" {
			t.Errorf("expected usedIndex to be 'INDEX_0', got '%s'", usedIndex)
		}

		// 注意：buildWhereCondition会在条件周围添加括号
		if sql != "WHERE (OrderId = ? AND (MerchantId = ?) AND AND (TransactionType = ?))" {
			t.Errorf("expected sql to be 'WHERE (OrderId = ? AND (MerchantId = ?) AND AND (TransactionType = ?))', got '%s'", sql)
		}

		// 使用args避免编译器警告
		_ = args
	})
}

// 测试collectConditionFields函数
func TestCollectConditionFields(t *testing.T) {
	t.Run("收集单个条件的字段", func(t *testing.T) {
		cond := ConditionEq("OrderId", "123456")
		fields := collectConditionFields(cond)

		if len(fields) != 1 || fields[0] != "OrderId" {
			t.Errorf("expected fields to be ['OrderId'], got %v", fields)
		}
	})

	t.Run("收集多个条件的字段", func(t *testing.T) {
		cond := ConditionEq("OrderId", "123456")
		cond.AddChild(ConditionEq("MerchantId", "M001"))
		cond.AddChild(ConditionEq("TransactionType", "02"))

		fields := collectConditionFields(cond)

		if len(fields) != 3 {
			t.Errorf("expected 3 fields, got %d", len(fields))
		}

		fieldSet := make(map[string]bool)
		for _, field := range fields {
			fieldSet[field] = true
		}

		if !fieldSet["OrderId"] || !fieldSet["MerchantId"] || !fieldSet["TransactionType"] {
			t.Errorf("expected fields to contain OrderId, MerchantId, TransactionType, got %v", fields)
		}
	})

	t.Run("收集嵌套条件的字段", func(t *testing.T) {
		innerCond := ConditionEq("MerchantId", "M001")
		innerCond.AddChild(ConditionEq("TransactionType", "02"))

		outerCond := ConditionEq("OrderId", "123456")
		outerCond.AddChild(innerCond)

		fields := collectConditionFields(outerCond)

		if len(fields) != 3 {
			t.Errorf("expected 3 fields, got %d", len(fields))
		}
	})
}

// 测试isPrimaryKeyUsed函数
func TestIsPrimaryKeyUsed(t *testing.T) {
	t.Run("使用主键引导列", func(t *testing.T) {
		primaryKeyColumns := []string{"ID"}
		usedFields := []string{"ID", "OrderId"}

		if !isPrimaryKeyUsed(usedFields, primaryKeyColumns) {
			t.Error("expected isPrimaryKeyUsed to return true")
		}
	})

	t.Run("不使用主键引导列", func(t *testing.T) {
		primaryKeyColumns := []string{"ID"}
		usedFields := []string{"OrderId", "MerchantId"}

		if isPrimaryKeyUsed(usedFields, primaryKeyColumns) {
			t.Error("expected isPrimaryKeyUsed to return false")
		}
	})

	t.Run("空主键列", func(t *testing.T) {
		primaryKeyColumns := []string{}
		usedFields := []string{"ID", "OrderId"}

		if isPrimaryKeyUsed(usedFields, primaryKeyColumns) {
			t.Error("expected isPrimaryKeyUsed to return false for empty primaryKeyColumns")
		}
	})
}

// 测试findMatchingIndex函数
func TestFindMatchingIndex(t *testing.T) {
	indexColumns := [][]string{
		{"OrderId", "MerchantId", "TransactionType"},
		{"TransmissionDatetime", "Stan", "TransactionType"},
		{"RetrievalReferenceNumber", "MerchantId", "TransactionType"},
		{"InsertTimestamp"},
	}

	t.Run("匹配第一个索引", func(t *testing.T) {
		usedFields := []string{"OrderId", "MerchantId"}
		usedIndex := findMatchingIndex(usedFields, indexColumns)

		if usedIndex != "INDEX_0" {
			t.Errorf("expected usedIndex to be 'INDEX_0', got '%s'", usedIndex)
		}
	})

	t.Run("匹配第二个索引", func(t *testing.T) {
		usedFields := []string{"TransmissionDatetime", "Stan"}
		usedIndex := findMatchingIndex(usedFields, indexColumns)

		if usedIndex != "INDEX_1" {
			t.Errorf("expected usedIndex to be 'INDEX_1', got '%s'", usedIndex)
		}
	})

	t.Run("匹配第三个索引", func(t *testing.T) {
		usedFields := []string{"RetrievalReferenceNumber"}
		usedIndex := findMatchingIndex(usedFields, indexColumns)

		if usedIndex != "INDEX_2" {
			t.Errorf("expected usedIndex to be 'INDEX_2', got '%s'", usedIndex)
		}
	})

	t.Run("匹配第四个索引", func(t *testing.T) {
		usedFields := []string{"InsertTimestamp"}
		usedIndex := findMatchingIndex(usedFields, indexColumns)

		if usedIndex != "INDEX_3" {
			t.Errorf("expected usedIndex to be 'INDEX_3', got '%s'", usedIndex)
		}
	})

	t.Run("不匹配任何索引", func(t *testing.T) {
		usedFields := []string{"ResultCode", "RiskLevel"}
		usedIndex := findMatchingIndex(usedFields, indexColumns)

		if usedIndex != "" {
			t.Errorf("expected usedIndex to be empty, got '%s'", usedIndex)
		}
	})

	t.Run("使用索引的非引导列", func(t *testing.T) {
		usedFields := []string{"MerchantId", "TransactionType"}
		usedIndex := findMatchingIndex(usedFields, indexColumns)

		if usedIndex != "" {
			t.Errorf("expected usedIndex to be empty when using non-leading columns, got '%s'", usedIndex)
		}
	})
}
