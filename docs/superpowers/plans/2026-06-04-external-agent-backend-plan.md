# External Agent Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an abstraction layer (`AgentBackend` interface) to Agent-TUI that connects to external AI coding agents, starting with opencode via HTTP REST API, plus a reverse-control HTTP server for bidirectional communication.

**Architecture:** Adapter pattern — `AgentBackend` interface defines the contract, `OpencodeBackend` implements it via subprocess-managed `opencode serve`. `BackendRegistry` manages multiple backends. Agent-TUI also exposes a small HTTP server for external agents to control the TUI. A new `ExternalAgent` role wraps the backend into the existing agent system.

**Tech Stack:** Go 1.26, standard library `net/http` + `encoding/json`, `context` for cancellation/timeouts.

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/backend/backend.go` | Create | `AgentBackend` interface, data models (`Session`, `Message`, `Event`, etc.) |
| `internal/backend/factory.go` | Create | `NewBackend(agentType, config)` factory |
| `internal/backend/registry.go` | Create | `BackendRegistry` — register, lookup, list backends |
| `internal/backend/process.go` | Create | `ProcessManager` — subprocess lifecycle, health check, restart logic |
| `internal/backend/opencode/backend.go` | Create | `OpencodeBackend` struct, `Start`/`Stop`/`Health` |
| `internal/backend/opencode/sessions.go` | Create | Session CRUD API calls |
| `internal/backend/opencode/messages.go` | Create | `SendMessage`/`SendMessageStream`/`GetMessages` |
| `internal/backend/opencode/commands.go` | Create | `ExecuteCommand`/`ExecuteShell`/`ReadFile`/`SearchText` |
| `internal/backend/opencode/events.go` | Create | SSE connection, event dispatch |
| `internal/server/server.go` | Create | Agent-TUI reverse-control HTTP server |
| `internal/server/handlers.go` | Create | TUI control handlers |
| `internal/service/agent_roles.go` | Modify | Add `ExternalAgent` + `TaskExternal` |
| `internal/service/task_orchestrator.go` | Modify | Register `TaskExternal` type |
| `internal/config/config.go` | Modify | Add backend config section |

---

### Task 1: AgentBackend interface and data models

**Files:**
- Create: `internal/backend/backend.go`

- [ ] **Step 1: Write test for the interface contract**

```go
// internal/backend/backend_test.go
package backend

import (
    "context"
    "testing"
    "time"
)

func TestAgentTypes(t *testing.T) {
    if TypeOpencode != "opencode" {
        t.Errorf("TypeOpencode = %q, want %q", TypeOpencode, "opencode")
    }
    if TypeClaudeCode != "claude-code" {
        t.Errorf("TypeClaudeCode = %q, want %q", TypeClaudeCode, "claude-code")
    }
}

func TestSessionModel(t *testing.T) {
    now := time.Now()
    s := Session{
        ID: "sess-1", Title: "test", CreatedAt: now, Status: "active",
    }
    if s.ID != "sess-1" || s.Title != "test" || s.Status != "active" {
        t.Errorf("Session fields not set correctly")
    }
}

func TestMessageModel(t *testing.T) {
    m := Message{Role: "user", Content: "hello"}
    if m.Role != "user" || m.Content != "hello" {
        t.Errorf("Message fields not set correctly")
    }
}

