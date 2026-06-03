package backend

import (
	"context"
	"testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	mock := &mockBackend{agentType: TypeOpencode}
	r.Register(TypeOpencode, mock)
	got, ok := r.Get(TypeOpencode)
	if !ok {
		t.Fatal("expected to find backend")
	}
	if got != mock {
		t.Errorf("Get returned wrong backend")
	}
}

func TestRegistry_GetAll(t *testing.T) {
	r := NewRegistry()
	m1 := &mockBackend{agentType: TypeOpencode}
	m2 := &mockBackend{agentType: TypeClaudeCode}
	r.Register(TypeOpencode, m1)
	r.Register(TypeClaudeCode, m2)
	all := r.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 backends, got %d", len(all))
	}
}

type mockBackend struct {
	agentType AgentType
}

func (m *mockBackend) Start(ctx context.Context) error { return nil }
func (m *mockBackend) Stop(ctx context.Context) error  { return nil }
func (m *mockBackend) Health(ctx context.Context) (*HealthInfo, error) {
	return &HealthInfo{Healthy: true}, nil
}
func (m *mockBackend) CreateSession(ctx context.Context, title string) (*Session, error) {
	return &Session{ID: "mock", Title: title}, nil
}
func (m *mockBackend) ListSessions(ctx context.Context) ([]*Session, error) { return nil, nil }
func (m *mockBackend) GetSession(ctx context.Context, id string) (*Session, error) { return nil, nil }
func (m *mockBackend) DeleteSession(ctx context.Context, id string) error { return nil }
func (m *mockBackend) SendMessage(ctx context.Context, sID string, msg *Message) (*MessageResult, error) {
	return &MessageResult{Content: "mock reply"}, nil
}
func (m *mockBackend) SendMessageStream(ctx context.Context, sID string, msg *Message, onChunk func(*Chunk)) error {
	return nil
}
func (m *mockBackend) GetMessages(ctx context.Context, sID string) ([]*Message, error) { return nil, nil }
func (m *mockBackend) ExecuteCommand(ctx context.Context, sID string, cmd string) (*CommandResult, error) {
	return nil, nil
}
func (m *mockBackend) ExecuteShell(ctx context.Context, cmd string) (*CommandResult, error) { return nil, nil }
func (m *mockBackend) ReadFile(ctx context.Context, path string) (string, error) { return "", nil }
func (m *mockBackend) SearchText(ctx context.Context, pattern string) ([]SearchResult, error) { return nil, nil }
func (m *mockBackend) Events(ctx context.Context) (<-chan *Event, error) { return nil, nil }
