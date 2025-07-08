package client

import (
	"ag-core/ag/ag_error"
	"context"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

func clientErrMidleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp interface{}) (err error) {
		err = next(ctx, req, resp)
		if err != nil {
			return err
		}
		// 提取rpcinfo
		ri := rpcinfo.GetRPCInfo(ctx)
		be := ri.Invocation().BizStatusErr() // BizStatusErr是通过协议层头传递的。TODO 要测试该方式是否能和spring-grpc端兼容
		if be != nil {
			// 如果是业务异常则转换成业务系统内部业务异常
			err = ag_error.NewBizStatusError(be.BizStatusCode(), be.BizMessage(), be.BizExtra())
		}
		return err
	}
}

func NewAgBizErrorMiddlewareOption() *client.Option {
	opt := client.WithMiddleware(clientErrMidleware)
	return &opt
}

/*
func clientErrHandler(ctx context.Context, err error) (rerr error) {
	if err != nil {
		return err
	}
	// 提取rpcinfo
	ri := rpcinfo.GetRPCInfo(ctx)
	be := ri.Invocation().BizStatusErr() // BizStatusErr是通过协议层头传递的。TODO 要测试该方式是否能和spring-grpc端兼容
	if be != nil {
		// 如果是业务异常则转换成业务系统内部业务异常
		rerr = ag_error.NewBizStatusError(be.BizStatusCode(), be.BizMessage(), be.BizExtra())
	}
	return rerr
}

func NewAgBizErrorHandler() *client.Option { // TODO client ErrorHandler不生效，待排查为什么，暂clientErrMidleware实现
	// opt := client.WithMiddleware(clientErrMidleware)
	opt := client.WithErrorHandler(clientErrHandler)
	return &opt
}
*/
