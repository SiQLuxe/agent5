package backend

import (
	"testing"
	"time"
)

func TestAgentTypes(t *testing.T) {
	if TypeOpencode != "opencode" {
		t.Errorf("TypeOpencode = %q, want %q", TypeOpencode, "opencode")
	}
	if TypeClaudeCode != "claude-code" {
		t.Errorf("TypeClaudeCode = %q, want %q", TypeClaudeCode, "claude-code")
	}
	if TypeCodex != "codex" {
		t.Errorf("TypeCodex = %q, want %q", TypeCodex, "codex")
	}
}

func TestSessionModel(t *testing.T) {
	now := time.Now()
	s := Session{
		ID: "sess-1", Title: "test", CreatedAt: now, Status: SessionActive,
	}
	if s.ID != "sess-1" || s.Title != "test" || s.Status != SessionActive {
		t.Errorf("Session fields not set correctly")
	}
}

func TestMessageModel(t *testing.T) {
	m := Message{Role: RoleUser, Content: "hello"}
	if m.Role != RoleUser || m.Content != "hello" {
		t.Errorf("Message fields not set correctly")
	}
}

func TestEventModel(t *testing.T) {
	e := Event{Type: "message.completed", Payload: "ok"}
	if e.Type != "message.completed" {
		t.Errorf("Event type not set correctly")
	}
}
