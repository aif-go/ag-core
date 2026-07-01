package agbatch

import (
	"sync"
	"time"
)

// BatchStatus represents the current status of a job or step execution.
type BatchStatus string

const (
	StatusStarting  BatchStatus = "STARTING"
	StatusStarted   BatchStatus = "STARTED"
	StatusRunning   BatchStatus = "RUNNING"
	StatusCompleted BatchStatus = "COMPLETED"
	StatusFailed    BatchStatus = "FAILED"
	StatusStopped   BatchStatus = "STOPPED"
)

// ExitStatus represents the exit code and description of an execution.
type ExitStatus struct {
	Code        string
	Description string
}

// ExecutionContext holds arbitrary key-value state shared across steps within a job.
type ExecutionContext struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewExecutionContext creates a new empty execution context.
func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{data: make(map[string]any)}
}

// Get returns the value for key, or nil if not present.
func (ec *ExecutionContext) Get(key string) any {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.data[key]
}

// Set stores a value for key.
func (ec *ExecutionContext) Set(key string, val any) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.data[key] = val
}

// JobExecution tracks the lifecycle of a single job run.
type JobExecution struct {
	ID          int64
	JobName     string
	Status      BatchStatus
	ExitStatus  *ExitStatus
	StartTime   time.Time
	EndTime     time.Time
	LastUpdated time.Time
	Context     *ExecutionContext
	StepExecs   []*StepExecution
	FailureExcs []error
	mu          sync.RWMutex
}

// NewJobExecution creates a new execution for the given job name.
func NewJobExecution(id int64, jobName string) *JobExecution {
	now := time.Now()
	return &JobExecution{
		ID:          id,
		JobName:     jobName,
		Status:      StatusStarting,
		StartTime:   now,
		LastUpdated: now,
		Context:     NewExecutionContext(),
	}
}

// StepExecution tracks the lifecycle of a single step run.
type StepExecution struct {
	ID          int64
	StepName    string
	Status      BatchStatus
	ExitStatus  *ExitStatus
	StartTime   time.Time
	EndTime     time.Time
	LastUpdated time.Time
	ReadCount   int64
	WriteCount  int64
	SkipCount   int64
	RetryCount  int64
	FilterCount int64
	Context     *ExecutionContext
	FailureExcs []error
	mu          sync.RWMutex
}

// NewStepExecution creates a new execution for the given step name.
func NewStepExecution(id int64, stepName string) *StepExecution {
	now := time.Now()
	return &StepExecution{
		ID:          id,
		StepName:    stepName,
		Status:      StatusStarting,
		StartTime:   now,
		LastUpdated: now,
		Context:     NewExecutionContext(),
	}
}

// --- Counters (safe for concurrent use) ---

func (se *StepExecution) IncRead() {
	se.mu.Lock()
	se.ReadCount++
	se.mu.Unlock()
}

func (se *StepExecution) IncWrite() {
	se.mu.Lock()
	se.WriteCount++
	se.mu.Unlock()
}

func (se *StepExecution) IncSkip() {
	se.mu.Lock()
	se.SkipCount++
	se.mu.Unlock()
}

func (se *StepExecution) IncRetry() {
	se.mu.Lock()
	se.RetryCount++
	se.mu.Unlock()
}
