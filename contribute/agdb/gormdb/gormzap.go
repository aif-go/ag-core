package gormdb

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
	gormlog "gorm.io/gorm/logger"
)

const ctxLoggerKey = "zapLogger"

type Logger struct {
	ZapLogger                 *zap.Logger
	SlowThreshold             time.Duration
	Colorful                  bool
	IgnoreRecordNotFoundError bool
	ParameterizedQueries      bool
	LogLevel                  gormlog.LogLevel
}

func NewZapGormLog(zapLogger *zap.Logger) gormlog.Interface {
	zapLogger.Level()
	return &Logger{
		ZapLogger:                 zapLogger,
		LogLevel:                  gormlog.Warn,
		SlowThreshold:             100 * time.Millisecond,
		Colorful:                  false,
		IgnoreRecordNotFoundError: false,
		ParameterizedQueries:      false,
	}
}

func (l *Logger) LogMode(level gormlog.LogLevel) gormlog.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlog.Info {
		l.logger(ctx).Sugar().Infof(msg, data...)
	}
}

// Warn print warn messages
func (l Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlog.Warn {
		l.logger(ctx).Sugar().Warnf(msg, data...)
	}
}

// Error print error messages
func (l Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlog.Error {
		l.logger(ctx).Sugar().Errorf(msg, data...)
	}
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlog.Silent {
		return
	}

	elapsed := time.Since(begin)
	elapsedStr := fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)
	logger := l.logger(ctx)
	switch {
	case err != nil && l.LogLevel >= gormlog.Error && (!errors.Is(err, gormlog.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			logger.Error("trace", zap.Error(err), zap.String("elapsed", elapsedStr), zap.Int64("rows", rows), zap.String("sql", sql))
		} else {
			logger.Error("trace", zap.Error(err), zap.String("elapsed", elapsedStr), zap.Int64("rows", rows), zap.String("sql", sql))
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlog.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			logger.Warn("trace", zap.String("slow", slowLog), zap.String("elapsed", elapsedStr), zap.Int64("rows", rows), zap.String("sql", sql))
		} else {
			logger.Warn("trace", zap.String("slow", slowLog), zap.String("elapsed", elapsedStr), zap.Int64("rows", rows), zap.String("sql", sql))
		}
	case l.LogLevel == gormlog.Info:
		sql, rows := fc()
		if rows == -1 {
			logger.Info("trace", zap.String("elapsed", elapsedStr), zap.Int64("rows", rows), zap.String("sql", sql))
		} else {
			logger.Info("trace", zap.String("elapsed", elapsedStr), zap.Int64("rows", rows), zap.String("sql", sql))
		}
	}
}

var (
	gormPackage = filepath.Join("gorm.io", "gorm")
)

func (l Logger) logger(ctx context.Context) *zap.Logger {
	logger := l.ZapLogger
	if ctx != nil {
		if c, ok := ctx.(*gin.Context); ok {
			ctx = c.Request.Context()
		}
		zl := ctx.Value(ctxLoggerKey)
		ctxLogger, ok := zl.(*zap.Logger)
		if ok {
			logger = ctxLogger
		}
	}

	for i := 2; i < 15; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.HasSuffix(file, "_test.go"):
		case strings.Contains(file, gormPackage):
		default:
			return logger.WithOptions(zap.AddCallerSkip(i - 1))
		}
	}
	return logger
}

/*
 */
