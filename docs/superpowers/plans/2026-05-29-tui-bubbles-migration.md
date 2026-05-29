# TUI Bubbles Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace custom Bubble Tea UI components with Bubbles v2 library components (viewport, textarea, help, key) and upgrade lipgloss v1→v2.

**Architecture:** Each component is replaced independently. Key+Help are standalone; Viewport and Textarea depend on lipgloss v2 but can be done in parallel. All components follow the same pattern: thin wrapper around a Bubbles model, configuration via theme colors.

**Tech Stack:** Go 1.26, charm.land/bubbletea/v2 v2.0.6, charm.land/bubbles/v2 v2.1.0, charm.land/lipgloss/v2

---

### Task 1: Upgrade Dependencies (lipgloss v2 + bubbles v2)

**Files:**
- Modify: `go.mod`
- Modify: `internal/ui/ui.go:9`
- Modify: `internal/ui/chatpanel.go:9`
- Modify: `internal/ui/composer/composer.go:6`
- Modify: `internal/ui/status/status.go:6`
- Modify: `internal/ui/tabbar/tabbar.go:7`

- [ ] **Step 1: Add bubbles v2 + upgrade lipgloss v2 in go.mod**

Run:
```bash
cd G:\mllm\agent5
go get charm.land/bubbles/v2@v2.1.0
go get charm.land/lipgloss/v2@latest
go mod tidy
```

- [ ] **Step 2: Fix all lipgloss import paths across the codebase**

In each of these 5 files, replace the import:
```go
"github.com/charmbracelet/lipgloss"
```
with:
```go
"charm.land/lipgloss/v2"
```

Files:
- `internal/ui/ui.go:9`
- `internal/ui/chatpanel.go:9`
- `internal/ui/composer/composer.go:6`
- `internal/ui/status/status.go:6`
- `internal/ui/tabbar/tabbar.go:7`

- [ ] **Step 3: Verify compilation**

Run:
```bash
cd G:\mllm\agent5
go build ./...
```

Expected: Successful compilation. If there are lipgloss v2 API changes that cause errors, fix those specific calls (the project does NOT use AdaptiveColor, so the migration should be straightforward).

- [ ] **Step 4: Run existing tests**

Run:
```bash
cd G:\mllm\agent5
go test ./internal/ui/... -v
```

Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "build: upgrade lipgloss v1->v2, add bubbles v2"
```

---

### Task 2: Create Keymap (bubbles/key)

**Files:**
- Create: `internal/ui/keymap.go`

- [ ] **Step 1: Create keymap.go with KeyMap struct**

Create `internal/ui/keymap.go`:

```go
package ui

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Quit           key.Binding
	NewSession     key.Binding
	CloseSession   key.Binding
	RenameSession  key.Binding
	NextSession    key.Binding
	PrevSession    key.Binding
	ToggleThinking key.Binding
	ToggleCollapse key.Binding
	Search         key.Binding
	ToggleTheme    key.Binding
	ShowHelp       key.Binding
	ScrollUp       key.Binding
	ScrollDown     key.Binding
	ScrollTop      key.Binding
	ScrollBottom   key.Binding
	SendMessage    key.Binding
}

