package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aif-go/ag-core/contribute/agdb/agdao"
	"github.com/aif-go/ag-core/contribute/agdb/conditonwhere"
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/dao"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"

	// gormibmdb 内部通过 sql.Open("go_ibm_db", dsn) 建立连接，需注册该驱动
	_ "github.com/ibmdb/go_ibm_db"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestInsertOne(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	TimePointer:= func() *time.Time {
			t := time.Now()
			return &t
		}
	res, err := tmTeacherDao.InsertOne(ctx, &model.TmTeacher{
		Name:    "test1",
		Address: "上海市浦东新区",
		Phone:   "13800000000",
		ClassId: "1",
		TimePointer: TimePointer(),
		CardNo:  "沪A123M1",
	})
	printEntity("InsertOne", res, err, t)
}

func TestInsertOneIgnoreZeroVal(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	res, err := tmTeacherDao.InsertOneIgnoreZeroValCols(ctx, &model.TmTeacher{
		Name:    "aaa3",
		Address: "上海市徐汇区",
		Phone:   "10000000003",
		ClassId: "",
		BoolTest: true,
		CardNo:  "xxxx",
	})
	printEntity("InsertOneIgnoreZeroVal", res, err, t)
}

func TestUpdateByPrimaryKey(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	// tmTeacher, err := tmTeacherDao.FindByPrimaryKey(ctx, 1)
	// printEntity("FindByPrimaryKey", tmTeacher, err, t)
	// if err != nil {
	// 	t.Errorf("FindByPrimaryKey failed: %v", err)
	// }
	// tmTeacher.Name = "test3B"
	// tmTeacher.Address = "xxxx"	
	// tmTeacher.CardNo = "1234567890"
    tmTeacher := &model.TmTeacher {
		Id:      402,
		Name:    "I00001",
		Address: "I000001",
		Phone:   "10000000003",
		ClassId: "1",
		CardNo:  "I000003",
	}
	tmTeacher.JpaVersion.Int64 = 5
	tmTeacher.JpaVersion.Valid = true
	res, err := tmTeacherDao.UpdateByPrimaryKey(ctx, tmTeacher)
	printEntity("UpdateByPrimaryKey", res, err, t)
}

func TestUpdaeByPrimaryKeyIngoreZeroValCols(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	tmTeacher:= &model.TmTeacher{
		Id:      402,
		Name:    "A002",
		Address: "A0002",
		Phone:   "A0000000002",
		ClassId: "2",
		CardNo:  "A00000002",
	}
	
	tmTeacher.JpaVersion.Int64 = 1
	tmTeacher.JpaVersion.Valid = true
	res, err := tmTeacherDao.UpdateByPrimaryKeyIngoreZeroValCols(ctx, tmTeacher)

	printEntity("UpdaeByPrimaryKeyIngoreZeroValCols", res, err, t)
}

func TestFindByStruct(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()
	res, err := tmTeacherDao.FindByStruct(ctx, &model.TmTeacher{
		// Id: 2,
		Name:    "aaa3",
		Address: "上海市徐汇区",
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
	dbType := "ibmdb" // 切换数据库类型：mysql / ibmdb

	opener := gormdb.GetDBOpener(dbType)
	if opener == nil {
		panic(fmt.Sprintf("不支持的数据库驱动: %s", dbType))
	}

	db, err := gorm.Open(opener(GetDSN(dbType)), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger: logger.Default.LogMode(logger.Info),
	})

	// 乐观锁由 model 中 optimisticlock.Version 字段自动生效（v1.1.3 无 New/Config 注册 API，
	// Version 通过实现 GORM clause 接口完成 WHERE version=X + SET version=version+1）。

	if err != nil {
		panic(err.Error())
	}

	sqldb, err := db.DB()
	if err != nil {
		panic(err.Error())
	}

	sqldb.SetMaxIdleConns(10)
	sqldb.SetMaxOpenConns(10)
	sqldb.SetConnMaxLifetime(time.Second)
	sqldb.SetConnMaxIdleTime(time.Second)

	repository := gormdb.NewRepository(db)

	return dao.NewTmTeacherDao(repository, &TestBaseDao{
		tbInfoOpts: []agdao.TbInfoOpt{
			agdao.WithTbNameStrategy(func(ctx context.Context, info *agdao.TableInfo) string {
				return "tm_teacher"
			}),
		},
	})
}

// GetDSN 根据数据库类型返回对应的连接字符串
func GetDSN(dbType string) string {
	switch dbType {
	case "mysql":
		return "root:root@tcp(localhost:3306)/process?parseTime=True&loc=Local"
	case "ibmdb":
		return "HOSTNAME=192.168.105.63;DATABASE=testdb;PORT=50003;UID=db2inst1;PWD=db2inst1;AUTHENTICATION=SERVER;CurrentSchema=db2inst1"
	default:
		panic(fmt.Sprintf("不支持的数据库类型: %s", dbType))
	}
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
