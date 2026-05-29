package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/example/agent-tui/internal/ai"
	"github.com/example/agent-tui/internal/data/history"
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
	if a.keyMap.Quit == 0 {
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
	h := history.NewHistory("")
	mockClient := &MockAIClientForApp{
		mockResponse: "Hello!",
	}
	aiAssistant := service.NewAIAssistant(mockClient, h)
	a.SetAIAssistant(aiAssistant)

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
		if msgs[1].Role != RoleAssistant || msgs[1].Content != "Hello!" {
			t.Fatalf("expected assistant response 'Hello!', got '%s'", msgs[1].Content)
		}
	}
	if a.IsLoading() {
		t.Fatal("expected loading cleared after stream completes")
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

func TestSendMessageStreamsContent(t *testing.T) {
	a := NewApp()
	a.newSession()
	h := history.NewHistory("")
	mockClient := &MockAIClientForApp{
		mockResponse: "Hello world",
	}
	aiAssistant := service.NewAIAssistant(mockClient, h)
	a.SetAIAssistant(aiAssistant)

	a.composer.SetInput("hi")
	a.sendMessage()

	s := a.activeSessionPtr()

	if len(s.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(s.Messages))
	}
	if s.Messages[1].Content != "Hello world" {
		t.Fatalf("expected streaming content 'Hello world', got '%s'", s.Messages[1].Content)
	}
	if a.IsLoading() {
		t.Fatal("expected loading cleared after stream completes")
	}
}

func TestSendMessageBlockedDuringLoading(t *testing.T) {
	a := NewApp()
	a.newSession()
	h := history.NewHistory("")
	mockClient := &MockAIClientForApp{
		mockResponse: "Hello!",
	}
	aiAssistant := service.NewAIAssistant(mockClient, h)
	a.SetAIAssistant(aiAssistant)
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