var DefaultKeyMap = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("Ctrl+C", "quit"),
	),
	NewSession: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("Ctrl+N", "new session"),
	),
	CloseSession: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("Ctrl+Q", "close session"),
	),
	RenameSession: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("Ctrl+E", "rename session"),
	),
	NextSession: key.NewBinding(
		key.WithKeys("alt+n", "alt+right"),
		key.WithHelp("Alt+N/→", "next session"),
	),
	PrevSession: key.NewBinding(
		key.WithKeys("alt+p", "alt+left"),
		key.WithHelp("Alt+P/←", "prev session"),
	),
	ToggleThinking: key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("Ctrl+Y", "toggle thinking"),
	),
	ToggleCollapse: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("Ctrl+L", "toggle collapse"),
	),
	Search: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("Ctrl+F", "search"),
	),
	ToggleTheme: key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("Ctrl+T", "toggle theme"),
	),
	ShowHelp: key.NewBinding(
		key.WithKeys("ctrl+g"),
		key.WithHelp("Ctrl+G", "show help"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("pgup", "ctrl+up"),
		key.WithHelp("PgUp/Ctrl+↑", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("pgdown", "ctrl+down"),
		key.WithHelp("PgDn/Ctrl+↓", "scroll down"),
	),
	ScrollTop: key.NewBinding(
		key.WithKeys("ctrl+home"),
		key.WithHelp("Ctrl+Home", "scroll to top"),
	),
	ScrollBottom: key.NewBinding(
		key.WithKeys("ctrl+end"),
		key.WithHelp("Ctrl+End", "scroll to bottom"),
	),
	SendMessage: key.NewBinding(
		key.WithKeys("ctrl+enter"),
		key.WithHelp("Ctrl+Enter", "send message"),
	),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		km.NewSession,
		km.NextSession,
		km.PrevSession,
		km.ToggleThinking,
		km.Search,
		km.ToggleTheme,
		km.ShowHelp,
	}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.NewSession, km.CloseSession, km.RenameSession, km.NextSession, km.PrevSession},
		{km.ToggleThinking, km.ToggleCollapse, km.Search, km.ToggleTheme, km.ShowHelp},
		{km.ScrollUp, km.ScrollDown, km.ScrollTop, km.ScrollBottom},
		{km.SendMessage, km.Quit},
	}
}
```

Note: Alt+1~9 for session switching is handled separately in ui.go Update() since it's a numbered range.

- [ ] **Step 2: Write keymap test**

Create `internal/ui/keymap_test.go`:

```go
package ui

import (
	"testing"

	"charm.land/bubbles/v2/key"
)

func TestKeyMap_RegisteredKeys(t *testing.T) {
	km := DefaultKeyMap
	bindings := km.ShortHelp()
	bindings = append(bindings, km.FullHelp()...)

	// Verify all major bindings exist
	cases := []struct {
		name  string
		b     key.Binding
		keys  []string
		help  string
	}{
		{"Quit", km.Quit, []string{"ctrl+c"}, "quit"},
		{"NewSession", km.NewSession, []string{"ctrl+n"}, "new session"},
		{"CloseSession", km.CloseSession, []string{"ctrl+q"}, "close session"},
		{"RenameSession", km.RenameSession, []string{"ctrl+e"}, "rename session"},
		{"NextSession", km.NextSession, []string{"alt+n", "alt+right"}, "next session"},
		{"PrevSession", km.PrevSession, []string{"alt+p", "alt+left"}, "prev session"},
		{"ToggleThinking", km.ToggleThinking, []string{"ctrl+y"}, "toggle thinking"},
		{"ToggleCollapse", km.ToggleCollapse, []string{"ctrl+l"}, "toggle collapse"},
		{"Search", km.Search, []string{"ctrl+f"}, "search"},
		{"ToggleTheme", km.ToggleTheme, []string{"ctrl+t"}, "toggle theme"},
		{"ShowHelp", km.ShowHelp, []string{"ctrl+g"}, "show help"},
		{"ScrollUp", km.ScrollUp, []string{"pgup", "ctrl+up"}, "scroll up"},
		{"ScrollDown", km.ScrollDown, []string{"pgdown", "ctrl+down"}, "scroll down"},
		{"ScrollTop", km.ScrollTop, []string{"ctrl+home"}, "scroll to top"},
		{"ScrollBottom", km.ScrollBottom, []string{"ctrl+end"}, "scroll to bottom"},
		{"SendMessage", km.SendMessage, []string{"ctrl+enter"}, "send message"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.b.Help().Key == "" {
				t.Error("binding has empty Help().Key")
			}
			if c.b.Help().Desc != c.help {
				t.Errorf("expected help desc %q, got %q", c.help, c.b.Help().Desc)
			}
		})
	}
}

func TestKeyMap_ShortHelp(t *testing.T) {
	km := DefaultKeyMap
	short := km.ShortHelp()
	if len(short) == 0 {
		t.Fatal("ShortHelp returned empty")
	}
}

func TestKeyMap_FullHelp(t *testing.T) {
	km := DefaultKeyMap
	full := km.FullHelp()
	if len(full) == 0 {
		t.Fatal("FullHelp returned empty")
	}
}
```

- [ ] **Step 3: Run tests**

```bash
cd G:\mllm\agent5
go test ./internal/ui/... -v -run TestKeyMap
```

Expected: All tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/keymap.go internal/ui/keymap_test.go
git commit -m "feat: add keymap with bubbles/key bindings"
```

---

### Task 3: Add Help Model + Keyboard Refactor in ui.go

**Files:**
- Modify: `internal/ui/ui.go`
- Modify: `internal/ui/ui_test.go`

