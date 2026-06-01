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
	ModeCommandPalette
)

type App struct {
	*tview.Application
	pages        *tview.Pages
	chatFlex     *tview.Flex
	statusBar    *status.StatusBar
	chatPanel    *ChatPanel
	composer     *composer.Composer
	tabDock      *tabbar.TabDock
	searchInput  *tview.InputField
	searchStatus *tview.TextView
	helpView     *tview.TextView
	sessions     []*Session
	activeSession int
	mode         AppMode
	isLoading    bool
	keyMap       KeyMap
	themeService *ThemeService
	aiAssistant  *service.AIAssistant
	inputHistory []string
	historyIndex int
	commandPalette  *CommandPalette
	commandRegistry *service.CommandRegistry
	skillExecutor   *service.SkillExecutor
	slashDetected   bool
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
	a.helpView.SetTextStyle(tcell.StyleDefault.Background(tcell.ColorDefault))

	// Build layout: StatusBar + ChatPanel + [CommandPalette] + Composer + TabDock
	chatFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	chatFlex.AddItem(a.statusBar, 1, 0, false)
	chatFlex.AddItem(a.chatPanel, 0, 1, false)
	a.commandPalette = NewCommandPalette()
	chatFlex.AddItem(a.commandPalette, 0, 0, false) // hidden by default
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

	// Register builtin commands
	a.registerBuiltinCommands()

	// Immediate slash detection
	a.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		if a.mode == ModeChat && a.composer != nil {
			text := a.composer.GetInput()
			if len(text) > 0 && text[0] == '/' {
				if !a.slashDetected {
					a.slashDetected = true
					a.enterCommandPalette(ShowSkills)
				}
			} else {
				a.slashDetected = false
			}
		}
		return false
	})

	a.SetRoot(a.pages, true)
	a.SetInputCapture(a.handleInput)
	a.SetFocus(a.composer)

	a.applyTheme()
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
	case ModeCommandPalette:
		if event.Key() == tcell.KeyEnter {
			cmd := a.commandPalette.SelectedCommand()
			if cmd != nil {
				a.executeCommand(cmd)
			}
			return nil
		}
		if event.Key() == tcell.KeyEsc {
			a.exitCommandPalette()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			cmd := a.commandPalette.SelectedCommand()
			if cmd != nil {
				text := a.composer.GetInput()
				if len(text) > 0 && text[0] == '/' {
					a.composer.SetInput("/" + cmd.Name + " ")
				}
				a.executeCommand(cmd)
			}
			return nil
		}
		return event
	}

	// Chat mode — global shortcuts only here
	switch {
	case event.Key() == tcell.KeyCtrlC:
		a.Stop()
		return nil
	case event.Key() == tcell.KeyCtrlF:
		a.enterSearch()
		return nil
	case event.Key() == tcell.KeyF1 || event.Key() == tcell.KeyCtrlO:
		a.enterHelp()
		return nil
	case event.Key() == tcell.KeyCtrlP:
		a.enterCommandPalette(ShowAll)
		return nil
	case event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModNone:
		if a.isLoading {
			return nil
		}
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
	case event.Key() == tcell.KeyUp:
		if a.historyIndex < len(a.inputHistory) {
			a.historyIndex++
			idx := len(a.inputHistory) - a.historyIndex
			a.composer.SetInput(a.inputHistory[idx])
		}
		return nil
	case event.Key() == tcell.KeyDown:
		if a.historyIndex > 0 {
			a.historyIndex--
			if a.historyIndex == 0 {
				a.composer.ClearInput()
			} else {
				idx := len(a.inputHistory) - a.historyIndex
				a.composer.SetInput(a.inputHistory[idx])
			}
		}
		return nil
	case event.Key() == tcell.KeyCtrlN:
		a.newSession()
		return nil
	case event.Key() == tcell.KeyCtrlW:
		a.closeSession()
		return nil
	case event.Key() == tcell.KeyCtrlR:
		a.renameSession()
		return nil
	case event.Key() == tcell.KeyTab:
		a.nextSession()
		return nil
	case event.Key() == tcell.KeyBacktab:
		a.prevSession()
		return nil
	case event.Key() == tcell.KeyCtrlT:
		if s := a.activeSessionPtr(); s != nil {
			s.ToggleThinking()
		}
		return nil
	case event.Key() == tcell.KeyCtrlY:
		if s := a.activeSessionPtr(); s != nil {
			s.ToggleCollapse()
		}
		return nil
	case event.Key() == tcell.KeyCtrlK:
		a.themeService.NextTheme()
		a.applyTheme()
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
	id := uuid.New().String()
	if a.aiAssistant != nil {
		id = a.aiAssistant.CreateSession("New Session")
	}
	s := NewSession(id, "New Session")
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
	a.inputHistory = append(a.inputHistory, text)
	a.historyIndex = 0
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
	a.chatPanel.SetSession(s)
	a.isLoading = true

	sessionID := s.ID
	sessionPtr := s // capture session pointer for goroutine
	go func() {
		var fullResponse string
		err := a.aiAssistant.ChatStream(sessionID, text, func(chunk string) {
			fullResponse += chunk
			a.QueueUpdateDraw(func() {
				// Use captured session pointer — always the correct session
				if len(sessionPtr.Messages) > 0 {
					sessionPtr.Messages[len(sessionPtr.Messages)-1].Content = fullResponse
					a.chatPanel.SetSession(sessionPtr)
				}
			})
		})

		a.QueueUpdateDraw(func() {
			a.isLoading = false
			if err != nil {
				sessionPtr.AddMessage(RoleSystem, "Error: "+err.Error())
			}
			a.chatPanel.SetSession(sessionPtr)
			a.chatPanel.ScrollToBottom()
			// Update tab label after first AI response
			label := sessionPtr.GenerateLabel()
			if label != "New Session" {
				sessionPtr.Label = label
				for i, s := range a.sessions {
					if s == sessionPtr {
						a.tabDock.UpdateTab(i, label)
						break
					}
				}
			}
		})
	}()
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
	a.composer.SetPromptColor(colors.InputPrompt)
	a.composer.SetAccentColor(hexToTCell(colors.Accent))
	a.tabDock.SetColors(tcell.ColorWhite, hexToTCell(colors.Accent), tcell.ColorGray, tcell.ColorDefault)
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

func (a *App) CommandRegistry() *service.CommandRegistry {
	return a.commandRegistry
}

func (a *App) SetSkillExecutor(executor *service.SkillExecutor) {
	a.skillExecutor = executor
}

func (a *App) enterCommandPalette(mode PaletteMode) {
	a.mode = ModeCommandPalette
	if a.commandRegistry != nil {
		cmds := a.commandRegistry.List()
		a.commandPalette.SetCommands(cmds)
	}
	a.commandPalette.SetMode(mode)
	a.chatFlex.RemoveItem(a.commandPalette)
	a.chatFlex.AddItem(a.commandPalette, 6, 0, false)
	a.SetFocus(a.commandPalette)
}

func (a *App) exitCommandPalette() {
	a.mode = ModeChat
	a.chatFlex.RemoveItem(a.commandPalette)
	a.chatFlex.AddItem(a.commandPalette, 0, 0, false)
	a.SetFocus(a.composer)
}

func (a *App) executeCommand(cmd *service.Command) {
	a.mode = ModeChat
	a.chatFlex.RemoveItem(a.commandPalette)
	a.chatFlex.AddItem(a.commandPalette, 0, 0, false)
	a.SetFocus(a.composer)

	switch cmd.Category {
	case service.CmdBuiltin:
		cmd.Action(a.composer.GetInput())
	case service.CmdSkill:
		a.executeSkill(cmd.Name)
	}
}

func (a *App) executeSkill(name string) {
	if a.skillExecutor == nil || a.activeSession < 0 {
		return
	}

	s := a.sessions[a.activeSession]
	text := a.composer.GetInput()
	pc := service.ParseCommand(text)
	if pc == nil {
		pc = &service.ParsedCommand{Name: name}
	}

	result, err := a.skillExecutor.Execute(pc)
	if err != nil {
		s.AddMessage(RoleSkill, "Error: "+err.Error())
	} else {
		s.AddMessage(RoleSkill, result)
	}
	s.Messages[len(s.Messages)-1].Label = name
	a.composer.ClearInput()
	a.chatPanel.SetSession(s)
}

func (a *App) registerBuiltinCommands() {
	r := service.NewCommandRegistry()

	r.Register(&service.Command{
		Name:        "New Session",
		Description: "Create a new chat session",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.newSession() },
	})
	r.Register(&service.Command{
		Name:        "Search",
		Description: "Search messages in current session",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.enterSearch() },
	})
	r.Register(&service.Command{
		Name:        "Toggle Thinking",
		Description: "Expand or collapse thinking blocks",
		Category:    service.CmdBuiltin,
		Action:      func(string) { if s := a.activeSessionPtr(); s != nil { s.ToggleThinking() } },
	})
	r.Register(&service.Command{
		Name:        "Toggle Collapse",
		Description: "Collapse or expand the last message",
		Category:    service.CmdBuiltin,
		Action:      func(string) { if s := a.activeSessionPtr(); s != nil { s.ToggleCollapse() } },
	})
	r.Register(&service.Command{
		Name:        "Next Theme",
		Description: "Switch to the next color theme",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.themeService.NextTheme(); a.applyTheme() },
	})
	r.Register(&service.Command{
		Name:        "Next Session",
		Description: "Switch to the next session tab",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.nextSession() },
	})
	r.Register(&service.Command{
		Name:        "Close Session",
		Description: "Close the current session",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.closeSession() },
	})
	r.Register(&service.Command{
		Name:        "Help",
		Description: "Show keyboard shortcuts",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.enterHelp() },
	})

	a.commandRegistry = r
}
