# TUI 多窗口系统实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现基于 Go 的 TUI 多窗口系统，包括窗口切换、语法高亮、Thinking 折叠等功能

**Architecture:** 使用 Bubble Tea 框架，组件化设计，支持四个窗口（Chat/Editor/Terminal/Logs），底部标签栏切换

**Tech Stack:** Go, Bubble Tea, Lipgloss, Chroma

---

## 文件结构

```
创建的新文件:
- internal/ui/windowmanager/windowmanager.go
- internal/ui/tabbar/tabbar.go
- internal/ui/windows/chat.go
- internal/ui/windows/editor.go
- internal/ui/windows/terminal.go
- internal/ui/windows/logs.go
- internal/ui/syntax/syntax.go

修改的文件:
- internal/ui/ui.go
- internal/ui/theme.go
- cmd/agent/main.go
```

---

## Task 1: 创建窗口管理器组件

**Files:**
- Create: `internal/ui/windowmanager/windowmanager.go`
- Create: `internal/ui/windowmanager/windowmanager_test.go`

- [ ] **Step 1: 创建窗口接口**

```go
package windowmanager

type Window interface {
    ID() string
    Name() string
    View(width, height int) string
    Update(msg interface{})
}
```

- [ ] **Step 2: 实现 WindowManager**

```go
package windowmanager

type WindowManager struct {
    windows      map[string]Window
    activeWindow string
    windowOrder  []string
}

func NewWindowManager() *WindowManager {
    return &WindowManager{
        windows:     make(map[string]Window),
        windowOrder: []string{},
    }
}

func (wm *WindowManager) AddWindow(w Window) {
    if _, exists := wm.windows[w.ID()]; !exists {
        wm.windows[w.ID()] = w
        wm.windowOrder = append(wm.windowOrder, w.ID())
    }
}

func (wm *WindowManager) RemoveWindow(id string) {
    delete(wm.windows, id)
    for i, wID := range wm.windowOrder {
        if wID == id {
            wm.windowOrder = append(wm.windowOrder[:i], wm.windowOrder[i+1:]...)
            break
        }
    }
}

func (wm *WindowManager) SwitchWindow(id string) error {
    if _, exists := wm.windows[id]; !exists {
        return fmt.Errorf("window %s not found", id)
    }
    wm.activeWindow = id
    return nil
}

func (wm *WindowManager) GetActiveWindow() Window {
    return wm.windows[wm.activeWindow]
}

func (wm *WindowManager) GetWindowIDs() []string {
    return wm.windowOrder
}

func (wm *WindowManager) GetWindowNames() []string {
    names := []string{}
    for _, id := range wm.windowOrder {
        names = append(names, wm.windows[id].Name())
    }
    return names
}

func (wm *WindowManager) GetActiveWindowID() string {
    return wm.activeWindow
}
```

- [ ] **Step 3: 编写测试**

```go
package windowmanager

import "testing"

func TestWindowManager_AddWindow(t *testing.T) {
    wm := NewWindowManager()
    wm.AddWindow(&mockWindow{id: "test", name: "Test"})
    
    if len(wm.GetWindowIDs()) != 1 {
        t.Error("Expected 1 window, got", len(wm.GetWindowIDs()))
    }
}

type mockWindow struct {
    id   string
    name string
}

func (m *mockWindow) ID() string              { return m.id }
func (m *mockWindow) Name() string            { return m.name }
func (m *mockWindow) View(w, h int) string    { return "" }
func (m *mockWindow) Update(msg interface{})  {}
```

- [ ] **Step 4: 运行测试**

```bash
go test -v ./internal/ui/windowmanager/...
```

- [ ] **Step 5: 提交**

```bash
git add internal/ui/windowmanager/
git commit -m "feat: add window manager component"
```

---

## Task 2: 创建标签栏组件

**Files:**
- Create: `internal/ui/tabbar/tabbar.go`
- Create: `internal/ui/tabbar/tabbar_test.go`

- [ ] **Step 1: 实现 TabBar**