- [ ] **Step 1: Add help.Model and KeyMap to main Model struct**

In `internal/ui/ui.go`, update the Model struct:

```go
type Model struct {
	width         int
	height        int
	sessions      []*Session
	activeSession int
	chatPanel     *ChatPanel
	tabDock       *tabbar.TabDock
	composer      *composer.Composer
	statusBar     *status.StatusBar
	themeService  *ThemeService
	keyMap        KeyMap
	help          help.Model
	isLoading     bool
	showHelp      bool
	lastText      string
	clearTime     time.Time
}
```

Add imports:
```go
import (
	"time"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/example/agent-tui/internal/ui/composer"
	"github.com/example/agent-tui/internal/ui/status"
	"github.com/example/agent-tui/internal/ui/tabbar"
	"github.com/google/uuid"
)
```

In `NewModel()`, initialize:
```go
return &Model{
	sessions:      sessions,
	activeSession: 0,
	chatPanel:     chatPanel,
	tabDock:       tabDock,
	composer:      composer.NewComposer(),
	statusBar:     status.NewStatusBar(),
	themeService:  themeService,
	keyMap:        DefaultKeyMap,
	help:          help.New(),
	isLoading:     false,
}
```

- [ ] **Step 2: Refactor Update() to use key.Matches()**

Replace the keyboard handling in `Update()`:

```go
case tea.KeyPressMsg:
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	if m.chatPanel.IsSearchMode() {
		return m, m.handleSearchKey(msg)
	}

	if m.isLoading {
		return m, nil
	}

	keyPress := msg

	// Printable characters: forward to composer
	if keyPress.Key().Text != "" {
		if keyPress.Key().Text == m.lastText && time.Since(m.clearTime) < 100*time.Millisecond {
			m.lastText = ""
			return m, nil
		}
		m.lastText = keyPress.Key().Text
		m.composer.AppendInput(keyPress.Key().Text)
		return m, nil
	}

	switch {
	case key.Matches(keyPress, m.keyMap.Quit):
		return m, tea.Quit
	case key.Matches(keyPress, m.keyMap.NewSession):
		m.newSession()
	case key.Matches(keyPress, m.keyMap.CloseSession):
		m.closeSession()
	case key.Matches(keyPress, m.keyMap.RenameSession):
		m.renameSession()
	case key.Matches(keyPress, m.keyMap.NextSession):
		m.nextSession()
	case key.Matches(keyPress, m.keyMap.PrevSession):
		m.prevSession()
	case key.Matches(keyPress, m.keyMap.ToggleThinking):
		if session := m.activeSessionPtr(); session != nil {
			session.ToggleThinking()
		}
	case key.Matches(keyPress, m.keyMap.ToggleCollapse):
		if session := m.activeSessionPtr(); session != nil {
			session.ToggleCollapse()
		}
	case key.Matches(keyPress, m.keyMap.Search):
		m.chatPanel.EnterSearch()
		m.composer.SetInput("")
	case key.Matches(keyPress, m.keyMap.ToggleTheme):
		m.themeService.NextTheme()
	case key.Matches(keyPress, m.keyMap.ShowHelp):
		m.showHelp = true
	case key.Matches(keyPress, m.keyMap.ScrollUp):
		m.chatPanel.ScrollUp(m.chatPanel.height / 2)
	case key.Matches(keyPress, m.keyMap.ScrollDown):
		m.chatPanel.ScrollDown(m.chatPanel.height / 2)
	case key.Matches(keyPress, m.keyMap.ScrollTop):
		m.chatPanel.ScrollToTop()
	case key.Matches(keyPress, m.keyMap.ScrollBottom):
		m.chatPanel.ScrollToBottom()
	case key.Matches(keyPress, m.keyMap.SendMessage):
		if strings.TrimSpace(m.composer.GetInput()) == "" {
			return m, nil
		}
		m.isLoading = true
		return m, m.submitMessageAsync()
	case keyPress.Key().String() == "enter":
		// Single Enter: ignored in textinput mode (used for sending in textarea mode)
		// In current impl it sends - this will change in Task 6
		if strings.TrimSpace(m.composer.GetInput()) == "" {
			return m, nil
		}
		m.isLoading = true
		return m, m.submitMessageAsync()
	case keyPress.Key().String() == "backspace":
		m.composer.Backspace()
	}
```

