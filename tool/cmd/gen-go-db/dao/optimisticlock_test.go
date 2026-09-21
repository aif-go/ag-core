package dao

import (
	"strings"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/table"
)

// lockTableData 单主键 + 乐观锁列表（锁列改名防硬编码：非 jpa_version）
func lockTableData() *table.TableData {
	return &table.TableData{
		ModuleName:  "github.com/aif-go/ag-core/tool/cmd/gen-go-db",
		TableName:   "tm_lock_demo",
		StructName:  "TmLockDemo",
		PrimaryKeys: []string{"id"},
		Columns: []table.ColumnData{
			{Name: "id", GoType: "int64", JsonTag: "Id", IsPrimaryKey: true},
			{Name: "version_no", GoType: "optimisticlock.Version", JsonTag: "VersionNo", IsOptimisticLock: true},
			{Name: "name", GoType: "string", JsonTag: "Name"},
		},
	}
}

// TestDAO_GenerateZeroValueCheck_OptimisticVersion 零值判断 switch 覆盖乐观锁列类型。
func TestDAO_GenerateZeroValueCheck_OptimisticVersion(t *testing.T) {
	code := generateZeroValueCheck([]table.ColumnData{
		{GoType: "optimisticlock.Version", JsonTag: "VersionNo"},
	})
	if !strings.Contains(code, "!entity.VersionNo.Valid") {
		t.Errorf("generateZeroValueCheck 未覆盖 optimisticlock.Version, got: %s", code)
	}
}

// TestDAO_LockCheck 乐观锁校验块生成：有锁列注入、无锁列为空。
func TestDAO_LockCheck(t *testing.T) {
	t.Run("有锁列", func(t *testing.T) {
		d := DaoTemplateData{TableData: lockTableData()}
		check := d.LockCheck()
		if !strings.Contains(check, "if !entity.VersionNo.Valid") {
			t.Errorf("LockCheck 未按 Valid 判断:\n%s", check)
		}
		if !strings.Contains(check, "when update,optimistic lock version is required") {
			t.Errorf("LockCheck 缺少明确报错文案:\n%s", check)
		}
	})
	t.Run("无锁列", func(t *testing.T) {
		td := lockTableData()
		for i := range td.Columns {
			td.Columns[i].IsOptimisticLock = false
		}
		d := DaoTemplateData{TableData: td}
		if d.LockCheck() != "" {
			t.Error("无锁列表不应生成任何校验代码")
		}
	})
}

// TestDAO_UpdateStmt_UpdateForm 更新形态统一（Save→Updates）与剔除语义保留。
func TestDAO_UpdateStmt_UpdateForm(t *testing.T) {
	d := DaoTemplateData{TableData: lockTableData()}

	full := d.UpdateStmt()
	if !strings.Contains(full, `db.Model(entity).Select("*").Updates(entity)`) {
		t.Errorf("单主键全字段更新应为 db.Model(entity).Select(...).Updates(entity):\n%s", full)
	}
	if strings.Contains(full, "Save(") {
		t.Errorf("UpdateByPrimaryKey 不应再使用 Save 形态:\n%s", full)
	}

	ignore := d.UpdateIgnoreStmt()
	if strings.Contains(ignore, `Select("*")`) {
		t.Errorf("UpdateIgnore 不应带 Select(\"*\")（保留零值剔除语义）:\n%s", ignore)
	}
	if !strings.Contains(ignore, "db.Model(entity).Updates(entity)") {
		t.Errorf("单主键部分更新形态不符:\n%s", ignore)
	}
}

// TestDAO_UpdateStmt_MultiPK 多主键形态走显式 where map。
func TestDAO_UpdateStmt_MultiPK(t *testing.T) {
	td := lockTableData()
	td.PrimaryKeys = []string{"id", "version_no"}
	td.Columns[1].IsPrimaryKey = true
	d := DaoTemplateData{TableData: td}

	if !strings.Contains(d.UpdateStmt(), `db.Model(&model.TmLockDemo{}).Where(where).Select("*").Updates(entity)`) {
		t.Errorf("多主键全字段更新应携带显式 Where:\n%s", d.UpdateStmt())
	}
	if !strings.Contains(d.UpdateIgnoreStmt(), `db.Model(&model.TmLockDemo{}).Where(where).Updates(entity)`) {
		t.Errorf("多主键部分更新应携带显式 Where:\n%s", d.UpdateIgnoreStmt())
	}
	if !strings.Contains(d.UpdatePreBlock(), "where := make(map[string]any)") {
		t.Errorf("多主键形态应生成 where map:\n%s", d.UpdatePreBlock())
	}
}

// TestDAO_GetDaoTemplate_LockCode 注入位置与无锁表零泄漏（渲染产物级断言）。
func TestDAO_GetDaoTemplate_LockCode(t *testing.T) {
	t.Run("有锁列表", func(t *testing.T) {
		code := GetDaoTemplate(lockTableData())
		if strings.Count(code, "when update,optimistic lock version is required") != 2 {
			t.Errorf("LockCheck 应注入两个更新方法, count=%d", strings.Count(code, "when update,optimistic lock version is required"))
		}
		if strings.Count(code, "!entity.VersionNo.Valid") != 2 {
			t.Errorf("锁列 Valid 判断应出现两次:\n%s", code)
		}
	})
	t.Run("无锁列表", func(t *testing.T) {
		td := lockTableData()
		for i := range td.Columns {
			td.Columns[i].IsOptimisticLock = false
		}
		td.Columns[1] = table.ColumnData{Name: "version_no", GoType: "int64", JsonTag: "VersionNo"}
		code := GetDaoTemplate(td)
		if strings.Contains(code, "optimisticlock") {
			t.Error("无锁列表产物不应出现任何乐观锁相关代码")
		}
	})
}
