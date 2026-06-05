package session

import (
	"testing"
)

func TestGetContextReturnsAllMessages(t *testing.T) {
	sm := NewManager()
	id := sm.CreateSession("test")
	sm.AddMessage(id, "user", "hello")
	sm.AddMessage(id, "assistant", "hi")
	sm.AddMessage(id, "user", "how are you?")

	ctx := sm.GetContext(id, 0)
	if len(ctx) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(ctx))
	}
	if ctx[0].Role != "user" || ctx[0].Content != "hello" {
		t.Fatalf("unexpected first message: %+v", ctx[0])
	}
}

func TestGetContextTruncatesOldest(t *testing.T) {
	sm := NewManager()
	id := sm.CreateSession("test")
	for i := 0; i < 10; i++ {
		sm.AddMessage(id, "user", "msg")
		sm.AddMessage(id, "assistant", "resp")
	}

	ctx := sm.GetContext(id, 4)
	if len(ctx) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(ctx))
	}
}

func TestGetContextReturnsEmptyForUnknownSession(t *testing.T) {
	sm := NewManager()
	ctx := sm.GetContext("nonexistent", 0)
	if len(ctx) != 0 {
		t.Fatal("expected empty context for unknown session")
	}
}
