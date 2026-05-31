# Go TUI Agent 工具设计文档

## 1. 概述

本项目旨在开发一个比 DeepSeek-TUI 更强大的终端 AI 编程智能体工具，使用 Go 语言和 Bubble Tea 框架构建。该工具将提供多模型支持、代码编辑、调试工具、插件系统等增强功能。

## 2. 需求分析

### 2.1 核心功能需求

| 功能类别 | 需求描述 | 优先级 |
|---------|---------|-------|
| **代码编辑器** | 支持语法高亮、代码补全、文件编辑 | 高 |
| **AI 助手集成** | 支持 OpenAI、DeepSeek、Anthropic、本地模型 | 高 |
| **项目管理** | 文件管理、Git 集成、任务管理 | 高 |
| **调试工具** | 断点调试、日志查看、性能分析 | 中 |
| **插件系统** | 支持第三方插件扩展 | 中 |
| **终端界面** | 现代化 TUI，支持多模式切换 | 高 |

### 2.2 非功能需求

| 类别 | 要求 |
|-----|-----|
| **跨平台** | 支持 Windows、macOS、Linux |
| **性能** | 启动时间 < 2 秒，内存占用 < 50MB |
| **扩展性** | 模块化设计，易于扩展 |
| **安全性** | 操作批准机制，敏感操作需确认 |

## 3. 架构设计

### 3.1 分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                    UI Layer (TUI)                          │
│  ┌─────────┐ ┌─────────────┐ ┌─────────────┐ ┌──────────┐ │
│  │ Editor  │ │ Chat Panel  │ │ File Tree   │ │ Status   │ │
│  └────┬────┘ └──────┬──────┘ └──────┬──────┘ └────┬─────┘ │
└───────┼─────────────┼───────────────┼─────────────┼───────┘
        │             │               │             │
┌───────▼─────────────▼───────────────▼─────────────▼───────┐
│                   Business Logic Layer                     │
│  ┌─────────────────┐ ┌─────────────────┐ ┌───────────────┐ │
│  │ CodeEditor      │ │ AI Assistant    │ │ ProjectMgr    │ │
│  │  - Syntax High │ │  - API Clients  │ │  - File I/O   │ │
│  │  - Auto Comp   │ │  - Prompt Mgmt  │ │  - Git        │ │
│  └─────────────────┘ └─────────────────┘ └───────────────┘ │
└───────┬─────────────────┬─────────────────┬───────────────┘
        │                 │                 │
