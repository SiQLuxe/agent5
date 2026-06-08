package tool

import (
	"context"
	"fmt"
	"testing"

	"github.com/example/agent-tui/internal/agent/tool/search"
)

type mockWebSearchProvider struct {
	name    string
	results []search.WebSearchResult
	err     error
}

func (m *mockWebSearchProvider) Search(_ context.Context, query string, count int) ([]search.WebSearchResult, error) {
	return m.results, m.err
}

func (m *mockWebSearchProvider) Name() string { return m.name }

func TestWebSearchToolName(t *testing.T) {
	tool := NewWebSearchTool(nil)
	if tool.Name() != "web_search" {
		t.Fatalf("expected name 'web_search', got %s", tool.Name())
	}
}

func TestWebSearchToolDescription(t *testing.T) {
	tool := NewWebSearchTool(nil)
	if tool.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestWebSearchToolSchema(t *testing.T) {
	tool := NewWebSearchTool(nil)
	schema := tool.Schema()
	if _, ok := schema.Parameters["query"]; !ok {
		t.Fatal("expected 'query' parameter")
	}
	if _, ok := schema.Parameters["count"]; !ok {
		t.Fatal("expected 'count' parameter")
	}
	if len(schema.Required) != 1 || schema.Required[0] != "query" {
		t.Fatal("expected 'query' as required")
	}
}

func TestWebSearchToolExecute(t *testing.T) {
	provider := &mockWebSearchProvider{
		name: "mock",
		results: []search.WebSearchResult{
			{Title: "Test Title", URL: "http://example.com", Snippet: "Test snippet", Content: "Full content"},
		},
	}
	tool := NewWebSearchTool(provider)
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"query": "test query",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	results, ok := result.Data.([]search.WebSearchResult)
	if !ok {
		t.Fatalf("expected []search.WebSearchResult, got %T", result.Data)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "Test Title" {
		t.Fatalf("expected 'Test Title', got %q", results[0].Title)
	}
	if results[0].URL != "http://example.com" {
		t.Fatalf("expected 'http://example.com', got %q", results[0].URL)
	}
	if results[0].Snippet != "Test snippet" {
		t.Fatalf("expected 'Test snippet', got %q", results[0].Snippet)
	}
	if results[0].Content != "Full content" {
		t.Fatalf("expected 'Full content', got %q", results[0].Content)
	}
}

func TestWebSearchToolProviderError(t *testing.T) {
	provider := &mockWebSearchProvider{
		name: "mock",
		err:  fmt.Errorf("provider error"),
	}
	tool := NewWebSearchTool(provider)
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"query": "test query",
	})
	if result.Success {
		t.Fatal("expected failure for provider error")
	}
	if result.Error == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestWebSearchToolMissingQuery(t *testing.T) {
	tool := NewWebSearchTool(nil)
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{})
	if result.Success {
		t.Fatal("expected failure for missing query")
	}
}

func TestWebSearchToolEmptyQuery(t *testing.T) {
	tool := NewWebSearchTool(nil)
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"query": "",
	})
	if result.Success {
		t.Fatal("expected failure for empty query")
	}
}

func TestWebSearchToolCustomCount(t *testing.T) {
	results := make([]search.WebSearchResult, 3)
	for i := range results {
		results[i] = search.WebSearchResult{Title: fmt.Sprintf("Result %d", i+1)}
	}
	provider := &mockWebSearchProvider{
		name:    "mock",
		results: results,
	}
	tool := NewWebSearchTool(provider)
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"query": "test",
		"count": float64(3),
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	data := result.Data.([]search.WebSearchResult)
	if len(data) != 3 {
		t.Fatalf("expected 3 results, got %d", len(data))
	}
}

func TestWebSearchToolDefaultCount(t *testing.T) {
	results := make([]search.WebSearchResult, 5)
	for i := range results {
		results[i] = search.WebSearchResult{Title: fmt.Sprintf("Result %d", i+1)}
	}
	provider := &mockWebSearchProvider{
		name:    "mock",
		results: results,
	}
	tool := NewWebSearchTool(provider)
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"query": "test",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	data := result.Data.([]search.WebSearchResult)
	if len(data) != 5 {
		t.Fatalf("expected 5 results, got %d", len(data))
	}
}
