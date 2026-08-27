package test

import (
	"context"
	"testing"

	"github.com/aif-go/ag-core/contribute/agdb/conditonwhere"
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"github.com/shopspring/decimal"
)

// TestStudentFindByCondition 覆盖 tm_student 的 FindByCondition（条件/分页/排序/异常场景）
func TestStudentFindByCondition(t *testing.T) {
	ctx := context.Background()
	studentDao := GetStudentRepository()

	// 测试表策略：清空整表 + 插入自备数据，测试结束清空整表（TotalCount 等精确断言可靠）
	cleanup := seedTmStudent(t, ctx, studentDao)
	defer cleanup()

	t.Run("场景1:基础等值条件查询-未分页", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "Alice")
		list, pageResult, err := studentDao.FindByCondition(ctx, condition, nil, nil)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if len(list) != 3 {
			t.Errorf("期望 3 条(name=Alice)，实际 %d 条: %v", len(list), list)
		}
		if pageResult != nil {
			t.Errorf("未分页时 pageResult 应为 nil, 实际: %+v", pageResult)
		}
	})

	t.Run("场景2:分页查询-返回正确分页信息", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "Alice")
		page := &gormdb.Page{PageNum: 1, PageSize: 2}
		list, pageResult, err := studentDao.FindByCondition(ctx, condition, nil, page)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if len(list) > 2 {
			t.Errorf("分页应最多返回 2 条, 实际 %d", len(list))
		}
		if pageResult == nil {
			t.Fatal("期望返回分页信息，但得到 nil")
		}
		if pageResult.TotalCount != 3 {
			t.Errorf("期望总数 3, 实际 %d", pageResult.TotalCount)
		}
		if pageResult.TotalPage != 2 {
			t.Errorf("期望总页数 2, 实际 %d", pageResult.TotalPage)
		}
		if pageResult.CurrentPage != 1 || pageResult.PageSize != 2 {
			t.Errorf("分页参数不符: %+v", pageResult)
		}
	})

	t.Run("场景3:IN条件查询", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().In("phone", "13800000000")
		list, _, err := studentDao.FindByCondition(ctx, condition, nil, nil)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("期望 2 条(phone=13800000000)，实际 %d 条", len(list))
		}
	})

	t.Run("场景4:排序查询-按student_no升序", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "Alice")
		orderBuilder := gormdb.NewOrderBuilder().Asc("student_no")
		list, _, err := studentDao.FindByCondition(ctx, condition, orderBuilder, nil)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if len(list) != 3 {
			t.Errorf("期望 3 条, 实际 %d 条", len(list))
		}
		for i := 1; i < len(list); i++ {
			if list[i-1].StudentNo > list[i].StudentNo {
				t.Errorf("排序不正确(应升序): %v", list)
			}
		}
	})

	t.Run("场景5:非法页码-PageNum=0-预期错误", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "Alice")
		page := &gormdb.Page{PageNum: 0, PageSize: 2}
		_, _, err := studentDao.FindByCondition(ctx, condition, nil, page)
		assertErrorContains(t, err, "pageNum must be ≥1")
	})

	t.Run("场景6:超过最后一页-返回空结果与分页信息", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "Alice")
		page := &gormdb.Page{PageNum: 999, PageSize: 2}
		list, pageResult, err := studentDao.FindByCondition(ctx, condition, nil, page)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("期望空列表, 实际 %d 条", len(list))
		}
		if pageResult == nil || pageResult.TotalCount != 3 {
			t.Errorf("期望总数 3, 实际: %+v", pageResult)
		}
	})

	t.Run("场景7:OR条件查询", func(t *testing.T) {
		// name=Alice(3条) OR name=Helen(1条)
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "Alice").Or().Eq("name", "Helen")
		list, _, err := studentDao.FindByCondition(ctx, condition, nil, nil)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if len(list) != 4 {
			t.Errorf("期望 4 条(Alice 3 + Helen 1)，实际 %d 条", len(list))
		}
	})

	t.Run("场景8:无匹配条件-返回空列表", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "不存在的学生XXX")
		list, _, err := studentDao.FindByCondition(ctx, condition, nil, nil)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("期望 0 条, 实际 %d 条", len(list))
		}
	})

	t.Run("场景9:非法页大小-PageSize=0-预期错误", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "Alice")
		page := &gormdb.Page{PageNum: 1, PageSize: 0}
		_, _, err := studentDao.FindByCondition(ctx, condition, nil, page)
		assertErrorContains(t, err, "pageSize must be ≥1")
	})

	t.Run("场景10:Group嵌套条件-括号优先级", func(t *testing.T) {
		// (name=Alice OR name=Helen) AND class_id=C01
		// 自备数据：Alice 中 NO001/NO010 的 class_id=C01，Helen(NO013) 的 class_id=C04 → 精确命中 2 条
		// 注意：必须用 Group + ConditionOrGroup 嵌套构造，才能得到 ((A OR B) AND C) 语义；
		// 若用 BeginGroup()/EndGroup()，root 为空时 group 直接成为 root，后续条件被并入组内，
		// SQL 变为 (name=? OR name=? AND class_id=?)，AND 优先级高于 OR，实际等价于 name=Alice(3 条)。
		condition := conditonwhere.NewWhereClauseBuilder().Group(
			conditonwhere.ConditionOrGroup(
				conditonwhere.ConditionEq("name", "Alice"),
				conditonwhere.ConditionEq("name", "Helen"),
			),
			conditonwhere.ConditionEq("class_id", "C01"),
		)
		list, _, err := studentDao.FindByCondition(ctx, condition, nil, nil)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("Group 嵌套期望命中 2 条, 实际 %d 条: %v", len(list), list)
		}
		t.Logf("Group 嵌套命中 %d 条", len(list))
	})
}

