# External Agent Backend Integration

## Goal

Add an abstraction layer to Agent-TUI that allows it to interact with and control external AI coding agents (opencode, Claude Code, Codex, etc.) through a unified `AgentBackend` interface, with bidirectional control support.

## Background

Agent-TUI currently has its own multi-agent system (Planning/Coding/Review/Execution agents) using local AI clients (OpenAI, DeepSeek, Anthropic, Local). There is no integration with external coding agents like opencode.

Opencode v1.15.13 exposes:
- HTTP REST API via `opencode serve` (session CRUD, messaging, commands, files, events/SSE)
- ACP protocol (JSON-RPC over stdio, for editor integration)
- JS/TS SDK (`@opencode-ai/sdk`)

Claude Code and Codex CLI primarily expose CLI/stdio interfaces.

## Architecture

### Package Structure

```
internal/backend/
  backend.go              # AgentBackend 接口 + 数据模型
  factory.go              # NewBackend(type, config) -> AgentBackend
  registry.go             # BackendRegistry: 管理多 backend 实例
  process.go              # 通用子进程生命周期管理

  opencode/
    backend.go            # OpencodeBackend struct + Start/Stop
    sessions.go           # 会话相关 API 调用
    messages.go           # 消息相关 API 调用
    commands.go           # 命令/文件 API 调用
    events.go             # SSE 事件监听 -> channel

internal/server/
  server.go               # Agent-TUI HTTP Server (反向控制)
  routes.go               # 路由注册
  handlers.go             # handler 实现
```

### AgentBackend 接口

```go
type AgentType string

const (
    TypeOpencode   AgentType = "opencode"
    TypeClaudeCode AgentType = "claude-code"
    TypeCodex      AgentType = "codex"
)

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

### 数据模型

```go
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

type CommandResult struct {
    ExitCode int    `json:"exit_code"`
    Stdout   string `json:"stdout"`
    Stderr   string `json:"stderr"`
}

type Event struct {
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}
```

## OpencodeBackend Adapter

### 进程管理

支持两种模式，通过配置切换：

```
模式 A (auto_start = true):
  1. 查找 opencode 二进制（配置路径或 $PATH）
  2. 启动 opencode serve --port 0
  3. 解析输出获取实际端口
  4. 设置 baseURL = http://127.0.0.1:{port}

模式 B (auto_start = false):
  1. 直接用配置的 api_url
  2. 调用 GET /global/health 验证连通性
```

#### SSE 共享策略

`Events()` 和 `SendMessageStream` 都依赖 opencode 的 `GET /event` SSE 流。设计一个共享 SSE 连接：

```
OpencodeBackend 持有一个 SSE 连接
  ↓
SSE 事件到达 → 根据 event.type 分发:
  ├── 有等待中的 stream callback → 调用 onChunk(chunk)
  └── 否则 → 推入 Events() 返回的 channel
```

当 `SendMessageStream` 被调用时：
1. 通过 `POST /session/:id/prompt_async` 发送消息（返回 204）
2. 在共享 SSE 连接上注册一个临时 listener，按 `sessionID` 过滤事件
3. 收到完整响应后取消临时 listener

### API 映射

| AgentBackend    | Opencode REST API                     |
|-----------------|---------------------------------------|
| Health          | GET /global/health                    |
| CreateSession   | POST /session                         |
| ListSessions    | GET /session                          |
| GetSession      | GET /session/:id                      |
| DeleteSession   | DELETE /session/:id                   |
| SendMessage     | POST /session/:id/message             |
| SendMessageStream | POST /session/:id/prompt_async + SSE |
| GetMessages     | GET /session/:id/message              |
| ExecuteCommand  | POST /session/:id/command             |
| ExecuteShell    | POST /session/:id/shell               |
| ReadFile        | GET /file/content?path=...            |
| SearchText      | GET /find?pattern=...                 |
| Events          | GET /event (SSE)                      |

### 反向控制: Agent-TUI HTTP Server

Agent-TUI 暴露小型 HTTP server，供外部 agent 回调或外部程序控制 TUI：

| Endpoint | Method | 用途 |
|---|---|---|
| /api/tui/append-prompt | POST | 向输入框追加文字 |
| /api/tui/submit-prompt | POST | 提交当前 prompt |
| /api/tui/show-toast | POST | 显示通知 |
| /api/tui/execute-command | POST | 执行内部命令 |
| /api/events | GET | SSE 事件流 |

### 调用链路

```
正向: 用户输入 → ExternalAgent → AgentBackend → OpencodeBackend → opencode serve → LLM
反向: opencode serve → {任意外部 HTTP 客户端} → Agent-TUI HTTP Server → TUI 操作
```

## 配置模型

```toml
[agent.backends]

  [agent.backends.opencode]
  type = "opencode"
  enabled = true
  auto_start = true
  binary = ""
  # auto_start = false 时使用:
  # api_url = "http://127.0.0.1:4096"
  # api_key = ""

  [agent.backends.claude-code]
  type = "claude-code"
  enabled = false
  binary = "claude"
