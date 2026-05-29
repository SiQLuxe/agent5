package ui

import (
	"fmt"
	"strings"
	"time"

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
	ts := fmt.Sprintf("[%s]%s[-]", theme.Timestamp, msg.Timestamp.Format("15:04"))

	switch msg.Role {
	case RoleUser:
		badge := fmt.Sprintf("[%s:%s:b] \U0001f5e3 [-:-:-]", theme.UserFg, theme.UserBg)
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 3)

	case RoleAssistant:
		badge := fmt.Sprintf("[%s:%s:b] \U0001f47e Chan [-:-:-]", theme.AssistantFg, theme.AssistantBg)
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")

		if msg.Thinking != nil {
			renderThinkingBlock(sb, msg.Thinking, width, theme)
		}
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 1)

	case RoleSystem:
		badge := fmt.Sprintf("[%s:%s:b] \u2699 sys [-:-:-]", theme.SystemFg, theme.SystemBg)
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")
		renderContent(sb, msg.Content, msg.Collapsed, width, theme, 1)
	}

	sb.WriteString("\n")
}

func renderContent(sb *strings.Builder, content string, collapsed bool, width int, theme ColorPalette, paddingLeft int) {
	padding := strings.Repeat(" ", paddingLeft)

	if collapsed {
		sb.WriteString(foldContent(content, theme.Text, theme.TextMuted, paddingLeft))
		return
	}

	highlighted := safeHighlight(content)
	sb.WriteString(fmt.Sprintf("[%s]%s%s[-]", theme.Text, padding, highlighted))
}

func safeHighlight(content string) string {
	if hasCJK(content) {
		return content
	}
	return syntax.Highlight(content, "")
}

func foldContent(content string, textColor, mutedColor string, paddingLeft int) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)
	padding := strings.Repeat(" ", paddingLeft)
	if totalLines <= 3 {
		return fmt.Sprintf("[%s]%s%s[-]", textColor, padding, content)
	}
	preview := strings.Join(lines[:3], "\n")
	indicator := fmt.Sprintf("[%s]%s\u25bc [expand %d lines][-]", mutedColor, padding, totalLines-3)
	return fmt.Sprintf("[%s]%s%s[-]\n%s", textColor, padding, preview, indicator)
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

	sb.WriteString(fmt.Sprintf("[%s]%s thinking %d chars%s[-]", theme.ThinkingFg, arrow, charCount, durationStr))
	sb.WriteString("\n")

	if t.Expanded {
		lines := strings.Split(t.Content, "\n")
		for _, line := range lines {
			sb.WriteString(fmt.Sprintf(" [%s]\u2502[%s:%s] %s[-:-]\n", theme.ThinkingBorder, theme.Text, theme.ThinkingBg, line))
		}
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
