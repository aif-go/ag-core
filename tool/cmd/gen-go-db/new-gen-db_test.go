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
	res, err := tmTeacherDao.InsertOneIgnoreZeroValCols(ctx, &model.TmTeacher{
		Name:    "aaa3",
		Address: "上海市徐汇区",
		Phone:   "10000000003",
		ClassId: "",
		CardNo:  "xxxx",
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
	})
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
