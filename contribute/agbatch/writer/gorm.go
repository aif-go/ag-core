package writer

import (
	"context"
	"fmt"

	agbatch "github.com/aif-go/ag-core/contribute/agbatch"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormBatchItemWriter writes items using GORM's Create in a transaction.
type GormBatchItemWriter[T any] struct {
	getDB     func(context.Context) *gorm.DB
	batchSize int
}

// NewGormBatchItemWriter creates a GORM batch writer.
// batchSize: 0 = single INSERT, >0 = CreateInBatches(batchSize).
func NewGormBatchItemWriter[T any](getDB func(context.Context) *gorm.DB, batchSize int) *GormBatchItemWriter[T] {
	return &GormBatchItemWriter[T]{getDB: getDB, batchSize: batchSize}
}

// Write persists items within a transaction.
func (w *GormBatchItemWriter[T]) Write(ctx context.Context, items []T) error {
	if len(items) == 0 {
		return nil
	}
	return w.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		if w.batchSize > 0 {
			return tx.CreateInBatches(items, w.batchSize).Error
		}
		return tx.Create(items).Error
	})
}

var _ agbatch.ItemWriter[any] = (*GormBatchItemWriter[any])(nil)

// GormUpsertItemWriter writes items using GORM's Clauses with ON CONFLICT.
type GormUpsertItemWriter[T any] struct {
	getDB       func(context.Context) *gorm.DB
	onConflicts []clause.Expression
}

// NewGormUpsertItemWriter creates a GORM upsert writer.
func NewGormUpsertItemWriter[T any](getDB func(context.Context) *gorm.DB, onConflict ...clause.Expression) *GormUpsertItemWriter[T] {
	return &GormUpsertItemWriter[T]{getDB: getDB, onConflicts: onConflict}
}

// Write upserts items within a transaction.
func (w *GormUpsertItemWriter[T]) Write(ctx context.Context, items []T) error {
	if len(items) == 0 {
		return nil
	}
	return w.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(w.onConflicts...).Create(items).Error
	})
}

var _ agbatch.ItemWriter[any] = (*GormUpsertItemWriter[any])(nil)

// Ensure imports used.
var _ = gorm.Expr
var _ = fmt.Sprintf
