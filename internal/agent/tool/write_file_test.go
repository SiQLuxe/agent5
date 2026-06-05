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

func TestWriteFileToolRejected(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "rejected.txt")
	os.WriteFile(file, []byte("original"), 0644)

	tool := &WriteFileTool{}
	result := tool.Execute(ToolContext{
		Context: context.Background(),
		Approval: func(toolName string, params map[string]interface{}, oldContent, newContent string) bool {
			if toolName != "write_file" {
				t.Fatalf("expected tool name 'write_file', got %s", toolName)
			}
			if oldContent != "original" {
				t.Fatalf("expected old content 'original', got %s", oldContent)
			}
			if newContent != "new content" {
				t.Fatalf("expected new content 'new content', got %s", newContent)
			}
			return false
		},
	}, map[string]interface{}{
		"path":    file,
		"content": "new content",
	})
	if result.Success {
		t.Fatal("expected failure when approval rejects")
	}
	if result.Error != "file write rejected by user" {
		t.Fatalf("expected rejection error, got: %s", result.Error)
	}
	data, _ := os.ReadFile(file)
	if string(data) != "original" {
		t.Fatalf("file should not have been modified, got: %s", string(data))
	}
}

func TestWriteFileSandboxWithin(t *testing.T) {
	sandbox := t.TempDir()
	ctx := ToolContext{
		SandboxDir: sandbox,
	}
	tool := &WriteFileTool{}
	path := filepath.Join(sandbox, "hello.py")
	result := tool.Execute(ctx, map[string]interface{}{
		"path":    path,
		"content": `print("hello")`,
	})
	if result.Error != "" {
		t.Fatalf("expected no error, got: %s", result.Error)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `print("hello")` {
		t.Fatalf("expected 'print(\"hello\")', got %q", string(data))
	}
}

func TestWriteFileSandboxRejectsOutside(t *testing.T) {
	sandbox := t.TempDir()
	ctx := ToolContext{
		SandboxDir: sandbox,
	}
	tool := &WriteFileTool{}
	outsidePath := filepath.Join(t.TempDir(), "evil.py")
	result := tool.Execute(ctx, map[string]interface{}{
		"path":    outsidePath,
		"content": `print("evil")`,
	})
	if result.Error == "" {
		t.Fatal("expected error for path outside sandbox, got none")
	}
}

func TestWriteFileSandboxEscape(t *testing.T) {
	sandbox := t.TempDir()
	ctx := ToolContext{
		SandboxDir: sandbox,
	}
	tool := &WriteFileTool{}
	parent := filepath.Dir(sandbox)
	sibling := filepath.Join(parent, filepath.Base(sandbox)+"-escape", "evil.py")
	result := tool.Execute(ctx, map[string]interface{}{
		"path":    sibling,
		"content": `print("evil")`,
	})
	if result.Error == "" {
		t.Fatal("expected error for sandbox escape attempt, got none")
	}
}

func TestWriteFileSandboxRelativePath(t *testing.T) {
	sandbox := t.TempDir()
	ctx := ToolContext{
		SandboxDir: sandbox,
	}
	tool := &WriteFileTool{}
	result := tool.Execute(ctx, map[string]interface{}{
		"path":    "hello.py",
		"content": `print("hello")`,
	})
	if result.Error != "" {
		t.Fatalf("expected no error, got: %s", result.Error)
	}
	expectedPath := filepath.Join(sandbox, "hello.py")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist", expectedPath)
	}
}

func TestWriteFileToolApproved(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "approved.txt")

	approved := false
	tool := &WriteFileTool{}
	result := tool.Execute(ToolContext{
		Context: context.Background(),
		Approval: func(toolName string, params map[string]interface{}, oldContent, newContent string) bool {
			approved = true
			return true
		},
	}, map[string]interface{}{
		"path":    file,
		"content": "new content",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if !approved {
		t.Fatal("expected approval callback to be called")
	}
	data, _ := os.ReadFile(file)
	if string(data) != "new content" {
		t.Fatalf("expected 'new content', got %s", string(data))
	}
}
