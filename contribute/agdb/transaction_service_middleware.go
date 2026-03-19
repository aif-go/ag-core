package agdb

import (
	ag_service "ag-core/ag/ag_service"
	"context"
	"fmt"
)

const (
	TransactionTag = "transaction"
)

type transactionPropagationKey struct{}

type TransactionMiddlewareProvider struct {
	tm TransactionManager
}

func TransactionPreOpt(callInfo *ag_service.CallInfo) error {
	ttag := callInfo.GetTag(TransactionTag)
	if ttag == nil {
		return nil
	}
	var tp TransactionPropagation
	switch v := ttag.(type) {
	case bool:
		if !v {
			return nil
		} else {
			tp = TRANSACTION_PROPAGATION_REQUIRED // 默认 REQUIRED 模式
		}
	case TransactionPropagation:
		tp = v
	default:
		return nil
	}
	if !IsSupportsTransaction(tp) {
		return fmt.Errorf("transaction propagation %v is not supported", tp)
	}

	err := callInfo.AddTag(transactionPropagationKey{}, tp)
	if err != nil {
		return err
	}
	return nil
}

// NewTransactionMiddlewareProvider creates and returns a new TransactionMiddlewareProvider instance.
// func NewTransactionMiddlewareProvider(tm TransactionManager) *TransactionMiddlewareProvider {
func NewTransactionMiddlewareProvider(tm TransactionManager) ag_service.MiddlewareProvider {
	return &TransactionMiddlewareProvider{tm: tm}
}

func (p *TransactionMiddlewareProvider) Condition(callInfo *ag_service.CallInfo) bool {
	ok := callInfo.HasTag(transactionPropagationKey{})
	return ok
}

func (p *TransactionMiddlewareProvider) Middleware() ag_service.MiddlewareFunc {
	return func(next ag_service.Endpoint) ag_service.Endpoint {
		return func(ctx context.Context, req interface{}) (resp interface{}, rerr error) {
			// 获取调用信息注册的事务传播方式
			cinfo := ag_service.GetCallInfoFromContext(ctx)
			// tp := TransactionRegistry.GetTransactionFlagForCall(cinfo)
			tp, ok := cinfo.GetTag(transactionPropagationKey{}).(TransactionPropagation)
			if !ok {
				tp = TRANSACTION_PROPAGATION_UNKNOWN
			}

			// 根据事务的传播方式执行事务的行为
			terr := WithTransaction(ctx, p.tm, tp, func(ctx context.Context) error {
				resp, rerr = next(ctx, req)
				return rerr
			})

			if terr != nil {
				return resp, terr
			}
			return

		}
	}
}

// // RegisterTransactionService 注册需要开启事务的服务，全局的事务注册入口
// var TransactionRegistry transactionRegistry = transactionRegistry{
// 	flags: make(map[string]TransactionPropagation),
// }

// type transactionRegistry struct {
// 	flags map[string]TransactionPropagation
// 	mutex sync.RWMutex
// }

// // registerTransactionFlag 通过key注册事务传播行为,默认值为TransactionPropagationRequired
// func (tr *transactionRegistry) registerTransactionFlag(tpkey string, tp ...TransactionPropagation) {
// 	tr.mutex.Lock()
// 	defer tr.mutex.Unlock()

// 	if tr.flags == nil {
// 		tr.flags = make(map[string]TransactionPropagation)
// 	}

// 	key := tpkey
// 	ltp := TRANSACTION_PROPAGATION_REQUIRED // 默认 REQUIRED 模式
// 	if len(tp) > 0 {
// 		ltp = tp[0]
// 	}

// 	tr.flags[key] = ltp
// }

// // getTransactionFlag 获取事务传播行为
// func (tr *transactionRegistry) getTransactionFlag(tpkey string) TransactionPropagation {
// 	tr.mutex.RLock() // 只读锁, 并发安全 TODO 评估读锁的性能影响情况
// 	defer tr.mutex.RUnlock()
// 	if tp, ok := tr.flags[tpkey]; ok {
// 		return tp
// 	}
// 	return TRANSACTION_PROPAGATION_UNKNOWN // 未知: 未指定事务传播行为
// }

// // RegTxForCall 通过callInfo注册事务传播行为
// func (tr *transactionRegistry) RegTxForCall(callInfo *ag_service.CallInfo, tp ...TransactionPropagation) {
// 	key := GetTransactionFlagKey(callInfo)
// 	tr.registerTransactionFlag(key, tp...)
// }

// // GetTransactionFlag 获取事务传播行为
// func (tr *transactionRegistry) GetTransactionFlagForCall(callInfo *ag_service.CallInfo) TransactionPropagation {
// 	key := GetTransactionFlagKey(callInfo)
// 	return tr.getTransactionFlag(key)
// }

// // GetTransactionFlagKey 获取事务传播行为的key
// func GetTransactionFlagKey(callInfo *ag_service.CallInfo) string {
// 	return fmt.Sprintf("%s.%s:%s", callInfo.ServiceInfo().PackageName(), callInfo.ServiceInfo().ServiceName(), callInfo.CallName())
// }
