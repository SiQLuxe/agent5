package status

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
}

func TestSetMode(t *testing.T) {
	s := New()
	s.SetMode("chat")
	if s.mode != "chat" {
		t.Fatalf("expected mode 'chat', got %q", s.mode)
	}
}

func TestSetTasks(t *testing.T) {
	s := New()
	s.SetTasks(3)
	if s.tasks != 3 {
		t.Fatalf("expected tasks 3, got %d", s.tasks)
	}
}

func TestSetConnected(t *testing.T) {
	s := New()
	s.SetConnected(true)
	if !s.connected {
		t.Fatal("expected connected=true")
	}
	s.SetConnected(false)
	if s.connected {
		t.Fatal("expected connected=false after reset")
	}
}

func TestSetBackgroundColor(t *testing.T) {
	s := New()
	s.SetBackgroundColor(tcell.ColorBlue)
}

func TestRefresh(t *testing.T) {
	s := New()
	s.SetMode("chat")
	s.SetTasks(2)
	s.SetConnected(true)
	text := s.GetText(false)
	if text == "" {
		t.Fatal("expected non-empty text after refresh")
	}
}
