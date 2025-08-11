package slogzap

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_log/agslog"
	"ag-core/ag/ag_log/logzap"
	"fmt"
	"log/slog"

	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
)

const (
	SlogZapPropertiesKeyPrefix = "aglog.zap"
)

type SlogZapProperties struct {
	Logs map[string]logzap.ZlogProperties
}

type SlogZapOption struct {
	ZapLogs         *zap.Logger
	AttrFromContext []agslog.SlogAttrFromContext
}

func BindSlogZapProperties(binder ag_conf.IBinder) (*SlogZapProperties, error) {
	prop := &SlogZapProperties{}
	err := binder.Bind(prop, SlogZapPropertiesKeyPrefix)
	if err != nil {
		fmt.Println("BindSlogZapProperties err:", err)
		// return nil, err
		return nil, nil // 日志配置加载问题，不中断应用
	}

	return prop, nil
}

func NewSlogHandler4ZapProps(props *SlogZapProperties) ([]slog.Handler, error) {
	if props == nil {
		fmt.Println("NewSlogHandler4ZapProps props is nil")
		return nil, nil // 日志加载异常不影响应用运行状态
	}

	var handlers []slog.Handler
	for k, v := range props.Logs {
		name := k
		zaplog := logzap.NewZapLogP(&v)
		opt := slogzap.Option{
			Level:     slog.LevelDebug,
			Logger:    zaplog,
			AddSource: true,
			// AttrFromContext: []func(ctx context.Context) []slog.Attr{afc},
		}
		handler := opt.NewZapHandler()
		nhandler := agslog.NewNamedHandler(name, handler)
		handlers = append(handlers, nhandler)
	}

	return handlers, nil
}

func NewSlog4Zap(log *zap.Logger) (slog.Handler, error) {
	opt := slogzap.Option{
		Level:     slog.LevelDebug,
		Logger:    log,
		AddSource: true,
	}
	handler := opt.NewZapHandler()
	nhandler := agslog.NewNamedHandler("slog4zap", handler)

	return nhandler, nil
}
