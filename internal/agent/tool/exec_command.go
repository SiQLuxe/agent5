package tool

import (
	"fmt"
	"os/exec"
	"strings"
)

type CmdOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type ExecCommandTool struct{}

func (t *ExecCommandTool) Name() string { return "exec_command" }

func (t *ExecCommandTool) Description() string {
	return "Execute a shell command and return output"
}

func (t *ExecCommandTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"command": {Type: "string", Description: "Shell command to execute"},
			"timeout": {Type: "integer", Description: "Timeout in seconds (default 30)"},
		},
		Required: []string{"command"},
	}
}

func (t *ExecCommandTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	cmdStr, _ := params["command"].(string)
	if cmdStr == "" {
		return ToolResult{Error: "command parameter is required"}
	}

	cmd := exec.CommandContext(ctx.Context, "sh", "-c", cmdStr)
	if ctx.SandboxDir != "" {
		cmd.Dir = ctx.SandboxDir
	}
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return ToolResult{Error: fmt.Sprintf("exec: %s", err)}
		}
	}

	out := CmdOutput{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}

	if exitCode != 0 {
		return ToolResult{
			Success: false,
			Data:    out,
			Error:   fmt.Sprintf("exit code %d", exitCode),
		}
	}
	return ToolResult{Success: true, Data: out}
}
