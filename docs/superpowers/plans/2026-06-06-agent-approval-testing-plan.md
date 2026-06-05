# Agent Approval Testing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add cross-platform sandbox path support, fix sandbox path traversal bug, add E2E approval integration test, and document manual TUI testing.

**Architecture:** Four independent changes: (1) OS-aware default SandboxDir in config loading, (2-3) sandbox traversal fix using `filepath.Rel` in read/write tools with tests, (4) new E2E test with MockLLMClient verifying full approval chain, (5) manual test documentation.

**Tech Stack:** Go, tview TUI, MockLLMClient, `testing` stdlib

---

### Task 1: Sandbox Path OS Auto-Detection

**Files:**
- Modify: `internal/data/config/config.go` (in `LoadConfig`, after TOML unmarshal)
- Test: `internal/data/config/config_test.go`

- [ ] **Step 1: Write failing test for OS-specific default sandbox**

Add to `config_test.go`:

```go
func TestDefaultSandboxDir(t *testing.T) {
    cfg := GetDefaultConfig()
    // Default config should have empty sandbox_dir
    for _, role := range cfg.AgentRoles {
        if role.SandboxDir != "" {
            t.Fatalf("expected empty sandbox_dir, got %q", role.SandboxDir)
        }
    }
}

func TestSandboxDirDefaultApplied(t *testing.T) {
    // Create a temp config file with empty sandbox_dir
    content := []byte(`
[[agent_role]]
name = "coder"
enabled = true
model = "test"
system_prompt = "test"
tools = ["read_file", "write_file"]
max_react_loop = 5
sandbox_dir = ""
`)
    path := filepath.Join(t.TempDir(), "config.toml")
    if err := os.WriteFile(path, content, 0644); err != nil {
        t.Fatal(err)
    }
    cfg, err := LoadConfig(path)
    if err != nil {
        t.Fatal(err)
    }
    if len(cfg.AgentRoles) != 1 {
        t.Fatalf("expected 1 role, got %d", len(cfg.AgentRoles))
    }
    role := cfg.AgentRoles[0]
    if role.SandboxDir == "" {
        t.Fatal("expected SandboxDir to be set to default, got empty")
    }
    // On Windows should contain TEMP, on Unix should contain /tmp
    if runtime.GOOS == "windows" {
        if !strings.Contains(role.SandboxDir, os.Getenv("TEMP")) {
            t.Fatalf("expected %q to contain TEMP dir", role.SandboxDir)
        }
    } else {
        if !strings.HasPrefix(role.SandboxDir, "/tmp/") {
            t.Fatalf("expected %q to start with /tmp/", role.SandboxDir)
        }
    }
}

func TestSandboxDirPreservesExplicit(t *testing.T) {
    content := []byte(`
[[agent_role]]
name = "coder"
enabled = true
model = "test"
system_prompt = "test"
tools = ["read_file", "write_file"]
max_react_loop = 5
sandbox_dir = "/custom/path"
`)
    path := filepath.Join(t.TempDir(), "config.toml")
    if err := os.WriteFile(path, content, 0644); err != nil {
        t.Fatal(err)
    }
    cfg, err := LoadConfig(path)
    if err != nil {
        t.Fatal(err)
    }
    if cfg.AgentRoles[0].SandboxDir != "/custom/path" {
        t.Fatalf("expected /custom/path, got %q", cfg.AgentRoles[0].SandboxDir)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -count=1 -run 'TestDefaultSandboxDir|TestSandboxDirDefaultApplied|TestSandboxDirPreservesExplicit' ./internal/data/config/`
Expected: Tests fail because default is not applied

- [ ] **Step 3: Implement OS-aware default in LoadConfig**

In `internal/data/config/config.go`, at the end of `LoadConfig`, after the agent roles are loaded:

```go
func applyDefaultSandbox(cfg *Config) {
    for i := range cfg.AgentRoles {
        if cfg.AgentRoles[i].SandboxDir == "" {
            if runtime.GOOS == "windows" {
                cfg.AgentRoles[i].SandboxDir = filepath.Join(os.Getenv("TEMP"), "agent-tui", "sandbox")
            } else {
                cfg.AgentRoles[i].SandboxDir = "/tmp/agent-tui/sandbox"
            }
        }
    }
}
```

