# TView Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace Bubble Tea v2 / Bubbles v2 / Lipgloss v2 with rivo/tview (tcell-backed) to fix Chinese IME input.

**Architecture:** tview.Application with tview.Pages root containing a vertical Flex layout (StatusBar→ChatPanel→Composer→TabDock) and overlay pages for Search and Help. Sub-packages (composer, status, tabbar) each expose a tview.Primitive that the main App assembles.

**Tech Stack:** github.com/rivo/tview (pulls in github.com/gdamore/tcell/v2), chroma (syntax highlighting, unchanged)

---

## File Structure

```
cmd/agent/main.go            — Create and run ui.App (rewrite)
internal/ui/
├── app.go                   — App struct, Application + Pages + Flex + InputCapture (NEW, replaces ui.go)
├── keymap.go                — tcell/tview key bindings (rewrite)
├── chatpanel.go             — tview.TextView wrapper (rewrite)
├── session.go               — Rewrite: remove lipgloss, use tview color tags
├── session_render_test.go   — UNCHANGED
├── session_test.go          — UNCHANGED
├── theme.go                 — UNCHANGED
├── theme_service.go         — UNCHANGED
├── theme_test.go            — UNCHANGED
├── ui.go                    — DELETED (replaced by app.go)
├── ui_test.go               — DELETED (replaced by app_test.go)
├── composer/
│   ├── composer.go          — tview.TextArea wrapper (rewrite)
│   └── composer_test.go     — Update for new API
├── status/
│   ├── status.go            — tview.TextView wrapper (rewrite)
│   └── status_test.go       — Update for new API
├── tabbar/
│   ├── tabbar.go            — tview.Box custom Draw (rewrite)
│   └── tabbar_test.go       — Update for new API
├── syntax/
│   ├── syntax.go            — UNCHANGED
│   └── syntax_test.go       — UNCHANGED
├── color/
│   ├── color.go             — NEW: hex→tcell.Color conversion
│   └── color_test.go        — NEW: tests for color conversion
```

**Deleted:** `internal/ui/editor/` (entire package), `internal/ui/ui.go`, `internal/ui/ui_test.go`

---

### Task 1: Add tview dependency

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Run go get to add tview**

```bash
cd G:\mllm\agent5
go get github.com/rivo/tview@latest
```

This adds tview and its tcell/v2 dependency to go.mod.

- [ ] **Step 2: Verify it resolves**

```bash
go list -m github.com/rivo/tview
```

Expected: prints the version, no error.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add rivo/tview dependency"
```

---

### Task 2: Create color conversion utility

**Files:**
- Create: `internal/ui/color/color.go`
- Create: `internal/ui/color/color_test.go`

- [ ] **Step 1: Create color package**

Create `internal/ui/color/color.go`:

```go
package color

import "github.com/gdamore/tcell/v2"

func HexToTCell(hex string) tcell.Color {
	if len(hex) == 0 {
		return tcell.ColorDefault
	}
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return tcell.ColorDefault
	}
	return tcell.NewHexColor(int32(
		(hexDigit(hex[0])<<20 |
			hexDigit(hex[1])<<16 |
			hexDigit(hex[2])<<12 |
			hexDigit(hex[3])<<8 |
			hexDigit(hex[4])<<4 |
			hexDigit(hex[5])),
	))
}

