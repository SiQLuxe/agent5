package backend

import (
	"context"
	"time"
)

type AgentType string

const (
	TypeOpencode   AgentType = "opencode"
	TypeClaudeCode AgentType = "claude-code"
	TypeCodex      AgentType = "codex"
)

type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MessageResult struct {
	SessionID string `json:"session_id"`
	MessageID string `json:"message_id"`
	Content   string `json:"content"`
}

type Chunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

type CommandResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

type SearchResult struct {
	Path       string `json:"path"`
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
}

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type HealthInfo struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version,omitempty"`
}

type BackendConfig struct {
	Type      AgentType `json:"type"`
	Enabled   bool      `json:"enabled"`
	AutoStart bool      `json:"auto_start"`
	Binary    string    `json:"binary"`
	APIURL    string    `json:"api_url,omitempty"`
	APIKey    string    `json:"api_key,omitempty"`
}

type AgentBackend interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health(ctx context.Context) (*HealthInfo, error)

	CreateSession(ctx context.Context, title string) (*Session, error)
	ListSessions(ctx context.Context) ([]*Session, error)
	GetSession(ctx context.Context, id string) (*Session, error)
	DeleteSession(ctx context.Context, id string) error

	SendMessage(ctx context.Context, sessionID string, msg *Message) (*MessageResult, error)
	SendMessageStream(ctx context.Context, sessionID string, msg *Message, onChunk func(*Chunk)) error
	GetMessages(ctx context.Context, sessionID string) ([]*Message, error)

	ExecuteCommand(ctx context.Context, sessionID string, command string) (*CommandResult, error)
	ExecuteShell(ctx context.Context, command string) (*CommandResult, error)

	ReadFile(ctx context.Context, path string) (string, error)
	SearchText(ctx context.Context, pattern string) ([]SearchResult, error)

	Events(ctx context.Context) (<-chan *Event, error)
}