Keep the Alt+1~9 session switching separately (not using key.Matches since it's dynamic):
```go
	// Alt+1~9 session switching
	switch keyPress.Key().String() {
	case "alt+1": m.switchToSession(0)
	case "alt+2": m.switchToSession(1)
	case "alt+3": m.switchToSession(2)
	case "alt+4": m.switchToSession(3)
	case "alt+5": m.switchToSession(4)
	case "alt+6": m.switchToSession(5)
	case "alt+7": m.switchToSession(6)
	case "alt+8": m.switchToSession(7)
	case "alt+9": m.switchToSession(8)
	}
```

- [ ] **Step 3: Refactor View() to use help.Model**

Replace `renderHelpOverlay()` and `buildHelpPanel()` with help.Model:

```go
func (m *Model) View() tea.View {
	statusContent := m.statusBar.View(m.width)

	var composerContent string
	if m.chatPanel.IsSearchMode() {
		composerContent = m.renderSearchBar()
	} else {
		composerContent = m.composer.View()
	}

	tabContent := m.tabDock.View()

	chatHeight := m.height - 4
	if chatHeight < 2 {
		chatHeight = 2
	}
	m.chatPanel.SetSize(m.width, chatHeight)
	chatContent := m.chatPanel.View()

	result := statusContent + "\n" + chatContent + "\n" + composerContent + "\n" + tabContent

	if m.showHelp {
		result = m.overlayHelp(result)
	}

	return tea.NewView(result)
}

func (m *Model) overlayHelp(underlying string) string {
	helpContent := m.help.View(m.keyMap)
	lines := strings.Split(underlying, "\n")
	helpLines := strings.Split(helpContent, "\n")
	helpHeight := len(helpLines)
	helpWidth := 0
	for _, l := range helpLines {
		if len(l) > helpWidth {
			helpWidth = len(l)
		}
	}

	totalLines := len(lines)
	startRow := (totalLines - helpHeight) / 2
	if startRow < 1 {
		startRow = 1
	}

	leftPad := (m.width - helpWidth) / 2
	if leftPad < 2 {
		leftPad = 2
	}

	var result []string
	for i, line := range lines {
		if i >= startRow && i < startRow+helpHeight {
			helpLine := helpLines[i-startRow]
			base := line
			if len(base) > leftPad {
				base = base[:leftPad]
			} else {
				base += strings.Repeat(" ", leftPad-len(base))
			}
			result = append(result, base+helpLine)
		} else {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}
```

Remove `buildHelpPanel()` and `formatHelpRow()` methods.

Keep `renderSearchBar()` method unchanged.

- [ ] **Step 4: Update ui_test.go**

In `internal/ui/ui_test.go`, update imports and ensure tests still compile. Update `TestModelCreation` to verify `help` and `keyMap` are initialized:

```go
func TestModelCreation(t *testing.T) {
	m := NewModel()
	if m.keyMap.Quit.Help().Key == "" {
		t.Error("keyMap not initialized")
	}
	// ... existing assertions
}
```

- [ ] **Step 5: Run tests**

```bash
cd G:\mllm\agent5
go test ./internal/ui/... -v -run TestKeyMap
go test ./internal/ui/... -v
```

Expected: All tests pass.

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: integrate help model and key-based keyboard dispatch"
```

---

### Task 4: Add session.RenderMessages()

**Files:**
- Modify: `internal/ui/session.go`
- Create: `internal/ui/session_render_test.go`

- [ ] **Step 1: Add RenderMessages() to session.go**

Append to `internal/ui/session.go`:

```go
import (
	"strings"
	"time"
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/example/agent-tui/internal/ui/syntax"
)

// RenderMessages renders all messages as a formatted string for viewport content.
func (s *Session) RenderMessages(width int, theme ColorPalette) string {
	if len(s.Messages) == 0 {
		return ""
	}
	contentWidth := width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	var sb strings.Builder
	for _, msg := range s.Messages {
		renderMessageToBuilder(&sb, msg, contentWidth, theme)
	}
	return sb.String()
}

func renderMessageToBuilder(sb *strings.Builder, msg Message, width int, theme ColorPalette) {
	ts := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Timestamp)).
		Render(msg.Timestamp.Format("15:04"))

	switch msg.Role {
	case RoleUser:
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.UserBg)).
			Foreground(lipgloss.Color(theme.UserFg)).
			Bold(true).
			Padding(0, 1).
			Render(" 🗣️ ")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 3)

	case RoleAssistant:
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.AssistantBg)).
			Foreground(lipgloss.Color(theme.AssistantFg)).
			Bold(true).
			Padding(0, 1).
			Render(" 👾 Chan ")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")

		if msg.Thinking != nil {
			renderThinkingBlock(sb, msg.Thinking, width, theme)
		}
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 1)

	case RoleSystem:
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.SystemBg)).
			Foreground(lipgloss.Color(theme.SystemFg)).
			Bold(true).
			Padding(0, 1).
			Render(" ⚙ sys ")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 1)
	}

	sb.WriteString("\n")
}

