package reader

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: cannot open sqlite: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE test_items (id INTEGER PRIMARY KEY, name TEXT, value INTEGER)`)
	if err != nil {
		t.Fatal(err)
	}
	for i := range 20 {
		db.Exec("INSERT INTO test_items VALUES (?, ?, ?)", i+1, fmt.Sprintf("item-%d", i+1), (i+1)*10)
	}
	return db
}

type testItem struct {
	ID    int
	Name  string
	Value int
}

func TestSqlCursorItemReader_Basic(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	reader := NewSqlCursorItemReader(db,
		"SELECT id, name, value FROM test_items ORDER BY id",
		func(rows *sql.Rows) (testItem, error) {
			var item testItem
			return item, rows.Scan(&item.ID, &item.Name, &item.Value)
		},
	)
	defer reader.Close()

	ctx := context.Background()
	count := 0
	for {
		item, err := reader.Read(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read error: %v", err)
		}
		count++
		if item.ID != count || item.Value != count*10 {
			t.Errorf("item %d: wrong id=%d value=%d", count, item.ID, item.Value)
		}
	}
	if count != 20 {
		t.Errorf("expected 20 items, got %d", count)
	}
}

func TestSqlCursorItemReader_Empty(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	reader := NewSqlCursorItemReader(db, "SELECT id FROM test_items WHERE id < 0",
		func(rows *sql.Rows) (int, error) { var x int; return x, rows.Scan(&x) })
	defer reader.Close()
	_, err := reader.Read(context.Background())
	if err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestSqlPagingItemReader_Basic(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	reader := NewSqlPagingItemReader(db,
		"SELECT id, name, value FROM test_items ORDER BY id LIMIT {limit} OFFSET {offset}",
		7,
		func(rows *sql.Rows) (testItem, error) {
			var item testItem
			return item, rows.Scan(&item.ID, &item.Name, &item.Value)
		},
	)
	ctx := context.Background()
	count := 0
	for {
		_, err := reader.Read(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 20 {
		t.Errorf("expected 20, got %d", count)
	}
}

func TestSqlPagingItemReader_Reset(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	reader := NewSqlPagingItemReader(db,
		"SELECT id FROM test_items ORDER BY id LIMIT {limit} OFFSET {offset}",
		3,
		func(rows *sql.Rows) (int, error) { var x int; return x, rows.Scan(&x) },
	)
	ctx := context.Background()
	for {
		_, err := reader.Read(ctx)
		if err == io.EOF {
			break
		}
	}
	reader.Reset()
	count := 0
	for {
		_, err := reader.Read(ctx)
		if err == io.EOF {
			break
		}
		count++
	}
	if count != 20 {
		t.Errorf("after reset, expected 20, got %d", count)
	}
}
