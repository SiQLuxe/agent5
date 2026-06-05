package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/example/agent-tui/internal/ai"
	"github.com/example/agent-tui/internal/agent/session"
	"github.com/example/agent-tui/internal/service"
)

type MockAIClientForApp struct {
	mockResponse string
}

func (m *MockAIClientForApp) ChatCompletion(req ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
	return &ai.ChatCompletionResponse{}, nil
}

func (m *MockAIClientForApp) ChatCompletionStream(req ai.ChatCompletionRequest, callback func(string)) error {
	if m.mockResponse != "" {
		callback(m.mockResponse)
	}
	return nil
}

func (m *MockAIClientForApp) SetAPIKey(key string)    {}
func (m *MockAIClientForApp) SetBaseURL(url string)    {}
func (m *MockAIClientForApp) GetModel() string          { return "mock" }
func (m *MockAIClientForApp) ListModels() ([]string, error) { return []string{"mock"}, nil }

func TestNewApp(t *testing.T) {
	a := NewApp()
	if a == nil {
		t.Fatal("NewApp returned nil")
	}
	if a.Application == nil {
		t.Fatal("expected non-nil Application")
	}
}

func TestNewAppSessions(t *testing.T) {
	a := NewApp()
	if len(a.sessions) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(a.sessions))
	}
}

func TestNewAppHasKeyMap(t *testing.T) {
	a := NewApp()
	if a.keyMap.NewSession == 0 {
		t.Fatal("expected non-zero key map")
	}
}

func TestNewAppHasThemeService(t *testing.T) {
	a := NewApp()
	if a.themeService == nil {
		t.Fatal("expected non-nil theme service")
	}
}

func TestAppNewSession(t *testing.T) {
	a := NewApp()
	a.newSession()
	if len(a.sessions) != 1 {
		t.Fatalf("expected 1 session after NewSession, got %d", len(a.sessions))
	}
	if a.activeSession != 0 {
		t.Fatalf("expected active session 0, got %d", a.activeSession)
	}
}

func TestSwitchSession(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.newSession()
	a.switchToSession(1)
	if a.activeSession != 1 {
		t.Fatalf("expected active session 1, got %d", a.activeSession)
	}
}

func TestSwitchSessionOutOfRange(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.switchToSession(5)
	a.switchToSession(-1)
	if a.activeSession != 0 {
		t.Fatalf("expected active session 0, got %d", a.activeSession)
	}
}

func TestNextSession(t *testing.T) {
	a := NewApp()
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
	a := NewApp()
	a.newSession()
	a.newSession()
	a.switchToSession(0)
	a.prevSession()
	if a.activeSession != 1 {
		t.Fatalf("expected active session 1 (wrap), got %d", a.activeSession)
	}
}

func TestCloseSession(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.newSession()
	a.closeSession()
	if len(a.sessions) != 1 {
		t.Fatalf("expected 1 session after close, got %d", len(a.sessions))
	}
}

func TestCloseSessionLastRemaining(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.closeSession()
	if len(a.sessions) != 1 {
		t.Fatalf("expected 1 session (can't close last), got %d", len(a.sessions))
	}
}

func TestSendMessage(t *testing.T) {
	a := NewApp()
	a.newSession()
	mockClient := &MockAIClientForApp{
		mockResponse: "Hello!",
	}
	sm := session.NewManager()
	sm.SetClient(mockClient)
	a.SetSessionManager(sm)

	a.composer.SetInput("hello")
	a.sendMessage()

	if a.composer.GetInput() != "" {
		t.Fatal("expected composer cleared after send")
	}
	if s := a.activeSessionPtr(); s != nil {
		msgs := s.Messages
		if len(msgs) != 2 {
			t.Fatalf("expected 2 messages (user + assistant), got %d", len(msgs))
		}
		if msgs[0].Role != RoleUser || msgs[0].Content != "hello" {
			t.Fatalf("unexpected first message: %+v", msgs[0])
		}
		if msgs[1].Role != RoleAssistant {
			t.Fatalf("expected assistant role, got %v", msgs[1].Role)
		}
	}
	if !a.IsLoading() {
		t.Fatal("expected loading after send")
	}
}