┌───────▼─────────────────▼─────────────────▼───────────────┐
│                    Data Access Layer                       │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐  │
│  │ Config   │ │ History  │ │ Plugins  │ │ AI Models    │  │
│  │ Storage  │ │ Storage  │ │ Registry │ │ Config Store │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────────┘  │
└───────────────────────────────────────────────────────────┘
```

### 3.2 模块职责

| 模块 | 层级 | 职责 |
|-----|------|------|
| `ui/` | UI层 | TUI 组件，处理用户交互和界面渲染 |
| `core/` | 业务逻辑层 | 核心业务逻辑，协调各功能模块 |
| `data/` | 数据访问层 | 数据持久化和配置管理 |
| `ai/` | AI层 | AI API 客户端，多模型支持 |

## 4. 目录结构

```
agent-tui/
├── cmd/
│   └── agent/              # 主入口
│       └── main.go
├── internal/
│   ├── ui/                 # TUI 层
│   │   ├── editor/         # 代码编辑器组件
│   │   ├── chat/           # 聊天面板组件
│   │   ├── filetree/       # 文件树组件
│   │   ├── status/         # 状态栏组件
│   │   └── layout/         # 布局管理
│   ├── core/               # 业务逻辑层
│   │   ├── editor/         # 编辑器逻辑
│   │   ├── assistant/      # AI 助手逻辑
│   │   ├── project/        # 项目管理逻辑
│   │   └── debugger/       # 调试器逻辑
│   ├── data/               # 数据访问层
│   │   ├── config/         # 配置存储
│   │   ├── history/        # 历史记录存储
│   │   ├── plugins/        # 插件注册表
│   │   └── models/         # AI 模型配置存储
│   └── ai/                 # AI API 客户端
│       ├── openai/         # OpenAI API
│       ├── deepseek/       # DeepSeek API
│       ├── anthropic/      # Anthropic API
│       └── local/          # 本地模型接口
├── pkg/                    # 公共库
├── plugins/                # 插件目录
├── configs/                # 配置文件模板
└── docs/                   # 文档
```

## 5. 核心组件设计

### 5.1 UI 层组件

#### 5.1.1 Editor 组件

**职责**：代码编辑器，支持语法高亮和代码补全

**接口设计**：
```go
type Editor interface {
    LoadFile(path string) error
    SaveFile(path string) error
    GetContent() string
    SetContent(content string)
    InsertAtCursor(text string)
    DeleteSelection()
    GetCursorPosition() (row, col int)
    SetCursorPosition(row, col int)
}
```

#### 5.1.2 ChatPanel 组件

**职责**：AI 对话面板，支持流式响应

**接口设计**：
```go
type ChatPanel interface {
    AddMessage(role string, content string)
    AppendToLastMessage(content string)
    Clear()
    GetMessages() []Message
    SetMessages(messages []Message)
}
```

#### 5.1.3 FileTree 组件

**职责**：文件浏览器，支持目录导航

**接口设计**：
```go
type FileTree interface {
    LoadDirectory(path string) error
    GetSelectedFile() string
    ExpandNode(path string)
    CollapseNode(path string)
    Refresh()
}
```

### 5.2 业务逻辑层

#### 5.2.1 CodeEditor 模块

**职责**：代码编辑核心逻辑

**功能**：
- 语法高亮（基于 Chroma）
- 代码补全（基于 AI 建议）
- 文件 I/O 操作

**接口设计**：
```go
type CodeEditorService interface {
    OpenFile(path string) (*Document, error)
    SaveFile(doc *Document) error
    HighlightSyntax(content string, language string) string
    GetCompletions(content string, position Position) []Completion
}
```

#### 5.2.2 AIAssistant 模块

**职责**：AI 助手核心逻辑

**功能**：
- 多模型支持（OpenAI、DeepSeek、Anthropic、本地）
- 工具调用管理
- 会话管理
- 流式响应处理

**接口设计**：
```go
type AIAssistant interface {
    SendMessage(messages []Message, model string) (<-chan string, error)
    GetModels() []ModelInfo
    SetActiveModel(modelID string)
    CreateSession() SessionID
    LoadSession(id SessionID) ([]Message, error)
    SaveSession(id SessionID, messages []Message) error
}
```

#### 5.2.3 ProjectManager 模块

**职责**：项目管理逻辑

**功能**：
- 文件管理（创建、删除、重命名）
- Git 集成（克隆、提交、推送）
- 任务管理

**接口设计**：
```go
type ProjectManager interface {
    ListFiles(path string) ([]FileInfo, error)
    CreateFile(path string, content string) error
    DeleteFile(path string) error
    RenameFile(oldPath, newPath string) error
    GitClone(url, dest string) error
    GitCommit(message string) error
    GitPush() error
}
```

#### 5.2.4 Debugger 模块

**职责**：调试工具逻辑

**功能**：
- 断点设置和管理
- 日志查看
- 性能分析

**接口设计**：
```go
type Debugger interface {
    SetBreakpoint(path string, line int) error
    RemoveBreakpoint(path string, line int) error
    GetBreakpoints() []Breakpoint
    StartDebugging() error
    StopDebugging() error
    StepOver() error
    StepInto() error
    GetLogs() []LogEntry
}
```

### 5.3 数据访问层

#### 5.3.1 ConfigStore

**职责**：配置管理

**存储结构**：
```go
type Config struct {
    APIKeys       APIKeys
    DefaultModel  string
    Theme         string
    ApprovalMode  ApprovalMode
    MaxSubagents  int
}

type APIKeys struct {
    OpenAI     string
    DeepSeek   string
    Anthropic  string
}
```

#### 5.3.2 HistoryStore

**职责**：会话历史存储

**存储方式**：SQLite 数据库

**表结构**：
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    name TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    role TEXT,
    content TEXT,
    timestamp TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);
```

### 5.4 AI 客户端层

#### 5.4.1 支持的模型

| 模型类型 | 支持模型 | API 端点 |
|---------|---------|---------|
| OpenAI | GPT-4, GPT-3.5 | https://api.openai.com/v1 |
| DeepSeek | DeepSeek V4 | https://api.deepseek.com |
| Anthropic | Claude 3 Opus/Sonnet/Haiku | https://api.anthropic.com/v1 |
| Local | 通过 OpenAI 兼容接口 | 自定义 |

**统一客户端接口**：
```go
type AIClient interface {
    Completion(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, error)
    GetModels() []ModelInfo
    ValidateAPIKey() error
}
```

## 6. 交互设计

### 6.1 主界面布局

