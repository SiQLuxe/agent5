# TUI 多窗口系统设计文档

**文档版本**: v1.0
**创建日期**: 2026-05-23
**作者**: Agent Designer

---

## 1. 需求概述

基于 Go 语言开发终端用户界面(TUI)，实现以下核心功能：

| 功能 | 描述 |
|------|------|
| 多窗口切换系统 | 底部标签栏，支持快捷键/鼠标切换，保留状态 |
| 代码风格文本显示 | 等宽字体，语法高亮，深色主题 |
| 内容交互功能 | Thinking 过程折叠/展开 |
| 状态与进度显示 | 实时任务状态显示 |
| 输入与交互 | 底部固定输入区，快捷键提示 |

### 1.1 界面布局

```
┌─────────────────────────────────────────────────────────────────┐
│  顶部状态栏：标题、当前模式、连接状态                            │
├─────────────────────────────────────────────────────────────────┤
│  中部主内容区：根据当前窗口切换显示                              │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  当前激活窗口内容                                        │    │
│  └─────────────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────────┤
│  底部标签栏：[Chat] [Editor] [Terminal] [Logs]              │
├─────────────────────────────────────────────────────────────────┤
│  底部输入区：输入框 + 快捷键提示                               │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 配色方案（VS Code 深色主题）

| 元素 | 颜色 | 十六进制 |
|------|------|---------|
| 主背景 | 深色背景 | `#1e1e1e` |
| 面板背景 | 面板背景 | `#252526` |
| 主文本 | 文本原色 | `#d4d4d4` |
| 次要文本 | 注释灰色 | `#858585` |
| 高亮色 | 蓝色高亮 | `#569cd6` |
| 边框 | 细微边框 | `#3c3c3c` |
| 成功 | 绿色 | `#4ec9b0` |
| 警告 | 橙色 | `#dcdcaa` |
| 错误 | 红色 | `#f14c4c` |

---

## 2. 架构设计

### 2.1 组件结构

```
┌─────────────────────────────────────────────────────────────┐
│                        Model (主状态)                       │
├─────────────────────────────────────────────────────────────┤
│  - windowManager: WindowManager  // 窗口管理器               │
│  - tabBar: TabBar              // 标签栏                    │
│  - composer: Composer          // 输入区                    │
│  - statusBar: StatusBar        // 状态栏                    │
├─────────────────────────────────────────────────────────────┤
│                     WindowManager                           │
├─────────────────────────────────────────────────────────────┤
│  - windows: map[string]Window  // 窗口集合                  │
│  - activeWindow: string        // 当前激活窗口ID            │
│  - SwitchWindow(id)            // 切换窗口                  │
│  - GetActiveWindow()           // 获取激活窗口              │
├─────────────────────────────────────────────────────────────┤
│                         TabBar                              │
├─────────────────────────────────────────────────────────────┤
│  - tabs: []Tab                 // 标签列表                  │
│  - activeTab: int              // 当前激活标签索引          │
│  - View()                      // 渲染标签栏                │
├─────────────────────────────────────────────────────────────┤
│                      Window (接口)                          │
├─────────────────────────────────────────────────────────────┤
│  - ID() string                 // 获取窗口ID                │
│  - Name() string               // 获取窗口名称              │
│  - View(width, height) string  // 渲染窗口内容              │
│  - Update(msg)                 // 处理消息                  │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 窗口类型

| 窗口 ID | 名称 | 用途 | 实现文件 |
|---------|------|------|---------|
| `chat` | Chat | AI 对话界面 | `ui/windows/chat.go` |
| `editor` | Editor | 代码编辑器 | `ui/windows/editor.go` |
| `terminal` | Terminal | 命令行终端 | `ui/windows/terminal.go` |
| `logs` | Logs | 日志查看 | `ui/windows/logs.go` |

### 2.3 快捷键设计

| 快捷键 | 功能 | 处理位置 |
|--------|------|---------|
| `Ctrl+1-4` | 切换到第 N 个窗口 | Model.Update |
| `Ctrl+Tab` | 循环切换窗口 | Model.Update |
| `Ctrl+Shift+Tab` | 反向循环切换 | Model.Update |
| `Ctrl+C` | 退出程序 | Model.Update |
| `Enter` | 发送消息 | Composer |
| `Ctrl+F` | 折叠/展开 Thinking | ChatWindow |
| `Ctrl+T` | 切换主题 | Model.Update |

---

## 3. 核心组件设计

### 3.1 WindowManager（窗口管理器）

```go
// internal/ui/windowmanager/windowmanager.go

