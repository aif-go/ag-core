package agbatch

import (
	"context"
	"errors"
	"io"
	"testing"
)

func TestReaderFunc(t *testing.T) {
	vals := []int{1, 2, 3}
	i := 0
	r := ReaderFunc[int](func(ctx context.Context) (int, error) {
		if i >= len(vals) {
			return 0, io.EOF
		}
		v := vals[i]
		i++
		return v, nil
	})

	ctx := context.Background()
	for _, want := range vals {
		got, err := r.Read(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	}
	_, err := r.Read(ctx)
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestProcessorFunc(t *testing.T) {
	p := ProcessorFunc[int, int](func(_ context.Context, item int) (int, error) {
		return item * 2, nil
	})
	got, err := p.Process(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if got != 10 {
		t.Errorf("got %d, want 10", got)
	}
}

func TestWriterFunc(t *testing.T) {
	var written []int
	w := WriterFunc[int](func(_ context.Context, items []int) error {
		written = append(written, items...)
		return nil
	})
	err := w.Write(context.Background(), []int{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 3 {
		t.Errorf("got %d items, want 3", len(written))
	}
}

func TestEndOfStream(t *testing.T) {
	if !endOfStream(ErrEndOfStream) {
		t.Error("ErrEndOfStream should be end of stream")
	}
	if !endOfStream(io.EOF) {
		t.Error("io.EOF should be end of stream")
	}
	if endOfStream(errors.New("other")) {
		t.Error("other errors should not be end of stream")
	}
}

func TestTaskletFunc(t *testing.T) {
	called := false
	tl := TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		called = true
		return nil
	})
	err := tl.Execute(context.Background(), NewStepExecution(1, "test"))
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("tasklet was not called")
	}
}
