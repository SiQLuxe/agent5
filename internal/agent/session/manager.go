package session

import (
	"sync"

	"github.com/example/agent-tui/internal/ai"
	"github.com/example/agent-tui/internal/data/history"
)

type ContextMessage struct {
	Role    string
	Content string
}

type Manager struct {
	mu      sync.RWMutex
	history *history.History
	client  ai.Client
	model   string
}

func NewManager() *Manager {
	return &Manager{
		history: history.NewHistory(""),
	}
}

func (m *Manager) SetClient(client ai.Client) {
	m.client = client
}

func (m *Manager) CreateSession(name string) string {
	return m.history.CreateSession(name)
}

func (m *Manager) ListSessions() []history.SessionInfo {
	return m.history.GetSessions()
}

func (m *Manager) AddMessage(sessionID, role, content string) error {
	return m.history.AddMessage(sessionID, role, content)
}

func (m *Manager) GetContext(sessionID string, maxTurns int) []ContextMessage {
	msgs := m.history.GetMessages(sessionID)
	if msgs == nil {
		return nil
	}
	if maxTurns > 0 && len(msgs) > maxTurns {
		msgs = msgs[len(msgs)-maxTurns:]
	}
	ctx := make([]ContextMessage, len(msgs))
	for i, msg := range msgs {
		ctx[i] = ContextMessage{Role: msg.Role, Content: msg.Content}
	}
	return ctx
}

func (m *Manager) SwitchModel(model string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.model = model
}

func (m *Manager) GetModel() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.model
}

func (m *Manager) ListModels() ([]string, error) {
	return []string{}, nil
}