func TestSendMessageNoAssistant(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.composer.SetInput("hello")
	a.sendMessage()
	if a.composer.GetInput() != "hello" {
		t.Fatal("expected composer preserved when no ai assistant")
	}
	if s := a.activeSessionPtr(); s != nil {
		if len(s.Messages) != 0 {
			t.Fatalf("expected 0 messages with no ai assistant, got %d", len(s.Messages))
		}
	}
}

func TestSendMessageEmpty(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.sendMessage()
	if s := a.activeSessionPtr(); s != nil {
		msgs := s.Messages
		if len(msgs) != 0 {
			t.Fatalf("expected 0 messages for empty send, got %d", len(msgs))
		}
	}
}

func TestEnterExitSearch(t *testing.T) {
	a := NewApp()
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
	a := NewApp()
	a.enterHelp()
	if a.mode != ModeHelp {
		t.Fatalf("expected ModeHelp, got %d", a.mode)
	}
	a.exitHelp()
	if a.mode != ModeChat {
		t.Fatalf("expected ModeChat after exit, got %d", a.mode)
	}
}

func TestAddWelcomeMessage(t *testing.T) {
	a := NewApp()
	a.AddWelcomeMessage()
	if len(a.sessions) != 1 {
		t.Fatalf("expected 1 session after welcome, got %d", len(a.sessions))
	}
	if s := a.activeSessionPtr(); s != nil {
		if len(s.Messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(s.Messages))
		}
	}
}

func TestSetLoading(t *testing.T) {
	a := NewApp()
	if a.IsLoading() {
		t.Fatal("expected not loading initially")
	}
	a.SetLoading(true)
	if !a.IsLoading() {
		t.Fatal("expected loading after SetLoading(true)")
	}
}

func TestSendMessageStartsStreaming(t *testing.T) {
	a := NewApp()
	a.newSession()
	mockClient := &MockAIClientForApp{
		mockResponse: "Hello world",
	}
	sm := session.NewManager()
	sm.SetClient(mockClient)
	a.SetSessionManager(sm)

	a.composer.SetInput("hi")
	a.sendMessage()

	s := a.activeSessionPtr()
	// Before goroutine runs: user + empty assistant messages
	if len(s.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(s.Messages))
	}
	if s.Messages[0].Role != RoleUser || s.Messages[0].Content != "hi" {
		t.Fatalf("unexpected first message: %+v", s.Messages[0])
	}
	if s.Messages[1].Role != RoleAssistant || s.Messages[1].Content != "" {
		t.Fatalf("expected empty assistant placeholder, got: %+v", s.Messages[1])
	}
	if !a.IsLoading() {
		t.Fatal("expected loading during stream")
	}
	if a.composer.GetInput() != "" {
		t.Fatal("expected composer cleared")
	}
}

func TestNewSessionRegistersInHistory(t *testing.T) {
	a := NewApp()
	mockClient := &MockAIClientForApp{
		mockResponse: "Hello!",
	}
	sm := session.NewManager()
	sm.SetClient(mockClient)
	a.SetSessionManager(sm)
	a.newSession()

	s := a.activeSessionPtr()
	if s == nil {
		t.Fatal("expected active session")
	}

	// SessionManager should have this session registered
	sessions := sm.ListSessions()
	found := false
	for _, si := range sessions {
		if si.ID == s.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("BUG: session not registered in session manager - ChatStream AddMessage will silently fail, API receives empty messages")
	}

	// Simulate what ChatStream does - AddMessage should succeed now
	err := sm.AddMessage(s.ID, "user", "hello")
	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}
	ctx := sm.GetContext(s.ID, 0)
	if len(ctx) != 1 {
		t.Fatalf("expected 1 message in context, got %d", len(ctx))
	}
	if ctx[0].Role != "user" || ctx[0].Content != "hello" {
		t.Fatalf("unexpected message: %+v", ctx[0])
	}
}

