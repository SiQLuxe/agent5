# Subagent Phase 3: Full TaskTool Execution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make TaskTool actually execute subagents (not just placeholder), add background/foreground mode, tool filtering by subagent_type, child session isolation, depth limiting, configurable concurrency via config.toml, and heartbeat/liveness detection.

**Architecture:** Add `SubagentRunner` interface in `tool` package to avoid circular deps (`runtime` imports `tool`), implement in `runtime` package. TaskTool creates child sessions, filters tools per subagent_type, runs agents in goroutines. Background mode injects result to parent session. SubagentManager enforces depth limit and heartbeat.

**Tech Stack:** Go 1.26, goroutines/channels, existing `session.Manager`, existing `tool.Registry`, existing `runtime.Agent`.

---

### Task 1: History + Session child session support

**Files:**
- Modify: `internal/data/history/history.go`
- Modify: `internal/agent/session/manager.go`

- [ ] **Step 1: Add ParentID to SessionInfo and CreateChildSession to History**

Edit `internal/data/history/history.go`:

```go
type SessionInfo struct {
	ID        string
	Name      string
	ParentID  string
	CreatedAt time.Time
}

func (h *History) CreateChildSession(name, parentID string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := uuid.New().String()
	h.sessions[id] = SessionInfo{
		ID:        id,
		Name:      name,
		ParentID:  parentID,
		CreatedAt: time.Now(),
	}
	h.messages[id] = []Message{}
	return id
}
```

- [ ] **Step 2: Add CreateChildSession to session.Manager**

Edit `internal/agent/session/manager.go`. After `CreateSession`:

```go
func (m *Manager) CreateChildSession(parentID, name string) string {
	childID := m.history.CreateChildSession(name, parentID)
	return childID
}
```

- [ ] **Step 3: Run history tests**

Run: `cd G:\mllm\agent5 && go test ./internal/data/history/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/data/history/history.go internal/agent/session/manager.go
git commit -m "feat: add child session support for subagents"
```

---

### Task 2: SubagentManager — Complete/Fail, depth limit, heartbeat

**Files:**
- Modify: `internal/agent/tool/subagent_manager.go`
- Modify: `internal/agent/tool/subagent_manager_test.go`

- [ ] **Step 1: Update tests for new methods**

Edit `internal/agent/tool/subagent_manager_test.go`. Add after `TestSubagentManagerGetReturnsCopy`:

```go
func TestSubagentManagerComplete(t *testing.T) {
	mgr := NewSubagentManager(5)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sa := mgr.Spawn(ctx, SubAgentConfig{Name: "test"})
	mgr.Complete(sa.ID, "done")

	got := mgr.Get(sa.ID)
	if got.Status != StatusCompleted {
		t.Errorf("expected '%s', got '%s'", StatusCompleted, got.Status)
	}
	if got.Result != "done" {
		t.Errorf("expected result 'done', got %s", got.Result)
	}
}

func TestSubagentManagerFail(t *testing.T) {
	mgr := NewSubagentManager(5)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sa := mgr.Spawn(ctx, SubAgentConfig{Name: "test"})
	mgr.Fail(sa.ID, "something went wrong")

	got := mgr.Get(sa.ID)
	if got.Status != StatusFailed {
		t.Errorf("expected '%s', got '%s'", StatusFailed, got.Status)
	}
	if got.Error != "something went wrong" {
		t.Errorf("expected error 'something went wrong', got %s", got.Error)
	}
}

func TestSubagentManagerDepthLimit(t *testing.T) {
	mgr := NewSubagentManager(5)
	ctx := context.Background()

	sa := mgr.Spawn(ctx, SubAgentConfig{Name: "deep", MaxDepth: 0})
	if sa.Status != StatusFailed {
		t.Errorf("expected failed (depth limit), got %s", sa.Status)
	}
	if sa.Error == "" {
		t.Fatal("expected error message about depth limit")
	}
}
```

- [ ] **Step 2: Run new tests to see them fail**

