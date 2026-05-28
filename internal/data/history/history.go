package history

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        string
	SessionID string
	Role      string
	Content   string
	Timestamp time.Time
}

type SessionInfo struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type History struct {
	mu       sync.RWMutex
	sessions map[string]SessionInfo
	messages map[string][]Message
}

func NewHistory(path string) *History {
	return &History{
		sessions: make(map[string]SessionInfo),
		messages: make(map[string][]Message),
	}
}

func (h *History) CreateSession(name string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := uuid.New().String()
	h.sessions[id] = SessionInfo{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
	}
	h.messages[id] = []Message{}
	return id
}

func (h *History) AddMessage(sessionID, role, content string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.sessions[sessionID]; !ok {
		return nil
	}

	h.messages[sessionID] = append(h.messages[sessionID], Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	return nil
}

func (h *History) GetMessages(sessionID string) []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.messages[sessionID]
}

func (h *History) GetSessions() []SessionInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	sessions := make([]SessionInfo, 0, len(h.sessions))
	for _, s := range h.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}