```go
package tabbar

import (
    "fmt"
    "strings"
    
    "github.com/charmbracelet/lipgloss"
)

type Tab struct {
    ID   string
    Name string
    Icon string
}

type TabBar struct {
    tabs      []Tab
    activeTab int
    width     int
}

func NewTabBar(tabs []Tab) *TabBar {
    return &TabBar{
        tabs:      tabs,
        activeTab: 0,
        width:     80,
    }
}

func (tb *TabBar) SetWidth(width int) {
    tb.width = width
}

func (tb *TabBar) SetActiveTab(index int) {
    if index >= 0 && index < len(tb.tabs) {
        tb.activeTab = index
    }
}

func (tb *TabBar) GetActiveTab() int {
    return tb.activeTab
}

func (tb *TabBar) View() string {
    var tabs []string
    
    for i, tab := range tb.tabs {
        style := lipgloss.NewStyle().
            Padding(0, 3).
            Margin(0, 1).
            BorderBottom(1).
            BorderForeground(lipgloss.Color("#3c3c3c"))
        
        if i == tb.activeTab {
            style = style.
                Background(lipgloss.Color("#252526")).
                BorderForeground(lipgloss.Color("#569cd6")).
                Foreground(lipgloss.Color("#569cd6"))
        } else {
            style = style.Foreground(lipgloss.Color("#858585"))
        }
        
        tabs = append(tabs, style.Render(tab.Name))
    }
    
    return lipgloss.NewStyle().
        Width(tb.width).
        Background(lipgloss.Color("#1e1e1e")).
        Render(strings.Join(tabs, ""))
}

func (tb *TabBar) HandleClick(x int) (string, bool) {
    offset := 0
    for i, tab := range tb.tabs {
        tabWidth := lipgloss.Width(tab.Name) + 8 // padding + margin
        if x >= offset && x < offset+tabWidth {
            tb.activeTab = i
            return tab.ID, true
        }
        offset += tabWidth
    }
    return "", false
}
```

- [ ] **Step 2: 编写测试**

```go
package tabbar

import "testing"

func TestTabBar_View(t *testing.T) {
    tabs := []Tab{
        {ID: "chat", Name: "Chat"},
        {ID: "editor", Name: "Editor"},
    }
    tb := NewTabBar(tabs)
    result := tb.View()
    
    if len(result) == 0 {
        t.Error("Expected non-empty view")
    }
}
```

- [ ] **Step 3: 运行测试**

```bash
go test -v ./internal/ui/tabbar/...
```

- [ ] **Step 4: 提交**

```bash
git add internal/ui/tabbar/
git commit -m "feat: add tab bar component"
```

---

## Task 3: 创建语法高亮工具

**Files:**
- Create: `internal/ui/syntax/syntax.go`
- Create: `internal/ui/syntax/syntax_test.go`

- [ ] **Step 1: 实现语法高亮**

```go
package syntax

import (
    "bytes"
    
    "github.com/alecthomas/chroma/v2"
    "github.com/alecthomas/chroma/v2/formatters"
    "github.com/alecthomas/chroma/v2/lexers"
    "github.com/alecthomas/chroma/v2/styles"
)

func Highlight(code, language string) string {
    return HighlightWithStyle(code, language, "monokai")
}

func HighlightWithStyle(code, language, styleName string) string {
    var lexer chroma.Lexer
    
    if language != "" {
        lexer = lexers.Get(language)
    }
    if lexer == nil {
        lexer = lexers.Analyse(code)
    }
    if lexer == nil {
        lexer = lexers.Fallback
    }
    
    style := styles.Get(styleName)
    if style == nil {
        style = styles.Fallback
    }
    
    formatter := formatters.Get("terminal256")
    if formatter == nil {
        formatter = formatters.Fallback
    }
    
    iterator, err := lexer.Tokenise(nil, code)
    if err != nil {
        return code
    }
    
    var buf bytes.Buffer
    err = formatter.Format(&buf, style, iterator)
    if err != nil {
        return code
    }
    
    return buf.String()
}

func DetectLanguage(code string) string {
    lexer := lexers.Analyse(code)
    if lexer != nil {
        return lexer.Config().Name
    }
    return ""
}
```

- [ ] **Step 2: 编写测试**

```go
package syntax

import "testing"

func TestHighlight(t *testing.T) {
    code := "func main() { fmt.Println(\"Hello\") }"
    result := Highlight(code, "go")
    
    if len(result) == 0 {
        t.Error("Expected non-empty result")
    }
}
```

- [ ] **Step 3: 运行测试**

```bash
go test -v ./internal/ui/syntax/...
```

- [ ] **Step 4: 提交**

```bash
git add internal/ui/syntax/
git commit -m "feat: add syntax highlighting"
```