// TestTeacherFindByCondition 覆盖 tm_teacher 的 FindByCondition（自备数据，保证确定性）
func TestTeacherFindByCondition(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()

	// 测试表策略：清空整表 + 插入自备数据，测试结束清空整表
	clearTable(t, "tm_teacher")
	defer clearTable(t, "tm_teacher")

	seed := &model.TmTeacher{
		Id: 90006, Name: "条件查询教师", Address: "查询地址006",
		Phone: "13900009006", ClassId: "T9", CardNo: "CARD006",
		Salary: decimal.NewFromFloat(1000),
	}
	seed2 := &model.TmTeacher{
		Id: 91006, Name: "条件查询教师B", Address: "查询地址B006",
		Phone: "13900009106", ClassId: "TB", CardNo: "CARDB006",
		Salary: decimal.NewFromFloat(2000),
	}
	for _, s := range []*model.TmTeacher{seed, seed2} {
		if _, err := teacherDao.InsertOne(ctx, s); err != nil {
			t.Fatalf("前置插入失败: %v", err)
		}
	}

	t.Run("场景1:按Name精确查询-命中自备数据", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "条件查询教师")
		list, _, err := teacherDao.FindByCondition(ctx, condition, nil, nil)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("期望 1 条, 实际 %d 条", len(list))
		}
		if list[0].Id != seed.Id {
			t.Errorf("期望命中自备数据 Id=%d, 实际 %+v", seed.Id, list[0])
		}
	})

	t.Run("场景2:分页查询-总数与页数正确", func(t *testing.T) {
		condition := conditonwhere.NewWhereClauseBuilder().Eq("name", "条件查询教师")
		page := &gormdb.Page{PageNum: 1, PageSize: 1}
		list, pageResult, err := teacherDao.FindByCondition(ctx, condition, nil, page)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("期望 1 条, 实际 %d 条", len(list))
		}
		if pageResult == nil || pageResult.TotalCount != 1 || pageResult.TotalPage != 1 {
			t.Errorf("分页信息不符: %+v", pageResult)
		}
	})

	t.Run("场景3:Between范围条件-命中区间内数据", func(t *testing.T) {
		// 用唯一 Name 限定匹配集合，避免表内其他数据干扰；
		// name=seed2 AND salary between 1500 and 2500 命中 seed2(2000)
		condition := conditonwhere.NewWhereClauseBuilder().
			Eq("name", seed2.Name).And().Between("salary", decimal.NewFromInt(1500), decimal.NewFromInt(2500))
		list, _, err := teacherDao.FindByCondition(ctx, condition, nil, nil)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if len(list) != 1 || list[0].Id != seed2.Id {
			t.Errorf("Between 应命中 seed2(Id=%d), 实际 %d 条: %v", seed2.Id, len(list), list)
		}
	})

	t.Run("场景4:Gt+排序+分页组合条件", func(t *testing.T) {
		// (name=seed OR name=seed2) AND salary>500，匹配集合=自备2条，对表内其他数据免疫
		// 注意：必须用 Group + ConditionOrGroup 嵌套得到 ((A OR B) AND C) 语义，
		// 若用 BeginGroup()/EndGroup() 会生成 (A OR B AND C)，AND 优先级更高导致语义偏差（此处数据恰好都满足 salary>500 才未暴露）
		condition := conditonwhere.NewWhereClauseBuilder().Group(
			conditonwhere.ConditionOrGroup(
				conditonwhere.ConditionEq("name", seed.Name),
				conditonwhere.ConditionEq("name", seed2.Name),
			),
			conditonwhere.ConditionGt("salary", decimal.NewFromInt(500)),
		)
		page := &gormdb.Page{PageNum: 1, PageSize: 1}
		orderBuilder := gormdb.NewOrderBuilder().Desc("salary")
		list, pageResult, err := teacherDao.FindByCondition(ctx, condition, orderBuilder, page)
		if err != nil {
			t.Fatalf("FindByCondition 不期望错误: %v", err)
		}
		if pageResult == nil || pageResult.TotalCount != 2 {
			t.Errorf("期望总数 2, 实际: %+v", pageResult)
		}
		if len(list) != 1 || list[0].Id != seed2.Id {
			t.Errorf("降序第1条应为 seed2(Id=%d), 实际 %d 条: %v", seed2.Id, len(list), list)
		}
	})
}
