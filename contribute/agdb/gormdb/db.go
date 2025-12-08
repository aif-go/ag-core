package gormdb

import (
	"ag-core/ag/ag_conf"
	"time"

	gormibmdb "github.com/ZhengweiHou/gorm_ibmdb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// func NewDB(conf *viper.Viper, l *zap.Logger) *gorm.DB {
func NewDB(env ag_conf.IConfigurableEnvironment, l logger.Interface) *gorm.DB {
	var (
		db  *gorm.DB
		err error
	)

	//logger := NewGormLog(l)
	logger := l
	driver := env.GetProperty("data.db.user.driver")
	dsn := env.GetProperty("data.db.user.dsn")

	// GORM doc: https://gorm.io/docs/connecting_to_the_database.html
	switch driver {
	case "ibmdb":
		db, err = gorm.Open(gormibmdb.Open(dsn), &gorm.Config{ // 数据库不可用会报异常
			Logger: logger,
		})
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger,
		})
	default:
		panic("unknown db driver")
	}
	if err != nil {
		panic(err)
	}
	db = db.Debug()

	// Connection Pool config
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}
