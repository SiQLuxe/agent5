package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestSelectableTextView_RendersText(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetText("hello")
	if tv.GetText(false) != "hello" {
		t.Fatalf("expected 'hello', got %q", tv.GetText(false))
	}
}

func TestSelectableTextView_Draw(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetRect(0, 0, 20, 5)
	tv.SetText("hello world")
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(20, 5)
	tv.Draw(screen)
	screen.Show()
}

func TestSelectableTextView_GetTextStripTags(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetDynamicColors(true)
	tv.SetText(`[red]hello[-] world`)
	result := tv.GetText(true)
	if result != "hello world" {
		t.Fatalf("expected 'hello world', got %q", result)
	}
}

func TestSelectableTextView_Clear(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetText("hello")
	tv.Clear()
	if tv.GetText(false) != "" {
		t.Fatalf("expected empty, got %q", tv.GetText(false))
	}
}

func TestSelectableTextView_ColorTags(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetDynamicColors(true)
	tv.SetText(`[red]hello[-] [blue]world[-]`)
	// GetText(stripAllTags=true) should strip color tags
	result := tv.GetText(true)
	if result != "hello world" {
		t.Fatalf("expected 'hello world', got %q", result)
	}
}

func TestSelectableTextView_Highlight(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetRegions(true)
	tv.SetText(`before ["id"]highlight[""] after`)
	tv.Highlight("id")
	if tv.GetText(true) != "before highlight after" {
		t.Fatalf("unexpected stripped text: %q", tv.GetText(true))
	}
	// Draw with highlight should not panic
	tv.SetRect(0, 0, 30, 3)
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(30, 3)
	tv.Draw(screen)
	screen.Show()
}

func TestSelectableTextView_WordWrap(t *testing.T) {
	tv := NewSelectableTextView()
	tv.SetWordWrap(true)
	tv.SetText("hello world foo bar")
	lines := tv.splitLines(10)
	if len(lines) < 2 {
		t.Fatalf("expected multiple lines from word wrap, got %d", len(lines))
	}
	// Each line should not exceed wrap width
	for i, line := range lines {
		if len(line) > 10 {
			t.Fatalf("line %d exceeds wrap width: %d > 10", i, len(line))
		}
	}
}

func TestSelectableTextView_Write(t *testing.T) {
	tv := NewSelectableTextView()
	n, err := tv.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("expected 5, got %d", n)
	}
	if tv.GetText(false) != "hello" {
		t.Fatalf("expected 'hello', got %q", tv.GetText(false))
	}
}
