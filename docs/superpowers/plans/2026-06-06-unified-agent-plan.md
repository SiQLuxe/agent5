# UnifiedAgent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the two parallel execution paths (AIAssistant + broken Runtime Agent) with a single UnifiedAgent that preserves conversation history, actually passes tool definitions to the LLM, and properly executes tool calls via ReAct loop.

**Architecture:** Extend `ai.Client` with tool calling support, create `SessionManager` wrapping `history.History` for context management, rewrite `runtime.Agent` to accept session context and use tools properly, update `Orchestrator`/`SkillExecutor`/`UI` to use the unified agent, remove `AIAssistant`.

**Tech Stack:** Go, OpenAI/DeepSeek API function calling

---

### Task 1: Extend ai.Client types with tool calling support

**Files:**
- Modify: `internal/ai/client.go`

**Overview:** Add tool definition and tool call types to `internal/ai/client.go`. The OpenAI-compatible `tools` and `tool_calls` fields will be added to request and response structs. These match the standard OpenAI function calling format which DeepSeek also uses.

- [ ] **Step 1: Add ToolDefinition and related types**

Add these types before `ChatCompletionRequest`:

```go
type ToolFunction struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Parameters  any    `json:"parameters"`
}

type ToolDefinition struct {
    Type     string       `json:"type"`
    Function ToolFunction `json:"function"`
}

type ToolCallFunction struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}

type ToolCall struct {
    ID       string           `json:"id"`
    Type     string           `json:"type"`
    Function ToolCallFunction `json:"function"`
}
```

- [ ] **Step 2: Extend ChatCompletionRequest with Tools field**

```go
type ChatCompletionRequest struct {
    Model    string           `json:"model"`
    Messages []Message        `json:"messages"`
    Tools    []ToolDefinition `json:"tools,omitempty"`
    Stream   bool             `json:"stream"`
}
```

- [ ] **Step 3: Extend ChatCompletionResponse with tool_calls support**

Replace the anonymous choice struct with named types to support tool_calls on the response message:

```go
type ResponseMessage struct {
    Role      string     `json:"role"`
    Content   string     `json:"content"`
    ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ResponseChoice struct {
    Index        int             `json:"index"`
    Message      ResponseMessage `json:"message"`
    FinishReason string          `json:"finish_reason"`
}

type ChatCompletionResponse struct {
    ID      string           `json:"id"`
    Object  string           `json:"object"`
    Created int64            `json:"created"`
    Model   string           `json:"model"`
    Choices []ResponseChoice `json:"choices"`
    Usage   struct {
        PromptTokens     int `json:"prompt_tokens"`
        CompletionTokens int `json:"completion_tokens"`
        TotalTokens      int `json:"total_tokens"`
    } `json:"usage"`
}
```

- [ ] **Step 4: Extend StreamingResponse with tool call delta support**

Add `ToolCalls` to the streaming delta so we can detect tool calls in streaming mode:

```go
type StreamingChoice struct {
    Index        int `json:"index"`
    Delta        struct {
        Role      string     `json:"role"`
        Content   string     `json:"content"`
        ToolCalls []ToolCall `json:"tool_calls,omitempty"`
    } `json:"delta"`
    FinishReason string `json:"finish_reason"`
}

type StreamingResponse struct {
    ID      string           `json:"id"`
    Object  string           `json:"object"`
    Created int64            `json:"created"`
    Model   string           `json:"model"`
    Choices []StreamingChoice `json:"choices"`
}
```

- [ ] **Step 5: Run build to verify no type breakage**

Run: `go build ./internal/ai/...`
Expected: PASS (ResponseMessage/ResponseChoice/StreamingChoice types are new, existing usage of anonymous structs is moved to named types)

- [ ] **Step 6: Commit**

```bash
git add internal/ai/client.go
git commit -m "feat(ai): add tool calling types to client"
```

---

### Task 2: Implement tool calling in OpenAIClient (and DeepSeek, Local)

**Files:**
- Modify: `internal/ai/openai.go`
- Modify: `internal/ai/deepseek.go`
- Modify: `internal/ai/local.go`
- Test: `internal/ai/openai_test.go` (create if not exists)

**Overview:** The OpenAI and DeepSeek APIs both use the same `tools`/`tool_calls` format. Modify `ChatCompletion` to serialize tools in the request and parse `tool_calls` from the response. For streaming, if the stream produces content deltas, return as text; if it produces tool call deltas, accumulate and return as a tool call. The Local client doesn't support tool calling — skip tools there.

- [ ] **Step 1: Write test for tool calling (non-streaming)**

```go
// internal/ai/openai_test.go
func TestOpenAIClientToolCallResponse(t *testing.T) {
    // Mock server that returns a response with tool_calls
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var req ChatCompletionRequest
        json.NewDecoder(r.Body).Decode(&req)
        if len(req.Tools) == 0 {
            t.Error("expected tools in request")
        }
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(ChatCompletionResponse{
            Choices: []ResponseChoice{
                {
                    Message: ResponseMessage{
                        Role:    "assistant",
                        Content: "",
                        ToolCalls: []ToolCall{
                            {
                                ID:   "call_123",
                                Type: "function",
                                Function: ToolCallFunction{
                                    Name:      "read_file",
                                    Arguments: `{"path":"/tmp/test.txt"}`,
                                },
                            },
                        },
                    },
                    FinishReason: "tool_calls",
                },
            },
        })
    }))
    defer server.Close()

    client := &OpenAIClient{
        apiKey:  "test",
        baseURL: server.URL,
        model:   "gpt-4",
        client:  &http.Client{},
    }

    resp, err := client.ChatCompletion(ChatCompletionRequest{
        Model: "gpt-4",
        Messages: []Message{{Role: "user", Content: "read a file"}},
        Tools: []ToolDefinition{{
            Type: "function",
            Function: ToolFunction{
                Name:        "read_file",
                Description: "read a file",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "path": map[string]interface{}{"type": "string"},
                    },
                    "required": []string{"path"},
                },
            },
        }},
    })
    if err != nil {
        t.Fatalf("ChatCompletion failed: %v", err)
    }
    if len(resp.Choices) == 0 {
        t.Fatal("expected at least one choice")
    }
    if len(resp.Choices[0].Message.ToolCalls) == 0 {
        t.Fatal("expected tool calls in response")
    }
    if resp.Choices[0].Message.ToolCalls[0].Function.Name != "read_file" {
        t.Fatalf("expected read_file tool call, got %s", resp.Choices[0].Message.ToolCalls[0].Function.Name)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ai/ -run TestOpenAIClientToolCallResponse -v`
