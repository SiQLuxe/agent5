package ui

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	s := NewSession("s1", "Test")
	if s.ID != "s1" {
		t.Errorf("expected ID 's1', got '%s'", s.ID)
	}
	if s.Label != "Test" {
		t.Errorf("expected Label 'Test', got '%s'", s.Label)
	}
	if len(s.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(s.Messages))
	}
}

func TestSession_AddMessage(t *testing.T) {
	s := NewSession("s1", "Test")
	s.AddMessage(RoleUser, "Hello")
	if len(s.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(s.Messages))
	}
	if s.Messages[0].Role != RoleUser {
		t.Errorf("expected RoleUser, got %v", s.Messages[0].Role)
	}
	if s.Messages[0].Content != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", s.Messages[0].Content)
	}
}

func TestSession_SetThinking(t *testing.T) {
	s := NewSession("s1", "Test")
	s.AddMessage(RoleAssistant, "")
	s.SetThinking("analyzing code...")
	if s.Messages[0].Thinking == nil {
		t.Fatal("expected Thinking to be set")
	}
	if s.Messages[0].Thinking.Content != "analyzing code..." {
		t.Errorf("expected 'analyzing code...', got '%s'", s.Messages[0].Thinking.Content)
	}
	if s.Messages[0].Thinking.Expanded != false {
		t.Error("expected Expanded to be false")
	}
}

func TestSession_ToggleThinking(t *testing.T) {
	s := NewSession("s1", "Test")
	s.AddMessage(RoleAssistant, "")
	s.SetThinking("thinking...")
	s.ToggleThinking()
	if s.Messages[0].Thinking.Expanded != true {
		t.Error("expected Expanded to be true after toggle")
	}
	s.ToggleThinking()
	if s.Messages[0].Thinking.Expanded != false {
		t.Error("expected Expanded to be false after second toggle")
	}
}

func TestSession_FinishThinking(t *testing.T) {
	s := NewSession("s1", "Test")
	s.AddMessage(RoleAssistant, "")
	s.SetThinking("thinking...")
	s.FinishThinking(3 * time.Second)
	if s.Messages[0].Thinking.Duration != 3*time.Second {
		t.Errorf("expected 3s, got %v", s.Messages[0].Thinking.Duration)
	}
}

func TestSession_GenerateLabel_SlashCommand(t *testing.T) {
	s := NewSession("s1", "New Session")
	s.AddMessage(RoleUser, "/fix lint errors")
	label := s.GenerateLabel()
	if label != "/fix lint errors" {
		t.Errorf("expected '/fix lint errors', got '%s'", label)
	}
}

func TestSession_GenerateLabel_Truncate(t *testing.T) {
	s := NewSession("s1", "New Session")
	s.AddMessage(RoleUser, "这是一个很长的消息用于测试截断功能")
	label := s.GenerateLabel()
	expected := "这是一个很长的消息用于测试截断功能"
	if label != expected {
		t.Errorf("expected '%s', got '%s'", expected, label)
	}
}

func TestSession_GenerateLabel_Default(t *testing.T) {
	s := NewSession("s1", "New Session")
	label := s.GenerateLabel()
	if label != "New Session" {
		t.Errorf("expected 'New Session', got '%s'", label)
	}
}

func TestRole_String(t *testing.T) {
	if RoleUser.String() != "user" {
		t.Errorf("expected 'user', got '%s'", RoleUser.String())
	}
	if RoleAssistant.String() != "assistant" {
		t.Errorf("expected 'assistant', got '%s'", RoleAssistant.String())
	}
	if RoleSystem.String() != "system" {
		t.Errorf("expected 'system', got '%s'", RoleSystem.String())
	}
}
