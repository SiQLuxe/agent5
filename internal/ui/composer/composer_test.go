package composer

import (
	"testing"
)

func TestComposer_NewComposer(t *testing.T) {
	c := NewComposer()
	if c == nil {
		t.Error("NewComposer returned nil")
	}
	if c.GetInput() != "" {
		t.Errorf("expected empty input, got '%s'", c.GetInput())
	}
}

func TestComposer_AppendInput(t *testing.T) {
	c := NewComposer()
	c.AppendInput("hello")
	if c.GetInput() != "hello" {
		t.Errorf("expected 'hello', got '%s'", c.GetInput())
	}
}

func TestComposer_AppendChinese(t *testing.T) {
	c := NewComposer()
	c.AppendInput("你好")
	if c.GetInput() != "你好" {
		t.Errorf("expected '你好', got '%s'", c.GetInput())
	}
}

func TestComposer_ClearInput(t *testing.T) {
	c := NewComposer()
	c.AppendInput("hello")
	c.ClearInput()
	if c.GetInput() != "" {
		t.Errorf("expected empty input after ClearInput, got '%s'", c.GetInput())
	}
}

func TestComposer_Backspace(t *testing.T) {
	c := NewComposer()
	c.AppendInput("hello")
	c.Backspace()
	if c.GetInput() != "hell" {
		t.Errorf("expected 'hell', got '%s'", c.GetInput())
	}
}

func TestComposer_BackspaceChinese(t *testing.T) {
	c := NewComposer()
	c.AppendInput("你好")
	c.Backspace()
	if c.GetInput() != "你" {
		t.Errorf("expected '你', got '%s'", c.GetInput())
	}
}

func TestComposer_BackspaceEmoji(t *testing.T) {
	c := NewComposer()
	c.AppendInput("👋")
	c.Backspace()
	if c.GetInput() != "" {
		t.Errorf("expected empty after backspace emoji, got '%s'", c.GetInput())
	}
}

func TestComposer_BackspaceEmpty(t *testing.T) {
	c := NewComposer()
	c.Backspace()
	if c.GetInput() != "" {
		t.Errorf("expected empty after backspace on empty, got '%s'", c.GetInput())
	}
}

func TestComposer_ViewEmpty(t *testing.T) {
	c := NewComposer()
	view := c.View()
	if view == "" {
		t.Error("View should not return empty string for empty composer")
	}
}

func TestComposer_ViewWithInput(t *testing.T) {
	c := NewComposer()
	c.AppendInput("hello")
	view := c.View()
	if view == "" {
		t.Error("View should not return empty string with input")
	}
}

func TestComposer_SetColors(t *testing.T) {
	c := NewComposer()
	colors := ComposerColors{
		Background: "#000",
		Prompt:     "#fff",
		Separator:  "#333",
		Text:       "#aaa",
	}
	c.SetColors(colors)
	if c.colors.Background != "#000" {
		t.Errorf("expected '#000', got '%s'", c.colors.Background)
	}
}
