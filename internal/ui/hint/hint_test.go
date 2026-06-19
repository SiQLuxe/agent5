package hint

import (
	"strings"
	"testing"
)

func TestNew_StartsEmpty(t *testing.T) {
	h := New()
	if got := h.GetText(true); got != "" {
		t.Fatalf("expected empty hint bar, got %q", got)
	}
}

func TestShow_WritesTextWithInfoTag(t *testing.T) {
	h := New()
	h.Show("再次按 Ctrl+C 退出", LevelInfo)

	stripped := h.GetText(true)
	if !strings.Contains(stripped, "再次按 Ctrl+C 退出") {
		t.Fatalf("stripped text missing message: %q", stripped)
	}

	raw := h.GetText(false)
	if !strings.Contains(raw, "[gray]") {
		t.Fatalf("raw text missing [gray] color tag: %q", raw)
	}
}

func TestShow_WarnUsesYellow(t *testing.T) {
	h := New()
	h.Show("careful", LevelWarn)
	if !strings.Contains(h.GetText(false), "[yellow]") {
		t.Fatalf("expected [yellow] tag for LevelWarn, got %q", h.GetText(false))
	}
}

func TestShow_ErrorUsesRed(t *testing.T) {
	h := New()
	h.Show("boom", LevelError)
	if !strings.Contains(h.GetText(false), "[red]") {
		t.Fatalf("expected [red] tag for LevelError, got %q", h.GetText(false))
	}
}

func TestClear_RemovesText(t *testing.T) {
	h := New()
	h.Show("temp", LevelInfo)
	h.Clear()
	if got := h.GetText(true); got != "" {
		t.Fatalf("expected empty after Clear, got %q", got)
	}
}

func TestShow_ReplacesPriorContent(t *testing.T) {
	h := New()
	h.Show("first", LevelInfo)
	h.Show("second", LevelWarn)

	stripped := h.GetText(true)
	if strings.Contains(stripped, "first") {
		t.Fatalf("expected second Show to replace first; got %q", stripped)
	}
	if !strings.Contains(stripped, "second") {
		t.Fatalf("expected stripped text to contain 'second'; got %q", stripped)
	}
}
