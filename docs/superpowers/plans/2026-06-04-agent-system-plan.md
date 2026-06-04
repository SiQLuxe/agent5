# Agent System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Tool-based Agent system with ReAct loop, multi-agent orchestration, external agent bridges, and TUI integration.

**Architecture:** Agents execute tasks using a ReAct loop (LLM → tool → observe → repeat). Tools are registered capabilities (read_file, write_file, exec_command, etc.). An Orchestrator manages task decomposition and dispatch across agents. External agents (opencode) appear as delegate tools via bridges.

**Tech Stack:** Go 1.26, stdlib `net/http` + `encoding/json`, DeepSeek/OpenAI function calling format, rivo/tview

---

### Phase 1: Tool System

The foundation: Tool interface, ToolRegistry, and all built-in Tool implementations.

### Task 1: Tool interface + ToolSchema + ToolResult + ToolContext

**Files:**
- Create: `internal/agent/tool/tool.go`
- Test: `internal/agent/tool/tool_test.go`

- [ ] **Step 1: Write the test**

```go
package tool

import (
	"context"
	"testing"
)

func TestToolSchemaValidation(t *testing.T) {
	schema := ToolSchema{
		Parameters: map[string]ParamSchema{
			"name": {Type: "string", Description: "A name"},
			"count": {Type: "integer", Description: "A count"},
		},
		Required: []string{"name"},
	}

	if len(schema.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(schema.Parameters))
	}
	if len(schema.Required) != 1 {
		t.Fatalf("expected 1 required, got %d", len(schema.Required))
	}
}

func TestToolResult(t *testing.T) {
	r := ToolResult{Success: true, Data: "hello"}
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data.(string) != "hello" {
		t.Fatalf("expected 'hello', got %v", r.Data)
	}
}

func TestToolContext(t *testing.T) {
	ctx := context.Background()
	tc := ToolContext{Context: ctx}
	if tc.Context == nil {
		t.Fatal("expected non-nil context")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestTool ./internal/agent/tool/`
Expected: FAIL with package not found

- [ ] **Step 3: Write the implementation**

```go
package tool

import (
	"context"
)

type ToolSchema struct {
	Parameters map[string]ParamSchema
	Required   []string
}

type ParamSchema struct {
	Type        string
	Description string
	Enum        []string
}

type ToolContext struct {
	Context context.Context
}

type ToolResult struct {
	Success bool
	Data    interface{}
	Error   string
}

type Tool interface {
	Name() string
	Description() string
	Schema() ToolSchema
	Execute(ctx ToolContext, params map[string]interface{}) ToolResult
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestTool ./internal/agent/tool/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/tool.go internal/agent/tool/tool_test.go
git commit -m "feat(agent): add Tool interface, ToolSchema, ToolResult, ToolContext"
```

### Task 2: ToolRegistry with Register/Get/List/AsToolDefinitions

**Files:**
- Modify: `internal/agent/tool/tool.go` (add registry)
- Create: `internal/agent/tool/registry.go`
- Test: `internal/agent/tool/registry_test.go`

- [ ] **Step 1: Write the failing test**

```go
package tool

import (
	"context"
	"testing"
)

type mockTool struct{}

func (m *mockTool) Name() string                     { return "mock_tool" }
func (m *mockTool) Description() string              { return "A mock tool" }
func (m *mockTool) Schema() ToolSchema                { return ToolSchema{} }
func (m *mockTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	return ToolResult{Success: true, Data: "ok"}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{})

	got, ok := r.Get("mock_tool")
	if !ok {
		t.Fatal("expected to find mock_tool")
	}
	if got.Name() != "mock_tool" {
		t.Fatalf("expected name mock_tool, got %s", got.Name())
	}
}

func TestRegistryGetMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("expected false for missing tool")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{})

	list := r.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(list))
	}
}

func TestRegistryAsToolDefinitions(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{})

	defs := r.AsToolDefinitions()
	if len(defs) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(defs))
	}

	def := defs[0]
	if def["type"] != "function" {
		t.Fatalf("expected type 'function', got %v", def["type"])
	}

	fn := def["function"].(map[string]interface{})
	if fn["name"] != "mock_tool" {
		t.Fatalf("expected name 'mock_tool', got %v", fn["name"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestRegistry ./internal/agent/tool/`
Expected: FAIL with NewRegistry not defined

- [ ] **Step 3: Write implementation**

```go
package tool

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) List() []Tool {
	list := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

func (r *Registry) AsToolDefinitions() []map[string]interface{} {
	defs := make([]map[string]interface{}, 0, len(r.tools))
	for _, t := range r.tools {
		schema := t.Schema()
		properties := make(map[string]interface{})
		for name, ps := range schema.Parameters {
			prop := map[string]interface{}{
				"type":        ps.Type,
				"description": ps.Description,
			}
			if len(ps.Enum) > 0 {
				prop["enum"] = ps.Enum
			}
			properties[name] = prop
		}
		defs = append(defs, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        t.Name(),
				"description": t.Description(),
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": properties,
					"required":   schema.Required,
				},
			},
		})
	}
	return defs
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestRegistry ./internal/agent/tool/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/registry.go internal/agent/tool/registry_test.go
git commit -m "feat(agent): add ToolRegistry with Register/Get/List/AsToolDefinitions"
```

### Task 3: read_file Tool

**Files:**
- Create: `internal/agent/tool/read_file.go`
- Test: `internal/agent/tool/read_file_test.go`

- [ ] **Step 1: Write the failing test**

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestReadFile ./internal/agent/tool/`
Expected: FAIL with ReadFileTool not defined

- [ ] **Step 3: Write implementation**

```go
package tool

import (
	"fmt"
	"os"
)

type ReadFileTool struct{}

func (t *ReadFileTool) Name() string { return "read_file" }

