# Agent Approval Testing Design

## Motivation

The agent file write approval system is fully implemented but lacks:
1. Cross-platform sandbox path support (currently hardcoded to Unix `/tmp/`)
2. A sandbox path traversal security fix
3. An end-to-end integration test verifying the full approval chain
4. Documentation for manual TUI testing with a real LLM

This spec addresses all four gaps in a single focused change.

## Design

### 1. Sandbox Path OS Auto-Detection

**File:** `internal/data/config/config.go`

In `LoadConfig()`, after loading TOML data, apply default `SandboxDir` per agent role:

```
if role.SandboxDir == "":
    if runtime.GOOS == "windows":
        role.SandboxDir = filepath.Join(os.Getenv("TEMP"), "agent-tui", "sandbox")
    else:
        role.SandboxDir = "/tmp/agent-tui/sandbox"
```

If `sandbox_dir` is explicitly set in config.toml, it is preserved as-is (user override).

### 2. Sandbox Path Traversal Fix

**Files:** `internal/agent/tool/write_file.go`, `internal/agent/tool/read_file.go`

Replace current `strings.HasPrefix` sandbox check with `filepath.Rel`:

```go
sandbox := filepath.Clean(ctx.SandboxDir)
target := filepath.Clean(path)
rel, err := filepath.Rel(sandbox, target)
if err != nil || strings.HasPrefix(rel, "..") {
    return ToolResult{Error: fmt.Sprintf("path %q is outside sandbox %q", path, ctx.SandboxDir)}
}
```

This correctly rejects `/tmp/sandbox-escape/evil.txt` when sandbox is `/tmp/sandbox`.

**Tests:** Add sandbox enforcement tests in `write_file_test.go` and `read_file_test.go`:
- Write/read within sandbox (relative and absolute paths)
- Write/read outside sandbox (rejected)
- Sandbox escape attempt (e.g., `/tmp/sandbox-escape/file.txt`)

### 3. E2E Integration Test

**File:** `e2e/approval_test.go` (new)

One test function:

**`TestApprovalFlow_WriteFile`**:
- `MockLLMClient` preset: `[tool_call(write_file, {path: "hello.py", content: 'print("hello")'}), final("done")]`
- `WriteFileTool` registered in `tool.NewRegistry()`
- `runtime.NewAgent()` with `ApprovalFn: func(...) bool { return true }`
- `t.TempDir()` as sandbox
- `agent.Execute()` with prompt "write hello.py"
- Assert file `hello.py` exists with content `print("hello")`
- Assert logger contains write_file entry

### 4. TUI Manual Test Guide

**File:** Manual test instructions in `docs/testing/approval-manual-test.md` (or inline in the design doc if placed elsewhere).

Steps:
1. Set `sandbox_dir = ""` in config.toml (or let OS auto-detect take effect)
2. Configure `default_client = "openai"` with valid API key
3. Run `go run ./cmd/agent`
4. In chat, send: "write a hello world python program to hello.py"
5. Observe approval modal with diff preview
6. Press `y` to approve

### Files Changed

| File | Change |
|------|--------|
| `internal/data/config/config.go` | OS-aware default SandboxDir |
| `internal/agent/tool/write_file.go` | Fix sandbox traversal check |
| `internal/agent/tool/read_file.go` | Fix sandbox traversal check |
| `internal/agent/tool/write_file_test.go` | Add sandbox enforcement tests |
| `internal/agent/tool/read_file_test.go` | Add sandbox enforcement tests |
| `e2e/approval_test.go` | New file: approval flow integration test |
| `docs/testing/approval-manual-test.md` | New file: manual test guide |

### Out of Scope

- `exec_command.go` shell hardcoding (`"sh"` for Unix)
- UI/ApprovalModal changes
- Agent/ReAct loop changes
- Config format changes
- Non-approval related e2e test improvements
