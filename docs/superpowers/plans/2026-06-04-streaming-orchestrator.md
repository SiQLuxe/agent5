# Streaming Orchestrator Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Pipe LLM token streaming through the Orchestrator → Agent → LLMClient layers so the UI shows incremental output instead of waiting for the full response.

**Architecture:** Add a streaming variant at each layer — `ChatWithToolsStream` on `LLMClient`, `ExecuteStream` on `Agent`, `DispatchStream` on `Orchestrator` — all accepting an `onChunk func(string)` callback that forwards token deltas from the `ai.Client.ChatCompletionStream` SSE parser up to the `ChatPanel.UpdateStreaming()`.

**Tech Stack:** Go, tview

---

### Task 1: Add ChatWithToolsStream to LLMClient interface + Mock

**Files:**
- Modify: `internal/agent/runtime/llm.go:14-16` (interface), `llm.go:18-34` (mock)

- [ ] **Step 1: Add ChatWithToolsStream to interface**

```go
type LLMClient interface {
	ChatWithTools(messages []Message, tools []map[string]interface{}, model string) (*LLMResponse, error)
	ChatWithToolsStream(messages []Message, tools []map[string]interface{}, model string, onChunk func(string)) (*LLMResponse, error)
}
```

- [ ] **Step 2: Add ChatWithToolsStream to MockLLMClient**

```go
func (m *MockLLMClient) ChatWithToolsStream(messages []Message, tools []map[string]interface{}, model string, onChunk func(string)) (*LLMResponse, error) {
	if m.index >= len(m.Responses) {
		return &LLMResponse{Type: "final", Content: "done"}, nil
	}
	resp := m.Responses[m.index]
	m.index++
	// Simulate streaming by calling onChunk once for the full content
	if onChunk != nil && resp.Content != "" {
		onChunk(resp.Content)
	}
	return &resp, nil
}
```

- [ ] **Step 3: Build and test**

Run: `go build ./internal/agent/runtime/...`
Expected: no errors

Run: `go test ./internal/agent/runtime/... -count=1 -short`
Expected: all tests pass (existing tests still use `ChatWithTools`; mock now also implements the new method)

- [ ] **Step 4: Commit**

```bash
git add internal/agent/runtime/llm.go
git commit -m "feat: add ChatWithToolsStream to LLMClient interface"
```

---

### Task 2: Add ExecuteStream + reactLoopStream to Agent

**Files:**
- Modify: `internal/agent/runtime/agent.go:46-49`
- Modify: `internal/agent/runtime/react.go`

- [ ] **Step 1: Add ExecuteStream method to Agent**

Append after `Execute` in `agent.go`:

```go
func (a *Agent) ExecuteStream(task string, onChunk func(string)) (string, error) {
	a.Memory.Clear()
	a.Logger.Clear()
	return a.reactLoopStream(task, onChunk)
}
```

- [ ] **Step 2: Add reactLoopStream method**

Append to `react.go` (full method, do NOT modify existing `reactLoop`):

```go
func (a *Agent) reactLoopStream(task string, onChunk func(string)) (string, error) {
	messages := []Message{
		{Role: "system", Content: a.Config.SystemPrompt},
		{Role: "user", Content: task},
	}

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
			a.Memory.Append("assistant", resp.Content)
			return resp.Content, nil
		}
	}

	return "", fmt.Errorf("max react loop iterations (%d) reached", a.Config.MaxReActLoop)
}
```

- [ ] **Step 3: Build and test**

Run: `go build ./internal/agent/runtime/...`
Expected: no errors

Run: `go test ./internal/agent/runtime/... -count=1 -short`
Expected: all pass

- [ ] **Step 4: Commit**

```bash
git add internal/agent/runtime/agent.go internal/agent/runtime/react.go
git commit -m "feat: add ExecuteStream and reactLoopStream to Agent"
```

---

### Task 3: Add DispatchStream to Orchestrator

**Files:**
- Modify: `internal/agent/orchestrator/orchestrator.go`