func (t *ReadFileTool) Description() string { return "Read the content of a file" }

func (t *ReadFileTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"path": {Type: "string", Description: "Path to the file"},
		},
		Required: []string{"path"},
	}
}

func (t *ReadFileTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return ToolResult{Error: "path parameter is required"}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ToolResult{Error: fmt.Sprintf("read file: %s", err)}
	}
	return ToolResult{Success: true, Data: string(data)}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestReadFile ./internal/agent/tool/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/read_file.go internal/agent/tool/read_file_test.go
git commit -m "feat(agent): add read_file tool"
```

### Task 4: write_file Tool

**Files:**
- Create: `internal/agent/tool/write_file.go`
- Test: `internal/agent/tool/write_file_test.go`

- [ ] **Step 1: Write the failing test**

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Expected: FAIL with WriteFileTool not defined

- [ ] **Step 3: Write implementation**

```go
package tool

import (
	"fmt"
	"os"
	"path/filepath"
)

type WriteFileTool struct{}

func (t *WriteFileTool) Name() string { return "write_file" }

func (t *WriteFileTool) Description() string { return "Write or create a file" }

func (t *WriteFileTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"path":    {Type: "string", Description: "Path to the file"},
			"content": {Type: "string", Description: "Content to write"},
		},
		Required: []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	path, _ := params["path"].(string)
	content, _ := params["content"].(string)
	if path == "" {
		return ToolResult{Error: "path parameter is required"}
	}

	dir := filepath.Dir(path)
	if dir != "." {
		os.MkdirAll(dir, 0755)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return ToolResult{Error: fmt.Sprintf("write file: %s", err)}
	}
	return ToolResult{Success: true, Data: fmt.Sprintf("wrote %d bytes to %s", len(content), path)}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestWriteFile ./internal/agent/tool/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/write_file.go internal/agent/tool/write_file_test.go
git commit -m "feat(agent): add write_file tool"
```

### Task 5: search_text Tool

**Files:**
- Create: `internal/agent/tool/search_text.go`
- Test: `internal/agent/tool/search_text_test.go`

- [ ] **Step 1: Write the failing test**

```go
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
	dir := t.TempDir()
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
	dir := t.TempDir()
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
```

- [ ] **Step 2: Write implementation**

```go
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

	cwd, _ := os.Getwd()
	var matches []SearchMatch
	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
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
				rel, _ := filepath.Rel(cwd, path)
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
```

- [ ] **Step 3: Run test to verify it passes**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestSearchText ./internal/agent/tool/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/agent/tool/search_text.go internal/agent/tool/search_text_test.go
git commit -m "feat(agent): add search_text tool"
```

### Task 6: exec_command Tool

**Files:**
- Create: `internal/agent/tool/exec_command.go`
- Test: `internal/agent/tool/exec_command_test.go`

- [ ] **Step 1: Write the failing test**

```go
package tool

import (
	"context"
	"testing"
)

func TestExecCommandToolName(t *testing.T) {
	tool := &ExecCommandTool{}
	if tool.Name() != "exec_command" {
		t.Fatalf("expected name 'exec_command', got %s", tool.Name())
	}
}

func TestExecCommandToolEcho(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping exec test in short mode")
	}
	tool := &ExecCommandTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"command": "echo hello",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	out, ok := result.Data.(CmdOutput)
	if !ok {
		t.Fatalf("expected CmdOutput, got %T", result.Data)
	}
	if out.Stdout != "hello\n" && out.Stdout != "hello" {
		t.Fatalf("expected 'hello', got %q", out.Stdout)
	}
}

func TestExecCommandToolFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping exec test in short mode")
	}
	tool := &ExecCommandTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"command": "exit 42",
	})
	if result.Success {
		t.Fatal("expected failure for exit 42")
	}
}

func TestExecCommandToolMissingParam(t *testing.T) {
	tool := &ExecCommandTool{}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{})
	if result.Success {
		t.Fatal("expected failure for missing param")
	}
}
```

- [ ] **Step 2: Write implementation**

```go
package tool

import (
	"fmt"
	"os/exec"
	"strings"
)

type CmdOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type ExecCommandTool struct{}

func (t *ExecCommandTool) Name() string { return "exec_command" }

func (t *ExecCommandTool) Description() string {
	return "Execute a shell command and return output"
}

func (t *ExecCommandTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"command": {Type: "string", Description: "Shell command to execute"},
			"timeout": {Type: "integer", Description: "Timeout in seconds (default 30)"},
		},
		Required: []string{"command"},
	}
}

func (t *ExecCommandTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	cmdStr, _ := params["command"].(string)
	if cmdStr == "" {
		return ToolResult{Error: "command parameter is required"}
	}

	cmd := exec.CommandContext(ctx.Context, "sh", "-c", cmdStr)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return ToolResult{Error: fmt.Sprintf("exec: %s", err)}
		}
	}

	out := CmdOutput{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}

	if exitCode != 0 {
		return ToolResult{
			Success: false,
			Data:    out,
			Error:   fmt.Sprintf("exit code %d", exitCode),
		}
	}
	return ToolResult{Success: true, Data: out}
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestExecCommand ./internal/agent/tool/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/agent/tool/exec_command.go internal/agent/tool/exec_command_test.go
git commit -m "feat(agent): add exec_command tool"
```

### Task 7: chat_llm Tool

**Files:**
- Create: `internal/agent/tool/chat_llm.go`
- Test: `internal/agent/tool/chat_llm_test.go`

- [ ] **Step 1: Write the failing test**