func TestShortcutNewSession_CtrlN(t *testing.T) {
	a := NewApp()
	ev := tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Ctrl+N to be consumed (nil)")
	}
	if len(a.sessions) != 1 {
		t.Fatalf("expected 1 session after Ctrl+N, got %d", len(a.sessions))
	}
}

func TestShortcutCloseSession_CtrlW(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.newSession()
	ev := tcell.NewEventKey(tcell.KeyCtrlW, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Ctrl+W to be consumed (nil)")
	}
	if len(a.sessions) != 1 {
		t.Fatalf("expected 1 session after close, got %d", len(a.sessions))
	}
}

func TestRenameSession(t *testing.T) {
	a := NewApp()
	a.newSession()
	s := a.activeSessionPtr()
	s.Label = "Old Name"

	a.enterRename()
	if a.mode != ModeRename {
		t.Fatalf("expected ModeRename, got %d", a.mode)
	}
	if a.renameInput.GetText() != "Old Name" {
		t.Fatalf("expected 'Old Name', got %q", a.renameInput.GetText())
	}
}

func TestRenameSessionApply(t *testing.T) {
	a := NewApp()
	a.newSession()
	s := a.activeSessionPtr()
	s.Label = "Old Name"

	a.enterRename()
	a.renameInput.SetText("New Name")
	a.applyRename()

	if a.mode != ModeChat {
		t.Fatalf("expected ModeChat after rename, got %d", a.mode)
	}
	if s.Label != "New Name" {
		t.Fatalf("expected 'New Name', got %q", s.Label)
	}
}

func TestRenameSessionCancel(t *testing.T) {
	a := NewApp()
	a.newSession()
	s := a.activeSessionPtr()
	s.Label = "Original"

	a.enterRename()
	a.renameInput.SetText("Changed")
	a.exitRename()

	if a.mode != ModeChat {
		t.Fatalf("expected ModeChat after exit, got %d", a.mode)
	}
	if s.Label != "Original" {
		t.Fatalf("expected unchanged label 'Original', got %q", s.Label)
	}
}

func TestRenameSessionEmptyName(t *testing.T) {
	a := NewApp()
	a.newSession()
	s := a.activeSessionPtr()
	s.Label = "Old"

	a.enterRename()
	a.renameInput.SetText("")
	a.applyRename()

	if s.Label != "New Session" {
		t.Fatalf("expected default 'New Session', got %q", s.Label)
	}
}

func TestShortcutRenameSession_CtrlR(t *testing.T) {
	a := NewApp()
	a.newSession()
	s := a.activeSessionPtr()
	s.Label = "Test Session"
	ev := tcell.NewEventKey(tcell.KeyCtrlR, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Ctrl+R to be consumed (nil)")
	}
	if a.mode != ModeRename {
		t.Fatalf("expected ModeRename, got %d", a.mode)
	}
}

func TestShortcutNextSession_AltDot(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.newSession()
	a.switchToSession(0)
	ev := tcell.NewEventKey(tcell.KeyRune, '.', tcell.ModAlt)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Alt+. to be consumed (nil)")
	}
	if a.activeSession != 1 {
		t.Fatalf("expected active session 1, got %d", a.activeSession)
	}
}

func TestShortcutPrevSession_AltComma(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.newSession()
	a.switchToSession(0)
	ev := tcell.NewEventKey(tcell.KeyRune, ',', tcell.ModAlt)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Alt+, to be consumed (nil)")
	}
	if a.activeSession != 1 {
		t.Fatalf("expected active session 1 (wrap), got %d", a.activeSession)
	}
}

func TestShortcutNextSession_MacOSOptionDot(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.newSession()
	a.switchToSession(0)
	ev := tcell.NewEventKey(tcell.KeyRune, '\u2265', tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Option+. (≥) to be consumed (nil)")
	}
	if a.activeSession != 1 {
		t.Fatalf("expected active session 1, got %d", a.activeSession)
	}
}

func TestShortcutPrevSession_MacOSOptionComma(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.newSession()
	a.switchToSession(0)
	ev := tcell.NewEventKey(tcell.KeyRune, '\u2264', tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Option+, (≤) to be consumed (nil)")
	}
	if a.activeSession != 1 {
		t.Fatalf("expected active session 1 (wrap), got %d", a.activeSession)
	}
}

