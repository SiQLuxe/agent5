# Subagent 并发执行 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add concurrent subagent execution to Agent-TUI in two phases — Phase 1 makes the existing Orchestrator dispatch independent tasks concurrently; Phase 2 gives the LLM a `task` tool to spawn subagents dynamically.

**Architecture:** Phase 1 adds `DispatchConcurrent()` to `internal/agent/orchestrator/orchestrator.go` using goroutines + semaphore channel + WaitGroup. Phase 2 adds `SubagentManager` and `TaskTool` in `internal/agent/tool/` — each subagent runs in its own goroutine with its own session, isolated context, and configurable concurrency cap.

**Tech Stack:** Go 1.26, goroutines/channels, existing `session.Manager`, existing `tool.Registry`, existing `agent/orchestrator` package.

---

### Task 1: Add DispatchConcurrent to orchestrator

**Files:**
- Modify: `internal/agent/orchestrator/orchestrator.go`
- Modify: `internal/agent/orchestrator/types.go`
- Test: `internal/agent/orchestrator/orchestrator_test.go`

- [ ] **Step 1: Add MaxConcurrent field and DispatchConcurrent method**

Edit `internal/agent/orchestrator/orchestrator.go`:

```go
package orchestrator

import (
	"fmt"
	"sync"
)

type Orchestrator struct {
	registry      *Registry
	decomposer    *Decomposer
	merger        *Merger
	MaxConcurrent int
}

func NewOrchestrator(reg *Registry, d *Decomposer, m *Merger) *Orchestrator {
	return &Orchestrator{
		registry:      reg,
		decomposer:    d,
		merger:        m,
		MaxConcurrent: 5,
	}
}

func (o *Orchestrator) DispatchConcurrent(sessionID string, task *Task) ([]*Task, error) {
	steps, err := o.decomposer.Decompose(task)
	if err != nil {
		return nil, fmt.Errorf("decompose: %w", err)
	}

	if len(steps) == 0 {
		return nil, nil
	}

	sem := make(chan struct{}, o.MaxConcurrent)
	type stepResult struct {
		step *Task
		err  error
	}
	resultCh := make(chan stepResult, len(steps))

	for _, step := range steps {
		sem <- struct{}{}
		go func(s *Task) {
			defer func() { <-sem }()
			agents := o.registry.FindByRole(string(s.Type))
			if len(agents) == 0 {
				s.Status = StatusFailed
				resultCh <- stepResult{s, fmt.Errorf("no agent found for role %q", s.Type)}
				return
			}
			agent := agents[0]
			s.Status = StatusRunning
			s.AgentID = agent.Name
			res, err := agent.Execute(sessionID, s.Content)
			if err != nil {
				s.Status = StatusFailed
				s.Error = err.Error()
				resultCh <- stepResult{s, err}
				return
			}
			s.Status = StatusCompleted
			s.Result = res
			resultCh <- stepResult{s, nil}
		}(step)
	}

	// Wait for all goroutines by filling the semaphore back up
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
	close(resultCh)

	var results []*Task
	var firstErr error
	for r := range resultCh {
		results = append(results, r.step)
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
	}

	return results, firstErr
}
```

- [ ] **Step 2: Write concurrent dispatch test**

Edit `internal/agent/orchestrator/orchestrator_test.go`, add after `TestOrchestratorNoAgent`:

