package gormdb

import (
	"ag-core/ag/ag_db"
	"context"
	"log"
	"time"
)

type TmMiddlewareContext struct {
	tm ag_db.TransactionManager
}

func NewTmMiddlewareContext(tm ag_db.TransactionManager) *TmMiddlewareContext {
	return &TmMiddlewareContext{tm: tm}
}

func (tmCtx *TmMiddlewareContext) TransactionMiddleware(
	method string,
	ctx context.Context,
	req interface{},
	next func(context.Context, interface{}) (interface{}, error),
) (res interface{}, err error) {
	start := time.Now()
	log.Printf("[%s] 准备添加事务", method)
	ctx, txCall := tmCtx.tm.WithTransaction(ctx)
	defer func() {
		txCall(err)
	}()

	err = tmCtx.tm.Transaction(ctx, func(ctx context.Context) error {
		res, err = next(ctx, req)
		return err
	})

	if err != nil {
		log.Printf("[%s] 事务执行失败 %v: %v", method, time.Since(start), err)
	} else {
		log.Printf("[%s] 事务执行成功 %v", method, time.Since(start))
	}
	return
}
