package orchestrator

import (
	"fmt"
)

type Orchestrator struct {
	registry   *Registry
	decomposer *Decomposer
	merger     *Merger
}

func NewOrchestrator(reg *Registry, d *Decomposer, m *Merger) *Orchestrator {
	return &Orchestrator{
		registry:   reg,
		decomposer: d,
		merger:     m,
	}
}

func (o *Orchestrator) Dispatch(task *Task) ([]*Task, error) {
	steps, err := o.decomposer.Decompose(task)
	if err != nil {
		return nil, fmt.Errorf("decompose: %w", err)
	}

	var results []*Task
	for _, step := range steps {
		agents := o.registry.FindByRole(string(step.Type))
		if len(agents) == 0 {
			step.Status = StatusFailed
			step.Error = fmt.Sprintf("no agent found for role %q", step.Type)
			results = append(results, step)
			return results, fmt.Errorf("%s", step.Error)
		}

		agent := agents[0]
		step.Status = StatusRunning
		step.AgentID = agent.Name

		result, err := agent.Execute(step.Content)
		if err != nil {
			step.Status = StatusFailed
			step.Error = err.Error()
			results = append(results, step)
			return results, err
		}

		step.Status = StatusCompleted
		step.Result = result
		results = append(results, step)
	}

	return results, nil
}

func (o *Orchestrator) DispatchAndMerge(task *Task) (string, error) {
	results, err := o.Dispatch(task)
	if err != nil {
		return "", err
	}
	return o.merger.Merge(results), nil
}
