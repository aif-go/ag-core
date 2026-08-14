package dao

import (
	"strings"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/table"
)

// TestDAO_PointerDispatch 验证 dao 模板对指针类型列生成 != nil 分派判断。
func TestDAO_PointerDispatch(t *testing.T) {
	t.Run("generateZeroValueCheck 指针类型生成 == nil", func(t *testing.T) {
		columns := []table.ColumnData{{GoType: "*time.Time", JsonTag: "createdAt"}}
		code := generateZeroValueCheck(columns)
		if !strings.Contains(code, "entity.createdAt == nil") {
			t.Errorf("generateZeroValueCheck 应生成 == nil 判断, got: %s", code)
		}
	})

	t.Run("generateZeroValueCheck time.Time 生成 IsZero", func(t *testing.T) {
		columns := []table.ColumnData{{GoType: "time.Time", JsonTag: "createdAt"}}
		code := generateZeroValueCheck(columns)
		if !strings.Contains(code, "entity.createdAt.IsZero()") {
			t.Errorf("generateZeroValueCheck 应生成 IsZero 判断, got: %s", code)
		}
	})

	t.Run("generateZeroValueCheck string 生成空串判断", func(t *testing.T) {
		columns := []table.ColumnData{{GoType: "string", JsonTag: "name"}}
		code := generateZeroValueCheck(columns)
		if !strings.Contains(code, `entity.name == ""`) {
			t.Errorf("generateZeroValueCheck 应生成 == \"\" 判断, got: %s", code)
		}
	})

	t.Run("GetDaoTemplate 主键与索引指针列生成 != nil", func(t *testing.T) {
		tableData := &table.TableData{
			ModuleName:  "github.com/aif-go/ag-core/tool/cmd/gen-go-db",
			TableName:   "tm_pointer",
			StructName:  "TmPointer",
			PrimaryKeys: []string{"created_at", "updated_at"},
			Columns: []table.ColumnData{
				{Name: "created_at", GoType: "*time.Time", JsonTag: "createdAt", IsPrimaryKey: true},
				{Name: "updated_at", GoType: "*time.Time", JsonTag: "updatedAt", IsPrimaryKey: true},
			},
			Indexes: []table.IndexData{
				{Name: "idx_pointer", Columns: []string{"created_at", "updated_at"}},
			},
		}
		code := GetDaoTemplate(tableData)
		for _, want := range []string{
			"entity.createdAt != nil",
			"entity.updatedAt != nil",
		} {
			if !strings.Contains(code, want) {
				t.Errorf("GetDaoTemplate 应生成指针 != nil 判断 %q, got:\n%s", want, code)
			}
		}
	})

	t.Run("GetDaoTemplate 索引引导列指针生成 != nil", func(t *testing.T) {
		tableData := &table.TableData{
			ModuleName:  "github.com/aif-go/ag-core/tool/cmd/gen-go-db",
			TableName:   "tm_pointer_idx",
			StructName:  "TmPointerIdx",
			PrimaryKeys: []string{"id"},
			Columns: []table.ColumnData{
				{Name: "id", GoType: "int64", JsonTag: "id", IsPrimaryKey: true},
				{Name: "created_date", GoType: "*time.Time", JsonTag: "createdDate"},
				{Name: "updated_date", GoType: "*time.Time", JsonTag: "updatedDate"},
			},
			Indexes: []table.IndexData{
				{Name: "idx_date", Columns: []string{"created_date", "updated_date"}},
			},
		}
		code := GetDaoTemplate(tableData)
		// 引导索引列生成指针非空判断；索引列纳入统一 where 拼接
		for _, want := range []string{
			"entity.createdDate != nil",
			"entity.ListZeroValueCols(true, false, true, false)",
		} {
			if !strings.Contains(code, want) {
				t.Errorf("GetDaoTemplate 应生成索引引导列指针 != nil 判断 %q, got:\n%s", want, code)
			}
		}
		// 次要索引列不再在索引块生成非空判断（统一由 ListZeroValueCols 过滤）
		for _, notWant := range []string{
			"entity.updatedDate != nil",
			"entity.createdDate != 0",
			"entity.updatedDate != 0",
		} {
			if strings.Contains(code, notWant) {
				t.Errorf("GetDaoTemplate 不应生成 %q, got:\n%s", notWant, code)
			}
		}
	})

	t.Run("generateZeroValueCheck bool 主键生成 !entity 判断", func(t *testing.T) {
		columns := []table.ColumnData{{GoType: "bool", JsonTag: "active"}}
		code := generateZeroValueCheck(columns)
		if !strings.Contains(code, "!entity.active") {
			t.Errorf("generateZeroValueCheck 应生成 !entity.active 判断, got: %s", code)
		}
		if strings.Contains(code, "entity.active == 0") {
			t.Errorf("generateZeroValueCheck bool 不应生成 == 0, got: %s", code)
		}
	})

	t.Run("GetDaoTemplate bool 主键与索引列生成非空判断", func(t *testing.T) {
		tableData := &table.TableData{
			ModuleName:  "github.com/aif-go/ag-core/tool/cmd/gen-go-db",
			TableName:   "tm_bool",
			StructName:  "TmBool",
			PrimaryKeys: []string{"active"},
			Columns: []table.ColumnData{
				{Name: "active", GoType: "bool", JsonTag: "active", IsPrimaryKey: true},
			},
			Indexes: []table.IndexData{
				{Name: "idx_bool", Columns: []string{"active"}},
			},
		}
		code := GetDaoTemplate(tableData)
		if !strings.Contains(code, "if entity.active {") {
			t.Errorf("GetDaoTemplate bool 主键应生成 entity.active 非空判断, got:\n%s", code)
		}
		for _, notWant := range []string{
			"entity.active != 0",
			"entity.active == 0",
		} {
			if strings.Contains(code, notWant) {
				t.Errorf("GetDaoTemplate bool 列不应生成 %q, got:\n%s", notWant, code)
			}
		}
	})

	t.Run("GetDaoTemplate 非指针主键回归生成 IsZero", func(t *testing.T) {
		tableData := &table.TableData{
			ModuleName:  "github.com/aif-go/ag-core/tool/cmd/gen-go-db",
			TableName:   "tm_time",
			StructName:  "TmTime",
			PrimaryKeys: []string{"created_at"},
			Columns: []table.ColumnData{
				{Name: "created_at", GoType: "time.Time", JsonTag: "createdAt", IsPrimaryKey: true},
			},
		}
		code := GetDaoTemplate(tableData)
		if !strings.Contains(code, "!entity.createdAt.IsZero()") {
			t.Errorf("GetDaoTemplate 非指针主键应生成 IsZero 判断, got:\n%s", code)
		}
	})
}
