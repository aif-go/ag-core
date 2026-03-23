package conditonwhere

import (
	"fmt"
	"reflect"
	"testing"
)

func TestWhereClauseBuilder_Basic(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *WhereClauseBuilder
		wantSQL  string
		wantArgs []interface{}
	}{
		{
			name: "简单 EQ 条件",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().Eq("name", "test")
			},
			wantSQL:  "name = ?",
			wantArgs: []interface{}{"test"},
		},
		{
			name: "多个 AND 条件",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					Eq("b", 2).
					Eq("c", 3)
			},
			wantSQL:  "a = ? AND b = ? AND c = ?",
			wantArgs: []interface{}{1, 2, 3},
		},
		{
			name: "Or 后接 AND 优先级",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					Or().
					Eq("b", 2).
					And().
					Eq("c", 3)
			},
			wantSQL:  "(a = ? OR b = ? AND c = ?)",
			wantArgs: []interface{}{1, 2, 3},
		},
		{
			name: "Or 链式调用",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("status", "active").
					Or().
					Eq("status", "pending")
			},
			wantSQL:  "status = ? OR status = ?",
			wantArgs: []interface{}{"active", "pending"},
		},
		{
			name: "多个 Or 连续",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					Or().
					Eq("b", 2).
					Or().
					Eq("c", 3)
			},
			wantSQL:  "a = ? OR b = ? OR c = ?",
			wantArgs: []interface{}{1, 2, 3},
		},
		{
			name: "Gt 条件",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().Gt("age", 18)
			},
			wantSQL:  "age > ?",
			wantArgs: []interface{}{18},
		},
		{
			name: "Lt 条件",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().Lt("price", 100)
			},
			wantSQL:  "price < ?",
			wantArgs: []interface{}{100},
		},
		{
			name: "Gte 条件",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().Gte("score", 60)
			},
			wantSQL:  "score >= ?",
			wantArgs: []interface{}{60},
		},
		{
			name: "Lte 条件",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().Lte("count", 10)
			},
			wantSQL:  "count <= ?",
			wantArgs: []interface{}{10},
		},
		{
			name: "Neq 条件",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().Neq("status", "deleted")
			},
			wantSQL:  "status != ?",
			wantArgs: []interface{}{"deleted"},
		},
		{
			name: "In 条件",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().In("status", "active", "pending", "draft")
			},
			wantSQL:  "status IN (?, ?, ?)",
			wantArgs: []interface{}{"active", "pending", "draft"},
		},
		{
			name: "NotIn 条件",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().NotIn("type", "system", "admin")
			},
			wantSQL:  "type NOT IN (?, ?)",
			wantArgs: []interface{}{"system", "admin"},
		},
		{
			name: "Between 条件",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().Between("age", 18, 65)
			},
			wantSQL:  "age BETWEEN ? AND ?",
			wantArgs: []interface{}{18, 65},
		},
		{
			name: "混合条件链式",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("name", "test").
					Gt("age", 18).
					In("status", "active", "pending")
			},
			wantSQL:  "name = ? AND age > ? AND status IN (?, ?)",
			wantArgs: []interface{}{"test", 18, "active", "pending"},
		},
		{
			name: "复杂 OR AND 混合",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					Or().
					Eq("b", 2).
					And().
					Eq("c", 3).
					Or().
					Eq("d", 4)
			},
			wantSQL:  "(a = ? OR b = ? AND c = ? OR d = ?)",
			wantArgs: []interface{}{1, 2, 3, 4},
		},
		{
			name: "空构建器",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder()
			},
			wantSQL:  "",
			wantArgs: nil,
		},
		{
			name: "And 方法显式调用",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					Or().
					Eq("b", 2).
					And()
			},
			wantSQL:  "a = ? OR b = ?",
			wantArgs: []interface{}{1, 2},
		},
		{
			name: "Or 后立即接 Or",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					Or().
					Or().
					Eq("b", 2)
			},
			wantSQL:  "a = ? OR b = ?",
			wantArgs: []interface{}{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			gotSQL, gotArgs, err := builder.Build()
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}
			if gotSQL != tt.wantSQL {
				t.Errorf("Build() SQL = %v, want %v", gotSQL, tt.wantSQL)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("Build() Args = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

// func TestWhereClauseBuilder_Nested(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		setup    func() *WhereClauseBuilder
// 		wantSQL  string
// 		wantArgs []interface{}
// 	}{
// 		{
// 			name: "Group 嵌套条件",
// 			setup: func() *WhereClauseBuilder {
// 				return NewWhereClauseBuilder().
// 					Eq("a", 1).
// 					Group(
// 						ConditionEq("b", 2),
// 						ConditionEq("c", 3),
// 					)
// 			},
// 			wantSQL:  "a = ? AND (b = ? AND c = ?)",
// 			wantArgs: []interface{}{1, 2, 3},
// 		},
// 		{
// 			name: "Group 使用 OR",
// 			setup: func() *WhereClauseBuilder {
// 				return NewWhereClauseBuilder().
// 					Eq("a", 1).
// 					Group(
// 						ConditionEq("b", 2).Or(),
// 						ConditionEq("c", 3),
// 					)
// 			},
// 			wantSQL:  "a = ? AND (b = ? AND c = ?)",
// 			wantArgs: []interface{}{1, 2, 3},
// 		},
// 		{
// 			name: "AndGroup",
// 			setup: func() *WhereClauseBuilder {
// 				return NewWhereClauseBuilder().
// 					Eq("a", 1).
// 					AndGroup(
// 						ConditionEq("b", 2),
// 						ConditionEq("c", 3),
// 					)
// 			},
// 			wantSQL:  "a = ? AND (b = ? AND c = ?)",
// 			wantArgs: []interface{}{1, 2, 3},
// 		},
// 		{
// 			name: "OrGroup",
// 			setup: func() *WhereClauseBuilder {
// 				return NewWhereClauseBuilder().
// 					Eq("a", 1).
// 					OrGroup(
// 						ConditionEq("b", 2),
// 						ConditionEq("c", 3),
// 					)
// 			},
// 			wantSQL:  "a = ? OR (b = ? OR c = ?)",
// 			wantArgs: []interface{}{1, 2, 3},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			builder := tt.setup()
// 			gotSQL, gotArgs, err := builder.Build()
// 			if err != nil {
// 				t.Fatalf("Build() error = %v", err)
// 			}
// 			if gotSQL != tt.wantSQL {
// 				t.Errorf("Build() SQL = %v, want %v", gotSQL, tt.wantSQL)
// 			}
// 			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
// 				t.Errorf("Build() Args = %v, want %v", gotArgs, tt.wantArgs)
// 			}
// 		})
// 	}
// }

// func TestWhereClauseBuilder_SetRoot(t *testing.T) {
// 	builder := NewWhereClauseBuilder().
// 		SetRoot(ConditionEq("id", 100)).
// 		Eq("status", "active")

// 	gotSQL, gotArgs, err := builder.Build()
// 	if err != nil {
// 		t.Fatalf("Build() error = %v", err)
// 	}
// 	wantSQL := "id = ? AND status = ?"
// 	wantArgs := []interface{}{100, "active"}

// 	if gotSQL != wantSQL {
// 		t.Errorf("Build() SQL = %v, want %v", gotSQL, wantSQL)
// 	}
// 	if !reflect.DeepEqual(gotArgs, wantArgs) {
// 		t.Errorf("Build() Args = %v, want %v", gotArgs, wantArgs)
// 	}
// }

// func TestWhereClauseBuilder_AddCondition(t *testing.T) {
// 	builder := NewWhereClauseBuilder().
// 		AddCondition(ConditionEq("a", 1)).
// 		AddCondition(ConditionNeq("b", 2))

// 	gotSQL, gotArgs, err := builder.Build()
// 	if err != nil {
// 		t.Fatalf("Build() error = %v", err)
// 	}
// 	wantSQL := "a = ? AND b != ?"
// 	wantArgs := []interface{}{1, 2}

// 	if gotSQL != wantSQL {
// 		t.Errorf("Build() SQL = %v, want %v", gotSQL, wantSQL)
// 	}
// 	if !reflect.DeepEqual(gotArgs, wantArgs) {
// 		t.Errorf("Build() Args = %v, want %v", gotArgs, wantArgs)
// 	}
// }

// func TestWhereClauseBuilder_AddConditions(t *testing.T) {
// 	builder := NewWhereClauseBuilder().
// 		AddConditions(
// 			ConditionEq("a", 1),
// 			ConditionEq("b", 2),
// 			ConditionEq("c", 3).Or(),
// 		)

// 	gotSQL, gotArgs, err := builder.Build()
// 	if err != nil {
// 		t.Fatalf("Build() error = %v", err)
// 	}
// 	wantSQL := "(a = ? AND b = ? OR c = ?)"
// 	wantArgs := []interface{}{1, 2, 3}

// 	if gotSQL != wantSQL {
// 		t.Errorf("Build() SQL = %v, want %v", gotSQL, wantSQL)
// 	}
// 	if !reflect.DeepEqual(gotArgs, wantArgs) {
// 		t.Errorf("Build() Args = %v, want %v", gotArgs, wantArgs)
// 	}
// }

func TestWhereClauseBuilder_IndexCheck(t *testing.T) {
	tests := []struct {
		name              string
		setup             func() *WhereClauseBuilder
		primaryKeyColumns []string
		indexColumns      [][]string
		wantSQL           string
		wantArgs          []interface{}
		wantIndex         string
		wantErr           bool
	}{
		{
			name: "使用主键",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().Eq("id", 100)
			},
			primaryKeyColumns: []string{"id"},
			indexColumns:      [][]string{},
			wantSQL:           "WHERE id = ?",
			wantArgs:          []interface{}{100},
			wantIndex:         "PRIMARY",
			wantErr:           false,
		},
		{
			name: "使用索引",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().Eq("user_id", 1)
			},
			primaryKeyColumns: []string{"id"},
			indexColumns:      [][]string{{"user_id", "name"}},
			wantSQL:           "WHERE user_id = ?",
			wantArgs:          []interface{}{1},
			wantIndex:         "INDEX_0",
			wantErr:           false,
		},
		{
			name: "无索引",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().Eq("name", "test")
			},
			primaryKeyColumns: []string{"id"},
			indexColumns:      [][]string{{"user_id", "name"}},
			wantSQL:           "",
			wantArgs:          nil,
			wantIndex:         "",
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			gotSQL, gotArgs, gotIndex, err := builder.BuildWithIndexCheck(
				tt.primaryKeyColumns,
				tt.indexColumns,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("BuildWithIndexCheck() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("BuildWithIndexCheck() error = %v", err)
			}

			if gotSQL != tt.wantSQL {
				t.Errorf("BuildWithIndexCheck() SQL = %v, want %v", gotSQL, tt.wantSQL)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("BuildWithIndexCheck() Args = %v, want %v", gotArgs, tt.wantArgs)
			}
			if gotIndex != tt.wantIndex {
				t.Errorf("BuildWithIndexCheck() Index = %v, want %v", gotIndex, tt.wantIndex)
			}
		})
	}
}

