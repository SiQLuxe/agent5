package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/example/agent-tui/internal/ui/composer"
	"github.com/example/agent-tui/internal/ui/status"
	"github.com/example/agent-tui/internal/ui/tabbar"
	"github.com/google/uuid"
)

type ViewMode string

const (
	ModeChat    ViewMode = "chat"
	ModeDefault ViewMode = ModeChat
)

type ChatResponseMsg struct {
	UserInput string
	Content   string
	Error     error
}

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
	isLoading     bool
	showHelp      bool
	lastText      string
	clearTime     time.Time
}

func NewModel() *Model {
	themeService := NewThemeService(nil)

	session := NewSession(uuid.New().String(), "New Session")
	sessions := []*Session{session}

	chatPanel := NewChatPanel(session)

	tabs := []tabbar.Tab{
		{ID: session.ID, Label: "New Session"},
	}
	tabDock := tabbar.NewTabDock(tabs)

	return &Model{
		sessions:      sessions,
		activeSession: 0,
		chatPanel:     chatPanel,
		tabDock:       tabDock,
		composer:      composer.NewComposer(),
		statusBar:     status.NewStatusBar(),
		themeService:  themeService,
		isLoading:     false,
	}
}

func (m *Model) SetAIAssistant(ai interface{}) {
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tabDock.SetWidth(msg.Width)
		m.composer.SetWidth(msg.Width)
		chatHeight := m.height - 4
		if chatHeight < 1 {
			chatHeight = 1
		}
		m.chatPanel.SetSize(msg.Width, chatHeight)

	case ChatResponseMsg:
		m.composer.ClearInput()
		m.lastText = ""
		m.clearTime = time.Now()
		session := m.activeSessionPtr()
		if session != nil {
			if msg.UserInput != "" {
				session.AddMessage(RoleUser, msg.UserInput)
			}
			if msg.Error != nil {
				session.AddMessage(RoleSystem, "❌ Error: "+msg.Error.Error())
			} else if msg.Content != "" {
				session.AddMessage(RoleAssistant, msg.Content)
			}
			m.tabDock.UpdateTabLabel(m.activeSession, session.GenerateLabel())
		}
		m.isLoading = false
		m.chatPanel.ScrollToBottom()

	case tea.PasteMsg:
		m.composer.AppendInput(msg.Content)

	case tea.KeyPressMsg:
		// Help panel: intercept all keys when visible
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		// Search mode: handle search-specific keys
		if m.chatPanel.IsSearchMode() {
			return m, m.handleSearchKey(msg)
		}

		if m.isLoading {
			return m, nil
		}

		key := msg.Key()

		// Printable characters: Key.Text is populated (space → " ")
		if key.Text != "" {
			if key.Text == m.lastText && time.Since(m.clearTime) < 100*time.Millisecond {
				m.lastText = ""
				return m, nil
			}
			m.lastText = key.Text
			m.composer.AppendInput(key.Text)
			return m, nil
		}

		// Special keys: Key.Text is empty, use String() for matching
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+g":
			m.showHelp = true
		case "ctrl+t":
			m.themeService.NextTheme()
		case "ctrl+n":
			m.newSession()
		case "ctrl+q":
			m.closeSession()
		case "ctrl+e":
			m.renameSession()
		case "ctrl+y":
			session := m.activeSessionPtr()
			if session != nil {
				session.ToggleThinking()
			}
		case "ctrl+l":
			session := m.activeSessionPtr()
			if session != nil {
				session.ToggleCollapse()
			}
		case "ctrl+f":
			m.chatPanel.EnterSearch()
			m.composer.SetInput("")
		// Tab switching: alt+n/p (ctrl+tab not supported by most terminals)
		case "alt+n", "alt+right":
			m.nextSession()
		case "alt+p", "alt+left":
			m.prevSession()
		// Session jump: alt+1~9 (ctrl+1~9 not supported by terminals)
		case "alt+1":
			m.switchToSession(0)
		case "alt+2":
			m.switchToSession(1)
		case "alt+3":
			m.switchToSession(2)
		case "alt+4":
			m.switchToSession(3)
		case "alt+5":
			m.switchToSession(4)
		case "alt+6":
			m.switchToSession(5)
		case "alt+7":
			m.switchToSession(6)
		case "alt+8":
			m.switchToSession(7)
		case "alt+9":
			m.switchToSession(8)
		case "pgup":
			m.chatPanel.ScrollUp(m.chatPanel.height / 2)
		case "pgdown":
			m.chatPanel.ScrollDown(m.chatPanel.height / 2)
		case "ctrl+up":
			m.chatPanel.ScrollUp(1)
		case "ctrl+down":
			m.chatPanel.ScrollDown(1)
		case "ctrl+home":
			m.chatPanel.ScrollToTop()
		case "ctrl+end":
			m.chatPanel.ScrollToBottom()
		case "enter":
			if strings.TrimSpace(m.composer.GetInput()) == "" {
				return m, nil
			}
			m.isLoading = true
			return m, m.submitMessageAsync()
		case "backspace":
			m.composer.Backspace()
		}
	case tea.KeyReleaseMsg:
		return m, nil
	}
	return m, nil
}