Run: `cd G:\mllm\agent5 && go test ./internal/agent/tool/ -v -run "TestSubagentManager(Complete|Fail|DepthLimit)"`
Expected: Tests fail (methods not defined)

- [ ] **Step 3: Add StatusCompleted + Complete/Fail + depth check + heartbeat**

Edit `internal/agent/tool/subagent_manager.go`:

Add to constants:
```go
const (
	StatusRunning   AgentStatus = "running"
	StatusFailed    AgentStatus = "failed"
	StatusCancelled AgentStatus = "cancelled"
	StatusCompleted AgentStatus = "completed"
)
```

Add to SubAgent struct after `Error` field:
```go
	lastHeartbeat time.Time
```

Modify `Spawn` to check MaxDepth and set heartbeat:

After `cfg.MaxDepth` check block (before `childCtx, cancel`):

```go
	if cfg.MaxDepth <= 0 {
		return &SubAgent{
			ID:     uuid.New().String(),
			Name:   cfg.Name,
			Config: cfg,
			Status: StatusFailed,
			Error:  "max nesting depth reached",
		}
	}
```

After `m.agents[sa.ID] = sa` in Spawn, add heartbeat goroutine:

```go
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-childCtx.Done():
				return
			case <-ticker.C:
				m.mu.Lock()
				if a, ok := m.agents[sa.ID]; ok {
					if time.Since(a.lastHeartbeat) > 2*30*time.Second {
						a.Status = StatusFailed
						a.Error = "agent heartbeat timeout"
						a.cancel()
					}
				}
				m.mu.Unlock()
			}
		}
	}()
```

Set lastHeartbeat when creating SubAgent:
```go
	sa := &SubAgent{
		ID:            uuid.New().String(),
		Name:          cfg.Name,
		Config:        cfg,
		Status:        StatusRunning,
		CreatedAt:     time.Now(),
		lastHeartbeat: time.Now(),
		ctx:           childCtx,
		cancel:        cancel,
	}
```

Add new methods after `RunningCount`:

```go
func (m *SubagentManager) Complete(id, result string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		a.Status = StatusCompleted
		a.Result = result
	}
}

func (m *SubagentManager) Fail(id, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		a.Status = StatusFailed
		a.Error = errMsg
	}
}

func (m *SubagentManager) Heartbeat(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		a.lastHeartbeat = time.Now()
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd G:\mllm\agent5 && go test ./internal/agent/tool/ -v -run "TestSubagentManager"`
Expected: All 9 tests pass

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/subagent_manager.go internal/agent/tool/subagent_manager_test.go
git commit -m "feat(tool): add Complete/Fail, depth limit, heartbeat to SubagentManager"
```

---

### Task 3: TaskTool full execution + SubagentRunner interface

**Files:**
- Modify: `internal/agent/tool/task.go`
- Modify: `internal/agent/tool/task_test.go`
- Create: `internal/agent/runtime/runner.go`

- [ ] **Step 1: Define SubagentRunner interface + update TaskTool schema**

Edit `internal/agent/tool/task.go`:

```go
package tool

import (
	"context"
	"fmt"
	"time"
)

type SubagentRunner interface {
	Run(sessionID, task, systemPrompt string, tools *Registry) (string, error)
}

type TaskTool struct {
	Manager *SubagentManager
	Runner  SubagentRunner
	Session *session.Manager
	Tools   *Registry
	Depth   int
}

func (t *TaskTool) Name() string { return "task" }

func (t *TaskTool) Description() string {
	return "Delegate a task to a specialized subagent. The subagent runs independently with its own context."
}

func (t *TaskTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"description": {
				Type:        "string",
				Description: "A short (3-5 word) description of the task",
			},
			"prompt": {
				Type:        "string",
				Description: "The detailed task for the subagent to perform",
			},
			"subagent_type": {
				Type:        "string",
				Description: "Type of subagent: general, explore, review",
			},
			"background": {
				Type:        "bool",
				Description: "Run in background (return agent_id immediately)",
			},
		},
		Required: []string{"description", "prompt", "subagent_type"},
	}
}

