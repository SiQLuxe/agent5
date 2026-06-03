# Slash Command Autocomplete 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将现有的 `tview.List` 风格的 inline 斜杠建议系统替换为专用的 `SuggestionMenu` 组件，支持两列对齐显示所有命令 + 底部状态栏。

**Architecture:** 新建 `internal/ui/suggestion/suggestion.go` 组件包，用 `tview.Box.SetDrawFunc` 自定义绘制实现两列对齐列表；`App.onComposerChange` 改为显示所有命令（builtin + skill）；`App.handleInput` 键盘路由改为操作新组件。

**Tech Stack:** Go, tview, tcell, go-runewidth

---

### Task 1: SuggestionMenu 组件

**Files:**
- Create: `internal/ui/suggestion/suggestion.go`
- Dependency: `internal/service/command_registry.go` (Command 类型)

- [ ] **Step 1: 创建包目录**

```bash
mkdir -p internal/ui/suggestion
```

- [ ] **Step 2: 编写 SuggestionMenu 结构体和构造函数**

`internal/ui/suggestion/suggestion.go`:

```go
package suggestion

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"

	"github.com/example/agent-tui/internal/service"
)

type SuggestionMenu struct {
	*tview.Flex
	items      []*service.Command
	filtered   []*service.Command
	filterText string
	selected   int
	list       *tview.Box
	headerBar  *tview.TextView
	statusBar  *tview.TextView
	visible    bool
}

func New() *SuggestionMenu {
	header := tview.NewTextView()
	header.SetDynamicColors(true)
	header.SetText("[::b]/ [::-]")
	header.SetTextAlign(tview.AlignLeft)

	status := tview.NewTextView()
	status.SetDynamicColors(true)
	status.SetText("[gray]↑↓ 选择  Enter 回填  Esc 关闭[-]")
	status.SetTextAlign(tview.AlignCenter)

	list := tview.NewBox()
	list.SetBackgroundColor(tcell.ColorDefault)

	m := &SuggestionMenu{
		Flex:      tview.NewFlex().SetDirection(tview.FlexRow),
		list:      list,
		headerBar: header,
		statusBar: status,
	}

	m.AddItem(header, 1, 0, false)
	m.AddItem(list, 0, 1, false)
	m.AddItem(status, 1, 0, false)
	m.SetBackgroundColor(tcell.ColorDefault)

	list.SetDrawFunc(m.drawList)

	return m
}
```

- [ ] **Step 3: 实现自定义绘制函数 `drawList`**

```go
func (m *SuggestionMenu) drawList(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
	nameColWidth := width * 40 / 100
	if nameColWidth > 20 {
		nameColWidth = 20
	}
	descColWidth := width - nameColWidth - 1

	for i, cmd := range m.filtered {
		if i >= height {
			break
		}
		rowY := y + i

		bg := tcell.ColorDefault
		fg := tcell.ColorWhite
		if i == m.selected {
			bg = tcell.ColorOrange
		}
		style := tcell.StyleDefault.Background(bg).Foreground(fg)

		// 绘制命令名（左对齐，固定宽度）
		name := cmd.Name
		if runewidth.StringWidth(name) > nameColWidth {
			name = runewidth.Truncate(name, nameColWidth-1, "…")
		}
		name = runewidth.FillRight(name, nameColWidth)
		drawString(screen, x, rowY, style, name)

		// 绘制分隔
		screen.SetContent(x+nameColWidth, rowY, ' ', nil, style)

		// 绘制描述（灰色，非选中状态）
		descStyle := tcell.StyleDefault.Background(bg).Foreground(tcell.ColorGray)
		if i == m.selected {
			descStyle = style
		}
		desc := cmd.Description
		if runewidth.StringWidth(desc) > descColWidth {
			desc = runewidth.Truncate(desc, descColWidth-1, "…")
		}
		drawString(screen, x+nameColWidth+1, rowY, descStyle, desc)
	}

	return x, y, width, height
}

func drawString(screen tcell.Screen, x, y int, style tcell.Style, s string) {
	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if w == 0 {
			continue
		}
		screen.SetContent(x, y, r, nil, style)
		x += w
	}
}
```

