package tool

import (
	"context"
)

type ToolSchema struct {
	Parameters map[string]ParamSchema
	Required   []string
}

type ParamSchema struct {
	Type        string
	Description string
	Enum        []string
}

type ToolContext struct {
	Context context.Context
}

type ToolResult struct {
	Success bool
	Data    interface{}
	Error   string
}

type Tool interface {
	Name() string
	Description() string
	Schema() ToolSchema
	Execute(ctx ToolContext, params map[string]interface{}) ToolResult
}
