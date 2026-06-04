# Agent System Design

Date: 2026-06-04

## Overview

Redesign the agent system from a hardcoded role-based architecture to a flexible Tool-based architecture. Each Agent is an execution unit with a set of Tools (capabilities), a ReAct loop for autonomous task completion, and support for multi-agent orchestration and external agent integration.

## Motivation

The current `AgentRole`/`TaskOrchestrator`/`CollaborationManager` architecture has fundamental limitations:
- All "agents" share the same LLM backend with only prompt differences
- No tool-use capability — agents can only chat, not read/write files or execute commands
- Multi-agent pipeline is hardcoded (4 steps, fixed order)
- ExternalAgent is dead code, never wired up
- No decision logging, no error recovery, no context management

## Architecture

```
internal/agent/
  tool/           Tool interface + built-in implementations
  runtime/        Agent runtime engine (ReAct loop)
  orchestrator/   Multi-agent orchestration
  bridge/         External agent adapters
```

### 1. Tool System (`internal/agent/tool/`)

**Core Interface:**

```go
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
    Context  context.Context
    Memory   *AgentMemory
    Logger   *AgentLogger
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

**Built-in Tools:**

| Tool | Parameters | Description |
|---|---|---|
| `read_file` | path (required) | Read file content |
| `write_file` | path (required), content (required) | Write/create file |
| `search_text` | pattern (required), include (optional) | Search codebase |
| `read_dir` | path (required) | List directory entries |
| `exec_command` | command (required), timeout (optional) | Run shell command |
| `git_diff` | (none) | Show working tree diff |
| `git_log` | max_count (optional) | Show recent commits |
| `chat_llm` | prompt (required), model (optional) | Call LLM |
| `delegate` | agent (required), task (required) | Delegate to sub-agent |
| `ask_user` | question (required) | Ask user for input |

**Registry:**

```go
type ToolRegistry struct {
    tools map[string]Tool
}
```

Supports `Register`, `Get`, `List`, `AsToolDefinitions` (for LLM function calling).

### 2. Agent Runtime (`internal/agent/runtime/`)

```go
type Agent struct {
    Name        string
    Config      AgentConfig
    Tools       *tool.ToolRegistry
    Memory      *AgentMemory
    Logger      *AgentLogger
}

type AgentConfig struct {
    Model           string
    SystemPrompt    string
    MaxReActLoop    int           // default 20
    Temperature     float64
    ContextLimit    int           // max messages before summarization
}
```

**Execution Flow (ReAct Loop):**

```
1. Build messages: [SystemPrompt, History..., Task]
2. Call LLM with tool definitions
3. LLM returns: thought + tool_call or final_answer
4. If tool_call:
   a. Validate params
   b. Execute Tool
   c. Append result as observation
   d. Go to step 2
5. If final_answer:
   a. Return result
```

**Memory:**

```go
type AgentMemory struct {
    ShortTerm   []Message        // current task messages
    WorkingSet  map[string]string // key-value working memory
}

type AgentLogger struct {
    entries    []LogEntry
    maxEntries int
}

type LogEntry struct {
    Timestamp  time.Time
    Phase      string   // thought / tool_call / tool_result / final
    Content    string
    ToolName   string
    ToolParams map[string]interface{}
    ToolResult *tool.ToolResult
    Duration   time.Duration
}
```

### 3. Multi-Agent Orchestration (`internal/agent/orchestrator/`)

```go
type Orchestrator struct {
    registry   *AgentRegistry
    decomposer *TaskDecomposer
    merger     *ResultMerger
}

type AgentRegistry struct {
    agents map[string]*runtime.Agent
    roles  map[string][]string // role tag → agent names
}
```

**Task Types:**

| Type | Description |
|---|---|
| `task_analyze` | Analyze requirements |
| `task_design` | Design architecture |
| `task_code` | Write code |
| `task_review` | Review code |
| `task_execute` | Run/execute |
| `task_research` | Research/investigate |
| `task_custom` | Custom task |

**TaskDecomposer** — strategies for splitting tasks:
- `sequential`: step-by-step pipeline
- `parallel`: split into independent parallel tasks
- `custom`: user-defined split

**ResultMerger** — strategies for merging results:
- `concat`: concatenate all results
- `summary`: LLM-summarize combined results
- `pick_best`: select best result (for parallel execution)

### 4. External Agent Bridge (`internal/agent/bridge/`)

```go
type OpencodeBridge struct {
    backend    backend.AgentBackend
    sessionIDs map[string]string  // taskID → sessionID
}