- [ ] **Step 4: 添加 `Filtered()` 方法**

```go
func (m *SuggestionMenu) Filtered() []*service.Command {
	return m.filtered
}
```

- [ ] **Step 5: 实现 SetCommands / SetFilter / applyFilter**

```go
func (m *SuggestionMenu) SetCommands(cmds []*service.Command) {
	m.items = cmds
	m.selected = 0
	m.applyFilter()
}

func (m *SuggestionMenu) SetFilter(text string) {
	m.filterText = text
	m.selected = 0
	m.applyFilter()
}

func (m *SuggestionMenu) applyFilter() {
	m.filtered = nil
	for _, cmd := range m.items {
		if strings.HasPrefix(strings.ToLower(cmd.Name), strings.ToLower(m.filterText)) {
			m.filtered = append(m.filtered, cmd)
		}
	}
}
```

- [ ] **Step 6: 实现 SelectNext / SelectPrev / Selected / Show / Hide / Visible / Height**

```go
func (m *SuggestionMenu) SelectNext() {
	if m.selected < len(m.filtered)-1 {
		m.selected++
	}
}

func (m *SuggestionMenu) SelectPrev() {
	if m.selected > 0 {
		m.selected--
	}
}

func (m *SuggestionMenu) Selected() *service.Command {
	if m.selected < 0 || m.selected >= len(m.filtered) {
		return nil
	}
	return m.filtered[m.selected]
}

func (m *SuggestionMenu) Show() {
	m.visible = true
}

func (m *SuggestionMenu) Hide() {
	m.visible = false
}

func (m *SuggestionMenu) Visible() bool {
	return m.visible
}

func (m *SuggestionMenu) Height() int {
	if !m.visible || len(m.filtered) == 0 {
		return 0
	}
	n := len(m.filtered)
	if n > 8 {
		n = 8
	}
	return n + 2 // items + header + status
}
```

- [ ] **Step 7: 检查并整理完整文件**

确保 `internal/ui/suggestion/suggestion.go` 包含所有 import：

```go
import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"

	"github.com/example/agent-tui/internal/service"
)
```

完整文件确认所有方法都已定义，无遗漏。

---

### Task 2: SuggestionMenu 单元测试

**Files:**
- Create: `internal/ui/suggestion/suggestion_test.go`

- [ ] **Step 1: 编写所有测试**

