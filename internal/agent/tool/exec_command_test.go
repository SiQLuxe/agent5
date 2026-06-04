package tool

import (
	"context"
	"testing"
)

func TestExecCommandToolName(t *testing.T) {
	tool := &ExecCommandTool{}
	if tool.Name() != "exec_command" {
		t.Fatalf("expected name 'exec_command', got %s", tool.Name())
	}
}

func TestExecCommandToolEcho(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping exec test in short mode")
	}
	tool := &ExecCommandTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"command": "echo hello",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	out, ok := result.Data.(CmdOutput)
	if !ok {
		t.Fatalf("expected CmdOutput, got %T", result.Data)
	}
	if out.Stdout != "hello\n" && out.Stdout != "hello" {
		t.Fatalf("expected 'hello', got %q", out.Stdout)
	}
}

func TestExecCommandToolFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping exec test in short mode")
	}
	tool := &ExecCommandTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"command": "exit 42",
	})
	if result.Success {
		t.Fatal("expected failure for exit 42")
	}
}

func TestExecCommandToolMissingParam(t *testing.T) {
	tool := &ExecCommandTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{})
	if result.Success {
		t.Fatal("expected failure for missing param")
	}
}