---

## Task 4: 创建窗口组件

**Files:**
- Create: `internal/ui/windows/chat.go`
- Create: `internal/ui/windows/editor.go`
- Create: `internal/ui/windows/terminal.go`
- Create: `internal/ui/windows/logs.go`

- [ ] **Step 1: 创建 ChatWindow**

```go
package windows

import (
    "strings"
    
    "github.com/charmbracelet/lipgloss"
    "github.com/example/agent-tui/internal/ui/syntax"
)

type Message struct {
    Role    string
    Content string
}

type ChatWindow struct {
    messages       []Message
    thinking       bool
    thinkingText   string
    thinkingFolded bool
}

func NewChatWindow() *ChatWindow {
    return &ChatWindow{
        messages:       []Message{},
        thinking:       false,
        thinkingText:   "thinking",
        thinkingFolded: false,
    }
}

func (w *ChatWindow) ID() string              { return "chat" }
func (w *ChatWindow) Name() string            { return "Chat" }
func (w *ChatWindow) Update(msg interface{})  {}

func (w *ChatWindow) ToggleThinkingFold() {
    w.thinkingFolded = !w.thinkingFolded
}

func (w *ChatWindow) AddMessage(role, content string) {
    w.messages = append(w.messages, Message{Role: role, Content: content})
}

func (w *ChatWindow) View(width, height int) string {
    style := lipgloss.NewStyle().
        Width(width).
        Height(height).
        Background(lipgloss.Color("#1e1e1e")).
        Padding(1, 2)
    
    var content strings.Builder
    
    for _, msg := range w.messages {
        roleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#569cd6"))
        contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#d4d4d4"))
        
        switch msg.Role {
        case "user":
            content.WriteString(roleStyle.Render("👤 You: "))
        case "assistant":
            content.WriteString(roleStyle.Render("🤖 Assistant: "))
        case "system":
            content.WriteString(roleStyle.Render("⚙️ System: "))
        }
        
        highlighted := syntax.Highlight(msg.Content, "")
        content.WriteString(contentStyle.Render(highlighted))
        content.WriteString("\n\n")
    }
    
    if w.thinking {
        thinkingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#dcdcaa"))
        if w.thinkingFolded {
            content.WriteString(thinkingStyle.Render("⏳ Thinking... (按 Ctrl+F 展开)"))
        } else {
            content.WriteString(thinkingStyle.Render("⏳ " + w.thinkingText + "..."))
        }
    }
    
    return style.Render(content.String())
}
```

- [ ] **Step 2: 创建 EditorWindow**

```go
package windows

import (
    "strings"
    
    "github.com/charmbracelet/lipgloss"
    "github.com/example/agent-tui/internal/ui/syntax"
)

type EditorWindow struct {
    content    string
    cursorX    int
    cursorY    int
    syntaxLang string
}

func NewEditorWindow() *EditorWindow {
    return &EditorWindow{
        content:    "// 代码编辑器\nfunc main() {\n    fmt.Println(\"Hello World\")\n}",
        cursorX:    0,
        cursorY:    0,
        syntaxLang: "go",
    }
}

func (w *EditorWindow) ID() string              { return "editor" }
func (w *EditorWindow) Name() string            { return "Editor" }
func (w *EditorWindow) Update(msg interface{})  {}

func (w *EditorWindow) View(width, height int) string {
    style := lipgloss.NewStyle().
        Width(width).
        Height(height).
        Background(lipgloss.Color("#1e1e1e")).
        Padding(1, 2)
    
    lines := strings.Split(w.content, "\n")
    highlighted := syntax.Highlight(w.content, w.syntaxLang)
    highlightedLines := strings.Split(highlighted, "\n")
    
    var result strings.Builder
    for i, line := range lines {
        // 添加行号
        lineNum := lipgloss.NewStyle().
            Foreground(lipgloss.Color("#858585")).
            Width(4).
            Render(fmt.Sprintf("%3d ", i+1))
        
        highlightedLine := line
        if i < len(highlightedLines) {
            highlightedLine = highlightedLines[i]
        }
        
        result.WriteString(lineNum + highlightedLine + "\n")
    }
    
    return style.Render(result.String())
}
```

- [ ] **Step 3: 创建 TerminalWindow**