```go
package tool

import (
	"context"
	"testing"
)

func TestChatLLMToolName(t *testing.T) {
	tool := &ChatLLMTool{Provider: &mockLLMProvider{}}
	if tool.Name() != "chat_llm" {
		t.Fatalf("expected name 'chat_llm', got %s", tool.Name())
	}
}

type mockLLMProvider struct{}

func (m *mockLLMProvider) Chat(model, systemPrompt, userPrompt string) (string, error) {
	return "mock response: " + userPrompt, nil
}

func TestChatLLMToolExecute(t *testing.T) {
	tool := &ChatLLMTool{Provider: &mockLLMProvider{}}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"prompt": "say hi",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Data.(string) != "mock response: say hi" {
		t.Fatalf("unexpected response: %v", result.Data)
	}
}

func TestChatLLMToolMissingPrompt(t *testing.T) {
	tool := &ChatLLMTool{Provider: &mockLLMProvider{}}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{})
	if result.Success {
		t.Fatal("expected failure for missing prompt")
	}
}
```

- [ ] **Step 2: Write implementation**

```go
package tool

type LLMProvider interface {
	Chat(model, systemPrompt, userPrompt string) (string, error)
}

type ChatLLMTool struct {
	Provider LLMProvider
}

func (t *ChatLLMTool) Name() string { return "chat_llm" }

func (t *ChatLLMTool) Description() string { return "Ask an LLM to respond to a prompt" }

func (t *ChatLLMTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"prompt": {Type: "string", Description: "The prompt to send to the LLM"},
			"model":  {Type: "string", Description: "Model name (optional, uses default if empty)"},
		},
		Required: []string{"prompt"},
	}
}

func (t *ChatLLMTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		return ToolResult{Error: "prompt parameter is required"}
	}
	model, _ := params["model"].(string)

	resp, err := t.Provider.Chat(model, "", prompt)
	if err != nil {
		return ToolResult{Error: err.Error()}
	}
	return ToolResult{Success: true, Data: resp}
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestChatLLM ./internal/agent/tool/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/agent/tool/chat_llm.go internal/agent/tool/chat_llm_test.go
git commit -m "feat(agent): add chat_llm tool"
```

---

### Phase 2: Agent Runtime

### Task 8: AgentMemory + AgentLogger

**Files:**
- Create: `internal/agent/runtime/memory.go`
- Create: `internal/agent/runtime/logger.go`
- Test: `internal/agent/runtime/memory_test.go`
- Test: `internal/agent/runtime/logger_test.go`

- [ ] **Step 1: Write failing test for memory**

```go
package runtime

import (
	"testing"
)

func TestAgentMemoryAppend(t *testing.T) {
	m := NewMemory()
	m.Append("user", "hello")
	m.Append("assistant", "hi")

	if len(m.ShortTerm) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(m.ShortTerm))
	}
	if m.ShortTerm[0].Role != "user" || m.ShortTerm[0].Content != "hello" {
		t.Fatalf("unexpected first message: %+v", m.ShortTerm[0])
	}
}

func TestAgentMemorySetGet(t *testing.T) {
	m := NewMemory()
	m.Set("key1", "value1")
	m.Set("key2", "value2")

	if m.Get("key1") != "value1" {
		t.Fatalf("expected 'value1', got %s", m.Get("key1"))
	}
	if m.Get("key3") != "" {
		t.Fatal("expected empty string for missing key")
	}
}

func TestAgentMemoryClear(t *testing.T) {
	m := NewMemory()
	m.Append("user", "hello")
	m.Set("key", "val")
	m.Clear()

	if len(m.ShortTerm) != 0 {
		t.Fatal("expected empty ShortTerm after clear")
	}
}
```

- [ ] **Step 2: Write memory implementation**

```go
package runtime

type Message struct {
	Role    string
	Content string
}

type Memory struct {
	ShortTerm  []Message
	WorkingSet map[string]string
}

func NewMemory() *Memory {
	return &Memory{
		ShortTerm:  make([]Message, 0),
		WorkingSet: make(map[string]string),
	}
}

func (m *Memory) Append(role, content string) {
	m.ShortTerm = append(m.ShortTerm, Message{Role: role, Content: content})
}

func (m *Memory) Set(key, value string) {
	m.WorkingSet[key] = value
}

func (m *Memory) Get(key string) string {
	return m.WorkingSet[key]
}

func (m *Memory) Clear() {
	m.ShortTerm = make([]Message, 0)
	m.WorkingSet = make(map[string]string)
}
```

- [ ] **Step 3: Run memory test**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestAgentMemory ./internal/agent/runtime/ -v`
Expected: PASS

- [ ] **Step 4: Write failing test for logger**

```go
func TestAgentLogger(t *testing.T) {
	log := NewLogger(100)
	log.Log("thought", "I should read the file", "", nil, nil, 0)

	if len(log.Entries()) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(log.Entries()))
	}
	if log.Entries()[0].Phase != "thought" {
		t.Fatalf("expected phase 'thought', got %s", log.Entries()[0].Phase)
	}
}

func TestAgentLoggerMaxEntries(t *testing.T) {
	log := NewLogger(3)
	for i := 0; i < 5; i++ {
		log.Log("thought", "msg", "", nil, nil, 0)
	}
	if len(log.Entries()) != 3 {
		t.Fatalf("expected 3 entries (max), got %d", len(log.Entries()))
	}
}

func TestAgentLoggerClear(t *testing.T) {
	log := NewLogger(100)
	log.Log("thought", "msg", "", nil, nil, 0)
	log.Clear()
	if len(log.Entries()) != 0 {
		t.Fatal("expected 0 entries after clear")
	}
}
```

- [ ] **Step 5: Write logger implementation**

```go
package runtime

import (
	"time"

	"github.com/example/agent-tui/internal/agent/tool"
)

type LogEntry struct {
	Timestamp  time.Time
	Phase      string
	Content    string
	ToolName   string
	ToolParams map[string]interface{}
	ToolResult *tool.ToolResult
	Duration   time.Duration
}

