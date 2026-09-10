package test

import (
	"context"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"github.com/shopspring/decimal"
)

// TestStudentUpdateByPrimaryKeyIngoreZeroValCols 覆盖 tm_student 的 UpdateByPrimaryKeyIngoreZeroValCols：
// 仅更新非零字段，零值字段不参与 SET，避免误清空其他列
func TestStudentUpdateByPrimaryKeyIngoreZeroValCols(t *testing.T) {
	ctx := context.Background()
	studentDao := GetStudentRepository()

	// 测试表策略：清空整表保证数据干净，测试结束清空整表
	clearTable(t, "tm_student")
	defer clearTable(t, "tm_student")

	seed := &model.TmStudent{
		TenantId: 90004, StudentNo: "TEST004", Name: "部分更新学生",
		Address: "原地址004", Phone: "13800009004", ClassId: "T9", CardNo: "ORGCARD004",
		Score: decimal.NewFromFloat(66.6),
	}
	if _, err := studentDao.InsertOne(ctx, seed); err != nil {
		t.Fatalf("前置插入失败: %v", err)
	}

	t.Run("场景1:仅更新指定非零字段-其他字段保持不变", func(t *testing.T) {
		before, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: seed.TenantId, StudentNo: seed.StudentNo})
		if err != nil || before == nil {
			t.Fatalf("加载原实体失败: %v", err)
		}

		// 仅设置主键 + 待更新的 support_update 列（class_id），其余保持零值
		patch := &model.TmStudent{TenantId: seed.TenantId, StudentNo: seed.StudentNo, ClassId: "C8"}
		affected, err := studentDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, patch)
		if err != nil {
			t.Fatalf("UpdateByPrimaryKeyIngoreZeroValCols 不期望错误: %v", err)
		}
		if affected != 1 {
			t.Errorf("期望影响 1 行，实际 %d", affected)
		}

		after, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: seed.TenantId, StudentNo: seed.StudentNo})
		if err != nil || after == nil {
			t.Fatalf("回查失败: %v", err)
		}
		if after.ClassId != "C8" {
			t.Errorf("class_id 未更新: %+v", after)
		}
		// 零值字段不应被改动
		if after.Name != before.Name {
			t.Errorf("name 不应被改动, got %q want %q", after.Name, before.Name)
		}
		if after.CardNo != before.CardNo {
			t.Errorf("card_no 不应被改动, got %q want %q", after.CardNo, before.CardNo)
		}
		if !after.Score.Equal(before.Score) {
			t.Errorf("score 不应被改动, got %v want %v", after.Score, before.Score)
		}
		t.Logf("部分更新成功, 回查结果: %+v", after)
	})

	t.Run("场景2:主键缺失-预期错误", func(t *testing.T) {
		_, err := studentDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, &model.TmStudent{Name: "x"})
		assertErrorContains(t, err, "primary key or unique key is required")
	})

	t.Run("场景3:更新不存在的主键-影响0行", func(t *testing.T) {
		affected, err := studentDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, &model.TmStudent{TenantId: 99999, StudentNo: "NOEXIST", ClassId: "C1"})
		if err != nil {
			t.Fatalf("更新不存在主键不期望错误: %v", err)
		}
		if affected != 0 {
			t.Errorf("期望影响 0 行，实际 %d", affected)
		}
	})
}

// TestTeacherUpdateByPrimaryKeyIngoreZeroValCols 覆盖 tm_teacher 的 UpdateByPrimaryKeyIngoreZeroValCols
func TestTeacherUpdateByPrimaryKeyIngoreZeroValCols(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()

	// 测试表策略：清空整表保证数据干净，测试结束清空整表
	clearTable(t, "tm_teacher")
	defer clearTable(t, "tm_teacher")

	seed := &model.TmTeacher{
		Id: 90004, Name: "部分更新教师", Address: "原地址004",
		Phone: "13900009004", ClassId: "T9", CardNo: "ORGCARD004",
		Salary: decimal.NewFromFloat(15000),
	}
	if _, err := teacherDao.InsertOne(ctx, seed); err != nil {
		t.Fatalf("前置插入失败: %v", err)
	}

	t.Run("场景1:仅更新指定非零字段-其他字段保持不变", func(t *testing.T) {
		before, err := teacherDao.FindByPrimaryKey(ctx, model.TmTeacherPrimaryKey(seed.Id))
		if err != nil || before == nil {
			t.Fatalf("加载原实体失败: %v", err)
		}

		patch := &model.TmTeacher{Id: seed.Id, CardNo: "UPDCARD004"}
		affected, err := teacherDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, patch)
		if err != nil {
			t.Fatalf("UpdateByPrimaryKeyIngoreZeroValCols 不期望错误: %v", err)
		}
		if affected != 1 {
			t.Errorf("期望影响 1 行，实际 %d", affected)
		}

		after, err := teacherDao.FindByPrimaryKey(ctx, model.TmTeacherPrimaryKey(seed.Id))
		if err != nil || after == nil {
			t.Fatalf("回查失败: %v", err)
		}
		if after.CardNo != "UPDCARD004" {
			t.Errorf("card_no 未更新: %+v", after)
		}
		if after.Name != before.Name {
			t.Errorf("name 不应被改动, got %q want %q", after.Name, before.Name)
		}
		if !after.Salary.Equal(before.Salary) {
			t.Errorf("salary 不应被改动, got %v want %v", after.Salary, before.Salary)
		}
		t.Logf("部分更新成功, 回查结果: %+v", after)
	})

	t.Run("场景2:主键缺失-预期错误", func(t *testing.T) {
		_, err := teacherDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, &model.TmTeacher{Name: "x"})
		assertErrorContains(t, err, "primary key or unique key is required")
	})

	t.Run("场景3:更新不存在的主键-影响0行", func(t *testing.T) {
		affected, err := teacherDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, &model.TmTeacher{Id: 999999999, CardNo: "X"})
		if err != nil {
			t.Fatalf("更新不存在主键不期望错误: %v", err)
		}
		if affected != 0 {
			t.Errorf("期望影响 0 行，实际 %d", affected)
		}
	})
}

// TestTmNoUpdateByPrimaryKeyIngoreZeroValCols 无主键无索引表 tm_no：Update 恒报错
func TestTmNoUpdateByPrimaryKeyIngoreZeroValCols(t *testing.T) {
	ctx := context.Background()
	tmNoDao := GetTmNoRepository()

	// 测试表策略：清空整表保证数据干净，测试结束清空整表
	clearTable(t, "tm_no")
	defer clearTable(t, "tm_no")

	t.Run("场景1:空实体更新-恒报错", func(t *testing.T) {
		_, err := tmNoDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, &model.TmNo{})
		assertErrorContains(t, err, "primary key or unique key is required")
	})

	t.Run("场景2:带字段更新-恒报错", func(t *testing.T) {
		_, err := tmNoDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, &model.TmNo{Name: "Alice"})
		assertErrorContains(t, err, "primary key or unique key is required")
	})
}