func (m *Model) handleSearchKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.Key()

	// Printable characters in search mode: update query
	if key.Text != "" {
		m.composer.AppendInput(key.Text)
		m.chatPanel.SetSearchQuery(m.composer.GetInput())
		return nil
	}

	switch key.String() {
	case "esc":
		m.chatPanel.ExitSearch()
		m.composer.ClearInput()
	case "enter":
		m.chatPanel.NextMatch()
	case "shift+enter":
		m.chatPanel.PrevMatch()
	case "backspace":
		m.composer.Backspace()
		m.chatPanel.SetSearchQuery(m.composer.GetInput())
	}
	return nil
}

func (m *Model) activeSessionPtr() *Session {
	if m.activeSession >= 0 && m.activeSession < len(m.sessions) {
		return m.sessions[m.activeSession]
	}
	return nil
}

func (m *Model) newSession() {
	session := NewSession(uuid.New().String(), "New Session")
	m.sessions = append(m.sessions, session)
	m.tabDock.AddTab(tabbar.Tab{ID: session.ID, Label: "New Session"})
	m.switchToSession(len(m.sessions) - 1)
}

func (m *Model) closeSession() {
	if len(m.sessions) <= 1 {
		return
	}
	idx := m.activeSession
	m.sessions = append(m.sessions[:idx], m.sessions[idx+1:]...)
	m.tabDock.RemoveTab(idx)
	if m.activeSession >= len(m.sessions) {
		m.activeSession = len(m.sessions) - 1
	}
	m.chatPanel.SetSession(m.activeSessionPtr())
}

func (m *Model) renameSession() {
	session := m.activeSessionPtr()
	if session == nil {
		return
	}
	// Use composer input as new label
	input := strings.TrimSpace(m.composer.GetInput())
	if input != "" {
		session.Label = input
		m.tabDock.UpdateTabLabel(m.activeSession, input)
		m.composer.ClearInput()
	}
}

func (m *Model) nextSession() {
	if len(m.sessions) == 0 {
		return
	}
	next := (m.activeSession + 1) % len(m.sessions)
	m.switchToSession(next)
}

func (m *Model) prevSession() {
	if len(m.sessions) == 0 {
		return
	}
	prev := (m.activeSession - 1 + len(m.sessions)) % len(m.sessions)
	m.switchToSession(prev)
}

func (m *Model) switchToSession(index int) {
	if index < 0 || index >= len(m.sessions) {
		return
	}
	m.activeSession = index
	m.tabDock.SetActiveTab(index)
	m.chatPanel.SetSession(m.sessions[index])
	m.statusBar.SetMode(m.sessions[index].Label)
}

func (m *Model) submitMessageAsync() tea.Cmd {
	return func() tea.Msg {
		input := m.composer.GetInput()
		return ChatResponseMsg{
			UserInput: input,
			Content:   fmt.Sprintf("Echo: %s", input),
		}
	}
}

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
		result = m.renderHelpOverlay(result)
	}

	v := tea.NewView(result)
	v.AltScreen = true
	return v
}

