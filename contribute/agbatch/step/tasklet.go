package step

import (
	"context"
	"fmt"
	"time"

	agbatch "github.com/aif-go/ag-core/contribute/agbatch"
)

// TaskletStep is a step that executes a single tasklet.
type TaskletStep struct {
	name     string
	tasklet  agbatch.Tasklet
	listener agbatch.StepListener
}

// NewTaskletStep creates a tasklet step.
func NewTaskletStep(name string, tasklet agbatch.Tasklet, listener agbatch.StepListener) *TaskletStep {
	return &TaskletStep{name: name, tasklet: tasklet, listener: listener}
}

// Name returns the step name.
func (s *TaskletStep) Name() string { return s.name }

// Execute runs the tasklet.
func (s *TaskletStep) Execute(ctx context.Context, _ *agbatch.JobExecution) (*agbatch.StepExecution, error) {
	stepExec := agbatch.NewStepExecution(0, s.name)
	stepExec.Status = agbatch.StatusStarted
	if s.listener != nil {
		if err := s.listener.BeforeStep(ctx, stepExec); err != nil {
			stepExec.Status = agbatch.StatusFailed
			return stepExec, fmt.Errorf("agbatch/step: before step hook: %w", err)
		}
	}
	err := s.tasklet.Execute(ctx, stepExec)
	stepExec.EndTime = time.Now()
	stepExec.LastUpdated = time.Now()
	if s.listener != nil { _ = s.listener.AfterStep(ctx, stepExec) }
	if err != nil {
		stepExec.Status = agbatch.StatusFailed
		stepExec.FailureExcs = append(stepExec.FailureExcs, err)
		return stepExec, err
	}
	stepExec.Status = agbatch.StatusCompleted
	stepExec.ExitStatus = &agbatch.ExitStatus{Code: "COMPLETED"}
	return stepExec, nil
}
