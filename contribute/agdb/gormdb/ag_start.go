package gormdb

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_log/agslog"

	"gorm.io/gorm/logger"
)

func NewAggormDbConfig(binder ag_conf.IBinder) (*Config, error) {
	cfg := NewDefaultConfig()
	err := binder.Bind(cfg, DBConfigPrefix)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// func FindGormLogger(conf *Config, slogLogger *slog.Logger) logger.Interface {
func FindGormLoggerFromAgslog(conf *Config) logger.Interface {
	name := conf.Logger.Name
	if name == "" {
		name = "agdb"
	}
	slogLogger := agslog.GetSlogByName(name)
	log := NewSLogGormLog(slogLogger)

	// 调试模式
	if conf.Logger.Debug {
		log = log.LogMode(logger.Info)
	}

	return log
}
