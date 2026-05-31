# Go TUI Agent 工具实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 开发一个比 DeepSeek-TUI 更强大的终端 AI 编程智能体工具，支持多模型、代码编辑、调试和插件系统。

**Architecture:** 分层架构（UI层 → 业务逻辑层 → 数据访问层 → AI层），使用 Bubble Tea 构建 TUI，SQLite 存储会话历史，支持 OpenAI/DeepSeek/Anthropic/本地模型。

**Tech Stack:** Go 1.22+, Bubble Tea, Lipgloss, Chroma, SQLite, go-git

---

## Phase 1: 项目初始化与基础框架

### Task 1: 创建项目目录结构

**Files:**
- Create: `agent-tui/cmd/agent/main.go`
- Create: `agent-tui/go.mod`
- Create: `agent-tui/go.sum`

- [ ] **Step 1: 创建项目根目录和 go.mod**

```bash
mkdir -p agent-tui/cmd/agent agent-tui/internal/ui agent-tui/internal/core agent-tui/internal/data agent-tui/internal/ai agent-tui/pkg agent-tui/plugins agent-tui/configs
cd agent-tui
go mod init github.com/example/agent-tui
```

- [ ] **Step 2: 添加依赖**

```bash
go get github.com/charmbracelet/bubbletea/v2
go get github.com/charmbracelet/lipgloss
go get github.com/alecthomas/chroma/v2
go get github.com/mattn/go-sqlite3
go get github.com/go-git/go-git/v5
go get github.com/google/uuid
```

- [ ] **Step 3: 创建主入口文件**

```go
// cmd/agent/main.go
package main

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/example/agent-tui/internal/ui"
)

func main() {
    p := tea.NewProgram(ui.NewModel())
    if _, err := p.Run(); err != nil {
        panic(err)
    }
}
```

- [ ] **Step 4: 验证项目构建**

```bash
go build -o agent-tui ./cmd/agent
./agent-tui
```
Expected: 显示空白 TUI 界面

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "init: project structure and dependencies"
```

---

### Task 2: 配置管理模块

**Files:**
- Create: `agent-tui/internal/data/config/config.go`
- Create: `agent-tui/internal/data/config/config_test.go`
- Create: `agent-tui/configs/config.example.toml`

- [ ] **Step 1: 编写配置结构和测试**

```go
// internal/data/config/config_test.go
package config

import "testing"

func TestLoadConfig(t *testing.T) {
    cfg, err := LoadConfig("../../configs/config.example.toml")
    if err != nil {
        t.Fatalf("failed to load config: %v", err)
    }
    if cfg.DefaultModel == "" {
        t.Error("DefaultModel should not be empty")
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/data/config/...
```
Expected: FAIL - config file not found

- [ ] **Step 3: 实现配置加载逻辑**

```go
// internal/data/config/config.go
package config

import (
    "os"
    "path/filepath"

    "github.com/BurntSushi/toml"
)

type APIKeys struct {
    OpenAI    string `toml:"openai"`
    DeepSeek  string `toml:"deepseek"`
    Anthropic string `toml:"anthropic"`
}

type Config struct {
    APIKeys       APIKeys `toml:"api_keys"`
    DefaultModel  string  `toml:"default_model"`
    Theme         string  `toml:"theme"`
    ApprovalMode  string  `toml:"approval_mode"`
    MaxSubagents  int     `toml:"max_subagents"`
}

func LoadConfig(path string) (*Config, error) {
    var cfg Config
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    if err := toml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}

func GetConfigPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".agent-tui", "config.toml")
}
```

- [ ] **Step 4: 创建配置示例文件**

```toml
# configs/config.example.toml
[api_keys]
openai = ""
deepseek = ""
anthropic = ""

default_model = "gpt-4"
theme = "dark"
approval_mode = "manual"
max_subagents = 3
```

- [ ] **Step 5: 运行测试验证通过**

```bash
go test -v ./internal/data/config/...
```
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/data/config configs
git commit -m "feat: config management module"
```

---

### Task 3: 会话历史存储模块

**Files:**
- Create: `agent-tui/internal/data/history/history.go`
- Create: `agent-tui/internal/data/history/history_test.go`

- [ ] **Step 1: 编写会话存储测试**

```go
// internal/data/history/history_test.go
package history

import "testing"

func TestCreateSession(t *testing.T) {
    h := NewHistory(":memory:")
    id := h.CreateSession("test-session")
    if id == "" {
        t.Error("session ID should not be empty")
    }
}

func TestAddMessage(t *testing.T) {
    h := NewHistory(":memory:")
    id := h.CreateSession("test")
    err := h.AddMessage(id, "user", "hello")
    if err != nil {
        t.Fatalf("failed to add message: %v", err)
    }
    msgs := h.GetMessages(id)
    if len(msgs) != 1 {
        t.Errorf("expected 1 message, got %d", len(msgs))
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/data/history/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现会话历史存储**

```go
// internal/data/history/history.go
package history

import (
    "database/sql"
    "time"

    _ "github.com/mattn/go-sqlite3"
    "github.com/google/uuid"
)

type Message struct {
    ID        string
    SessionID string
    Role      string
    Content   string
    Timestamp time.Time
}

type History struct {
    db *sql.DB
}

func NewHistory(path string) *History {
    db, _ := sql.Open("sqlite3", path)
    initDB(db)
    return &History{db: db}
}

func initDB(db *sql.DB) {
    db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
        id TEXT PRIMARY KEY,
        name TEXT,
        created_at TIMESTAMP,
        updated_at TIMESTAMP
    )`)
    db.Exec(`CREATE TABLE IF NOT EXISTS messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        session_id TEXT,
        role TEXT,
        content TEXT,
        timestamp TIMESTAMP,
        FOREIGN KEY (session_id) REFERENCES sessions(id)
    )`)
}

func (h *History) CreateSession(name string) string {
    id := uuid.New().String()
    now := time.Now()
    h.db.Exec("INSERT INTO sessions (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)",
        id, name, now, now)
    return id
}

func (h *History) AddMessage(sessionID, role, content string) error {
    _, err := h.db.Exec("INSERT INTO messages (session_id, role, content, timestamp) VALUES (?, ?, ?, ?)",
        sessionID, role, content, time.Now())
    return err
}

func (h *History) GetMessages(sessionID string) []Message {
    rows, _ := h.db.Query("SELECT id, role, content, timestamp FROM messages WHERE session_id = ? ORDER BY timestamp", sessionID)
    defer rows.Close()
    
    var msgs []Message
    for rows.Next() {
        var m Message
        m.SessionID = sessionID
        rows.Scan(&m.ID, &m.Role, &m.Content, &m.Timestamp)
        msgs = append(msgs, m)
    }
    return msgs
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/data/history/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/data/history
git commit -m "feat: session history storage module"
```

