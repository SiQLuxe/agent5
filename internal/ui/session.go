package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/example/agent-tui/internal/ui/syntax"
)

type Role int

const (
	RoleUser Role = iota
	RoleAssistant
	RoleSystem
)

func (r Role) String() string {
	switch r {
	case RoleUser:
		return "user"
	case RoleAssistant:
		return "assistant"
	case RoleSystem:
		return "system"
	default:
		return "unknown"
	}
}

type Thinking struct {
	Content  string
	Duration time.Duration
	Expanded bool
}

type Message struct {
	Role      Role
	Content   string
	Thinking  *Thinking
	Timestamp time.Time
	Collapsed bool // whether message content is folded
	lineCount int  // cached rendered line count (0 = uncached)
}

type Session struct {
	ID        string
	Label     string
	Messages  []Message
	CreatedAt time.Time
}

func NewSession(id, label string) *Session {
	return &Session{
		ID:        id,
		Label:     label,
		Messages:  []Message{},
		CreatedAt: time.Now(),
	}
}

func (s *Session) AddMessage(role Role, content string) {
	s.Messages = append(s.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}

func (s *Session) SetThinking(content string) {
	// Find or create thinking on last assistant message
	if len(s.Messages) > 0 && s.Messages[len(s.Messages)-1].Role == RoleAssistant {
		s.Messages[len(s.Messages)-1].Thinking = &Thinking{
			Content:  content,
			Expanded: false,
		}
	}
}

func (s *Session) FinishThinking(duration time.Duration) {
	if len(s.Messages) > 0 && s.Messages[len(s.Messages)-1].Thinking != nil {
		s.Messages[len(s.Messages)-1].Thinking.Duration = duration
	}
}

func (s *Session) ToggleThinking() {
	if len(s.Messages) > 0 && s.Messages[len(s.Messages)-1].Thinking != nil {
		s.Messages[len(s.Messages)-1].Thinking.Expanded = !s.Messages[len(s.Messages)-1].Thinking.Expanded
	}
}

// ToggleCollapse toggles the collapsed state of the last message.
func (s *Session) ToggleCollapse() {
	if len(s.Messages) > 0 {
		s.Messages[len(s.Messages)-1].Collapsed = !s.Messages[len(s.Messages)-1].Collapsed
		// Invalidate line count cache
		s.Messages[len(s.Messages)-1].lineCount = 0
	}
}

func (s *Session) GenerateLabel() string {
	if s.Label != "" && s.Label != "New Session" {
		return s.Label
	}
	if len(s.Messages) > 0 {
		first := s.Messages[0].Content
		// Check for slash command
		if len(first) > 0 && first[0] == '/' {
			return first
		}
		// Truncate to 20 chars
		runes := []rune(first)
		if len(runes) > 20 {
			return string(runes[:20]) + "..."
		}
		return first
	}
	return "New Session"
}

// RenderMessages renders all messages as a formatted string for viewport content.
func (s *Session) RenderMessages(width int, theme ColorPalette) string {
	if len(s.Messages) == 0 {
		return ""
	}
	contentWidth := width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	var sb strings.Builder
	for _, msg := range s.Messages {
		renderMessageToBuilder(&sb, msg, contentWidth, theme)
	}
	return sb.String()
}

func renderMessageToBuilder(sb *strings.Builder, msg Message, width int, theme ColorPalette) {
	ts := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Timestamp)).
		Render(msg.Timestamp.Format("15:04"))

	switch msg.Role {
	case RoleUser:
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.UserBg)).
			Foreground(lipgloss.Color(theme.UserFg)).
			Bold(true).
			Padding(0, 1).
			Render(" \U0001f5e3 ")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 3)

	case RoleAssistant:
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.AssistantBg)).
			Foreground(lipgloss.Color(theme.AssistantFg)).
			Bold(true).
			Padding(0, 1).
			Render(" \U0001f47e Chan ")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")

		if msg.Thinking != nil {
			renderThinkingBlock(sb, msg.Thinking, width, theme)
		}
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 1)

	case RoleSystem:
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.SystemBg)).
			Foreground(lipgloss.Color(theme.SystemFg)).
			Bold(true).
			Padding(0, 1).
			Render(" \u2699 sys ")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 1)
	}

	sb.WriteString("\n")
}

func renderContent(sb *strings.Builder, content string, collapsed bool, width int, theme ColorPalette, paddingLeft int) {
	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Text)).
		PaddingLeft(paddingLeft)

	if collapsed {
		sb.WriteString(foldContent(content, textStyle, theme))
		return
	}

	highlighted := safeHighlight(content)
	sb.WriteString(textStyle.Render(highlighted))
}

func safeHighlight(content string) string {
	if hasCJK(content) {
		return content
	}
	return syntax.Highlight(content, "")
}

func foldContent(content string, style lipgloss.Style, theme ColorPalette) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)
	if totalLines <= 3 {
		return style.Render(content)
	}
	preview := strings.Join(lines[:3], "\n")
	indicator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.TextMuted)).
		Render(fmt.Sprintf("\u25bc [expand %d lines]", totalLines-3))
	return style.Render(preview) + "\n" + indicator
}

func renderThinkingBlock(sb *strings.Builder, t *Thinking, width int, theme ColorPalette) {
	arrow := "\u25b6"
	if t.Expanded {
		arrow = "\u25bc"
	}

	charCount := len([]rune(t.Content))
	durationStr := ""
	if t.Duration > 0 {
		durationStr = fmt.Sprintf(" \u00b7 %.1fs", t.Duration.Seconds())
	}

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ThinkingFg)).
		Render(fmt.Sprintf("%s thinking %d chars%s", arrow, charCount, durationStr))
	sb.WriteString(header)
	sb.WriteString("\n")

	if t.Expanded {
		contentStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(theme.ThinkingBg)).
			BorderLeft(true).
			BorderStyle(lipgloss.Border{Left: "\u2502"}).
			BorderForeground(lipgloss.Color(theme.ThinkingBorder)).
			Padding(0, 1).
			MarginLeft(1).
			Width(width - 4)
		sb.WriteString(contentStyle.Render(t.Content))
		sb.WriteString("\n")
	}
}

func hasCJK(s string) bool {
	for _, r := range s {
		if (r >= 0x4E00 && r <= 0x9FFF) ||
			(r >= 0x3400 && r <= 0x4DBF) ||
			(r >= 0x20000 && r <= 0x2A6DF) ||
			(r >= 0xF900 && r <= 0xFAFF) ||
			(r >= 0x3040 && r <= 0x309F) ||
			(r >= 0x30A0 && r <= 0x30FF) ||
			(r >= 0xAC00 && r <= 0xD7AF) {
			return true
		}
	}
	return false
}


