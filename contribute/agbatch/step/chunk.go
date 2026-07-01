// Package step provides built-in Step implementations for agbatch.
package step

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	agbatch "github.com/aif-go/ag-core/contribute/agbatch"
	"github.com/destel/rill"
)

// ChunkStepConfig holds configuration for a chunk-oriented step.
type ChunkStepConfig struct {
	Name              string
	Reader            agbatch.ItemReader[any]
	Processor         agbatch.ItemProcessor[any, any]
	Writer            agbatch.ItemWriter[any]
	ChunkSize         int
	ProcessorPoolSize int
	Retry             agbatch.RetryPolicy
	Skip              agbatch.SkipPolicy
	SkipOnRead        agbatch.SkipPolicy
	ChunkTimeout      time.Duration
	Listener          agbatch.StepListener
	ChunkListener     agbatch.ChunkListener
}

// ChunkStep is a step that processes items in chunks using read-process-write.
type ChunkStep struct{ cfg ChunkStepConfig }

// NewChunkStep creates a chunk-oriented step.
func NewChunkStep(cfg ChunkStepConfig) *ChunkStep {
	if cfg.ChunkSize <= 0 { cfg.ChunkSize = 100 }
	if cfg.ProcessorPoolSize <= 0 { cfg.ProcessorPoolSize = 1 }
	return &ChunkStep{cfg: cfg}
}

// Name returns the step name.
func (s *ChunkStep) Name() string { return s.cfg.Name }

// Execute runs the chunk step.
func (s *ChunkStep) Execute(ctx context.Context, jobExec *agbatch.JobExecution) (*agbatch.StepExecution, error) {
	stepExec := agbatch.NewStepExecution(0, s.cfg.Name)
	stepExec.Status = agbatch.StatusStarted

	if s.cfg.Listener != nil {
		if err := s.cfg.Listener.BeforeStep(ctx, stepExec); err != nil {
			stepExec.Status = agbatch.StatusFailed
			return stepExec, fmt.Errorf("agbatch/step: before step hook: %w", err)
		}
	}

	items := rill.Generate(func(send func(any), sendErr func(error)) {
		for {
			select {
			case <-ctx.Done(): sendErr(ctx.Err()); return
			default:
			}
			item, err := s.cfg.Reader.Read(ctx)
			if err != nil {
				if agbatch.IsEndOfStream(err) { return }
				if s.cfg.SkipOnRead != nil && s.cfg.SkipOnRead.ShouldSkip(err, int(stepExec.SkipCount)) {
					stepExec.IncSkip()
					slog.DebugContext(ctx, "agbatch/step: skipping reader error", "step", s.cfg.Name, "err", err)
					continue
				}
				sendErr(err); return
			}
			stepExec.IncRead()
			send(item)
		}
	})

	timeout := s.cfg.ChunkTimeout
	if timeout == 0 { timeout = -1 }
	chunks := rill.Batch(items, s.cfg.ChunkSize, timeout)

	err := rill.ForEach(chunks, 1, func(chunk []any) error {
		select {
		case <-ctx.Done(): return ctx.Err()
		default:
		}
		return s.processChunk(ctx, stepExec, chunk)
	})

	stepExec.EndTime = time.Now()
	stepExec.LastUpdated = time.Now()
	if s.cfg.Listener != nil { _ = s.cfg.Listener.AfterStep(ctx, stepExec) }

	if err != nil {
		stepExec.Status = agbatch.StatusFailed
		stepExec.FailureExcs = append(stepExec.FailureExcs, err)
		return stepExec, err
	}
	stepExec.Status = agbatch.StatusCompleted
	stepExec.ExitStatus = &agbatch.ExitStatus{Code: "COMPLETED"}
	return stepExec, nil
}

func (s *ChunkStep) processChunk(ctx context.Context, stepExec *agbatch.StepExecution, chunk []any) error {
	if s.cfg.ChunkListener != nil {
		if err := s.cfg.ChunkListener.BeforeChunk(ctx, stepExec); err != nil { return err }
	}
	processed := make([]any, len(chunk))
	if s.cfg.ProcessorPoolSize > 1 {
		idxs := rill.FromSlice(makeIdxSlice(len(chunk)), nil)
		results := rill.Map(idxs, s.cfg.ProcessorPoolSize, func(idx int) (struct{idx int; value any}, error) {
			result, err := s.processOneItem(ctx, stepExec, chunk[idx])
			if err != nil { return struct{idx int; value any}{}, err }
			return struct{idx int; value any}{idx: idx, value: result}, nil
		})
		for r := range results {
			if r.Error != nil {
				if s.cfg.ChunkListener != nil { s.cfg.ChunkListener.OnChunkError(ctx, stepExec, r.Error) }
				return r.Error
			}
			processed[r.Value.idx] = r.Value.value
		}
	} else {
		for i, item := range chunk {
			result, err := s.processOneItem(ctx, stepExec, item)
			if err != nil {
				if s.cfg.ChunkListener != nil { s.cfg.ChunkListener.OnChunkError(ctx, stepExec, err) }
				return err
			}
			processed[i] = result
		}
	}

	filtered := make([]any, 0, len(processed))
	for _, item := range processed {
		if item != nil { filtered = append(filtered, item) }
	}
	if len(filtered) == 0 {
		if s.cfg.ChunkListener != nil { _ = s.cfg.ChunkListener.AfterChunk(ctx, stepExec) }
		return nil
	}
	if err := s.cfg.Writer.Write(ctx, filtered); err != nil {
		if s.cfg.ChunkListener != nil { s.cfg.ChunkListener.OnChunkError(ctx, stepExec, err) }
		return fmt.Errorf("agbatch/step: writer error in %q: %w", s.cfg.Name, err)
	}
	stepExec.IncWrite()
	if s.cfg.ChunkListener != nil { _ = s.cfg.ChunkListener.AfterChunk(ctx, stepExec) }
	return nil
}

