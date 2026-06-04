package e2e

import (
	"strings"
	"testing"

	"github.com/example/agent-tui/internal/agent/orchestrator"
	"github.com/example/agent-tui/internal/agent/runtime"
	"github.com/example/agent-tui/internal/agent/tool"
	"github.com/example/agent-tui/internal/ui"
)

func TestEndToEnd_ChatFlow(t *testing.T) {
	app := ui.NewApp()
	app.AddWelcomeMessage()

	app.AddChatMessage("user", "Hello")

	messages := app.GetChatMessages()
	if len(messages) < 1 {
		t.Error("Expected at least 1 message")
	}
}

func TestEndToEnd_MultipleMessages(t *testing.T) {
	app := ui.NewApp()
	app.AddWelcomeMessage()

	app.AddChatMessage("user", "First")
	app.AddChatMessage("assistant", "Second")

	messages := app.GetChatMessages()
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}
}

func TestEndToEnd_ComposerInput(t *testing.T) {
	app := ui.NewApp()

	app.SetComposerInput("Test input")
	if app.GetComposerInput() != "Test input" {
		t.Errorf("Expected 'Test input', got %s", app.GetComposerInput())
	}
}

func TestEndToEnd_OrchestratorDispatch(t *testing.T) {
	// Direct test of orchestration with ReAct loop (no TUI event loop needed)
	mock := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{Type: "final", Content: "Feature implemented"},
	})
	agent := runtime.NewAgent(runtime.Config{Name: "coder", SystemPrompt: "test"}, tool.NewRegistry(), mock)
	agentReg := orchestrator.NewRegistry()
	agentReg.Register("coder", agent, "task_execute")

	orch := orchestrator.NewOrchestrator(agentReg, orchestrator.NewDecomposer(), orchestrator.NewMerger())

	task := &orchestrator.Task{
		ID:      "test-1",
		Type:    orchestrator.TaskExecute,
		Content: "implement login feature",
	}

	results, err := orch.Dispatch(task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Result != "Feature implemented" {
		t.Fatalf("expected 'Feature implemented', got %s", results[0].Result)
	}
}

func TestEndToEnd_AgentReActSteps(t *testing.T) {
	// Direct test of ReAct loop through orchestrator (no TUI event loop needed)
	mock := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{
			Type: "tool_call",
			ToolCall: &runtime.ToolCall{
				Name:      "read_file",
				Arguments: map[string]interface{}{"path": "/tmp/test.go"},
			},
		},
		{Type: "final", Content: "Found and analyzed the file"},
	})

	reg := tool.NewRegistry()
	reg.Register(&mockReadTool{})
	agent := runtime.NewAgent(runtime.Config{Name: "analyzer", SystemPrompt: "test"}, reg, mock)
	agentReg := orchestrator.NewRegistry()
	agentReg.Register("analyzer", agent, "task_execute")

	orch := orchestrator.NewOrchestrator(agentReg, orchestrator.NewDecomposer(), orchestrator.NewMerger())

	task := &orchestrator.Task{
		ID:      "test-2",
		Type:    orchestrator.TaskExecute,
		Content: "analyze the codebase",
	}

	results, err := orch.Dispatch(task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	final := results[0]
	if !strings.Contains(final.Result, "analyzed") {
		t.Fatalf("expected result containing 'analyzed', got: %s", final.Result)
	}
	if len(agent.Logger.Entries()) < 2 {
		t.Fatalf("expected >= 2 log entries (tool_call + final), got %d", len(agent.Logger.Entries()))
	}
}

// mockReadTool used by TestEndToEnd_AgentReActSteps
type mockReadTool struct{}

func (m *mockReadTool) Name() string                     { return "read_file" }
func (m *mockReadTool) Description() string              { return "read a file" }
func (m *mockReadTool) Schema() tool.ToolSchema           { return tool.ToolSchema{} }
func (m *mockReadTool) Execute(ctx tool.ToolContext, params map[string]interface{}) tool.ToolResult {
	return tool.ToolResult{Success: true, Data: "file content"}
}