Expected: FAIL (OpenAIClient.ChatCompletion doesn't parse tool_calls yet)

- [ ] **Step 3: Implement tool calling in OpenAIClient.ChatCompletion**

````go
// openai.go — replace the json.Marshal(req) block
type openAIRequest struct {
    Model    string           `json:"model"`
    Messages []Message        `json:"messages"`
    Tools    []ToolDefinition `json:"tools,omitempty"`
    Stream   bool             `json:"stream,omitempty"`
}

func (c *OpenAIClient) ChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
    if req.Model == "" {
        req.Model = c.model
    }
    oReq := openAIRequest{
        Model:    req.Model,
        Messages: req.Messages,
        Tools:    req.Tools,
        Stream:   req.Stream,
    }

    data, err := json.Marshal(oReq)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }
    // ... rest unchanged, but response now uses ResponseChoice/ResponseMessage types
```

No other changes needed — the JSON serialization field names match the OpenAI API format (`tools`, `tool_calls`), and the response struct `ChatCompletionResponse` already uses the new `ResponseChoice`/`ResponseMessage` types.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ai/ -run TestOpenAIClientToolCallResponse -v`
Expected: PASS

- [ ] **Step 5: Apply same change to DeepSeekClient**

````go
// deepseek.go — same change as OpenAIClient
// Use openAIRequest struct and pass Tools field
type deepSeekRequest struct {
    Model    string           `json:"model"`
    Messages []Message        `json:"messages"`
    Tools    []ToolDefinition `json:"tools,omitempty"`
    Stream   bool             `json:"stream,omitempty"`
}
````

DeepSeek uses the identical API format as OpenAI. Apply the same pattern: serialize Tools in request, parse ToolCalls from response.

- [ ] **Step 6: Apply the same request struct change to ChatCompletionStream in OpenAIClient**

In `ChatCompletionStream`, use the same `openAIRequest` struct so `Tools` are serialized when streaming. The streamed JSON lines from OpenAI will contain delta with `tool_calls` if the model decides to call tools. For now, we don't need to parse them in streaming — the streaming path will be used only for text responses, while tool calls go through the non-streaming `ChatCompletion` path (this is handled in the adapter, Task 5).

- [ ] **Step 7: Ensure LocalClient ignores tools (no change needed)**

`LocalClient` currently just passes through to a local LLM. No tools support is needed — it will send text-only requests.

- [ ] **Step 8: Commit**

```bash
git add internal/ai/openai.go internal/ai/deepseek.go
git commit -m "feat(ai): implement tool calling in OpenAI and DeepSeek clients"
```

---

### Task 3: *(skipped — Anthropic not supported)*

---

### Task 3: Create SessionManager

**Files:**
- Create: `internal/agent/session/manager.go`
- Test: `internal/agent/session/manager_test.go`

- [ ] **Step 1: Write failing test for session manager**

```go
// internal/agent/session/manager_test.go
package session

import (
    "testing"
)

func TestGetContextReturnsAllMessages(t *testing.T) {
    sm := NewManager()
    id := sm.CreateSession("test")
    sm.AddMessage(id, "user", "hello")
    sm.AddMessage(id, "assistant", "hi")
    sm.AddMessage(id, "user", "how are you?")

    ctx := sm.GetContext(id, 0) // 0 = no limit
    if len(ctx) != 3 {
        t.Fatalf("expected 3 messages, got %d", len(ctx))
    }
    if ctx[0].Role != "user" || ctx[0].Content != "hello" {
        t.Fatalf("unexpected first message: %+v", ctx[0])
    }
}

func TestGetContextTruncatesOldest(t *testing.T) {
    sm := NewManager()
    id := sm.CreateSession("test")
    for i := 0; i < 10; i++ {
        sm.AddMessage(id, "user", "msg")
        sm.AddMessage(id, "assistant", "resp")
    }

    ctx := sm.GetContext(id, 4) // last 4 messages (2 user + 2 assistant)
    if len(ctx) != 4 {
        t.Fatalf("expected 4 messages, got %d", len(ctx))
    }
}

func TestGetContextReturnsEmptyForUnknownSession(t *testing.T) {
    sm := NewManager()
    ctx := sm.GetContext("nonexistent", 0)
    if len(ctx) != 0 {
        t.Fatal("expected empty context for unknown session")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/agent/session/ -v`