func hexDigit(c byte) int32 {
	switch {
	case c >= '0' && c <= '9':
		return int32(c - '0')
	case c >= 'a' && c <= 'f':
		return int32(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int32(c - 'A' + 10)
	default:
		return 0
	}
}
```

- [ ] **Step 2: Create tests**

Create `internal/ui/color/color_test.go`:

```go
package color

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestHexToTCell_Black(t *testing.T) {
	c := HexToTCell("#000000")
	if c == tcell.ColorDefault {
		t.Fatal("expected a valid color")
	}
}

func TestHexToTCell_White(t *testing.T) {
	c := HexToTCell("#ffffff")
	if c == tcell.ColorDefault {
		t.Fatal("expected a valid color")
	}
}

func TestHexToTCell_Red(t *testing.T) {
	c := HexToTCell("#ff0000")
	if c == tcell.ColorDefault {
		t.Fatal("expected a valid color")
	}
}

func TestHexToTCell_Empty(t *testing.T) {
	c := HexToTCell("")
	if c != tcell.ColorDefault {
		t.Fatal("expected default for empty string")
	}
}

func TestHexToTCell_Short(t *testing.T) {
	c := HexToTCell("#fff")
	if c != tcell.ColorDefault {
		t.Fatal("expected default for short string")
	}
}
```

- [ ] **Step 3: Run tests**

```bash
cd G:\mllm\agent5
go test ./internal/ui/color/ -v
```

Expected: all tests PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/color/
git commit -m "feat: add hex to tcell color conversion utility"
```

---

### Task 3: Rewrite status package

**Files:**
- Modify: `internal/ui/status/status.go`
- Modify: `internal/ui/status/status_test.go`

- [ ] **Step 1: Rewrite status.go**

Replace entire contents:

```go
package status

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type StatusBar struct {
	*tview.TextView
	mode      string
	tasks     int
	connected bool
}

func New() *StatusBar {
	s := &StatusBar{
		TextView: tview.NewTextView(),
	}
	s.SetDynamicColors(true)
	s.SetTextAlign(tview.AlignLeft)
	return s
}

func (s *StatusBar) SetMode(mode string) {
	s.mode = mode
	s.refresh()
}

func (s *StatusBar) SetTasks(n int) {
	s.tasks = n
	s.refresh()
}

func (s *StatusBar) SetConnected(v bool) {
	s.connected = v
	s.refresh()
}

func (s *StatusBar) SetBackgroundColor(color tcell.Color) {
	s.TextView.SetBackgroundColor(color)
}

func (s *StatusBar) refresh() {
	connStr := "●"
	if !s.connected {
		connStr = "○"
	}
	s.SetText(fmt.Sprintf("  %s  Mode: %s  Tasks: %d", connStr, s.mode, s.tasks))
}
```

- [ ] **Step 2: Run test to verify it fails (no tests yet)**

```bash
cd G:\mllm\agent5
go build ./internal/ui/status/
```

Expected: builds successfully with no errors.

- [ ] **Step 3: Rewrite status_test.go**

Replace entire contents:

```go
package status

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
}

func TestSetMode(t *testing.T) {
	s := New()
	s.SetMode("chat")
	if s.mode != "chat" {
		t.Fatalf("expected mode 'chat', got %q", s.mode)
	}
}

func TestSetTasks(t *testing.T) {
	s := New()
	s.SetTasks(3)
	if s.tasks != 3 {
		t.Fatalf("expected tasks 3, got %d", s.tasks)
	}
}

func TestSetConnected(t *testing.T) {
	s := New()
	s.SetConnected(true)
	if !s.connected {
		t.Fatal("expected connected=true")
	}
	s.SetConnected(false)
	if s.connected {
		t.Fatal("expected connected=false after reset")
	}
}

func TestSetBackgroundColor(t *testing.T) {
	s := New()
	s.SetBackgroundColor(tcell.ColorBlue)
}

func TestRefresh(t *testing.T) {
	s := New()
	s.SetMode("chat")
	s.SetTasks(2)
	s.SetConnected(true)
	text := s.GetText(false)
	if text == "" {
		t.Fatal("expected non-empty text after refresh")
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd G:\mllm\agent5
go test ./internal/ui/status/ -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/status/
git commit -m "feat: rewrite status bar with tview.TextView"
```

---

### Task 4: Rewrite composer package

**Files:**
- Modify: `internal/ui/composer/composer.go`
- Modify: `internal/ui/composer/composer_test.go`

- [ ] **Step 1: Rewrite composer.go**

Replace entire contents:

```go
package composer

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Composer struct {
	*tview.TextArea
}

func New() *Composer {
	c := &Composer{
		TextArea: tview.NewTextArea(),
	}
	c.SetWordWrap(true)
	if w := c.GetMaxHeight(); w == 0 || w > 8 {
		c.SetMaxHeight(8)
	}
	return c
}

func (c *Composer) SetInput(s string) {
	c.SetText(s, true)
}

func (c *Composer) GetInput() string {
	return c.GetText()
}

func (c *Composer) ClearInput() {
	c.SetText("", true)
}

func (c *Composer) SetBackgroundColor(color tcell.Color) {
	c.TextArea.SetBackgroundColor(color)
}
```

- [ ] **Step 2: Verify it builds**

```bash
cd G:\mllm\agent5
go build ./internal/ui/composer/
```

Expected: builds successfully.

- [ ] **Step 3: Rewrite composer_test.go**

Replace entire contents:

```go
package composer

import (
	"testing"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
}

func TestSetAndGetInput(t *testing.T) {
	c := New()
	c.SetInput("hello")
	if got := c.GetInput(); got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}

func TestClearInput(t *testing.T) {
	c := New()
	c.SetInput("hello")
	c.ClearInput()
	if got := c.GetInput(); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestClearInputEmpty(t *testing.T) {
	c := New()
	c.ClearInput()
	if got := c.GetInput(); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestHeightBounds(t *testing.T) {
	c := New()
	h := c.GetMaxHeight()
	if h == 0 || h > 8 {
		t.Fatalf("expected MaxHeight in [1,8], got %d", h)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd G:\mllm\agent5
go test ./internal/ui/composer/ -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/composer/
git commit -m "feat: rewrite composer with tview.TextArea"
```

---

### Task 5: Rewrite tabbar package

**Files:**
- Modify: `internal/ui/tabbar/tabbar.go`
- Modify: `internal/ui/tabbar/tabbar_test.go`

- [ ] **Step 1: Rewrite tabbar.go**

Replace entire contents:

```go
package tabbar

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Tab struct {
	ID    string
	Label string
}

type TabDock struct {
	*tview.Box
	tabs       []Tab
	active     int
	bgColor    tcell.Color
	activeFg   tcell.Color
	activeBg   tcell.Color
	inactiveFg tcell.Color
	inactiveBg tcell.Color
	onClick    func(idx int)
}

func New() *TabDock {
	t := &TabDock{
		Box:        tview.NewBox(),
		tabs:       []Tab{},
		active:     0,
		bgColor:    tcell.ColorDefault,
		activeFg:   tcell.ColorWhite,
		activeBg:   tcell.ColorBlue,
		inactiveFg: tcell.ColorGray,
		inactiveBg: tcell.ColorDefault,
	}
	t.SetDrawFunc(t.draw)
	return t
}

func (t *TabDock) AddTab(tab Tab) {
	t.tabs = append(t.tabs, tab)
}

func (t *TabDock) RemoveTab(index int) {
	if index < 0 || index >= len(t.tabs) {
		return
	}
	t.tabs = append(t.tabs[:index], t.tabs[index+1:]...)
	if t.active >= len(t.tabs) && len(t.tabs) > 0 {
		t.active = len(t.tabs) - 1
	}
}

func (t *TabDock) UpdateTab(index int, label string) {
	if index < 0 || index >= len(t.tabs) {
		return
	}
	t.tabs[index].Label = label
}

func (t *TabDock) SetActive(index int) {
	if index >= 0 && index < len(t.tabs) {
		t.active = index
	}
}

func (t *TabDock) ActiveIndex() int {
	return t.active
}

func (t *TabDock) TabCount() int {
	return len(t.tabs)
}

func (t *TabDock) SetOnClick(fn func(idx int)) {
	t.onClick = fn
	t.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick && t.onClick != nil {
			x, y := event.Position()
			_, _, w, _ := t.GetRect()
			// Calculate which tab was clicked based on x position
			idx := t.tabAtX(x, w)
			if idx >= 0 {
				t.onClick(idx)
				return tview.MouseActionConsumed, nil
			}
			return tview.MouseActionConsumed, nil
		}
		return action, event
	})
}

func (t *TabDock) tabAtX(x, width int) int {
	if len(t.tabs) == 0 || width <= 0 {
		return -1
	}
	// Calculate tab width distribution
	labelTotal := 0
	for _, tab := range t.tabs {
		labelTotal += len(tab.Label) + 4 // 2 spaces padding each side
	}
	remaining := width - labelTotal
	extraPerTab := 0
	if len(t.tabs) > 0 && remaining > 0 {
		extraPerTab = remaining / len(t.tabs)
	}
	cx := 0
	for i, tab := range t.tabs {
		tabW := len(tab.Label) + 4 + extraPerTab
		if x >= cx && x < cx+tabW {
			return i
		}
		cx += tabW
	}
	return -1
}

func (t *TabDock) SetColors(activeFg, activeBg, inactiveFg, inactiveBg tcell.Color) {
	t.activeFg = activeFg
	t.activeBg = activeBg
	t.inactiveFg = inactiveFg
	t.inactiveBg = inactiveBg
}

func (t *TabDock) draw(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
	if len(t.tabs) == 0 {
		return x, y, width, height
	}

	// Fill background
	for cy := y; cy < y+height; cy++ {
		for cx := x; cx < x+width; cx++ {
			screen.SetContent(cx, cy, ' ', nil, tcell.StyleDefault.Background(t.bgColor))
		}
	}

	// Calculate tab widths
	labelTotal := 0
	for _, tab := range t.tabs {
		labelTotal += len(tab.Label) + 4
	}
	remaining := width - labelTotal
	extraPerTab := 0
	if len(t.tabs) > 0 && remaining > 0 {
		extraPerTab = remaining / len(t.tabs)
	}

	cx := x
	for i, tab := range t.tabs {
		tabW := len(tab.Label) + 4 + extraPerTab
		fg, bg := t.inactiveFg, t.inactiveBg
		if i == t.active {
			fg, bg = t.activeFg, t.activeBg
		}
		style := tcell.StyleDefault.Foreground(fg).Background(bg)
		label := " " + tab.Label + " "
		for j, ch := range label {
			screen.SetContent(cx+j, y, ch, nil, style)
		}
		cx += tabW
	}

	return x, y, width, height
}
```

- [ ] **Step 2: Verify it builds**

```bash
cd G:\mllm\agent5
go build ./internal/ui/tabbar/
```

Expected: builds successfully.

- [ ] **Step 3: Rewrite tabbar_test.go**

Replace entire contents:

```go
package tabbar

import (
	"testing"
)

func TestNew(t *testing.T) {
	tb := New()
	if tb == nil {
		t.Fatal("New() returned nil")
	}
}

func TestAddTab(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "Session 1"})
	tb.AddTab(Tab{ID: "2", Label: "Session 2"})
	if tb.TabCount() != 2 {
		t.Fatalf("expected 2 tabs, got %d", tb.TabCount())
	}
}

func TestRemoveTab(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.AddTab(Tab{ID: "2", Label: "S2"})
	tb.RemoveTab(0)
	if tb.TabCount() != 1 {
		t.Fatalf("expected 1 tab after remove, got %d", tb.TabCount())
	}
	if tb.tabs[0].ID != "2" {
		t.Fatalf("expected remaining tab ID '2', got %q", tb.tabs[0].ID)
	}
}

func TestRemoveTabOutOfRange(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.RemoveTab(5) // should not panic
	if tb.TabCount() != 1 {
		t.Fatalf("expected 1 tab, got %d", tb.TabCount())
	}
	tb.RemoveTab(-1) // should not panic
	if tb.TabCount() != 1 {
		t.Fatalf("expected 1 tab, got %d", tb.TabCount())
	}
}

func TestUpdateTab(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "Old"})
	tb.UpdateTab(0, "New")
	if tb.tabs[0].Label != "New" {
		t.Fatalf("expected label 'New', got %q", tb.tabs[0].Label)
	}
}

func TestUpdateTabOutOfRange(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.UpdateTab(3, "X") // should not panic
	tb.UpdateTab(-1, "X") // should not panic
	if tb.tabs[0].Label != "S1" {
		t.Fatalf("expected unchanged label 'S1', got %q", tb.tabs[0].Label)
	}
}

func TestSetActive(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.AddTab(Tab{ID: "2", Label: "S2"})
	tb.SetActive(1)
	if tb.ActiveIndex() != 1 {
		t.Fatalf("expected active index 1, got %d", tb.ActiveIndex())
	}
}

func TestSetActiveOutOfRange(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.SetActive(5) // should not panic
	tb.SetActive(-1) // should not panic
}

func TestRemoveTabAdjustsActive(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.AddTab(Tab{ID: "2", Label: "S2"})
	tb.SetActive(0)
	tb.RemoveTab(0)
	// active should still be 0 (only tab left)
	if tb.ActiveIndex() != 0 {
		t.Fatalf("expected active index 0 after removal, got %d", tb.ActiveIndex())
	}
}

func TestSetColors(t *testing.T) {
	tb := New()
	// Just verify no panic
	tb.SetColors(tcell.ColorWhite, tcell.ColorBlue, tcell.ColorGray, tcell.ColorDefault)
}

func TestTabAtX(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "A"})
	tb.AddTab(Tab{ID: "2", Label: "B"})
	// tab 0 should start at x=0
	idx := tb.tabAtX(0, 20)
	if idx != 0 {
		t.Fatalf("expected tab index 0 at x=0, got %d", idx)
	}
}
```

Note: The test imports `tcell` — verify import path matches:

```go
import "github.com/gdamore/tcell/v2"
```

- [ ] **Step 4: Run tests**

```bash
cd G:\mllm\agent5
go test ./internal/ui/tabbar/ -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/tabbar/
git commit -m "feat: rewrite tabbar with tview.Box custom Draw"
```

---

### Task 6: Rewrite session.go (remove lipgloss)

**Files:**
- Modify: `internal/ui/session.go`

`session.go` uses lipgloss extensively for rendering timestamps, badges, content, thinking blocks. Replace all lipgloss formatting with tview color tags (`[fg:bg:attr]` syntax) that work with `tview.TextView.SetDynamicColors(true)`.

- [ ] **Step 1: Rewrite `renderMessageToBuilder`**

Remove all lipgloss imports and replace with `fmt.Sprintf` color tags:

```go
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/example/agent-tui/internal/ui/syntax"
)

// ... Role, Thinking, Message, Session structs and methods unchanged ...

// RenderMessages renders all messages as a formatted string for viewport content.
func (s *Session) RenderMessages(width int, theme ColorPalette) string {
	if len(s.Messages) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, msg := range s.Messages {
		renderMessageToBuilder(&sb, msg, width, theme)
	}
	return sb.String()
}

func renderMessageToBuilder(sb *strings.Builder, msg Message, width int, theme ColorPalette) {
	ts := fmt.Sprintf("[%s]%s[-]", theme.Timestamp, msg.Timestamp.Format("15:04"))

	switch msg.Role {
	case RoleUser:
		badge := fmt.Sprintf("[%s:%s:b] \U0001f5e3 [:-:-]", theme.UserFg, theme.UserBg)
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 3)

	case RoleAssistant:
		badge := fmt.Sprintf("[%s:%s:b] \U0001f47e Chan [:-:-]", theme.AssistantFg, theme.AssistantBg)
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")

		if msg.Thinking != nil {
			renderThinkingBlock(sb, msg.Thinking, width, theme)
		}
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 1)

	case RoleSystem:
		badge := fmt.Sprintf("[%s:%s:b] \u2699 sys [:-:-]", theme.SystemFg, theme.SystemBg)
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 1)
	}

	sb.WriteString("\n")
}

func renderContent(sb *strings.Builder, content string, collapsed bool, width int, theme ColorPalette, paddingLeft int) {
	padding := strings.Repeat(" ", paddingLeft)
	textPrefix := fmt.Sprintf("[%s]", theme.Text)
	textSuffix := "[-]"

	if collapsed {
		lines := strings.Split(content, "\n")
		totalLines := len(lines)
		if totalLines <= 3 {
			sb.WriteString(textPrefix + padding + content + textSuffix)
			return
		}
		preview := strings.Join(lines[:3], "\n")
		sb.WriteString(textPrefix + padding + preview + textSuffix)
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("[%s]%s\u25bc [expand %d lines][-]", theme.TextMuted, padding, totalLines-3))
		return
	}

	highlighted := safeHighlight(content)
	sb.WriteString(textPrefix + padding + highlighted + textSuffix)
}

func safeHighlight(content string) string {
	if hasCJK(content) {
		return content
	}
	return syntax.Highlight(content, "")
}

func foldContent(content string, theme ColorPalette) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)
	if totalLines <= 3 {
		return fmt.Sprintf("[%s]%s[-]", theme.Text, content)
	}
	preview := strings.Join(lines[:3], "\n")
	return fmt.Sprintf("[%s]%s[-]\n[%s]\u25bc [expand %d lines][-]", theme.Text, preview, theme.TextMuted, totalLines-3)
}

func renderThinkingBlock(sb *strings.Builder, t *Thinking, width int, theme ColorPalette) {
	arrow := "\u25b6"
	if t.Expanded {
		arrow = "\u25bc"
	}

	charCount := len([]rune(t.Content))
	durationStr := ""
	if t.Duration > 0 {
		durationStr = fmt.Sprintf(" \u00b7 %.1fs", t.Duration.Seconds())
	}

	sb.WriteString(fmt.Sprintf("[%s]%s thinking %d chars%s[-]", theme.ThinkingFg, arrow, charCount, durationStr))
	sb.WriteString("\n")

	if t.Expanded {
		lines := strings.Split(t.Content, "\n")
		for _, line := range lines {
			sb.WriteString(fmt.Sprintf("[%s:%s] \u2502 %s[-:-]", theme.ThinkingFg, theme.ThinkingBg, line))
			sb.WriteString("\n")
		}
	}
}

func hasCJK(s string) bool {
	for _, r := range s {
		if (r >= 0x4E00 && r <= 0x9FFF) ||
			(r >= 0x3400 && r <= 0x4DBF) ||
			(r >= 0x20000 && r <= 0x2A6DF) ||
			(r >= 0xF900 && r <= 0xFAFF) ||
			(r >= 0x3040 && r <= 0x309F) ||
			(r >= 0x30A0 && r <= 0x30FF) ||
			(r >= 0xAC00 && r <= 0xD7AF) {
			return true
		}
	}
	return false
}
```

Note: `foldContent` is still used by... let me verify. It was called from `renderContent` for the collapsed case. In the rewrite above, I inlined that logic into `renderContent`. So `foldContent` can remain as-is or be deleted. Keep it in case other callers exist.

- [ ] **Step 2: Verify build**

```bash
cd G:\mllm\agent5
go build ./internal/ui/ 2>&1
```

Expected: compiles without lipgloss import errors (may still have other errors due to old ui.go — that's OK).

- [ ] **Step 3: Run session tests**

```bash
cd G:\mllm\agent5
go test ./internal/ui/ -run TestRender -v
```

Expected: all session rendering tests PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/session.go
git commit -m "feat: replace lipgloss with tview color tags in session rendering"
```


### Task 7: Rewrite keymap

**Files:**
- Modify: `internal/ui/keymap.go`
- Modify: `internal/ui/keymap_test.go`

- [ ] **Step 1: Rewrite keymap.go**

Replace entire contents. Note: tview.InputCapture receives `*tcell.EventKey`, so bindings use tcell key constants.

```go
package ui

import "github.com/gdamore/tcell/v2"

type KeyMap struct {
	Quit           tcell.Key
	NewSession     tcell.Key
	CloseSession   tcell.Key
	RenameSession  tcell.Key
	NextSession    tcell.Key
	PrevSession    tcell.Key
	ToggleThinking tcell.Key
	ToggleCollapse tcell.Key
	Search         tcell.Key
	ToggleTheme    tcell.Key
	ShowHelp       tcell.Key
	ScrollUp       tcell.Key
	ScrollDown     tcell.Key
	ScrollTop      tcell.Key
	ScrollBottom   tcell.Key
	SendMessage    tcell.Key
	// Modifiers for SendMessage: Ctrl+Enter
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:           tcell.KeyRune, // 'q' checked via rune
		NewSession:     tcell.KeyRune, // 'n'
		CloseSession:   tcell.KeyRune, // 'w'
		RenameSession:  tcell.KeyRune, // 'r'
		NextSession:    tcell.KeyRune, // '.'
		PrevSession:    tcell.KeyRune, // ','
		ToggleThinking: tcell.KeyRune, // 't'
		ToggleCollapse: tcell.KeyRune, // 'y'
		Search:         tcell.KeyRune, // '/'
		ToggleTheme:    tcell.KeyRune, // 'T'
		ShowHelp:       tcell.KeyRune, // '?'
		ScrollUp:       tcell.KeyPgUp,
		ScrollDown:     tcell.KeyPgDn,
		ScrollTop:      tcell.KeyRune, // 'g'
		ScrollBottom:   tcell.KeyRune, // 'G'
		SendMessage:    tcell.KeyEnter, // Ctrl+Enter checked via Rune
	}
}

func (k KeyMap) ShortHelp() []string {
	return []string{
		"Ctrl+Enter: Send",
		"Alt+N: New",
		"Alt+W: Close",
		"?: Help",
	}
}

func (k KeyMap) FullHelp() []string {
	return []string{
		"Ctrl+Enter   Send message",
		"Alt+N        New session",
		"Alt+W        Close session",
		"Alt+R        Rename session",
		"Alt+.        Next session",
		"Alt+,        Previous session",
		"Alt+T        Toggle thinking",
		"Alt+Y        Toggle collapse",
		"Ctrl+F       Search",
		"Alt+Shift+T  Toggle theme",
		"PgUp/PgDn    Scroll chat",
		"g/G          Scroll top/bottom",
		"q            Quit",
	}
}
```

- [ ] **Step 2: Rewrite keymap_test.go**

```go
package ui

import (
	"testing"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()
	if km.Quit == 0 {
		t.Fatal("expected non-zero key mapping")
	}
	if km.SendMessage == 0 {
		t.Fatal("expected non-zero key mapping")
	}
}

func TestShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.ShortHelp()
	if len(help) == 0 {
		t.Fatal("expected non-empty short help")
	}
}

func TestFullHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.FullHelp()
	if len(help) == 0 {
		t.Fatal("expected non-empty full help")
	}
}
```

- [ ] **Step 3: Run tests**

Note: keymap_test.go is in package ui, which currently imports ui.go (still Bubble Tea). This test might not compile yet if it references types from the old ui.go. Let's just verify the test file parses correctly.

```bash
cd G:\mllm\agent5
go vet ./internal/ui/keymap.go
```

Expected: no errors (or import cycle errors from ui.go — that's expected, will be fixed in Task 8).

- [ ] **Step 4: Commit**

```bash
git add internal/ui/keymap.go internal/ui/keymap_test.go
git commit -m "feat: rewrite keymap with tcell key constants"
```

---

### Task 8: Rewrite chatpanel

**Files:**
- Modify: `internal/ui/chatpanel.go`

- [ ] **Step 1: Rewrite chatpanel.go**

```go
package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ChatPanel struct {
	*tview.TextView
	session *Session
	searchQuery string
	searchResults []int
	currentMatch int
}

func NewChatPanel() *ChatPanel {
	c := &ChatPanel{
		TextView: tview.NewTextView(),
	}
	c.SetDynamicColors(true)
	c.SetScrollable(true)
	c.SetWordWrap(true)
	c.SetRegions(true)
	return c
}

