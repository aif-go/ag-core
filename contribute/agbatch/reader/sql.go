// Package reader provides built-in ItemReader implementations for agbatch.
package reader

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	agbatch "github.com/aif-go/ag-core/contribute/agbatch"
)

// RowMapper maps a sql.Rows row to a typed item.
type RowMapper[T any] func(rows *sql.Rows) (T, error)

// --- SqlCursorItemReader ---

// SqlCursorItemReader reads items from a SQL database using a forward-only cursor.
// Corresponds to Spring Batch's JdbcCursorItemReader.
type SqlCursorItemReader[T any] struct {
	db      *sql.DB
	query   string
	args    []any
	mapper  RowMapper[T]
	rows    *sql.Rows
	started bool
	closed  bool
}

// NewSqlCursorItemReader creates a cursor-based SQL reader.
func NewSqlCursorItemReader[T any](db *sql.DB, query string, mapper RowMapper[T], args ...any) *SqlCursorItemReader[T] {
	return &SqlCursorItemReader[T]{db: db, query: query, args: args, mapper: mapper}
}

// Read returns the next item, or io.EOF when exhausted.
func (r *SqlCursorItemReader[T]) Read(ctx context.Context) (T, error) {
	var zero T
	if r.closed {
		return zero, io.EOF
	}
	if !r.started {
		rows, err := r.db.QueryContext(ctx, r.query, r.args...)
		if err != nil {
			return zero, fmt.Errorf("agbatch/reader: cursor query failed: %w", err)
		}
		r.rows = rows
		r.started = true
	}
	if !r.rows.Next() {
		r.Close()
		return zero, io.EOF
	}
	item, err := r.mapper(r.rows)
	if err != nil {
		r.Close()
		return zero, fmt.Errorf("agbatch/reader: row mapper failed: %w", err)
	}
	return item, nil
}

// Close releases the cursor.
func (r *SqlCursorItemReader[T]) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	if r.rows != nil {
		return r.rows.Close()
	}
	return nil
}

var _ agbatch.ItemReader[any] = (*SqlCursorItemReader[any])(nil)

// --- SqlPagingItemReader ---

// SqlPagingItemReader reads items with LIMIT/OFFSET pagination.
// Use {limit} and {offset} placeholders in the query.
type SqlPagingItemReader[T any] struct {
	db       *sql.DB
	query    string
	args     []any
	mapper   RowMapper[T]
	pageSize int
	offset   int
	buffer   []T
	bufIdx   int
	done     bool
}

// NewSqlPagingItemReader creates a paginating SQL reader.
// query must contain {limit} and {offset} placeholders.
func NewSqlPagingItemReader[T any](db *sql.DB, query string, pageSize int, mapper RowMapper[T], args ...any) *SqlPagingItemReader[T] {
	if pageSize <= 0 {
		pageSize = 100
	}
	return &SqlPagingItemReader[T]{db: db, query: query, args: args, mapper: mapper, pageSize: pageSize}
}

// Read returns the next item, fetching a new page when needed.
func (r *SqlPagingItemReader[T]) Read(ctx context.Context) (T, error) {
	var zero T
	if r.done && r.bufIdx >= len(r.buffer) {
		return zero, io.EOF
	}
	if r.bufIdx >= len(r.buffer) {
		page, err := r.fetchPage(ctx)
		if err != nil {
			return zero, err
		}
		if len(page) == 0 {
			r.done = true
			return zero, io.EOF
		}
		r.buffer = page
		r.bufIdx = 0
	}
	item := r.buffer[r.bufIdx]
	r.bufIdx++
	return item, nil
}

func (r *SqlPagingItemReader[T]) fetchPage(ctx context.Context) ([]T, error) {
	q := strings.ReplaceAll(r.query, "{limit}", fmt.Sprintf("%d", r.pageSize))
	q = strings.ReplaceAll(q, "{offset}", fmt.Sprintf("%d", r.offset))
	rows, err := r.db.QueryContext(ctx, q, r.args...)
	if err != nil {
		return nil, fmt.Errorf("agbatch/reader: paging query failed at offset %d: %w", r.offset, err)
	}
	defer rows.Close()
	var items []T
	for rows.Next() {
		item, err := r.mapper(rows)
		if err != nil {
			return nil, fmt.Errorf("agbatch/reader: row mapper failed: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("agbatch/reader: rows error: %w", err)
	}
	r.offset += r.pageSize
	if len(items) < r.pageSize {
		r.done = true
	}
	return items, nil
}

// Reset restarts from the first page.
func (r *SqlPagingItemReader[T]) Reset() {
	r.offset = 0
	r.buffer = nil
	r.bufIdx = 0
	r.done = false
}

var _ agbatch.ItemReader[any] = (*SqlPagingItemReader[any])(nil)
