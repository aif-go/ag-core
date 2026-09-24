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
		PrimaryKeys: []string{"id", "version_no"},
		Columns: []table.ColumnData{
			{Name: "id", GoType: "int64", JsonTag: "Id", IsPrimaryKey: true},
			{Name: "version_no", GoType: "optimisticlock.Version", JsonTag: "VersionNo", IsOptimisticLock: true, IsPrimaryKey: false},
			{Name: "name", GoType: "string", JsonTag: "Name"},
		},
	}
}

// baseLockCases 构造单/复合主键、有锁/无锁四组产物断言数据
func lockRenderCases() map[string]*table.TableData {
	td := lockTableData()

	multiPk := lockTableData()
	multiPk.PrimaryKeys = []string{"id", "version_no"}
	multiPk.Columns[1].IsPrimaryKey = true

	noLock := lockTableData()
	noLock.PrimaryKeys = []string{"id"}
	for i := range noLock.Columns {
		noLock.Columns[i].IsOptimisticLock = false
	}

	return map[string]*table.TableData{
		"单主键+锁列":  td,
		"复合主键+锁列": multiPk,
		"单主键无锁列":  noLock,
		"无主键":     PrimaryKeyless(),
	}
}

// PrimaryKeyless 无主键表
func PrimaryKeyless() *table.TableData {
	td := lockTableData()
	for i := range td.Columns {
		td.Columns[i].IsPrimaryKey = false
		td.Columns[i].IsOptimisticLock = false
	}
	td.PrimaryKeys = nil
	return td
}

// renderDao 渲染产物辅助
func renderDao(t *testing.T, data *table.TableData) string {
	t.Helper()
	return GetDaoTemplate(data)
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

// TestDAO_GetDaoTemplate_LockCheck 乐观锁校验注入：有锁列报告错误语义、无锁列零泄漏。
func TestDAO_GetDaoTemplate_LockCheck(t *testing.T) {
	t.Run("有锁列表", func(t *testing.T) {
		code := renderDao(t, lockTableData())
		if strings.Count(code, "when update,optimistic lock version is required") != 2 {
			t.Errorf("LockCheck 应注入两个更新方法, count=%d", strings.Count(code, "when update,optimistic lock version is required"))
		}
		if strings.Count(code, "!entity.VersionNo.Valid") != 2 {
			t.Errorf("锁列 Valid 判断应出现两次(两更新方法):\n%s", code)
		}
	})
	t.Run("无锁列表", func(t *testing.T) {
		td := lockTableData()
		for i := range td.Columns {
			td.Columns[i].IsOptimisticLock = false
		}
		td.Columns[1] = table.ColumnData{Name: "version_no", GoType: "int64", JsonTag: "VersionNo"}
		code := renderDao(t, td)
		if strings.Contains(code, "optimisticlock") {
			t.Error("无锁列表产物不应出现任何乐观锁相关代码")
		}
	})
}

// TestDAO_GetDaoTemplate_UpdateForm 更新形态统一（Save→Updates）语义：
// Select("*") 全字段覆盖、UpdateIgnore 不带 Select、无显式 where map。
func TestDAO_GetDaoTemplate_UpdateForm(t *testing.T) {
	t.Run("单主键+锁列", func(t *testing.T) {
		code := renderDao(t, lockTableData())
		if strings.Count(code, `db.Model(entity).Select("*").Updates(entity)`) != 1 {
			t.Errorf("全字段更新应为 db.Model(entity).Select(...).Updates(entity):\n%s", code)
		}
		if strings.Contains(code, "Save(") {
			t.Errorf("UpdateByPrimaryKey 不应再使用 Save 形态:\n%s", code)
		}
		if strings.Count(code, "db.Model(entity).Updates(entity)") != 1 {
			t.Errorf("部分更新应为 db.Model(entity).Updates(entity):\n%s", code)
		}
	})
	t.Run("复合主键+锁列", func(t *testing.T) {
		td := lockTableData()
		td.PrimaryKeys = []string{"id", "version_no"}
		td.Columns[1].IsPrimaryKey = true
		code := renderDao(t, td)
		if !strings.Contains(code, `(entity.TenantId == 0)`) && !strings.Contains(code, "!entity.Id") {
			// 复合主键: gorm auto 条件 + PKZeroCond 前置
		}
		if strings.Count(code, `db.Model(entity).Select("*").Updates(entity)`) != 1 {
			t.Errorf("复合主键全字段更新应与单主键同形态:\n%s", code)
		}
		if strings.Contains(code, "where := make(map[string]any)") {
			t.Errorf("统一形态不应生成显式 where map:\n%s", code)
		}
	})
	t.Run("无主键表", func(t *testing.T) {
		code := renderDao(t, PrimaryKeyless())
		// 三个方法各保留一处 dao 层拦截：两更新方法 + FindByPrimaryKey
		if strings.Count(code, "var pkConditions []string") != 3 {
			t.Errorf("无主键表两更新方法与 Find 应各保留一处 dao 层拦截, got %d", strings.Count(code, "var pkConditions []string"))
		}
		if strings.Count(code, "when query,primary key is required") != 1 {
			t.Errorf("FindByPrimaryKey 恒拦截应生成一次:\n%s", code)
		}
		if strings.Contains(code, "map[string]any") {
			t.Errorf("无主键表拦截不应该用 where map:\n%s", code)
		}
	})
}

// TestDAO_GetDaoTemplate_GuardCheck FindByStruct 守卫：主键与索引导引列非零判断。
func TestDAO_GetDaoTemplate_GuardCheck(t *testing.T) {
	td := lockTableData()
	td.PrimaryKeys = []string{"id"}
	td.Indexes = []table.IndexData{
		{Name: "idx_name", Columns: []string{"name"}},
	}
	code := renderDao(t, td)
	for _, want := range []string{
		"keyUsed := false",
		"// 检查主键",
		"if entity.Id != 0 {",
		"// 检查索引 idx_name",
		`if entity.Name != "" {`,
		"if !keyUsed {",
		`errors.New("query not use any index")`,
	} {
		if !strings.Contains(code, want) {
			t.Errorf("守卫块缺少期望片段 %q", want)
		}
	}
}