type Logger struct {
	entries    []LogEntry
	maxEntries int
}

func NewLogger(max int) *Logger {
	return &Logger{
		entries:    make([]LogEntry, 0, max),
		maxEntries: max,
	}
}

func (l *Logger) Log(phase, content, toolName string, params map[string]interface{}, result *tool.ToolResult, duration time.Duration) {
	entry := LogEntry{
		Timestamp:  time.Now(),
		Phase:      phase,
		Content:    content,
		ToolName:   toolName,
		ToolParams: params,
		ToolResult: result,
		Duration:   duration,
	}
	if len(l.entries) >= l.maxEntries {
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, entry)
}

func (l *Logger) Entries() []LogEntry {
	return l.entries
}

func (l *Logger) Clear() {
	l.entries = make([]LogEntry, 0, l.maxEntries)
}
```

- [ ] **Step 6: Run logger test**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestAgentLogger ./internal/agent/runtime/ -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/agent/runtime/memory.go internal/agent/runtime/logger.go internal/agent/runtime/memory_test.go internal/agent/runtime/logger_test.go
git commit -m "feat(agent): add AgentMemory and AgentLogger"
```

### Task 9: Agent struct + ReAct Loop

**Files:**
- Create: `internal/agent/runtime/agent.go`
- Create: `internal/agent/runtime/react.go`
- Create: `internal/agent/runtime/llm.go` (LLM call abstraction)
- Test: `internal/agent/runtime/agent_test.go`

- [ ] **Step 1: Write LLM call abstraction**

```go
package runtime

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ToolCall struct {
	Name      string
	Arguments map[string]interface{}
}

type LLMResponse struct {
	Type     string   // "tool_call" or "final"
	Content  string
	ToolCall *ToolCall
}

type LLMClient interface {
	ChatWithTools(messages []Message, tools []map[string]interface{}, model string) (*LLMResponse, error)
}

type MockLLMClient struct {
	Responses []LLMResponse
	index     int
}

func NewMockLLMClient(responses []LLMResponse) *MockLLMClient {
	return &MockLLMClient{Responses: responses}
}

func (m *MockLLMClient) ChatWithTools(messages []Message, tools []map[string]interface{}, model string) (*LLMResponse, error) {
	if m.index >= len(m.Responses) {
		return &LLMResponse{Type: "final", Content: "done"}, nil
	}
	resp := m.Responses[m.index]
	m.index++
	return &resp, nil
}
```

Add this to `internal/agent/runtime/llm.go`

- [ ] **Step 2: Write the Agent struct + config**

```go
package runtime

import (
	"fmt"

	"github.com/example/agent-tui/internal/agent/tool"
)

type Config struct {
	Name        string
	Model       string
	SystemPrompt string
	MaxReActLoop int
	Temperature  float64
	ContextLimit int
}

type Agent struct {
	Name   string
	Config Config
	Tools  *tool.Registry
	Memory *Memory
	Logger *Logger
	llm    LLMClient
}

func NewAgent(cfg Config, tools *tool.Registry, llm LLMClient) *Agent {
	if cfg.MaxReActLoop == 0 {
		cfg.MaxReActLoop = 20
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = 0.7
	}
	if cfg.ContextLimit == 0 {
		cfg.ContextLimit = 50
	}
	return &Agent{
		Name:   cfg.Name,
		Config: cfg,
		Tools:  tools,
		Memory: NewMemory(),
		Logger: NewLogger(100),
		llm:    llm,
	}
}

func (a *Agent) Execute(task string) (string, error) {
	a.Memory.Clear()
	a.Logger.Clear()
	return a.reactLoop(task)
}
```

Add to `internal/agent/runtime/agent.go`

- [ ] **Step 3: Write the failing test**

```go
package runtime

import (
	"testing"

	"github.com/example/agent-tui/internal/agent/tool"
)

type mockReadTool struct{}

func (m *mockReadTool) Name() string { return "read_file" }
func (m *mockReadTool) Description() string { return "read a file" }
func (m *mockReadTool) Schema() tool.ToolSchema {
	return tool.ToolSchema{
		Parameters: map[string]tool.ParamSchema{
			"path": {Type: "string", Description: "file path"},
		},
		Required: []string{"path"},
	}
}
func (m *mockReadTool) Execute(ctx tool.ToolContext, params map[string]interface{}) tool.ToolResult {
	return tool.ToolResult{Success: true, Data: "file content"}
}

func TestAgentSimpleFinal(t *testing.T) {
	reg := tool.NewRegistry()
	reg.Register(&mockReadTool{})

	mock := NewMockLLMClient([]LLMResponse{
		{Type: "final", Content: "Task complete"},
	})

	agent := NewAgent(Config{
		Name:         "test",
		Model:        "test-model",
		SystemPrompt: "You are a test agent",
	}, reg, mock)

	result, err := agent.Execute("do something")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result != "Task complete" {
		t.Fatalf("expected 'Task complete', got %s", result)
	}
}

func TestAgentToolCallThenFinal(t *testing.T) {
	reg := tool.NewRegistry()
	reg.Register(&mockReadTool{})

	mock := NewMockLLMClient([]LLMResponse{
		{
			Type: "tool_call",
			ToolCall: &ToolCall{
				Name:      "read_file",
				Arguments: map[string]interface{}{"path": "/tmp/test.txt"},
			},
		},
		{Type: "final", Content: "Done reading"},
	})

	agent := NewAgent(Config{
		Name:         "test",
		Model:        "test-model",
		SystemPrompt: "You are a test agent",
	}, reg, mock)

	result, err := agent.Execute("read a file")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result != "Done reading" {
		t.Fatalf("expected 'Done reading', got %s", result)
	}
	if len(agent.Logger.Entries()) < 2 {
		t.Fatal("expected at least 2 log entries")
	}
}
```

