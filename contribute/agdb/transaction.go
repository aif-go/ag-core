package ag_db

import (
	"context"
	"database/sql"
)

const CtxTxKey = "AicTxKey"

// TM TransactionManager
var TM TransactionManager // TODO TransactionManager

type TransactionManager interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	WithTransaction(ctx context.Context, opts ...*sql.TxOptions) (context.Context, func(error))
}