```go
func TestOrchestratorDispatchConcurrent(t *testing.T) {
	reg := NewRegistry()

	mockLLM := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{Type: "final", Content: "analyze done"},
		{Type: "final", Content: "code done"},
		{Type: "final", Content: "review done"},
	})
	sm := session.NewManager()
	analyzer := runtime.NewAgent(runtime.Config{Name: "analyzer"}, tool.NewRegistry(), mockLLM, sm)
	coder := runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), mockLLM, sm)
	reviewer := runtime.NewAgent(runtime.Config{Name: "reviewer"}, tool.NewRegistry(), mockLLM, sm)

	reg.Register("analyzer", analyzer, "task_analyze")
	reg.Register("coder", coder, "task_code")
	reg.Register("reviewer", reviewer, "task_review")

	d := NewDecomposer()
	d.Register(TaskDesign, func(task *Task) ([]*Task, error) {
		return []*Task{
			{ID: "t-analyze", Type: TaskAnalyze, Content: "analyze: " + task.Content},
			{ID: "t-code", Type: TaskCode, Content: "code: " + task.Content},
			{ID: "t-review", Type: TaskReview, Content: "review: " + task.Content},
		}, nil
	})

	orch := NewOrchestrator(reg, d, NewMerger())
	orch.MaxConcurrent = 3

	task := &Task{ID: "t1", Type: TaskDesign, Content: "add login feature"}
	results, err := orch.DispatchConcurrent("test-session", task)
	if err != nil {
		t.Fatalf("DispatchConcurrent failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Status != StatusCompleted {
			t.Errorf("expected completed status for %s, got %s", r.ID, r.Status)
		}
	}
}

func TestDispatchConcurrentPartialFailure(t *testing.T) {
	reg := NewRegistry()

	sm := session.NewManager()
	// analyzer will succeed
	okLLM := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{Type: "final", Content: "analysis ok"},
	})
	// coder will fail — no mock responses means empty
	failLLM := runtime.NewMockLLMClient(nil)
	reviewerLLM := runtime.NewMockLLMClient([]runtime.LLMResponse{
		{Type: "final", Content: "review ok"},
	})

	analyzer := runtime.NewAgent(runtime.Config{Name: "analyzer"}, tool.NewRegistry(), okLLM, sm)
	coder := runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), failLLM, sm)
	reviewer := runtime.NewAgent(runtime.Config{Name: "reviewer"}, tool.NewRegistry(), reviewerLLM, sm)

	reg.Register("analyzer", analyzer, "task_analyze")
	reg.Register("coder", coder, "task_code")
	reg.Register("reviewer", reviewer, "task_review")

	d := NewDecomposer()
	d.Register(TaskDesign, func(task *Task) ([]*Task, error) {
		return []*Task{
			{ID: "t-analyze", Type: TaskAnalyze, Content: "analyze"},
			{ID: "t-code", Type: TaskCode, Content: "code"},
			{ID: "t-review", Type: TaskReview, Content: "review"},
		}, nil
	})

	orch := NewOrchestrator(reg, d, NewMerger())
	orch.MaxConcurrent = 3

	task := &Task{ID: "t1", Type: TaskDesign, Content: "add login"}
	results, err := orch.DispatchConcurrent("test-session", task)
	if err == nil {
		t.Fatal("expected error for partial failure")
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}
```

- [ ] **Step 3: Run the tests**

```bash
cd G:\mllm\agent5
go test ./internal/agent/orchestrator/ -v -run "TestOrchestrator|TestDispatchConcurrent"
```

Expected: All existing tests pass + 2 new tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/agent/orchestrator/orchestrator.go internal/agent/orchestrator/orchestrator_test.go
git commit -m "feat(orchestrator): add DispatchConcurrent with goroutine pool"
```

---

### Task 2: Create SubagentManager

**Files:**
- Create: `internal/agent/tool/subagent_manager.go`
- Test: `internal/agent/tool/subagent_manager_test.go`

- [ ] **Step 1: Write failing manager test**

Create `internal/agent/tool/subagent_manager_test.go`:

```go
package tool

import (
	"context"
	"sync"
	"testing"
)

func TestSubagentManagerSpawn(t *testing.T) {
	mgr := NewSubagentManager(5)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sa := mgr.Spawn(ctx, SubAgentConfig{
		Name:         "test-agent",
		SystemPrompt: "You are a test agent",
		Model:        "test-model",
	})
	if sa.ID == "" {
		t.Fatal("expected non-empty agent ID")
	}
	if sa.Name != "test-agent" {
		t.Errorf("expected name 'test-agent', got %s", sa.Name)
	}
	if sa.Status != "running" {
		t.Errorf("expected status 'running', got %s", sa.Status)
	}

	got := mgr.Get(sa.ID)
	if got == nil {
		t.Fatal("expected to find spawned agent")
	}
	if got.ID != sa.ID {
		t.Errorf("expected ID %s, got %s", sa.ID, got.ID)
	}
}

func TestSubagentManagerConcurrencyLimit(t *testing.T) {
	mgr := NewSubagentManager(2) // max 2
	ctx := context.Background()

	sa1 := mgr.Spawn(ctx, SubAgentConfig{Name: "agent1"})
	sa2 := mgr.Spawn(ctx, SubAgentConfig{Name: "agent2"})
	sa3 := mgr.Spawn(ctx, SubAgentConfig{Name: "agent3"})

	if sa1.Status != "running" {
		t.Errorf("sa1 should be running, got %s", sa1.Status)
	}
	if sa2.Status != "running" {
		t.Errorf("sa2 should be running, got %s", sa2.Status)
	}
	if sa3.Status != "failed" {
		t.Errorf("sa3 should be failed (limit), got %s", sa3.Status)
	}
	if sa3.Error == "" {
		t.Fatal("expected error message for agent3")
	}
}

func TestSubagentManagerCancel(t *testing.T) {
	mgr := NewSubagentManager(5)
	ctx, cancel := context.WithCancel(context.Background())

	sa := mgr.Spawn(ctx, SubAgentConfig{Name: "test-agent"})
	cancel()
	if !mgr.IsCancelled(sa.ID) {
		t.Fatal("expected agent to be cancelled after context cancel")
	}
}

