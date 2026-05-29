package tabbar

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

type Tab struct {
	ID    string
	Label string
}

type TabDockColors struct {
	ActiveBg   string
	InactiveFg string
	Separator  string
	NewButton  string
	Background string
}

func DefaultColors() TabDockColors {
	return TabDockColors{
		ActiveBg:   "#569cd6",
		InactiveFg: "#858585",
		Separator:  "#444444",
		NewButton:  "#4ec9b0",
		Background: "#1a1a1a",
	}
}

type TabDock struct {
	tabs      []Tab
	activeTab int
	width     int
	colors    TabDockColors
}

func NewTabDock(tabs []Tab) *TabDock {
	return &TabDock{
		tabs:      tabs,
		activeTab: 0,
		width:     80,
		colors:    DefaultColors(),
	}
}

func (td *TabDock) SetWidth(width int) {
	td.width = width
}

func (td *TabDock) SetColors(c TabDockColors) {
	td.colors = c
}

func (td *TabDock) SetActiveTab(index int) {
	if index >= 0 && index < len(td.tabs) {
		td.activeTab = index
	}
}

func (td *TabDock) GetActiveTab() int {
	return td.activeTab
}

func (td *TabDock) ActiveTabID() string {
	if td.activeTab >= 0 && td.activeTab < len(td.tabs) {
		return td.tabs[td.activeTab].ID
	}
	return ""
}

func (td *TabDock) AddTab(tab Tab) {
	td.tabs = append(td.tabs, tab)
}

func (td *TabDock) RemoveTab(index int) {
	if index < 0 || index >= len(td.tabs) {
		return
	}
	td.tabs = append(td.tabs[:index], td.tabs[index+1:]...)
	if td.activeTab >= len(td.tabs) {
		td.activeTab = len(td.tabs) - 1
	}
	if td.activeTab < 0 {
		td.activeTab = 0
	}
}

func (td *TabDock) UpdateTabLabel(index int, label string) {
	if index >= 0 && index < len(td.tabs) {
		td.tabs[index].Label = label
	}
}

func (td *TabDock) TabCount() int {
	return len(td.tabs)
}

func (td *TabDock) View() string {
	if len(td.tabs) == 0 {
		return ""
	}

	var parts []string

	for i, tab := range td.tabs {
		if i > 0 {
			sep := lipgloss.NewStyle().
				Foreground(lipgloss.Color(td.colors.Separator)).
				Render(" | ")
			parts = append(parts, sep)
		}

		var tabStr string
		if i == td.activeTab {
			tabStr = lipgloss.NewStyle().
				Background(lipgloss.Color(td.colors.ActiveBg)).
				Foreground(lipgloss.Color("#ffffff")).
				Bold(true).
				Padding(0, 2).
				Render(tab.Label)
		} else {
			tabStr = lipgloss.NewStyle().
				Foreground(lipgloss.Color(td.colors.InactiveFg)).
				Padding(0, 2).
				Render(tab.Label)
		}
		parts = append(parts, tabStr)
	}

	// + new button
	plusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(td.colors.NewButton)).
		Bold(true).
		Padding(0, 1)

	parts = append(parts, lipgloss.NewStyle().
		Foreground(lipgloss.Color(td.colors.Separator)).
		Render(" | "))
	parts = append(parts, plusStyle.Render("+"))

	// Help hint on the right side
	helpHint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6a6a6a")).
		Render("⌃G 快捷键")

	content := strings.Join(parts, "")

	// Left side: tabs + new button, Right side: help hint
	left := lipgloss.NewStyle().
		Background(lipgloss.Color(td.colors.Background)).
		Render(content)

	right := lipgloss.NewStyle().
		Background(lipgloss.Color(td.colors.Background)).
		Render(helpHint)

	return lipgloss.NewStyle().
		Width(td.width).
		Background(lipgloss.Color(td.colors.Background)).
		Padding(0, 1).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, left, right))
}

func (td *TabDock) HandleClick(x int) (string, bool) {
	offset := 0
	for i, tab := range td.tabs {
		tabWidth := lipgloss.Width(tab.Label) + 4
		if i > 0 {
			tabWidth += 3
		}
		if x >= offset && x < offset+tabWidth {
			td.activeTab = i
			return tab.ID, true
		}
		offset += tabWidth
	}
	return "", false
}

func (td *TabDock) GetTabID(index int) (string, error) {
	if index < 0 || index >= len(td.tabs) {
		return "", fmt.Errorf("invalid tab index: %d", index)
	}
	return td.tabs[index].ID, nil
}
