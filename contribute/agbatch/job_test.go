package agbatch

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
)

func TestJobLauncher_SingleStep(t *testing.T) {
	// Simple job with one tasklet step
	tasklet := TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		return nil
	})

	job := NewJobBuilder("singleStepJob").
		Step(NewTaskletStep("step1", tasklet, nil)).
		Build()

	launcher := NewJobLauncher(NewInMemoryRepository())
	exec, err := launcher.Run(context.Background(), job)
	if err != nil {
		t.Fatalf("job failed: %v", err)
	}
	if exec.Status != StatusCompleted {
		t.Errorf("expected COMPLETED, got %s", exec.Status)
	}
	if len(exec.StepExecs) != 1 {
		t.Errorf("expected 1 step execution, got %d", len(exec.StepExecs))
	}
	if exec.StepExecs[0].Status != StatusCompleted {
		t.Errorf("step status: got %s, want %s", exec.StepExecs[0].Status, StatusCompleted)
	}
}

func TestJobLauncher_MultipleSteps(t *testing.T) {
	order := make([]string, 0)

	step1 := NewTaskletStep("step1", TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		order = append(order, "step1")
		return nil
	}), nil)

	step2 := NewTaskletStep("step2", TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		order = append(order, "step2")
		return nil
	}), nil)

	step3 := NewTaskletStep("step3", TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		order = append(order, "step3")
		return nil
	}), nil)

	job := NewJobBuilder("multiStepJob").
		Step(step1).
		Step(step2).
		Step(step3).
		Build()

	launcher := NewJobLauncher(NewInMemoryRepository())
	exec, err := launcher.Run(context.Background(), job)
	if err != nil {
		t.Fatalf("job failed: %v", err)
	}
	if len(exec.StepExecs) != 3 {
		t.Errorf("expected 3 step executions, got %d", len(exec.StepExecs))
	}
	if order[0] != "step1" || order[1] != "step2" || order[2] != "step3" {
		t.Errorf("wrong step order: %v", order)
	}
}

func TestJobLauncher_StepFailureStopsJob(t *testing.T) {
	step3ran := false

	step1 := NewTaskletStep("step1", TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		return nil
	}), nil)

	step2 := NewTaskletStep("step2", TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		return errors.New("step2 failure")
	}), nil)

	step3 := NewTaskletStep("step3", TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		step3ran = true
		return nil
	}), nil)

	job := NewJobBuilder("failFastJob").
		Step(step1).
		Step(step2).
		Step(step3).
		Build()

	launcher := NewJobLauncher(NewInMemoryRepository())
	exec, err := launcher.Run(context.Background(), job)
	if err == nil {
		t.Fatal("expected job failure")
	}
	if exec.Status != StatusFailed {
		t.Errorf("expected FAILED, got %s", exec.Status)
	}
	if step3ran {
		t.Error("step3 should not run after step2 failure")
	}
	if len(exec.StepExecs) != 2 {
		t.Errorf("expected 2 step executions, got %d", len(exec.StepExecs))
	}
}

func TestJobLauncher_ChunkStepInJob(t *testing.T) {
	items := make([]int, 25)
	for i := range items {
		items[i] = i
	}
	idx := atomic.Int64{}
	reader := ReaderFunc[int](func(_ context.Context) (int, error) {
		i := idx.Add(1) - 1
		if int(i) >= len(items) {
			return 0, io.EOF
		}
		return items[i], nil
	})

	processor := ProcessorFunc[int, int](func(_ context.Context, item int) (int, error) {
		return item * 3, nil
	})

	var written []int
	writer := WriterFunc[int](func(_ context.Context, items []int) error {
		written = append(written, items...)
		return nil
	})

	chunkStep := NewChunkStepBuilder[int, int]("chunkStep").
		Reader(reader).
		Processor(processor).
		Writer(writer).
		ChunkSize(5).
		Build()

	job := NewJobBuilder("chunkJob").
		Step(chunkStep).
		Build()

	launcher := NewJobLauncher(NewInMemoryRepository())
	exec, err := launcher.Run(context.Background(), job)
	if err != nil {
		t.Fatalf("job failed: %v", err)
	}
	if exec.Status != StatusCompleted {
		t.Errorf("expected COMPLETED, got %s", exec.Status)
	}
	if len(written) != 25 {
		t.Errorf("expected 25 written, got %d", len(written))
	}
	// Verify transformation
	for i, v := range written {
		if v != i*3 {
			t.Errorf("written[%d] = %d, want %d", i, v, i*3)
		}
	}
}

