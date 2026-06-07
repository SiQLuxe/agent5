package orchestrator

import (
	"fmt"
)

type Orchestrator struct {
	registry      *Registry
	decomposer    *Decomposer
	merger        *Merger
	MaxConcurrent int
}

func NewOrchestrator(reg *Registry, d *Decomposer, m *Merger) *Orchestrator {
	return &Orchestrator{
		registry:      reg,
		decomposer:    d,
		merger:        m,
		MaxConcurrent: 5,
	}
}

func (o *Orchestrator) Dispatch(sessionID string, task *Task) ([]*Task, error) {
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

		result, err := agent.Execute(sessionID, step.Content)
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

func (o *Orchestrator) DispatchConcurrent(sessionID string, task *Task) ([]*Task, error) {
	steps, err := o.decomposer.Decompose(task)
	if err != nil {
		return nil, fmt.Errorf("decompose: %w", err)
	}

	if len(steps) == 0 {
		return nil, nil
	}

	sem := make(chan struct{}, o.MaxConcurrent)
	type stepResult struct {
		step *Task
		err  error
	}
	resultCh := make(chan stepResult, len(steps))

	for _, step := range steps {
		sem <- struct{}{}
		go func(s *Task) {
			defer func() { <-sem }()
			agents := o.registry.FindByRole(string(s.Type))
			if len(agents) == 0 {
				s.Status = StatusFailed
				s.Error = fmt.Sprintf("no agent found for role %q", s.Type)
				resultCh <- stepResult{s, fmt.Errorf("%s", s.Error)}
				return
			}
			agent := agents[0]
			s.Status = StatusRunning
			s.AgentID = agent.Name
			res, err := agent.Execute(sessionID, s.Content)
			if err != nil {
				s.Status = StatusFailed
				s.Error = err.Error()
				resultCh <- stepResult{s, err}
				return
			}
			s.Status = StatusCompleted
			s.Result = res
			resultCh <- stepResult{s, nil}
		}(step)
	}

	// fill sem to capacity — waits for all goroutines to finish
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
	close(resultCh)

	var results []*Task
	var firstErr error
	for r := range resultCh {
		results = append(results, r.step)
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
	}

	return results, firstErr
}

func (o *Orchestrator) DispatchAndMerge(sessionID string, task *Task) (string, error) {
	results, err := o.Dispatch(sessionID, task)
	if err != nil {
		return "", err
	}
	return o.merger.Merge(results), nil
}

func (o *Orchestrator) DispatchStream(sessionID string, task *Task, onChunk func(string)) ([]*Task, error) {
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

		result, err := agent.ExecuteStream(sessionID, step.Content, onChunk)
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
