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
