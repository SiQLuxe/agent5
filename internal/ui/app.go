package ui

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/rivo/tview"

	"github.com/example/agent-tui/internal/agent/orchestrator"
	"github.com/example/agent-tui/internal/agent/runtime"
	"github.com/example/agent-tui/internal/agent/session"
	"github.com/example/agent-tui/internal/service"
	"github.com/example/agent-tui/internal/ui/composer"
	"github.com/example/agent-tui/internal/ui/status"
	"github.com/example/agent-tui/internal/ui/suggestion"
	"github.com/example/agent-tui/internal/ui/tabbar"
)

type AppMode int

const (
	ModeChat AppMode = iota
	ModeSearch
	ModeHelp
	ModeCommandPalette
	ModeRename
	ModeSkill
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
	sessionMgr  *session.Manager
	primaryAgent *runtime.Agent
	inputHistory []string
	historyIndex int
	commandPalette  *CommandPalette
	commandRegistry *service.CommandRegistry
	skillExecutor   *service.SkillExecutor
	skillRegistry   *service.SkillRegistry
	skillsDir       string
	suggestionMenu *suggestion.SuggestionMenu
	renameInput     *tview.InputField
	renamePage      *tview.Flex
	approvalModal   *ApprovalModal
	skillOverlay    *SkillOverlay
	orch            *orchestrator.Orchestrator
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

	// Build layout: StatusBar + ChatPanel + Composer + SuggestionBar + TabDock
	chatFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	chatFlex.AddItem(a.statusBar, 1, 0, false)
	chatFlex.AddItem(a.chatPanel, 0, 1, false)
	a.commandPalette = NewCommandPalette()
	a.suggestionMenu = suggestion.New()
	chatFlex.AddItem(a.suggestionMenu, 1, 0, false)
	chatFlex.AddItem(a.composer, 3, 0, true)
	chatFlex.AddItem(a.tabDock, 1, 0, false)
	a.chatFlex = chatFlex

	// Composer text change handler for command suggestions
	a.composer.SetOnTextChanged(func(text string) {
		a.onComposerChange(text)
	})

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

	// Rename overlay
	a.renameInput = tview.NewInputField()
	a.renameInput.SetLabel("Session name: ")
	a.renameInput.SetFieldWidth(0)
	renameFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	renameFlex.AddItem(nil, 0, 1, false)
	renameInner := tview.NewFlex().SetDirection(tview.FlexRow)
	renameInner.AddItem(a.renameInput, 1, 0, true)
	renameInner.AddItem(tview.NewTextView().SetText("[gray]Enter: confirm  Esc: cancel[-]").SetDynamicColors(true).SetTextAlign(tview.AlignCenter), 1, 0, false)
	renameFlex.AddItem(renameInner, 3, 0, true)
	renameFlex.AddItem(nil, 0, 1, false)
	renamePage := tview.NewFlex().SetDirection(tview.FlexColumn)
	renamePage.AddItem(nil, 0, 1, false)
	renamePage.AddItem(renameFlex, 50, 0, true)
	renamePage.AddItem(nil, 0, 1, false)
	a.renamePage = renamePage
	a.pages.AddPage("rename", renamePage, true, false)

	// Approval overlay page
	a.approvalModal = NewApprovalModal()
	a.pages.AddPage("approval", a.approvalModal, true, false)

	// Skill overlay page
	a.skillOverlay = NewSkillOverlay(a)
	skillFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	skillFlex.AddItem(nil, 0, 1, false)
	skillInner := tview.NewFlex().SetDirection(tview.FlexRow)
	skillInner.AddItem(nil, 0, 1, false)
	skillInner.AddItem(a.skillOverlay, 0, 3, true)
	skillInner.AddItem(nil, 0, 1, false)
	skillFlex.AddItem(skillInner, 80, 0, true)
	skillFlex.AddItem(nil, 0, 1, false)
	a.pages.AddPage("skill", skillFlex, true, false)

	// Register builtin commands
	a.registerBuiltinCommands()

	a.SetRoot(a.pages, true)
	a.SetInputCapture(a.handleInput)
	a.EnableMouse(true)
	a.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		return event, action
	})
	a.chatPanel.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		switch action {
		case tview.MouseScrollUp:
			a.chatPanel.SetAutoScroll(false)
			a.chatPanel.ScrollUp(3)
			return tview.MouseConsumed, nil
		case tview.MouseScrollDown:
			a.chatPanel.ScrollDown(3)
			return tview.MouseConsumed, nil
		}
		return action, event
	})
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
	case ModeRename:
		if event.Key() == tcell.KeyEnter {
			a.applyRename()
			return nil
		}
		if event.Key() == tcell.KeyEsc {
			a.exitRename()
			return nil
		}
		return event
	case ModeCommandPalette:
		if event.Key() == tcell.KeyEnter {
			cmd := a.commandPalette.SelectedCommand()
			if cmd != nil {
				a.executeCommand(cmd)
			}
			return nil
		}
		if event.Key() == tcell.KeyEsc {
			if a.commandPalette.FilterText() != "" {
				a.commandPalette.SetFilter("")
				a.commandPalette.GetFilterInput().SetText("")
			} else {
				a.exitCommandPalette()
			}
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
		if event.Key() == tcell.KeyDown {
			a.commandPalette.SelectNext()
			return nil
		}
		if event.Key() == tcell.KeyUp {
			a.commandPalette.SelectPrev()
			return nil
		}
		if event.Key() == tcell.KeyPgDn {
			for i := 0; i < 5; i++ {
				a.commandPalette.SelectNext()
			}
			return nil
		}
		if event.Key() == tcell.KeyPgUp {
			for i := 0; i < 5; i++ {
				a.commandPalette.SelectPrev()
			}
			return nil
		}
		return event
	case ModeSkill:
		if event.Key() == tcell.KeyCtrlS {
			a.exitSkillOverlay()
			return nil
		}
		// Handled by SkillOverlay's own InputCapture
		return nil
	}

	// Chat mode — global shortcuts only here
	switch {
	case event.Key() == tcell.KeyEsc:
		if a.suggestionMenu.Visible() {
			a.hideSuggestions()
			return nil
		}
	case event.Key() == tcell.KeyCtrlC:
		if a.chatPanel.HasSelection() {
			a.chatPanel.CopySelection()
			return nil
		}
		a.Stop()
		return nil
	case event.Key() == tcell.KeyCtrlF:
		a.enterSearch()
		return nil
	case event.Key() == tcell.KeyF1 || event.Key() == tcell.KeyCtrlO:
		a.enterHelp()
		return nil
	case event.Key() == tcell.KeyCtrlS:
		a.enterSkillOverlay()
		return nil
	case event.Key() == tcell.KeyCtrlP:
		a.enterCommandPalette(ShowAll)
		return nil
	case event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModNone:
		if a.suggestionMenu.Visible() {
			cmd := a.suggestionMenu.Selected()
			if cmd != nil && cmd.Action != nil && cmd.Category == service.CmdBuiltin {
				cmd.Action(cmd.Name)
				a.composer.ClearInput()
				a.hideSuggestions()
				return nil
			}
			a.hideSuggestions()
		}
		if a.isLoading {
			return nil
		}
		if strings.TrimSpace(a.composer.GetInput()) == "" {
			return nil
		}
		a.sendMessage()
		return nil
	case event.Key() == tcell.KeyEnter && (event.Modifiers()&tcell.ModCtrl != 0 || event.Modifiers()&tcell.ModAlt != 0):
		// Ctrl+Enter or Alt+Enter: insert newline (let fall through to TextArea)
		return event
	case event.Key() == tcell.KeyTab && a.suggestionMenu.Visible():
		cmd := a.suggestionMenu.Selected()
		if cmd != nil {
			a.composer.SetInput("/" + cmd.Name + " ")
			a.hideSuggestions()
		}
		return nil
	case event.Key() == tcell.KeyPgUp:
		a.chatPanel.ScrollUp(10)
		a.chatPanel.SetAutoScroll(false)
		return nil
	case event.Key() == tcell.KeyPgDn:
		a.chatPanel.ScrollDown(10)
		return nil
	case event.Rune() == 'g' && event.Modifiers() == tcell.ModNone:
		a.chatPanel.ScrollToTop()
		a.chatPanel.SetAutoScroll(false)
		return nil
	case event.Rune() == 'G' && event.Modifiers() == tcell.ModNone:
		a.chatPanel.ScrollToBottom()
		a.chatPanel.SetAutoScroll(true)
		return nil
	case event.Key() == tcell.KeyUp:
		if a.suggestionMenu.Visible() {
			a.suggestionMenu.SelectPrev()
		} else if a.historyIndex < len(a.inputHistory) {
			a.historyIndex++
			idx := len(a.inputHistory) - a.historyIndex
			a.composer.SetInput(a.inputHistory[idx])
		}
		return nil
	case event.Key() == tcell.KeyDown:
		if a.suggestionMenu.Visible() {
			a.suggestionMenu.SelectNext()
		} else if a.historyIndex > 0 {
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
	case event.Rune() == '\u00f1' || event.Rune() == '\u00d1': // Option+n/N → ñ/Ñ
		a.newSession()
		return nil
	case event.Rune() == '\u2020' || event.Rune() == '\u2021': // Option+w/W → †/‡
		a.closeSession()
		return nil
	case event.Rune() == '\u00ae': // Option+r → ®
		a.renameSession()
		return nil
	case event.Rune() == '\u2265': // Option+. → ≥
		a.nextSession()
		return nil
	case event.Rune() == '\u2264': // Option+, → ≤
		a.prevSession()
		return nil
	case event.Rune() == '\u2122': // Option+t → ™
		if s := a.activeSessionPtr(); s != nil {
			s.ToggleThinking()
		}
		return nil
	case event.Rune() == '\u00a5': // Option+y → ¥
		if s := a.activeSessionPtr(); s != nil {
			s.ToggleCollapse()
		}
		return nil
	case event.Rune() == '\u00df': // Option+s → ß
		a.themeService.NextTheme()
		a.applyTheme()
		return nil
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
	a.onComposerChange(a.composer.GetInput())
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
	a.onComposerChange(a.composer.GetInput())
}

// Session management

func (a *App) newSession() {
	id := uuid.New().String()
	if a.sessionMgr != nil {
		id = a.sessionMgr.CreateSession("New Session")
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
	a.SetFocus(a.composer)
}

func (a *App) renameSession() {
	a.enterRename()
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
	a.SetFocus(a.composer)
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
	a.hideSuggestions()

	s := a.activeSessionPtr()
	if s == nil {
		return
	}

	// /command → skill executor
	if strings.HasPrefix(text, "/") {
		pc := service.ParseCommand(text)
		if pc != nil && a.skillExecutor != nil {
			result, err := a.skillExecutor.Execute(pc)
			if err != nil {
				s.AddMessage(RoleSkill, "Error: "+err.Error())
			} else {
				s.AddMessage(RoleSkill, result)
			}
			s.Messages[len(s.Messages)-1].Label = pc.Name
		}
		a.composer.ClearInput()
		a.chatPanel.SetSession(s)
		return
	}

	// Normal text → dispatch through orchestrator (if configured), fallback to primary agent or session manager
	if a.orch == nil && a.primaryAgent == nil && a.sessionMgr == nil {
		// No agent system or AI client configured — cannot handle this message
		return
	}
	s.AddMessage(RoleUser, text)
	s.AddMessage(RoleAssistant, "")
	a.composer.ClearInput()
	a.chatPanel.SetSession(s)
	a.chatPanel.StartStreaming()
	a.isLoading = true

	sessionPtr := s
	go func() {
		if a.orch != nil {
			task := &orchestrator.Task{
				ID:      uuid.New().String(),
				Type:    orchestrator.TaskExecute,
				Content: text,
			}

			var streamBuf string
			results, err := a.orch.DispatchStream(sessionPtr.ID, task, func(chunk string) {
				a.QueueUpdateDraw(func() {
					streamBuf += chunk
					if len(sessionPtr.Messages) > 0 {
						sessionPtr.Messages[len(sessionPtr.Messages)-1].Content = streamBuf
						a.chatPanel.UpdateStreaming(streamBuf)
					}
				})
			})

			a.QueueUpdateDraw(func() {
				a.isLoading = false
				// Remove the placeholder assistant message
				if len(sessionPtr.Messages) > 0 && sessionPtr.Messages[len(sessionPtr.Messages)-1].Role == RoleAssistant {
					sessionPtr.Messages = sessionPtr.Messages[:len(sessionPtr.Messages)-1]
				}
				if err != nil {
					sessionPtr.AddMessage(RoleSystem, "Agent error: "+err.Error())
				}
				for _, r := range results {
					statusStr := string(r.Status)
					header := "[" + statusStr + "] " + string(r.Type) + " (agent: " + r.AgentID + ")"
					if r.Error != "" {
						sessionPtr.AddMessage(RoleSkill, header+"\n"+r.Error)
					} else if r.Result != "" {
						sessionPtr.AddMessage(RoleSkill, header+"\n"+r.Result)
					}
				}
				a.chatPanel.EndStreaming()
			})
			return
		}

		// Direct agent or session manager fallback
		if a.primaryAgent != nil {
			var streamBuf string
			_, err := a.primaryAgent.ExecuteStream(sessionPtr.ID, text, func(chunk string) {
				streamBuf += chunk
				a.QueueUpdateDraw(func() {
					if len(sessionPtr.Messages) > 0 {
						sessionPtr.Messages[len(sessionPtr.Messages)-1].Content = streamBuf
						a.chatPanel.UpdateStreaming(streamBuf)
					}
				})
			})
			a.QueueUpdateDraw(func() {
				a.isLoading = false
				if err != nil {
					sessionPtr.AddMessage(RoleSystem, "Error: "+err.Error())
				}
				a.chatPanel.EndStreaming()
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
		} else if a.sessionMgr != nil {
			var fullResponse string
			err := a.sessionMgr.ChatStream(sessionPtr.ID, text, func(chunk string) {
				fullResponse += chunk
				a.QueueUpdateDraw(func() {
					if len(sessionPtr.Messages) > 0 {
						sessionPtr.Messages[len(sessionPtr.Messages)-1].Content = fullResponse
						a.chatPanel.UpdateStreaming(fullResponse)
					}
				})
			})
			a.QueueUpdateDraw(func() {
				a.isLoading = false
				if err != nil {
					sessionPtr.AddMessage(RoleSystem, "Error: "+err.Error())
				}
				a.chatPanel.EndStreaming()
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
		}
	}()
}

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

func (a *App) showSuggestions() {
	a.suggestionMenu.Show()
	a.chatFlex.ResizeItem(a.suggestionMenu, a.suggestionMenu.Height(), 0)
}

func (a *App) hideSuggestions() {
	a.suggestionMenu.Hide()
	a.chatFlex.ResizeItem(a.suggestionMenu, 0, 0)
}

func (a *App) AddWelcomeMessage() {
	a.newSession()
	if s := a.activeSessionPtr(); s != nil {
		s.AddMessage(RoleSystem, "Hello! Welcome to the Agent TUI.")
	}
	a.chatPanel.SetSession(a.activeSessionPtr())
}

// Session manager and agent setters

func (a *App) SetSessionManager(sm *session.Manager) {
	a.sessionMgr = sm
}

func (a *App) SetPrimaryAgent(agent *runtime.Agent) {
	a.primaryAgent = agent
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

// TUIActionHandler implementations (for reverse-control server).

func (a *App) AppendPrompt(text string) {
	a.QueueUpdateDraw(func() {
		current := a.composer.GetInput()
		a.composer.SetInput(current + text)
	})
}

func (a *App) SubmitPrompt() {
	a.QueueUpdateDraw(func() {
		a.sendMessage()
	})
}

func (a *App) ShowToast(title, message, variant string) {
	a.QueueUpdateDraw(func() {
		a.statusBar.ShowMessage(message)
	})
	go func() {
		time.Sleep(3 * time.Second)
		a.QueueUpdateDraw(func() {
			a.statusBar.ClearMessage()
		})
	}()
}

func (a *App) ExecuteCommand(command string) {
	for _, cmd := range a.commandRegistry.List() {
		if cmd.Name == command {
			a.QueueUpdateDraw(func() {
				a.executeCommand(cmd)
			})
			return
		}
	}
}

func (a *App) CommandRegistry() *service.CommandRegistry {
	return a.commandRegistry
}

func (a *App) SetOrchestrator(orch *orchestrator.Orchestrator) {
	a.orch = orch
}

func (a *App) SetSkillExecutor(executor *service.SkillExecutor) {
	a.skillExecutor = executor
}

func (a *App) SetSkillRegistry(r *service.SkillRegistry) {
	a.skillRegistry = r
}

func (a *App) SetSkillsDir(dir string) {
	a.skillsDir = dir
}

func (a *App) ShowApproval() {
	a.pages.ShowPage("approval")
	a.SetFocus(a.approvalModal)
}

func (a *App) HideApproval() {
	a.pages.HidePage("approval")
}

func (a *App) ApprovalModal() *ApprovalModal {
	return a.approvalModal
}

func (a *App) enterCommandPalette(mode PaletteMode) {
	a.mode = ModeCommandPalette
	if a.commandRegistry != nil {
		cmds := a.commandRegistry.List()
		a.commandPalette.SetCommands(cmds)
	}
	a.commandPalette.SetMode(mode)
	a.commandPalette.SetFilter("")
	a.commandPalette.GetFilterInput().SetText("")
	a.pages.SwitchToPage("command")
	a.SetFocus(a.commandPalette.GetFilterInput())
}

func (a *App) exitCommandPalette() {
	a.mode = ModeChat
	a.pages.SwitchToPage("chat")
	a.SetFocus(a.composer)
	a.onComposerChange(a.composer.GetInput())
}

func (a *App) enterSkillOverlay() {
	a.mode = ModeSkill
	a.skillOverlay.LoadSkills()
	a.pages.SwitchToPage("skill")
	a.SetFocus(a.skillOverlay)
}

func (a *App) exitSkillOverlay() {
	a.mode = ModeChat
	a.pages.SwitchToPage("chat")
	a.SetFocus(a.composer)
	a.onComposerChange(a.composer.GetInput())
}

func (a *App) executeCommand(cmd *service.Command) {
	a.mode = ModeChat
	a.pages.SwitchToPage("chat")
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

func (a *App) enterRename() {
	s := a.activeSessionPtr()
	if s == nil {
		return
	}
	a.mode = ModeRename
	a.renameInput.SetText(s.Label)
	a.pages.SwitchToPage("rename")
	a.SetFocus(a.renameInput)
}

func (a *App) applyRename() {
	s := a.activeSessionPtr()
	if s == nil {
		a.exitRename()
		return
	}
	newName := a.renameInput.GetText()
	if newName == "" {
		newName = "New Session"
	}
	s.Label = newName
	for i, session := range a.sessions {
		if session == s {
			a.tabDock.UpdateTab(i, newName)
			break
		}
	}
	a.exitRename()
}

func (a *App) exitRename() {
	a.mode = ModeChat
	a.pages.SwitchToPage("chat")
	a.SetFocus(a.composer)
	a.onComposerChange(a.composer.GetInput())
}

func (a *App) reloadSkills() {
	if a.skillRegistry == nil || a.skillsDir == "" {
		return
	}
	if err := service.ReloadSkillsDir(a.skillRegistry, a.skillsDir); err != nil {
		return
	}
	if a.commandRegistry != nil {
		a.commandRegistry.ClearCategory(service.CmdSkill)
		a.commandRegistry.SyncSkills(a.skillRegistry)
	}
}

func (a *App) StartSkillWatcher() {
	if a.skillsDir == "" {
		return
	}
	go func() {
		mtimes := make(map[string]time.Time)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			changed := false
			filepath.Walk(a.skillsDir, func(path string, fi os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if fi.IsDir() || fi.Name() != "SKILL.md" {
					return nil
				}
				if old, ok := mtimes[path]; !ok || fi.ModTime() != old {
					changed = true
					mtimes[path] = fi.ModTime()
				}
				return nil
			})
			if changed {
				a.QueueUpdateDraw(func() { a.reloadSkills() })
			}
		}
	}()
}

func (a *App) registerBuiltinCommands() {
	r := service.NewCommandRegistry()

	r.Register(&service.Command{
		Name:        "New Session",
		Description: "新建会话",
		Shortcut:    "Ctrl+N / Alt+N",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.newSession() },
	})
	r.Register(&service.Command{
		Name:        "Close Session",
		Description: "关闭当前会话",
		Shortcut:    "Ctrl+W / Alt+W",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.closeSession() },
	})
	r.Register(&service.Command{
		Name:        "Next Session",
		Description: "切换到下一个会话",
		Shortcut:    "Alt+.",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.nextSession() },
	})
	r.Register(&service.Command{
		Name:        "Previous Session",
		Description: "切换到上一个会话",
		Shortcut:    "Alt+,",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.prevSession() },
	})
	r.Register(&service.Command{
		Name:        "Rename Session",
		Description: "重命名当前会话",
		Shortcut:    "Ctrl+R / Alt+R",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.enterRename() },
	})
	r.Register(&service.Command{
		Name:        "Search",
		Description: "搜索当前会话中的消息",
		Shortcut:    "Ctrl+F",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.enterSearch() },
	})
	r.Register(&service.Command{
		Name:        "Help",
		Description: "显示键盘快捷键列表",
		Shortcut:    "F1 / Ctrl+O",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.enterHelp() },
	})
	r.Register(&service.Command{
		Name:        "Toggle Thinking",
		Description: "展开或收起思考过程",
		Shortcut:    "Ctrl+T / Alt+T",
		Category:    service.CmdBuiltin,
		Action:      func(string) { if s := a.activeSessionPtr(); s != nil { s.ToggleThinking() } },
	})
	r.Register(&service.Command{
		Name:        "Toggle Collapse",
		Description: "折叠或展开最后一条消息",
		Shortcut:    "Ctrl+Y / Alt+Y",
		Category:    service.CmdBuiltin,
		Action:      func(string) { if s := a.activeSessionPtr(); s != nil { s.ToggleCollapse() } },
	})
	r.Register(&service.Command{
		Name:        "Next Theme",
		Description: "切换到下一个颜色主题",
		Shortcut:    "Ctrl+K / Alt+S",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.themeService.NextTheme(); a.applyTheme() },
	})
	r.Register(&service.Command{
		Name:        "Scroll to Top",
		Description: "滚动到消息顶部",
		Shortcut:    "g",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.chatPanel.ScrollToTop() },
	})
	r.Register(&service.Command{
		Name:        "Scroll to Bottom",
		Description: "滚动到消息底部",
		Shortcut:    "G",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.chatPanel.ScrollToBottom() },
	})
	r.Register(&service.Command{
		Name:        "Clear Input",
		Description: "清空输入框",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.composer.ClearInput() },
	})
	r.Register(&service.Command{
		Name:        "Reload Skills",
		Description: "从磁盘重新加载技能",
		Category:    service.CmdBuiltin,
		Action:      func(string) { a.reloadSkills() },
	})

	a.commandRegistry = r
}
