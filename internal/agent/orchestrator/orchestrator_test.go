package orchestrator

import (
	"testing"

	"github.com/example/agent-tui/internal/agent/runtime"
	"github.com/example/agent-tui/internal/agent/tool"
)

func TestOrchestratorDispatchSingle(t *testing.T) {
	reg := NewRegistry()
	mockLLM := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{Type: "final", Content: "done"},
	})
	agent := runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), mockLLM)
	reg.Register("coder", agent, "task_code")

	orch := NewOrchestrator(reg, NewDecomposer(), NewMerger())

	task := &Task{ID: "t1", Type: TaskCode, Content: "write code"}
	results, err := orch.Dispatch(task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestOrchestratorDispatchSplit(t *testing.T) {
	reg := NewRegistry()

	mockLLM := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{Type: "final", Content: "done"},
	})
	planner := runtime.NewAgent(runtime.Config{Name: "planner"}, tool.NewRegistry(), mockLLM)
	coder := runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), mockLLM)
	reviewer := runtime.NewAgent(runtime.Config{Name: "reviewer"}, tool.NewRegistry(), mockLLM)

	reg.Register("planner", planner, "task_analyze", "task_design")
	reg.Register("coder", coder, "task_code")
	reg.Register("reviewer", reviewer, "task_review")

	d := NewDecomposer()
	d.Register(TaskDesign, DefaultSequentialStrategy)

	orch := NewOrchestrator(reg, d, NewMerger())

	task := &Task{ID: "t1", Type: TaskDesign, Content: "add login feature"}
	results, err := orch.Dispatch(task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if len(results) != 4 {
		t.Fatalf("expected 4 results (decomposed), got %d", len(results))
	}
}

func TestOrchestratorNoAgent(t *testing.T) {
	reg := NewRegistry()
	orch := NewOrchestrator(reg, NewDecomposer(), NewMerger())

	task := &Task{ID: "t1", Type: TaskCode, Content: "write code"}
	_, err := orch.Dispatch(task)
	if err == nil {
		t.Fatal("expected error when no agent for task type")
	}
}