```go
package suggestion

import (
	"testing"

	"github.com/example/agent-tui/internal/service"
)

func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
	if m.headerBar == nil {
		t.Fatal("expected headerBar")
	}
	if m.list == nil {
		t.Fatal("expected list")
	}
	if m.statusBar == nil {
		t.Fatal("expected statusBar")
	}
}

func TestSetCommands(t *testing.T) {
	m := New()
	cmds := []*service.Command{
		{Name: "Search", Description: "搜索", Category: service.CmdBuiltin},
		{Name: "code-review", Description: "Review code", Category: service.CmdSkill},
	}
	m.SetCommands(cmds)
	if len(m.items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(m.items))
	}
}

func TestSetFilter(t *testing.T) {
	m := New()
	cmds := []*service.Command{
		{Name: "Search", Description: "搜索", Category: service.CmdBuiltin},
		{Name: "New Session", Description: "新建会话", Category: service.CmdBuiltin},
		{Name: "Scroll to Top", Description: "滚动到顶部", Category: service.CmdBuiltin},
	}
	m.SetCommands(cmds)
	m.SetFilter("Se")
	if len(m.filtered) != 1 {
		t.Fatalf("expected 1 filtered, got %d", len(m.filtered))
	}
	if m.filtered[0].Name != "Search" {
		t.Fatalf("expected Search, got %s", m.filtered[0].Name)
	}
}

func TestSetFilterCaseInsensitive(t *testing.T) {
	m := New()
	cmds := []*service.Command{
		{Name: "Search", Description: "搜索"},
		{Name: "scroll-to-top", Description: "滚动"},
	}
	m.SetCommands(cmds)
	m.SetFilter("scroll")
	if len(m.filtered) != 1 {
		t.Fatalf("expected 1 filtered, got %d", len(m.filtered))
	}
	if m.filtered[0].Name != "scroll-to-top" {
		t.Fatalf("expected scroll-to-top, got %s", m.filtered[0].Name)
	}
}

func TestSetFilterNoMatch(t *testing.T) {
	m := New()
	m.SetCommands([]*service.Command{{Name: "Search"}})
	m.SetFilter("xyz")
	if len(m.filtered) != 0 {
		t.Fatalf("expected 0 filtered, got %d", len(m.filtered))
	}
	if m.Visible() {
		t.Fatal("expected not visible with no match")
	}
}

func TestNavigation(t *testing.T) {
	m := New()
	cmds := []*service.Command{
		{Name: "A"}, {Name: "B"}, {Name: "C"},
	}
	m.SetCommands(cmds)
	m.SetFilter("")

	if m.selected != 0 {
		t.Fatalf("expected selected 0, got %d", m.selected)
	}
	m.SelectNext()
	if m.selected != 1 {
		t.Fatalf("expected selected 1, got %d", m.selected)
	}
	m.SelectNext()
	if m.selected != 2 {
		t.Fatalf("expected selected 2, got %d", m.selected)
	}
	m.SelectNext() // should stay at 2
	if m.selected != 2 {
		t.Fatalf("expected selected 2 (clamped), got %d", m.selected)
	}
	m.SelectPrev()
	if m.selected != 1 {
		t.Fatalf("expected selected 1, got %d", m.selected)
	}
	m.SelectPrev()
	if m.selected != 0 {
		t.Fatalf("expected selected 0, got %d", m.selected)
	}
	m.SelectPrev() // should stay at 0
	if m.selected != 0 {
		t.Fatalf("expected selected 0 (clamped), got %d", m.selected)
	}
}

func TestSelected(t *testing.T) {
	m := New()
	cmds := []*service.Command{
		{Name: "Alpha"}, {Name: "Beta"},
	}
	m.SetCommands(cmds)
	m.SetFilter("")

	sel := m.Selected()
	if sel == nil || sel.Name != "Alpha" {
		t.Fatalf("expected Alpha, got %v", sel)
	}
	m.SelectNext()
	sel = m.Selected()
	if sel == nil || sel.Name != "Beta" {
		t.Fatalf("expected Beta, got %v", sel)
	}
}

func TestSelectedEmpty(t *testing.T) {
	m := New()
	sel := m.Selected()
	if sel != nil {
		t.Fatal("expected nil when no commands")
	}
}

func TestShowHide(t *testing.T) {
	m := New()
	if m.Visible() {
		t.Fatal("expected not visible initially")
	}
	if m.Height() != 0 {
		t.Fatalf("expected height 0, got %d", m.Height())
	}
	m.SetCommands([]*service.Command{{Name: "Test"}})
	m.SetFilter("")
	m.Show()
	if !m.Visible() {
		t.Fatal("expected visible after Show()")
	}
	if m.Height() != 3 { // 1 item + header + status = 3
		t.Fatalf("expected height 3, got %d", m.Height())
	}
	m.Hide()
	if m.Visible() {
		t.Fatal("expected not visible after Hide()")
	}
	if m.Height() != 0 {
		t.Fatalf("expected height 0, got %d", m.Height())
	}
}

func TestHeightMax(t *testing.T) {
	m := New()
	cmds := make([]*service.Command, 20)
	for i := 0; i < 20; i++ {
		cmds[i] = &service.Command{Name: string(rune('A' + i))}
	}
	m.SetCommands(cmds)
	m.SetFilter("")
	m.Show()
	if m.Height() != 10 { // max 8 items + header + status = 10
		t.Fatalf("expected height 10 (capped), got %d", m.Height())
	}
}
```

