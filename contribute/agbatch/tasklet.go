package agbatch

import (
	"context"
	"fmt"
	"time"
)

// TaskletStep is a step that executes a single tasklet.
// For complex steps (ChunkStep, FlowStep, PartitionedStep), see the step/ sub-package.
type TaskletStep struct {
	name     string
	tasklet  Tasklet
	listener StepListener
}

// NewTaskletStep creates a tasklet step.
func NewTaskletStep(name string, tasklet Tasklet, listener StepListener) *TaskletStep {
	return &TaskletStep{name: name, tasklet: tasklet, listener: listener}
}

// Name returns the step name.
func (s *TaskletStep) Name() string { return s.name }

// Execute runs the tasklet.
func (s *TaskletStep) Execute(ctx context.Context, _ *JobExecution) (*StepExecution, error) {
	stepExec := NewStepExecution(0, s.name)
	stepExec.Status = StatusStarted
	if s.listener != nil {
		if err := s.listener.BeforeStep(ctx, stepExec); err != nil {
			stepExec.Status = StatusFailed
			return stepExec, fmt.Errorf("agbatch: before step hook: %w", err)
		}
	}
	err := s.tasklet.Execute(ctx, stepExec)
	stepExec.EndTime = time.Now()
	stepExec.LastUpdated = time.Now()
	if s.listener != nil {
		_ = s.listener.AfterStep(ctx, stepExec)
	}
	if err != nil {
		stepExec.Status = StatusFailed
		stepExec.FailureExcs = append(stepExec.FailureExcs, err)
		return stepExec, err
	}
	stepExec.Status = StatusCompleted
	stepExec.ExitStatus = &ExitStatus{Code: "COMPLETED"}
	return stepExec, nil
}
