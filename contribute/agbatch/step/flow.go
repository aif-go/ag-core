package step

import (
	"context"
	"fmt"
	"log/slog"

	agbatch "github.com/aif-go/ag-core/contribute/agbatch"
)

// FlowDecision evaluates the result of a step and decides the next step.
type FlowDecision = agbatch.FlowDecision
type FlowDecisionFunc = agbatch.FlowDecisionFunc

// FlowStep is a step that contains a graph of sub-steps with conditional transitions.
type FlowStep struct {
	name      string
	steps     map[string]agbatch.Step
	decisions map[string]FlowDecision
	firstStep string
	listener  agbatch.StepListener
}

// FlowStepBuilder constructs a FlowStep.
type FlowStepBuilder struct {
	name      string
	steps     map[string]agbatch.Step
	decisions map[string]FlowDecision
	firstStep string
	listener  agbatch.StepListener
}

func NewFlowStepBuilder(name string) *FlowStepBuilder {
	return &FlowStepBuilder{name: name, steps: make(map[string]agbatch.Step), decisions: make(map[string]FlowDecision)}
}

func (b *FlowStepBuilder) First(stepName string) *FlowStepBuilder             { b.firstStep = stepName; return b }
func (b *FlowStepBuilder) Listener(l agbatch.StepListener) *FlowStepBuilder   { b.listener = l; return b }
func (b *FlowStepBuilder) Step(name string, step agbatch.Step) *FlowStepBuilder { b.steps[name] = step; return b }
func (b *FlowStepBuilder) Next(fromStep string, decision FlowDecision) *FlowStepBuilder { b.decisions[fromStep] = decision; return b }
func (b *FlowStepBuilder) Decider(name string, step agbatch.Step, decision FlowDecision) *FlowStepBuilder {
	b.steps[name] = step; b.decisions[name] = decision; return b
}
func (b *FlowStepBuilder) Build() *FlowStep {
	return &FlowStep{name: b.name, steps: b.steps, decisions: b.decisions, firstStep: b.firstStep, listener: b.listener}
}

func (s *FlowStep) Name() string { return s.name }

func (s *FlowStep) Execute(ctx context.Context, jobExec *agbatch.JobExecution) (*agbatch.StepExecution, error) {
	current := s.firstStep
	if current == "" {
		return nil, fmt.Errorf("agbatch/step: flow %q has no first step", s.name)
	}
	visited := make(map[string]int)
	var lastExec *agbatch.StepExecution
	for current != "" {
		visited[current]++
		if visited[current] > 10 {
			return nil, fmt.Errorf("agbatch/step: flow %q cycle at %q", s.name, current)
		}
		step, ok := s.steps[current]
		if !ok { return nil, fmt.Errorf("agbatch/step: flow %q unknown step %q", s.name, current) }
		slog.InfoContext(ctx, "agbatch/step: flow executing", "flow", s.name, "step", current)
		exec, err := step.Execute(ctx, jobExec)
		lastExec = exec
		jobExec.StepExecs = append(jobExec.StepExecs, exec)
		if err != nil {
			slog.ErrorContext(ctx, "agbatch/step: flow step failed", "flow", s.name, "step", current, "err", err)
			dec, hasDec := s.decisions[current]
			if !hasDec { return exec, err }
			next, decErr := dec.Decide(exec)
			if decErr != nil { return exec, fmt.Errorf("agbatch/step: decision error: %w", decErr) }
			if next == "" { break }
			current = next; continue
		}
		dec, hasDec := s.decisions[current]
		if !hasDec { break }
		next, err := dec.Decide(exec)
		if err != nil { return exec, fmt.Errorf("agbatch/step: decision error: %w", err) }
		current = next
	}
	if lastExec == nil {
		lastExec = agbatch.NewStepExecution(0, s.name)
		lastExec.Status = agbatch.StatusCompleted
	}
	return lastExec, nil
}
