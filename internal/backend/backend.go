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

type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
)

type SessionStatus string

const (
	SessionActive   SessionStatus = "active"
	SessionArchived SessionStatus = "archived"
)

// Session represents a conversation session with an external agent.
type Session struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	CreatedAt time.Time     `json:"created_at"`
	Status    SessionStatus `json:"status"`
}

// Message represents a single message in a session.
type Message struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`
}

// MessageResult contains the response from sending a message to an agent.
type MessageResult struct {
	SessionID string `json:"session_id"`
	MessageID string `json:"message_id"`
	Content   string `json:"content"`
}

// Chunk represents a partial streaming response from an agent.
type Chunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

// CommandResult contains the output of a command execution.
type CommandResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// SearchResult represents a single match from a text search.
type SearchResult struct {
	Path       string `json:"path"`
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
}

// Event represents an event from the agent's event stream.
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// HealthInfo contains the health status of an agent backend.
type HealthInfo struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version,omitempty"`
}

// BackendConfig holds configuration for creating an agent backend.
type BackendConfig struct {
	Type      AgentType `json:"type"`
	Enabled   bool      `json:"enabled"`
	AutoStart bool      `json:"auto_start"`
	Binary    string    `json:"binary"`
	APIURL    string    `json:"api_url,omitempty"`
	APIKey    string    `json:"api_key,omitempty"`
}

// AgentBackend defines the interface for interacting with external AI coding agents.
// Implementations wrap specific agents (opencode, Claude Code, Codex, etc.)
// and handle the transport layer (HTTP, stdio, etc.) internally.
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