Expected: FAIL (package doesn't exist yet)

- [ ] **Step 3: Implement SessionManager**

```go
// internal/agent/session/manager.go
package session

import (
    "sync"
    "time"

    "github.com/example/agent-tui/internal/data/history"
    "github.com/google/uuid"
)

type ContextMessage struct {
    Role    string
    Content string
}

type Manager struct {
    mu      sync.RWMutex
    history *history.History
    model   string
}

func NewManager() *Manager {
    return &Manager{
        history: history.NewHistory(""),
    }
}

func (m *Manager) CreateSession(name string) string {
    return m.history.CreateSession(name)
}

func (m *Manager) ListSessions() []history.SessionInfo {
    return m.history.GetSessions()
}

func (m *Manager) AddMessage(sessionID, role, content string) error {
    return m.history.AddMessage(sessionID, role, content)
}

func (m *Manager) GetContext(sessionID string, maxTurns int) []ContextMessage {
    msgs := m.history.GetMessages(sessionID)
    if msgs == nil {
        return nil
    }
    if maxTurns > 0 && len(msgs) > maxTurns*2 {
        msgs = msgs[len(msgs)-maxTurns*2:]
    }
    ctx := make([]ContextMessage, len(msgs))
    for i, msg := range msgs {
        ctx[i] = ContextMessage{Role: msg.Role, Content: msg.Content}
    }
    return ctx
}

func (m *Manager) SwitchModel(model string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.model = model
}

func (m *Manager) GetModel() string {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.model
}

func (m *Manager) ListModels() ([]string, error) {
    // delegate to model list — caller sets this after construction
    return []string{}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/agent/session/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/agent/session/
git commit -m "feat: create SessionManager for conversation history"
```

---

### Task 4: Fix aiLLMAdapter to pass tools and parse tool_calls

**Files:**
- Modify: `cmd/agent/main.go` (aiLLMAdapter)

**Overview:** The current `aiLLMAdapter.ChatWithTools` ignores the `tools` parameter and always returns `"final"`. Fix it to:
1. Convert `[]map[string]interface{}` tools to `[]ai.ToolDefinition` (they're already in the right format — just marshal each map)
2. Call `ai.Client.ChatCompletion` with tools in the request
3. Parse `tool_calls` from the response and return `runtime.LLMResponse` with `Type: "tool_call"` and a populated `ToolCall`

- [ ] **Step 1: Rewrite aiLLMAdapter.ChatWithTools**

```go
func (a *aiLLMAdapter) ChatWithTools(msgs []runtime.Message, tools []map[string]interface{}, model string) (*runtime.LLMResponse, error) {
    if model == "" {
        model = a.model
    }

    aiMessages := make([]ai.Message, len(msgs))
    for i, m := range msgs {
        aiMessages[i] = ai.Message{Role: m.Role, Content: m.Content}
    }

    // Convert tools from map format to typed ToolDefinition
    aiTools := make([]ai.ToolDefinition, len(tools))
    for i, t := range tools {
        data, _ := json.Marshal(t)
        json.Unmarshal(data, &aiTools[i])
    }

    req := ai.ChatCompletionRequest{
        Model:    model,
        Messages: aiMessages,
        Tools:    aiTools,
    }

    resp, err := a.client.ChatCompletion(req)
    if err != nil {
        return nil, err
    }
    if len(resp.Choices) == 0 {
        return &runtime.LLMResponse{Type: "final", Content: ""}, nil
    }

    choice := resp.Choices[0]
    msg := choice.Message

    // Check for tool calls
    if len(msg.ToolCalls) > 0 {
        tc := msg.ToolCalls[0]
        var args map[string]interface{}
        json.Unmarshal([]byte(tc.Function.Arguments), &args)

        return &runtime.LLMResponse{
            Type:    "tool_call",
            Content: msg.Content,
            ToolCall: &runtime.ToolCall{
                Name:      tc.Function.Name,
                Arguments: args,
            },
        }, nil
    }

    return &runtime.LLMResponse{Type: "final", Content: msg.Content}, nil
}
```

- [ ] **Step 2: Rewrite aiLLMAdapter.ChatWithToolsStream**

Since tool calls are hard to handle in streaming (deltas arrive piece by piece), use non-streaming for the initial call that might return tools. If the response has no tool calls and contains text, stream it. This is a pragmatic simplification: the streaming path is only for text-only responses.

```go
func (a *aiLLMAdapter) ChatWithToolsStream(msgs []runtime.Message, tools []map[string]interface{}, model string, onChunk func(string)) (*runtime.LLMResponse, error) {
    // Try non-streaming first to check for tool calls
    resp, err := a.ChatWithTools(msgs, tools, model)
    if err != nil {
        return nil, err
    }
    if resp.Type == "tool_call" {
        return resp, nil
    }
    // Text response — stream it
    if model == "" {
        model = a.model
    }
    aiMessages := make([]ai.Message, len(msgs))
    for i, m := range msgs {
        aiMessages[i] = ai.Message{Role: m.Role, Content: m.Content}
    }
    req := ai.ChatCompletionRequest{
        Model:    model,
        Messages: aiMessages,
        Stream:   true,
    }
    var fullContent string
    err = a.client.ChatCompletionStream(req, func(chunk string) {
        fullContent += chunk
        onChunk(chunk)
    })
    if err != nil {
        return nil, err
    }
    return &runtime.LLMResponse{Type: "final", Content: fullContent}, nil
}
```

- [ ] **Step 3: Add `encoding/json` to main.go imports**

Add `"encoding/json"` to the import block in `cmd/agent/main.go`.

- [ ] **Step 4: Run build**

Run: `go build ./cmd/agent/`
Expected: PASS

- [ ] **Step 5: Run existing agent tests**

Run: `go test ./internal/agent/runtime/ -v`
Expected: PASS (the MockLLMClient is used in tests, not the real adapter)

- [ ] **Step 6: Commit**

```bash
git add cmd/agent/main.go
git commit -m "feat: fix aiLLMAdapter to pass tools and parse tool_calls"
```

---

### Task 5: Rewrite runtime Agent with session context

**Files:**
- Modify: `internal/agent/runtime/agent.go`
- Modify: `internal/agent/runtime/react.go`
- Modify: `internal/agent/runtime/memory.go` (delete — replaced by SessionManager)
- Modify: `internal/agent/runtime/agent_test.go`
- Delete: `internal/agent/runtime/memory_test.go`

**Overview:** The core change. The Agent will:
1. Accept a `SessionManager` at construction time
2. `ExecuteStream` takes a `sessionID` parameter, loads history from SessionManager
3. Removes `Memory.Clear()` — history is now maintained by SessionManager
4. After execution, saves user + assistant messages to SessionManager
5. The `Config.ContextLimit` controls how many past turns are included

- [ ] **Step 1: Update agent.go — add SessionManager dependency, change signature**

```go
package runtime

import (
    "github.com/example/agent-tui/internal/agent/session"
    "github.com/example/agent-tui/internal/agent/tool"
)

type Config struct {
    Name         string
    Model        string
    SystemPrompt string
    MaxReActLoop int
    Temperature  float64
    ContextLimit int
    SandboxDir   string
    ApprovalFn   func(toolName string, params map[string]interface{}, oldContent, newContent string) bool
}

type Agent struct {
    Name      string
    Config    Config
    Tools     *tool.Registry
    Session   *session.Manager
    Logger    *Logger
    llm       LLMClient
}

func NewAgent(cfg Config, tools *tool.Registry, llm LLMClient, sm *session.Manager) *Agent {
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
        Name:    cfg.Name,
        Config:  cfg,
        Tools:   tools,
        Session: sm,
        Logger:  NewLogger(100),
        llm:     llm,
    }
}

func (a *Agent) Execute(sessionID, task string) (string, error) {
    a.Logger.Clear()
    return a.reactLoop(sessionID, task)
}

func (a *Agent) ExecuteStream(sessionID, task string, onChunk func(string)) (string, error) {
    a.Logger.Clear()
    return a.reactLoopStream(sessionID, task, onChunk)
}
```

- [ ] **Step 2: Rewrite react.go — load history, inject context, save messages**

```go
package runtime

import (
    "fmt"
    "time"

    "github.com/example/agent-tui/internal/agent/session"
    "github.com/example/agent-tui/internal/agent/tool"
)

func (a *Agent) buildMessages(sessionID, task string) []Message {
    msgs := []Message{
        {Role: "system", Content: a.Config.SystemPrompt},
    }
    // Load conversation history from SessionManager
    ctx := a.Session.GetContext(sessionID, a.Config.ContextLimit)
    for _, m := range ctx {
        msgs = append(msgs, Message{Role: m.Role, Content: m.Content})
    }
    // Append current user message
    msgs = append(msgs, Message{Role: "user", Content: task})
    return msgs
}

func (a *Agent) reactLoop(sessionID, task string) (string, error) {
    messages := a.buildMessages(sessionID, task)

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
            tc := tool.ToolContext{
                Context:    nil,
                SandboxDir: a.Config.SandboxDir,
                Approval:   a.Config.ApprovalFn,
            }
            result := t.Execute(tc, resp.ToolCall.Arguments)
            a.Logger.Log("tool_result", fmt.Sprintf("%+v", result.Data), resp.ToolCall.Name, resp.ToolCall.Arguments, &result, time.Since(toolStart))

            if result.Success {
                content := fmt.Sprintf("%v", result.Data)
                messages = append(messages, Message{Role: "tool", Content: content})
            } else {
                messages = append(messages, Message{Role: "tool", Content: fmt.Sprintf("error: %s", result.Error)})
            }

        case "final":
            a.Session.AddMessage(sessionID, "user", task)
            a.Session.AddMessage(sessionID, "assistant", resp.Content)
            return resp.Content, nil
        }
    }

    return "", fmt.Errorf("max react loop iterations (%d) reached", a.Config.MaxReActLoop)
}

func (a *Agent) reactLoopStream(sessionID, task string, onChunk func(string)) (string, error) {
    messages := a.buildMessages(sessionID, task)

    for i := 0; i < a.Config.MaxReActLoop; i++ {
        start := time.Now()
        resp, err := a.llm.ChatWithToolsStream(messages, a.Tools.AsToolDefinitions(), a.Config.Model, onChunk)
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
            tc := tool.ToolContext{
                Context:    nil,
                SandboxDir: a.Config.SandboxDir,
                Approval:   a.Config.ApprovalFn,
            }
            result := t.Execute(tc, resp.ToolCall.Arguments)
            a.Logger.Log("tool_result", fmt.Sprintf("%+v", result.Data), resp.ToolCall.Name, resp.ToolCall.Arguments, &result, time.Since(toolStart))

            if result.Success {
                content := fmt.Sprintf("%v", result.Data)
                messages = append(messages, Message{Role: "tool", Content: content})
            } else {
                messages = append(messages, Message{Role: "tool", Content: fmt.Sprintf("error: %s", result.Error)})
            }

        case "final":
            a.Session.AddMessage(sessionID, "user", task)
            a.Session.AddMessage(sessionID, "assistant", resp.Content)
            return resp.Content, nil
        }
    }

    return "", fmt.Errorf("max react loop iterations (%d) reached", a.Config.MaxReActLoop)
}
```

- [ ] **Step 3: Update agent_test.go**

```go
package runtime

import (
    "testing"

    "github.com/example/agent-tui/internal/agent/session"
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

    sm := session.NewManager()
    sid := sm.CreateSession("test")

    agent := NewAgent(Config{
        Name:         "test",
        Model:        "test-model",
        SystemPrompt: "You are a test agent",
    }, reg, mock, sm)

    result, err := agent.Execute(sid, "do something")
    if err != nil {
        t.Fatalf("Execute failed: %v", err)
    }
    if result != "Task complete" {
        t.Fatalf("expected 'Task complete', got %s", result)
    }

    // Verify messages were saved to session
    ctx := sm.GetContext(sid, 0)
    if len(ctx) != 2 {
        t.Fatalf("expected 2 messages in session, got %d", len(ctx))
    }
    if ctx[0].Role != "user" || ctx[0].Content != "do something" {
        t.Fatalf("unexpected first message: %+v", ctx[0])
    }
    if ctx[1].Role != "assistant" || ctx[1].Content != "Task complete" {
        t.Fatalf("unexpected second message: %+v", ctx[1])
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

    sm := session.NewManager()
    sid := sm.CreateSession("test")

    agent := NewAgent(Config{
        Name:         "test",
        Model:        "test-model",
        SystemPrompt: "You are a test agent",
    }, reg, mock, sm)

    result, err := agent.Execute(sid, "read a file")
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

- [ ] **Step 4: Move `Message` type out of memory.go, then delete memory.go**

The `Message` type (used by `react.go` and `llm.go`) lives in `memory.go`, which will be deleted. Move it to a new file first:

Create `internal/agent/runtime/types.go`:
```go
package runtime

type Message struct {
    Role    string
    Content string
}
```

Then delete `memory.go` and `memory_test.go` — the `Memory` struct is replaced by `SessionManager`.

- [ ] **Step 5: Run build and tests**

Run: `go build ./internal/agent/runtime/ && go test ./internal/agent/runtime/ -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/agent/runtime/agent.go internal/agent/runtime/react.go internal/agent/runtime/agent_test.go
git rm internal/agent/runtime/memory.go internal/agent/runtime/memory_test.go
git commit -m "feat: rewrite runtime agent with session context and tools"
```

---

### Task 6: Update Orchestrator to use session-aware Agent

**Files:**
- Modify: `internal/agent/orchestrator/orchestrator.go`
- Modify: `internal/agent/orchestrator/registry.go`
- Modify: `internal/agent/orchestrator/types.go`

**Overview:** The Orchestrator currently calls `agent.Execute(step.Content)` which doesn't take a sessionID. Update it to pass the session ID through the chain. The registry stores `*runtime.Agent` references which already have the SessionManager.

- [ ] **Step 1: Update Orchestrator.DispatchStream to accept sessionID**

```go
func (o *Orchestrator) DispatchStream(sessionID string, task *Task, onChunk func(string)) ([]*Task, error) {
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
            return results, fmt.Errorf("%s", step.Error)
        }

        agent := agents[0]
        step.Status = StatusRunning
        step.AgentID = agent.Name

        result, err := agent.ExecuteStream(sessionID, step.Content, onChunk)
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
```

Similarly update `Dispatch` and `DispatchAndMerge` to accept `sessionID`.

- [ ] **Step 2: Run build**

Run: `go build ./internal/agent/orchestrator/`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/agent/orchestrator/
git commit -m "feat: update orchestrator to pass sessionID to agent"
```

---

### Task 7: Update SkillExecutor to use UnifiedAgent

**Files:**
- Modify: `internal/service/skill_executor.go`

**Overview:** Replace `AIAssistant` dependency with a minimal LLM caller interface that can be satisfied by the runtime Agent + SessionManager.

- [ ] **Step 1: Replace AIAssistant with SessionManager-based chat**

```go
package service

import (
    "fmt"
    "strings"
    "github.com/example/agent-tui/internal/agent/session"
    "github.com/example/agent-tui/internal/ai"
)

// LLMChatter provides a simple chat completion (no tools needed for skills)
type LLMChatter interface {
    Chat(sessionID, message string) (string, error)
}

type SkillExecutor struct {
    registry    *SkillRegistry
    chater      LLMChatter
}

func NewSkillExecutor(registry *SkillRegistry, chater LLMChatter) *SkillExecutor {
    return &SkillExecutor{registry: registry, chater: chater}
}

func (e *SkillExecutor) Execute(cmd *ParsedCommand) (string, error) {
    skill, ok := e.registry.Get(cmd.Name)
    if !ok {
        return "", fmt.Errorf("skill not found: %s", cmd.Name)
    }

    switch skill.Type {
    case SkillHandler:
        fn, ok := e.registry.GetHandler(skill.Name)
        if !ok {
            return "", fmt.Errorf("no handler registered for: %s", skill.Name)
        }
        return fn(SkillContext{Input: cmd.Args}), nil

    case SkillPrompt:
        prompt := skill.Prompt
        if cmd.Args != "" && strings.Contains(prompt, "%s") {
            prompt = strings.ReplaceAll(prompt, "%s", cmd.Args)
        }
        if e.chater != nil {
            return e.chater.Chat("skill-"+skill.Name, prompt)
        }
        return prompt, nil
    }

    return "", fmt.Errorf("unknown skill type")
}
```

- [ ] **Step 2: Run build**

Run: `go build ./internal/service/`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/service/skill_executor.go
git commit -m "feat: update SkillExecutor to use LLMChatter instead of AIAssistant"
```

---

### Task 9: Wire everything in main.go + update UI + delete AIAssistant

**Files:**
- Modify: `cmd/agent/main.go`
- Modify: `internal/ui/app.go`
- Delete: `internal/service/ai_assistant.go`
- Modify: `internal/ui/theme_service.go`

**Overview:** The final integration step. Wire the `SessionManager` + fixed `aiLLMAdapter` + `runtime.Agent` together. Update the UI to route all messages through the agent. Delete `AIAssistant`.

- [ ] **Step 1: Rewrite main.go wiring**

```go
// cmd/agent/main.go — key sections

// Create LLM client (unchanged)
aiClient, err := ai.NewClientFromConfig(cfg)

// Create shared SessionManager (replaces history + AIAssistant)
sm := session.NewManager()
sm.SwitchModel(cfg.DefaultClient)

// Create agent LLM adapter (now with working tools)
agentLLM := &aiLLMAdapter{client: aiClient, model: cfg.DefaultClient}

// Create tool registry (unchanged)
toolReg := tool.NewRegistry()
toolReg.Register(&tool.ReadFileTool{})
toolReg.Register(&tool.WriteFileTool{})
toolReg.Register(&tool.SearchTextTool{})
toolReg.Register(&tool.ExecCommandTool{})
if aiClient != nil {
    toolReg.Register(&tool.ChatLLMTool{Provider: &aiLLMProvider{client: aiClient}})
}

// Create agents with session manager
agentReg := orchestrator.NewRegistry()
for _, ac := range cfg.AgentRoles {
    if !ac.Enabled {
        continue
    }
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
        SandboxDir:   ac.SandboxDir,
        ApprovalFn:   approvalFn,
    }, agentTools, agentLLM, sm)
    agentReg.Register(ac.Name, agent, ...)
}

