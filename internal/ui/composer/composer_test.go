package composer

import (
	"testing"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
}

func TestSetAndGetInput(t *testing.T) {
	c := New()
	c.SetInput("hello")
	if got := c.GetInput(); got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}

func TestClearInput(t *testing.T) {
	c := New()
	c.SetInput("hello")
	c.ClearInput()
	if got := c.GetInput(); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestClearInputEmpty(t *testing.T) {
	c := New()
	c.ClearInput()
	if got := c.GetInput(); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestPromptDisplayed(t *testing.T) {
	c := New()
	if c.prompt == nil {
		t.Fatal("expected non-nil prompt")
	}
	if c.prompt.GetText(false) != "> " {
		t.Fatalf("expected prompt '> ', got %q", c.prompt.GetText(false))
	}
}
