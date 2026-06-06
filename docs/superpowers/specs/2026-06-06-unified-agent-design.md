# UnifiedAgent Design

## Problem

The codebase has two parallel execution paths for handling user messages:

1. **AIAssistant** (`internal/service/ai_assistant.go`): preserves conversation history but has no tool access or ReAct loop.
2. **Runtime Agent** (`internal/agent/runtime/`): has tool infrastructure and ReAct loop but clears memory each call (no conversation history), and its `aiLLMAdapter` ignores tool definitions entirely — tools are registered but never sent to the LLM.

Both paths degenerate to a single LLM call per user message. Neither is a complete solution.

## Goal

Replace both paths with a single **UnifiedAgent** that handles conversation history, tool execution, and ReAct reasoning in one cohesive flow.

## Architecture

### Layered Design

```
┌─────────────────────────────────────────────┐
│                  UI (app.go)                 │
├─────────────────────────────────────────────┤
│         Orchestrator (可选，任务编排)         │
├─────────────────────────────────────────────┤
│              UnifiedAgent                    │
│  ┌───────────┐  ┌──────────────────────┐    │
│  │ Session   │  │    ReAct Loop        │    │
│  │ Manager   │  │  ┌────────────────┐  │    │
│  │           │  │  │ LLM + Tools    │  │    │
│  │ history   │  │  │ Tool Registry  │  │    │
│  │ .History  │  │  └────────────────┘  │    │
│  └───────────┘  └──────────────────────┘    │
└─────────────────────────────────────────────┘
```

**Layers:**
- **UnifiedAgent**: single entry point. Receives `SessionID`, `Message`, `Tools`. Returns result.
- **SessionManager**: wraps `history.History`. Manages conversation context with configurable truncation.
- **ReAct Loop**: iterates thought → tool_call → observation → final.
- **LLM Client**: extended to support `tools` parameter and parse `tool_calls` responses.
- **Tool Registry**: unchanged — tools execute as before.
- **Orchestrator**: optional. Wraps UnifiedAgent. Defaults to pass-through when no decomposition strategy is configured.

### Orchestrator + UnifiedAgent

Orchestrator is an optional higher-level layer that delegates to UnifiedAgent:

```go
func (o *Orchestrator) DispatchStream(task, onChunk) {
    if needsDecompose(task) {
        steps := o.decomposer.Decompose(task)
        for _, step := range steps {
            result, _ := o.agent.ExecuteStream(step, onChunk)
        }
        return mergeAll(results)
    }
    return o.agent.ExecuteStream(task, onChunk)  // pass-through
}
```

This follows the same pattern as LangChain's AgentExecutor + Chain: Orchestrator handles *what* to do (step-level planning), UnifiedAgent handles *how* to do it (tool-level execution).

## Data Flow

```
User input → app.sendMessage(text)
  │
  ├─ "/cmd" → SkillExecutor
  │            └─ skill prompt → UnifiedAgent.ExecuteStream(session, text)
  │
  └─ normal text → Orchestrator (if present and task needs decomposition)
       │            └─ else pass-through
       │
       └─ UnifiedAgent.ExecuteStream(session, text, tools, onChunk)
            │
            ├─ SessionManager.GetContext(sessionID)
            │   → [sys, hist..., {user, text}]
            │
            ├─ ReAct Loop:
            │   ├─ LLM.ChatWithTools(messages, tools)
            │   │   → parse tool_call or final
            │   ├─ tool_call: execute → append result → loop
            │   └─ final: stream callback → break
            │
            └─ SessionManager.AddMessage(session, "user", text)
            └─ SessionManager.AddMessage(session, "assistant", result)
```

## Core Interface

```go
type ExecuteInput struct {
    SessionID string
    Message   string
}

type UnifiedAgent interface {
    ExecuteStream(ctx context.Context, input ExecuteInput, onChunk func(string)) (string, error)
    CreateSession() (string, error)
    ListSessions() ([]Session, error)
    SwitchModel(model string) error
    ListModels() ([]string, error)
}
```

## Error Handling

| Scenario | Handling |
|----------|----------|
| LLM returns invalid tool_call | Append error as tool result, let LLM recover |
| Tool execution fails | Return error message in tool result (no panic) |
| ReAct loop exhausted | Return partial result with warning |
| LLM API 4xx/5xx | Retry up to 3 times, then return user-visible error |
| Context window overflow | Sliding window: system + recent N turns (N configurable) |
| History + tools exceed limit | Truncate oldest history, keep system + tool definitions |

## Migration Plan

### Step 1: Extend LLM Client for tools

Files: `internal/ai/client.go`, `openai.go`, `deepseek.go`

- Add `Tools []ToolDefinition` field to `ChatCompletionRequest`
- Add `ToolCalls []ToolCall` to `ChatCompletionResponse`
- Implement serialization/deserialization per provider (OpenAI-compatible APIs only)
- Local client skips tools

### Step 2: Create SessionManager

New file: `internal/agent/session/manager.go`

- Thin wrapper around `history.History`
- `GetContext(sessionID, maxTurns)` returns truncated message history
- Methods: CreateSession, ListSessions, AddMessage, GetContext, DeleteSession
- Migrate SwitchModel/ListModels from AIAssistant

### Step 3: Rewrite UnifiedAgent

Files: `internal/agent/runtime/agent.go`, `react.go`

- `ExecuteStream` accepts `ExecuteInput` (SessionID + Message)
- Remove `Memory.Clear()` — build messages from SessionManager history
- ReAct Loop sends tools to LLM, parses tool_calls, executes tools
- Save user + assistant messages to SessionManager after completion
- Delete `runtime/memory.go` (replaced by SessionManager)

### Step 4: Update Orchestrator

File: `internal/agent/orchestrator/orchestrator.go`

- Use UnifiedAgent instead of old Agent
- Default to pass-through (no decomposition)

### Step 5: Update SkillExecutor

File: `internal/service/skill_executor.go`

- Change dependency from AIAssistant to UnifiedAgent

### Step 6: Update UI + main.go, delete AIAssistant

Files: `internal/ui/app.go`, `cmd/agent/main.go`

- Remove `internal/service/ai_assistant.go`
- Remove `aiLLMAdapter` from main.go
- UI routes all messages through UnifiedAgent
- Session creation/listing via SessionManager

## Testing

| Layer | What |
|-------|------|
| Unit | ReAct loop logic, message building, session CRUD, context truncation |
| Integration | Full UnifiedAgent flow with mock LLM, Orchestrator + Agent combo |
| E2E | Real LLM call with tool execution, multi-turn conversation |

## Deleted Files

- `internal/service/ai_assistant.go` — replaced by UnifiedAgent + SessionManager
- `internal/agent/runtime/memory.go` — replaced by SessionManager
- `cmd/agent/main.go` `aiLLMAdapter` — replaced by direct LLM client with tools

## Non-Goals

- Multi-agent routing across different roles (future Orchestrator enhancement)
- Task decomposition strategies beyond pass-through (future enhancement)
- Persistent session storage to disk (remains in-memory)
- Streaming tool execution progress in UI