// Create skill executor with SessionManager-based chatter
skillExecutor := service.NewSkillExecutor(skillRegistry, sm)

// Create orchestrator
orch := orchestrator.NewOrchestrator(agentReg, ...)

// Wire app
app := ui.NewApp()
app.SetSessionManager(sm)
app.SetSkillExecutor(skillExecutor)
app.SetSkillRegistry(skillRegistry)
app.SetOrchestrator(orch)
```

- [ ] **Step 2: Update app.go — replace AIAssistant with SessionManager**

Add to App struct:
```go
sessionMgr *session.Manager
```

Add setter:
```go
func (a *App) SetSessionManager(sm *session.Manager) {
    a.sessionMgr = sm
}
```

Update `newSession()`:
```go
func (a *App) newSession() {
    id := uuid.New().String()
    if a.sessionMgr != nil {
        id = a.sessionMgr.CreateSession("New Session")
    }
    s := NewSession(id, "New Session")
    a.sessions = append(a.sessions, s)
    ...
}
```

Update `sendMessage()` – remove AIAssistant path, use only agent/orchestrator:
```go
func (a *App) sendMessage() {
    text := a.composer.GetInput()
    if strings.TrimSpace(text) == "" {
        return
    }
    a.inputHistory = append(a.inputHistory, text)
    a.historyIndex = 0
    a.hideSuggestions()

    s := a.activeSessionPtr()
    if s == nil {
        return
    }

    if strings.HasPrefix(text, "/") {
        pc := service.ParseCommand(text)
        if pc != nil && a.skillExecutor != nil {
            result, err := a.skillExecutor.Execute(pc)
            if err != nil {
                s.AddMessage(RoleSkill, "Error: "+err.Error())
            } else {
                s.AddMessage(RoleSkill, result)
            }
            s.Messages[len(s.Messages)-1].Label = pc.Name
        }
        a.composer.ClearInput()
        return
    }

    if a.orch == nil && a.sessionMgr == nil {
        return
    }
    s.AddMessage(RoleUser, text)
    s.AddMessage(RoleAssistant, "")
    a.composer.ClearInput()
    a.chatPanel.SetSession(s)
    a.chatPanel.StartStreaming()
    a.isLoading = true

    sessionPtr := s
    go func() {
        if a.orch != nil {
            // Orchestrator path — pass sessionID
            task := &orchestrator.Task{
                ID:      uuid.New().String(),
                Type:    orchestrator.TaskExecute,
                Content: text,
            }
            var streamBuf string
            results, err := a.orch.DispatchStream(sessionPtr.ID, task, func(chunk string) { ... })
            // ... rest unchanged
            return
        }

        // Direct agent path (no orchestrator)
        var streamBuf string
        err := a.sessionMgr.ChatStream(sessionPtr.ID, text, func(chunk string) {
            streamBuf += chunk
            a.QueueUpdateDraw(func() {
                if len(sessionPtr.Messages) > 0 {
                    sessionPtr.Messages[len(sessionPtr.Messages)-1].Content = streamBuf
                    a.chatPanel.UpdateStreaming(streamBuf)
                }
            })
        })
        // ... handle result same as before
    }()
}
```

- [ ] **Step 3: Add ChatStream method to SessionManager**

```go
// internal/agent/session/manager.go
func (m *Manager) ChatStream(sessionID, message string, callback func(string)) error {
    // This is a simple chat without tools — used as fallback when no orchestrator
    // Uses the model directly without agent tools
    m.AddMessage(sessionID, "user", message)
    ctx := m.GetContext(sessionID, 0)

    aiMessages := make([]ai.Message, len(ctx))
    for i, msg := range ctx {
        aiMessages[i] = ai.Message{Role: msg.Role, Content: msg.Content}
    }

    req := ai.ChatCompletionRequest{
        Model:    m.GetModel(),
        Messages: aiMessages,
        Stream:   true,
    }

    var fullResponse string
    err := m.client.ChatCompletionStream(req, func(content string) {
        fullResponse += content
        callback(content)
    })
    if err == nil && fullResponse != "" {
        m.AddMessage(sessionID, "assistant", fullResponse)
    }
    return err
}
```

Wait — `SessionManager` doesn't have an `ai.Client` reference. I need to add one.

Actually, let me reconsider. The session manager should just manage sessions. The actual chatting should be done by the agent. If there's no orchestrator configured, the UI should create a default single-agent and use it directly.

Simpler approach: always use the agent (which has both SessionManager and LLM). No fallback ChatStream needed.

```go
// app.go — simplified sendMessage
go func() {
    // Always use agent path (we always have at least one agent)
    var streamBuf string
    _, err := a.agent.ExecuteStream(sessionPtr.ID, text, func(chunk string) {
        streamBuf += chunk
        a.QueueUpdateDraw(func() {
            if len(sessionPtr.Messages) > 0 {
                sessionPtr.Messages[len(sessionPtr.Messages)-1].Content = streamBuf
                a.chatPanel.UpdateStreaming(streamBuf)
            }
        })
    })
    // handle err...
}()
```

But we'd need a reference to the agent in the App. Let me add a `primaryAgent *runtime.Agent` field.

Actually, I think the cleaner approach for the plan is:
- The App gets a `primaryAgent *runtime.Agent` reference
- When orchestrator is configured, use orchestrator (which internally uses agents)
- When orchestrator is not configured, use primaryAgent directly
- Remove AIAssistant entirely

Let me write Task 9 more carefully.

Actually, I realize I'm getting into too much detail in the plan. Let me simplify — for Task 9, I'll describe what needs to happen at a high level with code snippets for the critical parts, like the design doc describes. The actual implementation will figure out the details.

But the writing-plans skill says "Every step must contain the actual content an engineer needs." So I need concrete code.

Let me think about this more practically. The key changes in Task 9:

1. `main.go`:
   - Create `session.NewManager()` instead of `history.NewHistory("")`
   - Pass `sm` to `runtime.NewAgent(..., sm)` 
   - Create `service.NewSkillExecutor(skillRegistry, sm)` instead of with `aiAssistant`
   - Remove `aiAssistant := service.NewAIAssistant(...)`
   - Wire `app.SetSessionManager(sm)`

2. `app.go`:
   - Add `sessionMgr *session.Manager` field, `SetSessionManager` setter
   - `newSession()`: use `a.sessionMgr.CreateSession()` instead of `a.aiAssistant.CreateSession()`
   - `sendMessage()`: simplify to always go through orchestrator or agent, remove AIAssistant branch
   - Remove `aiAssistant` field and `SetAIAssistant` method

3. Add `ai.Client` reference to `session.Manager` so it can do simple chat for skills.

Wait, actually the `SkillExecutor` now takes an `LLMChatter` interface. For skill prompts, it calls `Chat(sessionID, prompt)`. The `SessionManager` needs to satisfy this. So `SessionManager` needs an `ai.Client` reference.

Let me adjust the SessionManager to hold an ai.Client:

```go
type Manager struct {
    mu      sync.RWMutex
    history *history.History
    client  ai.Client
    model   string
}

