package tool

import (
	"fmt"
	"os"
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
	data, err := os.ReadFile(path)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("read file: %s", err)}
	}
	return ToolResult{Success: true, Data: string(data)}
}
