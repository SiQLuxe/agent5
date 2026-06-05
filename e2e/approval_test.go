package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/agent-tui/internal/agent/runtime"
	"github.com/example/agent-tui/internal/agent/tool"
)

func TestEndToEnd_ApprovalFlow(t *testing.T) {
	sandbox := t.TempDir()

	mockLLM := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{
			Type: "tool_call",
			ToolCall: &runtime.ToolCall{
				Name: "write_file",
				Arguments: map[string]interface{}{
					"path":    "hello.py",
					"content": `print("hello")`,
				},
			},
		},
		{
			Type:    "final",
			Content: "Done",
		},
	})

	toolReg := tool.NewRegistry()
	toolReg.Register(&tool.WriteFileTool{})

	approved := false
	approvalFn := func(toolName string, params map[string]interface{}, oldContent, newContent string) bool {
		approved = true
		if toolName != "write_file" {
			t.Errorf("expected tool_name 'write_file', got %q", toolName)
		}
		path, _ := params["path"].(string)
		if path != "hello.py" {
			t.Errorf("expected path 'hello.py', got %q", path)
		}
		return true
	}

	agent := runtime.NewAgent(runtime.Config{
		Name:         "test-coder",
		Model:        "mock",
		SystemPrompt: "You are a test agent",
		MaxReActLoop: 5,
		SandboxDir:   sandbox,
		ApprovalFn:   approvalFn,
	}, toolReg, mockLLM)

	_, err := agent.Execute("write a hello world program to hello.py")
	if err != nil {
		t.Fatalf("agent execute failed: %v", err)
	}

	if !approved {
		t.Fatal("expected approval callback to be invoked")
	}

	expectedPath := filepath.Join(sandbox, "hello.py")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("expected file %s to exist: %v", expectedPath, err)
	}
	if string(data) != `print("hello")` {
		t.Fatalf("expected content 'print(\"hello\")', got %q", string(data))
	}
}
