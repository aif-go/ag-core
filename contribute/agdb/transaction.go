package agdb

import (
	"context"
)

// const CtxTxKey = "AicTxKey"

type ctxTxKey struct{}

// TM TransactionManager
// var tm TransactionManager // TODO TransactionManager

// func GetTransactionManager() TransactionManager {
// 	return tm
// }
// func setTransactionManager(rtm TransactionManager) {
// 	tm = rtm
// }

// TransactionPropagation 事务传播行为
type TransactionPropagation int

const (
	TRANSACTION_PROPAGATION_UNKNOWN  TransactionPropagation = iota // 未知: 未指定事务传播行为
	TRANSACTION_PROPAGATION_SUPPORTS                               // 支持: 有事务则加入,无事务则非事务方式执行
	TRANSACTION_PROPAGATION_REQUIRED                               // 必须: 有事务则加入,无事务则开启新事务
	// TRANSACTION_PROPAGATION_MANDATORY                                   // 强制: 有事务则加入,无事务则报错
	// TRANSACTION_PROPAGATION_REQUIRED_NEW                                // 隔离：每次开启新事务,与其他事务隔离(挂起当前事务)
	// TRANSACTION_PROPAGATION_NOT_SUPPORTED                               // 不支持: 非事务方式执行,如果有当前事务则挂起当前事务
	// TRANSACTION_PROPAGATION_NEVER                                       // 非事务: 不开启事务,如果有当前事务则报错
	// TRANSACTION_PROPAGATION_NESTED                                      // 嵌套事务: 没有事务则新建事务，有事务则嵌套事务内执行
)

func IsSupportsTransaction(tp TransactionPropagation) bool {
	return tp == TRANSACTION_PROPAGATION_SUPPORTS || tp == TRANSACTION_PROPAGATION_REQUIRED
}

type TransactionManager interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	// TransactionWithTP(ctx context.Context, tp TransactionPropagation, fn func(ctx context.Context) error) error
	// WithTransaction(ctx context.Context, opts ...*sql.TxOptions) (context.Context, func(error) error)
}

// TxContextAbility 事务上下文能力,TransactionManager实现者应该继承该能力，并通过该能力
type TxContextAbility[T any] struct{}

// BindTxToCtx 将事务绑定到上下文
func (a *TxContextAbility[T]) BindTxToCtx(ctx context.Context, tx T) context.Context {
	// tx为any，是因为不同的db框架，其tx类型不同，此处无法确定tx的类型
	// TODO tx的类型不确定性存在一定风险，但此能力是由TransactionManager实现者继承与使用的，无业务开发使用场景，从框架实现上控制使用风险
	rctx := context.WithValue(ctx, ctxTxKey{}, tx)
	return rctx
}

// GetTxFromCtx 从上下文获取当前事务
func (a *TxContextAbility[T]) GetTxFromCtx(ctx context.Context) T {

	tx := ctx.Value(ctxTxKey{})
	if tx == nil {
		ntx := *new(T)
		return ntx
	}
	return tx.(T)
}

// WithTransaction 执行事务操作
func WithTransaction(
	ctx context.Context,
	tm TransactionManager,
	tp TransactionPropagation,
	fn func(ctx context.Context) error,
	// opts ...*sql.TxOptions, // FIXME 暂不支持事务选项，有需要再设计
) error {
	switch tp {
	case TRANSACTION_PROPAGATION_UNKNOWN:
		return fn(ctx)
	case TRANSACTION_PROPAGATION_SUPPORTS:
		return fn(ctx)
	case TRANSACTION_PROPAGATION_REQUIRED:
		return doTransactionWithRequired(ctx, tm, fn)
	default:
		return fn(ctx) // 默认同未知操作
	}
}

// HasCurrentTx 判断上下文是否存在当前事务
func HasCurrentTx(ctx context.Context) bool {
	return ctx.Value(ctxTxKey{}) != nil
}

func doTransactionWithRequired(ctx context.Context, tm TransactionManager, fn func(ctx context.Context) error) error {
	if HasCurrentTx(ctx) { // 存在当前事务
		return fn(ctx)
	} else { // 不存在当前事务
		terr := tm.Transaction(ctx, func(ctx context.Context) error {
			return fn(ctx)
		})
		return terr
	}
}
