package history

import "testing"

func TestCreateSession(t *testing.T) {
	h := NewHistory(":memory:")
	id := h.CreateSession("test-session")
	if id == "" {
		t.Error("session ID should not be empty")
	}
}

func TestAddMessage(t *testing.T) {
	h := NewHistory(":memory:")
	id := h.CreateSession("test")
	err := h.AddMessage(id, "user", "hello")
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}
	msgs := h.GetMessages(id)
	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs))
	}
}

func TestGetSessions(t *testing.T) {
	h := NewHistory(":memory:")
	h.CreateSession("session1")
	h.CreateSession("session2")
	
	sessions := h.GetSessions()
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}