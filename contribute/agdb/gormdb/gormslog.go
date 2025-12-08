package gormdb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/logger"
)

const ctxSlogLoggerKey = "slogLogger"

type GormSlogLogger struct {
	SlogLogger                *slog.Logger
	SlowThreshold             time.Duration
	Colorful                  bool
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
	LogLevel                  logger.LogLevel
}

func NewSLogGormLog(slogLogger *slog.Logger) logger.Interface {
	return &GormSlogLogger{
		SlogLogger:                slogLogger,
		LogLevel:                  logger.Warn,
		SlowThreshold:             100 * time.Millisecond,
		Colorful:                  false,
		IgnoreRecordNotFoundError: false,
		ParameterizedQueries:      false,
	}
}

func (l *GormSlogLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l GormSlogLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.logger(ctx).Info(fmt.Sprintf(msg, data...))
	}
}

// Warn print warn messages
func (l GormSlogLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.logger(ctx).Warn(fmt.Sprintf(msg, data...))
	}
}

// Error print error messages
func (l GormSlogLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.logger(ctx).Error(fmt.Sprintf(msg, data...))
	}
}

func (l GormSlogLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	elapsedStr := fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)
	logs := l.logger(ctx)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		logs.Log(ctx, slog.LevelError, "trace",
			"error", err,
			"elapsed", elapsedStr,
			"rows", rows,
			"sql", sql)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		logs.Log(ctx, slog.LevelWarn, "trace",
			"slow", slowLog,
			"elapsed", elapsedStr,
			"rows", rows,
			"sql", sql)
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		logs.Log(ctx, slog.LevelInfo, "trace",
			"elapsed", elapsedStr,
			"rows", rows,
			"sql", sql)
	}
}

var (
	gormSlogPackage = filepath.Join("gorm.io", "gorm")
)

func (l GormSlogLogger) logger(ctx context.Context) *slog.Logger {
	logger := l.SlogLogger
	if ctx != nil {
		if c, ok := ctx.(*gin.Context); ok {
			ctx = c.Request.Context()
		}
		sl := ctx.Value(ctxSlogLoggerKey)
		ctxLogger, ok := sl.(*slog.Logger)
		if ok {
			logger = ctxLogger
		}
	}

	for i := 2; i < 15; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.HasSuffix(file, "_test.go"):
		case strings.Contains(file, gormSlogPackage):
		default:
			return logger.With(
				slog.String("caller", file),
			)
		}
	}
	return logger
}
