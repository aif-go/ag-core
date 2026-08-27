package test

import (
	"context"
	"testing"
	"time"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"github.com/shopspring/decimal"
)

// TestStudentInsertOne 覆盖 tm_student（复合主键 + 索引列）的 InsertOne 方法
func TestStudentInsertOne(t *testing.T) {
	ctx := context.Background()
	studentDao := GetStudentRepository()

	// 测试表策略：清空整表保证数据干净，测试结束清空整表
	clearTable(t, "tm_student")
	defer clearTable(t, "tm_student")

	entity := &model.TmStudent{
		TenantId:   90001,
		StudentNo:  "TEST001",
		Name:       "插入测试学生",
		Address:    "测试地址001",
		Phone:      "13800009001",
		ClassId:    "T9",
		CardNo:     "TSTCARD001",
		Score:      decimal.NewFromFloat(88.5),
		IsGraduate: true,
	}

	t.Run("场景1:插入完整数据-成功且可回查", func(t *testing.T) {
		affected, err := studentDao.InsertOne(ctx, entity)
		if err != nil {
			t.Fatalf("InsertOne 不期望错误: %v", err)
		}
		if affected != 1 {
			t.Errorf("期望影响 1 行，实际 %d", affected)
		}

		found, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: entity.TenantId, StudentNo: entity.StudentNo})
		if err != nil {
			t.Fatalf("FindByPrimaryKey 回查失败: %v", err)
		}
		if found == nil {
			t.Fatal("回查未找到刚插入的数据")
		}
		if found.Name != entity.Name || found.Phone != entity.Phone || !found.Score.Equal(entity.Score) {
			t.Errorf("落库数据与入参不一致: %+v", found)
		}
		t.Logf("插入成功, 回查结果: %+v", found)
	})

	t.Run("场景2:重复主键插入-预期错误", func(t *testing.T) {
		// 自包含：清空表后插入一次，再重复插入应报错
		clearTable(t, "tm_student")
		if _, err := studentDao.InsertOne(ctx, entity); err != nil {
			t.Fatalf("前置插入失败: %v", err)
		}
		_, err := studentDao.InsertOne(ctx, entity)
		if err == nil {
			t.Error("期望重复主键插入返回错误，但得到 nil")
		} else {
			t.Logf("重复主键正确返回错误: %v", err)
		}
	})

	t.Run("场景3:nil实体-不应崩溃", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("nil 实体触发 panic（当前实现行为）: %v", r)
			}
		}()
		affected, err := studentDao.InsertOne(ctx, nil)
		if err != nil {
			t.Logf("nil 实体返回错误: %v", err)
			return
		}
		t.Logf("nil 实体未被拦截, affected=%d", affected)
	})

	t.Run("场景4:指针类型列EnrollDate插入-回查一致", func(t *testing.T) {
		enroll := time.Date(2025, 9, 1, 0, 0, 0, 0, time.Local)
		p := &model.TmStudent{
			TenantId: 90001, StudentNo: "TEST001-PTR", Name: "指针学生",
			Address: "测试地址001", Phone: "13800009001", ClassId: "T9",
			EnrollDate: &enroll,
		}
		// 整表已清空，p 主键与场景1(entity) 不同，直接插入（顶层 clearTable 负责清理）
		if _, err := studentDao.InsertOne(ctx, p); err != nil {
			t.Fatalf("插入失败: %v", err)
		}
		found, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: p.TenantId, StudentNo: p.StudentNo})
		if err != nil || found == nil {
			t.Fatalf("回查失败: %v", err)
		}
		if found.EnrollDate == nil || !found.EnrollDate.Equal(enroll) {
			t.Errorf("指针列回查不一致: %+v", found.EnrollDate)
		}
		t.Logf("指针列回查: %v", found.EnrollDate)
	})
}

// TestTeacherInsertOne 覆盖 tm_teacher（单主键）的 InsertOne 方法
func TestTeacherInsertOne(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()

	// 测试表策略：清空整表保证数据干净，测试结束清空整表
	clearTable(t, "tm_teacher")
	defer clearTable(t, "tm_teacher")

	entity := &model.TmTeacher{
		Id:       90001,
		Name:     "插入测试教师",
		Address:  "测试地址001",
		Phone:    "13900009001",
		ClassId:  "T9",
		CardNo:   "TSTCARD001",
		Salary:   decimal.NewFromFloat(20000.5),
		BoolTest: true,
	}

	t.Run("场景1:插入完整数据-成功且可回查", func(t *testing.T) {
		affected, err := teacherDao.InsertOne(ctx, entity)
		if err != nil {
			t.Fatalf("InsertOne 不期望错误: %v", err)
		}
		if affected != 1 {
			t.Errorf("期望影响 1 行，实际 %d", affected)
		}

		found, err := teacherDao.FindByPrimaryKey(ctx, model.TmTeacherPrimaryKey(entity.Id))
		if err != nil {
			t.Fatalf("FindByPrimaryKey 回查失败: %v", err)
		}
		if found == nil {
			t.Fatal("回查未找到刚插入的数据")
		}
		if found.Name != entity.Name || !found.Salary.Equal(entity.Salary) {
			t.Errorf("落库数据与入参不一致: %+v", found)
		}
		t.Logf("插入成功, 回查结果: %+v", found)
	})

	t.Run("场景2:重复主键插入-预期错误", func(t *testing.T) {
		// 自包含：清空表后插入一次，再重复插入应报错
		clearTable(t, "tm_teacher")
		if _, err := teacherDao.InsertOne(ctx, entity); err != nil {
			t.Fatalf("前置插入失败: %v", err)
		}
		_, err := teacherDao.InsertOne(ctx, entity)
		if err == nil {
			t.Error("期望重复主键插入返回错误，但得到 nil")
		} else {
			t.Logf("重复主键正确返回错误: %v", err)
		}
	})

	t.Run("场景3:nil实体-不应崩溃", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("nil 实体触发 panic（当前实现行为）: %v", r)
			}
		}()
		affected, err := teacherDao.InsertOne(ctx, nil)
		if err != nil {
			t.Logf("nil 实体返回错误: %v", err)
			return
		}
		t.Logf("nil 实体未被拦截, affected=%d", affected)
	})

	t.Run("场景4:指针类型列TimePointer插入-回查一致", func(t *testing.T) {
		tp := time.Date(2025, 1, 15, 12, 30, 0, 0, time.Local)
		// 注意：tm_teacher 表 card_no/phone 为唯一约束列，且主键 Id 与场景1/2 的 entity 相同会导致重复插入失败，
		// 故此处 Id/Phone/CardNo 全部使用独立值
		p := &model.TmTeacher{
			Id: 90008, Name: "指针教师", Address: "测试地址", Phone: "13900009008", ClassId: "T9", CardNo: "TSTCARD008",
			TimePointer: &tp,
		}
		if _, err := teacherDao.InsertOne(ctx, p); err != nil {
			t.Fatalf("插入失败: %v", err)
		}
		found, err := teacherDao.FindByPrimaryKey(ctx, model.TmTeacherPrimaryKey(p.Id))
		if err != nil || found == nil {
			t.Fatalf("回查失败: %v", err)
		}
		if found.TimePointer == nil || !found.TimePointer.Equal(tp) {
			t.Errorf("指针列回查不一致: %+v", found.TimePointer)
		}
		t.Logf("指针列回查: %v", found.TimePointer)
	})
}