func renderContent(sb *strings.Builder, content string, collapsed bool, width int, theme ColorPalette, paddingLeft int) {
	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Text)).
		PaddingLeft(paddingLeft)

	if collapsed {
		sb.WriteString(foldContent(content, textStyle, theme))
		return
	}

	highlighted := safeHighlight(content)
	sb.WriteString(textStyle.Render(highlighted))
}

func safeHighlight(content string) string {
	if hasCJK(content) {
		return content
	}
	return syntax.Highlight(content, "")
}

func foldContent(content string, style lipgloss.Style, theme ColorPalette) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)
	if totalLines <= 3 {
		return style.Render(content)
	}
	preview := strings.Join(lines[:3], "\n")
	indicator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextMuted)).
		Render(fmt.Sprintf("\u25bc [expand %d lines]", totalLines-3))
	return style.Render(preview) + "\n" + indicator
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

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ThinkingFg)).
		Render(fmt.Sprintf("%s thinking %d chars%s", arrow, charCount, durationStr))
	sb.WriteString(header)
	sb.WriteString("\n")

	if t.Expanded {
		contentStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.ThinkingBg)).
			BorderLeft(true).
			BorderStyle(lipgloss.Border{Left: "\u2502"}).
			BorderForeground(lipgloss.Color(theme.ThinkingBorder)).
			Padding(0, 1).
			MarginLeft(1).
			Width(width - 4)
		sb.WriteString(contentStyle.Render(t.Content))
		sb.WriteString("\n")
	}
}

// hasCJK returns true if the string contains any CJK characters
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

- [ ] **Step 2: Write RenderMessages tests**

Create `internal/ui/session_render_test.go`:

```go
package ui

import (
	"strings"
	"testing"
	"time"
)

func TestRenderMessages_Empty(t *testing.T) {
	s := NewSession("1", "test")
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestRenderMessages_UserMessage(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleUser, "hello world")
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "hello world") {
		t.Errorf("expected content in output, got: %s", result)
	}
	if !strings.Contains(result, "🗣️") {
		t.Error("expected user badge")
	}
}

func TestRenderMessages_AssistantMessage(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleAssistant, "I am an AI")
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "I am an AI") {
		t.Errorf("expected content in output, got: %s", result)
	}
	if !strings.Contains(result, "👾") {
		t.Error("expected assistant badge")
	}
}

func TestRenderMessages_SystemMessage(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleSystem, "system message")
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "system message") {
		t.Errorf("expected content in output, got: %s", result)
	}
	if !strings.Contains(result, "⚙ sys") {
		t.Error("expected system badge")
	}
}

func TestRenderMessages_ThinkingCollapsed(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleAssistant, "answer")
	s.Messages[0].Thinking = &Thinking{
		Content:  "thinking content",
		Expanded: false,
		Duration: time.Second,
	}
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "▶") {
		t.Error("expected collapsed thinking indicator ▶")
	}
	if strings.Contains(result, "thinking content") {
		t.Error("thinking content should not be visible when collapsed")
	}
}

func TestRenderMessages_ThinkingExpanded(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleAssistant, "answer")
	s.Messages[0].Thinking = &Thinking{
		Content:  "expanded thinking",
		Expanded: true,
		Duration: 2 * time.Second,
	}
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "▼") {
		t.Error("expected expanded thinking indicator ▼")
	}
	if !strings.Contains(result, "expanded thinking") {
		t.Error("thinking content should be visible when expanded")
	}
}

func TestRenderMessages_Collapsed(t *testing.T) {
	s := NewSession("1", "test")
	content := "line1\nline2\nline3\nline4\nline5"
	s.AddMessage(RoleUser, content)
	s.Messages[0].Collapsed = true
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "▼") {
		t.Error("expected collapsed indicator ▼")
	}
}

func TestRenderMessages_MultipleMessages(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleUser, "user msg")
	s.AddMessage(RoleAssistant, "assistant msg")
	s.AddMessage(RoleSystem, "system msg")
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "user msg") {
		t.Error("missing user message")
	}
	if !strings.Contains(result, "assistant msg") {
		t.Error("missing assistant message")
	}
	if !strings.Contains(result, "system msg") {
		t.Error("missing system message")
	}
}
```

