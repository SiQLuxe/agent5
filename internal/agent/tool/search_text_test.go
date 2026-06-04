package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSearchTextToolName(t *testing.T) {
	tool := &SearchTextTool{}
	if tool.Name() != "search_text" {
		t.Fatalf("expected name 'search_text', got %s", tool.Name())
	}
}

func TestSearchTextToolExecute(t *testing.T) {
	// Use a temp dir so we don't pollute the workspace
	prev, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(prev)

	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\nfunc Hello()"), 0644)
	os.WriteFile(filepath.Join(dir, "b.go"), []byte("package b\nfunc World()"), 0644)

	tool := &SearchTextTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"pattern": "Hello",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}

	matches, ok := result.Data.([]SearchMatch)
	if !ok {
		t.Fatalf("expected []SearchMatch, got %T", result.Data)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
}

func TestSearchTextToolNoMatch(t *testing.T) {
	prev, _ := os.Getwd()
	dir := t.TempDir()
	os.Chdir(dir)
	defer os.Chdir(prev)

	os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\nfunc Foo()"), 0644)

	tool := &SearchTextTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"pattern": "NonExistent",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	matches := result.Data.([]SearchMatch)
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(matches))
	}
}