func (m *Model) renderSearchBar() string {
	query := m.composer.GetInput()
	matches := m.chatPanel.searchMatches
	idx := m.chatPanel.searchIdx

	status := ""
	if len(matches) > 0 && idx >= 0 {
		status = fmt.Sprintf(" [%d/%d]", idx+1, len(matches))
	} else if query != "" {
		status = " [0/0]"
	}

	prompt := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#dcdcaa")).
		Bold(true).
		Render("🔍 ")

	inputText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d4d4d4")).
		Render(query)

	cursor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d4d4d4")).
		Background(lipgloss.Color("#d4d4d4")).
		Render(" ")

	matchStatus := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#858585")).
		Render(status)

	content := prompt + inputText + cursor + matchStatus

	return lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 2).
		BorderTop(true).
		BorderForeground(lipgloss.Color("#3c3c3c")).
		Render(content)
}

func (m *Model) renderHelpOverlay(underlying string) string {
	helpContent := m.buildHelpPanel()

	// Split underlying into lines, overlay help panel centered
	lines := strings.Split(underlying, "\n")
	totalLines := len(lines)

	helpLines := strings.Split(helpContent, "\n")
	helpHeight := len(helpLines)
	helpWidth := 0
	for _, l := range helpLines {
		if len(l) > helpWidth {
			helpWidth = len(l)
		}
	}

	// Center the help panel vertically and horizontally
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
			// Truncate/pad line to leftPad, then overlay help
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

func (m *Model) buildHelpPanel() string {
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#569cd6")).
		Background(lipgloss.Color("#2d2d30")).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#569cd6")).
		Bold(true).
		Align(lipgloss.Center)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d4d4d4")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 1)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#858585"))

	dividerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3c3c3c"))

	var rows []string
	rows = append(rows, titleStyle.Render("⌨  快捷键"))
	rows = append(rows, "")
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌃N", "新建会话"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌃Q", "关闭会话"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌃E", "重命名会话"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌃T", "切换主题"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌃Y", "折叠/展开思考"))
	rows = append(rows, dividerStyle.Render("─────────────────────"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌥N / ⌥→", "下一会话"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌥P / ⌥←", "上一会话"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌥1~9", "跳转到第N个会话"))
	rows = append(rows, dividerStyle.Render("─────────────────────"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "PgUp / PgDn", "滚动对话（半页）"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌃↑ / ⌃↓", "滚动对话（逐行）"))
	rows = append(rows, dividerStyle.Render("─────────────────────"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "Enter", "发送消息"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌫", "删除字符"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌃G", "快捷键帮助"))
	rows = append(rows, formatHelpRow(keyStyle, descStyle, "⌃C", "退出"))
	rows = append(rows, "")
	rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("#6a6a6a")).Align(lipgloss.Center).Render("按任意键关闭"))

	content := strings.Join(rows, "\n")
	return borderStyle.Render(content)
}

func formatHelpRow(keyStyle, descStyle lipgloss.Style, key, desc string) string {
	keyPart := keyStyle.Render(key)
	descPart := descStyle.Render(desc)
	// Pad key column to fixed width
	paddedKey := keyPart + strings.Repeat(" ", max(0, 12-lipgloss.Width(keyPart)))
	return paddedKey + " " + descPart
}

func (m *Model) SetMode(mode ViewMode) {
}

func (m *Model) GetMode() ViewMode {
	return ModeChat
}

func (m *Model) ToggleRightPanel() {
}

func (m *Model) AddChatMessage(role, content string) {
	session := m.activeSessionPtr()
	if session == nil {
		return
	}
	var r Role
	switch role {
	case "user":
		r = RoleUser
	case "assistant":
		r = RoleAssistant
	case "system":
		r = RoleSystem
	default:
		r = RoleUser
	}
	session.AddMessage(r, content)
}

func (m *Model) GetChatMessages() []Message {
	session := m.activeSessionPtr()
	if session == nil {
		return nil
	}
	return session.Messages
}

func (m *Model) SetChatInput(input string) {
}

func (m *Model) SubmitChatMessage() {
}

func (m *Model) SetComposerInput(input string) {
	m.composer.SetInput(input)
}

func (m *Model) GetComposerInput() string {
	return m.composer.GetInput()
}

func (m *Model) ClearComposerInput() {
	m.composer.ClearInput()
}

func (m *Model) SetPlan(plan string) {
}

func (m *Model) SetTodos(todos []string) {
}

func (m *Model) SetTasks(tasks []string) {
}

func (m *Model) SetAgents(agents []string) {
}

func (m *Model) IsLoading() bool {
	return m.isLoading
}