- [ ] **Step 3: Run tests**

```bash
cd G:\mllm\agent5
go test ./internal/ui/... -v -run TestRenderMessages
```

Expected: All tests pass.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat: add session.RenderMessages for viewport formatting"
```

---

### Task 5: Replace ChatPanel with bubbles/viewport

**Files:**
- Modify: `internal/ui/chatpanel.go`
- Modify: `internal/ui/ui.go`
- Modify: `internal/ui/ui_test.go`

- [ ] **Step 1: Rewrite chatpanel.go as viewport wrapper**

Replace the entire content of `internal/ui/chatpanel.go`:

```go
package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

type ChatPanel struct {
	session       *Session
	viewport      viewport.Model
	width         int
	height        int
	colors        ChatPanelColors
	currentTheme  ColorPalette
	searchMode    bool
	searchQuery   string
	searchIdx     int
	searchMatches []int
}

type ChatPanelColors struct {
	Background string
	TextMuted  string
}

func DefaultChatPanelColors() ChatPanelColors {
	return ChatPanelColors{
		Background: "#1e1e1e",
		TextMuted:  "#858585",
	}
}

func NewChatPanel(session *Session) *ChatPanel {
	vp := viewport.New()
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 3

	cp := &ChatPanel{
		session:  session,
		viewport: vp,
		width:    80,
		height:   24,
		colors:   DefaultChatPanelColors(),
	}
	cp.ApplyTheme(DefaultThemes[0].Colors)
	return cp
}

func (cp *ChatPanel) SetSession(s *Session) {
	cp.session = s
	cp.refreshContent()
}

func (cp *ChatPanel) Session() *Session {
	return cp.session
}

func (cp *ChatPanel) SetColors(c ChatPanelColors) {
	cp.colors = c
}

func (cp *ChatPanel) SetSize(width, height int) {
	cp.width = width
	cp.height = height
	cp.viewport.SetWidth(width)
	cp.viewport.SetHeight(height)
}

func (cp *ChatPanel) ScrollUp(lines int) {
	cp.viewport.ScrollUp(lines)
}

func (cp *ChatPanel) ScrollDown(lines int) {
	cp.viewport.ScrollDown(lines)
}

func (cp *ChatPanel) ScrollToBottom() {
	cp.viewport.GotoBottom()
}

func (cp *ChatPanel) ScrollToTop() {
	cp.viewport.GotoTop()
}

func (cp *ChatPanel) View() string {
	cp.refreshContent()
	result := cp.viewport.View()
	style := lipgloss.NewStyle().
		Width(cp.width).
		Height(cp.height).
		Background(lipgloss.Color(cp.colors.Background))
	return style.Render(result)
}

func (cp *ChatPanel) refreshContent() {
	if cp.session == nil || len(cp.session.Messages) == 0 {
		cp.viewport.SetContent("")
		return
	}
	content := cp.session.RenderMessages(cp.width, cp.currentTheme)
	cp.viewport.SetContent(content)
}

func (cp *ChatPanel) ApplyTheme(theme ColorPalette) {
	cp.currentTheme = theme
	cp.viewport.Style = lipgloss.NewStyle().
		Background(lipgloss.Color(theme.Background))
}

// Search methods
func (cp *ChatPanel) EnterSearch() {
	cp.searchMode = true
	cp.searchQuery = ""
	cp.searchIdx = -1
}

func (cp *ChatPanel) ExitSearch() {
	cp.searchMode = false
	cp.searchQuery = ""
	cp.searchIdx = -1
	cp.viewport.ClearHighlights()
}

func (cp *ChatPanel) IsSearchMode() bool {
	return cp.searchMode
}

func (cp *ChatPanel) SetSearchQuery(q string) {
	cp.searchQuery = q
	cp.searchMatches = nil
	cp.searchIdx = -1
	cp.viewport.ClearHighlights()
	if q == "" || cp.session == nil {
		return
	}

	lower := strings.ToLower(q)
	var matchLines [][]int
	lines := strings.Split(cp.session.RenderMessages(cp.width, cp.currentTheme), "\n")
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), lower) {
			matchLines = append(matchLines, []int{i, i})
		}
	}

	if len(matchLines) > 0 {
		cp.searchIdx = 0
		cp.viewport.SetHighlights(matchLines)
	}
}

