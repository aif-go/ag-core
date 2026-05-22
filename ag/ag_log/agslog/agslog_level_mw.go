package agslog

import (
	"ag-core/ag/ag_conf"
	"context"
	"errors"
	"log/slog"
	"strings"

	slogmulti "github.com/samber/slog-multi"
	"golang.org/x/sync/singleflight"
)

/*
 * 日志级别中间件
 */

const (
	LevelConfigPrefix = "aglog.level"
)

var levelKeyMap = map[string]slog.Level{
	"DEBUG": slog.LevelDebug,
	"INFO":  slog.LevelInfo,
	"WARN":  slog.LevelWarn,
	"ERROR": slog.LevelError,
}

type (
	LevelMwConfig map[string]string

	LevelMwConfigHelper struct {
		sfgrp       *singleflight.Group
		levelConfig *LevelMwConfig
	}
)

func NewLevelMwConfig(binder ag_conf.IBinder) (*LevelMwConfig, error) {
	conf := &LevelMwConfig{}
	err := binder.Bind(conf, LevelConfigPrefix)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func NewLevelMwConfigHelper(levelConfig *LevelMwConfig) *LevelMwConfigHelper {
	return &LevelMwConfigHelper{
		sfgrp:       &singleflight.Group{},
		levelConfig: levelConfig,
	}
}

func (lch *LevelMwConfigHelper) Enable(ctx context.Context, level slog.Level, next func(context.Context, slog.Level) bool) bool {
	if lch.levelConfig == nil || len(*lch.levelConfig) == 0 {
		return next(ctx, level)
	}

	startName := ctx.Value(HandlerStartCtxKey{})
	if startName == nil {
		return next(ctx, level)
	}

	if logKey, ok := startName.(string); ok {
		cnflevel, ok := lch.getLevel(logKey)

		// cnflevel, ok := lch.GetLevel(logKey)
		if !ok {
			cnflevel, ok = lch.GetLevel(const_rootLevelKey)
			if !ok {
				return next(ctx, level)
			}
		}

		lv := cnflevel

		enable := level >= lv
		return enable
	}

	return next(ctx, level)
}

func (lch *LevelMwConfigHelper) getLevel(key string) (slog.Level, bool) {
	cnflevel, err, _ := lch.sfgrp.Do(
		key,
		func() (interface{}, error) {
			lv, ok := lch.GetLevel(key)
			if !ok {
				return slog.LevelInfo, errors.New("level not found")
			}
			return lv, nil
		},
	)

	if err != nil {
		return slog.LevelInfo, false
	}

	return cnflevel.(slog.Level), true

}

func (lch *LevelMwConfigHelper) GetLevel(key string) (slog.Level, bool) {
	levelStr := (*lch.levelConfig)[key]
	if levelStr == "" {
		return slog.LevelInfo, false
	}

	levelStrKey := strings.ToUpper(levelStr)
	level, ok := levelKeyMap[levelStrKey]
	if !ok {
		return slog.LevelInfo, false
	}
	return level, true
}

func NewLevelMiddleware(levelHelper *LevelMwConfigHelper) slogmulti.Middleware {
	return slogmulti.NewEnabledInlineMiddleware(
		levelHelper.Enable,
	)
}
