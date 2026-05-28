package tabbar

import "testing"

func TestTabDock_View(t *testing.T) {
	tabs := []Tab{
		{ID: "chat", Label: "Chat"},
		{ID: "editor", Label: "Editor"},
	}
	td := NewTabDock(tabs)
	result := td.View()

	if len(result) == 0 {
		t.Error("Expected non-empty view")
	}
}

func TestTabDock_SetActiveTab(t *testing.T) {
	tabs := []Tab{
		{ID: "chat", Label: "Chat"},
		{ID: "editor", Label: "Editor"},
	}
	td := NewTabDock(tabs)
	td.SetActiveTab(1)

	if td.GetActiveTab() != 1 {
		t.Error("Expected active tab to be 1, got", td.GetActiveTab())
	}
}

func TestTabDock_GetTabID(t *testing.T) {
	tabs := []Tab{
		{ID: "chat", Label: "Chat"},
		{ID: "editor", Label: "Editor"},
	}
	td := NewTabDock(tabs)
	id, err := td.GetTabID(0)

	if err != nil {
		t.Error("Expected no error, got", err)
	}
	if id != "chat" {
		t.Error("Expected chat, got", id)
	}
}

func TestTabDock_AddTab(t *testing.T) {
	td := NewTabDock([]Tab{})
	td.AddTab(Tab{ID: "s1", Label: "Session 1"})
	if td.TabCount() != 1 {
		t.Errorf("Expected 1 tab, got %d", td.TabCount())
	}
}

func TestTabDock_RemoveTab(t *testing.T) {
	td := NewTabDock([]Tab{
		{ID: "s1", Label: "S1"},
		{ID: "s2", Label: "S2"},
	})
	td.RemoveTab(0)
	if td.TabCount() != 1 {
		t.Errorf("Expected 1 tab, got %d", td.TabCount())
	}
	if td.ActiveTabID() != "s2" {
		t.Errorf("Expected active 's2', got '%s'", td.ActiveTabID())
	}
}

func TestTabDock_UpdateTabLabel(t *testing.T) {
	td := NewTabDock([]Tab{
		{ID: "s1", Label: "Old"},
	})
	td.UpdateTabLabel(0, "New")
	if td.tabs[0].Label != "New" {
		t.Errorf("Expected 'New', got '%s'", td.tabs[0].Label)
	}
}

func TestTabDock_ActiveTabID(t *testing.T) {
	td := NewTabDock([]Tab{
		{ID: "s1", Label: "S1"},
	})
	if td.ActiveTabID() != "s1" {
		t.Errorf("Expected 's1', got '%s'", td.ActiveTabID())
	}
}
