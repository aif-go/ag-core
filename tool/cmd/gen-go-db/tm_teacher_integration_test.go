package main

import (
	"context"
	"os"
	"testing"

	"github.com/aif-go/ag-core/contribute/agdb/agdao"
	"github.com/aif-go/ag-core/contribute/agdb/conditonwhere"
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/dao"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var tmCtx = context.Background()
var tmDao dao.ITmTeacherDao
var tmRepository *gormdb.Repository

func TestMain(m *testing.M) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(localhost:3306)/process?parseTime=True&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get sql.DB: " + err.Error())
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(10)

	tmRepository = gormdb.NewRepository(db)
	tmDao = dao.NewTmTeacherDao(tmRepository, &TestBaseDao{
		tbInfoOpts: []agdao.TbInfoOpt{
			agdao.WithTbNameStrategy(func(ctx context.Context, info *agdao.TableInfo) string {
				return "tm_teacher"
			}),
		},
	})

	code := m.Run()
	os.Exit(code)
}

func cleanData(t *testing.T) {
	t.Helper()
	err := tmRepository.DB(tmCtx).Exec("DELETE FROM tm_teacher WHERE name LIKE 'test%'").Error
	if err != nil {
		t.Logf("cleanData warning: %v", err)
	}
}

func insertOne(t *testing.T) *model.TmTeacher {
	t.Helper()
	entity := &model.TmTeacher{
		Name:    "test1",
		Address: "上海市浦东新区",
		Phone:   "13800000000",
		ClassId: "1",
		CardNo:  "沪A123M1",
	}
	_, err := tmDao.InsertOne(tmCtx, entity)
	if err != nil {
		t.Fatalf("insertOne failed: %v", err)
	}
	return entity
}

// InsertOne 插入完整数据
func TestTmTeacher_Insert(t *testing.T) {
	cleanData(t)
	entity := &model.TmTeacher{
		Name:    "test1",
		Address: "上海市浦东新区",
		Phone:   "13800000000",
		ClassId: "1",
		CardNo:  "沪A123M1",
	}
	rows, err := tmDao.InsertOne(tmCtx, entity)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}
	if entity.Id <= 0 {
		t.Errorf("expected auto-assigned id > 0, got %d", entity.Id)
	}
}

