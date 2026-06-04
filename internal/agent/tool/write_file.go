package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WriteFileTool struct{}

func (t *WriteFileTool) Name() string { return "write_file" }

func (t *WriteFileTool) Description() string { return "Write or create a file" }

func (t *WriteFileTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"path":    {Type: "string", Description: "Path to the file"},
			"content": {Type: "string", Description: "Content to write"},
		},
		Required: []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	path, _ := params["path"].(string)
	content, _ := params["content"].(string)
	if path == "" {
		return ToolResult{Error: "path parameter is required"}
	}

	if ctx.SandboxDir != "" {
		if filepath.IsAbs(path) {
			if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(ctx.SandboxDir)) {
				return ToolResult{Error: fmt.Sprintf("path %q is outside sandbox %q", path, ctx.SandboxDir)}
			}
		} else {
			path = filepath.Join(ctx.SandboxDir, path)
		}
	}

	dir := filepath.Dir(path)
	if dir != "." {
		os.MkdirAll(dir, 0755)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return ToolResult{Error: fmt.Sprintf("write file: %s", err)}
	}
	return ToolResult{Success: true, Data: fmt.Sprintf("wrote %d bytes to %s", len(content), path)}
}
