package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ReadFileTool struct{}

func (t *ReadFileTool) Name() string { return "read_file" }

func (t *ReadFileTool) Description() string { return "Read the content of a file" }

func (t *ReadFileTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"path": {Type: "string", Description: "Path to the file"},
		},
		Required: []string{"path"},
	}
}

func (t *ReadFileTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ToolResult{Error: "path parameter is required"}
	}

	if ctx.SandboxDir != "" {
		sandbox := filepath.Clean(ctx.SandboxDir)
		if !filepath.IsAbs(path) {
			path = filepath.Join(sandbox, path)
		}
		target := filepath.Clean(path)
		rel, err := filepath.Rel(sandbox, target)
		if err != nil || strings.HasPrefix(rel, "..") {
			return ToolResult{Error: fmt.Sprintf("path %q is outside sandbox %q", path, ctx.SandboxDir)}
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("read file: %s", err)}
	}
	return ToolResult{Success: true, Data: string(data)}
}