func (t *TaskTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	description, _ := params["description"].(string)
	prompt, _ := params["prompt"].(string)
	subagentType, _ := params["subagent_type"].(string)
	background, _ := params["background"].(bool)

	if description == "" || prompt == "" || subagentType == "" {
		return ToolResult{Error: "description, prompt, and subagent_type are required"}
	}

	if t.Depth <= 0 {
		return ToolResult{Error: "max subagent nesting depth reached"}
	}
	if t.Runner == nil {
		return ToolResult{Error: "SubagentRunner not configured"}
	}

	subTools := filterTools(t.Tools, subagentType)

	cfg := SubAgentConfig{
		Name:         subagentType,
		SystemPrompt: buildSubagentPrompt(subagentType, description),
		Model:        subagentType,
		MaxDepth:     t.Depth - 1,
	}

	sa := t.Manager.Spawn(context.Background(), cfg)
	if sa.Status == StatusFailed {
		return ToolResult{
			Error: fmt.Sprintf("failed to spawn subagent: %s", sa.Error),
		}
	}

	sessionID := sa.ID
	if t.Session != nil {
		parentID := ""
		// Try to get parent session ID from context params
		if p, ok := params["session_id"].(string); ok {
			parentID = p
		}
		sessionID = t.Session.CreateChildSession(parentID, "subagent:"+subagentType)
	}

	if background {
		go func() {
			result, err := t.Runner.Run(sessionID, prompt, cfg.SystemPrompt, subTools)
			if err != nil {
				t.Manager.Fail(sa.ID, err.Error())
				if t.Session != nil {
					t.Session.AddMessage(sessionID, "system",
						fmt.Sprintf("[subagent:%s] failed: %s", sa.ID, err.Error()))
				}
				return
			}
			t.Manager.Complete(sa.ID, result)
		}()

		return ToolResult{
			Success: true,
			Data:    fmt.Sprintf("[subagent:%s] task %q started in background (session: %s)", sa.ID, description, sessionID),
		}
	}

	result, err := t.Runner.Run(sessionID, prompt, cfg.SystemPrompt, subTools)
	if err != nil {
		t.Manager.Fail(sa.ID, err.Error())
		return ToolResult{Error: err.Error()}
	}
	t.Manager.Complete(sa.ID, result)

	return ToolResult{
		Success: true,
		Data:    result,
	}
}

func filterTools(reg *Registry, agentType string) *Registry {
	out := NewRegistry()
	if reg == nil {
		return out
	}
	switch agentType {
	case "explore":
		for _, name := range []string{"read_file", "search_text"} {
			if t, ok := reg.Get(name); ok {
				out.Register(t)
			}
		}
	case "review":
		for _, name := range []string{"read_file", "search_text", "exec_command"} {
			if t, ok := reg.Get(name); ok {
				out.Register(t)
			}
		}
	default:
		for _, t := range reg.List() {
			out.Register(t)
		}
	}
	return out
}

func buildSubagentPrompt(agentType, description string) string {
	switch agentType {
	case "explore":
		return fmt.Sprintf("You are a read-only code exploration agent. Task: %s. Search and analyze code without making changes.", description)
	case "review":
		return fmt.Sprintf("You are a code review agent. Task: %s. Review code for bugs, security issues, and best practices.", description)
	default:
		return fmt.Sprintf("You are a general-purpose agent. Task: %s.", description)
	}
}
```

Add import for `time`.

- [ ] **Step 2: Update TaskTool tests**

Edit `internal/agent/tool/task_test.go`:

```go
package tool

import (
	"testing"
)

func TestTaskToolName(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr, Depth: 3}
	if tool.Name() != "task" {
		t.Errorf("expected name 'task', got %s", tool.Name())
	}
}