func TestEventModel(t *testing.T) {
    e := Event{Type: "message.completed", Payload: "ok"}
    if e.Type != "message.completed" {
        t.Errorf("Event type not set correctly")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/ -v 2>&1`
Expected: FAIL — package doesn't exist

- [ ] **Step 3: Create the interface and data models**

```go
// internal/backend/backend.go
package backend

import (
    "context"
    "time"
)

type AgentType string

const (
    TypeOpencode   AgentType = "opencode"
    TypeClaudeCode AgentType = "claude-code"
    TypeCodex      AgentType = "codex"
)

type Session struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    CreatedAt time.Time `json:"created_at"`
    Status    string    `json:"status"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type MessageResult struct {
    SessionID string `json:"session_id"`
    MessageID string `json:"message_id"`
    Content   string `json:"content"`
}

type Chunk struct {
    Content string `json:"content"`
    Done    bool   `json:"done"`
}

type CommandResult struct {
    ExitCode int    `json:"exit_code"`
    Stdout   string `json:"stdout"`
    Stderr   string `json:"stderr"`
}

type SearchResult struct {
    Path       string `json:"path"`
    LineNumber int    `json:"line_number"`
    Content    string `json:"content"`
}

type Event struct {
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}

type HealthInfo struct {
    Healthy bool   `json:"healthy"`
    Version string `json:"version,omitempty"`
}

type BackendConfig struct {
    Type      AgentType `json:"type"`
    Enabled   bool      `json:"enabled"`
    AutoStart bool      `json:"auto_start"`
    Binary    string    `json:"binary"`
    APIURL    string    `json:"api_url,omitempty"`
    APIKey    string    `json:"api_key,omitempty"`
}

type AgentBackend interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health(ctx context.Context) (*HealthInfo, error)

    CreateSession(ctx context.Context, title string) (*Session, error)
    ListSessions(ctx context.Context) ([]*Session, error)
    GetSession(ctx context.Context, id string) (*Session, error)
    DeleteSession(ctx context.Context, id string) error

    SendMessage(ctx context.Context, sessionID string, msg *Message) (*MessageResult, error)
    SendMessageStream(ctx context.Context, sessionID string, msg *Message, onChunk func(*Chunk)) error
    GetMessages(ctx context.Context, sessionID string) ([]*Message, error)

    ExecuteCommand(ctx context.Context, sessionID string, command string) (*CommandResult, error)
    ExecuteShell(ctx context.Context, command string) (*CommandResult, error)

    ReadFile(ctx context.Context, path string) (string, error)
    SearchText(ctx context.Context, pattern string) ([]SearchResult, error)

    Events(ctx context.Context) (<-chan *Event, error)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/backend/backend.go internal/backend/backend_test.go
git commit -m "feat(backend): add AgentBackend interface and data models"
```

---

### Task 2: Factory and Registry

**Files:**
- Create: `internal/backend/factory.go`
- Create: `internal/backend/registry.go`
- Create: `internal/backend/factory_test.go`
- Create: `internal/backend/registry_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/backend/factory_test.go
package backend

import (
    "testing"
)

func TestNewBackend_UnknownType(t *testing.T) {
    _, err := NewBackend("unknown", BackendConfig{})
    if err == nil {
        t.Errorf("expected error for unknown backend type")
    }
}

func TestNewBackend_Opencode(t *testing.T) {
    b, err := NewBackend(string(TypeOpencode), BackendConfig{
        Type: TypeOpencode, Enabled: true, AutoStart: false,
    })
    if err != nil {
        t.Fatalf("NewBackend(opencode) failed: %v", err)
    }
    if b == nil {
        t.Fatal("expected non-nil backend")
    }
}
```

```go
// internal/backend/registry_test.go
package backend

import (
    "testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
    r := NewRegistry()
    mock := &mockBackend{agentType: TypeOpencode}
    r.Register(TypeOpencode, mock)
    got, ok := r.Get(TypeOpencode)
    if !ok {
        t.Fatal("expected to find backend")
    }
    if got != mock {
        t.Errorf("Get returned wrong backend")
    }
}

func TestRegistry_GetAll(t *testing.T) {
    r := NewRegistry()
    m1 := &mockBackend{agentType: TypeOpencode}
    m2 := &mockBackend{agentType: TypeClaudeCode}
    r.Register(TypeOpencode, m1)
    r.Register(TypeClaudeCode, m2)
    all := r.GetAll()
    if len(all) != 2 {
        t.Errorf("expected 2 backends, got %d", len(all))
    }
}

type mockBackend struct {
    agentType AgentType
}

func (m *mockBackend) Start(ctx context.Context) error { return nil }
func (m *mockBackend) Stop(ctx context.Context) error { return nil }
func (m *mockBackend) Health(ctx context.Context) (*HealthInfo, error) {
    return &HealthInfo{Healthy: true}, nil
}
func (m *mockBackend) CreateSession(ctx context.Context, title string) (*Session, error) {
    return &Session{ID: "mock", Title: title}, nil
}
func (m *mockBackend) ListSessions(ctx context.Context) ([]*Session, error) { return nil, nil }
func (m *mockBackend) GetSession(ctx context.Context, id string) (*Session, error) { return nil, nil }
func (m *mockBackend) DeleteSession(ctx context.Context, id string) error { return nil }
func (m *mockBackend) SendMessage(ctx context.Context, sID string, msg *Message) (*MessageResult, error) {
    return &MessageResult{Content: "mock reply"}, nil
}
func (m *mockBackend) SendMessageStream(ctx context.Context, sID string, msg *Message, onChunk func(*Chunk)) error { return nil }
func (m *mockBackend) GetMessages(ctx context.Context, sID string) ([]*Message, error) { return nil, nil }
func (m *mockBackend) ExecuteCommand(ctx context.Context, sID string, cmd string) (*CommandResult, error) { return nil, nil }
func (m *mockBackend) ExecuteShell(ctx context.Context, cmd string) (*CommandResult, error) { return nil, nil }
func (m *mockBackend) ReadFile(ctx context.Context, path string) (string, error) { return "", nil }
func (m *mockBackend) SearchText(ctx context.Context, pattern string) ([]SearchResult, error) { return nil, nil }
func (m *mockBackend) Events(ctx context.Context) (<-chan *Event, error) { return nil, nil }

// Need context import
var _ = context.Background
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/ -v -run "TestNewBackend|TestRegistry"`
Expected: compilation errors (factory.go, registry.go don't exist yet)

- [ ] **Step 3: Implement factory**

```go
// internal/backend/factory.go
package backend

import (
    "fmt"
)

func NewBackend(agentType string, cfg BackendConfig) (AgentBackend, error) {
    switch AgentType(agentType) {
    case TypeOpencode:
        return newOpencodeBackend(cfg)
    default:
        return nil, fmt.Errorf("unknown backend type: %s", agentType)
    }
}

// newOpencodeBackend is defined in opencode/backend.go via the init pattern
// but we call it through an internal registration
var backendBuilders = map[AgentType]func(BackendConfig) (AgentBackend, error){}

func RegisterBackendBuilder(t AgentType, fn func(BackendConfig) (AgentBackend, error)) {
    backendBuilders[t] = fn
}
```

- [ ] **Step 4: Implement registry**

```go
// internal/backend/registry.go
package backend

import "sync"

type Registry struct {
    mu       sync.RWMutex
    backends map[AgentType]AgentBackend
}

func NewRegistry() *Registry {
    return &Registry{
        backends: make(map[AgentType]AgentBackend),
    }
}

func (r *Registry) Register(t AgentType, b AgentBackend) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.backends[t] = b
}

func (r *Registry) Get(t AgentType) (AgentBackend, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    b, ok := r.backends[t]
    return b, ok
}

func (r *Registry) GetAll() []AgentBackend {
    r.mu.RLock()
    defer r.mu.RUnlock()
    result := make([]AgentBackend, 0, len(r.backends))
    for _, b := range r.backends {
        result = append(result, b)
    }
    return result
}

func (r *Registry) ActiveBackends() []AgentBackend {
    all := r.GetAll()
    active := make([]AgentBackend, 0, len(all))
    for _, b := range all {
        h, err := b.Health(nil)
        if err == nil && h != nil && h.Healthy {
            active = append(active, b)
        }
    }
    return active
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/ -v -run "TestNewBackend|TestRegistry"`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/backend/factory.go internal/backend/registry.go internal/backend/factory_test.go internal/backend/registry_test.go
git commit -m "feat(backend): add factory and registry"
```

---

### Task 3: Process manager

**Files:**
- Create: `internal/backend/process.go`
- Create: `internal/backend/process_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/backend/process_test.go
package backend

import (
    "context"
    "testing"
    "time"
)

func TestNewProcessManager(t *testing.T) {
    pm := NewProcessManager("echo", []string{"hello"}, nil)
    if pm == nil {
        t.Fatal("expected non-nil ProcessManager")
    }
}

func TestProcessManager_StartStop(t *testing.T) {
    pm := NewProcessManager("echo", []string{"hello"}, nil)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := pm.Start(ctx)
    if err != nil {
        t.Fatalf("Start failed: %v", err)
    }
    defer pm.Stop(ctx)

    if !pm.Running() {
        t.Errorf("expected process to be running")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/ -v -run TestProcessManager`
Expected: compilation error

- [ ] **Step 3: Implement ProcessManager**

```go
// internal/backend/process.go
package backend

import (
    "context"
    "io"
    "os/exec"
    "sync"
    "time"
)

const (
    defaultStopTimeout  = 5 * time.Second
    maxRestartAttempts  = 3
)

type ProcessManager struct {
    cmdName        string
    cmdArgs        []string
    env            []string
    stdout         io.Writer
    stderr         io.Writer

    mu             sync.Mutex
    cmd            *exec.Cmd
    running        bool
    restartCount   int
}

func NewProcessManager(name string, args []string, env []string) *ProcessManager {
    return &ProcessManager{
        cmdName: name,
        cmdArgs: args,
        env:     env,
    }
}

func (pm *ProcessManager) Start(ctx context.Context) error {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    cmd := exec.CommandContext(ctx, pm.cmdName, pm.cmdArgs...)
    cmd.Env = pm.env
    if pm.stdout != nil {
        cmd.Stdout = pm.stdout
    }
    if pm.stderr != nil {
        cmd.Stderr = pm.stderr
    }

    if err := cmd.Start(); err != nil {
        return err
    }

    pm.cmd = cmd
    pm.running = true
    return nil
}

func (pm *ProcessManager) Stop(ctx context.Context) error {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    if pm.cmd == nil || !pm.running {
        return nil
    }

    if err := pm.cmd.Process.Signal(); err != nil {
        pm.cmd.Process.Kill()
    }

    done := make(chan error, 1)
    go func() {
        done <- pm.cmd.Wait()
    }()

    select {
    case <-done:
    case <-ctx.Done():
        pm.cmd.Process.Kill()
        <-done
    }

    pm.running = false
    pm.cmd = nil
    return nil
}

func (pm *ProcessManager) Running() bool {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    return pm.running
}

func (pm *ProcessManager) Restart(ctx context.Context) error {
    if err := pm.Stop(ctx); err != nil {
        return err
    }
    return pm.Start(ctx)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/ -v -run TestProcessManager`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/backend/process.go internal/backend/process_test.go
git commit -m "feat(backend): add ProcessManager for subprocess lifecycle"
```

---

### Task 4: OpencodeBackend skeleton (Start/Stop/Health)

**Files:**
- Create: `internal/backend/opencode/backend.go`
- Create: `internal/backend/opencode/backend_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/backend/opencode/backend_test.go
package opencode

import (
    "context"
    "testing"
    "time"
)

func TestNewBackend(t *testing.T) {
    b := NewBackend(Config{
        Binary: "opencode",
        AutoStart: false,
        APIURL: "http://127.0.0.1:4096",
    })
    if b == nil {
        t.Fatal("expected non-nil backend")
    }
}

func TestBackend_Name(t *testing.T) {
    b := NewBackend(Config{AutoStart: false, APIURL: "http://127.0.0.1:4096"})
    if b.Name() != "opencode" {
        t.Errorf("Name() = %q, want %q", b.Name(), "opencode")
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/opencode/ -v -run TestNewBackend`
Expected: compilation error

- [ ] **Step 3: Implement opencode backend skeleton**

```go
// internal/backend/opencode/backend.go
package opencode

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/example/agent-tui/internal/backend"
)

type Config struct {
    Binary    string `json:"binary"`
    AutoStart bool   `json:"auto_start"`
    APIURL    string `json:"api_url"`
    APIKey    string `json:"api_key"`
}

type OpencodeBackend struct {
    config    Config
    client    *http.Client
    baseURL   string
    procMgr   *backend.ProcessManager
    sseConn   *sseConnection
}

func NewBackend(cfg Config) *OpencodeBackend {
    return &OpencodeBackend{
        config: cfg,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (b *OpencodeBackend) Name() string {
    return string(backend.TypeOpencode)
}

func (b *OpencodeBackend) Start(ctx context.Context) error {
    if b.config.AutoStart {
        // Start opencode serve as subprocess
        args := []string{"serve"}
        if b.config.Binary == "" {
            b.config.Binary = "opencode"
        }
        b.procMgr = backend.NewProcessManager(b.config.Binary, args, nil)
        if err := b.procMgr.Start(ctx); err != nil {
            return fmt.Errorf("failed to start opencode: %w", err)
        }
        // Wait for server to be ready (poll health endpoint)
        // In auto-start mode, we need to detect the port
        b.baseURL = "http://127.0.0.1:4096" // default, may need port detection
    } else {
        b.baseURL = b.config.APIURL
    }

    // Health check
    h, err := b.Health(ctx)
    if err != nil {
        return fmt.Errorf("opencode health check failed: %w", err)
    }
    if !h.Healthy {
        return fmt.Errorf("opencode is not healthy")
    }
    return nil
}

func (b *OpencodeBackend) Stop(ctx context.Context) error {
    if b.sseConn != nil {
        b.sseConn.Close()
    }
    if b.procMgr != nil {
        return b.procMgr.Stop(ctx)
    }
    return nil
}

func (b *OpencodeBackend) Health(ctx context.Context) (*backend.HealthInfo, error) {
    url := b.baseURL + "/global/health"
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    resp, err := b.client.Do(req)
    if err != nil {
        return &backend.HealthInfo{Healthy: false}, nil
    }
    defer resp.Body.Close()
    return &backend.HealthInfo{Healthy: resp.StatusCode == http.StatusOK}, nil
}
```

- [ ] **Step 4: Fix test import and run**

Fix the test to use the correct import path:

```go
package opencode

import (
    "testing"
)

func TestNewBackend(t *testing.T) {
    b := NewBackend(Config{
        Binary:    "opencode",
        AutoStart: false,
        APIURL:    "http://127.0.0.1:4096",
    })
    if b == nil {
        t.Fatal("expected non-nil backend")
    }
}

func TestBackend_Name(t *testing.T) {
    b := NewBackend(Config{AutoStart: false, APIURL: "http://127.0.0.1:4096"})
    if b.Name() != "opencode" {
        t.Errorf("Name() = %q, want %q", b.Name(), "opencode")
    }
}
```

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/opencode/ -v -run TestNewBackend`
Expected: PASS

- [ ] **Step 5: Register OpencodeBackend in factory**

```go
// In internal/backend/opencode/backend.go, add init function:

func init() {
    backend.RegisterBackendBuilder(backend.TypeOpencode, func(cfg backend.BackendConfig) (backend.AgentBackend, error) {
        return NewBackend(Config{
            Binary:    cfg.Binary,
            AutoStart: cfg.AutoStart,
            APIURL:    cfg.APIURL,
            APIKey:    cfg.APIKey,
        }), nil
    })
}
```

- [ ] **Step 6: Commit**

```bash
git add internal/backend/opencode/
git commit -m "feat(backend): add OpencodeBackend skeleton with Start/Stop/Health"
```

---

### Task 5: OpencodeBackend sessions API

**Files:**
- Create: `internal/backend/opencode/sessions.go`
- Modify: `internal/backend/opencode/backend.go` (add helper method)

- [ ] **Step 1: Implement sessions API**

```go
// internal/backend/opencode/sessions.go
package opencode

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/example/agent-tui/internal/backend"
)

type opencodeSession struct {
    ID        string `json:"id"`
    Title     string `json:"title"`
    CreatedAt string `json:"created_at"`
    Status    string `json:"status"`
}

func (b *OpencodeBackend) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
    url := b.baseURL + path
    var reqBody []byte
    if body != nil {
        var err error
        reqBody, err = json.Marshal(body)
        if err != nil {
            return fmt.Errorf("marshal request: %w", err)
        }
    }

    req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
    if err != nil {
        return fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := b.client.Do(req)
    if err != nil {
        return fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("opencode API error: %s", resp.Status)
    }

    if result != nil {
        if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
            return fmt.Errorf("decode response: %w", err)
        }
    }
    return nil
}

func (b *OpencodeBackend) CreateSession(ctx context.Context, title string) (*backend.Session, error) {
    body := map[string]string{"title": title}
    var os opencodeSession
    if err := b.doRequest(ctx, "POST", "/session", body, &os); err != nil {
        return nil, err
    }
    return mapSession(&os), nil
}

func (b *OpencodeBackend) ListSessions(ctx context.Context) ([]*backend.Session, error) {
    var sessions []opencodeSession
    if err := b.doRequest(ctx, "GET", "/session", nil, &sessions); err != nil {
        return nil, err
    }
    result := make([]*backend.Session, len(sessions))
    for i, s := range sessions {
        result[i] = mapSession(&s)
    }
    return result, nil
}

func (b *OpencodeBackend) GetSession(ctx context.Context, id string) (*backend.Session, error) {
    var os opencodeSession
    if err := b.doRequest(ctx, "GET", "/session/"+id, nil, &os); err != nil {
        return nil, err
    }
    return mapSession(&os), nil
}

func (b *OpencodeBackend) DeleteSession(ctx context.Context, id string) error {
    return b.doRequest(ctx, "DELETE", "/session/"+id, nil, nil)
}

func mapSession(os *opencodeSession) *backend.Session {
    return &backend.Session{
        ID:     os.ID,
        Title:  os.Title,
        Status: os.Status,
    }
}
```

- [ ] **Step 2: Write tests with a test server**

```go
// internal/backend/opencode/sessions_test.go
package opencode

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/example/agent-tui/internal/backend"
)

func setupTestServer() (*httptest.Server, *OpencodeBackend) {
    mux := http.NewServeMux()
    server := httptest.NewServer(mux)

    sessions := []opencodeSession{
        {ID: "s1", Title: "session 1", Status: "active"},
    }

    mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "GET":
            json.NewEncoder(w).Encode(sessions)
        case "POST":
            var s opencodeSession
            json.NewDecoder(r.Body).Decode(&s)
            s.ID = "new-id"
            json.NewEncoder(w).Encode(s)
        }
    })
    mux.HandleFunc("/session/", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "GET":
            json.NewEncoder(w).Encode(&sessions[0])
        case "DELETE":
            w.WriteHeader(http.StatusNoContent)
        }
    })

    b := NewBackend(Config{AutoStart: false, APIURL: server.URL})
    return server, b
}

