# TUI Bubbles Migration Design

**Date**: 2026-05-29
**Status**: Draft
**Scope**: `internal/ui/` Bubbles 组件替换

---

## 1. Overview

将当前基于自实现组件的 Bubble Tea v2 TUI，替换为使用 Bubbles 官方组件库的方案。

**保留**：TabDock（自实现）

**替换为 Bubbles**：
- ChatPanel → `bubbles/viewport`
- Composer → `bubbles/textarea`（默认单行，自动增高）
- 键盘处理 → `bubbles/key`
- Help → `bubbles/help`

**升级**：
- StatusBar：保留逻辑，升级 lipgloss v2
- TabDock：保留自实现，升级 lipgloss v2
- Lipgloss: `github.com/charmbracelet/lipgloss v1.1.0` → `charm.land/lipgloss/v2`
- ColorPalette：同步升级到 lipgloss v2 API

**新增**：
- 鼠标滚轮支持（viewport + tea.WithMouseCellMotion）
- 多行输入（textarea DynamicHeight）

## 2. Dependency Changes

```
go.mod 变更：
+ charm.land/bubbles/v2 v2.1.0
~ github.com/charmbracelet/lipgloss v1.1.0 → charm.land/lipgloss/v2
```

三个模块必须一起升级：Bubbles v2 依赖 Bubble Tea v2 和 Lip Gloss v2。

当前 `charm.land/bubbletea/v2 v2.0.6` 已是 v2，无需变动。

## 3. Component Mapping

### 3.1 ChatPanel → bubbles/viewport

| 当前（自实现） | Bubbles 方案 |
|---|---|
| `ChatPanel` 自定义 viewport | `viewport.Model` |
| 手动 scrollbar 渲染 | 内置 scrollbar（MouseWheelEnabled） |
| 手动搜索高亮 | `SetHighlights()` / `HighlightNext()` / `HighlightPrevious()` |
| 手动 message 折行计算 | `SetContent()` 整个渲染 |
| 无鼠标滚轮 | `MouseWheelEnabled = true`, `MouseWheelDelta = 3` |

**渲染方式变更**：
当前：遍历 message → 渲染 → 按 line offset 切片
新：预渲染全部 message 为字符串 → `viewport.SetContent()` → viewport 自动处理滚动和可见区域

**Session.RenderMessages() 方法**：
新增方法用于将 session 的所有 messages 渲染为完整字符串（含角色 badge、时间戳、thinking 块、折叠状态），直接传给 viewport。

**搜索替换**：
- `bubbles/viewport` 支持 `SetHighlights(matches [][]int)`，传入 `[起始行, 结束行]` 数组
- 搜索时计算匹配行的行号范围，设为 highlights
- `HighlightNext()`/`HighlightPrevious()` 导航

### 3.2 Composer → bubbles/textarea (单行→多行)

| 当前（自实现） | Bubbles 方案 |
|---|---|
| `Composer` 手动 cursor | `textarea.Model` |
| 单行 | `DynamicHeight = true, MinHeight = 1, MaxHeight = 8` |
| 手动 backspace/cursor | 内置 Emacs 风格编辑 |
| 无 paste 支持 | 内置 `Paste()` command |
| Enter 发送消息 | Ctrl+Enter 发送消息, Enter 换行 |

**textarea 配置**：
```go
ta := textarea.New()
ta.SetWidth(width)
ta.SetHeight(1)          // 初始高度
ta.DynamicHeight = true  // 自适应高度
ta.MinHeight = 1
ta.MaxHeight = 8         // 防止撑满屏幕
ta.ShowLineNumbers = false
ta.Prompt = "❯ "
```

**发送机制变更**：
- 当前：Enter → submitMessage
- 新：Ctrl+Enter → submitMessage，Enter → InsertNewline（textarea 默认行为）

### 3.3 Keyboard → bubbles/key

**当前**：
```go
switch key.String() {
case "ctrl+c": return m, tea.Quit
case "ctrl+n": m.newSession()
case "ctrl+g": m.showHelp = true
...
}
```