```go
package windows

import (
    "fmt"
    "strings"
    
    "github.com/charmbracelet/lipgloss"
)

type TerminalWindow struct {
    history   []string
    input     string
    cursorPos int
}

func NewTerminalWindow() *TerminalWindow {
    return &TerminalWindow{
        history:   []string{"$ Welcome to Agent TUI Terminal"},
        input:     "",
        cursorPos: 0,
    }
}

func (w *TerminalWindow) ID() string              { return "terminal" }
func (w *TerminalWindow) Name() string            { return "Terminal" }
func (w *TerminalWindow) Update(msg interface{})  {}

func (w *TerminalWindow) ExecuteCommand(cmd string) {
    w.history = append(w.history, "$ "+cmd)
    w.history = append(w.history, fmt.Sprintf("Command executed: %s", cmd))
    w.input = ""
}

func (w *TerminalWindow) View(width, height int) string {
    style := lipgloss.NewStyle().
        Width(width).
        Height(height).
        Background(lipgloss.Color("#1e1e1e")).
        Padding(1, 2)
    
    var content strings.Builder
    
    for _, line := range w.history {
        content.WriteString(line + "\n")
    }
    
    prompt := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ec9b0")).Render("$ ")
    content.WriteString(prompt + w.input + lipgloss.NewStyle().Blink(true).Render("█"))
    
    return style.Render(content.String())
}
```

- [ ] **Step 4: 创建 LogsWindow**

```go
package windows

import (
    "fmt"
    "strings"
    "time"
    
    "github.com/charmbracelet/lipgloss"
)

type LogEntry struct {
    Time    time.Time
    Level   string
    Message string
}

type LogsWindow struct {
    logs      []LogEntry
    scrollPos int
}

func NewLogsWindow() *LogsWindow {
    return &LogsWindow{
        logs: []LogEntry{
            {Time: time.Now(), Level: "info", Message: "Application started"},
            {Time: time.Now(), Level: "info", Message: "Loading configuration..."},
            {Time: time.Now(), Level: "success", Message: "Configuration loaded"},
        },
        scrollPos: 0,
    }
}

func (w *LogsWindow) ID() string              { return "logs" }
func (w *LogsWindow) Name() string            { return "Logs" }
func (w *LogsWindow) Update(msg interface{})  {}

func (w *LogsWindow) AddLog(level, message string) {
    w.logs = append(w.logs, LogEntry{
        Time:    time.Now(),
        Level:   level,
        Message: message,
    })
}

func (w *LogsWindow) View(width, height int) string {
    style := lipgloss.NewStyle().
        Width(width).
        Height(height).
        Background(lipgloss.Color("#1e1e1e")).
        Padding(1, 2)
    
    var content strings.Builder
    
    for _, log := range w.logs {
        timeStr := log.Time.Format("15:04:05")
        timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#858585"))
        
        var levelStyle lipgloss.Style
        switch log.Level {
        case "error":
            levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#f14c4c"))
        case "warn":
            levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#dcdcaa"))
        case "success":
            levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ec9b0"))
        default:
            levelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#569cd6"))
        }
        
        content.WriteString(timeStyle.Render(timeStr) + " ")
        content.WriteString(levelStyle.Render("[" + log.Level + "] ") + " ")
        content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#d4d4d4")).Render(log.Message))
        content.WriteString("\n")
    }
    
    return style.Render(content.String())
}
```

- [ ] **Step 5: 提交**

```bash
git add internal/ui/windows/
git commit -m "feat: add window components"
```

---

## Task 5: 更新主 UI 模型

**Files:**
- Modify: `internal/ui/ui.go`

- [ ] **Step 1: 更新 ui.go**

