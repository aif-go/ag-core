package ag_service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sync"
)

var (
	ErrCallInfoLocked = errors.New("call info is locked")
	ErrTagKeyExist    = errors.New("tag key exist")
	ErrExtraKeyExist  = errors.New("extra key exist")
)

type (
	// agServiceInfoKey     struct{}
	agServiceCallInfoKey struct{}
)

// ServiceInfo 服务级别信息
type ServiceInfo struct {
	packageName string
	serviceName string
	handlerType interface{}
	// Extra       map[string]interface{}
}

func NewServiceInfo(packageName, serviceName string, handlerType interface{}) *ServiceInfo {
	return &ServiceInfo{
		packageName: packageName,
		serviceName: serviceName,
		handlerType: handlerType,
	}
}

// CallInfo 调用级别信息
type CallInfo struct {
	serviceInfo     *ServiceInfo
	callName        string
	clientStreaming bool
	serverStreaming bool
	Extra           map[string]interface{}
	tag             sync.Map
	// tag             map[interface{}]interface{}

	locked bool
	mu     sync.RWMutex
}

func NewCallInfo(serviceInfo *ServiceInfo, callName string, clientStreaming bool, serverStreaming bool) *CallInfo {
	return &CallInfo{
		serviceInfo:     serviceInfo,
		callName:        callName,
		clientStreaming: clientStreaming,
		serverStreaming: serverStreaming,
		Extra:           make(map[string]interface{}),
		tag:             sync.Map{},
	}
}

func (ci *CallInfo) AddTag(key interface{}, value interface{}) error {
	if ci.locked {
		slog.Warn("CallInfo is locked, cannot add tag")
		return ErrCallInfoLocked
	}

	if ci.HasTag(key) {
		return errors.Join(ErrTagKeyExist, fmt.Errorf("tag key %v exist", key))
	}

	ci.tag.Store(key, value)
	return nil
}

func (ci *CallInfo) GetTag(key interface{}) interface{} {
	if v, ok := ci.tag.Load(key); ok {
		return v
	}
	return nil
}

func (ci *CallInfo) HasTag(key interface{}) bool {
	_, ok := ci.tag.Load(key)
	return ok
}

type CallInfoOpt func(cinfo *CallInfo) error

// callInfoCtxBindMw 调用信息上下文绑定中间件
func callInfoCtxBindMw(cinfo *CallInfo) PrioritizedMiddlewareProvider {
	pmw := &SimplePrioritizedMiddleware{
		// Order: ServiceInfoMiddlewarePriorityNormal,
		Order: math.MinInt, // 最高级别,int最小值
		// Mw:    newServiceInfoCtxBinderMw(sinfo),
		Mw: func(next Endpoint) Endpoint {
			return func(ctx context.Context, req interface{}) (interface{}, error) {

				// 绑定服务信息到上下文
				ctx = context.WithValue(ctx, agServiceCallInfoKey{}, cinfo)

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
func GetCallInfoFromContext(ctx context.Context) *CallInfo {
	if rmd, ok := ctx.Value(agServiceCallInfoKey{}).(*CallInfo); ok {
		return rmd
	}
	// return &CallInfo{}
	return nil
}

func (ci *CallInfo) lock() {
	ci.mu.Lock()
	ci.locked = true
	ci.mu.Unlock()
}

// func (ci *CallInfo) unlock() {
// 	ci.mu.Lock()
// 	ci.locked = false
// 	ci.mu.Unlock()
// }

func (ci *CallInfo) ServiceInfo() *ServiceInfo {
	return ci.serviceInfo
}

func (ci *CallInfo) CallName() string {
	return ci.callName
}

func (ci *CallInfo) IsClientStreaming() bool {
	return ci.clientStreaming
}

func (ci *CallInfo) IsServerStreaming() bool {
	return ci.serverStreaming
}

func (ci *CallInfo) AddExtra(key string, value interface{}) error {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	_, ok := ci.Extra[key]
	if ok {
		return ErrExtraKeyExist
	}

	ci.Extra[key] = value
	return nil
}

func (ci *CallInfo) GetExtra(key string) interface{} {
	ci.mu.RLock()
	defer ci.mu.RUnlock()
	return ci.Extra[key]
}

func (si *ServiceInfo) PackageName() string {
	return si.packageName
}

func (si *ServiceInfo) ServiceName() string {
	return si.serviceName
}

func (si *ServiceInfo) HandlerType() interface{} {
	return si.handlerType
}