- [ ] **Step 4: Write the ReAct loop**

```go
package runtime

import (
	"fmt"
	"time"

	"github.com/example/agent-tui/internal/agent/tool"
)

func (a *Agent) reactLoop(task string) (string, error) {
	messages := []Message{
		{Role: "system", Content: a.Config.SystemPrompt},
		{Role: "user", Content: task},
	}

	for i := 0; i < a.Config.MaxReActLoop; i++ {
		start := time.Now()
		resp, err := a.llm.ChatWithTools(messages, a.Tools.AsToolDefinitions(), a.Config.Model)
		if err != nil {
			return "", fmt.Errorf("llm call: %w", err)
		}
		a.Logger.Log(string(resp.Type), resp.Content, "", nil, nil, time.Since(start))

		switch resp.Type {
		case "tool_call":
			if resp.ToolCall == nil {
				continue
			}
			t, ok := a.Tools.Get(resp.ToolCall.Name)
			if !ok {
				errMsg := fmt.Sprintf("tool %q not found", resp.ToolCall.Name)
				messages = append(messages, Message{Role: "tool", Content: errMsg})
				a.Logger.Log("tool_error", errMsg, resp.ToolCall.Name, resp.ToolCall.Arguments, nil, 0)
				continue
			}

			toolStart := time.Now()
			tc := tool.ToolContext{Context: nil}
			result := t.Execute(tc, resp.ToolCall.Arguments)
			a.Logger.Log("tool_result", fmt.Sprintf("%+v", result.Data), resp.ToolCall.Name, resp.ToolCall.Arguments, &result, time.Since(toolStart))

			if result.Success {
				content := fmt.Sprintf("%v", result.Data)
				messages = append(messages, Message{Role: "tool", Content: content})
			} else {
				messages = append(messages, Message{Role: "tool", Content: fmt.Sprintf("error: %s", result.Error)})
			}

		case "final":
			a.Memory.Append("assistant", resp.Content)
			return resp.Content, nil
		}
	}

	return "", fmt.Errorf("max react loop iterations (%d) reached", a.Config.MaxReActLoop)
}
```

Add to `internal/agent/runtime/react.go`

- [ ] **Step 5: Run test**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestAgent ./internal/agent/runtime/ -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/agent/runtime/agent.go internal/agent/runtime/react.go internal/agent/runtime/llm.go internal/agent/runtime/agent_test.go
git commit -m "feat(agent): add Agent struct and ReAct loop"
```

---

### Phase 3: Orchestrator

### Task 10: Task types + AgentRegistry

**Files:**
- Create: `internal/agent/orchestrator/types.go`
- Create: `internal/agent/orchestrator/registry.go`
- Test: `internal/agent/orchestrator/registry_test.go`

- [ ] **Step 1: Write types**

```go
package orchestrator

type TaskType string

const (
	TaskAnalyze  TaskType = "task_analyze"
	TaskDesign   TaskType = "task_design"
	TaskCode     TaskType = "task_code"
	TaskReview   TaskType = "task_review"
	TaskExecute  TaskType = "task_execute"
	TaskResearch TaskType = "task_research"
	TaskCustom   TaskType = "task_custom"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID      string
	Type    TaskType
	Content string
	Status  TaskStatus
	AgentID string
	Result  string
	Error   string
}
```

- [ ] **Step 2: Write registry**

```go
package orchestrator

import "github.com/example/agent-tui/internal/agent/runtime"

type Registry struct {
	agents map[string]*runtime.Agent
	roles  map[string][]string
}

func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]*runtime.Agent),
		roles:  make(map[string][]string),
	}
}

func (r *Registry) Register(name string, agent *runtime.Agent, roleTags ...string) {
	r.agents[name] = agent
	for _, tag := range roleTags {
		r.roles[tag] = append(r.roles[tag], name)
	}
}

func (r *Registry) Get(name string) (*runtime.Agent, bool) {
	a, ok := r.agents[name]
	return a, ok
}

func (r *Registry) List() []*runtime.Agent {
	list := make([]*runtime.Agent, 0, len(r.agents))
	for _, a := range r.agents {
		list = append(list, a)
	}
	return list
}

func (r *Registry) FindByRole(role string) []*runtime.Agent {
	names := r.roles[role]
	agents := make([]*runtime.Agent, 0, len(names))
	for _, n := range names {
		if a, ok := r.agents[n]; ok {
			agents = append(agents, a)
		}
	}
	return agents
}
```

- [ ] **Step 3: Write and run test**

```go
package orchestrator

import (
	"testing"

	"github.com/example/agent-tui/internal/agent/runtime"
	"github.com/example/agent-tui/internal/agent/tool"
)

func TestRegistryRegisterAndGet(t *testing.T) {
	reg := NewRegistry()
	agent := runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), nil)
	reg.Register("coder", agent, "code", "review")

	got, ok := reg.Get("coder")
	if !ok {
		t.Fatal("expected to find coder")
	}
	if got.Name != "coder" {
		t.Fatalf("expected name coder, got %s", got.Name)
	}
}

func TestRegistryFindByRole(t *testing.T) {
	reg := NewRegistry()
	reg.Register("planner", runtime.NewAgent(runtime.Config{Name: "planner"}, tool.NewRegistry(), nil), "analyze", "design")
	reg.Register("coder", runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), nil), "code")

	agents := reg.FindByRole("code")
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent for role 'code', got %d", len(agents))
	}
	agents = reg.FindByRole("analyze")
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent for role 'analyze', got %d", len(agents))
	}
	agents = reg.FindByRole("nonexistent")
	if len(agents) != 0 {
		t.Fatalf("expected 0 agents for missing role, got %d", len(agents))
	}
}
```

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestRegistry ./internal/agent/orchestrator/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/agent/orchestrator/types.go internal/agent/orchestrator/registry.go internal/agent/orchestrator/registry_test.go
git commit -m "feat(agent): add Task types and AgentRegistry"
```

