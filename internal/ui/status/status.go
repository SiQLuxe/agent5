package status

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

type StatusBar struct {
	mode       string
	status     string
	taskCount  int
	connection bool
}

func NewStatusBar() *StatusBar {
	return &StatusBar{
		mode:       "Chat",
		status:     "Ready",
		taskCount:  0,
		connection: true,
	}
}

func (sb *StatusBar) SetMode(mode string) {
	sb.mode = mode
}

func (sb *StatusBar) SetStatus(status string) {
	sb.status = status
}

func (sb *StatusBar) SetTaskCount(count int) {
	sb.taskCount = count
}

func (sb *StatusBar) SetConnected(connected bool) {
	sb.connection = connected
}

func (sb *StatusBar) View(width int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(1).
		Background(lipgloss.Color("#252526")).
		Padding(0, 2)

	left := lipgloss.NewStyle().Foreground(lipgloss.Color("#569cd6")).Render("Agent TUI")
	left += lipgloss.NewStyle().Foreground(lipgloss.Color("#858585")).Render(" | " + sb.mode)

	connectionStatus := "Connected"
	if !sb.connection {
		connectionStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("#f14c4c")).Render("Disconnected")
	}

	right := lipgloss.NewStyle().Foreground(lipgloss.Color("#858585")).Render(fmt.Sprintf(
		"Tasks: %d | Status: %s | %s",
		sb.taskCount,
		sb.status,
		connectionStatus,
	))

	return style.Render(
		lipgloss.JoinHorizontal(lipgloss.Left, left, right),
	)
}
