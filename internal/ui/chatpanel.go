package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/example/agent-tui/internal/ui/syntax"
)

type ChatPanelColors struct {
	Background     string
	UserFg         string
	AssistantBg    string
	AssistantFg    string
	SystemBg       string
	SystemFg       string
	Text           string
	TextMuted      string
	Timestamp      string
	ThinkingFg     string
	ThinkingBg     string
	ThinkingBorder string
}

func DefaultChatPanelColors() ChatPanelColors {
	return ChatPanelColors{
		Background:     "#1e1e1e",
		UserFg:         "#4ec9b0",
		AssistantBg:    "#28a745",
		AssistantFg:    "#ffffff",
		SystemBg:       "#9370db",
		SystemFg:       "#ffffff",
		Text:           "#d4d4d4",
		TextMuted:      "#858585",
		Timestamp:      "#555555",
		ThinkingFg:     "#dcdcaa",
		ThinkingBg:     "#2a2a1a",
		ThinkingBorder: "#dcdcaa",
	}
}

type ChatPanel struct {
	session *Session
	width   int
	height  int
	colors  ChatPanelColors
}

func NewChatPanel(session *Session) *ChatPanel {
	return &ChatPanel{
		session: session,
		width:   80,
		height:  24,
		colors:  DefaultChatPanelColors(),
	}
}

func (cp *ChatPanel) SetSession(s *Session) {
	cp.session = s
}

func (cp *ChatPanel) Session() *Session {
	return cp.session
}

func (cp *ChatPanel) SetColors(c ChatPanelColors) {
	cp.colors = c
}

func (cp *ChatPanel) SetSize(width, height int) {
	cp.width = width
	cp.height = height
}

func (cp *ChatPanel) View() string {
	style := lipgloss.NewStyle().
		Width(cp.width).
		Height(cp.height).
		Background(lipgloss.Color(cp.colors.Background)).
		Padding(1, 2)

	var content strings.Builder

	if cp.session == nil || len(cp.session.Messages) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(cp.colors.TextMuted)).
			Padding(2, 0)
		content.WriteString(emptyStyle.Render("开始新对话..."))
		return style.Render(content.String())
	}

	for _, msg := range cp.session.Messages {
		cp.renderMessage(&content, msg)
	}

	return style.Render(content.String())
}

func (cp *ChatPanel) renderMessage(sb *strings.Builder, msg Message) {
	// Timestamp
	ts := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cp.colors.Timestamp)).
		Render(msg.Timestamp.Format("15:04"))

	switch msg.Role {
	case RoleUser:
		// 👤 User badge with background
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(cp.colors.UserFg)).
			Foreground(lipgloss.Color("#1e1e1e")).
			Bold(true).
			Padding(0, 1).
			Render("👤")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")

		textStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(cp.colors.Text)).
			PaddingLeft(3)
		highlighted := syntax.Highlight(msg.Content, "")
		sb.WriteString(textStyle.Render(highlighted))
		sb.WriteString("\n\n")

	case RoleAssistant:
		// 🤖 Luxe badge
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(cp.colors.AssistantBg)).
			Foreground(lipgloss.Color(cp.colors.AssistantFg)).
			Bold(true).
			Padding(0, 1).
			Render("🤖 Luxe")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")

		// Thinking display
		if msg.Thinking != nil {
			cp.renderThinking(sb, msg.Thinking)
		}

		textStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(cp.colors.Text)).
			PaddingLeft(1)
		highlighted := syntax.Highlight(msg.Content, "")
		sb.WriteString(textStyle.Render(highlighted))
		sb.WriteString("\n\n")

	case RoleSystem:
		// ⚙️ System badge
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(cp.colors.SystemBg)).
			Foreground(lipgloss.Color(cp.colors.SystemFg)).
			Bold(true).
			Padding(0, 1).
			Render("⚙️ System")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")

		textStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(cp.colors.Text)).
			PaddingLeft(1)
		sb.WriteString(textStyle.Render(msg.Content))
		sb.WriteString("\n\n")
	}
}

func (cp *ChatPanel) renderThinking(sb *strings.Builder, t *Thinking) {
	arrow := "▶"
	if t.Expanded {
		arrow = "▼"
	}

	charCount := len([]rune(t.Content))
	durationStr := ""
	if t.Duration > 0 {
		durationStr = fmt.Sprintf(" · %.1fs", t.Duration.Seconds())
	}

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cp.colors.ThinkingFg)).
		Render(fmt.Sprintf("%s thinking %d chars%s", arrow, charCount, durationStr))
	sb.WriteString(header)
	sb.WriteString("\n")

	if t.Expanded {
		// Content area with yellow left border
		contentStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(cp.colors.ThinkingBg)).
			BorderLeft(true).
			BorderStyle(lipgloss.Border{Left: "│"}).
			BorderForeground(lipgloss.Color(cp.colors.ThinkingBorder)).
			Padding(0, 1).
			MarginLeft(1).
			Width(cp.width - 8)

		sb.WriteString(contentStyle.Render(t.Content))
		sb.WriteString("\n")
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