### Task 11: TaskDecomposer + ResultMerger

**Files:**
- Create: `internal/agent/orchestrator/decomposer.go`
- Create: `internal/agent/orchestrator/merger.go`
- Test: `internal/agent/orchestrator/decomposer_test.go`
- Test: `internal/agent/orchestrator/merger_test.go`

- [ ] **Step 1: Write decomposer**

```go
package orchestrator

type DecomposeStrategy func(task *Task) ([]*Task, error)

type Decomposer struct {
	strategies map[string]DecomposeStrategy
}

func NewDecomposer() *Decomposer {
	return &Decomposer{
		strategies: make(map[string]DecomposeStrategy),
	}
}

func (d *Decomposer) Register(taskType TaskType, strategy DecomposeStrategy) {
	d.strategies[string(taskType)] = strategy
}

func (d *Decomposer) Decompose(task *Task) ([]*Task, error) {
	strategy, ok := d.strategies[string(task.Type)]
	if !ok {
		return []*Task{task}, nil
	}
	return strategy(task)
}

func DefaultSequentialStrategy(task *Task) ([]*Task, error) {
	return []*Task{
		{ID: task.ID + "-analyze", Type: TaskAnalyze, Content: "Analyze: " + task.Content, Status: StatusPending},
		{ID: task.ID + "-design", Type: TaskDesign, Content: "Design: based on analysis", Status: StatusPending},
		{ID: task.ID + "-code", Type: TaskCode, Content: "Implement: based on design", Status: StatusPending},
		{ID: task.ID + "-review", Type: TaskReview, Content: "Review: verify implementation", Status: StatusPending},
	}, nil
}
```

- [ ] **Step 2: Write merger**

```go
package orchestrator

import "strings"

type MergeStrategy func(results []*Task) string

type Merger struct {
	strategy MergeStrategy
}

func NewMerger() *Merger {
	return &Merger{strategy: ConcatMerge}
}

func (m *Merger) SetStrategy(s MergeStrategy) {
	m.strategy = s
}

func (m *Merger) Merge(results []*Task) string {
	return m.strategy(results)
}

func ConcatMerge(results []*Task) string {
	var b strings.Builder
	for _, r := range results {
		if r.Status == StatusCompleted {
			b.WriteString("## " + string(r.Type) + "\n")
			b.WriteString(r.Result + "\n\n")
		} else if r.Status == StatusFailed {
			b.WriteString("## " + string(r.Type) + " (FAILED)\n")
			b.WriteString(r.Error + "\n\n")
		}
	}
	return b.String()
}
```

- [ ] **Step 3: Write and run tests**

```go
func TestDecomposerPassthrough(t *testing.T) {
	d := NewDecomposer()
	task := &Task{ID: "t1", Type: TaskCustom, Content: "do something"}
	steps, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected 1 step (passthrough), got %d", len(steps))
	}
}

func TestDecomposerSequential(t *testing.T) {
	d := NewDecomposer()
	d.Register(TaskAnalyze, DefaultSequentialStrategy)
	task := &Task{ID: "t1", Type: TaskAnalyze, Content: "add login"}
	steps, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}
	if len(steps) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(steps))
	}
	if steps[0].Type != TaskAnalyze {
		t.Fatalf("first step should be analyze")
	}
	if steps[3].Type != TaskReview {
		t.Fatalf("last step should be review")
	}
}

func TestConcatMerge(t *testing.T) {
	results := []*Task{
		{ID: "1", Type: TaskAnalyze, Status: StatusCompleted, Result: "Analysis done"},
		{ID: "2", Type: TaskCode, Status: StatusCompleted, Result: "Code done"},
		{ID: "3", Type: TaskReview, Status: StatusFailed, Error: "bugs found"},
	}
	output := ConcatMerge(results)
	if !strings.Contains(output, "Analysis done") {
		t.Fatal("expected Analysis done in output")
	}
	if !strings.Contains(output, "FAILED") {
		t.Fatal("expected FAILED marker")
	}
}
```

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestDecomposer\|TestConcatMerge ./internal/agent/orchestrator/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/agent/orchestrator/decomposer.go internal/agent/orchestrator/merger.go internal/agent/orchestrator/decomposer_test.go internal/agent/orchestrator/merger_test.go
git commit -m "feat(agent): add TaskDecomposer and ResultMerger"
```

### Task 12: Orchestrator (full dispatch)

**Files:**
- Create: `internal/agent/orchestrator/orchestrator.go`
- Test: `internal/agent/orchestrator/orchestrator_test.go`

- [ ] **Step 1: Write the failing test**

```go
package orchestrator

import (
	"testing"

	"github.com/example/agent-tui/internal/agent/runtime"
	"github.com/example/agent-tui/internal/agent/tool"
)

type mockAgent struct {
	name string
	resp string
	err  error
}

func (m *mockAgent) Execute(task string) (string, error) {
	return m.resp, m.err
}

func newMockAgent(name, resp string, err error) *runtime.Agent {
	return &runtime.Agent{Name: name}
}

