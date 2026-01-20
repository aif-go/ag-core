package gormdb

import (
	"context"
	"strings"

	agdb "ag-core/contribute/agdb"

	"gorm.io/gorm"
)

type Repository struct {
	agdb.TxContextAbility[*gorm.DB]
	db     *gorm.DB
	DbType string
}

func NewRepository(
	db *gorm.DB,
) *Repository {

	rep := &Repository{
		db: db,
	}

	// ag_db.TM = rep

	// 获取当前驱动对应的数据库类型
	rep.DbType = strings.ToUpper(db.Dialector.Name())
	// 将go dbm官方的驱动名替换为俗名IBM
	switch rep.DbType {
	case "GO_IBM_DB":
		rep.DbType = "DB2"
	}

	return rep
}

func NewTransactionManager(repository *Repository) agdb.TransactionManager {
	//TM = repository // 保留一个全局对象，方便事务操作
	return repository
}

func (r *Repository) DB(ctx context.Context) *gorm.DB {
	// 若上下文开启了事务则返回上下文事务
	v := r.GetTxFromCtx(ctx)
	if v != nil {
		return v
	}
	// 若未开启事务则返回新db，事务行为为默认方式
	return r.db.WithContext(ctx)
}

// 开启事务处理
func (r *Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// 将传入的业务处理fn包装到gorm的Transaction中处理
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = r.BindTxToCtx(ctx, tx)
		return fn(ctx)
	})
}

// func (r *Repository) WithTransaction(ctx context.Context, opts ...*sql.TxOptions) (context.Context, func(error) error) {
// 	tx := r.db.Begin(opts...)
// 	// txctx := context.WithValue(ctx, ag_db.CtxTxKey, tx)
// 	txctx := r.BindTxToCtx(ctx, tx)
// 	r.logger.Info("开启事务")
// 	return txctx, func(err error) error {

// 		if err != nil {
// 			r.logger.Info("事务回滚")
// 			tx.Rollback()
// 			return nil
// 		} else {
// 			select {
// 			case <-ctx.Done():
// 				r.logger.Info("事务回滚")
// 				tx.Rollback()
// 				return nil
// 			default:
// 				r.logger.Info("事务提交")
// 				return tx.Commit().Error
// 			}
// 		}
// 	}
// }