func TestTaskToolSchema(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr, Depth: 3}
	schema := tool.Schema()
	if schema.Parameters == nil {
		t.Fatal("expected non-nil parameters")
	}
	for _, name := range []string{"description", "prompt", "subagent_type"} {
		if _, ok := schema.Parameters[name]; !ok {
			t.Errorf("missing required parameter: %s", name)
		}
	}
	if _, ok := schema.Parameters["background"]; !ok {
		t.Error("missing optional parameter: background")
	}
}

type mockRunner struct {
	result string
	err    error
}

func (r *mockRunner) Run(sessionID, task, systemPrompt string, tools *Registry) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return r.result, nil
}

func TestTaskToolExecute(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{
		Manager: mgr,
		Runner:  &mockRunner{result: "task completed"},
		Depth:   3,
	}

	result := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "test task",
		"prompt":        "do something",
		"subagent_type": "general",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	data, ok := result.Data.(string)
	if !ok {
		t.Fatalf("expected string data, got %T", result.Data)
	}
	if data != "task completed" {
		t.Errorf("expected 'task completed', got %s", data)
	}
}

func TestTaskToolExecuteMissingFields(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr, Depth: 3}

	result := tool.Execute(ToolContext{}, map[string]interface{}{
		"description": "test",
	})
	if result.Success {
		t.Fatal("expected failure with missing prompt")
	}
	if result.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestTaskToolExecuteMaxConcurrent(t *testing.T) {
	mgr := NewSubagentManager(1)
	tool := &TaskTool{
		Manager: mgr,
		Runner:  &mockRunner{result: "ok"},
		Depth:   3,
	}

	r1 := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "task 1",
		"prompt":        "do 1",
		"subagent_type": "general",
	})
	if !r1.Success {
		t.Fatalf("first task should succeed: %s", r1.Error)
	}

	r2 := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "task 2",
		"prompt":        "do 2",
		"subagent_type": "general",
	})
	if r2.Success {
		t.Fatal("second task should fail (concurrency limit)")
	}
}

func TestTaskToolExecuteBackground(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{
		Manager: mgr,
		Runner:  &mockRunner{result: "background done"},
		Depth:   3,
	}

	result := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "bg task",
		"prompt":        "do bg work",
		"subagent_type": "general",
		"background":    true,
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	data, ok := result.Data.(string)
	if !ok {
		t.Fatalf("expected string data, got %T", result.Data)
	}
	if data == "" {
		t.Fatal("expected non-empty background result")
	}
}

func TestTaskToolExecuteDepthLimit(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{
		Manager: mgr,
		Runner:  &mockRunner{result: "ok"},
		Depth:   0,
	}

	result := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "deep task",
		"prompt":        "go deeper",
		"subagent_type": "general",
	})
	if result.Success {
		t.Fatal("expected failure due to depth limit")
	}
}
```

- [ ] **Step 3: Run TaskTool tests to see them fail**

Run: `cd G:\mllm\agent5 && go test ./internal/agent/tool/ -v -run "TestTaskTool"`
Expected: Tests fail (needs session import, need to update task.go)

But wait — `task.go` now imports `session` package. Does this create a circular dependency? Let me check:
- `tool` package does NOT currently import `session`
- `session` package imports `ai` and `history` — does not import `tool`
- So `tool` importing `session` is fine, no circular dep

But wait, `runtime` package imports both `session` and `tool`. And `tool` would now import `session`. So:
- `tool` → `session` ✓ (no cycle)
- `runtime` → `session` + `tool` ✓ (runtime imports tool, but tool doesn't import runtime)
- No circular dependency.

Good.

- [ ] **Step 4: Implement SubagentRunner in runtime package**

Create `internal/agent/runtime/runner.go`:

```go
package runtime

import (
	"github.com/example/agent-tui/internal/agent/session"
	"github.com/example/agent-tui/internal/agent/tool"
)

type AgentRunner struct {
	BaseConfig Config
	LLM        LLMClient
	Session    *session.Manager
}