func TestOrchestratorDispatchSingle(t *testing.T) {
	reg := NewRegistry()
	agent := runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), nil)
	reg.Register("coder", agent, "code")

	orch := NewOrchestrator(reg, NewDecomposer(), NewMerger())

	task := &Task{ID: "t1", Type: TaskCode, Content: "write code"}
	results, err := orch.Dispatch(task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestOrchestratorDispatchSplit(t *testing.T) {
	reg := NewRegistry()

	planner := runtime.NewAgent(runtime.Config{Name: "planner"}, tool.NewRegistry(), nil)
	coder := runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), nil)
	reviewer := runtime.NewAgent(runtime.Config{Name: "reviewer"}, tool.NewRegistry(), nil)

	reg.Register("planner", planner, "analyze", "design")
	reg.Register("coder", coder, "code")
	reg.Register("reviewer", reviewer, "review")

	d := NewDecomposer()
	d.Register(TaskDesign, DefaultSequentialStrategy)

	orch := NewOrchestrator(reg, d, NewMerger())

	task := &Task{ID: "t1", Type: TaskDesign, Content: "add login feature"}
	results, err := orch.Dispatch(task)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if len(results) != 4 {
		t.Fatalf("expected 4 results (decomposed), got %d", len(results))
	}
}

func TestOrchestratorNoAgent(t *testing.T) {
	reg := NewRegistry()
	orch := NewOrchestrator(reg, NewDecomposer(), NewMerger())

	task := &Task{ID: "t1", Type: TaskCode, Content: "write code"}
	_, err := orch.Dispatch(task)
	if err == nil {
		t.Fatal("expected error when no agent for task type")
	}
}
```

- [ ] **Step 2: Write implementation**

```go
package orchestrator

import (
	"fmt"

	"github.com/example/agent-tui/internal/agent/runtime"
)

type Orchestrator struct {
	registry   *Registry
	decomposer *Decomposer
	merger     *Merger
}

func NewOrchestrator(reg *Registry, d *Decomposer, m *Merger) *Orchestrator {
	return &Orchestrator{
		registry:   reg,
		decomposer: d,
		merger:     m,
	}
}

func (o *Orchestrator) Dispatch(task *Task) ([]*Task, error) {
	steps, err := o.decomposer.Decompose(task)
	if err != nil {
		return nil, fmt.Errorf("decompose: %w", err)
	}

	var results []*Task
	for _, step := range steps {
		agents := o.registry.FindByRole(string(step.Type))
		if len(agents) == 0 {
			step.Status = StatusFailed
			step.Error = fmt.Sprintf("no agent found for role %q", step.Type)
			results = append(results, step)
			return results, fmt.Errorf(step.Error)
		}

		agent := agents[0]
		step.Status = StatusRunning
		step.AgentID = agent.Name

		result, err := agent.Execute(step.Content)
		if err != nil {
			step.Status = StatusFailed
			step.Error = err.Error()
			results = append(results, step)
			return results, err
		}

		step.Status = StatusCompleted
		step.Result = result
		results = append(results, step)
	}

	return results, nil
}

func (o *Orchestrator) DispatchAndMerge(task *Task) (string, error) {
	results, err := o.Dispatch(task)
	if err != nil {
		return "", err
	}
	return o.merger.Merge(results), nil
}
```

- [ ] **Step 3: Run test**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestOrchestrator ./internal/agent/orchestrator/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/agent/orchestrator/orchestrator.go internal/agent/orchestrator/orchestrator_test.go
git commit -m "feat(agent): add Orchestrator with dispatch and merge"
```

---

### Phase 4: External Agent Bridge

### Task 13: OpencodeBridge

**Files:**
- Create: `internal/agent/bridge/opencode.go`
- Test: `internal/agent/bridge/opencode_test.go`

- [ ] **Step 1: Write the failing test**

```go
package bridge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/agent-tui/internal/agent/tool"
)

func TestOpencodeBridgeName(t *testing.T) {
	b := &OpencodeBridge{baseURL: "http://localhost:4096"}
	if b.Name() != "delegate_opencode" {
		t.Fatalf("expected name 'delegate_opencode', got %s", b.Name())
	}
}

func TestOpencodeBridgeSendTask(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"id": "sess_123", "title": "test"})
	})
	mux.HandleFunc("/session/sess_123/message", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"info": map[string]string{"id": "msg_1"},
			"parts": []map[string]string{
				{"type": "text", "text": "task result"},
			},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	b := NewOpencodeBridge(srv.URL, "")
	ctx := tool.ToolContext{Context: context.Background()}
	result := b.Execute(ctx, map[string]interface{}{
		"task": "implement login feature",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Data.(string) != "task result" {
		t.Fatalf("expected 'task result', got %v", result.Data)
	}
}

func TestOpencodeBridgeMissingTask(t *testing.T) {
	b := &OpencodeBridge{baseURL: "http://localhost:4096"}
	ctx := tool.ToolContext{Context: context.Background()}
	result := b.Execute(ctx, map[string]interface{}{})
	if result.Success {
		t.Fatal("expected failure for missing task param")
	}
}
```

- [ ] **Step 2: Write implementation**

```go
package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/example/agent-tui/internal/agent/tool"
)

type OpencodeBridge struct {
	baseURL    string
	apiKey     string
	client     *http.Client
	sessionID  string
}

func NewOpencodeBridge(baseURL, apiKey string) *OpencodeBridge {
	return &OpencodeBridge{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (b *OpencodeBridge) Name() string { return "delegate_opencode" }

func (b *OpencodeBridge) Description() string {
	return "Delegate a task to opencode (external AI coding agent)"
}

func (b *OpencodeBridge) Schema() tool.ToolSchema {
	return tool.ToolSchema{
		Parameters: map[string]tool.ParamSchema{
			"task": {Type: "string", Description: "Task description for opencode"},
		},
		Required: []string{"task"},
	}
}

func (b *OpencodeBridge) Execute(ctx tool.ToolContext, params map[string]interface{}) tool.ToolResult {
	task, _ := params["task"].(string)
	if task == "" {
		return tool.ToolResult{Error: "task parameter is required"}
	}

	// Create session if needed
	if b.sessionID == "" {
		sessionID, err := b.createSession(ctx.Context)
		if err != nil {
			return tool.ToolResult{Error: fmt.Sprintf("create session: %s", err)}
		}
		b.sessionID = sessionID
	}

	// Send task as message
	result, err := b.sendMessage(ctx.Context, task)
	if err != nil {
		return tool.ToolResult{Error: fmt.Sprintf("send message: %s", err)}
	}

	return tool.ToolResult{Success: true, Data: result}
}

func (b *OpencodeBridge) createSession(ctx context.Context) (string, error) {
	body := map[string]string{"title": "agent-delegated-task"}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/session", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.ID, nil
}

func (b *OpencodeBridge) sendMessage(ctx context.Context, task string) (string, error) {
	body := map[string]interface{}{
		"parts": []map[string]string{
			{"type": "text", "text": task},
		},
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/session/"+b.sessionID+"/message", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return "task delegated to opencode", nil
}
```

