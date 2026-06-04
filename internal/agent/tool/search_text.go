package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type SearchMatch struct {
	Path       string
	LineNumber int
	Content    string
}

type SearchTextTool struct{}

func (t *SearchTextTool) Name() string { return "search_text" }

func (t *SearchTextTool) Description() string { return "Search for text patterns in the codebase" }

func (t *SearchTextTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"pattern": {Type: "string", Description: "Text pattern to search (regex supported)"},
			"include": {Type: "string", Description: "File glob pattern (e.g. *.go)"},
		},
		Required: []string{"pattern"},
	}
}

func (t *SearchTextTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	pattern, _ := params["pattern"].(string)
	if pattern == "" {
		return ToolResult{Error: "pattern parameter is required"}
	}
	include, _ := params["include"].(string)

	re, err := regexp.Compile(pattern)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("invalid pattern: %s", err)}
	}

	root := ctx.SandboxDir
	if root == "" {
		root, _ = os.Getwd()
	}
	var matches []SearchMatch
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if include != "" {
			matched, _ := filepath.Match(include, filepath.Base(path))
			if !matched {
				return nil
			}
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		lines := regexp.MustCompile(`\n`).Split(string(data), -1)
		for i, line := range lines {
			if re.MatchString(line) {
				rel, _ := filepath.Rel(root, path)
				matches = append(matches, SearchMatch{
					Path:       rel,
					LineNumber: i + 1,
					Content:    line,
				})
			}
		}
		return nil
	})

	return ToolResult{Success: true, Data: matches}
}
