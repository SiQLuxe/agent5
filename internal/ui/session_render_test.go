package ui

import (
	"strings"
	"testing"
	"time"
)

func TestRenderMessages_Empty(t *testing.T) {
	s := NewSession("1", "test")
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestRenderMessages_UserMessage(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleUser, "hello world")
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "hello world") {
		t.Errorf("expected content in output, got: %s", result)
	}
	if !strings.Contains(result, "\U0001f5e3") {
		t.Error("expected user badge")
	}
}

func TestRenderMessages_AssistantMessage(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleAssistant, "I am an AI")
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "I am an AI") {
		t.Errorf("expected content in output, got: %s", result)
	}
	if !strings.Contains(result, "\U0001f47e") {
		t.Error("expected assistant badge")
	}
}

func TestRenderMessages_SystemMessage(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleSystem, "system message")
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "system message") {
		t.Errorf("expected content in output, got: %s", result)
	}
	if !strings.Contains(result, "\u2699 sys") {
		t.Error("expected system badge")
	}
}

func TestRenderMessages_ThinkingCollapsed(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleAssistant, "answer")
	s.Messages[0].Thinking = &Thinking{
		Content:  "thinking content",
		Expanded: false,
		Duration: time.Second,
	}
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "\u25b6") {
		t.Error("expected collapsed thinking indicator \u25b6")
	}
	if strings.Contains(result, "thinking content") {
		t.Error("thinking content should not be visible when collapsed")
	}
}

func TestRenderMessages_ThinkingExpanded(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleAssistant, "answer")
	s.Messages[0].Thinking = &Thinking{
		Content:  "expanded thinking",
		Expanded: true,
		Duration: 2 * time.Second,
	}
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "\u25bc") {
		t.Error("expected expanded thinking indicator \u25bc")
	}
	if !strings.Contains(result, "expanded thinking") {
		t.Error("thinking content should be visible when expanded")
	}
}

func TestRenderMessages_Collapsed(t *testing.T) {
	s := NewSession("1", "test")
	content := "line1\nline2\nline3\nline4\nline5"
	s.AddMessage(RoleUser, content)
	s.Messages[0].Collapsed = true
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "\u25bc") {
		t.Error("expected collapsed indicator \u25bc")
	}
}

func TestRenderMessages_MultipleMessages(t *testing.T) {
	s := NewSession("1", "test")
	s.AddMessage(RoleUser, "user msg")
	s.AddMessage(RoleAssistant, "assistant msg")
	s.AddMessage(RoleSystem, "system msg")
	result := s.RenderMessages(80, DefaultThemes[0].Colors)
	if !strings.Contains(result, "user msg") {
		t.Error("missing user message")
	}
	if !strings.Contains(result, "assistant msg") {
		t.Error("missing assistant message")
	}
	if !strings.Contains(result, "system msg") {
		t.Error("missing system message")
	}
}