---

## Phase 2: UI 组件开发

### Task 4: 基础布局管理器

**Files:**
- Create: `agent-tui/internal/ui/layout/layout.go`
- Create: `agent-tui/internal/ui/layout/layout_test.go`

- [ ] **Step 1: 编写布局测试**

```go
// internal/ui/layout/layout_test.go
package layout

import "testing"

func TestLayoutCreation(t *testing.T) {
    l := NewLayout()
    if l == nil {
        t.Error("layout should not be nil")
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/ui/layout/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现布局管理器**

```go
// internal/ui/layout/layout.go
package layout

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type Layout struct {
    width, height int
    sidebarWidth  int
}

func NewLayout() *Layout {
    return &Layout{sidebarWidth: 30}
}

func (l *Layout) SetSize(width, height int) {
    l.width = width
    l.height = height
}

func (l *Layout) GetSidebarStyle() lipgloss.Style {
    return lipgloss.NewStyle().
        Width(l.sidebarWidth).
        Height(l.height).
        Border(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("#626262"))
}

func (l *Layout) GetMainStyle() lipgloss.Style {
    return lipgloss.NewStyle().
        Width(l.width - l.sidebarWidth - 2).
        Height(l.height).
        Border(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("#626262"))
}

func (l *Layout) GetStatusBarStyle() lipgloss.Style {
    return lipgloss.NewStyle().
        Width(l.width).
        Height(1).
        Background(lipgloss.Color("#333333"))
}

