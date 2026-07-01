package step

import (
	"context"
	"fmt"
	"sync"
	"time"

	agbatch "github.com/aif-go/ag-core/contribute/agbatch"
	"github.com/destel/rill"
)

// Partitioner splits items into partitions.
type Partitioner[T any] interface {
	Partition(items []T, numPartitions int) [][]T
}

// PartitionedStepConfig holds configuration for a partitioned step.
type PartitionedStepConfig struct {
	Name          string
	Reader        agbatch.ItemReader[any]
	Partitioner   interface{ Partition([]any, int) [][]any }
	Processor     agbatch.ItemProcessor[any, any]
	Writer        agbatch.ItemWriter[any]
	ChunkSize     int
	NumPartitions int
	Retry         agbatch.RetryPolicy
	Skip          agbatch.SkipPolicy
	ChunkTimeout  time.Duration
	Listener      agbatch.StepListener
}

// PartitionedStep reads items, partitions them, and processes each partition concurrently.
type PartitionedStep struct{ cfg PartitionedStepConfig }

func NewPartitionedStep(cfg PartitionedStepConfig) *PartitionedStep {
	if cfg.ChunkSize <= 0 { cfg.ChunkSize = 100 }
	if cfg.NumPartitions <= 0 { cfg.NumPartitions = 4 }
	return &PartitionedStep{cfg: cfg}
}

func (s *PartitionedStep) Name() string { return s.cfg.Name }

func (s *PartitionedStep) Execute(ctx context.Context, jobExec *agbatch.JobExecution) (*agbatch.StepExecution, error) {
	stepExec := agbatch.NewStepExecution(0, s.cfg.Name)
	stepExec.Status = agbatch.StatusStarted
	if s.cfg.Listener != nil {
		if err := s.cfg.Listener.BeforeStep(ctx, stepExec); err != nil {
			stepExec.Status = agbatch.StatusFailed
			return stepExec, fmt.Errorf("agbatch/step: before hook: %w", err)
		}
	}

	var allItems []any
	for {
		select { case <-ctx.Done(): stepExec.Status = agbatch.StatusStopped; return stepExec, ctx.Err(); default: }
		item, err := s.cfg.Reader.Read(ctx)
		if err != nil {
			if agbatch.IsEndOfStream(err) { break }
			stepExec.Status = agbatch.StatusFailed
			return stepExec, fmt.Errorf("agbatch/step: reader error: %w", err)
		}
		stepExec.IncRead()
		allItems = append(allItems, item)
	}
	if len(allItems) == 0 {
		stepExec.Status = agbatch.StatusCompleted; stepExec.EndTime = time.Now()
		return stepExec, nil
	}

	partitions := s.cfg.Partitioner.Partition(allItems, s.cfg.NumPartitions)
	var wg sync.WaitGroup; var mu sync.Mutex; var firstErr error
	var totalRead, totalWrite, totalSkip, totalRetry int64

	for i, partition := range partitions {
		if len(partition) == 0 { continue }
		select { case <-ctx.Done(): stepExec.Status = agbatch.StatusStopped; return stepExec, ctx.Err(); default: }
		wg.Add(1)
		go func(pid int, items []any) {
			defer wg.Done()
			partExec := agbatch.NewStepExecution(0, fmt.Sprintf("%s-p%d", s.cfg.Name, pid))
			timeout := s.cfg.ChunkTimeout; if timeout == 0 { timeout = -1 }
			itemStream := rill.FromSlice(items, nil)
			chunks := rill.Batch(itemStream, s.cfg.ChunkSize, timeout)
			err := rill.ForEach(chunks, 1, func(chunk []any) error {
				select { case <-ctx.Done(): return ctx.Err(); default: }
				for _, item := range chunk {
					if _, e := processOneItem(ctx, partExec, item, s.cfg.Processor, s.cfg.Retry, s.cfg.Skip); e != nil { return e }
				}
				if err := s.cfg.Writer.Write(ctx, chunk); err != nil {
					return fmt.Errorf("agbatch/step: writer error: %w", err)
				}
				partExec.IncWrite()
				return nil
			})
			mu.Lock()
			if err != nil && firstErr == nil { firstErr = err }
			totalRead += partExec.ReadCount; totalWrite += partExec.WriteCount
			totalSkip += partExec.SkipCount; totalRetry += partExec.RetryCount
			mu.Unlock()
		}(i, partition)
	}
	wg.Wait()

	stepExec.ReadCount = totalRead; stepExec.WriteCount = totalWrite
	stepExec.SkipCount = totalSkip; stepExec.RetryCount = totalRetry
	stepExec.EndTime = time.Now(); stepExec.LastUpdated = time.Now()
	if s.cfg.Listener != nil { _ = s.cfg.Listener.AfterStep(ctx, stepExec) }
	if firstErr != nil { stepExec.Status = agbatch.StatusFailed; return stepExec, firstErr }
	stepExec.Status = agbatch.StatusCompleted
	return stepExec, nil
}

func processOneItem(ctx context.Context, exec *agbatch.StepExecution, item any, proc agbatch.ItemProcessor[any, any], retry agbatch.RetryPolicy, skip agbatch.SkipPolicy) (any, error) {
	exec.IncRead()
	var lastErr error
	for attempt := 0; ; attempt++ {
		if attempt > 0 { exec.IncRetry() }
		result, err := proc.Process(ctx, item)
		if err == nil { return result, nil }
		lastErr = err
		if skip != nil && skip.ShouldSkip(err, int(exec.SkipCount)) { exec.IncSkip(); return nil, nil }
		if retry != nil {
			if ok, delay := retry.ShouldRetry(attempt+1, err); ok {
				select { case <-ctx.Done(): return nil, ctx.Err(); case <-time.After(delay): }
				continue
			}
		}
		break
	}
	return nil, fmt.Errorf("agbatch/step: processor error: %w", lastErr)
}