func TestWhereClauseBuilder_ChainedNested(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *WhereClauseBuilder
		wantSQL  string
		wantArgs []interface{}
	}{

		{
			name: "简单的链式",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).And().Eq("b",2).And().Gt("c", 3)
					
			},
			wantSQL:  "a = ? AND b = ? AND c > ?",
			wantArgs: []interface{}{1, 2, 3},
		},
		{
			name: "简单嵌套 BeginGroup/EndGroup",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					BeginGroup().
					Eq("b", 2).
					Or().
					Eq("c", 3).
					EndGroup()
			},
			wantSQL:  "a = ? AND (b = ? OR c = ?)",
			wantArgs: []interface{}{1, 2, 3},
		},
		{
			name: "多层嵌套",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					BeginGroup().
					Eq("b", 2).
					BeginGroup().
					Eq("c", 3).
					Or().
					Eq("d", 4).
					EndGroup().
					EndGroup()
			},
			wantSQL:  "a = ? AND (b = ? AND (c = ? OR d = ?))",
			wantArgs: []interface{}{1, 2, 3, 4},
		},
		{
			name: "嵌套后接其他条件",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					BeginGroup().
					Eq("b", 2).
					Or().
					Eq("c", 3).
					EndGroup().
					Eq("d", 4)
			},
			wantSQL:  "a = ? AND (b = ? OR c = ?) AND d = ?",
			wantArgs: []interface{}{1, 2, 3, 4},
		},
		{
			name: "OR 后接嵌套组",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					Or().
					BeginGroup().
					Eq("b", 2).
					And().
					Eq("c", 3).
					EndGroup()
			},
			wantSQL:  "a = ? OR (b = ? AND c = ?)",
			wantArgs: []interface{}{1, 2, 3},
		},
		{
			name: "空嵌套组",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					BeginGroup().
					EndGroup().
					Eq("b", 2)
			},
			wantSQL:  "a = ? AND b = ?",
			wantArgs: []interface{}{1, 2},
		},
		{
			name: "连续嵌套组",
			setup: func() *WhereClauseBuilder {
				return NewWhereClauseBuilder().
					Eq("a", 1).
					BeginGroup().
					Eq("b", 2).
					Or().
					Eq("c", 3).
					EndGroup().
					BeginGroup().
					Eq("d", 4).
					And().
					Eq("e", 5).
					EndGroup()
			},
			wantSQL:  "a = ? AND (b = ? OR c = ?) AND (d = ? AND e = ?)",
			wantArgs: []interface{}{1, 2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			gotSQL, gotArgs, err := builder.Build()
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}
			if gotSQL != tt.wantSQL {
				t.Errorf("Build() SQL = %v, want %v", gotSQL, tt.wantSQL)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("Build() Args = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestConditionHelpers(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() string
		wantSQL  string
		wantArgs []interface{}
	}{
		{
			name: "ConditionEq 单独使用",
			setup: func() string {
				cond := ConditionEq("id", 1)
				return fmt.Sprintf("%v", cond)
			},
			wantSQL:  "{id = 1}",
			wantArgs: nil,
		},
		{
			name: "WhereCondition Or 方法",
			setup: func() string {
				cond := ConditionEq("a", 1)
				cond.Or()
				cond.AddChild(ConditionEq("b", 2))
				builder := NewWhereClauseBuilder().AddCondition(cond)
				sql, args, _ := builder.Build()
				return fmt.Sprintf("%s %v", sql, args)
			},
			wantSQL:  "a = ? OR b = ? [1 2]",
			wantArgs: nil,
		},
		{
			name: "WhereCondition And 方法",
			setup: func() string {
				cond := ConditionEq("a", 1)
				cond.And()
				cond.AddChild(ConditionEq("b", 2))
				builder := NewWhereClauseBuilder().AddCondition(cond)
				sql, args, _ := builder.Build()
				return fmt.Sprintf("%s %v", sql, args)
			},
			wantSQL:  "a = ? AND b = ? [1 2]",
			wantArgs: nil,
		},
		{
			name: "WhereCondition AddChild",
			setup: func() string {
				cond := ConditionEq("a", 1).AddChild(ConditionEq("b", 2))
				builder := NewWhereClauseBuilder().AddCondition(cond)
				sql, args, _ := builder.Build()
				return fmt.Sprintf("%s %v", sql, args)
			},
			wantSQL:  "(a = ? AND b = ?) [1 2]",
			wantArgs: nil,
		},
		{
			name: "ConditionGroup",
			setup: func() string {
				cond := ConditionGroup(
					ConditionEq("a", 1),
					ConditionEq("b", 2).Or(),
				)
				builder := NewWhereClauseBuilder().AddCondition(cond)
				sql, args, _ := builder.Build()
				return fmt.Sprintf("%s %v", sql, args)
			},
			wantSQL:  "(a = ? OR b = ?) [1 2]",
			wantArgs: nil,
		},
		{
			name: "ConditionAndGroup",
			setup: func() string {
				cond := ConditionAndGroup(
					ConditionEq("a", 1),
					ConditionEq("b", 2),
				)
				builder := NewWhereClauseBuilder().AddCondition(cond)
				sql, args, _ := builder.Build()
				return fmt.Sprintf("%s %v", sql, args)
			},
			wantSQL:  "(a = ? AND b = ?) [1 2]",
			wantArgs: nil,
		},
		{
			name: "ConditionOrGroup",
			setup: func() string {
				cond := ConditionOrGroup(
					ConditionEq("a", 1),
					ConditionEq("b", 2),
				)
				builder := NewWhereClauseBuilder().AddCondition(cond)
				sql, args, _ := builder.Build()
				return fmt.Sprintf("%s %v", sql, args)
			},
			wantSQL:  "(a = ? OR b = ?) [1 2]",
			wantArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setup()
			t.Logf("Result: %s", result)
		})
	}
}
