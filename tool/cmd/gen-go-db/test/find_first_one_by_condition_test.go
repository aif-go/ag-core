package test

import (
	"context"
	"testing"

	"github.com/aif-go/ag-core/contribute/agdb/conditonwhere"
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"github.com/shopspring/decimal"
)

// TestStudentFindFirstOneByCondition 覆盖 tm_student 的 FindFirstOneByCondition
func TestStudentFindFirstOneByCondition(t *testing.T) {
	ctx := context.Background()
	studentDao := GetStudentRepository()

	// 测试表策略：清空整表 + 插入自备数据，测试结束清空整表
	cleanup := seedTmStudent(t, ctx, studentDao)
	defer cleanup()

	t.Run("场景1:命中条件-返回第一条", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "Alice")
		res, err := studentDao.FindFirstOneByCondition(ctx, condition, nil)
		if err != nil {
			t.Fatalf("FindFirstOneByCondition 不期望错误: %v", err)
		}
		if res == nil {
			t.Fatal("期望返回记录，但得到 nil")
		}
		if res.Name != "Alice" {
			t.Errorf("期望 Name=Alice, 实际: %s", res.Name)
		}
		t.Logf("命中: %+v", res)
	})

	t.Run("场景2:带排序返回首条", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "Alice")
		orderBuilder := gormdb.NewOrderBuilder().Desc("student_no")
		res, err := studentDao.FindFirstOneByCondition(ctx, condition, orderBuilder)
		if err != nil {
			t.Fatalf("FindFirstOneByCondition 不期望错误: %v", err)
		}
		if res == nil {
			t.Fatal("期望返回记录，但得到 nil")
		}
		t.Logf("按 student_no 降序取首条: %+v", res)
	})

	t.Run("场景3:无匹配条件-不报错", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "不存在的学生YYY")
		res, err := studentDao.FindFirstOneByCondition(ctx, condition, nil)
		if err != nil {
			t.Fatalf("无匹配时不应报错: %v", err)
		}
		if res == nil {
			t.Logf("当前实现无匹配时返回 nil")
			return
		}
		// 当前实现（db.Limit(1).Find）无匹配时返回零值实体而非 nil
		if res.TenantId != 0 || res.StudentNo != "" {
			t.Errorf("期望零值实体, 实际: %+v", res)
		}
		t.Logf("无匹配时返回零值实体: %+v", res)
	})
}

// TestTeacherFindFirstOneByCondition 覆盖 tm_teacher 的 FindFirstOneByCondition（自备数据）
func TestTeacherFindFirstOneByCondition(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()

	// 测试表策略：清空整表 + 插入自备数据，测试结束清空整表
	clearTable(t, "tm_teacher")
	defer clearTable(t, "tm_teacher")

	seed := &model.TmTeacher{
		Id: 90007, Name: "首条查询教师", Address: "查询地址007",
		Phone: "13900009007", ClassId: "T9", CardNo: "CARD007",
		Salary: decimal.NewFromFloat(500),
	}
	seed2 := &model.TmTeacher{
		Id: 91007, Name: "首条查询教师B", Address: "查询地址B007",
		Phone: "13900009107", ClassId: "TB", CardNo: "CARDB007",
		Salary: decimal.NewFromFloat(1500),
	}
	for _, s := range []*model.TmTeacher{seed, seed2} {
		if _, err := teacherDao.InsertOne(ctx, s); err != nil {
			t.Fatalf("前置插入失败: %v", err)
		}
	}

	t.Run("场景1:命中自备数据-返回首条", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "首条查询教师")
		res, err := teacherDao.FindFirstOneByCondition(ctx, condition, nil)
		if err != nil {
			t.Fatalf("FindFirstOneByCondition 不期望错误: %v", err)
		}
		if res == nil {
			t.Fatal("期望返回记录，但得到 nil")
		}
		if res.Id != seed.Id {
			t.Errorf("期望命中自备数据 Id=%d, 实际 %+v", seed.Id, res)
		}
		t.Logf("命中: %+v", res)
	})

	t.Run("场景2:无匹配条件-不报错", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "不存在的教师YYY")
		res, err := teacherDao.FindFirstOneByCondition(ctx, condition, nil)
		if err != nil {
			t.Fatalf("无匹配时不应报错: %v", err)
		}
		if res == nil {
			t.Logf("当前实现无匹配时返回 nil")
			return
		}
		if res.Id != 0 {
			t.Errorf("期望零值实体, 实际: %+v", res)
		}
		t.Logf("无匹配时返回零值实体: %+v", res)
	})

	t.Run("场景3:Between范围条件-命中首条", func(t *testing.T) {
		// 用唯一 Name 限定匹配集合，避免表内其他数据干扰；
		// name=seed2 AND salary between 1000 and 2000 命中 seed2(1500)
		condition := conditonwhere.NewWhereClauseBuilder().
			Eq("name", seed2.Name).And().Between("salary", decimal.NewFromInt(1000), decimal.NewFromInt(2000))
		res, err := teacherDao.FindFirstOneByCondition(ctx, condition, nil)
		if err != nil {
			t.Fatalf("FindFirstOneByCondition 不期望错误: %v", err)
		}
		if res == nil {
			t.Fatal("期望命中记录，但得到 nil")
		}
		if res.Id != seed2.Id {
			t.Errorf("Between 应命中 seed2(Id=%d), 实际: %+v", seed2.Id, res)
		}
		t.Logf("Between 命中: %+v", res)
	})
}