func (r *AgentRunner) Run(sessionID, task, systemPrompt string, tools *tool.Registry) (string, error) {
	cfg := r.BaseConfig
	cfg.SystemPrompt = systemPrompt
	agent := NewAgent(cfg, tools, r.LLM, r.Session)
	return agent.Execute(sessionID, task)
}
```

- [ ] **Step 5: Run all tool package tests**

Run: `cd G:\mllm\agent5 && go test ./internal/agent/tool/ -v`
Expected: All tests pass (or at least TaskTool tests)

But wait — `task.go` now imports `session`. The `session` package is under `internal/agent/session/`. Let me double-check the import path:

The module is `github.com/example/agent-tui`. So import would be: `"github.com/example/agent-tui/internal/agent/session"`. Let me check existing imports in the codebase for the session package.

From runtime/agent.go:
```go
import (
	"github.com/example/agent-tui/internal/agent/session"
	"github.com/example/agent-tui/internal/agent/tool"
)
```

Good, that's the right path.

- [ ] **Step 6: Commit**

```bash
git add internal/agent/tool/task.go internal/agent/tool/task_test.go internal/agent/runtime/runner.go
git commit -m "feat: add SubagentRunner, full TaskTool execution, background mode, depth limit"
```

---

### Task 4: Wire everything in main.go

**Files:**
- Modify: `cmd/agent/main.go`

- [ ] **Step 1: Update main.go to use cfg.MaxSubagents and wire AgentRunner**

Edit `cmd/agent/main.go` around the tool registration section (after agent role setup, before orch creation):

Find the section:
```go
	subagentMgr := tool.NewSubagentManager(5)
	toolReg.Register(&tool.TaskTool{Manager: subagentMgr})
```

Replace with:
```go
	subagentMgr := tool.NewSubagentManager(cfg.MaxSubagents)
	agentRunner := &runtime.AgentRunner{
		BaseConfig: runtime.Config{
			Model:        cfg.DefaultClient,
			MaxReActLoop: 20,
			Temperature:  0.7,
			ContextLimit: 50,
			ApprovalFn:   approvalFn,
		},
		LLM:     agentLLM,
		Session: sm,
	}
	toolReg.Register(&tool.TaskTool{
		Manager: subagentMgr,
		Runner:  agentRunner,
		Session: sm,
		Tools:   toolReg,
		Depth:   3,
	})
```

Add import for `"github.com/example/agent-tui/internal/agent/runtime"` — already imported at line 14.

- [ ] **Step 2: Build**

Run: `cd G:\mllm\agent5 && go build ./cmd/agent/`
Expected: Build succeeds

- [ ] **Step 3: Run all tests**

Run: `cd G:\mllm\agent5 && go test ./... 2>&1`
Expected: All tests pass (except pre-existing opencode integration test failure)

- [ ] **Step 4: Commit**

```bash
git add cmd/agent/main.go
git commit -m "feat(cmd): wire SubagentRunner, config-driven concurrency, full TaskTool"
```

---

### Self-Review

**1. Spec coverage:**
- Child session isolation: Task 1 ✓
- Complete/Fail on Subagent: Task 2 ✓
- Depth limiting: Task 2 (Spawn check) + Task 3 (Depth field) ✓
- Heartbeat/liveness: Task 2 (30s ticker) ✓
- Full agent.Execute in TaskTool: Task 3 (Runner.Run) ✓
- Background mode: Task 3 (background param) ✓
- Tool filtering: Task 3 (filterTools) ✓
- Configurable concurrency: Task 4 (cfg.MaxSubagents) ✓

**2. Placeholder scan:** No TBD/TODO found.

**3. Type consistency:**
- `AgentRunner.Run(sessionID, task, systemPrompt, tools)` matches `SubagentRunner` interface signature ✓
- `SubAgentConfig.MaxDepth` used in Spawn (Task 2) and TaskTool (Task 3) ✓
- `tool.NewRegistry()` returns `*Registry` with `Register`, `Get`, `List` methods — used in filterTools ✓
- `session.Manager.CreateChildSession(parentID, name)` matches history method ✓
- `mockRunner` in test satisfies `SubagentRunner` interface ✓
