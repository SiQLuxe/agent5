package tool

import (
	"testing"
)

func TestTaskToolName(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr}
	if tool.Name() != "task" {
		t.Errorf("expected name 'task', got %s", tool.Name())
	}
}

func TestTaskToolSchema(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr}
	schema := tool.Schema()
	for _, name := range []string{"description", "prompt", "subagent_type"} {
		if _, ok := schema.Parameters[name]; !ok {
			t.Errorf("missing required parameter: %s", name)
		}
	}
}

func TestTaskToolExecute(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr}

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
	if data == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestTaskToolExecuteMissingFields(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr}

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
	mgr := NewSubagentManager(1) // max 1
	tool := &TaskTool{Manager: mgr}

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
