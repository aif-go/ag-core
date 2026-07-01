package reader

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	agbatch "github.com/aif-go/ag-core/contribute/agbatch"
	"gorm.io/gorm"
)

// --- GormCursorItemReader ---

// GormCursorItemReader reads items from a GORM query using a forward-only cursor.
type GormCursorItemReader[T any] struct {
	getDB   func(context.Context) *gorm.DB
	query   func(*gorm.DB) *gorm.DB
	rows    *sql.Rows
	started bool
	closed  bool
}

// NewGormCursorItemReader creates a GORM cursor reader.
// getDB: use agbatch.DBGetter(db) for plain GORM, or repo.DB for gormdb.Repository.
func NewGormCursorItemReader[T any](getDB func(context.Context) *gorm.DB, query func(*gorm.DB) *gorm.DB) *GormCursorItemReader[T] {
	return &GormCursorItemReader[T]{getDB: getDB, query: query}
}

// Read returns the next item, or io.EOF.
func (r *GormCursorItemReader[T]) Read(ctx context.Context) (T, error) {
	var zero T
	if r.closed {
		return zero, io.EOF
	}
	if !r.started {
		rows, err := r.query(r.getDB(ctx)).Rows()
		if err != nil {
			return zero, fmt.Errorf("agbatch/reader: gorm cursor query failed: %w", err)
		}
		r.rows = rows
		r.started = true
	}
	if !r.rows.Next() {
		r.Close()
		return zero, io.EOF
	}
	var item T
	if err := r.getDB(context.Background()).ScanRows(r.rows, &item); err != nil {
		r.Close()
		return zero, fmt.Errorf("agbatch/reader: gorm scan rows: %w", err)
	}
	return item, nil
}

// Close releases the cursor.
func (r *GormCursorItemReader[T]) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	if r.rows != nil {
		return r.rows.Close()
	}
	return nil
}

var _ agbatch.ItemReader[any] = (*GormCursorItemReader[any])(nil)

// --- GormPagingItemReader ---

// GormPagingItemReader reads items with GORM Offset/Limit pagination.
type GormPagingItemReader[T any] struct {
	getDB    func(context.Context) *gorm.DB
	query    func(*gorm.DB) *gorm.DB
	pageSize int
	page     int
	buffer   []T
	bufIdx   int
	done     bool
}

// NewGormPagingItemReader creates a GORM paginating reader.
func NewGormPagingItemReader[T any](getDB func(context.Context) *gorm.DB, pageSize int, query func(*gorm.DB) *gorm.DB) *GormPagingItemReader[T] {
	if pageSize <= 0 {
		pageSize = 100
	}
	return &GormPagingItemReader[T]{getDB: getDB, query: query, pageSize: pageSize}
}

// Read returns the next item, fetching a new page when needed.
func (r *GormPagingItemReader[T]) Read(ctx context.Context) (T, error) {
	var zero T
	if r.done && r.bufIdx >= len(r.buffer) {
		return zero, io.EOF
	}
	if r.bufIdx >= len(r.buffer) {
		var page []T
		if err := r.query(r.getDB(ctx)).
			Offset(r.page * r.pageSize).Limit(r.pageSize).
			Find(&page).Error; err != nil {
			return zero, fmt.Errorf("agbatch/reader: gorm paging query: %w", err)
		}
		if len(page) == 0 {
			r.done = true
			return zero, io.EOF
		}
		r.buffer = page
		r.bufIdx = 0
		r.page++
		if len(page) < r.pageSize {
			r.done = true
		}
	}
	item := r.buffer[r.bufIdx]
	r.bufIdx++
	return item, nil
}

// Reset restarts from the first page.
func (r *GormPagingItemReader[T]) Reset() {
	r.page = 0
	r.buffer = nil
	r.bufIdx = 0
	r.done = false
}

var _ agbatch.ItemReader[any] = (*GormPagingItemReader[any])(nil)
