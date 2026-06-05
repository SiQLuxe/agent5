package runtime

import (
	"testing"

	"github.com/example/agent-tui/internal/agent/session"
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

	sm := session.NewManager()
	sid := sm.CreateSession("test")

	agent := NewAgent(Config{
		Name:         "test",
		Model:        "test-model",
		SystemPrompt: "You are a test agent",
	}, reg, mock, sm)

	result, err := agent.Execute(sid, "do something")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result != "Task complete" {
		t.Fatalf("expected 'Task complete', got %s", result)
	}

	ctx := sm.GetContext(sid, 0)
	if len(ctx) != 2 {
		t.Fatalf("expected 2 messages in session, got %d", len(ctx))
	}
	if ctx[0].Role != "user" || ctx[0].Content != "do something" {
		t.Fatalf("unexpected first message: %+v", ctx[0])
	}
	if ctx[1].Role != "assistant" || ctx[1].Content != "Task complete" {
		t.Fatalf("unexpected second message: %+v", ctx[1])
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

	sm := session.NewManager()
	sid := sm.CreateSession("test")

	agent := NewAgent(Config{
		Name:         "test",
		Model:        "test-model",
		SystemPrompt: "You are a test agent",
	}, reg, mock, sm)

	result, err := agent.Execute(sid, "read a file")
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