func (c *ChatPanel) SetSession(session *Session) {
	c.session = session
	c.refresh()
}

func (c *ChatPanel) SetContent(content string) {
	c.SetText(content)
	c.ScrollToEnd()
}

func (c *ChatPanel) ScrollUp(lines int) {
	row, _ := c.GetScrollOffset()
	c.ScrollTo(row-lines, 0)
}

func (c *ChatPanel) ScrollDown(lines int) {
	row, _ := c.GetScrollOffset()
	c.ScrollTo(row+lines, 0)
}

func (c *ChatPanel) ScrollToTop() {
	c.ScrollTo(0, 0)
}

func (c *ChatPanel) ScrollToBottom() {
	c.ScrollToEnd()
}

func (c *ChatPanel) EnterSearch() {
	c.searchQuery = ""
	c.searchResults = nil
	c.currentMatch = -1
}

func (c *ChatPanel) ExitSearch() {
	c.searchQuery = ""
	c.searchResults = nil
	c.currentMatch = -1
	c.refresh()
}

func (c *ChatPanel) IsSearchMode() bool {
	// Search mode is tracked by the App, not the panel
	return false
}

func (c *ChatPanel) SetSearchQuery(query string) {
	c.searchQuery = query
	c.findMatches()
	c.Highlight()
	if len(c.searchResults) > 0 {
		c.currentMatch = 0
		c.Highlight(c.matchRegionID(c.searchResults[0]))
		c.ScrollTo(c.searchResults[0], 0)
	}
}