```

### BackendRegistry

```go
type Registry struct {
    backends map[AgentType]AgentBackend
}

func (r *Registry) Get(t AgentType) (AgentBackend, bool)
func (r *Registry) GetAll() []AgentBackend
func (r *Registry) ActiveBackends() []AgentBackend
```

## 与现有系统集成

### 新增 AgentRole

```go
// internal/service/agent_roles.go
type ExternalAgent struct {
    backend backend.AgentBackend
}

func NewExternalAgent(b backend.AgentBackend) *ExternalAgent
func (e *ExternalAgent) Execute(task *Task) (string, error)
func (e *ExternalAgent) GetRoleName() string
func (e *ExternalAgent) GetSupportedTaskTypes() []TaskType
```

```go
const TaskExternal TaskType = "external"
```

### 初始化流程

```
Agent-TUI 启动
  ↓
加载 TOML 配置
  ↓
遍历 [agent.backends], enabled=true 的:
  ↓
factory.NewBackend(type, config)
  ↓
backend.Start(ctx)           ← 拉起子进程 / 连接外部实例
  ↓
health check 通过后注册到 BackendRegistry
  ↓
TaskOrchestrator.RegisterAgent(TaskExternal, ExternalAgent{backend})
  ↓
Agent-TUI HTTP Server 启动
  ↓
TUI 界面就绪
```

## 错误处理

### 进程崩溃

- `ProcessManager.watch()` 监控子进程退出
- 自动重启最多 3 次，指数退避
- 超过上限后标记 backend 为 unhealthy，通知 TUI

### 断连重连

- 每次 API 调用前检查 health
- SSE 断开后自动重新订阅
- 重试策略: 3 次后标记 unhealthy

### 超时

- 普通 API 调用: 30s context timeout
- Stream: 无超时，由用户取消驱动
- 进程启动: 60s timeout

### Graceful Shutdown

```
SIGTERM/SIGINT
  ↓
依次 Stop 所有 backend (发送 SIGTERM → 等待 5s → SIGKILL)
  ↓
取消所有进行中的 stream
  ↓
停止 Agent-TUI HTTP Server
  ↓
退出
```

## 文件清单

| File | Action |
|------|--------|
| internal/backend/backend.go | Create — 接口 + 数据模型 |
| internal/backend/factory.go | Create — NewBackend |
| internal/backend/registry.go | Create — BackendRegistry |
| internal/backend/process.go | Create — ProcessManager |
| internal/backend/opencode/backend.go | Create — OpencodeBackend |
| internal/backend/opencode/sessions.go | Create |
| internal/backend/opencode/messages.go | Create |
| internal/backend/opencode/commands.go | Create |
| internal/backend/opencode/events.go | Create |
| internal/server/server.go | Create — HTTP server |
| internal/server/handlers.go | Create |
| internal/service/agent_roles.go | Modify — 添加 ExternalAgent |
| internal/service/task_orchestrator.go | Modify — 添加 TaskExternal |
| internal/config/config.go | Modify — 添加 backend 配置 |
| app.go | Modify — 初始化流程 |

## Future: 其他 Agent 适配器

后续 adapter 的传输方式:

| Agent | 传输层 | 实现方式 |
|-------|--------|---------|
| Claude Code | stdio | 启动 `claude` 子进程，解析结构化输出 |
| Codex CLI | stdio | 启动 `codex` 子进程，解析结构化输出 |

共享同一个 `AgentBackend` 接口，只需新增 adapter 包即可。
