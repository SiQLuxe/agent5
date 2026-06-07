package orchestrator

import (
	"testing"

	"github.com/example/agent-tui/internal/agent/runtime"
	"github.com/example/agent-tui/internal/agent/session"
	"github.com/example/agent-tui/internal/agent/tool"
)

func TestOrchestratorDispatchSingle(t *testing.T) {
	reg := NewRegistry()
	mockLLM := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{Type: "final", Content: "done"},
	})
	sm := session.NewManager()
	agent := runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), mockLLM, sm)
	reg.Register("coder", agent, "task_code")

	orch := NewOrchestrator(reg, NewDecomposer(), NewMerger())

	task := &Task{ID: "t1", Type: TaskCode, Content: "write code"}
	results, err := orch.Dispatch("test-session", task)
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
	sm := session.NewManager()
	planner := runtime.NewAgent(runtime.Config{Name: "planner"}, tool.NewRegistry(), mockLLM, sm)
	coder := runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), mockLLM, sm)
	reviewer := runtime.NewAgent(runtime.Config{Name: "reviewer"}, tool.NewRegistry(), mockLLM, sm)

	reg.Register("planner", planner, "task_analyze", "task_design")
	reg.Register("coder", coder, "task_code")
	reg.Register("reviewer", reviewer, "task_review")

	d := NewDecomposer()
	d.Register(TaskDesign, DefaultSequentialStrategy)

	orch := NewOrchestrator(reg, d, NewMerger())

	task := &Task{ID: "t1", Type: TaskDesign, Content: "add login feature"}
	results, err := orch.Dispatch("test-session", task)
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
	_, err := orch.Dispatch("test-session", task)
	if err == nil {
		t.Fatal("expected error when no agent for task type")
	}
}

func TestOrchestratorDispatchConcurrent(t *testing.T) {
	reg := NewRegistry()

	mockLLM := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{Type: "final", Content: "analyze done"},
		{Type: "final", Content: "code done"},
		{Type: "final", Content: "review done"},
	})
	sm := session.NewManager()
	analyzer := runtime.NewAgent(runtime.Config{Name: "analyzer"}, tool.NewRegistry(), mockLLM, sm)
	coder := runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), mockLLM, sm)
	reviewer := runtime.NewAgent(runtime.Config{Name: "reviewer"}, tool.NewRegistry(), mockLLM, sm)

	reg.Register("analyzer", analyzer, "task_analyze")
	reg.Register("coder", coder, "task_code")
	reg.Register("reviewer", reviewer, "task_review")

	d := NewDecomposer()
	d.Register(TaskDesign, func(task *Task) ([]*Task, error) {
		return []*Task{
			{ID: "t-analyze", Type: TaskAnalyze, Content: "analyze: " + task.Content},
			{ID: "t-code", Type: TaskCode, Content: "code: " + task.Content},
			{ID: "t-review", Type: TaskReview, Content: "review: " + task.Content},
		}, nil
	})

	orch := NewOrchestrator(reg, d, NewMerger())
	orch.MaxConcurrent = 3

	task := &Task{ID: "t1", Type: TaskDesign, Content: "add login feature"}
	results, err := orch.DispatchConcurrent("test-session", task)
	if err != nil {
		t.Fatalf("DispatchConcurrent failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Status != StatusCompleted {
			t.Errorf("expected completed status for %s, got %s", r.ID, r.Status)
		}
	}
}

func TestDispatchConcurrentPartialFailure(t *testing.T) {
	reg := NewRegistry()

	sm := session.NewManager()
	okLLM := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{Type: "final", Content: "analysis ok"},
	})
	reviewerLLM := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{Type: "final", Content: "review ok"},
	})

	analyzer := runtime.NewAgent(runtime.Config{Name: "analyzer"}, tool.NewRegistry(), okLLM, sm)
	reviewer := runtime.NewAgent(runtime.Config{Name: "reviewer"}, tool.NewRegistry(), reviewerLLM, sm)

	reg.Register("analyzer", analyzer, "task_analyze")
	reg.Register("reviewer", reviewer, "task_review")

	d := NewDecomposer()
	d.Register(TaskDesign, func(task *Task) ([]*Task, error) {
		return []*Task{
			{ID: "t-analyze", Type: TaskAnalyze, Content: "analyze"},
			{ID: "t-code", Type: TaskCode, Content: "code"},
			{ID: "t-review", Type: TaskReview, Content: "review"},
		}, nil
	})

	orch := NewOrchestrator(reg, d, NewMerger())
	orch.MaxConcurrent = 3

	task := &Task{ID: "t1", Type: TaskDesign, Content: "add login"}
	results, err := orch.DispatchConcurrent("test-session", task)
	if err == nil {
		t.Fatal("expected error for partial failure")
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	var failed bool
	for _, r := range results {
		if r.Status == StatusFailed {
			failed = true
			if r.Error == "" {
				t.Errorf("failed task %s has empty Error field", r.ID)
			}
		}
	}
	if !failed {
		t.Fatal("expected at least one failed subtask")
	}
}