- [ ] **Step 2: 运行测试**

```bash
go test ./internal/ui/suggestion/... -count=1 -v
```

Expected: ALL PASS

- [ ] **Step 3: Commit**

```bash
git add internal/ui/suggestion/
git commit -m "feat: add SuggestionMenu component with two-column aligned rendering"
```

---

### Task 3: 集成到 App

**Files:**
- Modify: `internal/ui/app.go`

- [ ] **Step 1: 在 App 结构体中替换字段**

在 `type App struct` 中：

替换：
```go
suggestionList  *tview.List
suggestionCmds  []*service.Command
```

为：
```go
suggestionMenu *suggestion.SuggestionMenu
```

添加 import：
```go
"github.com/example/agent-tui/internal/ui/suggestion"
```

移除不再需要的 import（如果没有其他地方使用 `tview.List` 的话保留）。

- [ ] **Step 2: 更新 NewApp() 构造函数**

在 `NewApp()` 中：

替换创建 `suggestionList` 的代码（约行 105-113）：
```go
a.suggestionMenu = suggestion.New()
```

替换布局顺序，将 SuggestionMenu 放在 ChatPanel 和 Composer 之间：
```go
// Build layout: StatusBar + ChatPanel + SuggestionMenu + Composer + TabDock
chatFlex := tview.NewFlex().SetDirection(tview.FlexRow)
chatFlex.AddItem(a.statusBar, 1, 0, false)
chatFlex.AddItem(a.chatPanel, 0, 1, false)
chatFlex.AddItem(a.suggestionMenu, 0, 0, false) // hidden by default
chatFlex.AddItem(a.composer, 3, 0, true)
chatFlex.AddItem(a.tabDock, 1, 0, false)
```

- [ ] **Step 3: 更新 onComposerChange**

```go
func (a *App) onComposerChange(text string) {
	if a.commandRegistry == nil {
		return
	}
	if strings.HasPrefix(text, "/") {
		prefix := strings.TrimPrefix(text, "/")
		var cmds []*service.Command
		for _, c := range a.commandRegistry.List() {
			if strings.HasPrefix(strings.ToLower(c.Name), strings.ToLower(prefix)) {
				cmds = append(cmds, c)
			}
		}
		a.suggestionMenu.SetCommands(cmds)
		a.suggestionMenu.SetFilter(prefix)
		if len(a.suggestionMenu.Filtered()) > 0 {
			a.showSuggestions()
		} else {
			a.hideSuggestions()
		}
		return
	}
	a.hideSuggestions()
}
```

注意：需要给 SuggestionMenu 加一个 `Filtered()` 方法返回 `m.filtered`：

```go
func (m *SuggestionMenu) Filtered() []*service.Command {
	return m.filtered
}
```

- [ ] **Step 4: 更新 showSuggestions / hideSuggestions**

```go
func (a *App) showSuggestions() {
	a.suggestionMenu.Show()
	a.chatFlex.RemoveItem(a.suggestionMenu)
	a.chatFlex.AddItem(a.suggestionMenu, a.suggestionMenu.Height(), 0, false)
}

func (a *App) hideSuggestions() {
	a.suggestionMenu.Hide()
	a.chatFlex.RemoveItem(a.suggestionMenu)
	a.chatFlex.AddItem(a.suggestionMenu, 0, 0, false)
}
```

- [ ] **Step 5: 更新 handleInput 键盘事件**

在 `handleInput` 的 Chat mode 中，找到所有引用 `a.suggestionCmds` 和 `a.suggestionList` 的地方，替换为 `a.suggestionMenu`：

