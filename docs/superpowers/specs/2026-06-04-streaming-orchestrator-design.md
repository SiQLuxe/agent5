# Streaming Orchestrator Design

## Problem

The Orchestrator path in `sendMessage()` uses blocking `Dispatch()` → `agent.Execute()` → `llm.ChatWithTools()` → `ai.Client.ChatCompletion()`. The entire response is received at once after a long wait, with no incremental output visible to the user. The streaming infrastructure exists at the UI layer (`ChatPanel.StartStreaming/UpdateStreaming/EndStreaming`) and the `ai.Client` layer (`ChatCompletionStream`), but the middle layers have no streaming plumbing.

## Approach

Add streaming through all 4 layers: LLMClient → Agent → Orchestrator → UI. Each layer gets a streaming variant that accepts an `onChunk func(string)` callback to forward token deltas.

## Changes

### 1. `internal/agent/runtime/llm.go` — LLMClient 接口

Add `ChatWithToolsStream`:

```go
type LLMClient interface {
    ChatWithTools(messages []Message, tools []map[string]interface{}, model string) (*LLMResponse, error)
    ChatWithToolsStream(messages []Message, tools []map[string]interface{}, model string, onChunk func(string)) (*LLMResponse, error)
}
```

Update `MockLLMClient` to implement the new method.

### 2. `cmd/agent/main.go` — aiLLMAdapter

Add `ChatWithToolsStream` implementation using `ai.Client.ChatCompletionStream`:

```go
func (a *aiLLMAdapter) ChatWithToolsStream(msgs []runtime.Message, tools []map[string]interface{}, model string, onChunk func(string)) (*runtime.LLMResponse, error) {
    if model == "" {
        model = a.model
    }
    req := ai.ChatCompletionRequest{Model: model, Stream: true}
    req.Messages = make([]ai.Message, len(msgs))
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

### 3. `internal/agent/runtime/agent.go` — Agent

Add `ExecuteStream`:

```go
func (a *Agent) ExecuteStream(task string, onChunk func(string)) (string, error) {
    a.Memory.Clear()
    a.Logger.Clear()
    return a.reactLoopStream(task, onChunk)
}
```

### 4. `internal/agent/runtime/react.go` — reactLoopStream

Duplicated from `reactLoop` with `ChatWithToolsStream` call:

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
            tc := tool.ToolContext{Context: nil, SandboxDir: a.Config.SandboxDir}
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

### 5. `internal/agent/orchestrator/orchestrator.go` — Orchestrator

Add `DispatchStream`:

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

### 6. `internal/ui/app.go` — UI Wiring

Replace `Dispatch` with `DispatchStream` in `sendMessage()`:

```go
if a.orch != nil {
    task := &orchestrator.Task{ID: uuid.New().String(), Type: orchestrator.TaskExecute, Content: text}
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

## Files Changed

| File | Change |
|------|--------|
| `internal/agent/runtime/llm.go` | Add `ChatWithToolsStream` to interface + mock |
| `internal/agent/runtime/react.go` | Add `reactLoopStream` |
| `internal/agent/runtime/agent.go` | Add `ExecuteStream` |
| `internal/agent/orchestrator/orchestrator.go` | Add `DispatchStream` |
| `cmd/agent/main.go` | Add `aiLLMAdapter.ChatWithToolsStream` |
| `internal/ui/app.go` | Wire `DispatchStream` with streaming callback |

## Scope

- Only the LLM response generation is streamed (tokens). Tool execution within the ReAct loop is still blocking.
- The `aiLLMAdapter` currently ignores tools and always returns `"final"`, so the ReAct loop effectively runs exactly one iteration. Adding proper tool support is out of scope.
