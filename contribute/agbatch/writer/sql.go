// Package writer provides built-in ItemWriter implementations for agbatch.
package writer

import (
	"context"
	"database/sql"
	"fmt"

	agbatch "github.com/aif-go/ag-core/contribute/agbatch"
)

// SqlBatchItemWriter writes items using a SQL statement within a transaction.
// Corresponds to Spring Batch's JdbcBatchItemWriter.
type SqlBatchItemWriter[T any] struct {
	db       *sql.DB
	query    string
	preparer func(item T) []any
}

// NewSqlBatchItemWriter creates a batch SQL writer.
func NewSqlBatchItemWriter[T any](db *sql.DB, query string, preparer func(item T) []any) *SqlBatchItemWriter[T] {
	return &SqlBatchItemWriter[T]{db: db, query: query, preparer: preparer}
}

// Write persists items within a transaction.
func (w *SqlBatchItemWriter[T]) Write(ctx context.Context, items []T) error {
	if len(items) == 0 {
		return nil
	}
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("agbatch/writer: begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx, w.query)
	if err != nil {
		return fmt.Errorf("agbatch/writer: prepare: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.ExecContext(ctx, w.preparer(item)...); err != nil {
			return fmt.Errorf("agbatch/writer: exec: %w", err)
		}
	}
	return tx.Commit()
}

var _ agbatch.ItemWriter[any] = (*SqlBatchItemWriter[any])(nil)