func (c *ChatPanel) NextMatch() {
	if len(c.searchResults) == 0 {
		return
	}
	c.currentMatch = (c.currentMatch + 1) % len(c.searchResults)
	c.Highlight()
	c.Highlight(c.matchRegionID(c.searchResults[c.currentMatch]))
	c.ScrollTo(c.searchResults[c.currentMatch], 0)
}

func (c *ChatPanel) PrevMatch() {
	if len(c.searchResults) == 0 {
		return
	}
	c.currentMatch--
	if c.currentMatch < 0 {
		c.currentMatch = len(c.searchResults) - 1
	}
	c.Highlight()
	c.Highlight(c.matchRegionID(c.searchResults[c.currentMatch]))
	c.ScrollTo(c.searchResults[c.currentMatch], 0)
}

func (c *ChatPanel) MatchCount() int {
	return len(c.searchResults)
}

func (c *ChatPanel) CurrentMatch() int {
	return c.currentMatch
}

func (c *ChatPanel) ApplyTheme(colors ColorPalette) {
	c.SetBackgroundColor(cellColor(colors.Background))
}

func (c *ChatPanel) refresh() {
	if c.session == nil {
		c.SetText("")
		return
	}
	c.SetText(c.session.RenderMessages())
	c.ScrollToEnd()
}

func (c *ChatPanel) findMatches() {
	c.searchResults = nil
	if c.searchQuery == "" || c.session == nil {
		return
	}
	query := strings.ToLower(c.searchQuery)
	content := strings.ToLower(c.GetText(false))
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, query) {
			c.searchResults = append(c.searchResults, i)
		}
	}
}

func (c *ChatPanel) matchRegionID(line int) string {
	return "match-" + itoa(line)
}
```

Also add a helper to `chatpanel.go` or create a small util:

```go
func cellColor(hex string) tcell.Color {
	if len(hex) == 0 {
		return tcell.ColorDefault
	}
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return tcell.ColorDefault
	}
	r := parseHex(hex[0:2])
	g := parseHex(hex[2:4])
	b := parseHex(hex[4:6])
	return tcell.NewHexColor(int32(r)<<16 | int32(g)<<8 | int32(b))
}

func parseHex(s string) int32 {
	if len(s) != 2 {
		return 0
	}
	return hexVal(s[0])<<4 | hexVal(s[1])
}