替换：
```go
case event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModNone:
    if a.suggestionCmds != nil {
        idx := a.suggestionList.GetCurrentItem()
        if idx >= 0 && idx < len(a.suggestionCmds) {
            cmd := a.suggestionCmds[idx]
            a.composer.SetInput("/" + cmd.Name + " ")
            a.hideSuggestions()
        }
        return nil
    }
```

为：
```go
case event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModNone:
    if a.suggestionMenu.Visible() {
        cmd := a.suggestionMenu.Selected()
        if cmd != nil {
            a.composer.SetInput("/" + cmd.Name + " ")
            a.hideSuggestions()
        }
        return nil
    }
```

替换 Tab 处理：
```go
case event.Key() == tcell.KeyTab && a.suggestionCmds != nil:
    idx := a.suggestionList.GetCurrentItem()
    if idx >= 0 && idx < len(a.suggestionCmds) {
        cmd := a.suggestionCmds[idx]
        a.composer.SetInput("/" + cmd.Name + " ")
        a.hideSuggestions()
    }
    return nil
```

为：
```go
case event.Key() == tcell.KeyTab && a.suggestionMenu.Visible():
    cmd := a.suggestionMenu.Selected()
    if cmd != nil {
        a.composer.SetInput("/" + cmd.Name + " ")
        a.hideSuggestions()
    }
    return nil
```

替换 ↑ 键处理中的 suggestion 部分：
```go
case event.Key() == tcell.KeyUp:
    if len(a.suggestionCmds) > 0 {
        idx := a.suggestionList.GetCurrentItem()
        if idx > 0 {
            a.suggestionList.SetCurrentItem(idx - 1)
        }
    } else if a.historyIndex < len(a.inputHistory) {
```

为：
```go
case event.Key() == tcell.KeyUp:
    if a.suggestionMenu.Visible() {
        a.suggestionMenu.SelectPrev()
    } else if a.historyIndex < len(a.inputHistory) {
```

在 Chat mode 的 switch 中，在 CtrlC 之前（第一优先级）添加 Esc 关闭菜单：

```go
case event.Key() == tcell.KeyEsc:
    if a.suggestionMenu.Visible() {
        a.hideSuggestions()
        return nil
    }
```

替换 ↓ 键处理中的 suggestion 部分：
```go
case event.Key() == tcell.KeyDown:
    if len(a.suggestionCmds) > 0 {
        idx := a.suggestionList.GetCurrentItem()
        if idx < a.suggestionList.GetItemCount()-1 {
            a.suggestionList.SetCurrentItem(idx + 1)
        }
    } else if a.historyIndex > 0 {
```

为：
```go
case event.Key() == tcell.KeyDown:
    if a.suggestionMenu.Visible() {
        a.suggestionMenu.SelectNext()
    } else if a.historyIndex > 0 {
```

- [ ] **Step 6: 更新 sendMessage 中对 suggestion 的引用**

在 `sendMessage()` 中找到 `a.hideSuggestions()`，确保其仍被调用（应该在 `sendMessage` 开头被调用，清理建议菜单）：
```go
func (a *App) sendMessage() {
	text := a.composer.GetInput()
	if strings.TrimSpace(text) == "" {
		return
	}
	a.inputHistory = append(a.inputHistory, text)
	a.historyIndex = 0
	a.hideSuggestions()  // 此行应该已经存在
    ...
```

- [ ] **Step 7: 构建验证**

```bash
go build ./cmd/agent/
```

Expected: 无编译错误

- [ ] **Step 8: Commit**

```bash
git add internal/ui/app.go
git commit -m "refactor: replace suggestionList with SuggestionMenu, show all commands"
```

---

### Task 4: 更新 App 测试

**Files:**
- Modify: `internal/ui/app_test.go`

- [ ] **Step 1: 更新 `TestSlashPassesThrough` 确保 `/` 仍然可以正常输入**

现有的测试应该仍然通过，因为 `/` 的输入由 composer 的 TextArea 处理（`return event` 放行）。

