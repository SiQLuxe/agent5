package runtime

import (
	"testing"

	"github.com/example/agent-tui/internal/agent/tool"
)

type mockReadTool struct{}

func (m *mockReadTool) Name() string { return "read_file" }
func (m *mockReadTool) Description() string { return "read a file" }
func (m *mockReadTool) Schema() tool.ToolSchema {
	return tool.ToolSchema{
		Parameters: map[string]tool.ParamSchema{
			"path": {Type: "string", Description: "file path"},
		},
		Required: []string{"path"},
	}
}
func (m *mockReadTool) Execute(ctx tool.ToolContext, params map[string]interface{}) tool.ToolResult {
	return tool.ToolResult{Success: true, Data: "file content"}
}

func TestAgentSimpleFinal(t *testing.T) {
	reg := tool.NewRegistry()
	reg.Register(&mockReadTool{})

	mock := NewMockLLMClient([]LLMResponse{
		{Type: "final", Content: "Task complete"},
	})

	agent := NewAgent(Config{
		Name:         "test",
		Model:        "test-model",
		SystemPrompt: "You are a test agent",
	}, reg, mock)

	result, err := agent.Execute("do something")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result != "Task complete" {
		t.Fatalf("expected 'Task complete', got %s", result)
	}
}

func TestAgentToolCallThenFinal(t *testing.T) {
	reg := tool.NewRegistry()
	reg.Register(&mockReadTool{})

	mock := NewMockLLMClient([]LLMResponse{
		{
			Type: "tool_call",
			ToolCall: &ToolCall{
				Name:      "read_file",
				Arguments: map[string]interface{}{"path": "/tmp/test.txt"},
			},
		},
		{Type: "final", Content: "Done reading"},
	})

	agent := NewAgent(Config{
		Name:         "test",
		Model:        "test-model",
		SystemPrompt: "You are a test agent",
	}, reg, mock)

	result, err := agent.Execute("read a file")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result != "Done reading" {
		t.Fatalf("expected 'Done reading', got %s", result)
	}
	if len(agent.Logger.Entries()) < 2 {
		t.Fatal("expected at least 2 log entries")
	}
}
