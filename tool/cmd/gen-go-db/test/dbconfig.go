//go:build db

package test

import (
	"context"
	"fmt"
	"os"
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

// GetTmNoRepository 获取 tm_no DAO 实例（无主键无索引表）
func GetTmNoRepository() dao.ITmNoDao {
	db := mustOpenDB()
	return dao.NewTmNoDao(gormdb.NewRepository(db), &TestBaseDao{
		tbInfoOpts: []agdao.TbInfoOpt{
			agdao.WithTbNameStrategy(func(ctx context.Context, info *agdao.TableInfo) string {
				return "tm_no"
			}),
		},
	})
}

// GetTmNoIndexRepository 获取 tm_no_index DAO 实例（有主键无索引表）
func GetTmNoIndexRepository() dao.ITmNoIndexDao {
	db := mustOpenDB()
	return dao.NewTmNoIndexDao(gormdb.NewRepository(db), &TestBaseDao{
		tbInfoOpts: []agdao.TbInfoOpt{
			agdao.WithTbNameStrategy(func(ctx context.Context, info *agdao.TableInfo) string {
				return "tm_no_index"
			}),
		},
	})
}

// GetTmNoPrimaryRepository 获取 tm_no_primary DAO 实例（无主键有索引表）
func GetTmNoPrimaryRepository() dao.ITmNoPrimaryDao {
	db := mustOpenDB()
	return dao.NewTmNoPrimaryDao(gormdb.NewRepository(db), &TestBaseDao{
		tbInfoOpts: []agdao.TbInfoOpt{
			agdao.WithTbNameStrategy(func(ctx context.Context, info *agdao.TableInfo) string {
				return "tm_no_primary"
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

// GetDSN 根据数据库类型返回对应的连接字符串（环境变量 MYSQL_DSN / DB2_DSN，§5.3）
func GetDSN(dbType string) string {
	switch dbType {
	case "mysql":
		dsn := os.Getenv("MYSQL_DSN")
		if dsn == "" {
			panic("未设置 MYSQL_DSN 环境变量")
		}
		return dsn
	case "ibmdb":
		dsn := os.Getenv("DB2_DSN")
		if dsn == "" {
			panic("未设置 DB2_DSN 环境变量")
		}
		return dsn
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
