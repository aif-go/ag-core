package writer

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSqlBatchItemWriter_Basic(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: %v", err)
	}
	defer db.Close()
	db.Exec("CREATE TABLE dest (id INTEGER PRIMARY KEY, name TEXT, value INTEGER)")

	writer := NewSqlBatchItemWriter(db,
		"INSERT INTO dest (id, name, value) VALUES (?, ?, ?)",
		func(item struct{ ID int; Name string; Value int }) []any {
			return []any{item.ID, item.Name, item.Value}
		},
	)
	items := []struct{ ID int; Name string; Value int }{
		{1, "a", 10}, {2, "b", 20}, {3, "c", 30},
	}
	if err := writer.Write(context.Background(), items); err != nil {
		t.Fatal(err)
	}
	var count int
	db.QueryRow("SELECT COUNT(*) FROM dest").Scan(&count)
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

func TestSqlBatchItemWriter_Empty(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skipf("skipping: %v", err)
	}
	defer db.Close()
	writer := NewSqlBatchItemWriter(db, "INSERT INTO t VALUES (?)", func(i int) []any { return []any{i} })
	if err := writer.Write(context.Background(), nil); err != nil {
		t.Fatal("empty write should not error")
	}
	_ = fmt.Sprintf
}