func (m *Manager) Chat(sessionID, message string) (string, error) {
    m.AddMessage(sessionID, "user", message)
    ctx := m.GetContext(sessionID, 0)
    aiMessages := make([]ai.Message, len(ctx))
    for i, msg := range ctx {
        aiMessages[i] = ai.Message{Role: msg.Role, Content: msg.Content}
    }
    req := ai.ChatCompletionRequest{
        Model:    m.GetModel(),
        Messages: aiMessages,
    }
    resp, err := m.client.ChatCompletion(req)
    if err != nil {
        return "", err
    }
    if len(resp.Choices) > 0 {
        content := resp.Choices[0].Message.Content
        m.AddMessage(sessionID, "assistant", content)
        return content, nil
    }
    return "", nil
}
```

OK, this is getting complex. Let me just write the plan now — it's detailed enough for an engineer to execute.<｜end▁of▁thinking｜>

<｜｜DSML｜｜tool_calls>
<｜｜DSML｜｜invoke name="write">
<｜｜DSML｜｜parameter name="content" string="true">## UnifiedAgent Implementation Plan (continued)

---

### Task 8: Wire SessionManager + Agent in main.go, update UI, delete AIAssistant

**Files:**
- Modify: `cmd/agent/main.go`
- Modify: `internal/ui/app.go`
- Modify: `internal/ui/theme_service.go`
- Modify: `internal/ui/app_test.go`
- Modify: `internal/agent/session/manager.go` (add Chat method + ai.Client field)
- Delete: `internal/service/ai_assistant.go`

**Overview:** The final integration. Wire the new `SessionManager` everywhere, update the UI to route all messages through the agent, and delete `AIAssistant`. The SessionManager gets a `Chat` method (simple LLM call without tools) to satisfy the `LLMChatter` interface used by `SkillExecutor`.

- [ ] **Step 1: Add ai.Client and Chat method to SessionManager**

```go
// internal/agent/session/manager.go — add fields and method