```go
package ui

import (
    "fmt"
    "strconv"
    
    "charm.land/bubbletea/v2"
    "github.com/charmbracelet/lipgloss"
    "github.com/example/agent-tui/internal/ui/composer"
    "github.com/example/agent-tui/internal/ui/status"
    "github.com/example/agent-tui/internal/ui/tabbar"
    "github.com/example/agent-tui/internal/ui/windowmanager"
    "github.com/example/agent-tui/internal/ui/windows"
)

type Model struct {
    width          int
    height         int
    windowManager  *windowmanager.WindowManager
    tabBar         *tabbar.TabBar
    composer       *composer.Composer
    statusBar      *status.StatusBar
    themeService   *ThemeService
    isLoading      bool
}

func NewModel() *Model {
    themeService := NewThemeService(nil)
    
    wm := windowmanager.NewWindowManager()
    wm.AddWindow(windows.NewChatWindow())
    wm.AddWindow(windows.NewEditorWindow())
    wm.AddWindow(windows.NewTerminalWindow())
    wm.AddWindow(windows.NewLogsWindow())
    wm.SwitchWindow("chat")
    
    tabs := []tabbar.Tab{
        {ID: "chat", Name: "Chat"},
        {ID: "editor", Name: "Editor"},
        {ID: "terminal", Name: "Terminal"},
        {ID: "logs", Name: "Logs"},
    }
    
    return &Model{
        windowManager: wm,
        tabBar:        tabbar.NewTabBar(tabs),
        composer:      composer.NewComposer(),
        statusBar:     status.NewStatusBar(),
        themeService:  themeService,
        isLoading:     false,
    }
}

func (m *Model) Init() tea.Cmd {
    return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.tabBar.SetWidth(msg.Width)
        m.composer.SetWidth(msg.Width)
    
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
        case "ctrl+t":
            m.themeService.NextTheme()
        case "ctrl+1":
            m.SwitchWindow(0)
        case "ctrl+2":
            m.SwitchWindow(1)
        case "ctrl+3":
            m.SwitchWindow(2)
        case "ctrl+4":
            m.SwitchWindow(3)
        case "ctrl+tab":
            m.NextWindow()
        case "ctrl+shift+tab":
            m.PrevWindow()
        case "enter":
            return m, m.submitMessageAsync()
        case "backspace":
            m.composer.Backspace()
        default:
            s := msg.String()
            if len(s) > 0 {
                m.composer.AppendInput(s)
            }
        }
    }
    return m, nil
}

func (m *Model) SwitchWindow(index int) {
    ids := m.windowManager.GetWindowIDs()
    if index >= 0 && index < len(ids) {
        m.windowManager.SwitchWindow(ids[index])
        m.tabBar.SetActiveTab(index)
    }
}

func (m *Model) NextWindow() {
    ids := m.windowManager.GetWindowIDs()
    current := 0
    for i, id := range ids {
        if id == m.windowManager.GetActiveWindowID() {
            current = i
            break
        }
    }
    next := (current + 1) % len(ids)
    m.SwitchWindow(next)
}

func (m *Model) PrevWindow() {
    ids := m.windowManager.GetWindowIDs()
    current := 0
    for i, id := range ids {
        if id == m.windowManager.GetActiveWindowID() {
            current = i
            break
        }
    }
    prev := (current - 1 + len(ids)) % len(ids)
    m.SwitchWindow(prev)
}

func (m *Model) submitMessageAsync() tea.Cmd {
    return func() tea.Msg {
        input := m.composer.GetInput()
        if chat, ok := m.windowManager.GetActiveWindow().(*windows.ChatWindow); ok {
            chat.AddMessage("user", input)
        }
        m.composer.ClearInput()
        return nil
    }
}

func (m *Model) View() string {
    statusContent := m.statusBar.View(m.width)
    tabContent := m.tabBar.View()
    composerContent := m.composer.View()
    
    activeWindow := m.windowManager.GetActiveWindow()
    mainHeight := m.height - 8 // 减去状态栏(2) + 标签栏(2) + 输入区(4)
    mainContent := activeWindow.View(m.width-4, mainHeight)
    
    mainStyle := lipgloss.NewStyle().
        Width(m.width).
        Background(lipgloss.Color("#1e1e1e"))
    
    return statusContent + "\n" + mainStyle.Render(mainContent) + "\n" + tabContent + "\n" + composerContent
}
```

- [ ] **Step 2: 运行测试**

```bash
go test -v ./internal/ui/...
```

- [ ] **Step 3: 提交**

```bash
git add internal/ui/ui.go
git commit -m "feat: update main UI model with multi-window support"
```

---

## Task 6: 更新状态栏组件

**Files:**
- Modify: `internal/ui/status/status.go`

- [ ] **Step 1: 更新 StatusBar**

