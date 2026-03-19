package ag_service

import "sync"

type AgServiceProxyBase struct {
	ServiceInfo *ServiceInfo
	CallInfos   map[string]*CallInfo

	mu        sync.RWMutex // 保护endpoints并发访问
	endpoints map[string]Endpoint
}

// func DefaultAgServiceProxyBase() AgServiceProxyBase {
// 	return AgServiceProxyBase{
// 		MethodInfos: make(map[string]CallInfo),
// 		endpoints:   make(map[string]Endpoint),
// 	}
// }

func NewAgServiceProxyBase(serviceInfo *ServiceInfo, callInfos map[string]*CallInfo) AgServiceProxyBase {
	return AgServiceProxyBase{
		ServiceInfo: serviceInfo,
		CallInfos:   callInfos,
		endpoints:   make(map[string]Endpoint),
	}
}

// registerEndpoint 注册单个端点
func (p *AgServiceProxyBase) RegisterEndpoint(callName string, endpoint Endpoint) {
	p.mu.Lock()
	p.endpoints[callName] = endpoint
	p.mu.Unlock()
}

func (p *AgServiceProxyBase) GetEndpoint(callName string) Endpoint {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.endpoints[callName]
}