func TestSubagentManagerList(t *testing.T) {
	mgr := NewSubagentManager(5)
	ctx := context.Background()

	mgr.Spawn(ctx, SubAgentConfig{Name: "alpha"})
	mgr.Spawn(ctx, SubAgentConfig{Name: "beta"})

	agents := mgr.List()
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}
}
```

- [ ] **Step 2: Run tests to see them fail**

```bash
cd G:\mllm\agent5
go test ./internal/agent/tool/ -v -run "TestSubagentManager"
```

Expected: Build failure (no SubagentManager defined).

- [ ] **Step 3: Create SubagentManager implementation**

Create `internal/agent/tool/subagent_manager.go`:

```go
package tool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SubAgentConfig struct {
	Name         string
	SystemPrompt string
	Model        string
	MaxDepth     int
}

type SubAgent struct {
	ID        string
	Name      string
	Config    SubAgentConfig
	Status    string
	Result    string
	Error     string
	CreatedAt time.Time
	cancel    context.CancelFunc
}

type SubagentManager struct {
	mu       sync.RWMutex
	agents   map[string]*SubAgent
	max      int
}

func NewSubagentManager(max int) *SubagentManager {
	if max <= 0 {
		max = 5
	}
	return &SubagentManager{
		agents: make(map[string]*SubAgent),
		max:    max,
	}
}

func (m *SubagentManager) Spawn(ctx context.Context, cfg SubAgentConfig) *SubAgent {
	m.mu.Lock()
	defer m.mu.Unlock()

	running := 0
	for _, a := range m.agents {
		if a.Status == "running" {
			running++
		}
	}
	if running >= m.max {
		return &SubAgent{
			ID:     uuid.New().String(),
			Name:   cfg.Name,
			Config: cfg,
			Status: "failed",
			Error:  fmt.Sprintf("max concurrent agents reached (%d)", m.max),
		}
	}

	childCtx, cancel := context.WithCancel(ctx)
	sa := &SubAgent{
		ID:        uuid.New().String(),
		Name:      cfg.Name,
		Config:    cfg,
		Status:    "running",
		CreatedAt: time.Now(),
		cancel:    cancel,
	}
	m.agents[sa.ID] = sa

	go func() {
		<-childCtx.Done()
		m.mu.Lock()
		if a, ok := m.agents[sa.ID]; ok && a.Status == "running" {
			a.Status = "cancelled"
		}
		m.mu.Unlock()
	}()

	return sa
}

func (m *SubagentManager) Get(id string) *SubAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.agents[id]
}

func (m *SubagentManager) List() []*SubAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*SubAgent, 0, len(m.agents))
	for _, a := range m.agents {
		out = append(out, a)
	}
	return out
}

func (m *SubagentManager) Cancel(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		a.cancel()
		a.Status = "cancelled"
	}
}

func (m *SubagentManager) IsCancelled(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if a, ok := m.agents[id]; ok {
		return a.Status == "cancelled"
	}
	return false
}

func (m *SubagentManager) RunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, a := range m.agents {
		if a.Status == "running" {
			count++
		}
	}
	return count
}
```

- [ ] **Step 4: Run tests to make them pass**

```bash
cd G:\mllm\agent5
go test ./internal/agent/tool/ -v -run "TestSubagentManager"
```

Expected: All 4 tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/agent/tool/subagent_manager.go internal/agent/tool/subagent_manager_test.go
git commit -m "feat(tool): add SubagentManager with concurrency cap"
```

---

### Task 3: Create TaskTool

**Files:**
- Create: `internal/agent/tool/task.go`
- Test: `internal/agent/tool/task_test.go`

- [ ] **Step 1: Write failing task tool test**

Create `internal/agent/tool/task_test.go`:

```go
package tool

import (
	"testing"
)

func TestTaskToolName(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr}
	if tool.Name() != "task" {
		t.Errorf("expected name 'task', got %s", tool.Name())
	}
}

func TestTaskToolSchema(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr}
	schema := tool.Schema()
	if schema.Parameters == nil {
		t.Fatal("expected non-nil parameters")
	}
	for _, name := range []string{"description", "prompt", "subagent_type"} {
		if _, ok := schema.Parameters[name]; !ok {
			t.Errorf("missing required parameter: %s", name)
		}
	}
}

func TestTaskToolExecute(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr}

	result := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":    "test task",
		"prompt":         "do something",
		"subagent_type":  "general",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	data, ok := result.Data.(string)
	if !ok {
		t.Fatalf("expected string data, got %T", result.Data)
	}
	if data == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestTaskToolExecuteMissingFields(t *testing.T) {
	mgr := NewSubagentManager(5)
	tool := &TaskTool{Manager: mgr}

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
	mgr := NewSubagentManager(1) // max 1
	tool := &TaskTool{Manager: mgr}

	// first succeeds
	r1 := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "task 1",
		"prompt":        "do 1",
		"subagent_type": "general",
	})
	if !r1.Success {
		t.Fatalf("first task should succeed: %s", r1.Error)
	}

	// second fails (limit)
	r2 := tool.Execute(ToolContext{}, map[string]interface{}{
		"description":   "task 2",
		"prompt":        "do 2",
		"subagent_type": "general",
	})
	if r2.Success {
		t.Fatal("second task should fail (concurrency limit)")
	}
}
```