```
┌─────────────────────────────────────────────────────────────┐
│  FileTree  │              Editor/Chat                      │
│  ┌──────┐  │  ┌─────────────────────────────────────────┐  │
│  │ .git │  │  │                                         │  │
│  │ src/ │  │  │  [Chat Mode]                            │  │
│  │ main.go│ │  │  You: What's the best way to...        │  │
│  │ utils/ │ │  │  AI: Let me analyze this for you...    │  │
│  └──────┘  │  │                                         │  │
│            │  └─────────────────────────────────────────┘  │
│            │  ┌─────────────────────────────────────────┐  │
│            │  │ Status: Connected | Model: GPT-4       │  │
│            │  │ Mode: Agent | Tasks: 0/3               │  │
│            │  └─────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 6.2 模式切换

| 模式 | 快捷键 | 功能描述 | 批准行为 |
|-----|-------|---------|---------|
| **Normal** | `Tab` | 标准聊天模式 | 手动批准写操作和 Shell |
| **Plan** | `Tab` | 只读规划模式 | 禁止所有写操作 |
| **Agent** | `Tab` | 智能体模式 | 自动批准文件写入，手动批准 Shell |
| **YOLO** | `Tab` | 全自动模式 | 自动批准所有操作 |

### 6.3 快捷键设计

| 快捷键 | 功能 |
|-------|------|
| `Ctrl+N` | 新建文件 |
| `Ctrl+O` | 打开文件 |
| `Ctrl+S` | 保存文件 |
| `Ctrl+Q` | 退出 |
| `Tab` | 切换模式 |
| `Ctrl+Tab` | 切换面板 |
| `/help` | 显示帮助 |
| `/model` | 切换模型 |

## 7. 技术栈

| 分类 | 技术 | 版本 |
|-----|------|-----|
| **语言** | Go | 1.22+ |
| **TUI 框架** | Bubble Tea | latest |
| **样式** | Lipgloss | latest |
| **语法高亮** | Chroma | latest |
| **数据库** | SQLite | 3.45+ |
| **HTTP 客户端** | net/http | 内置 |
| **Git 库** | go-git | latest |

## 8. 安全性设计

### 8.1 操作批准机制

| 操作类型 | 风险等级 | 默认批准行为 |
|---------|---------|-------------|
| 文件读取 | 低 | 自动批准 |
| 文件写入 | 中 | 手动批准（Agent 模式自动） |
| Shell 执行 | 高 | 手动批准（YOLO 模式自动） |
| Git 推送 | 高 | 手动批准 |

### 8.2 API Key 安全

- API Key 存储在加密配置文件中
- 支持环境变量覆盖
- 不记录或打印 API Key

## 9. 测试策略

### 9.1 单元测试

- 每个模块至少 80% 覆盖率
- 测试边界条件和错误场景

### 9.2 集成测试

- 测试模块间交互
- 测试 AI API 客户端

### 9.3 端到端测试

- 测试完整用户流程
- 测试模式切换和操作批准

## 10. 部署方案

### 10.1 构建

```bash
# 构建
go build -o agent-tui ./cmd/agent

# 交叉编译
GOOS=windows GOARCH=amd64 go build -o agent-tui.exe ./cmd/agent
GOOS=darwin GOARCH=amd64 go build -o agent-tui ./cmd/agent
GOOS=linux GOARCH=amd64 go build -o agent-tui ./cmd/agent
```

### 10.2 安装

```bash
# 本地安装
go install ./cmd/agent

# 或下载预编译二进制
```

## 11. 里程碑

| 阶段 | 时间 | 目标 |
|-----|------|-----|
| Phase 1 | 第 1-2 周 | 项目初始化，基础框架搭建 |
| Phase 2 | 第 3-4 周 | UI 组件开发，布局管理 |
| Phase 3 | 第 5-6 周 | AI 客户端集成，多模型支持 |
| Phase 4 | 第 7-8 周 | 核心功能完成，测试 |
| Phase 5 | 第 9-10 周 | 插件系统，调试工具 |
| Phase 6 | 第 11-12 周 | 发布准备，文档完善 |

## 12. 扩展计划

### 12.1 短期（1-3 个月）

- 基础功能完善
- 性能优化
- 跨平台测试

### 12.2 中期（3-6 个月）

- 插件市场
- 团队协作功能
- 云端同步

### 12.3 长期（6-12 个月）

- 代码审查工具集成
- CI/CD 集成
- 高级调试功能

---

**文档版本**: v1.0  
**创建日期**: 2026-05-22  
**作者**: Agent Designer