package fanout

import (
	"ag-core/ag/ag_log/agslog"
	"fmt"
	"log/slog"

	slogmulti "github.com/samber/slog-multi"
)

const (
	AgSlogFanoutPropertiesKeyPrefix = "aglog.fanout"
)

// AgSlogFanoutProperties 日志分发给多个handler
type AgSlogFanoutProperties struct {
	FanoutHandler map[string][]string
}

func NewFanoutHandlerFactorys(props *AgSlogFanoutProperties) ([]*agslog.HandlerFactory, error) {
	factories := make([]*agslog.HandlerFactory, 0)
	for name, handlers := range props.FanoutHandler {
		// 创建fanout handler工厂
		// 创建局部变量副本
		handlerscopy := handlers
		factory := agslog.NewHandlerFactory(
			name,
			getDoGetHandlerFunc(handlerscopy),
		)
		factories = append(factories, factory)
	}
	return factories, nil
}

func getDoGetHandlerFunc(
	fanoutHandlerNames []string,
) func(getHandler func(handlerName string) (slog.Handler, error)) (slog.Handler, error) {
	return func(getHandler func(handlerName string) (slog.Handler, error)) (slog.Handler, error) {
		subHandlers := make([]slog.Handler, 0)
		for _, handlerName := range fanoutHandlerNames {
			// 根据handlername获取handler
			subhandler, err := getHandler(handlerName)
			if err != nil {
				return nil, err
			}

			subHandlers = append(subHandlers, subhandler)
		}

		if len(subHandlers) == 0 {
			return nil, fmt.Errorf("agslog: fanout handler %s not found", fanoutHandlerNames)
		}

		fanoutHandler := slogmulti.Fanout(subHandlers...)

		return fanoutHandler, nil
	}
}
