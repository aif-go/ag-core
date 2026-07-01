package agbatch

import (
	"context"
	"sync"
	"sync/atomic"
)

// JobRepository persists job and step execution metadata.
// The default in-memory implementation is suitable for single-node usage.
// Implement a persistent version (e.g., DB-backed) for production restartability.
type JobRepository interface {
	// SaveJobExecution persists a job execution.
	SaveJobExecution(ctx context.Context, exec *JobExecution) error
	// UpdateJobExecution updates an existing job execution.
	UpdateJobExecution(ctx context.Context, exec *JobExecution) error
	// GetJobExecution retrieves a job execution by ID.
	GetJobExecution(ctx context.Context, id int64) (*JobExecution, error)
	// SaveStepExecution persists a step execution.
	SaveStepExecution(ctx context.Context, exec *StepExecution) error
	// UpdateStepExecution updates an existing step execution.
	UpdateStepExecution(ctx context.Context, exec *StepExecution) error
	// GetStepExecution retrieves a step execution by ID.
	GetStepExecution(ctx context.Context, id int64) (*StepExecution, error)
}

// InMemoryRepository is an in-memory JobRepository, suitable for testing
// and single-node deployments where restartability is not required.
type InMemoryRepository struct {
	mu         sync.RWMutex
	jobExecs   map[int64]*JobExecution
	stepExecs  map[int64]*StepExecution
	nextJobID  atomic.Int64
	nextStepID atomic.Int64
}

// NewInMemoryRepository creates a new in-memory repository.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		jobExecs:  make(map[int64]*JobExecution),
		stepExecs: make(map[int64]*StepExecution),
	}
}

// NextJobID returns the next job execution ID.
func (r *InMemoryRepository) NextJobID() int64 { return r.nextJobID.Add(1) }

// NextStepID returns the next step execution ID.
func (r *InMemoryRepository) NextStepID() int64 { return r.nextStepID.Add(1) }

func (r *InMemoryRepository) SaveJobExecution(_ context.Context, exec *JobExecution) error {
	r.mu.Lock()
	r.jobExecs[exec.ID] = exec
	r.mu.Unlock()
	return nil
}

func (r *InMemoryRepository) UpdateJobExecution(_ context.Context, exec *JobExecution) error {
	r.mu.Lock()
	r.jobExecs[exec.ID] = exec
	r.mu.Unlock()
	return nil
}

func (r *InMemoryRepository) GetJobExecution(_ context.Context, id int64) (*JobExecution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	exec, ok := r.jobExecs[id]
	if !ok {
		return nil, nil
	}
	return exec, nil
}

func (r *InMemoryRepository) SaveStepExecution(_ context.Context, exec *StepExecution) error {
	r.mu.Lock()
	r.stepExecs[exec.ID] = exec
	r.mu.Unlock()
	return nil
}

func (r *InMemoryRepository) UpdateStepExecution(_ context.Context, exec *StepExecution) error {
	r.mu.Lock()
	r.stepExecs[exec.ID] = exec
	r.mu.Unlock()
	return nil
}

func (r *InMemoryRepository) GetStepExecution(_ context.Context, id int64) (*StepExecution, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	exec, ok := r.stepExecs[id]
	if !ok {
		return nil, nil
	}
	return exec, nil
}
