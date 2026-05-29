package composer

import (
	"strings"
	"testing"
)

func TestNewComposer(t *testing.T) {
	c := NewComposer()
	if c == nil {
		t.Fatal("expected non-nil Composer")
	}
}

func TestComposerSetInput(t *testing.T) {
	c := NewComposer()
	c.SetInput("hello")
	if c.GetInput() != "hello" {
		t.Errorf("expected 'hello', got '%s'", c.GetInput())
	}
}

func TestComposerClearInput(t *testing.T) {
	c := NewComposer()
	c.SetInput("hello")
	c.ClearInput()
	if c.GetInput() != "" {
		t.Errorf("expected empty, got '%s'", c.GetInput())
	}
}

func TestComposerAppendInput(t *testing.T) {
	c := NewComposer()
	c.SetInput("hel")
	c.AppendInput("lo")
	if c.GetInput() != "hello" {
		t.Errorf("expected 'hello', got '%s'", c.GetInput())
	}
}

func TestComposerBackspace(t *testing.T) {
	c := NewComposer()
	c.SetInput("hello")
	c.Backspace()
	if c.GetInput() != "hell" {
		t.Errorf("expected 'hell', got '%s'", c.GetInput())
	}
}

func TestComposerView(t *testing.T) {
	c := NewComposer()
	c.SetWidth(80)
	c.SetInput("test")
	result := c.View()
	if !strings.Contains(result, "test") {
		t.Errorf("expected view to contain 'test', got: %s", result)
	}
}
