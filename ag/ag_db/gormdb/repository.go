package gormdb

import (
	"ag-core/ag/ag_db"
	"context"
	"database/sql"
	"log/slog"

	"gorm.io/gorm"
)

// func InitializeTM(tm TransactionManager) {
// 	TM = tm
// }

type Repository struct {
	db *gorm.DB
	//tm     TransactionManager // 循环注入
	//logger *log.Logger
	logger *slog.Logger
}

func NewRepository(
	logger *slog.Logger,
	//	logger *log.Logger,
	db *gorm.DB,
	// tm TransactionManager, // 循环注入
) *Repository {

	rep := &Repository{
		db:     db,
		logger: logger,
		//tm:     tm,
	}
	ag_db.TM = rep
	return rep
}

func NewTransactionManager(repository *Repository) ag_db.TransactionManager {
	//TM = repository // 保留一个全局对象，方便事务操作
	return repository
}

func (r *Repository) DB(ctx context.Context) *gorm.DB {
	// 若上下文开启了事务则返回上下文事务
	v := ctx.Value(ag_db.CtxTxKey)
	if v != nil {
		if tx, ok := v.(*gorm.DB); ok {
			r.logger.Info("已开启事务")
			return tx
		}
	}
	r.logger.Info("新事务")
	// 若未开启事务则返回新db，事务行为为默认方式
	return r.db.WithContext(ctx)
}

// 开启事务处理
func (r *Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// TODO 存在当前事务时，处理事务传播逻辑
	// 将传入的业务处理fn包装到gorm的Transaction中处理
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, ag_db.CtxTxKey, tx) // 创建子ctx，置入事务对象，后续数据库操作要使用该事务对象
		r.logger.Info("开启事务")
		return fn(ctx)
	})
}

func (r *Repository) WithTransaction(ctx context.Context, opts ...*sql.TxOptions) (context.Context, func(error)) {
	tx := r.db.Begin(opts...)
	txctx := context.WithValue(ctx, ag_db.CtxTxKey, tx)
	r.logger.Info("开启事务")
	return txctx, func(err error) {

		if err != nil {
			r.logger.Info("事务回滚")
			tx.Rollback()
		} else {
			select {
			case <-ctx.Done():
				r.logger.Info("事务回滚")
				tx.Rollback()
				return
			default:
				r.logger.Info("事务提交")
				tx.Commit()
			}
		}
	}
}
