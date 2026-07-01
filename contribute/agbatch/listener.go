package agbatch

import "context"

// JobListener receives callbacks during job lifecycle events.
type JobListener interface {
	BeforeJob(ctx context.Context, exec *JobExecution) error
	AfterJob(ctx context.Context, exec *JobExecution) error
}

// StepListener receives callbacks during step lifecycle events.
type StepListener interface {
	BeforeStep(ctx context.Context, exec *StepExecution) error
	AfterStep(ctx context.Context, exec *StepExecution) error
}

// ChunkListener receives callbacks around chunk processing.
type ChunkListener interface {
	BeforeChunk(ctx context.Context, exec *StepExecution) error
	AfterChunk(ctx context.Context, exec *StepExecution) error
	OnChunkError(ctx context.Context, exec *StepExecution, err error)
}

// --- No-op implementations for embedding ---

type noopJobListener struct{}

func (noopJobListener) BeforeJob(context.Context, *JobExecution) error { return nil }
func (noopJobListener) AfterJob(context.Context, *JobExecution) error  { return nil }

// JobListenerFunc adapts standalone before/after functions to JobListener.
func JobListenerFunc(before, after func(ctx context.Context, exec *JobExecution) error) JobListener {
	return &funcJobListener{before: before, after: after}
}

type funcJobListener struct {
	noopJobListener
	before, after func(ctx context.Context, exec *JobExecution) error
}

func (l *funcJobListener) BeforeJob(ctx context.Context, exec *JobExecution) error {
	if l.before != nil {
		return l.before(ctx, exec)
	}
	return nil
}

func (l *funcJobListener) AfterJob(ctx context.Context, exec *JobExecution) error {
	if l.after != nil {
		return l.after(ctx, exec)
	}
	return nil
}

type noopStepListener struct{}

func (noopStepListener) BeforeStep(context.Context, *StepExecution) error { return nil }
func (noopStepListener) AfterStep(context.Context, *StepExecution) error  { return nil }

// StepListenerFunc adapts standalone before/after functions to StepListener.
func StepListenerFunc(before, after func(ctx context.Context, exec *StepExecution) error) StepListener {
	return &funcStepListener{before: before, after: after}
}

type funcStepListener struct {
	noopStepListener
	before, after func(ctx context.Context, exec *StepExecution) error
}

func (l *funcStepListener) BeforeStep(ctx context.Context, exec *StepExecution) error {
	if l.before != nil {
		return l.before(ctx, exec)
	}
	return nil
}

func (l *funcStepListener) AfterStep(ctx context.Context, exec *StepExecution) error {
	if l.after != nil {
		return l.after(ctx, exec)
	}
	return nil
}
