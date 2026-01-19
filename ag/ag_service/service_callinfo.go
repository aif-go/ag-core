package ag_service

import (
	"context"
	"math"
)

type (
	agServiceInfoKey     struct{}
	agServiceCallInfoKey struct{}
)

// ServiceInfo 服务级别信息
type ServiceInfo struct {
	PackageName string
	ServiceName string
	HandlerType interface{}
	// Extra       map[string]interface{}
}

// CallInfo 调用级别信息
type CallInfo struct {
	ServiceInfo
	CallName        string
	ClientStreaming bool
	ServerStreaming bool
	Extra           map[string]interface{}
}

func (ci *CallInfo) Clone() CallInfo {
	ci2 := *ci
	if ci.Extra != nil {
		ci2.Extra = make(map[string]interface{}, len(ci.Extra))
		for k, v := range ci.Extra {
			ci2.Extra[k] = v
		}
	}
	return ci2
}

type CallInfoOpt func(cinfo *CallInfo)

// callInfoCtxBindMw 调用信息上下文绑定中间件
func callInfoCtxBindMw(cinfo CallInfo) PrioritizedMiddlewareProvider {
	pmw := &SimplePrioritizedMiddleware{
		// Order: ServiceInfoMiddlewarePriorityNormal,
		Order: math.MinInt, // 最高级别,int最小值
		// Mw:    newServiceInfoCtxBinderMw(sinfo),
		Mw: func(next Endpoint) Endpoint {
			return func(ctx context.Context, req interface{}) (interface{}, error) {

				ci := cinfo.Clone()
				// 绑定服务信息到上下文
				ctx = context.WithValue(ctx, agServiceCallInfoKey{}, ci)

				// TEST 模拟从ctx中获取cinfo并添加信息
				// ci2 := GetCallInfoFromContext(ctx)
				// ci2.Extra["hzw"] = rand.Intn(1000) // 随机数 TODO 修改原始，原始map
				// slog.Error("===TEST===, call chain hzw: %d", ci2.Extra["hzw"])
				// ci2.CallName = "xxxxxxx"

				return next(ctx, req)
			}
		},
	}
	return pmw
}

// GetMdFromContext 从上下文中提取元数据
func GetCallInfoFromContext(ctx context.Context) CallInfo {
	if rmd, ok := ctx.Value(agServiceCallInfoKey{}).(CallInfo); ok {
		return rmd
	}
	return CallInfo{}
}