- [ ] **Step 3: Run test**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test -run TestOpencodeBridge ./internal/agent/bridge/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/agent/bridge/opencode.go internal/agent/bridge/opencode_test.go
git commit -m "feat(agent): add OpencodeBridge (external agent delegate tool)"
```

---

### Phase 5: Wire Everything Together

### Task 14: Config changes + main.go wiring

**Files:**
- Modify: `internal/data/config/config.go`
- Modify: `cmd/agent/main.go`

- [ ] **Step 1: Read config.go and main.go**

```bash
cd /Users/luxe/Downloads/agentx/agent5 && cat internal/data/config/config.go cmd/agent/main.go
```

- [ ] **Step 2: Add AgentConfig to config.go**

```go
type AgentSection struct {
	Agents []AgentConfig `toml:"agent"`
}

type AgentConfig struct {
	Name         string   `toml:"name"`
	Enabled      bool     `toml:"enabled"`
	Model        string   `toml:"model"`
	SystemPrompt string   `toml:"system_prompt"`
	Tools        []string `toml:"tools"`
	MaxReActLoop int      `toml:"max_react_loop"`
}
```

- [ ] **Step 3: Wire in main.go**

After AI client initialization and before UI setup, add:

```go
import (
    "github.com/example/agent-tui/internal/agent/tool"
    "github.com/example/agent-tui/internal/agent/runtime"
    "github.com/example/agent-tui/internal/agent/orchestrator"
)

// Initialize tool registry with all built-in tools
toolReg := tool.NewRegistry()
toolReg.Register(&tool.ReadFileTool{})
toolReg.Register(&tool.WriteFileTool{})
toolReg.Register(&tool.SearchTextTool{})
toolReg.Register(&tool.ExecCommandTool{})

// If LLM provider configured, add chat_llm tool
if aiClient != nil {
    toolReg.Register(&tool.ChatLLMTool{Provider: &aiLLMProvider{client: aiClient}})
}

// Create LLM client for agents
agentLLM := runtime.NewLLMAdapter(aiClient, cfg.DefaultModel)

// Create agents from config
agentReg := orchestrator.NewRegistry()
for _, ac := range cfg.AgentSection.Agents {
    if !ac.Enabled {
        continue
    }
    // Filter tools by config
    agentTools := tool.NewRegistry()
    for _, name := range ac.Tools {
        if t, ok := toolReg.Get(name); ok {
            agentTools.Register(t)
        }
    }
    agent := runtime.NewAgent(runtime.Config{
        Name:         ac.Name,
        Model:        ac.Model,
        SystemPrompt: ac.SystemPrompt,
        MaxReActLoop: ac.MaxReActLoop,
    }, agentTools, agentLLM)
    agentReg.Register(ac.Name, agent, string(orchestrator.TaskAnalyze), string(orchestrator.TaskCode), string(orchestrator.TaskReview))
}

// Create orchestrator
orch := orchestrator.NewOrchestrator(agentReg, orchestrator.NewDecomposer(), orchestrator.NewMerger())
```

- [ ] **Step 4: Commit**

```bash
git add internal/data/config/config.go cmd/agent/main.go
git commit -m "feat: wire agent system into main.go and config"
```

### Task 15: Clean up old dead code

**Files:**
- Remove: `internal/service/agent_roles.go` (or replace with bridge to new system)
- Keep: `internal/service/collaboration_manager.go` (replace with delegator to new Orchestrator)
- Keep: `internal/service/task_orchestrator.go` (replace with delegator to new Orchestrator)

- [ ] **Step 1: Replace task_orchestrator.go with delegation to new system**

```go
package service

import (
    "github.com/example/agent-tui/internal/agent/orchestrator"
)

type TaskOrchestrator struct {
    orch *orchestrator.Orchestrator
}

func NewTaskOrchestrator(orch *orchestrator.Orchestrator) *TaskOrchestrator {
    return &TaskOrchestrator{orch: orch}
}
```

- [ ] **Step 2: Remove old agent_roles.go**

Remove PlanningAgent, CodingAgent, ReviewAgent, ExecutionAgent — replaced by generic Agent with different system prompts. Keep ExternalAgent as a bridge to the new system.

- [ ] **Step 3: Commit**

```bash
git rm internal/service/agent_roles.go
git add internal/service/task_orchestrator.go
git commit -m "refactor: remove old AgentRole implementations, delegate to new agent system"
```

---

### Phase 6: Tests

### Task 16: Run all tests and fix

- [ ] **Step 1: Run all existing tests**

```bash
cd /Users/luxe/Downloads/agentx/agent5 && go test -count=1 ./...
```

- [ ] **Step 2: Run race detector**

```bash
cd /Users/luxe/Downloads/agentx/agent5 && go test -race -count=1 ./...
```

- [ ] **Step 3: Run build**

```bash
cd /Users/luxe/Downloads/agentx/agent5 && go build ./...
```

Fix any issues found.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore: fix tests after agent system migration"
```
