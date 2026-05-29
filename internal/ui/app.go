package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/rivo/tview"

	"github.com/example/agent-tui/internal/service"
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
	pages         *tview.Pages
	chatFlex      *tview.Flex
	statusBar     *status.StatusBar
	chatPanel     *ChatPanel
	composer      *composer.Composer
	tabDock       *tabbar.TabDock
	searchInput   *tview.InputField
	searchStatus  *tview.TextView
	helpView      *tview.TextView
	sessions      []*Session
	activeSession int
	mode          AppMode
	isLoading     bool
	keyMap        KeyMap
	themeService  *ThemeService
	aiAssistant   *service.AIAssistant
}

func NewApp() *App {
	a := &App{
		Application:   tview.NewApplication(),
		pages:         tview.NewPages(),
		statusBar:     status.New(),
		chatPanel:     NewChatPanel(),
		composer:      composer.New(),
		tabDock:       tabbar.New(),
		searchInput:   tview.NewInputField(),
		searchStatus:  tview.NewTextView(),
		helpView:      tview.NewTextView(),
		sessions:      []*Session{},
		activeSession: -1,
		mode:          ModeChat,
		keyMap:        DefaultKeyMap(),
		themeService:  NewThemeService(nil),
	}

	// Wire tab click
	a.tabDock.SetOnClick(func(idx int) {
		a.switchToSession(idx)
	})

	// Setup search input
	a.searchInput.SetChangedFunc(func(text string) {
		a.chatPanel.SetSearchQuery(text)
	})

	// Setup search help text
	a.searchStatus.SetDynamicColors(true)
	a.searchStatus.SetText("[gray]Enter: next  Shift+Enter: prev  Esc: exit[-]")

	// Setup help view
	a.helpView.SetDynamicColors(true)
	a.helpView.SetText(strings.Join(a.keyMap.FullHelp(), "\n"))
	a.helpView.SetTextAlign(tview.AlignLeft)
	a.helpView.SetBorder(true)
	a.helpView.SetTitle(" Help ")

	// Build layout: StatusBar + ChatPanel + Composer + TabDock
	chatFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	chatFlex.AddItem(a.statusBar, 1, 0, false)
	chatFlex.AddItem(a.chatPanel, 0, 1, false)
	chatFlex.AddItem(a.composer, 3, 0, true)
	chatFlex.AddItem(a.tabDock, 1, 0, false)
	a.chatFlex = chatFlex

	// Search overlay: centered box with InputField + status
	searchFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	searchFlex.AddItem(nil, 0, 1, false)
	searchInner := tview.NewFlex().SetDirection(tview.FlexRow)
	searchInner.AddItem(a.searchInput, 1, 0, true)
	searchInner.AddItem(a.searchStatus, 1, 0, false)
	searchFlex.AddItem(searchInner, 2, 0, true)
	searchFlex.AddItem(nil, 0, 1, false)
	searchPage := tview.NewFlex().SetDirection(tview.FlexColumn)
	searchPage.AddItem(nil, 0, 1, false)
	searchPage.AddItem(searchFlex, 40, 0, true)
	searchPage.AddItem(nil, 0, 1, false)

	// Help overlay: centered box with key bindings
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
	switch a.mode {
	case ModeHelp:
		if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
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
			if event.Modifiers()&tcell.ModShift != 0 {
				a.chatPanel.PrevMatch()
			} else {
				a.chatPanel.NextMatch()
			}
			return nil
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			text := a.searchInput.GetText()
			if len(text) > 0 {
				a.searchInput.SetText(text[:len(text)-1])
				a.chatPanel.SetSearchQuery(a.searchInput.GetText())
			}
			return nil
		default:
			if event.Rune() != 0 {
				return event
			}
			return nil
		}
	}

	// Chat mode — global shortcuts only here
	switch {
	case event.Key() == tcell.KeyCtrlC:
		a.Stop()
		return nil
	case event.Rune() == 'q' && event.Modifiers() == tcell.ModNone:
		a.Stop()
		return nil
	case event.Key() == tcell.KeyCtrlF:
		a.enterSearch()
		return nil
	case event.Rune() == '?' && event.Modifiers() == tcell.ModNone:
		a.enterHelp()
		return nil
	case event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModNone:
		// Enter to send
		if strings.TrimSpace(a.composer.GetInput()) == "" {
			return nil
		}
		a.sendMessage()
		return nil
	case event.Key() == tcell.KeyEnter && (event.Modifiers()&tcell.ModCtrl != 0 || event.Modifiers()&tcell.ModAlt != 0):
		// Ctrl+Enter or Alt+Enter: insert newline (let fall through to TextArea)
		return event
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
	case event.Modifiers()&tcell.ModAlt != 0:
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
		case 'S':
			a.themeService.NextTheme()
			a.applyTheme()
			return nil
		}
		// Alt+{1-9} session switching
		if event.Rune() >= '1' && event.Rune() <= '9' {
			idx := int(event.Rune() - '1')
			if idx < len(a.sessions) {
				a.switchToSession(idx)
			}
			return nil
		}
	}

	return event
}

// Input handlers

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

// Session management

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
	// Placeholder
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

// Messaging

func (a *App) sendMessage() {
	text := a.composer.GetInput()
	if strings.TrimSpace(text) == "" {
		return
	}
	if a.aiAssistant == nil {
		return
	}
	s := a.activeSessionPtr()
	if s == nil {
		return
	}
	s.AddMessage(RoleUser, text)
	s.AddMessage(RoleAssistant, "")
	a.composer.ClearInput()
	a.isLoading = true
	a.chatPanel.SetSession(s)
}

func (a *App) AddWelcomeMessage() {
	a.newSession()
	if s := a.activeSessionPtr(); s != nil {
		s.AddMessage(RoleSystem, "Hello! Welcome to the Agent TUI.")
	}
	a.chatPanel.SetSession(a.activeSessionPtr())
}

// AI assistant (placeholder)

func (a *App) SetAIAssistant(ai *service.AIAssistant) {
	a.aiAssistant = ai
}

// Theme

func (a *App) applyTheme() {
	colors := a.themeService.CurrentTheme().Colors
	a.statusBar.SetBackgroundColor(hexToTCell(colors.Background))
	a.chatPanel.ApplyTheme(colors)
	a.composer.SetBackgroundColor(hexToTCell(colors.Background))
}

// Loading state (for tests)

func (a *App) SetLoading(v bool) {
	a.isLoading = v
}

func (a *App) IsLoading() bool {
	return a.isLoading
}

// AddChatMessage adds a message to the active session (for tests).
func (a *App) AddChatMessage(role, content string) {
	if s := a.activeSessionPtr(); s != nil {
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
		s.AddMessage(r, content)
		a.chatPanel.SetSession(s)
	}
}

// GetChatMessages returns all messages from the active session (for tests).
func (a *App) GetChatMessages() []Message {
	if s := a.activeSessionPtr(); s != nil {
		return s.Messages
	}
	return nil
}

// SetComposerInput sets the composer input text (for tests).
func (a *App) SetComposerInput(input string) {
	a.composer.SetInput(input)
}

// GetComposerInput returns the composer input text (for tests).
func (a *App) GetComposerInput() string {
	return a.composer.GetInput()
}