type WindowManager struct {
    windows       map[string]Window
    activeWindow  string
    windowOrder   []string  // 保持窗口顺序
}

func NewWindowManager() *WindowManager

func (wm *WindowManager) AddWindow(w Window)
func (wm *WindowManager) RemoveWindow(id string)
func (wm *WindowManager) SwitchWindow(id string) error
func (wm *WindowManager) GetActiveWindow() Window
func (wm *WindowManager) GetWindowIDs() []string
func (wm *WindowManager) GetWindowNames() []string
```

### 3.2 TabBar（标签栏）

```go
// internal/ui/tabbar/tabbar.go

type Tab struct {
    ID    string
    Name  string
    Icon  string
}

type TabBar struct {
    tabs      []Tab
    activeTab int
    width     int
}

func NewTabBar(tabs []Tab) *TabBar

func (tb *TabBar) SetWidth(width int)
func (tb *TabBar) SetActiveTab(index int)
func (tb *TabBar) View() string
func (tb *TabBar) HandleClick(x int) (string, bool)  // 返回窗口ID和是否点击
```

### 3.3 ChatWindow（聊天窗口）

```go
// internal/ui/windows/chat.go

type ChatWindow struct {
    messages     []Message
    thinking     bool
    thinkingText string
    thinkingFolded bool  // Thinking 是否折叠
}

type Message struct {
    Role    string
    Content string
}

func NewChatWindow() *ChatWindow

func (w *ChatWindow) ID() string
func (w *ChatWindow) Name() string
func (w *ChatWindow) View(width, height int) string
func (w *ChatWindow) Update(msg tea.Msg)
func (w *ChatWindow) ToggleThinkingFold()
func (w *ChatWindow) AddMessage(role, content string)
```

### 3.4 EditorWindow（编辑器窗口）

```go
// internal/ui/windows/editor.go

type EditorWindow struct {
    content    string
    cursorX    int
    cursorY    int
    syntaxLang string  // 语法高亮语言
}

func NewEditorWindow() *EditorWindow

func (w *EditorWindow) ID() string
func (w *EditorWindow) Name() string
func (w *EditorWindow) View(width, height int) string
func (w *EditorWindow) Update(msg tea.Msg)
```

### 3.5 TerminalWindow（终端窗口）

```go
// internal/ui/windows/terminal.go

type TerminalWindow struct {
    history    []string
    input      string
    cursorPos  int
}

func NewTerminalWindow() *TerminalWindow

func (w *TerminalWindow) ID() string
func (w *TerminalWindow) Name() string
func (w *TerminalWindow) View(width, height int) string
func (w *TerminalWindow) Update(msg tea.Msg)
func (w *TerminalWindow) ExecuteCommand(cmd string)
```

### 3.6 LogsWindow（日志窗口）

```go
// internal/ui/windows/logs.go

type LogsWindow struct {
    logs      []LogEntry
    scrollPos int
}

type LogEntry struct {
    Time     time.Time
    Level    string  // info, warn, error, debug
    Message  string
}

func NewLogsWindow() *LogsWindow

func (w *LogsWindow) ID() string
func (w *LogsWindow) Name() string
func (w *LogsWindow) View(width, height int) string
func (w *LogsWindow) Update(msg tea.Msg)
func (w *LogsWindow) AddLog(level, message string)
```

### 3.7 SyntaxHighlighter（语法高亮）

```go
// internal/ui/syntax/syntax.go

package syntax

import (
    "github.com/alecthomas/chroma/v2"
)

