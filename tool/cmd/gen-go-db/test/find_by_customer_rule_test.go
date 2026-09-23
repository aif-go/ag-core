//go:build db

package test

import (
	"context"
	"testing"

	"github.com/aif-go/ag-core/contribute/agdb/conditonwhere"
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
)

// TestTeacherFindByCustomerRule_Validations 覆盖 FindByCustomerRule 的参数校验分支（纯代码校验，不依赖数据库）
func TestTeacherFindByCustomerRule_Validations(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()

	t.Run("场景1:ctx为nil-预期错误", func(t *testing.T) {
		_, err := teacherDao.FindByCustomerRule(nil, &gormdb.NameingSqlArgInfo{}, nil)
		assertErrorContains(t, err, "ctx is nil")
	})

	t.Run("场景2:namingInfo为nil-预期错误", func(t *testing.T) {
		_, err := teacherDao.FindByCustomerRule(ctx, nil, nil)
		assertErrorContains(t, err, "namingInfo is nil")
	})

	t.Run("场景3:SqlName为空-预期错误", func(t *testing.T) {
		_, err := teacherDao.FindByCustomerRule(ctx, &gormdb.NameingSqlArgInfo{}, nil)
		assertErrorContains(t, err, "namingInfo.SqlName is empty")
	})

	t.Run("场景4:请求参数类型不匹配-预期错误", func(t *testing.T) {
		namingInfo := &gormdb.NameingSqlArgInfo{
			SqlName: "FindByNameNadAddress",
			ReqType: &model.TmTeacherFindByNameNadAddressArg{},
		}
		// args 类型与 ReqType 不一致
		_, err := teacherDao.FindByCustomerRule(ctx, namingInfo, &model.TmTeacherFindByPhoneArg{FieldMask: conditonwhere.NewFieldMask()})
		assertErrorContains(t, err, "req type not match")
	})

	t.Run("场景5:未知SqlName-预期错误", func(t *testing.T) {
		namingInfo := &gormdb.NameingSqlArgInfo{
			SqlName: "NonExistentRule",
			ReqType: &model.TmTeacherFindByPhoneArg{},
		}
		_, err := teacherDao.FindByCustomerRule(ctx, namingInfo, &model.TmTeacherFindByPhoneArg{FieldMask: conditonwhere.NewFieldMask()})
		assertErrorContains(t, err, "not found naming sql")
	})
}

// seedCustomerRuleTeachers 清空 tm_teacher 并插入 FindByCustomerRule 业务用例自备数据
// （card_no/phone 为唯一约束列，必须独立取值），返回清理函数。
func seedCustomerRuleTeachers(t *testing.T, ctx context.Context, teacherDao interface {
	InsertOne(ctx context.Context, entity *model.TmTeacher) (int64, error)
}) func() {
	t.Helper()
	clearTable(t, "tm_teacher")
	seeds := []*model.TmTeacher{
		{Id: 94001, Name: "CustomerRuleAlice", Address: "北京市海淀区", Phone: "13800009401", ClassId: "C1", CardNo: "CRULE001"},
		{Id: 94002, Name: "CustomerRuleAlice", Address: "上海市浦东新区", Phone: "13800009402", ClassId: "C2", CardNo: "CRULE002"},
		{Id: 94003, Name: "CustomerRuleBob", Address: "北京市海淀区", Phone: "13800009403", ClassId: "C3", CardNo: "CRULE003"},
	}
	for _, s := range seeds {
		if _, err := teacherDao.InsertOne(ctx, s); err != nil {
			t.Fatalf("插入 tm_teacher 自备数据失败: %v", err)
		}
	}
	return func() { clearTable(t, "tm_teacher") }
}

// TestTeacherFindByCustomerRule_FindByNameNadAddress 覆盖自定义规则 FindByNameNadAddress（非分页）：
// MySQL 命名 SQL 已注册（InitTmTeacherNamingSql → InitTmTeacherMYSQL），强断言命中条数与结果集
func TestTeacherFindByCustomerRule_FindByNameNadAddress(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()
	cleanup := seedCustomerRuleTeachers(t, ctx, teacherDao)
	defer cleanup()

	fieldMask := conditonwhere.NewFieldMask()
	fieldMask.Set("Name")
	fieldMask.Set("Address")
	args := &model.TmTeacherFindByNameNadAddressArg{
		FieldMask: fieldMask,
		Name:      "CustomerRuleAlice",
		Address:   "北京市海淀区",
	}
	namingInfo := &gormdb.NameingSqlArgInfo{
		SqlName: "FindByNameNadAddress",
		ReqType: &model.TmTeacherFindByNameNadAddressArg{},
	}

	res, err := teacherDao.FindByCustomerRule(ctx, namingInfo, args)
	if err != nil {
		t.Fatalf("FindByCustomerRule 不期望错误: %v", err)
	}

	list, ok := res.([]*model.TmTeacherFindByNameNadAddressRes)
	if !ok {
		t.Fatalf("返回类型不符, got %T", res)
	}
	if len(list) != 1 {
		t.Fatalf("期望命中 1 条(Name=CustomerRuleAlice AND Address=北京市海淀区)，实际 %d 条: %v", len(list), list)
	}
	if list[0].Name != "CustomerRuleAlice" || list[0].Address != "北京市海淀区" || list[0].Phone != "13800009401" {
		t.Errorf("结果集字段不符: %+v", list[0])
	}
}

