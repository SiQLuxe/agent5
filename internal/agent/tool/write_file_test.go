package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileToolName(t *testing.T) {
	tool := &WriteFileTool{}
	if tool.Name() != "write_file" {
		t.Fatalf("expected name 'write_file', got %s", tool.Name())
	}
}

func TestWriteFileToolExecute(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "new.txt")

	tool := &WriteFileTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"path":    file,
		"content": "hello world",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "hello world" {
		t.Fatalf("expected 'hello world', got %s", string(data))
	}
}

func TestWriteFileToolOverwrite(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "overwrite.txt")
	os.WriteFile(file, []byte("old"), 0644)

	tool := &WriteFileTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"path":    file,
		"content": "new content",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "new content" {
		t.Fatalf("expected 'new content', got %s", string(data))
	}
}

func TestWriteFileToolMissingParams(t *testing.T) {
	tool := &WriteFileTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{})
	if result.Success {
		t.Fatal("expected failure for missing params")
	}
}