func (cp *ChatPanel) NextMatch() {
	if cp.searchIdx < 0 {
		return
	}
	cp.viewport.HighlightNext()
}

func (cp *ChatPanel) PrevMatch() {
	if cp.searchIdx < 0 {
		return
	}
	cp.viewport.HighlightPrevious()
}

func (cp *ChatPanel) Update(msg tea.Msg) {
	cp.viewport.Update(msg)
}
```

- [ ] **Step 2: Integrate with ui.go**

In `internal/ui/ui.go`:

```go
// On theme change:
m.chatPanel.ApplyTheme(m.themeService.CurrentTheme().Colors)

// On session switch:
m.chatPanel.SetSession(m.sessions[index])

// On WindowSizeMsg:
m.chatPanel.SetSize(msg.Width, chatHeight)
```

- [ ] **Step 3: Update ui_test.go**

Update tests to use new ChatPanel API.

- [ ] **Step 4: Build and test**

```bash
cd G:\mllm\agent5
go build ./...
go test ./internal/ui/... -v
```

Expected: Compiles and tests pass.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: replace custom chat panel with bubbles/viewport"
```

---

### Task 6: Replace Composer with bubbles/textarea

**Files:**
- Modify: `internal/ui/composer/composer.go`
- Modify: `internal/ui/composer/composer_test.go`
- Modify: `internal/ui/ui.go`

- [ ] **Step 1: Rewrite composer.go as textarea wrapper**

Replace `internal/ui/composer/composer.go`:

```go
package composer

import (
	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
)

type ComposerColors struct {
	Background string
	Prompt     string
	Text       string
}

func DefaultColors() ComposerColors {
	return ComposerColors{
		Background: "#1a1a1a",
		Prompt:     "#4ec9b0",
		Text:       "#d4d4d4",
	}
}

type Composer struct {
	textarea textarea.Model
	width    int
	colors   ComposerColors
}

func NewComposer() *Composer {
	ta := textarea.New()
	ta.DynamicHeight = true
	ta.MinHeight = 1
	ta.MaxHeight = 8
	ta.ShowLineNumbers = false
	ta.Prompt = "\u276f "  // ❯

	c := &Composer{
		textarea: ta,
		width:    80,
		colors:   DefaultColors(),
	}
	c.applyColors()
	return c
}

func (c *Composer) SetWidth(width int) {
	c.width = width
	c.textarea.SetWidth(width)
}

func (c *Composer) SetColors(colors ComposerColors) {
	c.colors = colors
	c.applyColors()
}

func (c *Composer) SetInput(input string) {
	c.textarea.SetText(input)
}

func (c *Composer) GetInput() string {
	return c.textarea.Text()
}

func (c *Composer) ClearInput() {
	c.textarea.SetText("")
}

func (c *Composer) AppendInput(char string) {
	c.textarea.InsertText(char)
}

func (c *Composer) Backspace() {
	c.textarea.DeleteCharacterBackward()
}

func (c *Composer) View() string {
	return c.textarea.View()
}

func (c *Composer) applyColors() {
	s := textarea.DefaultStyles(false)
	s.Focused.Base = lipgloss.NewStyle().
		Background(lipgloss.Color(c.colors.Background)).
		BorderTop(true).
		BorderForeground(lipgloss.Color("#3c3c3c"))
	s.Focused.Text = lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.colors.Text))
	s.Focused.Prompt = lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.colors.Prompt))
	s.Focused.CursorLine = lipgloss.NewStyle()
	s.Focused.LineNumber = lipgloss.NewStyle()
	c.textarea.SetStyles(s)
}
```

- [ ] **Step 2: Update composer_test.go**

Replace `internal/ui/composer/composer_test.go`:

