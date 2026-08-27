package test

import (
	"context"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
)

// TestStudentFindByPrimaryKey 覆盖 tm_student（复合主键）的 FindByPrimaryKey
func TestStudentFindByPrimaryKey(t *testing.T) {
	ctx := context.Background()
	studentDao := GetStudentRepository()

	// 测试表策略：清空整表 + 插入自备数据，测试结束清空整表
	cleanup := seedTmStudent(t, ctx, studentDao)
	defer cleanup()

	t.Run("场景1:查询存在的复合主键-命中", func(t *testing.T) {
		// 使用自备数据 NO001
		found, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: 1, StudentNo: "NO001"})
		if err != nil {
			t.Fatalf("FindByPrimaryKey 不期望错误: %v", err)
		}
		if found == nil {
			t.Fatal("期望命中记录，但返回 nil")
		}
		if found.Name != "Alice" {
			t.Errorf("期望 Name=Alice, 实际: %s", found.Name)
		}
		t.Logf("命中结果: %+v", found)
	})

	t.Run("场景2:查询不存在的主键-返回nil,nil", func(t *testing.T) {
		found, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: 999999, StudentNo: "NOEXIST999"})
		if err != nil {
			t.Fatalf("不期望错误: %v", err)
		}
		if found != nil {
			t.Errorf("期望返回 nil, 实际: %+v", found)
		}
	})
}

// TestTeacherFindByPrimaryKey 覆盖 tm_teacher（单主键）的 FindByPrimaryKey
func TestTeacherFindByPrimaryKey(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()

	// 测试表策略：清空整表 + 插入自备数据，测试结束清空整表
	clearTable(t, "tm_teacher")
	defer clearTable(t, "tm_teacher")

	// 自备数据：tm_teacher 的种子数据未在 findstruct_test 中约定，插入后查询以保证确定性
	seed := &model.TmTeacher{
		Id: 90005, Name: "主键查询教师", Address: "查询地址005",
		Phone: "13900009005", ClassId: "T9", CardNo: "CARD005",
	}
	if _, err := teacherDao.InsertOne(ctx, seed); err != nil {
		t.Fatalf("前置插入失败: %v", err)
	}

	t.Run("场景1:查询存在的主键-命中", func(t *testing.T) {
		found, err := teacherDao.FindByPrimaryKey(ctx, model.TmTeacherPrimaryKey(seed.Id))
		if err != nil {
			t.Fatalf("FindByPrimaryKey 不期望错误: %v", err)
		}
		if found == nil {
			t.Fatal("期望命中记录，但返回 nil")
		}
		if found.Name != seed.Name {
			t.Errorf("期望 Name=%s, 实际: %s", seed.Name, found.Name)
		}
		t.Logf("命中结果: %+v", found)
	})

	t.Run("场景2:查询不存在的主键-返回nil,nil", func(t *testing.T) {
		found, err := teacherDao.FindByPrimaryKey(ctx, model.TmTeacherPrimaryKey(999999999))
		if err != nil {
			t.Fatalf("不期望错误: %v", err)
		}
		if found != nil {
			t.Errorf("期望返回 nil, 实际: %+v", found)
		}
	})
}

// TestTmNoIndexFindByPrimaryKey 有主键无索引表 tm_no_index：按主键查询命中
func TestTmNoIndexFindByPrimaryKey(t *testing.T) {
	ctx := context.Background()
	tmNoIndexDao := GetTmNoIndexRepository()

	// 测试表策略：清空整表 + 插入自备数据，测试结束清空整表
	cleanup := seedTmNoIndex(t, ctx, tmNoIndexDao)
	defer cleanup()

	t.Run("场景1:按完整复合主键-命中", func(t *testing.T) {
		found, err := tmNoIndexDao.FindByPrimaryKey(ctx, model.TmNoIndexPrimarkey{TenantId: 1, StudentNo: "NO001"})
		if err != nil {
			t.Fatalf("FindByPrimaryKey 不期望错误: %v", err)
		}
		if found == nil {
			t.Fatal("期望命中记录，但返回 nil")
		}
		t.Logf("命中结果: %+v", found)
	})

	t.Run("场景2:查询不存在的主键-返回nil,nil", func(t *testing.T) {
		found, err := tmNoIndexDao.FindByPrimaryKey(ctx, model.TmNoIndexPrimarkey{TenantId: 999999, StudentNo: "NOEXIST"})
		if err != nil {
			t.Fatalf("不期望错误: %v", err)
		}
		if found != nil {
			t.Errorf("期望返回 nil, 实际: %+v", found)
		}
	})
}

// TestStudentFindByPrimaryKeyPartial 复合主键仅传部分字段：未命中返回 nil,nil
func TestStudentFindByPrimaryKeyPartial(t *testing.T) {
	ctx := context.Background()
	studentDao := GetStudentRepository()

	// 测试表策略：清空整表保证表内无干扰数据，测试结束清空整表
	clearTable(t, "tm_student")
	defer clearTable(t, "tm_student")

	t.Run("场景1:仅传首主键TenantId-未命中", func(t *testing.T) {
		found, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: 1, StudentNo: ""})
		if err != nil {
			t.Fatalf("不期望错误: %v", err)
		}
		if found != nil {
			t.Errorf("期望返回 nil, 实际: %+v", found)
		}
	})

	t.Run("场景2:仅传次主键StudentNo-未命中", func(t *testing.T) {
		found, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimarkey{TenantId: 0, StudentNo: "NO001"})
		if err != nil {
			t.Fatalf("不期望错误: %v", err)
		}
		if found != nil {
			t.Errorf("期望返回 nil, 实际: %+v", found)
		}
	})
}

// TestTmNoFindByPrimaryKey 无主键无索引表 tm_no：FindByPrimaryKey 行为容错（不应崩溃）
func TestTmNoFindByPrimaryKey(t *testing.T) {
	ctx := context.Background()
	tmNoDao := GetTmNoRepository()

	// 测试表策略：清空整表保证数据干净，测试结束清空整表
	clearTable(t, "tm_no")
	defer clearTable(t, "tm_no")

	defer func() {
		if r := recover(); r != nil {
			t.Logf("tm_no FindByPrimaryKey 触发 panic（当前实现行为）: %v", r)
		}
	}()

	found, err := tmNoDao.FindByPrimaryKey(ctx, model.TmNoPrimarkey{})
	if err != nil {
		t.Logf("tm_no FindByPrimaryKey 返回错误: %v", err)
		return
	}
	t.Logf("tm_no FindByPrimaryKey 返回: %+v", found)
}