// TestTeacherFindByCustomerRule_FindByPhone 覆盖自定义规则 FindByPhone（分页）：
// 强断言分页 TotalCount/TotalPage/页内记录
func TestTeacherFindByCustomerRule_FindByPhone(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()
	cleanup := seedCustomerRuleTeachers(t, ctx, teacherDao)
	defer cleanup()

	fieldMask := conditonwhere.NewFieldMask()
	fieldMask.Set("Phone")
	args := &model.TmTeacherFindByPhoneArg{
		Page:      gormdb.Page{PageNum: 1, PageSize: 2},
		FieldMask: fieldMask,
		Phone:     "13800009401",
	}
	namingInfo := &gormdb.NameingSqlArgInfo{
		SqlName: "FindByPhone",
		ReqType: &model.TmTeacherFindByPhoneArg{},
	}

	res, err := teacherDao.FindByCustomerRule(ctx, namingInfo, args)
	if err != nil {
		t.Fatalf("FindByCustomerRule 不期望错误: %v", err)
	}

	pageRes, ok := res.(*model.TmTeacherFindByPhonePageRes)
	if !ok {
		t.Fatalf("返回类型不符, got %T", res)
	}
	if pageRes.PageResult.TotalCount != 1 {
		t.Errorf("期望 TotalCount=1，实际 %d", pageRes.PageResult.TotalCount)
	}
	if pageRes.PageResult.TotalPage != 1 {
		t.Errorf("期望 TotalPage=1，实际 %d", pageRes.PageResult.TotalPage)
	}
	if pageRes.PageResult.CurrentPage != 1 || pageRes.PageResult.PageSize != 2 {
		t.Errorf("分页参数不符: %+v", pageRes.PageResult)
	}
	if len(pageRes.ResultList) != 1 {
		t.Fatalf("期望本页 1 条，实际 %d 条", len(pageRes.ResultList))
	}
	if pageRes.ResultList[0].Id != 94001 || pageRes.ResultList[0].Phone != "13800009401" {
		t.Errorf("页内记录不符: %+v", pageRes.ResultList[0])
	}
}

// TestTeacherFindByCustomerRule_FieldMask 覆盖 FieldMask 未设置 / 仅非前导列场景（强断言错误路径）
func TestTeacherFindByCustomerRule_FieldMask(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()
	cleanup := seedCustomerRuleTeachers(t, ctx, teacherDao)
	defer cleanup()

	t.Run("场景1:FieldMask未设置任何字段-预期错误", func(t *testing.T) {
		args := &model.TmTeacherFindByNameNadAddressArg{FieldMask: conditonwhere.NewFieldMask()}
		namingInfo := &gormdb.NameingSqlArgInfo{
			SqlName: "FindByNameNadAddress",
			ReqType: &model.TmTeacherFindByNameNadAddressArg{},
		}
		_, err := teacherDao.FindByCustomerRule(ctx, namingInfo, args)
		assertErrorContains(t, err, "未设置方法FindByNameNadAddress对应的sql条件的参数值")
	})

	t.Run("场景2:仅设置非前导列Address-预期索引校验失败", func(t *testing.T) {
		fieldMask := conditonwhere.NewFieldMask()
		fieldMask.Set("Address")
		args := &model.TmTeacherFindByNameNadAddressArg{FieldMask: fieldMask, Address: "北京市海淀区"}
		namingInfo := &gormdb.NameingSqlArgInfo{
			SqlName: "FindByNameNadAddress",
			ReqType: &model.TmTeacherFindByNameNadAddressArg{},
		}
		_, err := teacherDao.FindByCustomerRule(ctx, namingInfo, args)
		assertErrorContains(t, err, "query not use any index")
	})

	t.Run("场景3:FieldMask设置Name(前导列)-索引校验通过", func(t *testing.T) {
		fieldMask := conditonwhere.NewFieldMask()
		fieldMask.Set("Name")
		args := &model.TmTeacherFindByNameNadAddressArg{FieldMask: fieldMask, Name: "CustomerRuleAlice"}
		namingInfo := &gormdb.NameingSqlArgInfo{
			SqlName: "FindByNameNadAddress",
			ReqType: &model.TmTeacherFindByNameNadAddressArg{},
		}
		// 仅设置前导列 Name → FieldMask 过滤后的 newwhere 含 name → 索引校验通过不报错；
		// 注意：实际执行的命名 SQL 固定为 (name = @Name AND address = @Address)，Address 未传为空串，
		// 因此结果集不保证命中，此处仅验证"索引校验通过、不报 query not use any index"。
		res, err := teacherDao.FindByCustomerRule(ctx, namingInfo, args)
		if err != nil {
			t.Fatalf("仅前导列 Name 不应报索引错误: %v", err)
		}
		if _, ok := res.([]*model.TmTeacherFindByNameNadAddressRes); !ok {
			t.Fatalf("返回类型不符, got %T", res)
		}
	})
}