```go
package status

import (
    "fmt"
    
    "github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
    mode       string
    status     string
    taskCount  int
    connection bool
}

func NewStatusBar() *StatusBar {
    return &StatusBar{
        mode:       "Chat",
        status:     "Ready",
        taskCount:  0,
        connection: true,
    }
}

func (sb *StatusBar) SetMode(mode string) {
    sb.mode = mode
}

func (sb *StatusBar) SetStatus(status string) {
    sb.status = status
}

func (sb *StatusBar) SetTaskCount(count int) {
    sb.taskCount = count
}

func (sb *StatusBar) SetConnected(connected bool) {
    sb.connection = connected
}

func (sb *StatusBar) View(width int) string {
    style := lipgloss.NewStyle().
        Width(width).
        Height(1).
        Background(lipgloss.Color("#252526")).
        Padding(0, 2)
    
    left := lipgloss.NewStyle().Foreground(lipgloss.Color("#569cd6")).Render("Agent TUI")
    left += lipgloss.NewStyle().Foreground(lipgloss.Color("#858585")).Render(" | " + sb.mode)
    
    connectionStatus := "Connected"
    if !sb.connection {
        connectionStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("#f14c4c")).Render("Disconnected")
    }
    
    right := lipgloss.NewStyle().Foreground(lipgloss.Color("#858585")).Render(fmt.Sprintf(
        "Tasks: %d | Status: %s | %s",
        sb.taskCount,
        sb.status,
        connectionStatus,
    ))
    
    return style.Render(
        lipgloss.JoinHorizontal(lipgloss.Center, left, right),
    )
}
```

- [ ] **Step 2: 提交**

```bash
git add internal/ui/status/status.go
git commit -m "feat: update status bar"
```

---

## Task 7: 更新主题配置

**Files:**
- Modify: `internal/ui/theme.go`

- [ ] **Step 1: 更新主题配置**

```go
package ui

type ColorPalette struct {
    Background   string
    PanelBg      string
    Text         string
    TextMuted    string
    Border       string
    Accent       string
    Success      string
    Warning      string
    Error        string
    UserFg       string
    UserBg       string
    AssistantFg  string
    AssistantBg  string
    SystemFg     string
    SystemBg     string
    Loading      string
}

type Theme struct {
    Name        string
    Description string
    Colors      ColorPalette
}

var DefaultThemes = []Theme{
    {
        Name:        "Dark",
        Description: "VS Code dark theme",
        Colors: ColorPalette{
            Background:   "#1e1e1e",
            PanelBg:      "#252526",
            Text:         "#d4d4d4",
            TextMuted:    "#858585",
            Border:       "#3c3c3c",
            Accent:       "#569cd6",
            Success:      "#4ec9b0",
            Warning:      "#dcdcaa",
            Error:        "#f14c4c",
            UserFg:       "#ffffff",
            UserBg:       "#0066cc",
            AssistantFg:  "#ffffff",
            AssistantBg:  "#28a745",
            SystemFg:     "#ffffff",
            SystemBg:     "#9370db",
            Loading:      "#ffa500",
        },
    },
}
```

- [ ] **Step 2: 提交**

```bash
git add internal/ui/theme.go
git commit -m "feat: update theme colors"
```

---

## Task 8: 集成测试与编译

**Files:**
- Modify: `cmd/agent/main.go`

- [ ] **Step 1: 检查 main.go**

```bash
cat cmd/agent/main.go
```

- [ ] **Step 2: 编译测试**

```bash
go build -o agent-tui ./cmd/agent
```

- [ ] **Step 3: 运行所有测试**

```bash
go test -v ./...
```

- [ ] **Step 4: 运行程序**

```bash
./agent-tui
```

- [ ] **Step 5: 提交**

```bash
git add -A
git commit -m "feat: complete TUI multi-window implementation"
```

---

## 实施总结

| Task | 描述 | 预计改动 |
|------|------|---------|
| 1 | 创建窗口管理器 | 创建 2 files |
| 2 | 创建标签栏 | 创建 2 files |
| 3 | 创建语法高亮 | 创建 2 files |
| 4 | 创建窗口组件 | 创建 4 files |
| 5 | 更新主 UI 模型 | 修改 1 file |
| 6 | 更新状态栏 | 修改 1 file |
| 7 | 更新主题配置 | 修改 1 file |
| 8 | 集成测试 | 编译验证 |

**总计**: 创建 10 files，修改 3 files

---

## 自检清单

- [x] Spec 覆盖：所有功能需求都有对应任务
- [x] 占位符扫描：无 TBD/TODO
- [x] 类型一致性：Window 接口在所有窗口中保持一致