type Manager struct {
    mu      sync.RWMutex
    history *history.History
    client  ai.Client
    model   string
}

func (m *Manager) SetClient(client ai.Client) {
    m.client = client
}

// Chat satisfies service.LLMChatter for SkillExecutor
func (m *Manager) Chat(sessionID, message string) (string, error) {
    m.AddMessage(sessionID, "user", message)
    ctx := m.GetContext(sessionID, 0)
    aiMessages := make([]ai.Message, len(ctx))
    for i, msg := range ctx {
        aiMessages[i] = ai.Message{Role: msg.Role, Content: msg.Content}
    }
    req := ai.ChatCompletionRequest{
        Model:    m.GetModel(),
        Messages: aiMessages,
    }
    resp, err := m.client.ChatCompletion(req)
    if err != nil {
        return "", err
    }
    if len(resp.Choices) > 0 {
        content := resp.Choices[0].Message.Content
        m.AddMessage(sessionID, "assistant", content)
        return content, nil
    }
    return "", nil
}
```

Also add `ChatStream` for direct chat (fallback when no orchestrator):

```go
func (m *Manager) ChatStream(sessionID, message string, callback func(string)) error {
    m.AddMessage(sessionID, "user", message)
    ctx := m.GetContext(sessionID, 0)
    aiMessages := make([]ai.Message, len(ctx))
    for i, msg := range ctx {
        aiMessages[i] = ai.Message{Role: msg.Role, Content: msg.Content}
    }
    req := ai.ChatCompletionRequest{
        Model:  m.GetModel(),
        Messages: aiMessages,
        Stream: true,
    }
    var fullResponse string
    err := m.client.ChatCompletionStream(req, func(content string) {
        fullResponse += content
        callback(content)
    })
    if err == nil && fullResponse != "" {
        m.AddMessage(sessionID, "assistant", fullResponse)
    }
    return err
}
```

Add import for `"github.com/example/agent-tui/internal/ai"`.

- [ ] **Step 2: Rewrite main.go — wire new components, remove AIAssistant**

```go
// cmd/agent/main.go — key wiring changes