func (l *Layout) Update(msg tea.Msg) *Layout {
    if msg, ok := msg.(tea.WindowSizeMsg); ok {
        l.SetSize(msg.Width, msg.Height-1)
    }
    return l
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/ui/layout/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/layout
git commit -m "feat: layout manager component"
```

---

### Task 5: 文件树组件

**Files:**
- Create: `agent-tui/internal/ui/filetree/filetree.go`
- Create: `agent-tui/internal/ui/filetree/filetree_test.go`

- [ ] **Step 1: 编写文件树测试**

```go
// internal/ui/filetree/filetree_test.go
package filetree

import "testing"

func TestLoadDirectory(t *testing.T) {
    ft := NewFileTree()
    err := ft.LoadDirectory(".")
    if err != nil {
        t.Fatalf("failed to load directory: %v", err)
    }
    if len(ft.Nodes) == 0 {
        t.Error("directory should have files")
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/ui/filetree/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现文件树组件**

```go
// internal/ui/filetree/filetree.go
package filetree

import (
    "os"
    "path/filepath"
)

type FileNode struct {
    Name     string
    Path     string
    IsDir    bool
    Children []*FileNode
    Expanded bool
}

type FileTree struct {
    Root     *FileNode
    Nodes    []*FileNode
    Selected int
}

func NewFileTree() *FileTree {
    return &FileTree{Selected: 0}
}

func (ft *FileTree) LoadDirectory(path string) error {
    entries, err := os.ReadDir(path)
    if err != nil {
        return err
    }
    
    ft.Nodes = []*FileNode{}
    for _, entry := range entries {
        node := &FileNode{
            Name:  entry.Name(),
            Path:  filepath.Join(path, entry.Name()),
            IsDir: entry.IsDir(),
        }
        ft.Nodes = append(ft.Nodes, node)
    }
    return nil
}

func (ft *FileTree) SelectNext() {
    if ft.Selected < len(ft.Nodes)-1 {
        ft.Selected++
    }
}

func (ft *FileTree) SelectPrev() {
    if ft.Selected > 0 {
        ft.Selected--
    }
}

func (ft *FileTree) ToggleExpand(index int) {
    if index >= 0 && index < len(ft.Nodes) {
        ft.Nodes[index].Expanded = !ft.Nodes[index].Expanded
    }
}

func (ft *FileTree) GetSelectedFile() string {
    if ft.Selected >= 0 && ft.Selected < len(ft.Nodes) {
        return ft.Nodes[ft.Selected].Path
    }
    return ""
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/ui/filetree/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/filetree
git commit -m "feat: file tree component"
```

---

### Task 6: 聊天面板组件

**Files:**
- Create: `agent-tui/internal/ui/chat/chat.go`
- Create: `agent-tui/internal/ui/chat/chat_test.go`

- [ ] **Step 1: 编写聊天面板测试**

```go
// internal/ui/chat/chat_test.go
package chat

import "testing"

func TestAddMessage(t *testing.T) {
    cp := NewChatPanel()
    cp.AddMessage("user", "hello")
    
    msgs := cp.GetMessages()
    if len(msgs) != 1 {
        t.Errorf("expected 1 message, got %d", len(msgs))
    }
    if msgs[0].Content != "hello" {
        t.Errorf("expected 'hello', got '%s'", msgs[0].Content)
    }
}

func TestAppendToLastMessage(t *testing.T) {
    cp := NewChatPanel()
    cp.AddMessage("assistant", "hello")
    cp.AppendToLastMessage(" world")
    
    msgs := cp.GetMessages()
    if msgs[0].Content != "hello world" {
        t.Errorf("expected 'hello world', got '%s'", msgs[0].Content)
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/ui/chat/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现聊天面板组件**

```go
// internal/ui/chat/chat.go
package chat

import "github.com/charmbracelet/bubbletea"

type Message struct {
    Role    string
    Content string
}

type ChatPanel struct {
    Messages []Message
    Input    string
}

func NewChatPanel() *ChatPanel {
    return &ChatPanel{Messages: []Message{}}
}

func (cp *ChatPanel) AddMessage(role, content string) {
    cp.Messages = append(cp.Messages, Message{Role: role, Content: content})
}

func (cp *ChatPanel) AppendToLastMessage(content string) {
    if len(cp.Messages) > 0 {
        cp.Messages[len(cp.Messages)-1].Content += content
    }
}

func (cp *ChatPanel) Clear() {
    cp.Messages = []Message{}
}

func (cp *ChatPanel) GetMessages() []Message {
    return cp.Messages
}

func (cp *ChatPanel) SetMessages(messages []Message) {
    cp.Messages = messages
}

func (cp *ChatPanel) Update(msg tea.Msg) (*ChatPanel, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "enter":
            if cp.Input != "" {
                cp.AddMessage("user", cp.Input)
                cp.Input = ""
            }
        case "backspace":
            if len(cp.Input) > 0 {
                cp.Input = cp.Input[:len(cp.Input)-1]
            }
        default:
            cp.Input += msg.String()
        }
    }
    return cp, nil
}

func (cp *ChatPanel) View() string {
    view := ""
    for _, msg := range cp.Messages {
        prefix := "[User] "
        if msg.Role == "assistant" {
            prefix = "[AI] "
        }
        view += prefix + msg.Content + "\n"
    }
    view += "\n> " + cp.Input
    return view
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/ui/chat/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/chat
git commit -m "feat: chat panel component"
```

---

### Task 7: 状态栏组件

**Files:**
- Create: `agent-tui/internal/ui/status/status.go`
- Create: `agent-tui/internal/ui/status/status_test.go`

- [ ] **Step 1: 编写状态栏测试**

```go
// internal/ui/status/status_test.go
package status

import "testing"

func TestStatusBarUpdate(t *testing.T) {
    sb := NewStatusBar()
    sb.SetModel("GPT-4")
    sb.SetMode("Agent")
    
    if sb.Model != "GPT-4" {
        t.Errorf("expected model 'GPT-4', got '%s'", sb.Model)
    }
    if sb.Mode != "Agent" {
        t.Errorf("expected mode 'Agent', got '%s'", sb.Mode)
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/ui/status/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现状态栏组件**

```go
// internal/ui/status/status.go
package status

import (
    "fmt"
    "github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
    Model     string
    Mode      string
    Connected bool
    Tasks     int
    Total     int
}

func NewStatusBar() *StatusBar {
    return &StatusBar{
        Model:     "GPT-4",
        Mode:      "Normal",
        Connected: true,
        Tasks:     0,
        Total:     0,
    }
}

func (sb *StatusBar) SetModel(model string) {
    sb.Model = model
}

func (sb *StatusBar) SetMode(mode string) {
    sb.Mode = mode
}

func (sb *StatusBar) SetConnected(connected bool) {
    sb.Connected = connected
}

func (sb *StatusBar) SetTasks(current, total int) {
    sb.Tasks = current
    sb.Total = total
}

func (sb *StatusBar) View() string {
    status := "Disconnected"
    if sb.Connected {
        status = "Connected"
    }
    
    statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
    modelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff00"))
    modeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffff"))
    
    return fmt.Sprintf(
        "%s | %s | %s | Tasks: %d/%d",
        statusStyle.Render(status),
        modelStyle.Render("Model: "+sb.Model),
        modeStyle.Render("Mode: "+sb.Mode),
        sb.Tasks, sb.Total,
    )
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/ui/status/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/status
git commit -m "feat: status bar component"
```

---

### Task 8: 代码编辑器组件

**Files:**
- Create: `agent-tui/internal/ui/editor/editor.go`
- Create: `agent-tui/internal/ui/editor/editor_test.go`

- [ ] **Step 1: 编写编辑器测试**

```go
// internal/ui/editor/editor_test.go
package editor

import "testing"

func TestLoadFile(t *testing.T) {
    e := NewEditor()
    err := e.LoadFile("test.txt")
    // File doesn't exist, should error
    if err == nil {
        t.Error("expected error for non-existent file")
    }
}

func TestContentOperations(t *testing.T) {
    e := NewEditor()
    e.SetContent("hello world")
    
    if e.GetContent() != "hello world" {
        t.Errorf("expected 'hello world', got '%s'", e.GetContent())
    }
    
    e.InsertAtCursor("test")
    if e.GetContent() != "testhello world" {
        t.Errorf("expected 'testhello world', got '%s'", e.GetContent())
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/ui/editor/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现代码编辑器组件**

```go
// internal/ui/editor/editor.go
package editor

import (
    "os"

    "github.com/alecthomas/chroma/v2/quick"
)

type Editor struct {
    Content       string
    CursorRow     int
    CursorCol     int
    FilePath      string
    Language      string
}

func NewEditor() *Editor {
    return &Editor{
        Content:   "",
        CursorRow: 0,
        CursorCol: 0,
    }
}

func (e *Editor) LoadFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    e.Content = string(data)
    e.FilePath = path
    return nil
}

func (e *Editor) SaveFile(path string) error {
    return os.WriteFile(path, []byte(e.Content), 0644)
}

func (e *Editor) GetContent() string {
    return e.Content
}

func (e *Editor) SetContent(content string) {
    e.Content = content
}

func (e *Editor) InsertAtCursor(text string) {
    lines := []rune(e.Content)
    pos := e.getCursorPosition()
    if pos > len(lines) {
        pos = len(lines)
    }
    e.Content = string(lines[:pos]) + text + string(lines[pos:])
    e.CursorCol += len(text)
}

func (e *Editor) DeleteSelection() {
    e.Content = ""
}

func (e *Editor) GetCursorPosition() (int, int) {
    return e.CursorRow, e.CursorCol
}

func (e *Editor) SetCursorPosition(row, col int) {
    e.CursorRow = row
    e.CursorCol = col
}

func (e *Editor) getCursorPosition() int {
    lines := []rune(e.Content)
    pos := 0
    currentRow := 0
    
    for i, ch := range lines {
        if currentRow == e.CursorRow {
            pos = i + e.CursorCol
            break
        }
        if ch == '\n' {
            currentRow++
        }
    }
    return pos
}

func (e *Editor) Highlight() string {
    var buf bytes.Buffer
    quick.Highlight(&buf, e.Content, e.Language, "terminal256", "monokai")
    return buf.String()
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/ui/editor/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/editor
git commit -m "feat: code editor component"
```

---

### Task 9: 主 UI 模型整合

**Files:**
- Create: `agent-tui/internal/ui/ui.go`
- Modify: `agent-tui/cmd/agent/main.go`

- [ ] **Step 1: 创建主 UI 模型**

```go
// internal/ui/ui.go
package ui

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/example/agent-tui/internal/ui/chat"
    "github.com/example/agent-tui/internal/ui/editor"
    "github.com/example/agent-tui/internal/ui/filetree"
    "github.com/example/agent-tui/internal/ui/layout"
    "github.com/example/agent-tui/internal/ui/status"
)

type Model struct {
    Layout     *layout.Layout
    FileTree   *filetree.FileTree
    ChatPanel  *chat.ChatPanel
    Editor     *editor.Editor
    StatusBar  *status.StatusBar
    ActivePane string
}

func NewModel() Model {
    return Model{
        Layout:     layout.NewLayout(),
        FileTree:   filetree.NewFileTree(),
        ChatPanel:  chat.NewChatPanel(),
        Editor:     editor.NewEditor(),
        StatusBar:  status.NewStatusBar(),
        ActivePane: "chat",
    }
}

func (m Model) Init() tea.Cmd {
    m.FileTree.LoadDirectory(".")
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.Layout.SetSize(msg.Width, msg.Height-1)
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+tab":
            m.switchPane()
        case "q":
            return m, tea.Quit
        }
    }
    
    var cmd tea.Cmd
    m.ChatPanel, cmd = m.ChatPanel.Update(msg)
    return m, cmd
}

func (m *Model) switchPane() {
    panes := []string{"chat", "editor", "filetree"}
    for i, pane := range panes {
        if pane == m.ActivePane {
            m.ActivePane = panes[(i+1)%len(panes)]
            break
        }
    }
}

func (m Model) View() string {
    sidebar := m.Layout.GetSidebarStyle().Render(m.FileTree.View())
    
    var mainContent string
    if m.ActivePane == "chat" {
        mainContent = m.ChatPanel.View()
    } else if m.ActivePane == "editor" {
        mainContent = m.Editor.GetContent()
    }
    
    main := m.Layout.GetMainStyle().Render(mainContent)
    statusBar := m.StatusBar.View()
    
    return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main) + "\n" + statusBar
}
```

- [ ] **Step 2: 更新主入口文件**

```go
// cmd/agent/main.go
package main

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/example/agent-tui/internal/ui"
)

