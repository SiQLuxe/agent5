package tool

import (
	"testing"
)

func TestTaskToolName(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr, Depth: 3}
	if tool.Name() != "task" {
		t.Errorf("expected name 'task', got %s", tool.Name())
	}
}

func TestTaskToolSchema(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr, Depth: 3}
	schema := tool.Schema()
	if schema.Parameters == nil {
		t.Fatal("expected non-nil parameters")
	}
	for _, name := range []string{"description", "prompt", "subagent_type"} {
		if _, ok := schema.Parameters[name]; !ok {
			t.Errorf("missing required parameter: %s", name)
		}
	}
	if _, ok := schema.Parameters["background"]; !ok {
		t.Error("missing optional parameter: background")
	}
}

type mockRunner struct {
	result string
	err    error
}

func (r *mockRunner) Run(sessionID, task, systemPrompt string, tools *Registry) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return r.result, nil
}

func TestTaskToolExecute(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{
		Manager: mgr,
		Runner:  &mockRunner{result: "task completed"},
		Depth:   3,
	}

	result := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "test task",
		"prompt":        "do something",
		"subagent_type": "general",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	data, ok := result.Data.(string)
	if !ok {
		t.Fatalf("expected string data, got %T", result.Data)
	}
	if data != "task completed" {
		t.Errorf("expected 'task completed', got %s", data)
	}
}

func TestTaskToolExecuteMissingFields(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr, Depth: 3}

	result := tool.Execute(ToolContext{}, map[string]interface{}{
		"description": "test",
	})
	if result.Success {
		t.Fatal("expected failure with missing prompt")
	}
	if result.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestTaskToolExecuteMaxConcurrent(t *testing.T) {
	mgr := NewSubagentManager(1)
	tool := &TaskTool{
		Manager: mgr,
		Runner:  &mockRunner{result: "ok"},
		Depth:   3,
	}

	r1 := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "task 1",
		"prompt":        "do 1",
		"subagent_type": "general",
	})
	if !r1.Success {
		t.Fatalf("first task should succeed: %s", r1.Error)
	}

	r2 := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "task 2",
		"prompt":        "do 2",
		"subagent_type": "general",
	})
	if r2.Success {
		t.Fatal("second task should fail (concurrency limit)")
	}
}

func TestTaskToolExecuteBackground(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{
		Manager: mgr,
		Runner:  &mockRunner{result: "background done"},
		Depth:   3,
	}

	result := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "bg task",
		"prompt":        "do bg work",
		"subagent_type": "general",
		"background":    true,
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	data, ok := result.Data.(string)
	if !ok {
		t.Fatalf("expected string data, got %T", result.Data)
	}
	if data == "" {
		t.Fatal("expected non-empty background result")
	}
}

func TestTaskToolExecuteDepthLimit(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{
		Manager: mgr,
		Runner:  &mockRunner{result: "ok"},
		Depth:   0,
	}

	result := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "deep task",
		"prompt":        "go deeper",
		"subagent_type": "general",
	})
	if result.Success {
		t.Fatal("expected failure due to depth limit")
	}
}
