package async

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_log/agslog"
	"fmt"
	"log/slog"
)

// BindAsyncConfig 绑定异步配置
func BindAsyncLogConfig(binder ag_conf.IBinder) (*AsyncGlobalProperties, error) {
	conf := &AsyncGlobalProperties{}
	err := binder.Bind(conf, AglogAsyncKeyPrefix)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// BuildAsyncHandlerFactorys 构建异步 handler工厂
func BuildAsyncHandlerFactorys(conf *AsyncGlobalProperties) ([]*agslog.HandlerFactory, error) {
	hfacs := []*agslog.HandlerFactory{}

	if conf == nil || conf.Logs == nil || len(conf.Logs) == 0 {
		return hfacs, nil
	}

	for name, logConfig := range conf.Logs {
		cname := name
		gname := logConfig.Group
		logname := logConfig.Log

		gconfig, ok := conf.Groups[gname]
		if !ok {
			slog.Warn("async log config group not found", "group", gname)
			gname = "default"
			gconfig = defaultAsyncGroupConfig
		}

		// wg := GetWorkerGroup("default", gconfig)

		factory := agslog.NewHandlerFactory(
			cname,
			func(geth func(string) (slog.Handler, error)) (slog.Handler, error) {
				oriHandler, err := geth(logname)
				if err != nil {
					return nil, err
				}

				if oriHandler == nil {
					return nil, fmt.Errorf("original handler not found for %s", logname)
				}
				return NewAsyncHandler(oriHandler, gname, &gconfig), nil
			},
		)

		hfacs = append(hfacs, factory)
	}

	return hfacs, nil
}
