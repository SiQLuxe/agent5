package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileToolName(t *testing.T) {
	tool := &ReadFileTool{}
	if tool.Name() != "read_file" {
		t.Fatalf("expected name 'read_file', got %s", tool.Name())
	}
}

func TestReadFileToolSchema(t *testing.T) {
	tool := &ReadFileTool{}
	schema := tool.Schema()
	if _, ok := schema.Parameters["path"]; !ok {
		t.Fatal("expected 'path' parameter")
	}
	if len(schema.Required) != 1 || schema.Required[0] != "path" {
		t.Fatal("expected 'path' as required")
	}
}

func TestReadFileToolExecute(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	os.WriteFile(file, []byte("hello world"), 0644)

	tool := &ReadFileTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"path": file,
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Data.(string) != "hello world" {
		t.Fatalf("expected 'hello world', got %v", result.Data)
	}
}

func TestReadFileToolMissingFile(t *testing.T) {
	tool := &ReadFileTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"path": "/nonexistent/file.txt",
	})
	if result.Success {
		t.Fatal("expected failure for missing file")
	}
	if result.Error == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestReadFileToolMissingParam(t *testing.T) {
	tool := &ReadFileTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{})
	if result.Success {
		t.Fatal("expected failure for missing param")
	}
}

func TestReadFileSandboxWithin(t *testing.T) {
	sandbox := t.TempDir()
	testFile := filepath.Join(sandbox, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx := ToolContext{SandboxDir: sandbox}
	tool := &ReadFileTool{}
	result := tool.Execute(ctx, map[string]interface{}{"path": testFile})
	if result.Error != "" {
		t.Fatalf("expected no error, got: %s", result.Error)
	}
}

func TestReadFileSandboxRejectsOutside(t *testing.T) {
	sandbox := t.TempDir()
	outsideDir := t.TempDir()
	testFile := filepath.Join(outsideDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx := ToolContext{SandboxDir: sandbox}
	tool := &ReadFileTool{}
	result := tool.Execute(ctx, map[string]interface{}{"path": testFile})
	if result.Error == "" {
		t.Fatal("expected error for path outside sandbox, got none")
	}
}

func TestReadFileSandboxEscape(t *testing.T) {
	sandbox := t.TempDir()
	parent := filepath.Dir(sandbox)
	escapePath := filepath.Join(parent, filepath.Base(sandbox)+"-escape", "test.txt")
	if err := os.MkdirAll(filepath.Dir(escapePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(escapePath, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx := ToolContext{SandboxDir: sandbox}
	tool := &ReadFileTool{}
	result := tool.Execute(ctx, map[string]interface{}{"path": escapePath})
	if result.Error == "" {
		t.Fatal("expected error for sandbox escape attempt, got none")
	}
}

func TestReadFileSandboxRelativePath(t *testing.T) {
	sandbox := t.TempDir()
	testFile := filepath.Join(sandbox, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	ctx := ToolContext{SandboxDir: sandbox}
	tool := &ReadFileTool{}
	result := tool.Execute(ctx, map[string]interface{}{"path": "test.txt"})
	if result.Error != "" {
		t.Fatalf("expected no error, got: %s", result.Error)
	}
}