aiClient, err := ai.NewClientFromConfig(cfg)

// Create shared SessionManager (replaces history.History + AIAssistant)
sm := session.NewManager()
sm.SetClient(aiClient)
sm.SwitchModel(cfg.DefaultClient)

// Create agent LLM adapter (now with working tools)
agentLLM := &aiLLMAdapter{client: aiClient, model: cfg.DefaultClient}

// Tool registry (unchanged)
toolReg := tool.NewRegistry()
toolReg.Register(&tool.ReadFileTool{})
toolReg.Register(&tool.WriteFileTool{})
toolReg.Register(&tool.SearchTextTool{})
toolReg.Register(&tool.ExecCommandTool{})
if aiClient != nil {
    toolReg.Register(&tool.ChatLLMTool{Provider: &aiLLMProvider{client: aiClient}})
}

// Skill registry (unchanged)
skillRegistry := service.NewSkillRegistry()
service.LoadSkillsDir(skillRegistry, "skills")

// SkillExecutor uses SessionManager (satisfies LLMChatter)
skillExecutor := service.NewSkillExecutor(skillRegistry, sm)

// Create agents with session manager
agentReg := orchestrator.NewRegistry()
for _, ac := range cfg.AgentRoles {
    if !ac.Enabled { continue }
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
        SandboxDir:   ac.SandboxDir,
        ApprovalFn:   approvalFn,
    }, agentTools, agentLLM, sm)
    agentReg.Register(ac.Name, agent,
        string(orchestrator.TaskExecute),
        string(orchestrator.TaskAnalyze),
        string(orchestrator.TaskDesign),
        string(orchestrator.TaskCode),
        string(orchestrator.TaskReview),
    )
}
orch := orchestrator.NewOrchestrator(agentReg, orchestrator.NewDecomposer(), orchestrator.NewMerger())

// Wire app — no AIAssistant
app := ui.NewApp()
app.SetSessionManager(sm)
app.SetPrimaryAgent(nil)       // set to first agent if no orchestrator, or nil
app.SetSkillExecutor(skillExecutor)
app.SetSkillRegistry(skillRegistry)
app.SetSkillsDir("skills")
app.SetOrchestrator(orch)      // when orch is set, it's used; when nil, primaryAgent is used
```

Delete the lines:
- `h := history.NewHistory("")`
- `aiAssistant := service.NewAIAssistant(aiClient, h)`
- `app.SetAIAssistant(aiAssistant)`

- [ ] **Step 3: Update app.go**

Add fields and setters:
```go
// App struct — replace aiAssistant with sessionMgr
sessionMgr  *session.Manager
primaryAgent *runtime.Agent  // used when orch is nil

func (a *App) SetSessionManager(sm *session.Manager) {
    a.sessionMgr = sm
}