func Highlight(code, language string) string
func HighlightWithStyle(code, language, style string) string
func DetectLanguage(code string) string  // 自动检测语言
```

---

## 4. 主 UI 模型

```go
// internal/ui/ui.go

type Model struct {
    width           int
    height          int
    windowManager   *windowmanager.WindowManager
    tabBar          *tabbar.TabBar
    composer        *composer.Composer
    statusBar       *status.StatusBar
    themeService    *ThemeService
    isLoading       bool
}

func NewModel() *Model

func (m *Model) Init() tea.Cmd
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m *Model) View() string
```

### 4.1 View 布局结构

```go
func (m *Model) View() string {
    // 顶部状态栏
    statusContent := m.statusBar.View()
    
    // 主内容区
    activeWindow := m.windowManager.GetActiveWindow()
    mainContent := activeWindow.View(m.width, m.height - 10)  // 减去状态栏和标签栏高度
    
    // 底部标签栏
    tabContent := m.tabBar.View()
    
    // 输入区
    composerContent := m.composer.View()
    
    // 组合布局
    return statusContent + "\n" + mainContent + "\n" + tabContent + "\n" + composerContent
}
```

---

## 5. 状态与进度显示

### 5.1 StatusBar（状态栏）

```go
// internal/ui/status/status.go

type StatusBar struct {
    mode        string  // 当前模式
    status      string  // 运行中/完成/失败
    taskCount   int     // 任务计数
    connection  bool    // 连接状态
}

func NewStatusBar() *StatusBar

func (sb *StatusBar) SetMode(mode string)
func (sb *StatusBar) SetStatus(status string)
func (sb *StatusBar) SetTaskCount(count int)
func (sb *StatusBar) SetConnected(connected bool)
func (sb *StatusBar) View() string
```

### 5.2 状态显示格式

```
Agent TUI | Chat | Connected | Tasks: 3/5 | Status: Running
```

---

## 6. 输入与交互

### 6.1 Composer（输入区）

```go
// internal/ui/composer/composer.go

type Composer struct {
    width       int
    input       string
    placeholder string
    isLoading   bool
}

func NewComposer() *Composer

func (c *Composer) SetWidth(width int)
func (c *Composer) SetInput(input string)
func (c *Composer) GetInput() string
func (c *Composer) ClearInput()
func (c *Composer) AppendInput(char string)
func (c *Composer) Backspace()
func (c *Composer) SetLoading(loading bool)
func (c *Composer) View() string
```

### 6.2 快捷键提示栏

```
按 Ctrl+1-4 切换窗口 | Ctrl+Tab 循环切换 | Enter 发送 | Ctrl+C 退出 | Ctrl+T 主题
```

---

## 7. 依赖与技术栈

| 依赖 | 版本 | 用途 |
|------|------|------|
| bubbletea | v2 | TUI 框架 |
| lipgloss | latest | 样式渲染 |
| chroma | v2 | 语法高亮 |
| go-isatty | latest | 终端检测 |

---

## 8. 文件结构

```
internal/
├── ui/
│   ├── windowmanager/
│   │   └── windowmanager.go
│   ├── tabbar/
│   │   └── tabbar.go
│   ├── windows/
│   │   ├── chat.go
│   │   ├── editor.go
│   │   ├── terminal.go
│   │   └── logs.go
│   ├── syntax/
│   │   └── syntax.go
│   ├── composer/
│   │   └── composer.go
│   ├── status/
│   │   └── status.go
│   ├── ui.go
│   ├── ui_test.go
│   ├── theme.go
│   └── theme_service.go
```

---

## 9. 实施步骤

1. 创建窗口管理器组件
2. 创建标签栏组件
3. 创建四个窗口组件（Chat/Editor/Terminal/Logs）
4. 创建语法高亮工具
5. 更新主 UI 模型
6. 更新主题配置
7. 集成测试

---

## 10. 测试计划

| 测试类型 | 测试内容 |
|----------|---------|
| 单元测试 | 窗口切换、标签栏点击、语法高亮 |
| 集成测试 | 多窗口状态保持、快捷键操作 |
| E2E 测试 | 完整交互流程 |