func main() {
    p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        panic(err)
    }
}
```

- [ ] **Step 3: 构建并测试**

```bash
go build -o agent-tui ./cmd/agent
./agent-tui
```
Expected: 显示 TUI 界面，包含文件树、聊天面板和状态栏

- [ ] **Step 4: Commit**

```bash
git add internal/ui ui.go cmd/agent/main.go
git commit -m "feat: main UI model integration"
```

---

## Phase 3: AI 客户端集成

### Task 10: AI 客户端接口定义

**Files:**
- Create: `agent-tui/internal/ai/client.go`
- Create: `agent-tui/internal/ai/client_test.go`

- [ ] **Step 1: 编写接口测试**

```go
// internal/ai/client_test.go
package ai

import "testing"

func TestClientInterface(t *testing.T) {
    // Test that interface is properly defined
    var _ AIClient = &mockClient{}
}

type mockClient struct{}

func (m *mockClient) Completion(req CompletionRequest) (<-chan CompletionResponse, error) {
    return nil, nil
}

func (m *mockClient) GetModels() []ModelInfo {
    return nil
}

func (m *mockClient) ValidateAPIKey() error {
    return nil
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/ai/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现 AI 客户端接口**

```go
// internal/ai/client.go
package ai

import "context"

type ModelInfo struct {
    ID     string
    Name   string
    MaxTokens int
}

type CompletionRequest struct {
    Model    string
    Messages []Message
    Stream   bool
}

type Message struct {
    Role    string
    Content string
}

type CompletionResponse struct {
    Content string
    Done    bool
    Error   error
}

type AIClient interface {
    Completion(ctx context.Context, req CompletionRequest) (<-chan CompletionResponse, error)
    GetModels() []ModelInfo
    ValidateAPIKey() error
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/ai/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ai/client.go internal/ai/client_test.go
git commit -m "feat: AI client interface"
```

---

### Task 11: OpenAI 客户端实现

**Files:**
- Create: `agent-tui/internal/ai/openai/openai.go`
- Create: `agent-tui/internal/ai/openai/openai_test.go`

- [ ] **Step 1: 编写 OpenAI 客户端测试**

```go
// internal/ai/openai/openai_test.go
package openai

import "testing"

func TestNewClient(t *testing.T) {
    c := NewClient("test-api-key")
    if c == nil {
        t.Error("client should not be nil")
    }
}

func TestGetModels(t *testing.T) {
    c := NewClient("test-api-key")
    models := c.GetModels()
    if len(models) == 0 {
        t.Error("should return default models")
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/ai/openai/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现 OpenAI 客户端**

```go
// internal/ai/openai/openai.go
package openai

import (
    "context"
    "encoding/json"
    "io"
    "net/http"
    "strings"
    
    "github.com/example/agent-tui/internal/ai"
)

type Client struct {
    apiKey string
    baseURL string
}

func NewClient(apiKey string) *Client {
    return &Client{
        apiKey: apiKey,
        baseURL: "https://api.openai.com/v1",
    }
}

func (c *Client) GetModels() []ai.ModelInfo {
    return []ai.ModelInfo{
        {ID: "gpt-4", Name: "GPT-4", MaxTokens: 8192},
        {ID: "gpt-4o", Name: "GPT-4 Omni", MaxTokens: 128000},
        {ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", MaxTokens: 16384},
    }
}

func (c *Client) ValidateAPIKey() error {
    return nil // In real implementation, make API call
}

func (c *Client) Completion(ctx context.Context, req ai.CompletionRequest) (<-chan ai.CompletionResponse, error) {
    ch := make(chan ai.CompletionResponse)
    
    go func() {
        defer close(ch)
        
        messages := make([]map[string]string, len(req.Messages))
        for i, msg := range req.Messages {
            messages[i] = map[string]string{"role": msg.Role, "content": msg.Content}
        }
        
        payload := map[string]interface{}{
            "model":    req.Model,
            "messages": messages,
            "stream":   req.Stream,
        }
        
        jsonData, _ := json.Marshal(payload)
        httpReq, _ := http.NewRequest("POST", c.baseURL+"/chat/completions", strings.NewReader(string(jsonData)))
        httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
        httpReq.Header.Set("Content-Type", "application/json")
        
        client := &http.Client{}
        resp, err := client.Do(httpReq)
        if err != nil {
            ch <- ai.CompletionResponse{Error: err}
            return
        }
        defer resp.Body.Close()
        
        if req.Stream {
            c.handleStream(resp.Body, ch)
        } else {
            c.handleNonStream(resp.Body, ch)
        }
    }()
    
    return ch, nil
}

func (c *Client) handleStream(body io.ReadCloser, ch chan ai.CompletionResponse) {
    buf := make([]byte, 1024)
    for {
        n, err := body.Read(buf)
        if err != nil {
            break
        }
        ch <- ai.CompletionResponse{Content: string(buf[:n]), Done: false}
    }
    ch <- ai.CompletionResponse{Done: true}
}

func (c *Client) handleNonStream(body io.ReadCloser, ch chan ai.CompletionResponse) {
    data, _ := io.ReadAll(body)
    ch <- ai.CompletionResponse{Content: string(data), Done: true}
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/ai/openai/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ai/openai
git commit -m "feat: OpenAI client implementation"
```

---

### Task 12: DeepSeek 客户端实现

**Files:**
- Create: `agent-tui/internal/ai/deepseek/deepseek.go`
- Create: `agent-tui/internal/ai/deepseek/deepseek_test.go`

- [ ] **Step 1: 编写 DeepSeek 客户端测试**

```go
// internal/ai/deepseek/deepseek_test.go
package deepseek

import "testing"

func TestNewClient(t *testing.T) {
    c := NewClient("test-api-key")
    if c == nil {
        t.Error("client should not be nil")
    }
}

func TestGetModels(t *testing.T) {
    c := NewClient("test-api-key")
    models := c.GetModels()
    if len(models) == 0 {
        t.Error("should return default models")
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/ai/deepseek/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现 DeepSeek 客户端**

```go
// internal/ai/deepseek/deepseek.go
package deepseek

import (
    "context"
    "encoding/json"
    "io"
    "net/http"
    "strings"
    
    "github.com/example/agent-tui/internal/ai"
)

type Client struct {
    apiKey string
    baseURL string
}

func NewClient(apiKey string) *Client {
    return &Client{
        apiKey: apiKey,
        baseURL: "https://api.deepseek.com/v1",
    }
}

func (c *Client) GetModels() []ai.ModelInfo {
    return []ai.ModelInfo{
        {ID: "deepseek-reasoner", Name: "DeepSeek Reasoner", MaxTokens: 1048576},
        {ID: "deepseek-chat", Name: "DeepSeek Chat", MaxTokens: 1048576},
        {ID: "deepseek-coder", Name: "DeepSeek Coder", MaxTokens: 1048576},
    }
}

func (c *Client) ValidateAPIKey() error {
    return nil
}

func (c *Client) Completion(ctx context.Context, req ai.CompletionRequest) (<-chan ai.CompletionResponse, error) {
    ch := make(chan ai.CompletionResponse)
    
    go func() {
        defer close(ch)
        
        messages := make([]map[string]string, len(req.Messages))
        for i, msg := range req.Messages {
            messages[i] = map[string]string{"role": msg.Role, "content": msg.Content}
        }
        
        payload := map[string]interface{}{
            "model":    req.Model,
            "messages": messages,
            "stream":   req.Stream,
        }
        
        jsonData, _ := json.Marshal(payload)
        httpReq, _ := http.NewRequest("POST", c.baseURL+"/chat/completions", strings.NewReader(string(jsonData)))
        httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
        httpReq.Header.Set("Content-Type", "application/json")
        
        client := &http.Client{}
        resp, err := client.Do(httpReq)
        if err != nil {
            ch <- ai.CompletionResponse{Error: err}
            return
        }
        defer resp.Body.Close()
        
        if req.Stream {
            c.handleStream(resp.Body, ch)
        } else {
            c.handleNonStream(resp.Body, ch)
        }
    }()
    
    return ch, nil
}

func (c *Client) handleStream(body io.ReadCloser, ch chan ai.CompletionResponse) {
    buf := make([]byte, 1024)
    for {
        n, err := body.Read(buf)
        if err != nil {
            break
        }
        ch <- ai.CompletionResponse{Content: string(buf[:n]), Done: false}
    }
    ch <- ai.CompletionResponse{Done: true}
}

func (c *Client) handleNonStream(body io.ReadCloser, ch chan ai.CompletionResponse) {
    data, _ := io.ReadAll(body)
    ch <- ai.CompletionResponse{Content: string(data), Done: true}
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/ai/deepseek/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ai/deepseek
git commit -m "feat: DeepSeek client implementation"
```

---

### Task 13: Anthropic 客户端实现

**Files:**
- Create: `agent-tui/internal/ai/anthropic/anthropic.go`
- Create: `agent-tui/internal/ai/anthropic/anthropic_test.go`

- [ ] **Step 1: 编写 Anthropic 客户端测试**

```go
// internal/ai/anthropic/anthropic_test.go
package anthropic

import "testing"

func TestNewClient(t *testing.T) {
    c := NewClient("test-api-key")
    if c == nil {
        t.Error("client should not be nil")
    }
}

func TestGetModels(t *testing.T) {
    c := NewClient("test-api-key")
    models := c.GetModels()
    if len(models) == 0 {
        t.Error("should return default models")
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/ai/anthropic/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现 Anthropic 客户端**

```go
// internal/ai/anthropic/anthropic.go
package anthropic

import (
    "context"
    "encoding/json"
    "io"
    "net/http"
    "strings"
    
    "github.com/example/agent-tui/internal/ai"
)

type Client struct {
    apiKey string
    baseURL string
}

func NewClient(apiKey string) *Client {
    return &Client{
        apiKey: apiKey,
        baseURL: "https://api.anthropic.com/v1",
    }
}

func (c *Client) GetModels() []ai.ModelInfo {
    return []ai.ModelInfo{
        {ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", MaxTokens: 200000},
        {ID: "claude-3-sonnet-20240229", Name: "Claude 3 Sonnet", MaxTokens: 200000},
        {ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku", MaxTokens: 200000},
    }
}

func (c *Client) ValidateAPIKey() error {
    return nil
}

func (c *Client) Completion(ctx context.Context, req ai.CompletionRequest) (<-chan ai.CompletionResponse, error) {
    ch := make(chan ai.CompletionResponse)
    
    go func() {
        defer close(ch)
        
        messages := make([]map[string]string, len(req.Messages))
        for i, msg := range req.Messages {
            messages[i] = map[string]string{"role": msg.Role, "content": msg.Content}
        }
        
        payload := map[string]interface{}{
            "model":    req.Model,
            "messages": messages,
            "stream":   req.Stream,
        }
        
        jsonData, _ := json.Marshal(payload)
        httpReq, _ := http.NewRequest("POST", c.baseURL+"/messages", strings.NewReader(string(jsonData)))
        httpReq.Header.Set("x-api-key", c.apiKey)
        httpReq.Header.Set("Content-Type", "application/json")
        httpReq.Header.Set("anthropic-version", "2023-06-01")
        
        client := &http.Client{}
        resp, err := client.Do(httpReq)
        if err != nil {
            ch <- ai.CompletionResponse{Error: err}
            return
        }
        defer resp.Body.Close()
        
        if req.Stream {
            c.handleStream(resp.Body, ch)
        } else {
            c.handleNonStream(resp.Body, ch)
        }
    }()
    
    return ch, nil
}

func (c *Client) handleStream(body io.ReadCloser, ch chan ai.CompletionResponse) {
    buf := make([]byte, 1024)
    for {
        n, err := body.Read(buf)
        if err != nil {
            break
        }
        ch <- ai.CompletionResponse{Content: string(buf[:n]), Done: false}
    }
    ch <- ai.CompletionResponse{Done: true}
}

func (c *Client) handleNonStream(body io.ReadCloser, ch chan ai.CompletionResponse) {
    data, _ := io.ReadAll(body)
    ch <- ai.CompletionResponse{Content: string(data), Done: true}
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/ai/anthropic/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ai/anthropic
git commit -m "feat: Anthropic client implementation"
```

---

### Task 14: 本地模型客户端实现

**Files:**
- Create: `agent-tui/internal/ai/local/local.go`
- Create: `agent-tui/internal/ai/local/local_test.go`

- [ ] **Step 1: 编写本地模型客户端测试**

```go
// internal/ai/local/local_test.go
package local

import "testing"

func TestNewClient(t *testing.T) {
    c := NewClient("http://localhost:11434")
    if c == nil {
        t.Error("client should not be nil")
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/ai/local/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现本地模型客户端**

```go
// internal/ai/local/local.go
package local

import (
    "context"
    "encoding/json"
    "io"
    "net/http"
    "strings"
    
    "github.com/example/agent-tui/internal/ai"
)

type Client struct {
    baseURL string
}

func NewClient(baseURL string) *Client {
    return &Client{baseURL: baseURL}
}

func (c *Client) GetModels() []ai.ModelInfo {
    return []ai.ModelInfo{
        {ID: "llama3", Name: "Llama 3", MaxTokens: 8192},
        {ID: "mistral", Name: "Mistral", MaxTokens: 8192},
        {ID: "phi3", Name: "Phi-3", MaxTokens: 4096},
    }
}

func (c *Client) ValidateAPIKey() error {
    return nil
}

func (c *Client) Completion(ctx context.Context, req ai.CompletionRequest) (<-chan ai.CompletionResponse, error) {
    ch := make(chan ai.CompletionResponse)
    
    go func() {
        defer close(ch)
        
        messages := make([]map[string]string, len(req.Messages))
        for i, msg := range req.Messages {
            messages[i] = map[string]string{"role": msg.Role, "content": msg.Content}
        }
        
        payload := map[string]interface{}{
            "model":    req.Model,
            "messages": messages,
            "stream":   req.Stream,
        }
        
        jsonData, _ := json.Marshal(payload)
        httpReq, _ := http.NewRequest("POST", c.baseURL+"/v1/chat/completions", strings.NewReader(string(jsonData)))
        httpReq.Header.Set("Content-Type", "application/json")
        
        client := &http.Client{}
        resp, err := client.Do(httpReq)
        if err != nil {
            ch <- ai.CompletionResponse{Error: err}
            return
        }
        defer resp.Body.Close()
        
        if req.Stream {
            c.handleStream(resp.Body, ch)
        } else {
            c.handleNonStream(resp.Body, ch)
        }
    }()
    
    return ch, nil
}

func (c *Client) handleStream(body io.ReadCloser, ch chan ai.CompletionResponse) {
    buf := make([]byte, 1024)
    for {
        n, err := body.Read(buf)
        if err != nil {
            break
        }
        ch <- ai.CompletionResponse{Content: string(buf[:n]), Done: false}
    }
    ch <- ai.CompletionResponse{Done: true}
}

func (c *Client) handleNonStream(body io.ReadCloser, ch chan ai.CompletionResponse) {
    data, _ := io.ReadAll(body)
    ch <- ai.CompletionResponse{Content: string(data), Done: true}
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/ai/local/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ai/local
git commit -m "feat: local model client implementation"
```

---

## Phase 4: 业务逻辑层整合

### Task 15: AI 助手服务

**Files:**
- Create: `agent-tui/internal/core/assistant/assistant.go`
- Create: `agent-tui/internal/core/assistant/assistant_test.go`

- [ ] **Step 1: 编写 AI 助手测试**

```go
// internal/core/assistant/assistant_test.go
package assistant

import "testing"

func TestCreateSession(t *testing.T) {
    a := NewAIAssistant(nil)
    id := a.CreateSession()
    if id == "" {
        t.Error("session ID should not be empty")
    }
}

func TestSetActiveModel(t *testing.T) {
    a := NewAIAssistant(nil)
    a.SetActiveModel("gpt-4")
    if a.activeModel != "gpt-4" {
        t.Errorf("expected model 'gpt-4', got '%s'", a.activeModel)
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/core/assistant/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现 AI 助手服务**

```go
// internal/core/assistant/assistant.go
package assistant

import (
    "context"
    
    "github.com/example/agent-tui/internal/ai"
    "github.com/google/uuid"
)

type AIAssistant struct {
    client      ai.AIClient
    activeModel string
    sessions    map[string][]ai.Message
}

func NewAIAssistant(client ai.AIClient) *AIAssistant {
    return &AIAssistant{
        client:      client,
        activeModel: "gpt-4",
        sessions:    make(map[string][]ai.Message),
    }
}

func (a *AIAssistant) SendMessage(messages []ai.Message, model string) (<-chan string, error) {
    req := ai.CompletionRequest{
        Model:    model,
        Messages: messages,
        Stream:   true,
    }
    
    ch, err := a.client.Completion(context.Background(), req)
    if err != nil {
        return nil, err
    }
    
    resultCh := make(chan string)
    go func() {
        defer close(resultCh)
        for resp := range ch {
            if resp.Error != nil {
                return
            }
            resultCh <- resp.Content
        }
    }()
    
    return resultCh, nil
}

func (a *AIAssistant) GetModels() []ai.ModelInfo {
    if a.client != nil {
        return a.client.GetModels()
    }
    return []ai.ModelInfo{}
}

func (a *AIAssistant) SetActiveModel(modelID string) {
    a.activeModel = modelID
}

func (a *AIAssistant) CreateSession() string {
    id := uuid.New().String()
    a.sessions[id] = []ai.Message{}
    return id
}

func (a *AIAssistant) LoadSession(id string) ([]ai.Message, error) {
    msgs, ok := a.sessions[id]
    if !ok {
        return nil, nil
    }
    return msgs, nil
}

func (a *AIAssistant) SaveSession(id string, messages []ai.Message) error {
    a.sessions[id] = messages
    return nil
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/core/assistant/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/core/assistant
git commit -m "feat: AI assistant service"
```

---

### Task 16: 项目管理服务

**Files:**
- Create: `agent-tui/internal/core/project/project.go`
- Create: `agent-tui/internal/core/project/project_test.go`

- [ ] **Step 1: 编写项目管理测试**

```go
// internal/core/project/project_test.go
package project

import "testing"

func TestListFiles(t *testing.T) {
    p := NewProjectManager()
    files, err := p.ListFiles(".")
    if err != nil {
        t.Fatalf("failed to list files: %v", err)
    }
    if len(files) == 0 {
        t.Error("should return files")
    }
}

func TestCreateFile(t *testing.T) {
    p := NewProjectManager()
    err := p.CreateFile("test_create.txt", "test content")
    if err != nil {
        t.Fatalf("failed to create file: %v", err)
    }
    
    files, _ := p.ListFiles(".")
    found := false
    for _, f := range files {
        if f.Name == "test_create.txt" {
            found = true
            break
        }
    }
    if !found {
        t.Error("created file should exist")
    }
    
    p.DeleteFile("test_create.txt")
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/core/project/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现项目管理服务**

```go
// internal/core/project/project.go
package project

import (
    "os"
    "path/filepath"
    
    "github.com/go-git/go-git/v5"
)

type FileInfo struct {
    Name    string
    Path    string
    IsDir   bool
    Size    int64
    ModTime string
}

type ProjectManager struct{}

func NewProjectManager() *ProjectManager {
    return &ProjectManager{}
}

func (p *ProjectManager) ListFiles(path string) ([]FileInfo, error) {
    entries, err := os.ReadDir(path)
    if err != nil {
        return nil, err
    }
    
    var files []FileInfo
    for _, entry := range entries {
        info, _ := entry.Info()
        files = append(files, FileInfo{
            Name:    entry.Name(),
            Path:    filepath.Join(path, entry.Name()),
            IsDir:   entry.IsDir(),
            Size:    info.Size(),
            ModTime: info.ModTime().String(),
        })
    }
    return files, nil
}

func (p *ProjectManager) CreateFile(path string, content string) error {
    return os.WriteFile(path, []byte(content), 0644)
}

func (p *ProjectManager) DeleteFile(path string) error {
    return os.Remove(path)
}

func (p *ProjectManager) RenameFile(oldPath, newPath string) error {
    return os.Rename(oldPath, newPath)
}

func (p *ProjectManager) GitClone(url, dest string) error {
    _, err := git.PlainClone(dest, false, &git.CloneOptions{
        URL: url,
    })
    return err
}

func (p *ProjectManager) GitCommit(message string) error {
    return nil
}

func (p *ProjectManager) GitPush() error {
    return nil
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/core/project/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/core/project
git commit -m "feat: project manager service"
```

---

### Task 17: 代码编辑器服务

**Files:**
- Create: `agent-tui/internal/core/editor/editor.go`
- Create: `agent-tui/internal/core/editor/editor_test.go`

- [ ] **Step 1: 编写代码编辑器测试**

```go
// internal/core/editor/editor_test.go
package editor

import "testing"

func TestOpenFile(t *testing.T) {
    e := NewCodeEditorService()
    doc, err := e.OpenFile("test.txt")
    // File doesn't exist
    if err == nil {
        t.Error("expected error for non-existent file")
    }
    _ = doc
}

func TestHighlightSyntax(t *testing.T) {
    e := NewCodeEditorService()
    highlighted := e.HighlightSyntax("func main() {}", "go")
    if highlighted == "" {
        t.Error("highlighted content should not be empty")
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/core/editor/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现代码编辑器服务**

```go
// internal/core/editor/editor.go
package editor

import (
    "os"
    
    "github.com/alecthomas/chroma/v2/quick"
)

type Document struct {
    Path    string
    Content string
    Language string
}

type Completion struct {
    Text string
    Kind string
}

type CodeEditorService struct{}

func NewCodeEditorService() *CodeEditorService {
    return &CodeEditorService{}
}

func (e *CodeEditorService) OpenFile(path string) (*Document, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    return &Document{
        Path:    path,
        Content: string(data),
        Language: getLanguage(path),
    }, nil
}

func (e *CodeEditorService) SaveFile(doc *Document) error {
    return os.WriteFile(doc.Path, []byte(doc.Content), 0644)
}

func (e *CodeEditorService) HighlightSyntax(content string, language string) string {
    if language == "" {
        language = "text"
    }
    var buf bytes.Buffer
    quick.Highlight(&buf, content, language, "terminal256", "monokai")
    return buf.String()
}

func (e *CodeEditorService) GetCompletions(content string, position Position) []Completion {
    return []Completion{}
}

func getLanguage(path string) string {
    ext := filepath.Ext(path)
    switch ext {
    case ".go":
        return "go"
    case ".py":
        return "python"
    case ".js":
        return "javascript"
    case ".ts":
        return "typescript"
    case ".json":
        return "json"
    case ".md":
        return "markdown"
    default:
        return "text"
    }
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/core/editor/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/core/editor
git commit -m "feat: code editor service"
```

---

### Task 18: 调试器服务

**Files:**
- Create: `agent-tui/internal/core/debugger/debugger.go`
- Create: `agent-tui/internal/core/debugger/debugger_test.go`

- [ ] **Step 1: 编写调试器测试**

```go
// internal/core/debugger/debugger_test.go
package debugger

import "testing"

func TestSetBreakpoint(t *testing.T) {
    d := NewDebugger()
    err := d.SetBreakpoint("test.go", 10)
    if err != nil {
        t.Fatalf("failed to set breakpoint: %v", err)
    }
    
    breakpoints := d.GetBreakpoints()
    if len(breakpoints) != 1 {
        t.Errorf("expected 1 breakpoint, got %d", len(breakpoints))
    }
}

func TestRemoveBreakpoint(t *testing.T) {
    d := NewDebugger()
    d.SetBreakpoint("test.go", 10)
    err := d.RemoveBreakpoint("test.go", 10)
    if err != nil {
        t.Fatalf("failed to remove breakpoint: %v", err)
    }
    
    breakpoints := d.GetBreakpoints()
    if len(breakpoints) != 0 {
        t.Errorf("expected 0 breakpoints, got %d", len(breakpoints))
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/core/debugger/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现调试器服务**

```go
// internal/core/debugger/debugger.go
package debugger

import "time"

type Breakpoint struct {
    Path string
    Line int
}

type LogEntry struct {
    Timestamp time.Time
    Level     string
    Message   string
}

type Debugger struct {
    breakpoints []Breakpoint
    logs        []LogEntry
    isRunning   bool
}

func NewDebugger() *Debugger {
    return &Debugger{
        breakpoints: []Breakpoint{},
        logs:        []LogEntry{},
        isRunning:   false,
    }
}

func (d *Debugger) SetBreakpoint(path string, line int) error {
    for _, bp := range d.breakpoints {
        if bp.Path == path && bp.Line == line {
            return nil
        }
    }
    d.breakpoints = append(d.breakpoints, Breakpoint{Path: path, Line: line})
    return nil
}

func (d *Debugger) RemoveBreakpoint(path string, line int) error {
    for i, bp := range d.breakpoints {
        if bp.Path == path && bp.Line == line {
            d.breakpoints = append(d.breakpoints[:i], d.breakpoints[i+1:]...)
            return nil
        }
    }
    return nil
}

func (d *Debugger) GetBreakpoints() []Breakpoint {
    return d.breakpoints
}

func (d *Debugger) StartDebugging() error {
    d.isRunning = true
    d.logs = append(d.logs, LogEntry{
        Timestamp: time.Now(),
        Level:     "INFO",
        Message:   "Debugging started",
    })
    return nil
}

func (d *Debugger) StopDebugging() error {
    d.isRunning = false
    d.logs = append(d.logs, LogEntry{
        Timestamp: time.Now(),
        Level:     "INFO",
        Message:   "Debugging stopped",
    })
    return nil
}

func (d *Debugger) StepOver() error {
    if !d.isRunning {
        return nil
    }
    d.logs = append(d.logs, LogEntry{
        Timestamp: time.Now(),
        Level:     "DEBUG",
        Message:   "Step over",
    })
    return nil
}

func (d *Debugger) StepInto() error {
    if !d.isRunning {
        return nil
    }
    d.logs = append(d.logs, LogEntry{
        Timestamp: time.Now(),
        Level:     "DEBUG",
        Message:   "Step into",
    })
    return nil
}

func (d *Debugger) GetLogs() []LogEntry {
    return d.logs
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -v ./internal/core/debugger/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/core/debugger
git commit -m "feat: debugger service"
```

---

## Phase 5: 插件系统与调试工具

### Task 19: 插件注册表

**Files:**
- Create: `agent-tui/internal/data/plugins/plugins.go`
- Create: `agent-tui/internal/data/plugins/plugins_test.go`

- [ ] **Step 1: 编写插件注册表测试**

```go
// internal/data/plugins/plugins_test.go
package plugins

import "testing"

func TestRegisterPlugin(t *testing.T) {
    r := NewPluginRegistry()
    err := r.RegisterPlugin("test-plugin", "/path/to/plugin")
    if err != nil {
        t.Fatalf("failed to register plugin: %v", err)
    }
    
    plugins := r.GetPlugins()
    if len(plugins) != 1 {
        t.Errorf("expected 1 plugin, got %d", len(plugins))
    }
}

func TestLoadPlugin(t *testing.T) {
    r := NewPluginRegistry()
    r.RegisterPlugin("test-plugin", "/path/to/plugin")
    
    plugin, err := r.LoadPlugin("test-plugin")
    if err != nil {
        t.Fatalf("failed to load plugin: %v", err)
    }
    if plugin == nil {
        t.Error("plugin should not be nil")
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -v ./internal/data/plugins/...
```
Expected: FAIL - type not defined

- [ ] **Step 3: 实现插件注册表**

```go
// internal/data/plugins/plugins.go
package plugins

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Plugin struct {
    Name        string
    Path        string
    Description string
    Version     string
    Enabled     bool
}

type PluginRegistry struct {
    plugins map[string]*Plugin
}

func NewPluginRegistry() *PluginRegistry {
    return &PluginRegistry{
        plugins: make(map[string]*Plugin),
    }
}

func (r *PluginRegistry) RegisterPlugin(name, path string) error {
    plugin := &Plugin{
        Name:        name,
        Path:        path,
        Description: "",
        Version:     "1.0.0",
        Enabled:     true,
    }
    
    configPath := filepath.Join