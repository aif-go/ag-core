package test

import (
	"context"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"github.com/shopspring/decimal"
)

// TestStudentUpdateByPrimaryKey 覆盖 tm_student（复合主键）的 UpdateByPrimaryKey
func TestStudentUpdateByPrimaryKey(t *testing.T) {
	ctx := context.Background()
	studentDao := GetStudentRepository()

	// 测试表策略：清空整表保证数据干净，测试结束清空整表
	clearTable(t, "tm_student")
	defer clearTable(t, "tm_student")

	// 前置准备：插入一条待更新数据
	seed := &model.TmStudent{
		TenantId: 90003, StudentNo: "TEST003", Name: "待更新学生",
		Address: "原地址003", Phone: "13800009003", ClassId: "T9", CardNo: "ORGCARD003",
	}
	if _, err := studentDao.InsertOne(ctx, seed); err != nil {
		t.Fatalf("前置插入失败: %v", err)
	}

	t.Run("场景1:更新支持列-成功", func(t *testing.T) {
		loaded, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: seed.TenantId, StudentNo: seed.StudentNo})
		if err != nil || loaded == nil {
			t.Fatalf("加载原实体失败: %v", err)
		}
		loaded.ClassId = "C9"
		loaded.CardNo = "UPDCARD003"

		affected, err := studentDao.UpdateByPrimaryKey(ctx, loaded)
		if err != nil {
			t.Fatalf("UpdateByPrimaryKey 不期望错误: %v", err)
		}
		if affected != 1 {
			t.Errorf("期望影响 1 行，实际 %d", affected)
		}

		found, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: seed.TenantId, StudentNo: seed.StudentNo})
		if err != nil || found == nil {
			t.Fatalf("回查失败: %v", err)
		}
		if found.ClassId != "C9" || found.CardNo != "UPDCARD003" {
			t.Errorf("更新列未生效: %+v", found)
		}
		t.Logf("更新成功, 回查结果: %+v", found)
	})

	t.Run("场景2:主键缺失-TenantId=0-预期错误", func(t *testing.T) {
		_, err := studentDao.UpdateByPrimaryKey(ctx, &model.TmStudent{StudentNo: seed.StudentNo, Name: "x", ClassId: "C1"})
		assertErrorContains(t, err, "primary key or unique key is required")
	})

	t.Run("场景3:主键缺失-StudentNo为空-预期错误", func(t *testing.T) {
		_, err := studentDao.UpdateByPrimaryKey(ctx, &model.TmStudent{TenantId: seed.TenantId, Name: "x", ClassId: "C1"})
		assertErrorContains(t, err, "primary key or unique key is required")
	})

	t.Run("场景4:更新不存在的主键-Save upsert插入-影响1行", func(t *testing.T) {
		t.Log("注意：GORM Save 语义：UPDATE 影响 0 行时回退 INSERT（upsert），故不存在的主键会被插入，影响 1 行")
		notExist := &model.TmStudent{
			TenantId: 99999, StudentNo: "NOEXIST999", Name: "不存在",
			Address: "a", Phone: "13800000000", ClassId: "C1",
		}
		affected, err := studentDao.UpdateByPrimaryKey(ctx, notExist)
		if err != nil {
			t.Fatalf("更新不存在主键不期望错误: %v", err)
		}
		// GORM Save 语义：UPDATE 影响 0 行时回退 INSERT（upsert），故不存在的主键会被插入，影响 1 行
		if affected != 1 {
			t.Errorf("期望影响 1 行（Save upsert 插入），实际 %d", affected)
		}
	})

	t.Run("场景5:Save全字段更新-零值被写入", func(t *testing.T) {
		// UpdateByPrimaryKey 使用 Save 会全字段更新，零值也会写入（区别于 IngoreZeroValCols）
		loaded, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: seed.TenantId, StudentNo: seed.StudentNo})
		if err != nil || loaded == nil {
			t.Fatalf("加载原实体失败: %v", err)
		}
		loaded.Score = decimal.NewFromInt(0)
		loaded.ClassId = "C7"
		affected, err := studentDao.UpdateByPrimaryKey(ctx, loaded)
		if err != nil {
			t.Fatalf("UpdateByPrimaryKey 不期望错误: %v", err)
		}
		if affected != 1 {
			t.Errorf("期望影响 1 行，实际 %d", affected)
		}
		found, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: seed.TenantId, StudentNo: seed.StudentNo})
		if err != nil || found == nil {
			t.Fatalf("回查失败: %v", err)
		}
		if !found.Score.IsZero() {
			t.Errorf("Save 应全字段更新, score 应为零值: %+v", found)
		}
		if found.ClassId != "C7" {
			t.Errorf("class_id 未更新: %+v", found)
		}
		t.Logf("全字段更新回查: %+v", found)
	})
}

