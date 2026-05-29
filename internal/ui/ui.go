package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	keyMap        KeyMap
	help          help.Model
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
		keyMap:        DefaultKeyMap,
		help:          help.New(),
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
			m.chatPanel.ApplyTheme(m.themeService.CurrentTheme().Colors)
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
			return m, nil  // Enter = newline in textarea, ignore here
		case keyPress.Key().String() == "backspace":
			m.composer.Backspace()
		}

		// Alt+1~9 session switching
		switch keyPress.Key().String() {
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
		result = m.overlayHelp(result)
	}

	return tea.NewView(result)
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
