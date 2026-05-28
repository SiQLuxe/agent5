package service

import (
	"testing"

	"github.com/example/agent-tui/internal/data/history"
)

func TestTaskOrchestrator(t *testing.T) {
	h := history.NewHistory("")
	o := NewTaskOrchestrator(h)

	mockAgent := &MockAgent{}
	o.RegisterAgent(TaskCode, mockAgent)

	task := &Task{
		ID:      "test1",
		Type:    TaskCode,
		Content: "Write a function",
	}

	err := o.Dispatch(task)
	if err != nil {
		t.Errorf("Dispatch failed: %v", err)
	}

	if task.Status != StatusCompleted {
		t.Errorf("Expected completed status, got %s", task.Status)
	}

	if task.Result != "mock result" {
		t.Errorf("Expected 'mock result', got %s", task.Result)
	}
}

func TestTaskOrchestrator_NoAgent(t *testing.T) {
	h := history.NewHistory("")
	o := NewTaskOrchestrator(h)

	task := &Task{
		ID:      "test2",
		Type:    TaskAnalyze,
		Content: "Analyze something",
		Status:  StatusPending,
	}

	err := o.Dispatch(task)
	if err == nil {
		t.Error("Expected error when no agent registered")
	}

	if task.Status != StatusPending {
		t.Errorf("Expected pending status when dispatch fails")
	}
}

func TestTaskOrchestrator_AddTask(t *testing.T) {
	h := history.NewHistory("")
	o := NewTaskOrchestrator(h)

	task := &Task{
		ID:      "test3",
		Type:    TaskCode,
		Content: "Test task",
	}

	o.AddTask(task)

	tasks := o.GetTasks()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task in queue, got %d", len(tasks))
	}

	if tasks[0].ID != "test3" {
		t.Errorf("Expected task ID 'test3', got %s", tasks[0].ID)
	}
}

func TestTaskOrchestrator_GetAgentForTask(t *testing.T) {
	h := history.NewHistory("")
	o := NewTaskOrchestrator(h)

	mockAgent := &MockAgent{}
	o.RegisterAgent(TaskDesign, mockAgent)

	agent, ok := o.GetAgentForTask(TaskDesign)
	if !ok {
		t.Error("Expected agent to be found")
	}

	if agent == nil {
		t.Error("Expected non-nil agent")
	}

	_, ok = o.GetAgentForTask(TaskReview)
	if ok {
		t.Error("Expected no agent for TaskReview")
	}
}

type MockAgent struct{}

func (m *MockAgent) Execute(task *Task) (string, error) {
	return "mock result", nil
}

func (m *MockAgent) GetRoleName() string {
	return "MockAgent"
}

func (m *MockAgent) GetSupportedTaskTypes() []TaskType {
	return []TaskType{TaskCode}
}