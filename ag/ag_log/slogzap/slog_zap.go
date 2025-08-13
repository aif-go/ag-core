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

// BindSlogZapProperties 绑定slogzap配置
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

// NewSlogHandler4ZapProps 根据zap日志配置创建slogzap handler
func NewSlogHandler4ZapProps(props *SlogZapProperties) ([]slog.Handler, error) {
	if props == nil {
		fmt.Println("NewSlogHandler4ZapProps props is nil")
		return nil, nil // 日志加载异常不影响应用运行状态
	}

	var handlers []slog.Handler
	// 遍历多个配置，每个配置创建一个zaplog
	for k, v := range props.Logs {
		name := k
		// 根据配置初始化zaplog
		zaplog := logzap.NewZapLogP(&v)

		// 创建slogzap封装
		opt := slogzap.Option{
			Level:     slog.LevelDebug, // TODO 应该从配置中读取
			AddSource: true,            // TODO 应该从配置中读取，是否添加caller信息
			// AttrFromContext: []func(ctx context.Context) []slog.Attr{afc}, // 自定义的从context中提取日志属性的函数，是否能从tophandler层级进行相关处理
			Logger: zaplog,
		}
		handler := opt.NewZapHandler()

		// 封装成NamedHandler，便于管理维护
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
