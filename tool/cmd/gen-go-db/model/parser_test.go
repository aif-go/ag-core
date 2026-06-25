package model

import (
	"testing"
)

func TestGetGoType(t *testing.T) {
	tests := []struct {
		name     string
		sqlType  string
		want     string
	}{
		{name: "decimal plain", sqlType: "decimal", want: "decimal.Decimal"},
		{name: "decimal with precision", sqlType: "decimal(18,2)", want: "decimal.Decimal"},
		{name: "decimal uppercase", sqlType: "DECIMAL", want: "decimal.Decimal"},
		{name: "decimal uppercase with precision", sqlType: "DECIMAL(18,2)", want: "decimal.Decimal"},
		{name: "int unchanged", sqlType: "int", want: "int"},
		{name: "int64 unchanged", sqlType: "bigint", want: "int64"},
		{name: "float unchanged", sqlType: "float", want: "float64"},
		{name: "double unchanged", sqlType: "double", want: "float64"},
		{name: "varchar unchanged", sqlType: "varchar", want: "string"},
		{name: "datetime unchanged", sqlType: "datetime", want: "time.Time"},
		{name: "bool unchanged", sqlType: "boolean", want: "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getGoType(tt.sqlType)
			if got != tt.want {
				t.Errorf("getGoType(%q) = %q, want %q", tt.sqlType, got, tt.want)
			}
		})
	}
}
