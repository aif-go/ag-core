package gormdb

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
	"fmt"
	"sync"
	"time"

	gormibmdb "github.com/ZhengweiHou/gorm_ibmdb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// func NewDB(conf *viper.Viper, l *zap.Logger) *gorm.DB {
func NewDB(env ag_conf.IConfigurableEnvironment, l logger.Interface) (*gorm.DB, error) {
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
		return nil, fmt.Errorf("unknown db driver %s", driver)
	}
	if err != nil {
		return nil, err
	}
	db = db.Debug()

	// Connection Pool config
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func NewDB_V2(cfg *Config, l logger.Interface) (*gorm.DB, error) {
	// 1. 基础配置校验
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if cfg.User.Driver == "" {
		return nil, fmt.Errorf("db driver is empty")
	}
	if cfg.User.DSN == "" {
		return nil, fmt.Errorf("db dsn is empty")
	}

	// 2. 获取驱动 opener
	opener := GetDBOpener(cfg.User.Driver)
	if opener == nil {
		return nil, fmt.Errorf("unsupported db driver: %s", cfg.User.Driver)
	}

	// 3. GORM 配置
	gormConf := &gorm.Config{
		Logger: l,
	}

	// 4. 打开连接
	db, err := gorm.Open(opener(cfg.User.DSN), gormConf)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// 5. 调试模式 FIXME 调整到ag_start.go创建logger时设置
	// if cfg.Debug {
	// 	db = db.Debug()
	// }

	// 6. 获取底层 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql db: %w", err)
	}

	// 7. 连接池配置（带合法校验）
	if cfg.Pool.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.Pool.MaxIdleConns)
	}

	if cfg.Pool.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.Pool.MaxOpenConns)
	}

	if cfg.Pool.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.Pool.ConnMaxLifetime) * time.Second)
	}

	if cfg.Pool.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Pool.ConnMaxIdleTime) * time.Second)
	}

	// 8. 关键：Ping 测试连接是否真实可用
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	return db, nil
}

var (
	dbOpenerMap map[string]DBOpener
	mutex       sync.RWMutex
)

func init() {
	dbOpenerMap = make(map[string]DBOpener)
	_ = RegisterDBOpener("ibmdb", gormibmdb.Open)
	_ = RegisterDBOpener("mysql", mysql.Open)
}

type DBOpener func(dsn string) gorm.Dialector

// RegisterDBOpener 注册驱动（支持判重）
func RegisterDBOpener(driver string, opener DBOpener) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := dbOpenerMap[driver]; exists {
		return fmt.Errorf("db driver %s already registered", driver)
	}

	dbOpenerMap[driver] = opener
	return nil
}

func GetDBOpener(driver string) DBOpener {
	mutex.RLock()
	defer mutex.RUnlock()
	return dbOpenerMap[driver]
}