- [ ] **Step 2: 添加斜杠菜单集成测试**

```go
func TestSlashShowsAllCommands(t *testing.T) {
	a := NewApp()
	if a.commandRegistry == nil {
		t.Fatal("expected command registry")
	}
	a.composer.SetInput("/")
	// onComposerChange 会被 TextArea.SetChangedFunc 触发
	a.onComposerChange("/")
	if !a.suggestionMenu.Visible() {
		t.Fatal("expected suggestion menu visible when typing /")
	}
	if a.suggestionMenu.Height() <= 2 {
		t.Fatal("expected menu with items")
	}
}

func TestSlashFiltering(t *testing.T) {
	a := NewApp()
	a.composer.SetInput("/Se")
	a.onComposerChange("/Se")
	if !a.suggestionMenu.Visible() {
		t.Fatal("expected menu visible")
	}
	sel := a.suggestionMenu.Selected()
	if sel == nil || sel.Name != "Search" {
		t.Fatalf("expected first match 'Search', got %v", sel)
	}
}

func TestSlashEnterFillsCommand(t *testing.T) {
	a := NewApp()
	a.onComposerChange("/Se")
	// Enter 回填
	ev := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Enter consumed (nil)")
	}
	if !strings.HasPrefix(a.composer.GetInput(), "/Search ") {
		t.Fatalf("expected composer to contain '/Search ', got %q", a.composer.GetInput())
	}
	if a.suggestionMenu.Visible() {
		t.Fatal("expected menu hidden after fill")
	}
}

func TestSlashEscHides(t *testing.T) {
	a := NewApp()
	a.onComposerChange("/")
	if !a.suggestionMenu.Visible() {
		t.Fatal("expected menu visible")
	}
	// 模拟 Esc 关闭
	a.hideSuggestions()
	if a.suggestionMenu.Visible() {
		t.Fatal("expected menu hidden after Esc")
	}
}

func TestSlashNoMatchHides(t *testing.T) {
	a := NewApp()
	a.onComposerChange("/zzz_nonexistent")
	if a.suggestionMenu.Visible() {
		t.Fatal("expected menu hidden when no match")
	}
}

func TestSlashNoSlashHides(t *testing.T) {
	a := NewApp()
	a.onComposerChange("/")
	if !a.suggestionMenu.Visible() {
		t.Fatal("expected menu visible with /")
	}
	a.onComposerChange("hello")
	if a.suggestionMenu.Visible() {
		t.Fatal("expected menu hidden without / prefix")
	}
}

func TestSlashIncludesBuiltins(t *testing.T) {
	a := NewApp()
	a.onComposerChange("/")
	// 应该包含内置命令
	cmds := a.suggestionMenu.Filtered()
	found := false
	for _, c := range cmds {
		if c.Name == "New Session" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'New Session' builtin in suggestion results")
	}
}
```

- [ ] **Step 3: 确认已有测试不受影响**

检查所有现有测试中引用 `a.suggestionList` 或 `a.suggestionCmds` 的地方并更新：

在 `TestSendMessage` 及其相关测试中，验证 `a.composer.GetInput() == ""` 等行为不变。

在 `TestSlashPassesThrough` — 确认 `/` 字符正常输入到 composer。

- [ ] **Step 4: 运行全部测试**

```bash
go test ./... -count=1
```

Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/app_test.go
git commit -m "test: add integration tests for slash command autocomplete"
```

---

### Task 5: 最终验证

- [ ] **Step 1: 完整构建**

```bash
go build ./cmd/agent/
```

- [ ] **Step 2: 完整测试**

```bash
go test ./... -count=1 -v 2>&1 | tail -50
```

Expected: 所有测试通过，无 panic

- [ ] **Step 3: 最终 commit（如果还有未提交的修改）**

```bash
git status
# 如果有未提交的修改文件
git add -A
git commit -m "chore: final cleanup for slash command autocomplete"
```
