package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aif-go/ag-core/contribute/agdb/agdao"
	"github.com/aif-go/ag-core/contribute/agdb/conditonwhere"
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/dao"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"github.com/shopspring/decimal"
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

	// 初始清理：确保无残留数据
	if err := tmRepository.DB(tmCtx).Exec("DELETE FROM tm_teacher WHERE name LIKE 'UT_%'").Error; err != nil {
		panic("initial cleanup failed: " + err.Error())
	}

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
	err := tmRepository.DB(tmCtx).Exec("DELETE FROM tm_teacher WHERE name LIKE 'UT_%'").Error
	if err != nil {
		t.Logf("cleanData warning: %v", err)
	}
}

func insertOne(t *testing.T) *model.TmTeacher {
	t.Helper()
	entity := &model.TmTeacher{
		Name:    "UT_Alice",
		Address: "北京中关村",
		Phone:   "13912345678",
		ClassId: "A1",
		CardNo:  "CC_10001",
		Salary:  decimal.NewFromFloat(10000.51),
	}
	_, err := tmDao.InsertOne(tmCtx, entity)
	if err != nil {
		t.Fatalf("insertOne failed: %v", err)
	}
	return entity
}

func TestTmTeacher_Insert(t *testing.T) {
	cleanData(t)
	entity := &model.TmTeacher{
		Name:    "UT_Bob",
		Address: "上海张江",
		Phone:   "15987654321",
		ClassId: "B2",
		CardNo:  "CC_20001",
		Salary:  decimal.NewFromFloat(20000.75),
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

func TestTmTeacher_InsertSkipZero(t *testing.T) {
	cleanData(t)
	entity := &model.TmTeacher{
		Name:    "UT_Carol",
		Address: "深圳南山",
		Phone:   "17712345678",
		Salary:  decimal.NewFromFloat(30000.25),
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
	t.Logf("FindByPrimaryKey result: %+v", entity)
	if entity.Name != "UT_Alice" {
		t.Errorf("expected name=UT_Alice, got %s", entity.Name)
	}
	if entity.Address != "北京中关村" {
		t.Errorf("expected address=北京中关村, got %s", entity.Address)
	}
	if entity.Phone != "13912345678" {
		t.Errorf("expected phone=13912345678, got %s", entity.Phone)
	}
	if entity.ClassId != "A1" {
		t.Errorf("expected classId=A1, got %s", entity.ClassId)
	}
	if entity.CardNo != "CC_10001" {
		t.Errorf("expected cardNo=CC_10001, got %s", entity.CardNo)
	}
}

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
	for i, item := range list {
		t.Logf("FindByStruct result[%d]: %+v", i, item)
	}
	if list[0].Name != "UT_Alice" {
		t.Errorf("expected name=UT_Alice, got %s", list[0].Name)
	}
}

func TestTmTeacher_ListByIndex(t *testing.T) {
	cleanData(t)
	insertOne(t)

	list, err := tmDao.FindByStruct(tmCtx, &model.TmTeacher{
		Phone: "13912345678",
	})
	if err != nil {
		t.Fatalf("FindByStruct failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty result list")
	}
	for i, item := range list {
		t.Logf("FindByStruct by index result[%d]: %+v", i, item)
	}
}

func TestTmTeacher_ListByName(t *testing.T) {
	cleanData(t)
	insertOne(t)

	list, err := tmDao.FindByStruct(tmCtx, &model.TmTeacher{
		Name: "UT_Alice",
	})
	if err != nil {
		t.Fatalf("FindByStruct failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty result list")
	}
	for i, item := range list {
		t.Logf("FindByStruct by name result[%d]: %+v", i, item)
	}
}

func TestTmTeacher_ListByNameAndAddress(t *testing.T) {
	cleanData(t)
	insertOne(t)

	list, err := tmDao.FindByStruct(tmCtx, &model.TmTeacher{
		Name:    "UT_Alice",
		Address: "北京中关村",
	})
	if err != nil {
		t.Fatalf("FindByStruct failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty result list")
	}
	for i, item := range list {
		t.Logf("FindByStruct by name+address result[%d]: %+v", i, item)
	}
}

func TestTmTeacher_Update(t *testing.T) {
	cleanData(t)
	inserted := insertOne(t)

	inserted.Name = "UT_Alice_Updated"
	inserted.Address = "杭州西湖"
	inserted.Phone = "18812345678"
	inserted.ClassId = "A9"
	inserted.CardNo = "CC_99999"

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
	t.Logf("FindByPrimaryKey after update result: %+v", entity)
	if entity.Name != "UT_Alice_Updated" {
		t.Errorf("expected name=UT_Alice_Updated, got %s", entity.Name)
	}
	if entity.Address != "杭州西湖" {
		t.Errorf("expected address=杭州西湖, got %s", entity.Address)
	}
	if entity.Phone != "18812345678" {
		t.Errorf("expected phone=18812345678, got %s", entity.Phone)
	}
	if entity.ClassId != "A9" {
		t.Errorf("expected classId=A9, got %s", entity.ClassId)
	}
	if entity.CardNo != "CC_99999" {
		t.Errorf("expected cardNo=CC_99999, got %s", entity.CardNo)
	}
}

func TestTmTeacher_UpdateSkipZero(t *testing.T) {
	cleanData(t)
	inserted := insertOne(t)

	update := &model.TmTeacher{
		Id:    inserted.Id,
		Name:  "UT_David",
		Phone: "16612345678",
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

	t.Logf("FindByPrimaryKey after skip-zero update result: %+v", entity)
	if entity.Name != "UT_David" {
		t.Errorf("expected name=UT_David, got %s", entity.Name)
	}
	if entity.Phone != "16612345678" {
		t.Errorf("expected phone=16612345678, got %s", entity.Phone)
	}
	if entity.Address != "北京中关村" {
		t.Errorf("expected address unchanged=北京中关村, got %s", entity.Address)
	}
}

func TestTmTeacher_ListWithCondition(t *testing.T) {
	cleanData(t)
	insertOne(t)

	cond := conditonwhere.NewWhereClauseBuilder().
		Eq("name", "UT_Alice")
	order := gormdb.NewOrderBuilder().
		Asc("id")

	list, _, err := tmDao.FindByCondition(tmCtx, cond, order, nil)
	if err != nil {
		t.Fatalf("FindByCondition failed: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty result list")
	}
	for i, item := range list {
		t.Logf("FindByCondition result[%d]: %+v", i, item)
	}
	for _, entity := range list {
		if entity.Name != "UT_Alice" {
			t.Errorf("expected name=UT_Alice, got %s", entity.Name)
		}
	}
}

func TestTmTeacher_ListWithConditionAndPage(t *testing.T) {
	cleanData(t)
	for i := 0; i < 5; i++ {
		entity := &model.TmTeacher{
			Name:    "UT_PageUser",
			Address: "测试地址",
			Phone:   fmt.Sprintf("1000000000%d", i),
			ClassId: "P1",
			CardNo:  fmt.Sprintf("PAGE_CARD_%d", i),
			Salary:  decimal.NewFromFloat(40000.00),
		}
		_, err := tmDao.InsertOne(tmCtx, entity)
		if err != nil {
			t.Fatalf("InsertOne failed: %v", err)
		}
	}

	cond := conditonwhere.NewWhereClauseBuilder().
		Eq("name", "UT_PageUser")
	page := &gormdb.Page{
		PageNum:  1,
		PageSize: 2,
	}

	list, pageRes, err := tmDao.FindByCondition(tmCtx, cond, nil, page)
	if err != nil {
		t.Fatalf("FindByCondition with page failed: %v", err)
	}
	for i, item := range list {
		t.Logf("result[%d]: %+v", i, item)
	}
	t.Logf("pageRes: %+v", pageRes)
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

func TestTmTeacher_GetFirst(t *testing.T) {
	cleanData(t)
	insertOne(t)

	cond := conditonwhere.NewWhereClauseBuilder().
		Eq("name", "UT_Alice")

	entity, err := tmDao.FindFirstOneByCondition(tmCtx, cond, nil)
	if err != nil {
		t.Fatalf("FindFirstOneByCondition failed: %v", err)
	}
	if entity == nil {
		t.Fatal("expected non-nil entity")
	}
	t.Logf("FindFirstOneByCondition result: %+v", entity)
	if entity.Name != "UT_Alice" {
		t.Errorf("expected name=UT_Alice, got %s", entity.Name)
	}
}

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
	for i, item := range resList {
		t.Logf("CustomRule result[%d]: %+v", i, item)
	}
}

func TestTmTeacher_CustomRuleWithPage(t *testing.T) {
	cleanData(t)
	insertOne(t)

	args := &model.TmTeacherFindByPhoneArg{
		FieldMask: conditonwhere.NewFieldMask(),
		Phone:     "13912345678",
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
	for i, item := range pageRes.ResultList {
		t.Logf("CustomRuleWithPage result[%d]: %+v", i, item)
	}
	t.Logf("CustomRuleWithPage pageRes: %+v", pageRes)
	if pageRes.TotalCount <= 0 {
		t.Errorf("expected total count > 0, got %d", pageRes.TotalCount)
	}
}

func TestTmTeacher_TransactionRollback(t *testing.T) {
	cleanData(t)

	rollbackName := "UT_Rollback"
	sentinelErr := os.ErrInvalid
	err := tmRepository.Transaction(tmCtx, func(ctx context.Context) error {
		_, err := tmDao.InsertOne(ctx, &model.TmTeacher{
			Name:    rollbackName,
			Address: "回滚地址",
			Phone:   "10000000001",
			ClassId: "R1",
			CardNo:  "ROLLBACK",
			Salary:  decimal.NewFromFloat(50000.99),
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
	for i, item := range list {
		t.Logf("FindByCondition after rollback result[%d]: %+v", i, item)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 records after rollback, got %d", len(list))
	}
}

func TestTmTeacher_TransactionCommit(t *testing.T) {
	cleanData(t)

	commitName := "UT_Commit"
	err := tmRepository.Transaction(tmCtx, func(ctx context.Context) error {
		_, err := tmDao.InsertOne(ctx, &model.TmTeacher{
			Name:    commitName,
			Address: "提交地址",
			Phone:   "10000000002",
			ClassId: "C1",
			CardNo:  "COMMIT_CARD",
			Salary:  decimal.NewFromFloat(60000.10),
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
	for i, item := range list {
		t.Logf("FindByCondition after commit result[%d]: %+v", i, item)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 committed record, got %d", len(list))
	}
}
