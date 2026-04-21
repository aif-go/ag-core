package gormdb

const (
	DBConfigPrefix = "data.db"
)

type Config struct {
	User   UserConfig
	Pool   PoolConfig
	Logger LoggerConfig
}

type UserConfig struct {
	Driver string
	DSN    string
}

type PoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int // 连接最大存活时间 单位：秒 默认0 表示不限制
	ConnMaxIdleTime int // 连接最大空闲时间 单位：秒 默认0 表示不限制
}

type LoggerConfig struct {
	Name  string
	Debug bool
}

func NewDefaultConfig() *Config {
	return &Config{
		Logger: LoggerConfig{
			Name:  "agdb",
			Debug: false,
		},
	}
}
