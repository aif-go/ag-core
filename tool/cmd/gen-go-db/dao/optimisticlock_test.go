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

// TestDAO_UpdateStmt_UpdateForm 更新形态统一（Save→Updates、复合主键同用实体自动 WHERE）。
func TestDAO_UpdateStmt_UpdateForm(t *testing.T) {
	d := DaoTemplateData{TableData: lockTableData()}

	full := d.UpdateStmt()
	if !strings.Contains(full, `db.Model(entity).Select("*").Updates(entity)`) {
		t.Errorf("全字段更新应为 db.Model(entity).Select(...).Updates(entity):\n%s", full)
	}
	if strings.Contains(full, "Save(") {
		t.Errorf("UpdateByPrimaryKey 不应再使用 Save 形态:\n%s", full)
	}
	if strings.Contains(full, "Where(where)") {
		t.Errorf("统一形态不应携带显式 where map:\n%s", full)
	}

	ignore := d.UpdateIgnoreStmt()
	if strings.Contains(ignore, `Select("*")`) {
		t.Errorf("UpdateIgnore 不应带 Select(\"*\")（保留零值剔除语义）:\n%s", ignore)
	}
	if !strings.Contains(ignore, "db.Model(entity).Updates(entity)") {
		t.Errorf("部分更新形态不符:\n%s", ignore)
	}

	// 有主键表（单/复合）均不生成显式 where map
	for _, td := range []*table.TableData{lockTableData(), multiPkTableData()} {
		if strings.Contains((&DaoTemplateData{TableData: td}).UpdatePreBlock(), "where := make(map[string]any)") {
			t.Errorf("%s 有主键，不应生成显式 where map", td.StructName)
		}
	}

	// 无主键表：dao 层恒拦截（守卫语句，非 SQL 特例形态），无 unchanged where map
	noPk := lockTableData()
	noPk.PrimaryKeys = nil
	for i := range noPk.Columns {
		noPk.Columns[i].IsPrimaryKey = false
	}
	pre := (&DaoTemplateData{TableData: noPk}).UpdatePreBlock()
	if !strings.Contains(pre, "pkConditions") || !strings.Contains(pre, "primary key is required") {
		t.Errorf("无主键表应保留 dao 层拦截:\n%s", pre)
	}
	if strings.Contains(pre, "map[string]any") {
		t.Errorf("无主键表拦截不应使用 where map:\n%s", pre)
	}
}

// multiPkTableData 复合主键 + 乐观锁列
func multiPkTableData() *table.TableData {
	td := lockTableData()
	td.PrimaryKeys = []string{"id", "version_no"}
	td.Columns[1].IsPrimaryKey = true
	return td
}

// TestDAO_UpdateStmt_MultiPK 复合主键：全键零值校验 + 同统一执行语句（无 where map）。
func TestDAO_UpdateStmt_MultiPK(t *testing.T) {
	d := DaoTemplateData{TableData: multiPkTableData()}

	if !strings.Contains(d.UpdateStmt(), `db.Model(entity).Select("*").Updates(entity)`) {
		t.Errorf("复合主键全字段更新应与单主键同形态:\n%s", d.UpdateStmt())
	}
	if !strings.Contains(d.UpdateIgnoreStmt(), "db.Model(entity).Updates(entity)") {
		t.Errorf("复合主键部分更新应与单主键同形态:\n%s", d.UpdateIgnoreStmt())
	}
	pre := d.UpdatePreBlock()
	if !strings.Contains(pre, `entity.Id == 0`) || !strings.Contains(pre, `!entity.VersionNo.Valid`) {
		t.Errorf("复合主键应校验全部主键列:\n%s", pre)
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