func TestCreateSession(t *testing.T) {
    srv, b := setupTestServer()
    defer srv.Close()

    sess, err := b.CreateSession(context.Background(), "test session")
    if err != nil {
        t.Fatalf("CreateSession failed: %v", err)
    }
    if sess.Title != "test session" {
        t.Errorf("Title = %q, want %q", sess.Title, "test session")
    }
}

func TestListSessions(t *testing.T) {
    srv, b := setupTestServer()
    defer srv.Close()

    sessions, err := b.ListSessions(context.Background())
    if err != nil {
        t.Fatalf("ListSessions failed: %v", err)
    }
    if len(sessions) != 1 {
        t.Errorf("expected 1 session, got %d", len(sessions))
    }
}

func TestDeleteSession(t *testing.T) {
    srv, b := setupTestServer()
    defer srv.Close()

    err := b.DeleteSession(context.Background(), "s1")
    if err != nil {
        t.Fatalf("DeleteSession failed: %v", err)
    }
}
```

- [ ] **Step 3: Run tests**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/opencode/ -v -run "TestCreateSession|TestListSessions|TestDeleteSession"`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/backend/opencode/sessions.go internal/backend/opencode/sessions_test.go
git commit -m "feat(backend): implement OpencodeBackend sessions API"
```

---

### Task 6: OpencodeBackend messages API

**Files:**
- Create: `internal/backend/opencode/messages.go`
- Create: `internal/backend/opencode/messages_test.go`

- [ ] **Step 1: Implement messages API**

```go
// internal/backend/opencode/messages.go
package opencode

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/example/agent-tui/internal/backend"
)

type opencodeMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type opencodePart struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

type opencodePromptBody struct {
    Parts []opencodePart `json:"parts"`
}

type opencodeMessageResponse struct {
    Info struct {
        ID string `json:"id"`
    } `json:"info"`
    Parts []opencodePart `json:"parts"`
}

func (b *OpencodeBackend) SendMessage(ctx context.Context, sessionID string, msg *backend.Message) (*backend.MessageResult, error) {
    body := opencodePromptBody{
        Parts: []opencodePart{
            {Type: "text", Text: msg.Content},
        },
    }

    var resp opencodeMessageResponse
    if err := b.doRequest(ctx, "POST", "/session/"+sessionID+"/message", body, &resp); err != nil {
        return nil, err
    }

    result := &backend.MessageResult{
        SessionID: sessionID,
        MessageID: resp.Info.ID,
    }
    for _, p := range resp.Parts {
        if p.Type == "text" {
            result.Content += p.Text
        }
    }
    return result, nil
}

func (b *OpencodeBackend) SendMessageStream(ctx context.Context, sessionID string, msg *backend.Message, onChunk func(*backend.Chunk)) error {
    // Send message asynchronously
    body := opencodePromptBody{
        Parts: []opencodePart{
            {Type: "text", Text: msg.Content},
        },
    }
    if err := b.doRequest(ctx, "POST", "/session/"+sessionID+"/prompt_async", body, nil); err != nil {
        return err
    }

    // Listen for response via SSE stream
    events, err := b.Events(ctx)
    if err != nil {
        return err
    }

    for event := range events {
        if event.Type == "message.completed" {
            // Extract content from event payload
            payload, ok := event.Payload.(map[string]interface{})
            if ok {
                if content, exists := payload["content"]; exists {
                    onChunk(&backend.Chunk{Content: fmt.Sprintf("%v", content), Done: false})
                }
            }
            onChunk(&backend.Chunk{Done: true})
            return nil
        }
        // Forward partial content if available
        if event.Type == "message.chunk" {
            payload, ok := event.Payload.(map[string]interface{})
            if ok {
                if content, exists := payload["content"]; exists {
                    onChunk(&backend.Chunk{Content: fmt.Sprintf("%v", content), Done: false})
                }
            }
        }
    }
    return nil
}

