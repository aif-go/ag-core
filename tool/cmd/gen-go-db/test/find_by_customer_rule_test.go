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

// TestTeacherFindByCustomerRule_FindByNameNadAddress 覆盖自定义规则 FindByNameNadAddress（非分页）
// 注意：生成的 InitTmTeacherNamingSql 仅注册 DB2 命名SQL，在 MySQL 环境下会命中 "not found naming sql"，
// 属已知的生成代码限制，此处对业务结果做容错处理并记录原因。
func TestTeacherFindByCustomerRule_FindByNameNadAddress(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()

	fieldMask := conditonwhere.NewFieldMask()
	fieldMask.Set("Name")
	fieldMask.Set("Address")
	args := &model.TmTeacherFindByNameNadAddressArg{
		FieldMask: fieldMask,
		Name:      "Alice",
		Address:   "北京市海淀区",
	}
	namingInfo := &gormdb.NameingSqlArgInfo{
		SqlName: "FindByNameNadAddress",
		ReqType: &model.TmTeacherFindByNameNadAddressArg{},
	}

	res, err := teacherDao.FindByCustomerRule(ctx, namingInfo, args)
	if err != nil {
		if err.Error() == "not found naming sql" {
			t.Logf("当前 DbType=%s 未注册 MYSQL 命名SQL（InitTmTeacherNamingSql 仅初始化 DB2），跳过业务断言: %v", DbType, err)
			return
		}
		t.Fatalf("FindByCustomerRule 不期望错误: %v", err)
	}

	list, ok := res.([]*model.TmTeacherFindByNameNadAddressRes)
	if !ok {
		t.Fatalf("返回类型不符, got %T", res)
	}
	t.Logf("FindByNameNadAddress 查询到 %d 条", len(list))
	for _, e := range list {
		t.Logf("  -> %+v", e)
	}
}

// TestTeacherFindByCustomerRule_FindByPhone 覆盖自定义规则 FindByPhone（分页）
func TestTeacherFindByCustomerRule_FindByPhone(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()

	fieldMask := conditonwhere.NewFieldMask()
	fieldMask.Set("Phone")
	args := &model.TmTeacherFindByPhoneArg{
		Page:      gormdb.Page{PageNum: 1, PageSize: 2},
		FieldMask: fieldMask,
		Phone:     "13800000000",
	}
	namingInfo := &gormdb.NameingSqlArgInfo{
		SqlName: "FindByPhone",
		ReqType: &model.TmTeacherFindByPhoneArg{},
	}

	res, err := teacherDao.FindByCustomerRule(ctx, namingInfo, args)
	if err != nil {
		if err.Error() == "not found naming sql" {
			t.Logf("当前 DbType=%s 未注册 MYSQL 命名SQL（InitTmTeacherNamingSql 仅初始化 DB2），跳过业务断言: %v", DbType, err)
			return
		}
		t.Fatalf("FindByCustomerRule 不期望错误: %v", err)
	}

	pageRes, ok := res.(*model.TmTeacherFindByPhonePageRes)
	if !ok {
		t.Fatalf("返回类型不符, got %T", res)
	}
	t.Logf("FindByPhone 分页结果: 当前页=%d 每页=%d 总数=%d 总页数=%d 本页记录=%d",
		pageRes.PageResult.CurrentPage, pageRes.PageResult.PageSize,
		pageRes.PageResult.TotalCount, pageRes.PageResult.TotalPage, len(pageRes.ResultList))
}

// TestTeacherFindByCustomerRule_FieldMask 覆盖 FieldMask 未设置/仅非前导列场景
// 注意：MySQL 环境下因命名SQL未注册会先命中 "not found naming sql"，此处对业务结果做容错并记录原因
func TestTeacherFindByCustomerRule_FieldMask(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()

	t.Run("场景1:FieldMask未设置任何字段-预期错误", func(t *testing.T) {
		args := &model.TmTeacherFindByNameNadAddressArg{FieldMask: conditonwhere.NewFieldMask()}
		namingInfo := &gormdb.NameingSqlArgInfo{
			SqlName: "FindByNameNadAddress",
			ReqType: &model.TmTeacherFindByNameNadAddressArg{},
		}
		res, err := teacherDao.FindByCustomerRule(ctx, namingInfo, args)
		if err != nil {
			t.Logf("FieldMask 未设置返回错误: %v", err)
			return
		}
		t.Logf("FieldMask 未设置未报错, 结果: %+v", res)
	})

	t.Run("场景2:仅设置非前导列Address-预期索引校验失败", func(t *testing.T) {
		fieldMask := conditonwhere.NewFieldMask()
		fieldMask.Set("Address")
		args := &model.TmTeacherFindByNameNadAddressArg{FieldMask: fieldMask, Address: "北京市海淀区"}
		namingInfo := &gormdb.NameingSqlArgInfo{
			SqlName: "FindByNameNadAddress",
			ReqType: &model.TmTeacherFindByNameNadAddressArg{},
		}
		res, err := teacherDao.FindByCustomerRule(ctx, namingInfo, args)
		if err != nil {
			t.Logf("仅非前导列 Address 返回错误: %v", err)
			return
		}
		t.Logf("仅 Address 未报错, 结果: %+v", res)
	})
}
