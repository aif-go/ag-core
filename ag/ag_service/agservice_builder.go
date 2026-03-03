package ag_service

import (
	"sort"
)

// AgServiceBuilder 服务调用链构建器接口
type AgServiceBuilder interface {
	BuildEndpointChain(cinfo CallInfo, middlewares []MiddlewareProvider, actual Endpoint) (Endpoint, error)
}

// NewAgServiceBuilder 创建服务调用链构建器
func NewAgServiceBuilder(ciOpts []CallInfoOpt, globalMWs []MiddlewareProvider, gpMWs []PrioritizedMiddlewareProvider, globalMWFuns []MiddlewareFunc) (AgServiceBuilder, error) {
	mws := globalMWs

	// 合并全局中间件和全局优先中间件
	for _, gpMw := range gpMWs {
		mws = append(mws, gpMw)
	}

	for _, mw := range globalMWFuns {
		mws = append(mws, &SimpleMiddleware{
			Mw: mw,
		})
	}

	// 注入相关增强能力，如对CallInfo的增强扩展
	return &agServiceBuilder{
		callInfoOpts: ciOpts,
		globalMWs:    mws,
	}, nil
}

// agServiceBuilder 服务调用链构建器实现
type agServiceBuilder struct {
	callInfoOpts []CallInfoOpt
	globalMWs    []MiddlewareProvider
}

func (as *agServiceBuilder) BuildEndpointChain(cinfo CallInfo, customMws []MiddlewareProvider, actual Endpoint) (Endpoint, error) {
	// 通过增强，对callInfo进行相关增强，如事务标志
	ci := cinfo
	for _, opt := range as.callInfoOpts {
		opt(&ci)
	}

	mws := append(as.globalMWs, customMws...)

	return buildEndpointChain(ci, mws, actual)
}

func buildEndpointChain(cinfo CallInfo, middlewares []MiddlewareProvider, actual Endpoint) (Endpoint, error) {
	// 转换为PrioritizedMiddleware
	pmws := make([]PrioritizedMiddlewareProvider, 0, len(middlewares)+1)

	// pmws = append(pmws, ServiceInfoCtxPMw(sinfo)) // 添加服务信息上下文绑定中间件
	pmws = append(pmws, callInfoCtxBindMw(cinfo)) // 添加服务信息上下文绑定中间件

	for _, mw := range middlewares {
		if cond, ok := mw.(MiddleWareCondition); ok {
			if !cond.Condition(cinfo) {
				continue
			}
		}
		// 判断mw是否为PrioritizedServerMiddleware
		if _, ok := mw.(PrioritizedMiddlewareProvider); ok {
			pmws = append(pmws, mw.(PrioritizedMiddlewareProvider))
		} else {
			pmws = append(pmws, &SimplePrioritizedMiddleware{
				Order: ServiceInfoMiddlewarePriorityNormal, // 默认优先级为普通优先级
				Mw:    mw.Middleware(),
			})
		}
	}

	// 排序
	sort.Sort(middlewareProviderSorter(pmws))

	// 构建责任链
	endpintChain := actual
	for i := len(pmws) - 1; i >= 0; i-- {
		endpintChain = pmws[i].Middleware()(endpintChain)
	}
	return endpintChain, nil
}