type ClaudeCodeBridge struct {
    // similar, different transport
}
```

Each bridge implements the `Tool` interface, so external agents appear as a `delegate` tool to the calling Agent.

Bridge lifecycle:
1. `Start()` — create session
2. `Execute()` — send task, get response
3. `Stop()` — close/cleanup session

### 5. Integration with Existing Code

**Config changes** (`config.go`):
```go
type AgentConfig struct {
    Enabled     bool     `toml:"enabled"`
    Model       string   `toml:"model"`
    Tools       []string `toml:"tools"`
    MaxReAct    int      `toml:"max_react_loop"`
}
```

**Main.go wiring:**
- Create ToolRegistry with built-in tools
- Create Agent instances from config
- Create Orchestrator
- Wire TUI commands: `/agent`, `/delegate`, `/agents`

**Existing code migration:**
- Remove `agent_roles.go` (PlanningAgent, CodingAgent, etc.) — replaced by generic Agent with different system prompts
- Rewrite `task_orchestrator.go` — replaced by new Orchestrator
- Keep `ExternalAgent` concept but implement as bridge Tool

## Detailed Implementation: Tool System

### Tool Schema Format

Each Tool declares its parameters for LLM function calling:

```go
func (t *ReadFileTool) Schema() ToolSchema {
    return ToolSchema{
        Parameters: map[string]ParamSchema{
            "path": {Type: "string", Description: "File path relative to project root"},
        },
        Required: []string{"path"},
    }
}
```

The `AsToolDefinitions()` method converts to a format compatible with OpenAI/DeepSeek function calling:

```go
func (r *ToolRegistry) AsToolDefinitions() []map[string]interface{} {
    // Returns: [{type: "function", function: {name, description, parameters: {type: "object", properties: {...}, required: [...]}}}]
}
```

### Supported LLM Formats

DeepSeek and OpenAI use the same function calling format (`tools` + `tool_choice`). Anthropic uses a different format (`tool_use` content blocks). The ToolDefinition converter handles both.

## Detailed Implementation: ReAct Loop

The `reactLoop` function is the core of the Agent:

```go
func (a *Agent) reactLoop(ctx ToolContext, task string) (string, error) {
    messages := a.buildMessages(task)
    
    for i := 0; i < a.Config.MaxReActLoop; i++ {
        llmResp := a.callLLM(ctx, messages)
        a.Logger.Log(llmResp)
        
        switch llmResp.Type {
        case "tool_call":
            result := a.executeTool(ctx, llmResp.ToolCall)
            messages = append(messages, result.asObservation())
            a.Memory.ShortTerm = append(a.Memory.ShortTerm, result.asMessage())
            
        case "final":
            a.Memory.ShortTerm = append(a.Memory.ShortTerm, finalMessage)
            return llmResp.Content, nil
        }
    }
    return "", fmt.Errorf("max react loop iterations reached")
}
```

**Error handling:**
- Tool execution error → returned as observation, LLM decides next action
- LLM call error → retry with backoff (max 3 retries)
- Context overflow → summarize oldest messages, keep recent

## Detailed Implementation: Task Decomposition

```go
type TaskDecomposer struct {
    strategies map[string]DecomposeStrategy
}

type DecomposeStrategy func(task *Task) ([]*Task, error)

func (d *TaskDecomposer) Decompose(task *Task) ([]*Task, error) {
    strategy, ok := d.strategies[string(task.Type)]
    if !ok {
        return d.defaultStrategy(task)
    }
    return strategy(task)
}
```

**Sequential strategy** (default for analyze/design tasks):
```
Input: "Add login feature"
Output:
  1. [analyze] 分析登录功能需求
  2. [design]  设计登录API和数据库
  3. [code]    实现登录代码
  4. [review]  审查登录代码
```

**Parallel strategy** (for independent sub-tasks):
```
Input: "Refactor utils package"
Output:
  1. [code] 重构 string_utils.go
  2. [code] 重构 file_utils.go (并行执行)
  3. [code] 重构 net_utils.go  (并行执行)
```

## Detailed Implementation: TUI Integration

The TUI gets a new `/agent` command and an Agent execution panel:

- **`/agent list`** — show available agents and their tools
- **`/agent run "task"`** — execute task with default agent
- **`/agent run --agent coder "task"`** — execute with specific agent
- **`/delegate "task" to opencode`** — delegate to external agent

UI components (added/modified):
- `internal/ui/agent_panel.go` — shows agent execution status, current tool, progress
- `internal/ui/agent_log_view.go` — expandable decision log viewer

## Testing Strategy

**Unit tests** (for each package):
- `tool/*_test.go` — test each tool with mocked filesystem/commands
- `runtime/*_test.go` — test ReAct loop with mock LLM
- `orchestrator/*_test.go` — test decomposition, dispatch, merging

**Integration tests** (`internal/agent/`):
- Full execution with real LLM (short test, skip with `testing.Short()`)
- Multi-agent workflow with mock LLM

## Migration

**Phase 1** — Create `internal/agent/` package structure with Tool system (no existing code changed)
- `internal/agent/tool/tool.go` — interface + registry
- `internal/agent/tool/read_file.go` + write_file + search_text + read_dir + exec_command + chat_llm
- Tests for each tool

**Phase 2** — Agent runtime
- `internal/agent/runtime/agent.go`
- `internal/agent/runtime/memory.go`
- `internal/agent/runtime/logger.go`
- `internal/agent/runtime/react_loop.go`
- Tests with mock LLM

**Phase 3** — Orchestrator
- `internal/agent/orchestrator/orchestrator.go`
- `internal/agent/orchestrator/registry.go`
- `internal/agent/orchestrator/decomposer.go`
- `internal/agent/orchestrator/merger.go`
- Remove old agent_roles.go task_orchestrator.go collaboration_manager.go or replace

**Phase 4** — Bridge
- `internal/agent/bridge/opencode.go`
- Wire through existing AgentBackend

**Phase 5** — TUI integration + config wiring
- Agent panel
- Agent log view
- main.go wiring
- Config changes