func hexVal(c byte) int32 {
	switch {
	case c >= '0' && c <= '9':
		return int32(c - '0')
	case c >= 'a' && c <= 'f':
		return int32(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int32(c - 'A' + 10)
	default:
		return 0
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
```

- [ ] **Step 2: Verify build**

```bash
cd G:\mllm\agent5
go vet ./internal/ui/chatpanel.go
```

Expected: parses with no errors (may not fully compile due to ui.go dependency — that's OK).

- [ ] **Step 3: Commit**

```bash
git add internal/ui/chatpanel.go
git commit -m "feat: rewrite chatpanel with tview.TextView"
```

---

### Task 9: Create main App (replaces ui.go)

**Files:**
- Delete: `internal/ui/ui.go`
- Create: `internal/ui/app.go`
- Delete: `internal/ui/ui_test.go`
- Create: `internal/ui/app_test.go`

This is the largest task — creates the main Application, Pages layout, InputCapture, and all component wiring.

- [ ] **Step 1: Create app.go**

```go
package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/rivo/tview"

	"github.com/example/agent-tui/internal/ui/composer"
	"github.com/example/agent-tui/internal/ui/status"
	"github.com/example/agent-tui/internal/ui/tabbar"
)

type AppMode int

const (
	ModeChat AppMode = iota
	ModeSearch
	ModeHelp
)

type App struct {
	*tview.Application
	pages      *tview.Pages
	chatFlex   *tview.Flex
	statusBar  *status.StatusBar
	chatPanel  *ChatPanel
	composer   *composer.Composer
	tabDock    *tabbar.TabDock
	searchInput *tview.InputField
	helpView   *tview.TextView
	sessions   []*Session
	activeSession int
	mode       AppMode
	isLoading  bool
	keyMap     KeyMap
	themeService *ThemeService
	aiAssistant interface{}
}

func NewApp(aiAssistant interface{}) *App {
	a := &App{
		Application:   tview.NewApplication(),
		pages:         tview.NewPages(),
		statusBar:     status.New(),
		chatPanel:     NewChatPanel(),
		composer:      composer.New(),
		tabDock:       tabbar.New(),
		searchInput:   tview.NewInputField(),
		helpView:      tview.NewTextView(),
		sessions:      []*Session{},
		activeSession: -1,
		mode:          ModeChat,
		keyMap:        DefaultKeyMap(),
		themeService:  NewThemeService(),
		aiAssistant:   aiAssistant,
	}

	// Wire tab click
	a.tabDock.SetOnClick(func(idx int) {
		a.switchToSession(idx)
	})

	// Setup search input
	a.searchInput.SetLabel("/")
	a.searchInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			a.chatPanel.NextMatch()
		case tcell.KeyEsc:
			a.exitSearch()
		case tcell.KeyBackspace2, tcell.KeyBackspace:
			// handled by InputCapture
		}
	})
	a.searchInput.SetChangedFunc(func(text string) {
		a.chatPanel.SetSearchQuery(text)
	})

	// Setup help view
	a.helpView.SetDynamicColors(true)
	a.helpView.SetText(strings.Join(a.keyMap.FullHelp(), "\n"))
	a.helpView.SetTextAlign(tview.AlignLeft)
	a.helpView.SetBorder(true)
	a.helpView.SetTitle(" Help ")

	// Build layout
	chatFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	chatFlex.AddItem(a.statusBar, 1, 0, false)
	chatFlex.AddItem(a.chatPanel, 0, 1, false)
	chatFlex.AddItem(a.composer, 0, 0, true)  // auto-height, focusable
	chatFlex.AddItem(a.tabDock, 1, 0, false)

	// Search page
	searchFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	searchFlex.AddItem(nil, 0, 1, false)
	searchInner := tview.NewFlex().SetDirection(tview.FlexRow)
	searchInner.AddItem(a.searchInput, 1, 0, true)
	searchFlex.AddItem(searchInner, 3, 0, true)
	searchFlex.AddItem(nil, 0, 1, false)
	searchPage := tview.NewFlex().SetDirection(tview.FlexColumn)
	searchPage.AddItem(nil, 0, 1, false)
	searchPage.AddItem(searchFlex, 40, 0, true)
	searchPage.AddItem(nil, 0, 1, false)

	// Help page
	helpFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	helpFlex.AddItem(nil, 0, 1, false)
	helpInner := tview.NewFlex().SetDirection(tview.FlexRow)
	helpInner.AddItem(nil, 0, 1, false)
	helpInner.AddItem(a.helpView, 0, 1, true)
	helpInner.AddItem(nil, 0, 1, false)
	helpFlex.AddItem(helpInner, 50, 0, true)
	helpFlex.AddItem(nil, 0, 1, false)

	a.pages.AddPage("chat", chatFlex, true, true)
	a.pages.AddPage("search", searchPage, true, false)
	a.pages.AddPage("help", helpFlex, true, false)

	a.SetRoot(a.pages, true)
	a.SetInputCapture(a.handleInput)
	a.SetFocus(a.composer)

	return a
}

func (a *App) handleInput(event *tcell.EventKey) *tcell.EventKey {
	// Handle based on current mode
	switch a.mode {
	case ModeHelp:
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter || event.Rune() == 'q' {
			a.exitHelp()
			return nil
		}
		return nil
	case ModeSearch:
		switch event.Key() {
		case tcell.KeyEsc:
			a.exitSearch()
			return nil
		case tcell.KeyEnter:
			a.chatPanel.NextMatch()
			return nil
		case tcell.KeyBackspace2, tcell.KeyBackspace:
			text := a.searchInput.GetText()
			if len(text) > 0 {
				a.searchInput.SetText(text[:len(text)-1])
				a.chatPanel.SetSearchQuery(a.searchInput.GetText())
			}
			return nil
		default:
			if event.Rune() != 0 {
				return event // let InputField handle it
			}
			return nil
		}
	}

	// Chat mode — handle global shortcuts
	switch {
	case event.Key() == tcell.KeyCtrlC || (event.Rune() == 'q' && event.Modifiers() == tcell.ModNone):
		a.Stop()
		return nil
	case event.Key() == tcell.KeyCtrlF:
		a.enterSearch()
		return nil
	case event.Rune() == '?' && event.Modifiers() == tcell.ModNone:
		a.enterHelp()
		return nil
	case event.Key() == tcell.KeyCtrlEnter:
		if strings.TrimSpace(a.composer.GetInput()) == "" {
			return nil
		}
		a.sendMessage()
		return nil
	case event.Key() == tcell.KeyPgUp:
		a.chatPanel.ScrollUp(10)
		return nil
	case event.Key() == tcell.KeyPgDn:
		a.chatPanel.ScrollDown(10)
		return nil
	case event.Rune() == 'g' && event.Modifiers() == tcell.ModNone:
		a.chatPanel.ScrollToTop()
		return nil
	case event.Rune() == 'G' && event.Modifiers() == tcell.ModNone:
		a.chatPanel.ScrollToBottom()
		return nil
	case event.Modifiers() == tcell.ModAlt:
		switch event.Rune() {
		case 'n', 'N':
			a.newSession()
			return nil
		case 'w', 'W':
			a.closeSession()
			return nil
		case 'r', 'R':
			a.renameSession()
			return nil
		case '.':
			a.nextSession()
			return nil
		case ',':
			a.prevSession()
			return nil
		case 't', 'T':
			if s := a.activeSessionPtr(); s != nil {
				s.ToggleThinking()
			}
			return nil
		case 'y', 'Y':
			if s := a.activeSessionPtr(); s != nil {
				s.ToggleCollapse()
			}
			return nil
		case 'T', 'S':
			a.themeService.NextTheme()
			a.applyTheme()
			return nil
		}
		// Alt+{1-9} for session switching
		if event.Rune() >= '1' && event.Rune() <= '9' {
			idx := int(event.Rune() - '1')
			if idx < len(a.sessions) {
				a.switchToSession(idx)
			}
			return nil
		}
	}

	return event // pass through to focused element (composer)
}