func (a *App) SetPrimaryAgent(agent *runtime.Agent) {
    a.primaryAgent = agent
}
```

Update `newSession()`:
```go
func (a *App) newSession() {
    id := uuid.New().String()
    if a.sessionMgr != nil {
        id = a.sessionMgr.CreateSession("New Session")
    }
    s := NewSession(id, "New Session")
    a.sessions = append(a.sessions, s)
    a.tabDock.AddTab(tabbar.Tab{ID: s.ID, Label: "New Session"})
    a.switchToSession(len(a.sessions) - 1)
}
```

Update `sendMessage()` — simplify to single path:
```go
go func() {
    if a.orch != nil {
        task := &orchestrator.Task{
            ID:      uuid.New().String(),
            Type:    orchestrator.TaskExecute,
            Content: text,
        }
        var streamBuf string
        results, err := a.orch.DispatchStream(sessionPtr.ID, task, func(chunk string) {
            a.QueueUpdateDraw(func() {
                streamBuf += chunk
                if len(sessionPtr.Messages) > 0 {
                    sessionPtr.Messages[len(sessionPtr.Messages)-1].Content = streamBuf
                    a.chatPanel.UpdateStreaming(streamBuf)
                }
            })
        })
        a.QueueUpdateDraw(func() {
            a.isLoading = false
            if len(sessionPtr.Messages) > 0 && sessionPtr.Messages[len(sessionPtr.Messages)-1].Role == RoleAssistant {
                sessionPtr.Messages = sessionPtr.Messages[:len(sessionPtr.Messages)-1]
            }
            if err != nil {
                sessionPtr.AddMessage(RoleSystem, "Agent error: "+err.Error())
            }
            for _, r := range results {
                statusStr := string(r.Status)
                header := "[" + statusStr + "] " + string(r.Type) + " (agent: " + r.AgentID + ")"
                if r.Error != "" {
                    sessionPtr.AddMessage(RoleSkill, header+"\n"+r.Error)
                } else if r.Result != "" {
                    sessionPtr.AddMessage(RoleSkill, header+"\n"+r.Result)
                }
            }
            a.chatPanel.EndStreaming()
        })
        return
    }

    // Direct agent or session manager fallback
    if a.primaryAgent != nil {
        var streamBuf string
        _, err := a.primaryAgent.ExecuteStream(sessionPtr.ID, text, func(chunk string) {
            streamBuf += chunk
            a.QueueUpdateDraw(func() {
                if len(sessionPtr.Messages) > 0 {
                    sessionPtr.Messages[len(sessionPtr.Messages)-1].Content = streamBuf
                    a.chatPanel.UpdateStreaming(streamBuf)
                }
            })
        })
        a.QueueUpdateDraw(func() {
            a.isLoading = false
            if err != nil {
                sessionPtr.AddMessage(RoleSystem, "Error: "+err.Error())
            }
            a.chatPanel.EndStreaming()
        })
    } else if a.sessionMgr != nil {
        // Simple chat without tools (fallback)
        var fullResponse string
        err := a.sessionMgr.ChatStream(sessionPtr.ID, text, func(chunk string) {
            fullResponse += chunk
            a.QueueUpdateDraw(func() {
                if len(sessionPtr.Messages) > 0 {
                    sessionPtr.Messages[len(sessionPtr.Messages)-1].Content = fullResponse
                    a.chatPanel.UpdateStreaming(fullResponse)
                }
            })
        })
        a.QueueUpdateDraw(func() {
            a.isLoading = false
            if err != nil {
                sessionPtr.AddMessage(RoleSystem, "Error: "+err.Error())
            }
            a.chatPanel.EndStreaming()
            label := sessionPtr.GenerateLabel()
            if label != "New Session" {
                sessionPtr.Label = label
                for i, s := range a.sessions {
                    if s == sessionPtr {
                        a.tabDock.UpdateTab(i, label)
                        break
                    }
                }
            }
        })
    }
}()
```

Remove the `aiAssistant` field and `SetAIAssistant` method from App.

- [ ] **Step 4: Update theme_service.go to use SessionManager**

```go
// internal/ui/theme_service.go
type ThemeService struct {
    sessionMgr *session.Manager
}

func NewThemeService(sm *session.Manager) *ThemeService {
    return &ThemeService{sessionMgr: sm}
}

// Replace ts.aiAssistant.Chat(...) with ts.sessionMgr.Chat(...)
```

- [ ] **Step 5: Delete ai_assistant.go**

Run: `rm internal/service/ai_assistant.go`

- [ ] **Step 6: Update app_test.go**

Update references from `AIAssistant` to `SessionManager`:

```go
// Replace:
//   aiAssistant := service.NewAIAssistant(mockClient, h)
//   a.SetAIAssistant(aiAssistant)
// With:
//   sm := session.NewManager()
//   sm.SetClient(mockClient)
//   a.SetSessionManager(sm)
```

- [ ] **Step 7: Build and test**

Run: `go build ./cmd/agent/ && go test ./...`
Expected: All PASS

- [ ] **Step 8: Commit**

```bash
git add cmd/agent/main.go internal/ui/app.go internal/ui/theme_service.go internal/ui/app_test.go internal/agent/session/manager.go
git rm internal/service/ai_assistant.go
git commit -m "feat: wire UnifiedAgent, delete AIAssistant, update UI"
```

---

### Task 9: Cleanup — remove unused Memory references

**Files:**
- Check: `internal/agent/runtime/`
- Check: `internal/agent/orchestrator/`

**Overview:** Ensure no remaining references to the deleted `Memory` type or `Clear()` calls. Run a full build and test suite.

- [ ] **Step 1: Search for remaining Memory references**

Run: `grep -r "Memory" internal/ --include="*.go"`
Expected: No references to `runtime.Memory` or `a.Memory` remain.

- [ ] **Step 2: Search for Clear() calls on wrong type**

Run: `grep -r "\.Clear()" internal/ --include="*.go"`
Expected: Only `a.Logger.Clear()` in agent.go.

- [ ] **Step 3: Full build**

Run: `go build ./... && go vet ./...`
Expected: PASS

- [ ] **Step 4: Full test suite**

Run: `go test ./...`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "chore: cleanup unused Memory references after UnifiedAgent migration"
```

---

## File Change Summary

| Action | File |
|--------|------|
| Modify | `internal/ai/client.go` — add tool types, extend request/response |
| Modify | `internal/ai/openai.go` — serialize tools, parse tool_calls |
| Modify | `internal/ai/deepseek.go` — same as openai |
| Create | `internal/agent/session/manager.go` — SessionManager |
| Modify | `internal/agent/runtime/agent.go` — accept sessionID, no Memory.Clear |
| Modify | `internal/agent/runtime/react.go` — load history from SessionManager |
| Delete | `internal/agent/runtime/memory.go` — replaced by SessionManager |
| Modify | `cmd/agent/main.go` — fix adapter, wire new components |
| Modify | `internal/agent/orchestrator/orchestrator.go` — pass sessionID |
| Modify | `internal/service/skill_executor.go` — use LLMChatter interface |
| Delete | `internal/service/ai_assistant.go` — replaced entirely |
| Modify | `internal/ui/app.go` — remove AIAssistant, use sessionMgr |
| Modify | `internal/ui/theme_service.go` — use sessionMgr |
| Modify | `internal/ui/app_test.go` — update test wiring |
