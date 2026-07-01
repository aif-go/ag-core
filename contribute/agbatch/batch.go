// Package agbatch provides a batch processing framework based on github.com/destel/rill,
// aligned with Spring Batch concepts: Job, Step, ItemReader, ItemProcessor, ItemWriter,
// chunk-oriented processing, retry/skip policies, and job repository.
package agbatch

import (
	"context"
	"errors"
	"io"
)

// ErrEndOfStream signals that the reader has no more items.
// Readers should return this error (or io.EOF) when exhausted.
var ErrEndOfStream = errors.New("agbatch: end of stream")

// ItemReader reads items one at a time. Implementations must be safe for
// sequential use; the framework does not call Read concurrently on the same reader.
//
// Return ErrEndOfStream (or io.EOF) to signal no more items. Any other error
// aborts the step.
type ItemReader[T any] interface {
	Read(ctx context.Context) (T, error)
}

// ItemProcessor processes a single item, returning the transformed item or an error.
// The framework may call Process concurrently across different items within a chunk,
// so implementations must be safe for concurrent use when step concurrency > 1.
type ItemProcessor[T, R any] interface {
	Process(ctx context.Context, item T) (R, error)
}

// ItemWriter writes a chunk of items. The framework guarantees that items passed
// to Write are the result of processing a single chunk — the slice length will
// never exceed the configured chunk size.
type ItemWriter[T any] interface {
	Write(ctx context.Context, items []T) error
}

// ReaderFunc adapts a function to the ItemReader interface.
type ReaderFunc[T any] func(ctx context.Context) (T, error)

func (f ReaderFunc[T]) Read(ctx context.Context) (T, error) { return f(ctx) }

// ProcessorFunc adapts a function to the ItemProcessor interface.
type ProcessorFunc[T, R any] func(ctx context.Context, item T) (R, error)

func (f ProcessorFunc[T, R]) Process(ctx context.Context, item T) (R, error) { return f(ctx, item) }

// WriterFunc adapts a function to the ItemWriter interface.
type WriterFunc[T any] func(ctx context.Context, items []T) error

func (f WriterFunc[T]) Write(ctx context.Context, items []T) error { return f(ctx, items) }

// Tasklet is a single-step task within a job. Use for steps that don't fit
// Step is a phase within a job.
type Step interface {
	Name() string
	Execute(ctx context.Context, jobExec *JobExecution) (*StepExecution, error)
}

// Tasklet is a single-step task within a job. Use for steps that don't fit
// the read-process-write pattern (e.g., file cleanup, stored procedure call).
type Tasklet interface {
	Execute(ctx context.Context, exec *StepExecution) error
}

// TaskletFunc adapts a function to the Tasklet interface.
type TaskletFunc func(ctx context.Context, exec *StepExecution) error

func (f TaskletFunc) Execute(ctx context.Context, exec *StepExecution) error { return f(ctx, exec) }

// IsEndOfStream reports whether err signals end of stream (ErrEndOfStream or io.EOF).
func IsEndOfStream(err error) bool {
	return errors.Is(err, ErrEndOfStream) || errors.Is(err, io.EOF)
}
