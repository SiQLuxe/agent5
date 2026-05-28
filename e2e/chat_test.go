package e2e

import (
	"testing"
	"github.com/example/agent-tui/internal/ui"
)

func TestEndToEnd_ChatFlow(t *testing.T) {
	// Create UI model
	model := ui.NewModel()
	
	// Add message directly
	model.AddChatMessage("user", "Hello")
	
	// Verify message was added
	messages := model.GetChatMessages()
	if len(messages) < 1 {
		t.Error("Expected at least 1 message")
	}
}

func TestEndToEnd_MultipleMessages(t *testing.T) {
	model := ui.NewModel()
	
	model.AddChatMessage("user", "First")
	model.AddChatMessage("assistant", "Second")
	
	messages := model.GetChatMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
}

func TestEndToEnd_ComposerInput(t *testing.T) {
	model := ui.NewModel()
	
	// Test composer input
	model.SetComposerInput("Test input")
	if model.GetComposerInput() != "Test input" {
		t.Errorf("Expected 'Test input', got %s", model.GetComposerInput())
	}
}