// TestTeacherUpdateByPrimaryKey 覆盖 tm_teacher（单主键）的 UpdateByPrimaryKey
func TestTeacherUpdateByPrimaryKey(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()

	// 测试表策略：清空整表保证数据干净，测试结束清空整表
	clearTable(t, "tm_teacher")
	defer clearTable(t, "tm_teacher")

	seed := &model.TmTeacher{
		Id: 90003, Name: "待更新教师", Address: "原地址003",
		Phone: "13900009003", ClassId: "T9", CardNo: "ORGCARD003",
	}
	if _, err := teacherDao.InsertOne(ctx, seed); err != nil {
		t.Fatalf("前置插入失败: %v", err)
	}

	t.Run("场景1:更新支持列-成功", func(t *testing.T) {
		loaded, err := teacherDao.FindByPrimaryKey(ctx, model.TmTeacherPrimaryKey(seed.Id))
		if err != nil || loaded == nil {
			t.Fatalf("加载原实体失败: %v", err)
		}
		loaded.ClassId = "C9"
		loaded.CardNo = "UPDCARD003"

		affected, err := teacherDao.UpdateByPrimaryKey(ctx, loaded)
		if err != nil {
			t.Fatalf("UpdateByPrimaryKey 不期望错误: %v", err)
		}
		if affected != 1 {
			t.Errorf("期望影响 1 行，实际 %d", affected)
		}

		found, err := teacherDao.FindByPrimaryKey(ctx, model.TmTeacherPrimaryKey(seed.Id))
		if err != nil || found == nil {
			t.Fatalf("回查失败: %v", err)
		}
		if found.ClassId != "C9" || found.CardNo != "UPDCARD003" {
			t.Errorf("更新列未生效: %+v", found)
		}
		t.Logf("更新成功, 回查结果: %+v", found)
	})

	t.Run("场景2:主键缺失-Id=0-预期错误", func(t *testing.T) {
		_, err := teacherDao.UpdateByPrimaryKey(ctx, &model.TmTeacher{Name: "x", ClassId: "C1"})
		assertErrorContains(t, err, "primary key or unique key is required")
	})

	t.Run("场景3:更新不存在的主键-Save upsert插入-影响1行", func(t *testing.T) {
		t.Log("注意：GORM Save 语义：UPDATE 影响 0 行时回退 INSERT（upsert），故不存在的主键会被插入，影响 1 行")
		notExist := &model.TmTeacher{Id: 999999999, Name: "不存在", Address: "a", Phone: "13900000000", ClassId: "C1"}
		affected, err := teacherDao.UpdateByPrimaryKey(ctx, notExist)
		if err != nil {
			t.Fatalf("更新不存在主键不期望错误: %v", err)
		}
		// GORM Save 语义：UPDATE 影响 0 行时回退 INSERT（upsert），故不存在的主键会被插入，影响 1 行
		if affected != 1 {
			t.Errorf("期望影响 1 行（Save upsert 插入），实际 %d", affected)
		}
	})
}

// TestTmNoUpdateByPrimaryKey 无主键无索引表 tm_no：生成代码守卫恒报错，Update 一律拦截
func TestTmNoUpdateByPrimaryKey(t *testing.T) {
	ctx := context.Background()
	tmNoDao := GetTmNoRepository()

	// 测试表策略：清空整表保证数据干净，测试结束清空整表
	clearTable(t, "tm_no")
	defer clearTable(t, "tm_no")

	t.Run("场景1:空实体更新-恒报错", func(t *testing.T) {
		_, err := tmNoDao.UpdateByPrimaryKey(ctx, &model.TmNo{})
		assertErrorContains(t, err, "primary key or unique key is required")
	})

	t.Run("场景2:带字段更新-恒报错", func(t *testing.T) {
		_, err := tmNoDao.UpdateByPrimaryKey(ctx, &model.TmNo{Name: "Alice", Score: decimal.NewFromInt(95)})
		assertErrorContains(t, err, "primary key or unique key is required")
	})
}
