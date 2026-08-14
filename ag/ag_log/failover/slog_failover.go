package failover

import (
	"fmt"
	"log/slog"

	"github.com/aif-go/ag-core/ag/ag_conf"
	"github.com/aif-go/ag-core/ag/ag_log/agslog"
	slogmulti "github.com/samber/slog-multi"
)

const (
	AgSlogFailoverPropertiesKeyPrefix = "aglog.failover"
)

// AgSlogFailoverProperties failover 日志配置
// Logs: key 为 failover handler 名，value 为按优先级排序的子 handler 名列表
type AgSlogFailoverProperties struct {
	Logs map[string][]string
}

// BindAgSLogFailoverProperties 绑定 failover 配置
func BindAgSLogFailoverProperties(binder ag_conf.IBinder) (*AgSlogFailoverProperties, error) {
	prop := &AgSlogFailoverProperties{}
	err := binder.Bind(prop, AgSlogFailoverPropertiesKeyPrefix)
	if err != nil {
		fmt.Printf("BindAgSLogFailoverProperties err: %v", err)
		return prop, nil
	}
	return prop, nil
}

// NewFailoverHandlerFactorys 为每个 failover 名创建 HandlerFactory
func NewFailoverHandlerFactorys(props *AgSlogFailoverProperties) ([]*agslog.HandlerFactory, error) {
	factories := make([]*agslog.HandlerFactory, 0)
	if props == nil || props.Logs == nil {
		return factories, nil
	}
	for name, handlers := range props.Logs {
		handlerscopy := handlers
		factory := agslog.NewHandlerFactory(name, getDoGetHandlerFunc(handlerscopy))
		factories = append(factories, factory)
	}
	return factories, nil
}

func getDoGetHandlerFunc(failoverHandlerNames []string) func(getHandler func(string) (slog.Handler, error)) (slog.Handler, error) {
	return func(getHandler func(string) (slog.Handler, error)) (slog.Handler, error) {
		subHandlers := make([]slog.Handler, 0, len(failoverHandlerNames))
		for _, handlerName := range failoverHandlerNames {
			subhandler, err := getHandler(handlerName)
			if err != nil || subhandler == nil {
				fmt.Printf("agslog: failover handler %s not found, will ignore and continue, err: %v", handlerName, err)
				continue
			}
			subHandlers = append(subHandlers, subhandler)
		}

		if len(subHandlers) == 0 {
			return nil, fmt.Errorf("agslog: failover handler %v not found", failoverHandlerNames)
		}

		return slogmulti.Failover()(subHandlers...), nil
	}
}
