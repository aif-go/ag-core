package async

import "time"

const (
	AglogAsyncKeyPrefix = "aglog.async"
)

var (
	defaultAsyncGroupConfig = AsyncGroupConfig{
		Worker:          1,
		Queue:           10000,
		FullStrategy:    "drop_new",
		ShutdownTimeout: time.Second,
	}
)

// 异步全局配置
type AsyncGlobalProperties struct {
	// Worker 组定义（可复用）
	Groups map[string]AsyncGroupConfig

	// 异步日志实例定义
	Logs map[string]AsyncLogConfig
}

// Worker 组配置
type AsyncGroupConfig struct {
	Worker       int    `value:"${:1}"`
	Queue        int    `value:"${:10000}"`
	FullStrategy string `value:"${:drop_new}"`
	// 可选：其他 worker 配置
	ShutdownTimeout time.Duration `value:"${:1s}"`
}

// 异步日志实例配置
type AsyncLogConfig struct {
	Group string // 引用 group 名称
	Log   string // 引用底层 log 名称
}
