package ui

import "time"

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
