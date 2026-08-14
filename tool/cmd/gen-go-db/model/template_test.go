package model

import (
	"strings"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/table"
)

// TestIsSpecialColumn_OptimisticLock 验证乐观锁列被视为特殊列（ListZeroValueCols 的 filterSpecial 排除）。
func TestIsSpecialColumn_OptimisticLock(t *testing.T) {
	if !isSpecialColumn(table.ColumnData{IsOptimisticLock: true}) {
		t.Errorf("IsOptimisticLock 列应视为特殊列")
	}
	if isSpecialColumn(table.ColumnData{}) {
		t.Errorf("普通列不应视为特殊列")
	}
	if !isSpecialColumn(table.ColumnData{IsJavaVersion: true}) {
		t.Errorf("javaVersion 列应仍视为特殊列（回归）")
	}
}

// TestGenerateStruct_OptimisticLockField 验证结构体字段类型生成乐观锁列时输出 optimisticlock.Version。
func TestGenerateStruct_OptimisticLockField(t *testing.T) {
	td := &table.TableData{
		StructName: "TmOptLock",
		Columns: []table.ColumnData{
			{Name: "id", JsonTag: "Id", GoType: "int64", GormTag: "column:id"},
			{Name: "jpa_version", JsonTag: "JpaVersion", GoType: "optimisticlock.Version", GormTag: "column:jpa_version"},
		},
	}
	code := generateStruct(td)
	if !strings.Contains(code, "JpaVersion optimisticlock.Version") {
		t.Errorf("结构体应生成乐观锁字段类型 optimisticlock.Version, got:\n%s", code)
	}
}
