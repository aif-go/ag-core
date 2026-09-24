package table

import "testing"

// TestZeroCheckExpr_Matrix 零值表达式矩阵：零值/非零值双形态逐类型锁定语义
func TestZeroCheckExpr_Matrix(t *testing.T) {
	cases := []struct {
		name    string
		col     ColumnData
		zero    string // invert=false 期望
		notZero string // invert=true 期望
	}{
		{col: ColumnData{GoType: "*time.Time", JsonTag: "TimePointer"}, zero: "entity.TimePointer == nil", notZero: "entity.TimePointer != nil"},
		{col: ColumnData{GoType: "*decimal.Decimal", JsonTag: "Amount"}, zero: "entity.Amount == nil", notZero: "entity.Amount != nil"},
		{col: ColumnData{GoType: "string", JsonTag: "Name"}, zero: `entity.Name == ""`, notZero: `entity.Name != ""`},
		{col: ColumnData{GoType: "time.Time", JsonTag: "CreateTime"}, zero: "entity.CreateTime.IsZero()", notZero: "!entity.CreateTime.IsZero()"},
		{col: ColumnData{GoType: "decimal.Decimal", JsonTag: "Salary"}, zero: "entity.Salary.IsZero()", notZero: "!entity.Salary.IsZero()"},
		{col: ColumnData{GoType: "optimisticlock.Version", JsonTag: "JpaVersion"}, zero: "!entity.JpaVersion.Valid", notZero: "entity.JpaVersion.Valid"},
		{col: ColumnData{GoType: "bool", JsonTag: "IsGraduate"}, zero: "!entity.IsGraduate", notZero: "entity.IsGraduate"},
		{col: ColumnData{GoType: "int64", JsonTag: "Id"}, zero: "entity.Id == 0", notZero: "entity.Id != 0"},
		{col: ColumnData{GoType: "float64", JsonTag: "Ratio"}, zero: "entity.Ratio == 0", notZero: "entity.Ratio != 0"},
	}

	for _, tc := range cases {
		if got := ZeroCheckExpr("entity", tc.col, false); got != tc.zero {
			t.Errorf("%s 零值: got %q want %q", tc.col.JsonTag, got, tc.zero)
		}
		if got := ZeroCheckExpr("entity", tc.col, true); got != tc.notZero {
			t.Errorf("%s 非零值: got %q want %q", tc.col.JsonTag, got, tc.notZero)
		}
	}
}