Call `applyDefaultSandbox(&cfg)` in `LoadConfig` before returning. Also add `"runtime"` to imports.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -count=1 -run 'TestDefaultSandboxDir|TestSandboxDirDefaultApplied|TestSandboxDirPreservesExplicit' ./internal/data/config/`
Expected: All pass

- [ ] **Step 5: Commit**

```bash
git add internal/data/config/config.go internal/data/config/config_test.go
git commit -m "feat: OS-aware default sandbox directory path"
```

---

### Task 2: Sandbox Traversal Fix in write_file + Tests

**Files:**
- Modify: `internal/agent/tool/write_file.go` (sandbox check, lines 33-39)
- Modify: `internal/agent/tool/write_file_test.go` (add sandbox enforcement tests)

- [ ] **Step 1: Write failing sandbox enforcement tests for write_file.go**

Add tests to `write_file_test.go`:

```go
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
    // escape by using a sibling directory name that starts with sandbox prefix
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
    // relative path should be resolved within sandbox
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
```

- [ ] **Step 2: Run tests to verify they fail for the right reasons**

Run: `go test -count=1 -run 'TestWriteFileSandbox' ./internal/agent/tool/`
Expected: `TestWriteFileSandboxEscape` fails because sandbox check doesn't detect escape; others may pass or fail depending on current behavior.

- [ ] **Step 3: Fix sandbox check in write_file.go**

Replace lines 33-39:

```go
if ctx.SandboxDir != "" {
    sandbox := filepath.Clean(ctx.SandboxDir)
    target := filepath.Clean(path)
    rel, err := filepath.Rel(sandbox, target)
    if err != nil || strings.HasPrefix(rel, "..") {
        return ToolResult{Error: fmt.Sprintf("path %q is outside sandbox %q", path, ctx.SandboxDir)}
    }
}
```

Remove the existing `strings.HasPrefix` check and the relative path joining logic (it's now handled by `filepath.Rel`). The relative path still needs to be resolved:

```go
if ctx.SandboxDir != "" {
    sandbox := filepath.Clean(ctx.SandboxDir)
    if !filepath.IsAbs(path) {
        path = filepath.Join(sandbox, path)
    }
    target := filepath.Clean(path)
    rel, err := filepath.Rel(sandbox, target)
    if err != nil || strings.HasPrefix(rel, "..") {
        return ToolResult{Error: fmt.Sprintf("path %q is outside sandbox %q", path, ctx.SandboxDir)}
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -count=1 -run 'TestWriteFile' ./internal/agent/tool/`
Expected: All 8+ write_file tests pass (including existing approval and basic tests)

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/write_file.go internal/agent/tool/write_file_test.go
git commit -m "fix: sandbox path traversal detection in write_file tool"
```

---

### Task 3: Sandbox Traversal Fix in read_file + Tests

**Files:**
- Modify: `internal/agent/tool/read_file.go` (sandbox check, lines 31-38)
- Modify: `internal/agent/tool/read_file_test.go` (add sandbox enforcement tests)

- [ ] **Step 1: Write failing sandbox enforcement tests for read_file.go**

Add to `read_file_test.go`:

```go
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
    // create the escape file
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -count=1 -run 'TestReadFileSandbox' ./internal/agent/tool/`
Expected: `TestReadFileSandboxEscape` fails; sandbox escape not detected.

- [ ] **Step 3: Fix sandbox check in read_file.go**

Replace the sandbox check (lines 31-38) with the same pattern as write_file:

```go
if ctx.SandboxDir != "" {
    sandbox := filepath.Clean(ctx.SandboxDir)
    if !filepath.IsAbs(path) {
        path = filepath.Join(sandbox, path)
    }
    target := filepath.Clean(path)
    rel, err := filepath.Rel(sandbox, target)
    if err != nil || strings.HasPrefix(rel, "..") {
        return ToolResult{Error: fmt.Sprintf("path %q is outside sandbox %q", path, ctx.SandboxDir)}
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -count=1 -run 'TestReadFile' ./internal/agent/tool/`
Expected: All 8+ read_file tests pass (4 existing + 4 new)

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/read_file.go internal/agent/tool/read_file_test.go
git commit -m "fix: sandbox path traversal detection in read_file tool"
```

---

### Task 4: E2E Approval Integration Test

**Files:**
- Create: `e2e/approval_test.go` (new file)

- [ ] **Step 1: Write the E2E approval test**

Create `e2e/approval_test.go`:

```go
package e2e

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/example/agent-tui/internal/agent/runtime"
    "github.com/example/agent-tui/internal/agent/tool"
)

func TestApprovalFlow_WriteFile(t *testing.T) {
    sandbox := t.TempDir()

    // Mock LLM: first response asks to write file, second is final
    mockLLM := runtime.NewMockLLMClient([]runtime.LLMResponse{
        {
            Type: "tool_call",
            ToolCalls: []runtime.ToolCall{
                {
                    ID:   "call_1",
                    Name: "write_file",
                    Arguments: map[string]interface{}{
                        "path":    "hello.py",
                        "content": `print("hello")`,
                    },
                },
            },
        },
        {
            Type:    "final",
            Content: "Done",
        },
    })

    toolReg := tool.NewRegistry()
    toolReg.Register(&tool.WriteFileTool{})

    approved := false
    approvalFn := func(toolName string, params map[string]interface{}, oldContent, newContent string) bool {
        approved = true
        if toolName != "write_file" {
            t.Errorf("expected tool_name 'write_file', got %q", toolName)
        }
        path, _ := params["path"].(string)
        if path != "hello.py" {
            t.Errorf("expected path 'hello.py', got %q", path)
        }
        return true
    }

    agent := runtime.NewAgent(runtime.Config{
        Name:         "test-coder",
        Model:        "mock",
        SystemPrompt: "You are a test agent",
        MaxReActLoop: 5,
        SandboxDir:   sandbox,
        ApprovalFn:   approvalFn,
    }, toolReg, mockLLM)

    result := agent.Execute("write a hello world program to hello.py")
    if result.Error != "" {
        t.Fatalf("agent execute failed: %v", result.Error)
    }

    if !approved {
        t.Fatal("expected approval callback to be invoked")
    }

    expectedPath := filepath.Join(sandbox, "hello.py")
    data, err := os.ReadFile(expectedPath)
    if err != nil {
        t.Fatalf("expected file %s to exist: %v", expectedPath, err)
    }
    if string(data) != `print("hello")` {
        t.Fatalf("expected content 'print(\"hello\")', got %q", string(data))
    }
}
```

- [ ] **Step 2: Run test to verify it passes**

Run: `go test -count=1 -run TestApprovalFlow ./e2e/`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add e2e/approval_test.go
git commit -m "test: add E2E approval flow integration test"
```

---

### Task 5: Manual Test Guide

**Files:**
- Create: `docs/testing/approval-manual-test.md`

- [ ] **Step 1: Write the manual test guide**

Create `docs/testing/approval-manual-test.md`:

```markdown
# Manual Test: Agent File Write Approval Modal

## Prerequisites

- Real LLM configured in `configs/config.toml` (e.g., `default_client = "openai"` with API key)
- `sandbox_dir` set to empty string (`sandbox_dir = ""`) to let OS auto-detect
- Project builds successfully: `go build ./cmd/agent`

## Steps

1. Run the TUI:
   ```bash
   go run ./cmd/agent
   ```

2. In the chat panel, send:
   ```
   write a hello world python program to hello.py
   ```

3. Observe that the agent processes the request through the ReAct loop and attempts to call `write_file`.

4. An approval modal overlay should appear showing:
   - File path: `hello.py` (resolved within sandbox)
   - Diff preview: new file with `print("hello")`
   - Prompt: `Approve? (y/n/d)`

5. Press `y` to approve — verify the file is written
6. Press `n` to reject — verify the file is NOT written and agent reports rejection
7. Press `d` to view full diff details

## Verification

After approving, check the sandbox directory for the file:
```bash
# Windows: %TEMP%\agent-tui\sandbox\
# Linux: /tmp/agent-tui/sandbox/
ls hello.py
cat hello.py  # should print "hello"
```

## Expected Results

| Action | Result |
|--------|--------|
| Approve (y) | File written, agent continues |
| Reject (n) | File not written, agent reports error |
| Diff (d) | Full diff shown in modal |
```

- [ ] **Step 2: Commit**

```bash
git add docs/testing/approval-manual-test.md
git commit -m "docs: add manual approval testing guide"
```

---

### Self-Review Checklist

1. **Spec coverage:** Does each spec requirement map to a task?
   - OS-aware sandbox default → Task 1
   - Sandbox traversal fix → Tasks 2, 3
   - Sandbox enforcement tests → Tasks 2, 3 (inline in each)
   - E2E approval test → Task 4
   - Manual test guide → Task 5
   All covered.

2. **Placeholder scan:** No TBD, TODO, or vague steps. All code is explicit.

3. **Type consistency:** `ApprovalFn` signature `func(string, map[string]interface{}, string, string) bool` matches existing type. `runtime.NewMockLLMClient` and `runtime.LLMResponse` match existing patterns. `tool.WriteFileTool`, `ToolContext.SandboxDir` match existing usage.

4. **Test consistency:** Existing tests (`TestWriteFileToolRejected`, `TestWriteFileToolApproved`) are unaffected; new sandbox tests use different function names.
