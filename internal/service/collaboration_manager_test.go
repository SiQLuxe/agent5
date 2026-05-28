package service

import (
	"testing"

	"github.com/example/agent-tui/internal/data/history"
)

func TestCollaborationManager(t *testing.T) {
	h := history.NewHistory("")
	o := NewTaskOrchestrator(h)
	cm := NewCollaborationManager(o)

	o.RegisterAgent(TaskAnalyze, &MockAgent{})
	o.RegisterAgent(TaskDesign, &MockAgent{})
	o.RegisterAgent(TaskCode, &MockAgent{})
	o.RegisterAgent(TaskReview, &MockAgent{})

	task := &Task{
		ID:      "test",
		Type:    TaskAnalyze,
		Content: "分析任务",
		Status:  StatusPending,
	}

	steps, err := cm.AutoDecompose(task)
	if err != nil {
		t.Errorf("AutoDecompose failed: %v", err)
	}

	if len(steps) != 4 {
		t.Errorf("Expected 4 steps, got %d", len(steps))
	}

	expectedTypes := []TaskType{TaskAnalyze, TaskDesign, TaskCode, TaskReview}
	for i, step := range steps {
		if step.Type != expectedTypes[i] {
			t.Errorf("Step %d: expected type %s, got %s", i, expectedTypes[i], step.Type)
		}
	}
}

func TestCollaborationManager_ExecuteWorkflow(t *testing.T) {
	h := history.NewHistory("")
	o := NewTaskOrchestrator(h)
	cm := NewCollaborationManager(o)

	o.RegisterAgent(TaskCode, &MockAgent{})

	steps := []*Task{
		{ID: "step1", Type: TaskCode, Content: "Task 1", Status: StatusPending},
		{ID: "step2", Type: TaskCode, Content: "Task 2", Status: StatusPending},
	}

	results, err := cm.ExecuteWorkflow(steps)
	if err != nil {
		t.Errorf("ExecuteWorkflow failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Status != StatusCompleted {
			t.Errorf("Expected completed status, got %s", result.Status)
		}
	}
}

func TestCollaborationManager_AutoDecomposeDesign(t *testing.T) {
	h := history.NewHistory("")
	o := NewTaskOrchestrator(h)
	cm := NewCollaborationManager(o)

	task := &Task{
		ID:      "design1",
		Type:    TaskDesign,
		Content: "设计任务",
		Status:  StatusPending,
	}

	steps, err := cm.AutoDecompose(task)
	if err != nil {
		t.Errorf("AutoDecompose failed: %v", err)
	}

	if len(steps) != 4 {
		t.Errorf("Expected 4 steps for design task, got %d", len(steps))
	}

	expectedTypes := []TaskType{TaskDesign, TaskCode, TaskReview, TaskExecute}
	for i, step := range steps {
		if step.Type != expectedTypes[i] {
			t.Errorf("Step %d: expected type %s, got %s", i, expectedTypes[i], step.Type)
		}
	}
}

func TestCollaborationManager_AutoDecomposeDefault(t *testing.T) {
	h := history.NewHistory("")
	o := NewTaskOrchestrator(h)
	cm := NewCollaborationManager(o)

	task := &Task{
		ID:      "exec1",
		Type:    TaskExecute,
		Content: "执行任务",
		Status:  StatusPending,
	}

	steps, err := cm.AutoDecompose(task)
	if err != nil {
		t.Errorf("AutoDecompose failed: %v", err)
	}

	if len(steps) != 1 {
		t.Errorf("Expected 1 step for default task type, got %d", len(steps))
	}

	if steps[0].ID != "exec1" {
		t.Errorf("Expected task ID 'exec1', got %s", steps[0].ID)
	}
}