func (a *App) enterSearch() {
	a.mode = ModeSearch
	a.searchInput.SetText("")
	a.pages.SwitchToPage("search")
	a.SetFocus(a.searchInput)
	a.chatPanel.EnterSearch()
}

func (a *App) exitSearch() {
	a.mode = ModeChat
	a.pages.SwitchToPage("chat")
	a.SetFocus(a.composer)
	a.chatPanel.ExitSearch()
}

func (a *App) enterHelp() {
	a.mode = ModeHelp
	a.helpView.SetText(strings.Join(a.keyMap.FullHelp(), "\n"))
	a.pages.SwitchToPage("help")
	a.SetFocus(a.helpView)
}

func (a *App) exitHelp() {
	a.mode = ModeChat
	a.pages.SwitchToPage("chat")
	a.SetFocus(a.composer)
}

func (a *App) newSession() {
	s := NewSession(uuid.New().String(), "New Session")
	a.sessions = append(a.sessions, s)
	a.tabDock.AddTab(tabbar.Tab{ID: s.ID, Label: "New Session"})
	a.switchToSession(len(a.sessions) - 1)
}

func (a *App) closeSession() {
	if len(a.sessions) <= 1 {
		return
	}
	idx := a.activeSession
	a.sessions = append(a.sessions[:idx], a.sessions[idx+1:]...)
	a.tabDock.RemoveTab(idx)
	if a.activeSession >= len(a.sessions) {
		a.activeSession = len(a.sessions) - 1
	}
	a.chatPanel.SetSession(a.sessions[a.activeSession])
}

func (a *App) renameSession() {
	// Placeholder — could prompt for new name
}

func (a *App) nextSession() {
	if len(a.sessions) == 0 {
		return
	}
	idx := (a.activeSession + 1) % len(a.sessions)
	a.switchToSession(idx)
}

func (a *App) prevSession() {
	if len(a.sessions) == 0 {
		return
	}
	idx := a.activeSession - 1
	if idx < 0 {
		idx = len(a.sessions) - 1
	}
	a.switchToSession(idx)
}

func (a *App) switchToSession(idx int) {
	if idx < 0 || idx >= len(a.sessions) {
		return
	}
	a.activeSession = idx
	a.tabDock.SetActive(idx)
	a.chatPanel.SetSession(a.sessions[idx])
	a.composer.ClearInput()
}

func (a *App) activeSessionPtr() *Session {
	if a.activeSession >= 0 && a.activeSession < len(a.sessions) {
		return a.sessions[a.activeSession]
	}
	return nil
}

func (a *App) sendMessage() {
	text := a.composer.GetInput()
	if strings.TrimSpace(text) == "" {
		return
	}
	if s := a.activeSessionPtr(); s != nil {
		s.AddMessage("user", text)
		// Placeholder: send to AI and get response
		s.AddMessage("assistant", "Echo: "+text)
	}
	a.composer.ClearInput()
	a.chatPanel.SetSession(a.activeSessionPtr())
}

func (a *App) applyTheme() {
	colors := a.themeService.CurrentTheme().Colors
	a.statusBar.SetBackgroundColor(cellColor(colors.Background))
	a.chatPanel.ApplyTheme(colors)
	a.composer.SetBackgroundColor(cellColor(colors.Background))
}
```

- [ ] **Step 2: Build to check compilation**

```bash
cd G:\mllm\agent5
go build ./...
```

Expected: builds with no errors.

- [ ] **Step 3: Create app_test.go**

```go
package ui

import (
	"testing"
)

func TestNewApp(t *testing.T) {
	a := NewApp(nil)
	if a == nil {
		t.Fatal("NewApp returned nil")
	}
	if a.Application == nil {
		t.Fatal("expected non-nil Application")
	}
}

func TestNewAppSessions(t *testing.T) {
	a := NewApp(nil)
	if len(a.sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(a.sessions))
	}
}

func TestNewAppHasKeyMap(t *testing.T) {
	a := NewApp(nil)
	if a.keyMap.Quit == 0 {
		t.Fatal("expected non-zero key map")
	}
}

func TestNewAppHasThemeService(t *testing.T) {
	a := NewApp(nil)
	if a.themeService == nil {
		t.Fatal("expected non-nil theme service")
	}
}

func TestNewSession(t *testing.T) {
	a := NewApp(nil)
	a.newSession()
	if len(a.sessions) != 1 {
		t.Fatalf("expected 1 session after NewSession, got %d", len(a.sessions))
	}
	if a.activeSession != 0 {
		t.Fatalf("expected active session 0, got %d", a.activeSession)
	}
}

func TestSwitchSession(t *testing.T) {
	a := NewApp(nil)
	a.newSession()
	a.newSession()
	a.switchToSession(1)
	if a.activeSession != 1 {
		t.Fatalf("expected active session 1, got %d", a.activeSession)
	}
}

func TestSwitchSessionOutOfRange(t *testing.T) {
	a := NewApp(nil)
	a.newSession()
	a.switchToSession(5) // should not panic
	a.switchToSession(-1) // should not panic
	if a.activeSession != 0 {
		t.Fatalf("expected active session 0, got %d", a.activeSession)
	}
}