**新**：
```go
// keymap.go
type KeyMap struct {
    Quit         key.Binding
    NewSession   key.Binding
    CloseSession key.Binding
    RenameSession key.Binding
    NextSession  key.Binding
    PrevSession  key.Binding
    SwitchToSession1..9 key.Binding
    ToggleThinking   key.Binding
    ToggleCollapse   key.Binding
    Search          key.Binding
    ToggleTheme     key.Binding
    ShowHelp        key.Binding
    ScrollUp        key.Binding
    ScrollDown      key.Binding
    SendMessage     key.Binding
}

func (km KeyMap) ShortHelp() []key.Binding
func (km KeyMap) FullHelp() [][]key.Binding
```

使用 `key.Matches(msg, km.SomeBinding)` 替代 `switch case`。

### 3.4 Help → bubbles/help

**当前**：手写 help panel（60 行字符串拼接）
**新**：
```go
helpModel := help.New()
helpModel.SetWidth(width)

// 主 Model 实现 KeyMap 接口
// View 中调用 helpModel.View(m.keyMap)
```

- 短模式：单行显示
- 长模式（Ctrl+G）：表格显示所有快捷键
- help.ShowAll 控制展开/折叠

### 3.5 StatusBar（保留 + lipgloss v2）

保持现有逻辑不变，仅升级 lipgloss v2 API：
- `github.com/charmbracelet/lipgloss` → `charm.land/lipgloss/v2`
- `lipgloss.Color()` → `lipgloss.Color()` (API 兼容)
- 主要变更：`AdaptiveColor` 改为 `Color` + 显式 `isDark`

### 3.6 TabDock（保留 + lipgloss v2）

保持自实现不变，仅升级 lipgloss v2 API。ColorPalette 中的 tab 颜色通过主题系统传入。

## 4. Mouse Support

**启用**：`tea.NewProgram(model, tea.WithMouseCellMotion())`

**viewport**：`viewport.MouseWheelEnabled = true`

**tabbar**：可通过鼠标点击切换 tab。TabDock 已有 `HandleClick(x)` 方法，鼠标事件 X 坐标匹配。

## 5. Layout & Rendering

```
StatusBar (lipgloss v2, 1行)
─────────────────────────────
viewport (bubbles, 可伸缩)
  包含全部 messages，鼠标滚轮滚动
─────────────────────────────
textarea (bubbles, DynamicHeight)
  单行到多行自动增高
─────────────────────────────
TabDock (保留, 1行)
  session tabs | + | Ctrl+G 快捷键
```

View 函数：
```go
func (m *Model) View() tea.View {
    statusV := m.statusBar.View(m.width)
    chatV := m.viewport.View()          // bubbles/viewport
    chatV.AltScreen = true
    inputV := m.textarea.View()          // bubbles/textarea
    tabV := m.tabDock.View()
    
    result := statusV + "\n" + chatV + "\n" + inputV + "\n" + tabV
    if m.showHelp {
        helpV := m.help.View(m.keyMap)
        result = m.overlayHelp(result, helpV)
    }
    return tea.NewView(result)
}
```

## 6. Update Flow (Message Pipeline)

```
用户输入 → tea.KeyPressMsg
  ├─ textarea.Update() → 输入处理（换行、粘贴、光标）
  ├─ key.Matches() 检测快捷键
  │   ├─ SendMessage → 提交 AI 请求 → ChatResponseMsg
  │   ├─ NewSession → 创建 session
  │   ├─ ShowHelp → toggle help
  │   └─ ...
  └─ viewport.Update() → 滚动处理（键盘/鼠标）

AI 响应 → ChatResponseMsg
  ├─ session.AddMessage() 添加消息
  ├─ 重新渲染全部 messages → viewport.SetContent()
  └─ viewport.GotoBottom()

窗口变化 → tea.WindowSizeMsg
  ├─ statusBar.SetWidth(msg.Width)
  ├─ viewport.SetWidth(msg.Width)
  ├─ viewport.SetHeight(chatHeight)
  ├─ textarea.SetWidth(msg.Width)
  └─ tabDock.SetWidth(msg.Width)
```

## 7. session.go 新增方法

```go
func (s *Session) RenderMessages(width int, theme ColorPalette) string {
    // 将 session 的 messages 渲染为完整字符串
    // 包含：角色 badge、时间戳、thinking 块、折叠状态
    // 用于传给 viewport.SetContent()
}
```

## 8. Theme System 迁移

当前 ColorPalette 保持不变，Bubbles 组件的样式通过 `SetStyles()` 统一设置：

