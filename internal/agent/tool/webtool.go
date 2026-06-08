package tool

import (
	"fmt"

	"github.com/example/agent-tui/internal/agent/tool/search"
)

type WebSearchTool struct {
	provider search.WebSearchProvider
}

func NewWebSearchTool(provider search.WebSearchProvider) *WebSearchTool {
	return &WebSearchTool{provider: provider}
}

func (t *WebSearchTool) Name() string { return "web_search" }

func (t *WebSearchTool) Description() string {
	return "Search the web for information. Returns a list of results with title, URL, and snippet."
}

func (t *WebSearchTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"query": {
				Type:        "string",
				Description: "The search query",
			},
			"count": {
				Type:        "integer",
				Description: "Number of results to return (default 5)",
			},
		},
		Required: []string{"query"},
	}
}

func (t *WebSearchTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return ToolResult{Error: "web_search: query is required"}
	}

	count := 5
	if c, ok := params["count"].(float64); ok {
		count = int(c)
	}

	results, err := t.provider.Search(ctx.Context, query, count)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("web_search: %s", err)}
	}

	return ToolResult{Success: true, Data: results}
}