// InsertOneIgnoreZeroValCols 插入跳过零值列
func TestTmTeacher_InsertSkipZero(t *testing.T) {
	cleanData(t)
	entity := &model.TmTeacher{
		Name:    "test2",
		Address: "上海市徐汇区",
		Phone:   "13800000001",
	}
	rows, err := tmDao.InsertOneIgnoreZeroValCols(tmCtx, entity)
	if err != nil {
		t.Fatalf("InsertOneIgnoreZeroValCols failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}
	if entity.Id <= 0 {
		t.Errorf("expected auto-assigned id > 0, got %d", entity.Id)
	}
}

// FindByPrimaryKey 存在记录时返回实体
func TestTmTeacher_Get(t *testing.T) {
	cleanData(t)
	inserted := insertOne(t)

	entity, err := tmDao.FindByPrimaryKey(tmCtx, model.TmTeacherPrimaryKey(inserted.Id))
	if err != nil {
		t.Fatalf("FindByPrimaryKey failed: %v", err)
	}
	if entity == nil {
		t.Fatal("expected non-nil entity")
	}
	if entity.Name != "test1" {
		t.Errorf("expected name=test1, got %s", entity.Name)
	}
	if entity.Address != "上海市浦东新区" {
		t.Errorf("expected address=上海市浦东新区, got %s", entity.Address)
	}
	if entity.Phone != "13800000000" {
		t.Errorf("expected phone=13800000000, got %s", entity.Phone)
	}
	if entity.ClassId != "1" {
		t.Errorf("expected classId=1, got %s", entity.ClassId)
	}
	if entity.CardNo != "沪A123M1" {
		t.Errorf("expected cardNo=沪A123M1, got %s", entity.CardNo)
	}
}

// FindByPrimaryKey 不存在记录时返回 nil
func TestTmTeacher_GetNotFound(t *testing.T) {
	cleanData(t)

	entity, err := tmDao.FindByPrimaryKey(tmCtx, model.TmTeacherPrimaryKey(999999))
	if err != nil {
		t.Fatalf("FindByPrimaryKey failed: %v", err)
	}
	if entity != nil {
		t.Fatal("expected nil for non-existent record")
	}
}

// FindByStruct 按主键 id 查询（走主键快速路径）
func TestTmTeacher_List(t *testing.T) {
	cleanData(t)
	inserted := insertOne(t)

	list, err := tmDao.FindByStruct(tmCtx, &model.TmTeacher{
		Id: inserted.Id,
	})
	if err != nil {
		t.Fatalf("FindByStruct failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty result list")
	}
	if list[0].Name != "test1" {
		t.Errorf("expected name=test1, got %s", list[0].Name)
	}
}

// FindByStruct 按索引列 phone 查询（走索引路径）
func TestTmTeacher_ListByIndex(t *testing.T) {
	cleanData(t)
	insertOne(t)

	list, err := tmDao.FindByStruct(tmCtx, &model.TmTeacher{
		Phone: "13800000000",
	})
	if err != nil {
		t.Fatalf("FindByStruct failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty result list")
	}
}

// FindByStruct 按索引列 name 查询（走复合索引 tm_teacher_name_IDX）
func TestTmTeacher_ListByName(t *testing.T) {
	cleanData(t)
	insertOne(t)

	list, err := tmDao.FindByStruct(tmCtx, &model.TmTeacher{
		Name: "test1",
	})
	if err != nil {
		t.Fatalf("FindByStruct failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty result list")
	}
}

// FindByStruct 按索引列 name+address 组合查询（走复合索引全部列）
func TestTmTeacher_ListByNameAndAddress(t *testing.T) {
	cleanData(t)
	insertOne(t)

	list, err := tmDao.FindByStruct(tmCtx, &model.TmTeacher{
		Name:    "test1",
		Address: "上海市浦东新区",
	})
	if err != nil {
		t.Fatalf("FindByStruct failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty result list")
	}
}

// UpdateByPrimaryKey 全字段更新后重新查询确认
func TestTmTeacher_Update(t *testing.T) {
	cleanData(t)
	inserted := insertOne(t)

	inserted.Name = "test1_updated"
	inserted.Address = "北京市朝阳区"
	inserted.Phone = "13900000000"
	inserted.ClassId = "2"
	inserted.CardNo = "沪B999M9"

	rows, err := tmDao.UpdateByPrimaryKey(tmCtx, inserted)
	if err != nil {
		t.Fatalf("UpdateByPrimaryKey failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	entity, err := tmDao.FindByPrimaryKey(tmCtx, model.TmTeacherPrimaryKey(inserted.Id))
	if err != nil {
		t.Fatalf("FindByPrimaryKey after update failed: %v", err)
	}
	if entity.Name != "test1_updated" {
		t.Errorf("expected name=test1_updated, got %s", entity.Name)
	}
	if entity.Address != "北京市朝阳区" {
		t.Errorf("expected address=北京市朝阳区, got %s", entity.Address)
	}
	if entity.Phone != "13900000000" {
		t.Errorf("expected phone=13900000000, got %s", entity.Phone)
	}
	if entity.ClassId != "2" {
		t.Errorf("expected classId=2, got %s", entity.ClassId)
	}
	if entity.CardNo != "沪B999M9" {
		t.Errorf("expected cardNo=沪B999M9, got %s", entity.CardNo)
	}
}

// UpdateByPrimaryKeyIngoreZeroValCols 更新跳过零值列，未赋值字段保持不变
func TestTmTeacher_UpdateSkipZero(t *testing.T) {
	cleanData(t)
	inserted := insertOne(t)

	update := &model.TmTeacher{
		Id:    inserted.Id,
		Name:  "test2_updated",
		Phone: "13800000001",
	}
	rows, err := tmDao.UpdateByPrimaryKeyIngoreZeroValCols(tmCtx, update)
	if err != nil {
		t.Fatalf("UpdateByPrimaryKeyIngoreZeroValCols failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	entity, err := tmDao.FindByPrimaryKey(tmCtx, model.TmTeacherPrimaryKey(inserted.Id))
	if err != nil {
		t.Fatalf("FindByPrimaryKey after update failed: %v", err)
	}
	if entity.Name != "test2_updated" {
		t.Errorf("expected name=test2_updated, got %s", entity.Name)
	}
	if entity.Phone != "13800000001" {
		t.Errorf("expected phone=13800000001, got %s", entity.Phone)
	}
	if entity.Address != "上海市浦东新区" {
		t.Errorf("expected address unchanged=上海市浦东新区, got %s", entity.Address)
	}
}

// FindByCondition 带排序的条件查询
func TestTmTeacher_ListWithCondition(t *testing.T) {
	cleanData(t)
	insertOne(t)

	cond := conditonwhere.NewWhereClauseBuilder().
		Eq("name", "test1")
	order := gormdb.NewOrderBuilder().
		Asc("id")

	list, _, err := tmDao.FindByCondition(tmCtx, cond, order, nil)
	if err != nil {
		t.Fatalf("FindByCondition failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty result list")
	}
	for _, entity := range list {
		if entity.Name != "test1" {
			t.Errorf("expected name=test1, got %s", entity.Name)
		}
	}
}

// FindByCondition 带分页的条件查询
func TestTmTeacher_ListWithConditionAndPage(t *testing.T) {
	cleanData(t)
	for i := 0; i < 5; i++ {
		entity := &model.TmTeacher{
			Name:    "test_page",
			Address: "地址",
			Phone:   "13800000000",
			ClassId: "1",
			CardNo:  "card",
		}
		_, err := tmDao.InsertOne(tmCtx, entity)
		if err != nil {
			t.Fatalf("InsertOne failed: %v", err)
		}
	}

	cond := conditonwhere.NewWhereClauseBuilder().
		Eq("name", "test_page")
	page := &gormdb.Page{
		PageNum:  1,
		PageSize: 2,
	}

	list, pageRes, err := tmDao.FindByCondition(tmCtx, cond, nil, page)
	if err != nil {
		t.Fatalf("FindByCondition with page failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 records per page, got %d", len(list))
	}
	if pageRes.TotalCount != 5 {
		t.Errorf("expected total count 5, got %d", pageRes.TotalCount)
	}
	if pageRes.CurrentPage != 1 {
		t.Errorf("expected current page 1, got %d", pageRes.CurrentPage)
	}
	if pageRes.PageSize != 2 {
		t.Errorf("expected page size 2, got %d", pageRes.PageSize)
	}
	if pageRes.TotalPage != 3 {
		t.Errorf("expected total page 3, got %d", pageRes.TotalPage)
	}
}

// FindFirstOneByCondition 条件查询返回单条记录
func TestTmTeacher_GetFirst(t *testing.T) {
	cleanData(t)
	insertOne(t)

	cond := conditonwhere.NewWhereClauseBuilder().
		Eq("name", "test1")

	entity, err := tmDao.FindFirstOneByCondition(tmCtx, cond, nil)
	if err != nil {
		t.Fatalf("FindFirstOneByCondition failed: %v", err)
	}
	if entity == nil {
		t.Fatal("expected non-nil entity")
	}
	if entity.Name != "test1" {
		t.Errorf("expected name=test1, got %s", entity.Name)
	}
}

// FindByCustomerRule 非分页命名查询（FindByNameNadAddress）
func TestTmTeacher_CustomRule(t *testing.T) {
	cleanData(t)
	inserted := insertOne(t)

	args := &model.TmTeacherFindByNameNadAddressArg{
		FieldMask: conditonwhere.NewFieldMask(),
		Name:      inserted.Name,
		Address:   inserted.Address,
	}
	args.FieldMask.Set("Name")
	args.FieldMask.Set("Address")

	res, err := tmDao.FindByCustomerRule(tmCtx, dao.FindByNameNadAddressNamingInfo, args)
	if err != nil {
		t.Fatalf("FindByCustomerRule failed: %v", err)
	}
	resList, ok := res.([]*model.TmTeacherFindByNameNadAddressRes)
	if !ok {
		t.Fatal("expected result type []*TmTeacherFindByNameNadAddressRes")
	}
	if len(resList) == 0 {
		t.Fatal("expected non-empty result list")
	}
}

// FindByCustomerRule 分页命名查询（FindByPhone）
func TestTmTeacher_CustomRuleWithPage(t *testing.T) {
	cleanData(t)
	insertOne(t)

	args := &model.TmTeacherFindByPhoneArg{
		FieldMask: conditonwhere.NewFieldMask(),
		Phone:     "13800000000",
		Page: gormdb.Page{
			PageNum:  1,
			PageSize: 10,
		},
	}
	args.FieldMask.Set("Phone")

	res, err := tmDao.FindByCustomerRule(tmCtx, dao.FindByPhoneNamingInfo, args)
	if err != nil {
		t.Fatalf("FindByCustomerRule failed: %v", err)
	}
	pageRes, ok := res.(*model.TmTeacherFindByPhonePageRes)
	if !ok {
		t.Fatal("expected result type *TmTeacherFindByPhonePageRes")
	}
	if len(pageRes.ResultList) == 0 {
		t.Fatal("expected non-empty result list")
	}
	if pageRes.TotalCount <= 0 {
		t.Errorf("expected total count > 0, got %d", pageRes.TotalCount)
	}
}

// Transaction 事务提交后数据持久化
// Transaction 事务回滚：fn 返回 error 时插入不持久化
func TestTmTeacher_TransactionRollback(t *testing.T) {
	cleanData(t)

	rollbackName := "test_rollback"
	sentinelErr := os.ErrInvalid
	err := tmRepository.Transaction(tmCtx, func(ctx context.Context) error {
		_, err := tmDao.InsertOne(ctx, &model.TmTeacher{
			Name:    rollbackName,
			Address: "rollback_addr",
			Phone:   "13800000998",
			ClassId: "9",
			CardNo:  "rollback_card",
		})
		if err != nil {
			return err
		}
		return sentinelErr
	})
	if err != sentinelErr {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	cond := conditonwhere.NewWhereClauseBuilder().
		Eq("name", rollbackName)
	list, _, err := tmDao.FindByCondition(tmCtx, cond, nil, nil)
	if err != nil {
		t.Fatalf("FindByCondition failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 records after rollback, got %d", len(list))
	}
}

// Transaction 事务提交：fn 返回 nil 时插入持久化
func TestTmTeacher_TransactionCommit(t *testing.T) {
	cleanData(t)

	commitName := "test_commit"
	err := tmRepository.Transaction(tmCtx, func(ctx context.Context) error {
		_, err := tmDao.InsertOne(ctx, &model.TmTeacher{
			Name:    commitName,
			Address: "commit_addr",
			Phone:   "13800000997",
			ClassId: "9",
			CardNo:  "commit_card",
		})
		return err
	})
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	cond := conditonwhere.NewWhereClauseBuilder().
		Eq("name", commitName)
	list, _, err := tmDao.FindByCondition(tmCtx, cond, nil, nil)
	if err != nil {
		t.Fatalf("FindByCondition failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 committed record, got %d", len(list))
	}
}
