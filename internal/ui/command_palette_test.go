package ui

import (
	"testing"

	"github.com/example/agent-tui/internal/service"
)

func TestNewCommandPalette(t *testing.T) {
	p := NewCommandPalette()
	if p == nil {
		t.Fatal("expected non-nil")
	}
	if p.GetFilterInput() == nil {
		t.Fatal("expected filter input")
	}
}

func TestCommandPaletteFilterPrefix(t *testing.T) {
	p := NewCommandPalette()
	cmds := []*service.Command{
		{Name: "Search", Description: "Search messages", Category: service.CmdBuiltin},
		{Name: "New Session", Description: "Create a session", Category: service.CmdBuiltin},
		{Name: "help-me", Description: "Help", Category: service.CmdSkill},
	}
	p.SetCommands(cmds)
	p.SetMode(ShowAll)

	if p.GetItemCount() != 3 {
		t.Fatalf("expected 3 items, got %d", p.GetItemCount())
	}

	p.SetFilter("Se")
	if p.GetItemCount() != 1 {
		t.Fatalf("expected 1 item after filter 'Se', got %d", p.GetItemCount())
	}
	sel := p.SelectedCommand()
	if sel == nil || sel.Name != "Search" {
		t.Fatalf("expected Search, got %v", sel)
	}
}

func TestCommandPaletteFilterSkillsOnly(t *testing.T) {
	p := NewCommandPalette()
	cmds := []*service.Command{
		{Name: "Search", Category: service.CmdBuiltin},
		{Name: "code-review", Category: service.CmdSkill},
	}
	p.SetCommands(cmds)
	p.SetMode(ShowSkills)

	if p.GetItemCount() != 1 {
		t.Fatalf("expected 1 skill, got %d", p.GetItemCount())
	}
}

func TestCommandPaletteNavigation(t *testing.T) {
	p := NewCommandPalette()
	cmds := []*service.Command{
		{Name: "A", Category: service.CmdBuiltin},
		{Name: "B", Category: service.CmdBuiltin},
		{Name: "C", Category: service.CmdBuiltin},
	}
	p.SetCommands(cmds)
	p.SetMode(ShowAll)

	if p.GetItemCount() != 3 {
		t.Fatalf("expected 3")
	}

	p.SelectNext()
	sel := p.SelectedCommand()
	if sel == nil || sel.Name != "B" {
		t.Fatalf("expected B after SelectNext, got %v", sel)
	}

	p.SelectPrev()
	sel = p.SelectedCommand()
	if sel == nil || sel.Name != "A" {
		t.Fatalf("expected A after SelectPrev, got %v", sel)
	}

	// Should not go below 0
	p.SelectPrev()
	sel = p.SelectedCommand()
	if sel == nil || sel.Name != "A" {
		t.Fatalf("expected A at bottom, got %v", sel)
	}
}

func TestCommandPaletteSelectedEmpty(t *testing.T) {
	p := NewCommandPalette()
	sel := p.SelectedCommand()
	if sel != nil {
		t.Fatal("expected nil when no commands")
	}
}