func (b *OpencodeBackend) GetMessages(ctx context.Context, sessionID string) ([]*backend.Message, error) {
    var resp struct {
        Info struct {
            ID string `json:"id"`
        } `json:"info"`
        Parts []opencodePart `json:"parts"`
    }

    // Opencode returns an array of {info, parts}
    var messages []struct {
        Info struct {
            ID   string `json:"id"`
            Role string `json:"role"`
        } `json:"info"`
        Parts []opencodePart `json:"parts"`
    }

    if err := b.doRequest(ctx, "GET", "/session/"+sessionID+"/message", nil, &messages); err != nil {
        return nil, err
    }

    result := make([]*backend.Message, len(messages))
    for i, m := range messages {
        content := ""
        for _, p := range m.Parts {
            if p.Type == "text" {
                content += p.Text
            }
        }
        result[i] = &backend.Message{
            Role:    m.Info.Role,
            Content: content,
        }
    }
    return result, nil
}
```

- [ ] **Step 2: Write tests**

```go
// internal/backend/opencode/messages_test.go
package opencode

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/example/agent-tui/internal/backend"
)

func setupMessagesTestServer() (*httptest.Server, *OpencodeBackend) {
    mux := http.NewServeMux()
    server := httptest.NewServer(mux)

    mux.HandleFunc("/session/s1/message", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "GET":
            resp := []map[string]interface{}{
                {
                    "info": map[string]interface{}{
                        "id": "msg1", "role": "assistant",
                    },
                    "parts": []map[string]interface{}{
                        {"type": "text", "text": "Hello from opencode"},
                    },
                },
            }
            json.NewEncoder(w).Encode(resp)
        case "POST":
            if r.URL.Query().Get("async") == "" {
                // Sync: return immediately
                resp := map[string]interface{}{
                    "info": map[string]interface{}{"id": "msg2"},
                    "parts": []map[string]interface{}{
                        {"type": "text", "text": "Sync reply"},
                    },
                }
                json.NewEncoder(w).Encode(resp)
            }
        }
    })
    mux.HandleFunc("/session/s1/prompt_async", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNoContent)
    })
    mux.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
        // SSE not easily testable via httptest; return empty for now
        w.WriteHeader(http.StatusOK)
    })

    b := NewBackend(Config{AutoStart: false, APIURL: server.URL})
    return server, b
}

func TestSendMessage(t *testing.T) {
    srv, b := setupMessagesTestServer()
    defer srv.Close()

    result, err := b.SendMessage(context.Background(), "s1", &backend.Message{
        Role: "user", Content: "hello",
    })
    if err != nil {
        t.Fatalf("SendMessage failed: %v", err)
    }
    if result.Content == "" {
        t.Errorf("expected non-empty content")
    }
}

func TestGetMessages(t *testing.T) {
    srv, b := setupMessagesTestServer()
    defer srv.Close()

    msgs, err := b.GetMessages(context.Background(), "s1")
    if err != nil {
        t.Fatalf("GetMessages failed: %v", err)
    }
    if len(msgs) == 0 {
        t.Errorf("expected at least one message")
    }
}
```

- [ ] **Step 3: Run tests**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/opencode/ -v -run "TestSendMessage|TestGetMessages"`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/backend/opencode/messages.go internal/backend/opencode/messages_test.go
git commit -m "feat(backend): implement OpencodeBackend messages API"
```

---

### Task 7: OpencodeBackend commands and files API

**Files:**
- Create: `internal/backend/opencode/commands.go`
- Create: `internal/backend/opencode/commands_test.go`

- [ ] **Step 1: Implement commands/files API**

```go
// internal/backend/opencode/commands.go
package opencode