func TestTabPassesThrough(t *testing.T) {
	a := NewApp()
	ev := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result == nil {
		t.Fatal("expected Tab to pass through (non-nil)")
	}
}

func TestShortcutToggleThinking_CtrlT(t *testing.T) {
	a := NewApp()
	a.newSession()
	s := a.activeSessionPtr()
	s.Messages = append(s.Messages, Message{Role: RoleAssistant, Content: "test"})
	s.Messages[0].Thinking = &Thinking{Content: "thinking...", Expanded: false}
	ev := tcell.NewEventKey(tcell.KeyCtrlT, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Ctrl+T to be consumed (nil)")
	}
	if !s.Messages[0].Thinking.Expanded {
		t.Fatal("expected Thinking.Expanded to be true after toggle")
	}
}

func TestShortcutToggleCollapse_CtrlY(t *testing.T) {
	a := NewApp()
	a.newSession()
	s := a.activeSessionPtr()
	s.Messages = append(s.Messages, Message{Role: RoleAssistant, Content: "test", Collapsed: false})
	ev := tcell.NewEventKey(tcell.KeyCtrlY, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Ctrl+Y to be consumed (nil)")
	}
	if !s.Messages[0].Collapsed {
		t.Fatal("expected Message.Collapsed to be true after toggle")
	}
}

func TestShortcutToggleTheme_CtrlK_Consumed(t *testing.T) {
	a := NewApp()
	ev := tcell.NewEventKey(tcell.KeyCtrlK, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Ctrl+K to be consumed (nil)")
	}
}

func TestSendMessageBlockedDuringLoading(t *testing.T) {
	a := NewApp()
	a.newSession()
	mockClient := &MockAIClientForApp{
		mockResponse: "Hello!",
	}
	sm := session.NewManager()
	sm.SetClient(mockClient)
	a.SetSessionManager(sm)
	a.SetLoading(true)
	a.composer.SetInput("hello")

	ev := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	result := a.handleInput(ev)

	if result != nil {
		t.Fatal("expected Enter to be consumed (nil) during loading")
	}
	if s := a.activeSessionPtr(); s != nil {
		if len(s.Messages) != 0 {
			t.Fatal("expected no messages added during loading")
		}
	}
}

func TestCommandPaletteCtrlP(t *testing.T) {
	a := NewApp()
	ev := tcell.NewEventKey(tcell.KeyCtrlP, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Ctrl+P consumed (nil)")
	}
	if a.mode != ModeCommandPalette {
		t.Fatalf("expected ModeCommandPalette, got %d", a.mode)
	}
}

func TestSlashPassesThrough(t *testing.T) {
	a := NewApp()
	ev := tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone)
	result := a.handleInput(ev)
	if result == nil {
		t.Fatal("expected '/' to pass through (non-nil)")
	}
}

func TestCommandPaletteEscExits(t *testing.T) {
	a := NewApp()
	a.enterCommandPalette(ShowAll)
	ev := tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Esc consumed (nil)")
	}
	if a.mode != ModeChat {
		t.Fatalf("expected ModeChat, got %d", a.mode)
	}
}

func TestCommandPaletteBuiltinsRegistered(t *testing.T) {
	a := NewApp()
	if a.commandRegistry == nil {
		t.Fatal("expected commandRegistry to be set")
	}
	cmds := a.commandRegistry.List()
	if len(cmds) == 0 {
		t.Fatal("expected builtin commands")
	}
	found := false
	for _, c := range cmds {
		if c.Name == "New Session" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'New Session' command")
	}
}

func TestCommandPaletteAllNewCommands(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.newSession()
	a.switchToSession(0)

	names := []string{
		"Previous Session",
		"Rename Session",
		"Scroll to Top",
		"Scroll to Bottom",
		"Clear Input",
		"Reload Skills",
	}
	for _, name := range names {
		cmd := a.commandRegistry.Get(name)
		if cmd == nil {
			t.Fatalf("command %q not registered", name)
		}
		if cmd.Category != service.CmdBuiltin {
			t.Fatalf("command %q expected builtin", name)
		}
	}
}

