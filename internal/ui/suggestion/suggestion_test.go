package suggestion

import (
	"testing"

	"github.com/example/agent-tui/internal/service"
)

func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
	if m.headerBar == nil {
		t.Fatal("expected headerBar")
	}
	if m.list == nil {
		t.Fatal("expected list")
	}
	if m.statusBar == nil {
		t.Fatal("expected statusBar")
	}
}

func TestSetCommands(t *testing.T) {
	m := New()
	cmds := []*service.Command{
		{Name: "Search", Description: "搜索", Category: service.CmdBuiltin},
		{Name: "code-review", Description: "Review code", Category: service.CmdSkill},
	}
	m.SetCommands(cmds)
	if len(m.items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(m.items))
	}
}

func TestSetFilter(t *testing.T) {
	m := New()
	cmds := []*service.Command{
		{Name: "Search", Description: "搜索", Category: service.CmdBuiltin},
		{Name: "New Session", Description: "新建会话", Category: service.CmdBuiltin},
		{Name: "Scroll to Top", Description: "滚动到顶部", Category: service.CmdBuiltin},
	}
	m.SetCommands(cmds)
	m.SetFilter("Se")
	if len(m.filtered) != 1 {
		t.Fatalf("expected 1 filtered, got %d", len(m.filtered))
	}
	if m.filtered[0].Name != "Search" {
		t.Fatalf("expected Search, got %s", m.filtered[0].Name)
	}
}

func TestSetFilterCaseInsensitive(t *testing.T) {
	m := New()
	cmds := []*service.Command{
		{Name: "Search", Description: "搜索"},
		{Name: "scroll-to-top", Description: "滚动"},
	}
	m.SetCommands(cmds)
	m.SetFilter("scroll")
	if len(m.filtered) != 1 {
		t.Fatalf("expected 1 filtered, got %d", len(m.filtered))
	}
	if m.filtered[0].Name != "scroll-to-top" {
		t.Fatalf("expected scroll-to-top, got %s", m.filtered[0].Name)
	}
}

func TestSetFilterNoMatch(t *testing.T) {
	m := New()
	m.SetCommands([]*service.Command{{Name: "Search"}})
	m.SetFilter("xyz")
	if len(m.filtered) != 0 {
		t.Fatalf("expected 0 filtered, got %d", len(m.filtered))
	}
}

func TestNavigation(t *testing.T) {
	m := New()
	cmds := []*service.Command{
		{Name: "A"}, {Name: "B"}, {Name: "C"},
	}
	m.SetCommands(cmds)
	m.SetFilter("")

	if m.selected != 0 {
		t.Fatalf("expected selected 0, got %d", m.selected)
	}
	m.SelectNext()
	if m.selected != 1 {
		t.Fatalf("expected selected 1, got %d", m.selected)
	}
	m.SelectNext()
	if m.selected != 2 {
		t.Fatalf("expected selected 2, got %d", m.selected)
	}
	m.SelectNext() // should stay at 2
	if m.selected != 2 {
		t.Fatalf("expected selected 2 (clamped), got %d", m.selected)
	}
	m.SelectPrev()
	if m.selected != 1 {
		t.Fatalf("expected selected 1, got %d", m.selected)
	}
	m.SelectPrev()
	if m.selected != 0 {
		t.Fatalf("expected selected 0, got %d", m.selected)
	}
	m.SelectPrev() // should stay at 0
	if m.selected != 0 {
		t.Fatalf("expected selected 0 (clamped), got %d", m.selected)
	}
}

func TestSelected(t *testing.T) {
	m := New()
	cmds := []*service.Command{
		{Name: "Alpha"}, {Name: "Beta"},
	}
	m.SetCommands(cmds)
	m.SetFilter("")

	sel := m.Selected()
	if sel == nil || sel.Name != "Alpha" {
		t.Fatalf("expected Alpha, got %v", sel)
	}
	m.SelectNext()
	sel = m.Selected()
	if sel == nil || sel.Name != "Beta" {
		t.Fatalf("expected Beta, got %v", sel)
	}
}

func TestSelectedEmpty(t *testing.T) {
	m := New()
	sel := m.Selected()
	if sel != nil {
		t.Fatal("expected nil when no commands")
	}
}

func TestShowHide(t *testing.T) {
	m := New()
	if m.Visible() {
		t.Fatal("expected not visible initially")
	}
	if m.Height() != 0 {
		t.Fatalf("expected height 0, got %d", m.Height())
	}
	m.SetCommands([]*service.Command{{Name: "Test"}})
	m.SetFilter("")
	m.Show()
	if !m.Visible() {
		t.Fatal("expected visible after Show()")
	}
	if m.Height() != 3 { // 1 item + header + status = 3
		t.Fatalf("expected height 3, got %d", m.Height())
	}
	m.Hide()
	if m.Visible() {
		t.Fatal("expected not visible after Hide()")
	}
	if m.Height() != 0 {
		t.Fatalf("expected height 0, got %d", m.Height())
	}
}

func TestHeightMax(t *testing.T) {
	m := New()
	cmds := make([]*service.Command, 20)
	for i := 0; i < 20; i++ {
		cmds[i] = &service.Command{Name: string(rune('A' + i))}
	}
	m.SetCommands(cmds)
	m.SetFilter("")
	m.Show()
	if m.Height() != 10 { // max 8 items + header + status = 10
		t.Fatalf("expected height 10 (capped), got %d", m.Height())
	}
}

func TestFiltered(t *testing.T) {
	m := New()
	m.SetCommands([]*service.Command{
		{Name: "Alpha", Description: "First"},
		{Name: "Beta", Description: "Second"},
	})
	m.SetFilter("A")
	if len(m.Filtered()) != 1 {
		t.Fatalf("expected 1 filtered, got %d", len(m.Filtered()))
	}
	if m.Filtered()[0].Name != "Alpha" {
		t.Fatalf("expected Alpha, got %s", m.Filtered()[0].Name)
	}
}
