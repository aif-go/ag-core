package agbatch

// JobBuilder provides a fluent API for constructing jobs.
//
// Steps are built using the step/ sub-package builders:
//
//	import agstep "github.com/aif-go/ag-core/contribute/agbatch/step"
//
//	job := agbatch.NewJobBuilder("myJob").
//	    Step(agstep.NewChunkStepBuilder[In, Out]("step1").Reader(r).Processor(p).Writer(w).Build()).
//	    Step(agstep.NewTaskletStep("cleanup", myTasklet, nil)).
//	    Build()
type JobBuilder struct {
	name       string
	steps      []Step
	listener   JobListener
	repository JobRepository
}

// NewJobBuilder creates a new job builder.
func NewJobBuilder(name string) *JobBuilder {
	return &JobBuilder{name: name}
}

// Step adds a step to the job. Steps execute in order.
func (b *JobBuilder) Step(step Step) *JobBuilder {
	b.steps = append(b.steps, step)
	return b
}

// Listener sets the job lifecycle listener.
func (b *JobBuilder) Listener(l JobListener) *JobBuilder {
	b.listener = l
	return b
}

// Repository sets the job repository (nil = InMemoryRepository).
func (b *JobBuilder) Repository(r JobRepository) *JobBuilder {
	b.repository = r
	return b
}

// Build constructs the Job.
func (b *JobBuilder) Build() *Job {
	return NewJob(JobConfig{
		Name: b.name, Steps: b.steps, Listener: b.listener, Repository: b.repository,
	})
}