import (
    "context"
    "encoding/json"
    "fmt"
    "net/url"

    "github.com/example/agent-tui/internal/backend"
)

type opencodeCommandBody struct {
    Command   string `json:"command"`
    Arguments string `json:"arguments,omitempty"`
}

type opencodeShellBody struct {
    Command string `json:"command"`
}

func (b *OpencodeBackend) ExecuteCommand(ctx context.Context, sessionID string, command string) (*backend.CommandResult, error) {
    body := opencodeCommandBody{Command: command}
    var resp struct {
        Parts []struct {
            Type string `json:"type"`
            Text string `json:"text"`
        } `json:"parts"`
    }
    if err := b.doRequest(ctx, "POST", "/session/"+sessionID+"/command", body, &resp); err != nil {
        return nil, err
    }

    result := &backend.CommandResult{}
    for _, p := range resp.Parts {
        result.Stdout += p.Text
    }
    return result, nil
}

func (b *OpencodeBackend) ExecuteShell(ctx context.Context, command string) (*backend.CommandResult, error) {
    body := opencodeShellBody{Command: command}
    var resp struct {
        Parts []struct {
            Type string `json:"type"`
            Text string `json:"text"`
        } `json:"parts"`
    }
    if err := b.doRequest(ctx, "POST", "/session/shell", body, &resp); err != nil {
        return nil, err
    }

    result := &backend.CommandResult{}
    for _, p := range resp.Parts {
        result.Stdout += p.Text
    }
    return result, nil
}

func (b *OpencodeBackend) ReadFile(ctx context.Context, path string) (string, error) {
    params := url.Values{}
    params.Set("path", path)

    var resp struct {
        Type    string `json:"type"`
        Content string `json:"content"`
    }
    if err := b.doRequest(ctx, "GET", "/file/content?"+params.Encode(), nil, &resp); err != nil {
        return "", err
    }
    return resp.Content, nil
}

func (b *OpencodeBackend) SearchText(ctx context.Context, pattern string) ([]backend.SearchResult, error) {
    params := url.Values{}
    params.Set("pattern", pattern)

    var matches []struct {
        Path       string `json:"path"`
        Lines      string `json:"lines"`
        LineNumber int    `json:"line_number"`
    }
    if err := b.doRequest(ctx, "GET", "/find?"+params.Encode(), nil, &matches); err != nil {
        return nil, err
    }

    results := make([]backend.SearchResult, len(matches))
    for i, m := range matches {
        results[i] = backend.SearchResult{
            Path: m.Path, LineNumber: m.LineNumber, Content: m.Lines,
        }
    }
    return results, nil
}
```

- [ ] **Step 2: Write tests**

```go
// internal/backend/opencode/commands_test.go
package opencode

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func setupCommandsTestServer() (*httptest.Server, *OpencodeBackend) {
    mux := http.NewServeMux()
    server := httptest.NewServer(mux)

    mux.HandleFunc("/session/s1/command", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "parts": []map[string]interface{}{
                {"type": "text", "text": "command output"},
            },
        })
    })
    mux.HandleFunc("/file/content", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "type":    "raw",
            "content": "file content here",
        })
    })
    mux.HandleFunc("/find", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode([]map[string]interface{}{
            {"path": "main.go", "lines": "func main()", "line_number": 1},
        })
    })

    b := NewBackend(Config{AutoStart: false, APIURL: server.URL})
    return server, b
}

func TestExecuteCommand(t *testing.T) {
    srv, b := setupCommandsTestServer()
    defer srv.Close()

    result, err := b.ExecuteCommand(context.Background(), "s1", "/help")
    if err != nil {
        t.Fatalf("ExecuteCommand failed: %v", err)
    }
    if !strings.Contains(result.Stdout, "command output") {
        t.Errorf("expected command output in stdout")
    }
}

func TestReadFile(t *testing.T) {
    srv, b := setupCommandsTestServer()
    defer srv.Close()

    content, err := b.ReadFile(context.Background(), "main.go")
    if err != nil {
        t.Fatalf("ReadFile failed: %v", err)
    }
    if content != "file content here" {
        t.Errorf("content = %q, want %q", content, "file content here")
    }
}
```

- [ ] **Step 3: Run tests**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/opencode/ -v -run "TestExecuteCommand|TestReadFile"`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/backend/opencode/commands.go internal/backend/opencode/commands_test.go
git commit -m "feat(backend): implement OpencodeBackend commands and files API"
```

---

### Task 8: OpencodeBackend SSE events

**Files:**
- Create: `internal/backend/opencode/events.go`
- Create: `internal/backend/opencode/events_test.go`

- [ ] **Step 1: Implement SSE event handling**

```go
// internal/backend/opencode/events.go
package opencode

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "sync"

    "github.com/example/agent-tui/internal/backend"
)

type sseConnection struct {
    mu       sync.Mutex
    ctx      context.Context
    cancel   context.CancelFunc
    out      chan *backend.Event
    closed   bool
}

func (b *OpencodeBackend) Events(ctx context.Context) (<-chan *backend.Event, error) {
    b.mu.Lock()
    defer b.mu.Unlock()

    if b.sseConn != nil && !b.sseConn.closed {
        return b.sseConn.out, nil
    }

    ctx, cancel := context.WithCancel(ctx)
    b.sseConn = &sseConnection{
        ctx:    ctx,
        cancel: cancel,
        out:    make(chan *backend.Event, 100),
    }

    go b.sseLoop(ctx, b.sseConn.out)
    return b.sseConn.out, nil
}