func TestJobListener_BeforeAndAfter(t *testing.T) {
	events := make([]string, 0)

	listener := JobListenerFunc(
		func(_ context.Context, _ *JobExecution) error {
			events = append(events, "before")
			return nil
		},
		func(_ context.Context, _ *JobExecution) error {
			events = append(events, "after")
			return nil
		},
	)

	step := NewTaskletStep("s1", TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		events = append(events, "step")
		return nil
	}), nil)

	job := NewJobBuilder("listenerJob").
		Step(step).
		Listener(listener).
		Build()

	launcher := NewJobLauncher(NewInMemoryRepository())
	_, err := launcher.Run(context.Background(), job)
	if err != nil {
		t.Fatalf("job failed: %v", err)
	}

	if len(events) != 3 || events[0] != "before" || events[1] != "step" || events[2] != "after" {
		t.Errorf("wrong event order: %v", events)
	}
}

func TestJobListener_BeforeError(t *testing.T) {
	listener := JobListenerFunc(
		func(_ context.Context, _ *JobExecution) error {
			return errors.New("before job failure")
		},
		nil,
	)

	stepRan := false
	step := NewTaskletStep("s1", TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		stepRan = true
		return nil
	}), nil)

	job := NewJobBuilder("beforeFailJob").
		Step(step).
		Listener(listener).
		Build()

	launcher := NewJobLauncher(NewInMemoryRepository())
	_, err := launcher.Run(context.Background(), job)
	if err == nil {
		t.Fatal("expected error from before job listener")
	}
	if stepRan {
		t.Error("step should not run after beforeJob error")
	}
}

func TestJobLauncher_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before job starts

	step := NewTaskletStep("step1", TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		return nil
	}), nil)

	job := NewJobBuilder("cancelJob").
		Step(step).
		Build()

	launcher := NewJobLauncher(NewInMemoryRepository())
	_, err := launcher.Run(ctx, job)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestJobLauncher_Repository(t *testing.T) {
	repo := NewInMemoryRepository()

	step := NewTaskletStep("s1", TaskletFunc(func(ctx context.Context, exec *StepExecution) error {
		return nil
	}), nil)

	job := NewJobBuilder("repoJob").
		Step(step).
		Repository(repo).
		Build()

	launcher := NewJobLauncher(repo)
	exec, err := launcher.Run(context.Background(), job)
	if err != nil {
		t.Fatalf("job failed: %v", err)
	}

	// Verify it was saved
	saved, _ := repo.GetJobExecution(context.Background(), exec.ID)
	if saved == nil {
		t.Error("job execution not found in repository")
	}
	if saved.Status != StatusCompleted {
		t.Errorf("saved status: got %s, want %s", saved.Status, StatusCompleted)
	}
}

func TestExecutionContext(t *testing.T) {
	ec := NewExecutionContext()
	ec.Set("key1", "value1")
	ec.Set("key2", 42)

	if ec.Get("key1") != "value1" {
		t.Error("key1 mismatch")
	}
	if ec.Get("key2") != 42 {
		t.Error("key2 mismatch")
	}
	if ec.Get("nonexistent") != nil {
		t.Error("should return nil for nonexistent key")
	}
}

func TestJobExecution_Lifecycle(t *testing.T) {
	exec := NewJobExecution(1, "testJob")
	if exec.Status != StatusStarting {
		t.Errorf("initial status: got %s, want %s", exec.Status, StatusStarting)
	}
	if exec.JobName != "testJob" {
		t.Errorf("job name: got %s, want testJob", exec.JobName)
	}
	if exec.ID != 1 {
		t.Errorf("job id: got %d, want 1", exec.ID)
	}
	if exec.Context == nil {
		t.Error("execution context should not be nil")
	}
}

func TestStepExecution_Counters(t *testing.T) {
	exec := NewStepExecution(1, "testStep")
	exec.IncRead()
	exec.IncRead()
	exec.IncWrite()
	exec.IncSkip()
	exec.IncRetry()
	exec.IncRetry()

	if exec.ReadCount != 2 {
		t.Errorf("read count: got %d, want 2", exec.ReadCount)
	}
	if exec.WriteCount != 1 {
		t.Errorf("write count: got %d, want 1", exec.WriteCount)
	}
	if exec.SkipCount != 1 {
		t.Errorf("skip count: got %d, want 1", exec.SkipCount)
	}
	if exec.RetryCount != 2 {
		t.Errorf("retry count: got %d, want 2", exec.RetryCount)
	}
}

func TestJob_EmptySteps(t *testing.T) {
	job := NewJobBuilder("emptyJob").Build()
	launcher := NewJobLauncher(NewInMemoryRepository())
	exec, err := launcher.Run(context.Background(), job)
	if err != nil {
		t.Fatalf("empty job should succeed: %v", err)
	}
	if exec.Status != StatusCompleted {
		t.Errorf("expected COMPLETED, got %s", exec.Status)
	}
	if len(exec.StepExecs) != 0 {
		t.Errorf("expected 0 step executions, got %d", len(exec.StepExecs))
	}
}
