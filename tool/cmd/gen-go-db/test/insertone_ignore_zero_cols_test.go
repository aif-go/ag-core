package test

import (
	"context"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
)

// TestStudentInsertOneIgnoreZeroValCols 覆盖 tm_student 的 InsertOneIgnoreZeroValCols：
// 剔除普通列（score/enroll_date/is_graduate）的零值，仅插入主键+索引列，由数据库写入默认值
func TestStudentInsertOneIgnoreZeroValCols(t *testing.T) {
	ctx := context.Background()
	studentDao := GetStudentRepository()

	// 测试表策略：清空整表保证数据干净，测试结束清空整表
	clearTable(t, "tm_student")
	defer clearTable(t, "tm_student")

	entity := &model.TmStudent{
		TenantId:  90002,
		StudentNo: "TEST002",
		Name:      "忽略零值学生",
		Address:   "忽略地址002",
		Phone:     "13800009002",
		ClassId:   "T9",
		CardNo:    "TSTCARD002",
		// Score / EnrollDate / IsGraduate 保持零值，应被剔除不进 INSERT
	}

	t.Run("场景1:剔除普通列零值-插入成功", func(t *testing.T) {
		affected, err := studentDao.InsertOneIgnoreZeroValCols(ctx, entity)
		if err != nil {
			t.Fatalf("InsertOneIgnoreZeroValCols 不期望错误: %v", err)
		}
		if affected != 1 {
			t.Errorf("期望影响 1 行，实际 %d", affected)
		}

		found, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: entity.TenantId, StudentNo: entity.StudentNo})
		if err != nil {
			t.Fatalf("回查失败: %v", err)
		}
		if found == nil {
			t.Fatal("回查未找到数据")
		}
		// 非零列应完整落库
		if found.Name != entity.Name || found.Address != entity.Address || found.CardNo != entity.CardNo {
			t.Errorf("非零列落库不一致: %+v", found)
		}
		// 被剔除的零值普通列应保持零值（数据库默认值）
		if !found.Score.IsZero() {
			t.Errorf("期望 score 因剔除而为零值, 实际: %v", found.Score)
		}
		t.Logf("插入成功, 回查结果: %+v", found)
	})

	t.Run("场景2:零值列不进INSERT-数据库默认生效", func(t *testing.T) {
		// 自包含：清空表后重新插入验证
		clearTable(t, "tm_student")
		if _, err := studentDao.InsertOneIgnoreZeroValCols(ctx, entity); err != nil {
			t.Fatalf("前置插入失败: %v", err)
		}
		found, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: entity.TenantId, StudentNo: entity.StudentNo})
		if err != nil || found == nil {
			t.Fatalf("回查失败: %v", err)
		}
		// bool 零值被剔除后，落库为 false（数据库默认）
		if found.IsGraduate {
			t.Errorf("期望 IsGraduate 保持 false, 实际: %v", found.IsGraduate)
		}
		t.Logf("零值列剔除后回查: %+v", found)
	})

	t.Run("场景3:nil实体-返回错误不崩溃", func(t *testing.T) {
		affected, err := studentDao.InsertOneIgnoreZeroValCols(ctx, nil)
		if err == nil {
			t.Errorf("期望 nil 实体报错, 实际 affected=%d err=nil", affected)
		} else {
			t.Logf("nil 实体正确返回错误: %v", err)
		}
	})
}

// TestTeacherInsertOneIgnoreZeroValCols 覆盖 tm_teacher 的 InsertOneIgnoreZeroValCols
func TestTeacherInsertOneIgnoreZeroValCols(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()

	// 测试表策略：清空整表保证数据干净，测试结束清空整表
	clearTable(t, "tm_teacher")
	defer clearTable(t, "tm_teacher")

	entity := &model.TmTeacher{
		Id:      90002,
		Name:    "忽略零值教师",
		Address: "忽略地址002",
		Phone:   "13900009002",
		ClassId: "T9",
		CardNo:  "TSTCARD002",
		// Salary / TimePointer / BoolTest 保持零值，应被剔除不进 INSERT
	}

	t.Run("场景1:剔除普通列零值-插入成功", func(t *testing.T) {
		affected, err := teacherDao.InsertOneIgnoreZeroValCols(ctx, entity)
		if err != nil {
			t.Fatalf("InsertOneIgnoreZeroValCols 不期望错误: %v", err)
		}
		if affected != 1 {
			t.Errorf("期望影响 1 行，实际 %d", affected)
		}

		found, err := teacherDao.FindByPrimaryKey(ctx, model.TmTeacherPrimaryKey(entity.Id))
		if err != nil {
			t.Fatalf("回查失败: %v", err)
		}
		if found == nil {
			t.Fatal("回查未找到数据")
		}
		if found.Name != entity.Name || found.ClassId != entity.ClassId {
			t.Errorf("非零列落库不一致: %+v", found)
		}
		if !found.Salary.IsZero() {
			t.Errorf("期望 salary 因剔除而为零值, 实际: %v", found.Salary)
		}
		t.Logf("插入成功, 回查结果: %+v", found)
	})

	t.Run("场景2:nil实体-返回错误不崩溃", func(t *testing.T) {
		affected, err := teacherDao.InsertOneIgnoreZeroValCols(ctx, nil)
		if err == nil {
			t.Errorf("期望 nil 实体报错, 实际 affected=%d err=nil", affected)
		} else {
			t.Logf("nil 实体正确返回错误: %v", err)
		}
	})
}