func (b *OpencodeBackend) sseLoop(ctx context.Context, out chan<- *backend.Event) {
    defer close(out)

    req, err := http.NewRequestWithContext(ctx, "GET", b.baseURL+"/event", nil)
    if err != nil {
        return
    }

    resp, err := b.client.Do(req)
    if err != nil {
        return
    }
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    var eventType string
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "event: ") {
            eventType = strings.TrimPrefix(line, "event: ")
        } else if strings.HasPrefix(line, "data: ") {
            data := strings.TrimPrefix(line, "data: ")
            var payload interface{}
            if err := json.Unmarshal([]byte(data), &payload); err != nil {
                payload = data
            }
            select {
            case out <- &backend.Event{Type: eventType, Payload: payload}:
            case <-ctx.Done():
                return
            }
        }
    }
}

func (b *OpencodeBackend) CloseSSE() {
    b.mu.Lock()
    defer b.mu.Unlock()
    if b.sseConn != nil {
        b.sseConn.cancel()
        b.sseConn.closed = true
    }
}
```

- [ ] **Step 2: Write tests**

```go
// internal/backend/opencode/events_test.go
package opencode

import (
    "context"
    "fmt"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func TestEvents(t *testing.T) {
    mux := http.NewServeMux()
    server := httptest.NewServer(mux)

    mux.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.Header().Set("Connection", "keep-alive")
        flusher, ok := w.(http.Flusher)
        if !ok {
            return
        }
        fmt.Fprintf(w, "event: message.chunk\ndata: {\"content\":\"hello\"}\n\n")
        flusher.Flush()
        fmt.Fprintf(w, "event: message.completed\ndata: {\"content\":\"world\"}\n\n")
        flusher.Flush()
    })

    b := NewBackend(Config{AutoStart: false, APIURL: server.URL})
    defer server.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    events, err := b.Events(ctx)
    if err != nil {
        t.Fatalf("Events failed: %v", err)
    }

    var count int
    for evt := range events {
        count++
        if count >= 2 {
            break
        }
    }
    if count < 2 {
        t.Errorf("expected at least 2 events, got %d", count)
    }
}
```

- [ ] **Step 3: Run tests**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/backend/opencode/ -v -run "TestEvents"`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/backend/opencode/events.go internal/backend/opencode/events_test.go
git commit -m "feat(backend): implement OpencodeBackend SSE events"
```

---

### Task 9: Agent-TUI reverse-control HTTP server

**Files:**
- Create: `internal/server/server.go`
- Create: `internal/server/handlers.go`
- Create: `internal/server/server_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/server/server_test.go
package server

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestAppendPrompt(t *testing.T) {
    s := New(nil)
    ts := httptest.NewServer(s.mux())
    defer ts.Close()

    body := map[string]string{"text": "hello"}
    data, _ := json.Marshal(body)
    resp, err := http.Post(ts.URL+"/api/tui/append-prompt", "application/json", bytes.NewReader(data))
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    if resp.StatusCode != http.StatusOK {
        t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/server/ -v`
Expected: compilation error

- [ ] **Step 3: Implement server**

```go
// internal/server/server.go
package server

import (
    "context"
    "encoding/json"
    "net/http"
    "sync"
)

type TUIActionHandler interface {
    AppendPrompt(text string)
    SubmitPrompt()
    ShowToast(title, message, variant string)
    ExecuteCommand(command string)
}

type Server struct {
    handler TUIActionHandler
    srv     *http.Server
    mu      sync.Mutex
}

func New(handler TUIActionHandler) *Server {
    return &Server{handler: handler}
}

func (s *Server) mux() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("POST /api/tui/append-prompt", s.handleAppendPrompt)
    mux.HandleFunc("POST /api/tui/submit-prompt", s.handleSubmitPrompt)
    mux.HandleFunc("POST /api/tui/show-toast", s.handleShowToast)
    mux.HandleFunc("POST /api/tui/execute-command", s.handleExecuteCommand)
    return mux
}

func (s *Server) Start(ctx context.Context, addr string) error {
    s.mu.Lock()
    s.srv = &http.Server{
        Addr:    addr,
        Handler: s.mux(),
    }
    s.mu.Unlock()
    return s.srv.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.srv != nil {
        return s.srv.Shutdown(ctx)
    }
    return nil
}
```

```go
// internal/server/handlers.go
package server

import (
    "encoding/json"
    "net/http"
)

type appendPromptRequest struct {
    Text string `json:"text"`
}

type showToastRequest struct {
    Title   string `json:"title,omitempty"`
    Message string `json:"message"`
    Variant string `json:"variant,omitempty"`
}

type executeCommandRequest struct {
    Command string `json:"command"`
}

func (s *Server) handleAppendPrompt(w http.ResponseWriter, r *http.Request) {
    var req appendPromptRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if s.handler != nil {
        s.handler.AppendPrompt(req.Text)
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *Server) handleSubmitPrompt(w http.ResponseWriter, r *http.Request) {
    if s.handler != nil {
        s.handler.SubmitPrompt()
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *Server) handleShowToast(w http.ResponseWriter, r *http.Request) {
    var req showToastRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if s.handler != nil {
        s.handler.ShowToast(req.Title, req.Message, req.Variant)
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *Server) handleExecuteCommand(w http.ResponseWriter, r *http.Request) {
    var req executeCommandRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if s.handler != nil {
        s.handler.ExecuteCommand(req.Command)
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
```

- [ ] **Step 4: Fix tests to work without handler**

```go
// internal/server/server_test.go
package server

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestServer_AppendPrompt(t *testing.T) {
    s := New(nil)
    ts := httptest.NewServer(s.mux())
    defer ts.Close()

    body := map[string]string{"text": "hello"}
    data, _ := json.Marshal(body)
    resp, err := http.Post(ts.URL+"/api/tui/append-prompt", "application/json", bytes.NewReader(data))
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    if resp.StatusCode != http.StatusOK {
        t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
    }
}

func TestServer_SubmitPrompt(t *testing.T) {
    s := New(nil)
    ts := httptest.NewServer(s.mux())
    defer ts.Close()

    resp, err := http.Post(ts.URL+"/api/tui/submit-prompt", "application/json", nil)
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    if resp.StatusCode != http.StatusOK {
        t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
    }
}

func TestServer_ShowToast(t *testing.T) {
    s := New(nil)
    ts := httptest.NewServer(s.mux())
    defer ts.Close()

    body := map[string]string{"message": "hello", "variant": "success"}
    data, _ := json.Marshal(body)
    resp, err := http.Post(ts.URL+"/api/tui/show-toast", "application/json", bytes.NewReader(data))
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    if resp.StatusCode != http.StatusOK {
        t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
    }
}
```

- [ ] **Step 5: Run tests**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/server/ -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/server/
git commit -m "feat(server): add reverse-control HTTP server for TUI actions"
```

---

### Task 10: Config changes

**Files:**
- Find and modify: `internal/config/config.go`

- [ ] **Step 1: Locate existing config structure**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && grep -n "type Config struct" internal/config/config.go`
Expected: find the existing Config struct

- [ ] **Step 2: Read full config file**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && cat internal/config/config.go`

- [ ] **Step 3: Add backend config section**

Based on the existing config structure, add:

```go
// In config.go, add to Config struct:

type BackendConfig struct {
    Type      string `toml:"type"`
    Enabled   bool   `toml:"enabled"`
    AutoStart bool   `toml:"auto_start"`
    Binary    string `toml:"binary"`
    APIURL    string `toml:"api_url,omitempty"`
    APIKey    string `toml:"api_key,omitempty"`
}

type Config struct {
    // ... existing fields ...

    Agent struct {
        Backends map[string]BackendConfig `toml:"backends"`
    } `toml:"agent"`
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/config/config.go
git commit -m "feat(config): add agent backend configuration"
```

---

### Task 11: Integrate ExternalAgent into existing system

**Files:**
- Modify: `internal/service/agent_roles.go`
- Modify: `internal/service/task_orchestrator.go`

- [ ] **Step 1: Add ExternalAgent type**

```go
// In internal/service/agent_roles.go, add:

const TaskExternal TaskType = "external"

type ExternalAgent struct {
    backend backend.AgentBackend
}

func NewExternalAgent(b backend.AgentBackend) *ExternalAgent {
    return &ExternalAgent{backend: b}
}

func (e *ExternalAgent) Execute(task *Task) (string, error) {
    result, err := e.backend.SendMessage(context.Background(), task.ID, &backend.Message{
        Role:    "user",
        Content: task.Content,
    })
    if err != nil {
        return "", err
    }
    return result.Content, nil
}

func (e *ExternalAgent) GetRoleName() string {
    return "外部Agent"
}

func (e *ExternalAgent) GetSupportedTaskTypes() []TaskType {
    return []TaskType{TaskExternal}
}
```

- [ ] **Step 2: Register TaskExternal in orchestrator**

```go
// In internal/service/task_orchestrator.go, no change needed —
// TaskExternal is registered via the same RegisterAgent mechanism.
// The type is already handled by the generic TaskType map.
```

- [ ] **Step 3: Write test for ExternalAgent**

```go
// internal/service/agent_roles_test.go
package service

import (
    "testing"
    "context"

    "github.com/example/agent-tui/internal/backend"
)

type mockBackend struct {
    backend.AgentBackend
    reply string
}

func (m *mockBackend) SendMessage(ctx context.Context, sessionID string, msg *backend.Message) (*backend.MessageResult, error) {
    return &backend.MessageResult{Content: m.reply}, nil
}

func TestExternalAgent(t *testing.T) {
    b := &mockBackend{reply: "mock response"}
    agent := NewExternalAgent(b)

    task := &Task{
        ID: "test-1", Type: TaskExternal,
        Content: "hello",
    }
    result, err := agent.Execute(task)
    if err != nil {
        t.Fatalf("Execute failed: %v", err)
    }
    if result != "mock response" {
        t.Errorf("result = %q, want %q", result, "mock response")
    }
}

func TestExternalAgent_RoleName(t *testing.T) {
    agent := NewExternalAgent(&mockBackend{})
    if name := agent.GetRoleName(); name != "外部Agent" {
        t.Errorf("role name = %q, want %q", name, "外部Agent")
    }
}
```

- [ ] **Step 4: Run tests**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go test ./internal/service/ -v -run TestExternalAgent`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/service/agent_roles.go internal/service/agent_roles_test.go
git commit -m "feat(agent): add ExternalAgent role wrapping AgentBackend"
```

---

### Task 12: Wire initialization in App

**Files:**
- Modify: `app.go`

- [ ] **Step 1: Read current app.go initialization**

Run: `cat /Users/luxe/Downloads/agentx/agent5/app.go | head -100`

- [ ] **Step 2: Add backend initialization**

In the initialization function (likely `NewApp` or `Run`), add after existing service setup:

```go
// Initialize backend registry
backendRegistry := backend.NewRegistry()

// Start configured backends
for name, cfg := range config.Agent.Backends {
    if !cfg.Enabled {
        continue
    }
    b, err := backend.NewBackend(cfg.Type, cfg)
    if err != nil {
        log.Printf("Failed to create backend %s: %v", name, err)
        continue
    }
    if err := b.Start(context.Background()); err != nil {
        log.Printf("Failed to start backend %s: %v", name, err)
        continue
    }
    backendRegistry.Register(backend.AgentType(cfg.Type), b)

    // Register ExternalAgent for this backend
    externalAgent := service.NewExternalAgent(b)
    orchestrator.RegisterAgent(service.TaskExternal, externalAgent)
}

// Start reverse-control server
ctrlServer := server.New(app) // app implements TUIActionHandler
go ctrlServer.Start(context.Background(), ":0") // random port
```

- [ ] **Step 3: Ensure build compiles**

Run: `cd /Users/luxe/Downloads/agentx/agent5 && go build ./...`
Expected: success

- [ ] **Step 4: Commit**

```bash
git add app.go
git commit -m "feat(app): wire backend initialization and reverse-control server"
```

---

## Self-Review Checklist

- [ ] Every spec requirement maps to at least one task
- [ ] No TBDs, TODOs, or placeholder code
- [ ] Type/method names consistent across all tasks
- [ ] Each task produces independently testable code
- [ ] Commands with expected output specified
