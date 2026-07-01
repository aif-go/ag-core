package agbatch

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// JobConfig holds the configuration for a job.
type JobConfig struct {
	Name       string
	Steps      []Step
	Listener   JobListener
	Repository JobRepository // nil = in-memory (no persistence)
}

// Job is a complete batch process consisting of one or more steps.
// Steps execute sequentially in the order they are defined.
type Job struct {
	cfg JobConfig
}

// NewJob creates a new job with the given configuration.
func NewJob(cfg JobConfig) *Job {
	if cfg.Repository == nil {
		cfg.Repository = NewInMemoryRepository()
	}
	return &Job{cfg: cfg}
}

// Name returns the job name.
func (j *Job) Name() string { return j.cfg.Name }

// Steps returns the steps in this job.
func (j *Job) Steps() []Step { return j.cfg.Steps }

// Repository returns the job repository.
func (j *Job) Repository() JobRepository { return j.cfg.Repository }

// listener returns the job listener (or noop if nil).
func (j *Job) listener() JobListener {
	if j.cfg.Listener != nil {
		return j.cfg.Listener
	}
	return noopJobListener{}
}

// JobLauncher is the entry point for running jobs.
type JobLauncher struct {
	repo JobRepository
}

// NewJobLauncher creates a new job launcher.
// repo is the default repository; individual jobs may override it.
func NewJobLauncher(repo JobRepository) *JobLauncher {
	if repo == nil {
		repo = NewInMemoryRepository()
	}
	return &JobLauncher{repo: repo}
}

// Run executes a job synchronously (blocks until completion or failure).
// Returns the job execution with its status and any error.
func (l *JobLauncher) Run(ctx context.Context, job *Job) (*JobExecution, error) {
	return l.runJob(ctx, job)
}

// runJob is the internal job execution engine.
func (l *JobLauncher) runJob(ctx context.Context, job *Job) (*JobExecution, error) {
	repo := job.Repository()
	if repo == nil {
		repo = l.repo
	}

	var jobID int64
	if memRepo, ok := repo.(*InMemoryRepository); ok {
		jobID = memRepo.NextJobID()
	}
	jobExec := NewJobExecution(jobID, job.Name())
	jobExec.Status = StatusStarting

	_ = repo.SaveJobExecution(ctx, jobExec)

	listener := job.listener()

	// Before job hook
	if err := listener.BeforeJob(ctx, jobExec); err != nil {
		jobExec.Status = StatusFailed
		jobExec.ExitStatus = &ExitStatus{Code: "FAILED", Description: err.Error()}
		_ = repo.UpdateJobExecution(ctx, jobExec)
		return jobExec, fmt.Errorf("agbatch: before job hook failed for %q: %w", job.Name(), err)
	}

	jobExec.Status = StatusRunning
	_ = repo.UpdateJobExecution(ctx, jobExec)

	slog.InfoContext(ctx, "agbatch: job started", "job", job.Name(), "steps", len(job.Steps()))

	// Execute steps sequentially
	var stepID int64
	if memRepo, ok := repo.(*InMemoryRepository); ok {
		stepID = memRepo.NextStepID() - 1 // will be incremented per step
	}

	for _, step := range job.Steps() {
		select {
		case <-ctx.Done():
			jobExec.Status = StatusStopped
			jobExec.EndTime = time.Now()
			jobExec.ExitStatus = &ExitStatus{Code: "STOPPED", Description: ctx.Err().Error()}
			_ = repo.UpdateJobExecution(ctx, jobExec)
			return jobExec, ctx.Err()
		default:
		}

		if memRepo, ok := repo.(*InMemoryRepository); ok {
			stepID = memRepo.NextStepID()
		}
		slog.InfoContext(ctx, "agbatch: step starting", "job", job.Name(), "step", step.Name())

		stepExec, err := step.Execute(ctx, jobExec)
		stepExec.ID = stepID
		_ = repo.SaveStepExecution(ctx, stepExec)
		jobExec.StepExecs = append(jobExec.StepExecs, stepExec)

		if err != nil {
			jobExec.Status = StatusFailed
			jobExec.EndTime = time.Now()
			jobExec.ExitStatus = &ExitStatus{Code: "FAILED", Description: err.Error()}
			jobExec.FailureExcs = append(jobExec.FailureExcs, err)
			_ = repo.UpdateJobExecution(ctx, jobExec)

			_ = listener.AfterJob(ctx, jobExec)
			slog.ErrorContext(ctx, "agbatch: job failed", "job", job.Name(), "step", step.Name(), "err", err)
			return jobExec, err
		}

		slog.InfoContext(ctx, "agbatch: step completed", "job", job.Name(), "step", step.Name(),
			"status", stepExec.Status, "read", stepExec.ReadCount, "write", stepExec.WriteCount,
			"skip", stepExec.SkipCount, "retry", stepExec.RetryCount)
	}

	jobExec.Status = StatusCompleted
	jobExec.EndTime = time.Now()
	jobExec.ExitStatus = &ExitStatus{Code: "COMPLETED"}
	_ = repo.UpdateJobExecution(ctx, jobExec)

	_ = listener.AfterJob(ctx, jobExec)
	slog.InfoContext(ctx, "agbatch: job completed", "job", job.Name(), "steps", len(jobExec.StepExecs))

	return jobExec, nil
}