- [ ] **Step 2: Run tests to see them fail**

```bash
cd G:\mllm\agent5
go test ./internal/agent/tool/ -v -run "TestTaskTool"
```

Expected: Build failure (no TaskTool defined).

- [ ] **Step 3: Create TaskTool implementation**

Create `internal/agent/tool/task.go`:

```go
package tool

import (
	"context"
	"fmt"
	"time"
)

type TaskTool struct {
	Manager *SubagentManager
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
		},
		Required: []string{"description", "prompt", "subagent_type"},
	}
}

func (t *TaskTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	description, _ := params["description"].(string)
	prompt, _ := params["prompt"].(string)
	subagentType, _ := params["subagent_type"].(string)

	if description == "" || prompt == "" || subagentType == "" {
		return ToolResult{Error: "description, prompt, and subagent_type are required"}
	}

	cfg := SubAgentConfig{
		Name:         subagentType,
		SystemPrompt: buildSubagentPrompt(subagentType, description),
		Model:        subagentType,
	}

	sa := t.Manager.Spawn(context.Background(), cfg)
	if sa.Status == "failed" {
		return ToolResult{
			Error: fmt.Sprintf("failed to spawn subagent: %s", sa.Error),
		}
	}

	// In Phase 2, the subagent would actually run agent.Execute here.
	// For now, return a placeholder result.
	result := fmt.Sprintf("[subagent:%s] task %q started at %s (agent: %s)",
		sa.ID, description, sa.CreatedAt.Format(time.RFC3339), subagentType)

	return ToolResult{
		Success: true,
		Data:    result,
	}
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

- [ ] **Step 4: Run tests to make them pass**

```bash
cd G:\mllm\agent5
go test ./internal/agent/tool/ -v -run "TestTaskTool"
```

Expected: All 5 tests pass.

- [ ] **Step 5: Run full tool package tests**

```bash
cd G:\mllm\agent5
go test ./internal/agent/tool/ -v
```

Expected: All tests pass (existing + new).

- [ ] **Step 6: Commit**

```bash
git add internal/agent/tool/task.go internal/agent/tool/task_test.go
git commit -m "feat(tool): add TaskTool for subagent delegation"
```

---

### Task 4: Wire TaskTool into main.go

**Files:**
- Modify: `cmd/agent/main.go`

- [ ] **Step 1: Register TaskTool in main.go**

Edit `cmd/agent/main.go`. After `toolReg.Register(&tool.ChatLLMTool{...})`, add:

```go
	subagentMgr := tool.NewSubagentManager(5)
	toolReg.Register(&tool.TaskTool{Manager: subagentMgr})
```

Find the exact insertion point (around line 179).

- [ ] **Step 2: Build**

```bash
cd G:\mllm\agent5
go build ./cmd/agent/
```

Expected: Build succeeds with no errors.

- [ ] **Step 3: Run all tests**

```bash
cd G:\mllm\agent5
go test ./...
```

Expected: All tests pass.

- [ ] **Step 4: Commit**

```bash
git add cmd/agent/main.go
git commit -m "feat(cmd): wire TaskTool into agent registry"
```

---

### Self-Review

**1. Spec coverage:**
- Phase 1 (DispatchConcurrent): Task 1 ✓
- Phase 2 (SubagentManager): Task 2 ✓
- Phase 2 (TaskTool): Task 3 ✓
- Phase 2 (wiring): Task 4 ✓

**2. Placeholder scan:** No TBD/TODO found.

**3. Type consistency:** All types match between tests and implementation.
- `SubAgentConfig` fields match across test and impl
- `TaskTool.Manager` field type matches `*SubagentManager`
- `ToolResult{Success, Data, Error}` matches existing pattern

**4. Missing from spec (deferred to future):**
- Background mode (Phase 3)
- Per-subagent tool filtering (current implementation gives all tools)
- Depth limiting (SubAgentConfig.MaxDepth exists but unused)
- Progress streaming to UI
- Full agent.Execute integration in TaskTool (currently returns placeholder)