func TestCommandScrollToTop(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.AddChatMessage("user", "hello")
	cmd := a.commandRegistry.Get("Scroll to Top")
	if cmd == nil {
		t.Fatal("Scroll to Top not registered")
	}
	cmd.Action("")
}

func TestCommandScrollToBottom(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.AddChatMessage("user", "hello")
	cmd := a.commandRegistry.Get("Scroll to Bottom")
	if cmd == nil {
		t.Fatal("Scroll to Bottom not registered")
	}
	cmd.Action("")
}

func TestCommandClearInput(t *testing.T) {
	a := NewApp()
	a.composer.SetInput("some text")
	cmd := a.commandRegistry.Get("Clear Input")
	if cmd == nil {
		t.Fatal("Clear Input not registered")
	}
	cmd.Action("")
	if a.composer.GetInput() != "" {
		t.Fatal("expected cleared input")
	}
}

func TestCommandPreviousSession(t *testing.T) {
	a := NewApp()
	a.newSession()
	a.newSession()
	a.switchToSession(1)
	cmd := a.commandRegistry.Get("Previous Session")
	if cmd == nil {
		t.Fatal("Previous Session not registered")
	}
	cmd.Action("")
	if a.activeSession != 0 {
		t.Fatalf("expected session 0, got %d", a.activeSession)
	}
}

func TestCommandPaletteHasSkillsCategory(t *testing.T) {
	a := NewApp()
	cr := service.NewCommandRegistry()
	sr := service.NewSkillRegistry()
	sr.Register(&service.Skill{Name: "test-skill", Description: "test"})
	cr.SyncSkills(sr)
	cr.Register(&service.Command{
		Name: "test-skill", Description: "test",
		Category: service.CmdSkill,
	})
	// Use the app's command palette
	a.commandRegistry = cr
	a.enterCommandPalette(ShowSkills)
	itemCount := a.commandPalette.GetItemCount()
	if itemCount < 1 {
		t.Fatal("expected at least 1 skill in palette")
	}
}

func TestSlashShowsAllCommands(t *testing.T) {
	a := NewApp()
	if a.commandRegistry == nil {
		t.Fatal("expected command registry")
	}
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
	a.onComposerChange("/Se")
	if !a.suggestionMenu.Visible() {
		t.Fatal("expected menu visible")
	}
	cmds := a.suggestionMenu.Filtered()
	if len(cmds) == 0 {
		t.Fatal("expected at least one match for /Se")
	}
	sel := a.suggestionMenu.Selected()
	if sel == nil {
		t.Fatal("expected a selected command")
	}
}

func TestSlashEnterFillsCommand(t *testing.T) {
	a := NewApp()
	a.onComposerChange("/Se")
	ev := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Enter consumed (nil)")
	}
	input := a.composer.GetInput()
	if !strings.HasPrefix(input, "/") || !strings.HasSuffix(input, " ") {
		t.Fatalf("expected composer to contain '/<cmd> ', got %q", input)
	}
	if a.suggestionMenu.Visible() {
		t.Fatal("expected menu hidden after fill")
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

func TestSlashNoMatchHides(t *testing.T) {
	a := NewApp()
	a.onComposerChange("/zzz_nonexistent")
	if a.suggestionMenu.Visible() {
		t.Fatal("expected menu hidden when no match")
	}
}

func TestSlashEscHides(t *testing.T) {
	a := NewApp()
	a.onComposerChange("/")
	if !a.suggestionMenu.Visible() {
		t.Fatal("expected menu visible")
	}
	ev := tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone)
	result := a.handleInput(ev)
	if result != nil {
		t.Fatal("expected Esc consumed (nil)")
	}
	if a.suggestionMenu.Visible() {
		t.Fatal("expected menu hidden after Esc")
	}
}

func TestSlashIncludesBuiltins(t *testing.T) {
	a := NewApp()
	a.onComposerChange("/")
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
