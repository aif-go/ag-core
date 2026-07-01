package agbatch

import "fmt"

// FlowDecision evaluates the result of a step and decides which step to execute next.
type FlowDecision interface {
	Decide(exec *StepExecution) (nextStep string, err error)
}

// FlowDecisionFunc adapts a function to FlowDecision.
type FlowDecisionFunc func(exec *StepExecution) (string, error)

func (f FlowDecisionFunc) Decide(exec *StepExecution) (string, error) { return f(exec) }

// OnStatus returns a decision that branches based on step execution status.
func OnStatus(statusToStep map[BatchStatus]string) FlowDecision {
	return FlowDecisionFunc(func(exec *StepExecution) (string, error) {
		next, ok := statusToStep[exec.Status]
		if !ok {
			return "", fmt.Errorf("agbatch: no transition for status %q", exec.Status)
		}
		return next, nil
	})
}

// OnExitCode returns a decision that branches based on exit code.
func OnExitCode(codes map[string]string) FlowDecision {
	return FlowDecisionFunc(func(exec *StepExecution) (string, error) {
		code := ""
		if exec.ExitStatus != nil {
			code = exec.ExitStatus.Code
		}
		next, ok := codes[code]
		if !ok {
			return "", fmt.Errorf("agbatch: no transition for exit code %q", code)
		}
		return next, nil
	})
}

// NextAlways returns a decision that always goes to the given step.
func NextAlways(nextStep string) FlowDecision {
	return FlowDecisionFunc(func(_ *StepExecution) (string, error) { return nextStep, nil })
}

// StopAlways returns a decision that always stops.
func StopAlways() FlowDecision {
	return FlowDecisionFunc(func(_ *StepExecution) (string, error) { return "", nil })
}