```go
// 根据当前主题设置 bubbles 组件样式
func applyThemeToViewport(vp *viewport.Model, theme ColorPalette) {
    vp.Style = lipgloss.NewStyle().
        Background(lipgloss.Color(theme.Background))
}

func applyThemeToTextarea(ta *textarea.Model, theme ColorPalette) {
    s := textarea.DefaultStyles(false)  // false = dark mode
    s.Focused.Base = lipgloss.NewStyle().
        Background(lipgloss.Color(theme.InputBg))
    s.Focused.Prompt = lipgloss.NewStyle().
        Foreground(lipgloss.Color(theme.InputPrompt))
    ta.SetStyles(s)
}
```

## 9. Keyboard Shortcuts

| 快捷键 | 功能 |
|--------|------|
| Ctrl+N | 新建会话 |
| Ctrl+Q | 关闭会话 |
| Ctrl+E | 重命名会话 |
| Ctrl+T | 切换主题 |
| Ctrl+Y | 折叠/展开 thinking |
| Ctrl+L | 折叠/展开消息 |
| Ctrl+F | 搜索 |
| Ctrl+G | 快捷键帮助 |
| Ctrl+C | 退出 |
| Alt+N / Alt+→ | 下一个会话 |
| Alt+P / Alt+← | 上一个会话 |
| Alt+1~9 | 跳转到第 N 个会话 |
| PgUp / PgDn | 滚动半页 |
| Ctrl+↑ / Ctrl+↓ | 滚动一行 |
| Ctrl+Home / Ctrl+End | 滚动到顶部/底部 |
| Ctrl+Enter | 发送消息（textarea 模式） |
| Enter | 换行（textarea 模式） |

## 10. Files to Modify

| File | Change |
|---|---|
| `go.mod` / `go.sum` | 添加 `charm.land/bubbles/v2`，升级 lipgloss v2 |
| `internal/ui/ui.go` | 重写 Model：viewport + textarea + keymap + help 集成 |
| `internal/ui/chatpanel.go` | 重写为 viewport wrapper，渲染逻辑移至 session.RenderMessages() |
| `internal/ui/composer/composer.go` | 重写为 textarea wrapper，或移除 |
| `internal/ui/status/status.go` | lipgloss v2 升级 |
| `internal/ui/tabbar/tabbar.go` | lipgloss v2 升级 |
| `internal/ui/session.go` | 新增 `RenderMessages()` 方法 |
| `internal/ui/theme.go` | ColorPalette 添加 Bubbles 组件样式字段 |
| `internal/ui/session_test.go` | 新增 `RenderMessages()` 测试 |
| `cmd/agent/main.go` | 添加 `tea.WithMouseCellMotion()` |
| `internal/ui/keymap.go` | **新建**：key.Binding 定义 |

## 11. Migration Order

按依赖关系分步迁移，每步可独立编译和测试：

### Step 1: 依赖升级
- go.mod 添加 `charm.land/bubbles/v2`，升级 `charm.land/lipgloss/v2`
- 解决 lipgloss v1→v2 编译错误

### Step 2: Key + Help
- 新建 `keymap.go`
- 修改 `ui.go` Update 方法使用 `key.Matches`
- 添加 `help.Model` 替代手写 help panel
- 验证：Ctrl+G 显示帮助、所有快捷键仍然生效

### Step 3: Mouse Support
- `tea.WithMouseCellMotion()` + viewport 鼠标滚轮
- TabDock 鼠标点击切换

### Step 4: Viewport（ChatPanel）
- 替换 `ChatPanel` 为 `viewport.Model`
- `session.RenderMessages()` 新增
- 搜索功能迁移到 `viewport.SetHighlights()`
- 验证：滚动、搜索、折叠、thinking 展开

### Step 5: Textarea（Composer）
- 替换 `Composer` 为 `textarea.Model`
- Enter → 换行，Ctrl+Enter → 发送
- `DynamicHeight` 配置
- 验证：多行输入、粘贴、发送

### Step 6: 清理
- 删除未使用的自实现代码
- 更新测试文件
- 完整 UI 验证

## 12. Out of Scope

- TabDock 替换为 bubbles 组件（保留自实现）
- editor 模块（与 main TUI 分离，不涉及）
- AI 主题生成功能（theme_service.go 交互不变）
- 会话持久化到磁盘