func TestNextSession(t *testing.T) {
	a := NewApp(nil)
	a.newSession()
	a.newSession()
	a.switchToSession(0)
	a.nextSession()
	if a.activeSession != 1 {
		t.Fatalf("expected active session 1, got %d", a.activeSession)
	}
	a.nextSession()
	if a.activeSession != 0 {
		t.Fatalf("expected active session 0 (wrap), got %d", a.activeSession)
	}
}

func TestPrevSession(t *testing.T) {
	a := NewApp(nil)
	a.newSession()
	a.newSession()
	a.switchToSession(0)
	a.prevSession()
	if a.activeSession != 1 {
		t.Fatalf("expected active session 1 (wrap), got %d", a.activeSession)
	}
}

func TestCloseSession(t *testing.T) {
	a := NewApp(nil)
	a.newSession()
	a.newSession()
	a.closeSession()
	if len(a.sessions) != 1 {
		t.Fatalf("expected 1 session after close, got %d", len(a.sessions))
	}
}

func TestCloseSessionLastRemaining(t *testing.T) {
	a := NewApp(nil)
	a.newSession()
	a.closeSession() // should not close the last session
	if len(a.sessions) != 1 {
		t.Fatalf("expected 1 session (can't close last), got %d", len(a.sessions))
	}
}

func TestSendMessage(t *testing.T) {
	a := NewApp(nil)
	a.newSession()
	a.composer.SetInput("hello")
	a.sendMessage()
	if a.composer.GetInput() != "" {
		t.Fatal("expected composer cleared after send")
	}
	if s := a.activeSessionPtr(); s != nil {
		msgs := s.Messages()
		if len(msgs) != 2 {
			t.Fatalf("expected 2 messages (user + echo), got %d", len(msgs))
		}
	}
}

func TestSendMessageEmpty(t *testing.T) {
	a := NewApp(nil)
	a.newSession()
	a.sendMessage() // empty input — should not add message
	if s := a.activeSessionPtr(); s != nil {
		msgs := s.Messages()
		if len(msgs) != 0 {
			t.Fatalf("expected 0 messages for empty send, got %d", len(msgs))
		}
	}
}

func TestEnterExitSearch(t *testing.T) {
	a := NewApp(nil)
	a.enterSearch()
	if a.mode != ModeSearch {
		t.Fatalf("expected ModeSearch, got %d", a.mode)
	}
	a.exitSearch()
	if a.mode != ModeChat {
		t.Fatalf("expected ModeChat after exit, got %d", a.mode)
	}
}

func TestEnterExitHelp(t *testing.T) {
	a := NewApp(nil)
	a.enterHelp()
	if a.mode != ModeHelp {
		t.Fatalf("expected ModeHelp, got %d", a.mode)
	}
	a.exitHelp()
	if a.mode != ModeChat {
		t.Fatalf("expected ModeChat after exit, got %d", a.mode)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd G:\mllm\agent5
go test ./internal/ui/ -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git rm internal/ui/ui.go internal/ui/ui_test.go
git commit -m "feat: create main App with tview.Application, Pages, InputCapture"
```

---

### Task 10: Update cmd/agent/main.go

**Files:**
- Modify: `cmd/agent/main.go`

- [ ] **Step 1: Rewrite main.go**

Replace contents (remove all Bubble Tea references):

```go
package main

import (
	"log"

	"github.com/example/agent-tui/internal/ai"
	"github.com/example/agent-tui/internal/data/config"
	"github.com/example/agent-tui/internal/data/history"
	"github.com/example/agent-tui/internal/service"
	"github.com/example/agent-tui/internal/ui"
)

func main() {
	cfg, err := config.LoadConfig("configs/config.toml")
	if err != nil {
		cfg = config.GetDefaultConfig()
	}
	if cfg.DefaultClient == "" {
		cfg.DefaultClient = "local"
	}

	aiClient, err := ai.NewClientFromConfig(cfg)
	if err != nil {
		log.Fatalf("failed to create AI client: %v", err)
	}

	h := history.NewHistory("")
	aiAssistant := service.NewAIAssistant(aiClient, h)

	app := ui.NewApp(aiAssistant)
	app.AddWelcomeMessage()

	if err := app.Run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}
```

- [ ] **Step 2: Add AddWelcomeMessage to App**

In `app.go`, add:

```go
func (a *App) AddWelcomeMessage() {
	a.newSession()
	if s := a.activeSessionPtr(); s != nil {
		s.AddMessage("system", "Hello! Welcome to the Agent TUI.")
	}
	a.chatPanel.SetSession(a.activeSessionPtr())
}
```

- [ ] **Step 3: Build**

```bash
cd G:\mllm\agent5
go build ./cmd/agent
```

Expected: builds successfully with no errors.

- [ ] **Step 4: Run tests**

```bash
cd G:\mllm\agent5
go test ./... 2>&1
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/agent/main.go internal/ui/app.go
git commit -m "feat: update main.go for tview Application"
```

---

### Task 11: Remove editor package

**Files:**
- Delete: `internal/ui/editor/editor.go`
- Delete: `internal/ui/editor/editor_test.go`

- [ ] **Step 1: Delete editor files**

```bash
cd G:\mllm\agent5
git rm -r internal/ui/editor/
```

- [ ] **Step 2: Build**

```bash
cd G:\mllm\agent5
go build ./...
```

Expected: builds successfully with no errors.

- [ ] **Step 3: Commit**

```bash
git commit -m "chore: remove editor package (no longer needed)"
```

---

### Task 12: Update remaining tests and clean up

**Files:**
- Modify: various test files
- Verify full test suite

- [ ] **Step 1: Update ui package top-level test**

Ensure `internal/ui/app_test.go` has all relevant tests from the previous `ui_test.go`.

- [ ] **Step 2: Full test run**

```bash
cd G:\mllm\agent5
go test ./... -count=1 2>&1
```

Expected: all packages report PASS with no failures.

- [ ] **Step 3: Run vet**

```bash
cd G:\mllm\agent5
go vet ./...
```

Expected: no warnings or errors.

---

### Task 13: Remove old dependencies

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Remove old UI dependencies**

```bash
cd G:\mllm\agent5
go mod edit -droprequire charm.land/bubbletea/v2
go mod edit -droprequire charm.land/bubbles/v2
go mod edit -droprequire charm.land/lipgloss/v2
go mod tidy
```

- [ ] **Step 2: Verify build still works**

```bash
cd G:\mllm\agent5
go build ./...
go test ./...
```

Expected: all tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: remove old UI deps (bubbletea, bubbles, lipgloss)"
```
