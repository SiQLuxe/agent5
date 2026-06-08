package tool

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestWebFetchToolName(t *testing.T) {
	tool := NewWebFetchTool()
	if tool.Name() != "web_fetch" {
		t.Fatalf("expected 'web_fetch', got %q", tool.Name())
	}
}

func TestWebFetchToolDescription(t *testing.T) {
	tool := NewWebFetchTool()
	if tool.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestWebFetchToolSchema(t *testing.T) {
	tool := NewWebFetchTool()
	schema := tool.Schema()
	if _, ok := schema.Parameters["url"]; !ok {
		t.Fatal("expected 'url' parameter")
	}
	if _, ok := schema.Parameters["selector"]; !ok {
		t.Fatal("expected 'selector' parameter")
	}
	if len(schema.Required) != 1 || schema.Required[0] != "url" {
		t.Fatal("expected 'url' as required")
	}
}

func TestWebFetchToolMissingURL(t *testing.T) {
	tool := NewWebFetchTool()
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{})
	if result.Success {
		t.Fatal("expected failure for missing URL")
	}
}

func TestWebFetchToolEmptyURL(t *testing.T) {
	tool := NewWebFetchTool()
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"url": "",
	})
	if result.Success {
		t.Fatal("expected failure for empty URL")
	}
}

func TestWebFetchToolInvalidURL(t *testing.T) {
	tool := NewWebFetchTool()
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"url": "not-a-valid-url:::",
	})
	if result.Success {
		t.Fatal("expected failure for invalid URL")
	}
}

func TestWebFetchToolUnsupportedScheme(t *testing.T) {
	tool := NewWebFetchTool()
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"url": "ftp://example.com/file",
	})
	if result.Success {
		t.Fatal("expected failure for unsupported scheme")
	}
}

func TestWebFetchToolHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tool := NewWebFetchTool()
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"url": server.URL,
	})
	if result.Success {
		t.Fatal("expected failure for HTTP error")
	}
}

func TestWebFetchToolSuccess(t *testing.T) {
	html := `<html><head><title>Test Page</title></head><body><p>Hello, world!</p></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	tool := NewWebFetchTool()
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"url": server.URL,
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	data, ok := result.Data.(WebFetchResult)
	if !ok {
		t.Fatalf("expected WebFetchResult, got %T", result.Data)
	}
	if data.Title != "Test Page" {
		t.Fatalf("expected 'Test Page', got %q", data.Title)
	}
	if data.Content == "" {
		t.Fatal("expected non-empty content")
	}
	if data.URL != server.URL {
		t.Fatalf("expected %q, got %q", server.URL, data.URL)
	}
}

func TestWebFetchToolCustomSelector(t *testing.T) {
	html := `<html><head><title>Page</title></head><body>
		<div id="main"><p>Main content</p></div>
		<div id="sidebar"><p>Sidebar content</p></div>
	</body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}))
	defer server.Close()

	tool := NewWebFetchTool()
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"url":      server.URL,
		"selector": "#main",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	data := result.Data.(WebFetchResult)
	if !strings.Contains(data.Content, "Main content") {
		t.Fatalf("expected content to contain 'Main content', got %q", data.Content)
	}
	if strings.Contains(data.Content, "Sidebar content") {
		t.Fatal("expected content NOT to contain 'Sidebar content'")
	}
}
