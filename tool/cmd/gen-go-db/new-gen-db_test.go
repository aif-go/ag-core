package main

import (
	"context"
	"testing"
	"time"

	"github.com/aif-go/ag-core/contribute/agdb/agdao"
	"github.com/aif-go/ag-core/contribute/agdb/conditonwhere"
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/dao"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestInsertOne(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	res, err := tmTeacherDao.InsertOne(ctx, &model.TmTeacher{
		Name:    "test1",
		Address: "上海市浦东新区",
		Phone:   "13800000000",
		ClassId: "1",
		CardNo:  "沪A123M1",
	})
	printEntity("InsertOne", res, err, t)
}

func TestInsertOneIgnoreZeroVal(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	res, err := tmTeacherDao.InsertOne(ctx, &model.TmTeacher{
		Name:    "test2",
		Address: "上海市徐汇区",
		Phone:   "13800000001",
		ClassId: "",
		CardNo:  "",
	})
	printEntity("InsertOneIgnoreZeroVal", res, err, t)
}

func TestUpdateByPrimaryKey(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	res, err := tmTeacherDao.UpdateByPrimaryKey(ctx, &model.TmTeacher{
		Id:      1,
		Name:    "test1B",
		Address: "上海市浦东新区",
		Phone:   "1380000000B",
		ClassId: "",
		CardNo:  "沪BBBB",
	})
	printEntity("UpdateByPrimaryKey", res, err, t)
}

func TestUpdaeByPrimaryKeyIngoreZeroValCols(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	res, err := tmTeacherDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, &model.TmTeacher{
		Id:      2,
		Name:    "test2",
		Address: "上海市浦东新区",
		Phone:   "1380000000x",
		ClassId: "",
		CardNo:  "沪AAAA",
	})

	printEntity("UpdaeByPrimaryKeyIngoreZeroValCols", res, err, t)
}

func TestFindByStruct(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	res, err := tmTeacherDao.FindByStruct(ctx, &model.TmTeacher{
		// Id: 2,
		// Name:    "test1",
		Address: "上海市浦东新区",
	})
	printList("FindByStruct", res, err, t)
}

func TestFindByCustomRule(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	args:=&model.TmTeacherFindByNameNadAddressArg{
		FieldMask: conditonwhere.NewFieldMask(),
	}
	args.WithName("Alice").WithAddress("北京")
	res, err := tmTeacherDao.FindByCustomerRule(ctx, dao.FindByNameNadAddressNamingInfo, args)

	resEntity, ok := res.([]*model.TmTeacherFindByNameNadAddressRes)
	if !ok {
		t.Errorf("FindByCustomerRule failed: %v", err)
	}
	printList("TestFindByCustomRule", resEntity, err, t)
}

func TestFindByCustomRuleByPageMysql(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	res, err := tmTeacherDao.FindByCustomerRule(ctx, dao.FindByPhoneNamingInfo, &model.TmTeacherFindByPhoneArg{
		Phone: "13800000000",
		Page: gormdb.Page{
			PageNum:  4,
			PageSize: 3,
		},
	})

	resEntity, ok := res.(*model.TmTeacherFindByPhonePageRes)
	t.Log("TestFindByCustomRuleByPageMysql", resEntity.PageResult)
	if !ok {
		t.Errorf("FindByCustomerRuleByPage failed: %v", err)
	}
	printList("TestFindByCustomRuleByPageMysql", resEntity.ResultList, err, t)
}

// func TestUpdateDynamic(t *testing.T) {
// 	ctx := context.Background()
// 	tmTeacherDao := GetRepository()
// 	res, err := tmTeacherDao.UpdateDynamic(ctx, &model.TmTeacher{
// 		Id: 2,
// 		// Name:    "test2",
// 		Address: "上海市浦东新区",
// 		Phone:   "1380000000x",
// 		ClassId: "5",
// 		CardNo:  "沪A5678",
// 	}, []string{dao.TmTeacherColumn.ClassId})

// 	printEntity("TestUpdateDynamic", res, err, t)
// }

func GetRepository() dao.ITmTeacherDao {

	dsn := "root:root@tcp(localhost:3306)/process?parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	// 	// 此处测试ok
	// 	// res:=db.Raw("select * from tm_media_act where app_id in ?",[]string{"1" , "2"}).Find(&[]*model.TmMediaAct{})

	// 	// fmt.Println(res.RowsAffected)

	sqldb, err := db.DB()
	if err != nil {
		panic(err.Error())
	}

	sqldb.SetMaxIdleConns(10)
	sqldb.SetMaxOpenConns(10)
	sqldb.SetConnMaxLifetime(time.Second)
	sqldb.SetConnMaxIdleTime(time.Second)

	// 	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	// 		Level: slog.LevelDebug, // 输出级别：Debug及以上
	// 	})
	// 	logger := slog.New(jsonHandler)
	repository := gormdb.NewRepository(db)

	return dao.NewTmTeacherDao(repository, &TestBaseDao{
		tbInfoOpts: []agdao.TbInfoOpt{
			agdao.WithTbNameStrategy(func(ctx context.Context, info *agdao.TableInfo) string {
				return "tm_teacher"
			}),
		},
	})
}

type TestBaseDao struct {
	tbInfoOpts []agdao.TbInfoOpt
}

func (dao *TestBaseDao) ApplyTbInfoOpts(ctx context.Context, info *agdao.TableInfo) {
	for _, opt := range dao.tbInfoOpts {
		opt(ctx, info)
	}
}

func (dao *TestBaseDao) RegTbInfoOpt(opts ...agdao.TbInfoOpt) {
	dao.tbInfoOpts = append(dao.tbInfoOpts, opts...)
}

// func TestInsertOne(t *testing.T) {
// 	// 测试插入一条数据
// 	err := dao.InsertOne(&model.User{
// 		Username: "testuser",
// 		Password: "testpass",
// 	})
// 	if err != nil {
// 		t.Errorf("InsertOne failed: %v", err)
// 	}
// }

func printList[T any](name string, list []T, err error, t *testing.T) {

	t.Logf(" *** %v *** ", name)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	if len(list) == 0 {
		t.Log("未查询到预期的数据")
		t.Fail()
	}
	for _, entity := range list {
		t.Logf("结果:%v \n", entity)
	}
	t.Logf(" *** %v end *** ", name)
}

func printEntity(name string, entity interface{}, err error, t *testing.T) {
	t.Logf(" *** %v *** ", name)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	t.Logf("查询结果%v \n", entity)
	t.Logf(" *** %v end *** ", name)
}
