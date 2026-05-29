package e2e

import (
	"testing"
	"github.com/example/agent-tui/internal/ui"
)

func TestEndToEnd_ChatFlow(t *testing.T) {
	app := ui.NewApp()
	app.AddWelcomeMessage()

	app.AddChatMessage("user", "Hello")

	messages := app.GetChatMessages()
	if len(messages) < 1 {
		t.Error("Expected at least 1 message")
	}
}

func TestEndToEnd_MultipleMessages(t *testing.T) {
	app := ui.NewApp()
	app.AddWelcomeMessage()

	app.AddChatMessage("user", "First")
	app.AddChatMessage("assistant", "Second")

	messages := app.GetChatMessages()
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}
}

func TestEndToEnd_ComposerInput(t *testing.T) {
	app := ui.NewApp()

	app.SetComposerInput("Test input")
	if app.GetComposerInput() != "Test input" {
		t.Errorf("Expected 'Test input', got %s", app.GetComposerInput())
	}
}