func (s *ChunkStep) processOneItem(ctx context.Context, stepExec *agbatch.StepExecution, item any) (any, error) {
	var lastErr error
	for attempt := 0; ; attempt++ {
		if attempt > 0 { stepExec.IncRetry() }
		result, err := s.cfg.Processor.Process(ctx, item)
		if err == nil { return result, nil }
		lastErr = err
		if s.cfg.Skip != nil && s.cfg.Skip.ShouldSkip(err, int(stepExec.SkipCount)) {
			stepExec.IncSkip()
			slog.DebugContext(ctx, "agbatch/step: skipping item", "step", s.cfg.Name, "err", err)
			return nil, nil
		}
		if s.cfg.Retry != nil {
			if shouldRetry, delay := s.cfg.Retry.ShouldRetry(attempt+1, err); shouldRetry {
				select { case <-ctx.Done(): return nil, ctx.Err(); case <-time.After(delay): }
				continue
			}
		}
		break
	}
	return nil, fmt.Errorf("agbatch/step: processor error in %q: %w", s.cfg.Name, lastErr)
}

func makeIdxSlice(n int) []int { s := make([]int, n); for i := range s { s[i] = i }; return s }

// ── ChunkStepBuilder ──────────────────────────────────────────────

// ChunkStepBuilder provides a fluent API for constructing chunk-oriented steps.
type ChunkStepBuilder[T, R any] struct {
	name              string
	reader            agbatch.ItemReader[T]
	processor         agbatch.ItemProcessor[T, R]
	writer            agbatch.ItemWriter[R]
	chunkSize         int
	processorPoolSize int
	retry             agbatch.RetryPolicy
	skip              agbatch.SkipPolicy
	skipOnRead        agbatch.SkipPolicy
	chunkTimeout      time.Duration
	listener          agbatch.StepListener
	chunkListener     agbatch.ChunkListener
}

// NewChunkStepBuilder creates a chunk step builder.
func NewChunkStepBuilder[T, R any](name string) *ChunkStepBuilder[T, R] {
	return &ChunkStepBuilder[T, R]{name: name, chunkSize: 100, processorPoolSize: 1}
}

func (b *ChunkStepBuilder[T, R]) Reader(r agbatch.ItemReader[T]) *ChunkStepBuilder[T, R] { b.reader = r; return b }
func (b *ChunkStepBuilder[T, R]) Processor(p agbatch.ItemProcessor[T, R]) *ChunkStepBuilder[T, R] { b.processor = p; return b }
func (b *ChunkStepBuilder[T, R]) Writer(w agbatch.ItemWriter[R]) *ChunkStepBuilder[T, R] { b.writer = w; return b }
func (b *ChunkStepBuilder[T, R]) ChunkSize(n int) *ChunkStepBuilder[T, R] { b.chunkSize = n; return b }
func (b *ChunkStepBuilder[T, R]) ProcessorPoolSize(n int) *ChunkStepBuilder[T, R] { b.processorPoolSize = n; return b }
func (b *ChunkStepBuilder[T, R]) RetryPolicy(p agbatch.RetryPolicy) *ChunkStepBuilder[T, R] { b.retry = p; return b }
func (b *ChunkStepBuilder[T, R]) SkipPolicy(p agbatch.SkipPolicy) *ChunkStepBuilder[T, R] { b.skip = p; return b }
func (b *ChunkStepBuilder[T, R]) SkipOnRead(p agbatch.SkipPolicy) *ChunkStepBuilder[T, R] { b.skipOnRead = p; return b }
func (b *ChunkStepBuilder[T, R]) ChunkTimeout(d time.Duration) *ChunkStepBuilder[T, R] { b.chunkTimeout = d; return b }
func (b *ChunkStepBuilder[T, R]) Listener(l agbatch.StepListener) *ChunkStepBuilder[T, R] { b.listener = l; return b }
func (b *ChunkStepBuilder[T, R]) ChunkListener(l agbatch.ChunkListener) *ChunkStepBuilder[T, R] { b.chunkListener = l; return b }

// Build constructs the ChunkStep.
func (b *ChunkStepBuilder[T, R]) Build() *ChunkStep {
	return NewChunkStep(ChunkStepConfig{
		Name: b.name, ChunkSize: b.chunkSize, ProcessorPoolSize: b.processorPoolSize,
		Reader: &anyReader[T]{b.reader}, Processor: &anyProcessor[T, R]{b.processor}, Writer: &anyWriter[R]{b.writer},
		Retry: b.retry, Skip: b.skip, SkipOnRead: b.skipOnRead, ChunkTimeout: b.chunkTimeout,
		Listener: b.listener, ChunkListener: b.chunkListener,
	})
}

// Internal any-type adapters.
type anyReader[T any] struct{ delegate agbatch.ItemReader[T] }
func (r *anyReader[T]) Read(ctx context.Context) (any, error) { return r.delegate.Read(ctx) }

type anyProcessor[T, R any] struct{ delegate agbatch.ItemProcessor[T, R] }
func (p *anyProcessor[T, R]) Process(ctx context.Context, item any) (any, error) { return p.delegate.Process(ctx, item.(T)) }

type anyWriter[T any] struct{ delegate agbatch.ItemWriter[T] }
func (w *anyWriter[T]) Write(ctx context.Context, items []any) error {
	typed := make([]T, len(items))
	for i, item := range items { if item != nil { typed[i] = item.(T) } }
	return w.delegate.Write(ctx, typed)
}