```go
package composer

import (
	"strings"
	"testing"
)

func TestNewComposer(t *testing.T) {
	c := NewComposer()
	if c == nil {
		t.Fatal("expected non-nil Composer")
	}
}

func TestComposerSetInput(t *testing.T) {
	c := NewComposer()
	c.SetInput("hello")
	if c.GetInput() != "hello" {
		t.Errorf("expected 'hello', got '%s'", c.GetInput())
	}
}

func TestComposerClearInput(t *testing.T) {
	c := NewComposer()
	c.SetInput("hello")
	c.ClearInput()
	if c.GetInput() != "" {
		t.Errorf("expected empty, got '%s'", c.GetInput())
	}
}

func TestComposerAppendInput(t *testing.T) {
	c := NewComposer()
	c.SetInput("hel")
	c.AppendInput("lo")
	if c.GetInput() != "hello" {
		t.Errorf("expected 'hello', got '%s'", c.GetInput())
	}
}

func TestComposerBackspace(t *testing.T) {
	c := NewComposer()
	c.SetInput("hello")
	c.Backspace()
	if c.GetInput() != "hell" {
		t.Errorf("expected 'hell', got '%s'", c.GetInput())
	}
}

func TestComposerView(t *testing.T) {
	c := NewComposer()
	c.SetWidth(80)
	c.SetInput("test")
	result := c.View()
	if !strings.Contains(result, "test") {
		t.Errorf("expected view to contain 'test', got: %s", result)
	}
}
```

- [ ] **Step 3: Update ui.go for textarea send behavior**

In `internal/ui/ui.go`, change the send behavior:

```go
// Replace the "enter" case in Update:
case keyPress.Key().String() == "enter":
	return m, nil  // Enter = newline in textarea, ignore here
```

Also update `submitMessageAsync()` - it still uses `m.composer.GetInput()` which now returns `m.textarea.Text()` through the wrapper.

- [ ] **Step 4: Build and test**

```bash
cd G:\mllm\agent5
go build ./...
go test ./internal/ui/... -v
```

Expected: Compiles and tests pass.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: replace custom composer with bubbles/textarea"
```

---

### Task 7: Add Mouse Support + Update main.go

**Files:**
- Modify: `cmd/agent/main.go`
- Modify: `internal/ui/tabbar/tabbar.go`
- Modify: `internal/ui/ui.go`

- [ ] **Step 1: Add mouse support to Bubble Tea program**

In `cmd/agent/main.go`, change:

```go
p := tea.NewProgram(model, tea.WithMouseCellMotion())
```

- [ ] **Step 2: Handle mouse events in ui.go**

Add mouse event handling in `ui.go` Update():

```go
case tea.MouseMsg:
	if m.showHelp {
		return m, nil
	}
	// Forward to viewport for mouse wheel scrolling
	m.chatPanel.Update(msg)
	// TabDock click handling
	if msg.Y == m.height-1 {
		tabID, ok := m.tabDock.HandleClick(msg.X)
		if ok {
			for i, s := range m.sessions {
				if s.ID == tabID {
					m.switchToSession(i)
					break
				}
			}
		}
	}
```

- [ ] **Step 3: Build and verify**

```bash
cd G:\mllm\agent5
go build ./...
```

Expected: Compiles successfully.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "feat: add mouse support (scroll + tab click)"
```

---

### Task 8: Cleanup Remaining Files

**Files:**
- Modify: `internal/ui/status/status.go` (lipgloss v2 already imported, verify compiles)
- Modify: `internal/ui/tabbar/tabbar.go` (lipgloss v2 already imported, verify compiles)
- Modify: `internal/ui/theme.go` (no lipgloss import, verify nothing needs changing)

- [ ] **Step 1: Verify StatusBar compiles with lipgloss v2**

```bash
cd G:\mllm\agent5
go build ./internal/ui/status/...
```

Expected: Compiles.

- [ ] **Step 2: Verify TabDock compiles with lipgloss v2**

```bash
cd G:\mllm\agent5
go build ./internal/ui/tabbar/...
```

Expected: Compiles.

- [ ] **Step 3: Remove unused code from chatpanel.go**

After rewriting chatpanel.go, ensure no dead functions remain. Verify:

```bash
cd G:\mllm\agent5
go vet ./internal/ui/...
```

Expected: No warnings.

- [ ] **Step 4: Run all tests**

```bash
cd G:\mllm\agent5
go test ./... -v
```

Expected: All tests pass.

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "chore: cleanup unused code after bubbles migration"
```

---

### Verification

- [ ] **Full compilation check**

```bash
cd G:\mllm\agent5
go build ./...
go vet ./...
```

- [ ] **Full test suite**

```bash
cd G:\mllm\agent5
go test ./... -v
```

- [ ] **Final status check**

```bash
cd G:\mllm\agent5
git status
git log --oneline -10
```

Expected: Clean working tree, green test suite, all 8 tasks committed.
