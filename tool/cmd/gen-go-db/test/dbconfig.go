package test

import (
	"context"
	"fmt"
	"time"

	"github.com/aif-go/ag-core/contribute/agdb/agdao"
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/dao"
	_ "github.com/ibmdb/go_ibm_db"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DbType 数据库类型：mysql / ibmdb
var DbType string = "mysql"

// GetRepository 获取 tm_teacher DAO 实例
func GetRepository() dao.ITmTeacherDao {
	db := mustOpenDB()
	return dao.NewTmTeacherDao(gormdb.NewRepository(db), &TestBaseDao{
		tbInfoOpts: []agdao.TbInfoOpt{
			agdao.WithTbNameStrategy(func(ctx context.Context, info *agdao.TableInfo) string {
				return "tm_teacher"
			}),
		},
	})
}

// GetStudentRepository 获取 tm_student DAO 实例
func GetStudentRepository() dao.ITmStudentDao {
	db := mustOpenDB()
	return dao.NewTmStudentDao(gormdb.NewRepository(db), &TestBaseDao{
		tbInfoOpts: []agdao.TbInfoOpt{
			agdao.WithTbNameStrategy(func(ctx context.Context, info *agdao.TableInfo) string {
				return "tm_student"
			}),
		},
	})
}

// mustOpenDB 建立数据库连接，失败直接 panic
func mustOpenDB() *gorm.DB {
	opener := gormdb.GetDBOpener(DbType)
	if opener == nil {
		panic(fmt.Sprintf("不支持的数据库驱动: %s", DbType))
	}

	db, err := gorm.Open(opener(GetDSN(DbType)), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Info),
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

	return db
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

// TestBaseDao 测试用 BaseDao 实现
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
