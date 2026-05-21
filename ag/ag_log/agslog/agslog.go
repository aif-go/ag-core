package agslog

import (
	"log/slog"
	"sync/atomic"
)

const (
	const_topLoggerName = "agslog_top"
	const_rootLevelKey  = "root"
)

// 允许应用获取slog实例
func GetSlog() *slog.Logger {
	return TopLogger()
}

func GetSlogByName(name string) *slog.Logger {
	builder := DefaultBuilder()
	logger := builder.GetSlogByName(name)
	if logger == nil {
		return TopLogger()
	}
	return logger
}

var (
	// builder 是默认的Builder
	builder atomic.Pointer[Builder]
	// TopLogger 是默认的slog.Logger
	topLogger atomic.Pointer[slog.Logger]
)

func init() {
	b := newBuilder()
	tlog := b.GetSlogByName(const_topLoggerName)
	topLogger.Store(tlog) // topLogger 默认是slog.DefaultLogger
	// topLogger.Store(slog.Default()) // topLogger 默认是slog.DefaultLogger
	builder.Store(b)
}

// TopLogger 返回默认的顶层slog.Logger
func TopLogger() *slog.Logger {
	return topLogger.Load()
}

// SetDefaultTop 设置默认的顶层slog.Logger
// func SetDefaultTop(l *slog.Logger) {
// 	topLogger.Store(l)
// }

func DefaultBuilder() *Builder {
	return builder.Load()
}
