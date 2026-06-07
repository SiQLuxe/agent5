# Subagent 并发执行系统设计

## 概述

为 Agent-TUI 添加并发 subagent 能力，分三阶段落地。参考 CodeWhale（LLM 驱动的动态调度）、OpenCode（`task` 工具 + 独立 Session）和 Codex（配置驱动并行）的设计，结合 Go 的 goroutine/channel 模型。

## Phase 1: Orchestrator 并发分解

### 目标
在现有 `agent/orchestrator.Orchestrator` 上增加并发 `DispatchConcurrent()` 方法，无依赖的子任务并行执行。

### 改动

**internal/agent/orchestrator/orchestrator.go** — 新增方法：

```go
type Orchestrator struct {
    registry   *Registry
    decomposer *Decomposer
    merger     *Merger
    maxConcurrent int  // 并发上限，默认 5
}

func (o *Orchestrator) DispatchConcurrent(sessionID string, task *Task) ([]*Task, error)
```

### 数据流

```
task → Decomposer.Decompose()
    ↓
子任务列表 []*Task
    ↓
semaphore chan struct{} (maxConcurrent)
    ↓
for each step → go func() { agent.Execute(); results <- result }
    ↓
wg.Wait() → 收集 results → Merger.Merge()
```

### 并发控制

- 用 buffered channel 做 semaphore，控制 goroutine 上限
- 每个子任务跑 `agent.Execute()`（同步调用，goroutine 包装）
- 完成后结果写入 channel，主 goroutine 收集

### 错误处理

- 单个子任务失败不影响其他子任务（尽量继续）
- 最终结果中标记每个子任务的 `Status`（completed / failed）
- 调用者通过 `Merger.Merge()` 看到完整状态

### 测试

- 现有 `orchestrator_test.go` 补充 `TestDispatchConcurrent`
- 验证并发执行（mock agent sleep + 检查总耗时 < 串行耗时）

### 估算改动

| 文件 | 改动 |
|------|------|
| `orchestrator/orchestrator.go` | +40 行 (`DispatchConcurrent`) |
| `orchestrator/orchestrator_test.go` | +60 行 (测试) |
| `orchestrator/types.go` | +1 字段 (`maxConcurrent`) |
| 合计 | ~100 行 |

---

## Phase 2: task 工具 + SubagentManager

### 目标

给 LLM 提供 `task` 工具，让 LLM 自主 spawn 子 agent。参考 OpenCode 的 `task` 工具设计。

### 新增组件

**internal/agent/tool/task.go** — `TaskTool` 工具：

```go
type TaskTool struct {
    manager *SubagentManager
}

func (t *TaskTool) Name() string           { return "task" }
func (t *TaskTool) Description() string    { return "Delegate a task to a subagent" }
func (t *TaskTool) Schema() ToolSchema     { /* description, prompt, subagent_type, background */ }
func (t *TaskTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult
```

**internal/agent/tool/subagent_manager.go** — `SubagentManager`：

```go
type SubAgentConfig struct {
    Name         string
    SystemPrompt string
    Model        string
    Tools        []string  // 允许的工具列表
    MaxDepth     int       // 嵌套深度，默认 3
}

type SubAgent struct {
    ID      string
    Name    string
    Config  SubAgentConfig
    Session *session.Manager
    Cancel  context.CancelFunc
    Status  string  // running / completed / failed / cancelled
    Result  string
    Created time.Time
}

type SubagentManager struct {
    mu     sync.RWMutex
    agents map[string]*SubAgent
    max    int  // 并发上限
}
```

### 工具接口

```
task(description, prompt, subagent_type, background?)
  → Foreground: 返回子 agent 结果文本
  → Background: 返回 agent_id, 完成后注入通知到父 session
```

### SubAgent 类型

| 类型 | 用途 | 默认工具 |
|------|------|---------|
| `general` | 通用多步骤任务 | 全部 |
| `explore` | 只读代码库搜索 | read_file, search_text |
| `review` | 代码审查 | read_file, search_text, exec_command(test) |

### 生命周期

```
task() 调用
    ↓
SubagentManager.spawn()  → 检查并发上限 + 嵌套深度
    ↓
创建子 Session (parentID 关联)
    ↓
goroutine: agent.Execute(sessionID, prompt)
    ├── foreground: 直接返回结果
    └── background: 返回 agent_id, 完成时 inject 消息到父 session
```

### 估算改动

| 文件 | 改动 |
|------|------|
| `agent/tool/task.go` | 新文件 ~150 行 |
| `agent/tool/subagent_manager.go` | 新文件 ~120 行 |
| `agent/tool/subagent_manager_test.go` | 新文件 ~100 行 |
| `agent/tool/task_test.go` | 新文件 ~80 行 |
| `cmd/agent/main.go` | +5 行 (注册 TaskTool) |
| 合计 | ~455 行 |

---

## Phase 3: 完善

- **进度推送** — Subagent 的实时消息通过 channel 推回 UI
- **agent_open/eval/close** — 更细粒度的生命周期控制（CodeWhale 风格）
- **嵌套深度限制** — 默认 3 层
- **并发上限可配置** — `config.toml` 中 `max_subagents` 字段
- **存活检测** — heartbeat 超时自动 cancel 卡死的 agent

---

## 关键设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 触发方式 | `task` 单工具（OpenCode 风格） | 比 agent_open/eval/close 三工具更简单，Go 中实现更直接 |
| 并发模型 | goroutine + channel | Go 原生模式，无需三方库，与现有 `session.Manager` 兼容 |
| Session 隔离 | 独立子 Session + parentID | 复用现有 session 系统，消息历史隔离，父 agent 可通过 agent_eval 查询 |
| 上下文传递 | Prompt 文本传入（不 fork context） | Go 当前无 prefix cache 概念，简单直接 |
| 前后台模式 | Foreground 为主，Background 为实验特性 | 前台语义简单可靠，后台需要额外通知机制 |
| 工具过滤 | 子 agent 启动时从全局 registry 白名单筛选构建子 registry | 复用 `main.go:218` 现有模式，`Tools` 为空时继承父 agent 全部工具 |
| 后台通知 | Background 完成时调用 `session.Manager.AddMessage()` 向父 session 注入 system 消息，含 `<subagent id="..." status="completed">` | 父 agent 在下一次 LLM 调用时自动看到通知 |
