package tabbar

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestNew(t *testing.T) {
	tb := New()
	if tb == nil {
		t.Fatal("New() returned nil")
	}
}

func TestAddTab(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "Session 1"})
	tb.AddTab(Tab{ID: "2", Label: "Session 2"})
	if tb.TabCount() != 2 {
		t.Fatalf("expected 2 tabs, got %d", tb.TabCount())
	}
}

func TestRemoveTab(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.AddTab(Tab{ID: "2", Label: "S2"})
	tb.RemoveTab(0)
	if tb.TabCount() != 1 {
		t.Fatalf("expected 1 tab after remove, got %d", tb.TabCount())
	}
	if tb.tabs[0].ID != "2" {
		t.Fatalf("expected remaining tab ID '2', got %q", tb.tabs[0].ID)
	}
}

func TestRemoveTabOutOfRange(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.RemoveTab(5)
	if tb.TabCount() != 1 {
		t.Fatalf("expected 1 tab, got %d", tb.TabCount())
	}
	tb.RemoveTab(-1)
	if tb.TabCount() != 1 {
		t.Fatalf("expected 1 tab, got %d", tb.TabCount())
	}
}

func TestUpdateTab(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "Old"})
	tb.UpdateTab(0, "New")
	if tb.tabs[0].Label != "New" {
		t.Fatalf("expected label 'New', got %q", tb.tabs[0].Label)
	}
}

func TestUpdateTabOutOfRange(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.UpdateTab(3, "X")
	tb.UpdateTab(-1, "X")
	if tb.tabs[0].Label != "S1" {
		t.Fatalf("expected unchanged label 'S1', got %q", tb.tabs[0].Label)
	}
}

func TestSetActive(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.AddTab(Tab{ID: "2", Label: "S2"})
	tb.SetActive(1)
	if tb.ActiveIndex() != 1 {
		t.Fatalf("expected active index 1, got %d", tb.ActiveIndex())
	}
}

func TestSetActiveOutOfRange(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.SetActive(5)
	tb.SetActive(-1)
}

func TestRemoveTabAdjustsActive(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "S1"})
	tb.AddTab(Tab{ID: "2", Label: "S2"})
	tb.SetActive(0)
	tb.RemoveTab(0)
	if tb.ActiveIndex() != 0 {
		t.Fatalf("expected active index 0 after removal, got %d", tb.ActiveIndex())
	}
}

func TestSetColors(t *testing.T) {
	tb := New()
	tb.SetColors(tcell.ColorWhite, tcell.ColorBlue, tcell.ColorGray, tcell.ColorDefault)
}

func TestTabAtX(t *testing.T) {
	tb := New()
	tb.AddTab(Tab{ID: "1", Label: "A"})
	tb.AddTab(Tab{ID: "2", Label: "B"})
	idx := tb.tabAtX(0, 20)
	if idx != 0 {
		t.Fatalf("expected tab index 0 at x=0, got %d", idx)
	}
}