- [ ] **Step 1: Add DispatchStream method**

Append after `DispatchAndMerge` in `orchestrator.go`:

```go
func (o *Orchestrator) DispatchStream(task *Task, onChunk func(string)) ([]*Task, error) {
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

		result, err := agent.ExecuteStream(step.Content, onChunk)
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

- [ ] **Step 2: Build and test**

Run: `go build ./internal/agent/orchestrator/...`
Expected: no errors

Run: `go test ./internal/agent/orchestrator/... -count=1 -short`
Expected: all pass

- [ ] **Step 3: Commit**

```bash
git add internal/agent/orchestrator/orchestrator.go
git commit -m "feat: add DispatchStream to Orchestrator"
```

---

### Task 4: Add ChatWithToolsStream to aiLLMAdapter

**Files:**
- Modify: `cmd/agent/main.go`

- [ ] **Step 1: Add ChatWithToolsStream to aiLLMAdapter**

Append after `ChatWithTools` method in `cmd/agent/main.go`:

```go
func (a *aiLLMAdapter) ChatWithToolsStream(msgs []runtime.Message, tools []map[string]interface{}, model string, onChunk func(string)) (*runtime.LLMResponse, error) {
	if model == "" {
		model = a.model
	}
	req := ai.ChatCompletionRequest{
		Model:    model,
		Messages: make([]ai.Message, len(msgs)),
		Stream:   true,
	}
	for i, m := range msgs {
		req.Messages[i] = ai.Message{Role: m.Role, Content: m.Content}
	}
	var fullContent string
	err := a.client.ChatCompletionStream(req, func(chunk string) {
		fullContent += chunk
		onChunk(chunk)
	})
	if err != nil {
		return nil, err
	}
	return &runtime.LLMResponse{Type: "final", Content: fullContent}, nil
}
```

- [ ] **Step 2: Build and test**

Run: `go build ./cmd/agent`
Expected: no errors

Run: `go test ./... -count=1 -short`
Expected: all pass

- [ ] **Step 3: Commit**

```bash
git add cmd/agent/main.go
git commit -m "feat: add ChatWithToolsStream to aiLLMAdapter"
```

---

### Task 5: Wire DispatchStream in app.go

**Files:**
- Modify: `internal/ui/app.go:586-618` (orchestrator branch of sendMessage)

- [ ] **Step 1: Replace Dispatch with DispatchStream in orchestrator branch**

Find the orchestrator branch (around line 586) and replace the entire block:

Current code (lines 586-618):
```go
		if a.orch != nil {
			task := &orchestrator.Task{
				ID:      uuid.New().String(),
				Type:    orchestrator.TaskExecute,
				Content: text,
			}
			results, err := a.orch.Dispatch(task)
			a.QueueUpdateDraw(func() {
				a.isLoading = false
				// Remove the placeholder assistant message
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
				a.chatPanel.SetSession(sessionPtr)
				a.chatPanel.ScrollToBottom()
			})
			return
		}
```

Replace with:
```go
		if a.orch != nil {
			task := &orchestrator.Task{
				ID:      uuid.New().String(),
				Type:    orchestrator.TaskExecute,
				Content: text,
			}

			var streamBuf string
			results, err := a.orch.DispatchStream(task, func(chunk string) {
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
				// Remove the placeholder assistant message
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
```

- [ ] **Step 2: Build and test**

Run: `go build ./cmd/agent`
Expected: no errors

Run: `go test ./internal/ui/... -count=1 -short`
Expected: all pass

- [ ] **Step 3: Commit**

```bash
git add internal/ui/app.go
git commit -m "feat: wire DispatchStream in sendMessage orchestrator path"
```

---

### Task 6: End-to-end verification

- [ ] **Step 1: Full build and test**

Run: `make build`
Expected: builds `build/agent` binary

Run: `make test`
Expected: all tests pass

- [ ] **Step 2: Commit any remaining files**

```bash
git add -A
git commit -m "chore: finalize streaming orchestrator implementation"
```
