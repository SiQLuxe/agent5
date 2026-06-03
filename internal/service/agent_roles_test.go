package service

import (
	"context"
	"testing"

	"github.com/example/agent-tui/internal/backend"
)

type mockAgentBackend struct {
	reply string
}

func (m *mockAgentBackend) Start(ctx context.Context) error { return nil }
func (m *mockAgentBackend) Stop(ctx context.Context) error  { return nil }
func (m *mockAgentBackend) Health(ctx context.Context) (*backend.HealthInfo, error) {
	return &backend.HealthInfo{Healthy: true}, nil
}
func (m *mockAgentBackend) CreateSession(ctx context.Context, title string) (*backend.Session, error) {
	return &backend.Session{ID: "mock", Title: title}, nil
}
func (m *mockAgentBackend) ListSessions(ctx context.Context) ([]*backend.Session, error) { return nil, nil }
func (m *mockAgentBackend) GetSession(ctx context.Context, id string) (*backend.Session, error) {
	return nil, nil
}
func (m *mockAgentBackend) DeleteSession(ctx context.Context, id string) error { return nil }
func (m *mockAgentBackend) SendMessage(ctx context.Context, sID string, msg *backend.Message) (*backend.MessageResult, error) {
	return &backend.MessageResult{Content: m.reply}, nil
}
func (m *mockAgentBackend) SendMessageStream(ctx context.Context, sID string, msg *backend.Message, onChunk func(*backend.Chunk)) error {
	return nil
}
func (m *mockAgentBackend) GetMessages(ctx context.Context, sID string) ([]*backend.Message, error) {
	return nil, nil
}
func (m *mockAgentBackend) ExecuteCommand(ctx context.Context, sID string, cmd string) (*backend.CommandResult, error) {
	return nil, nil
}
func (m *mockAgentBackend) ExecuteShell(ctx context.Context, cmd string) (*backend.CommandResult, error) {
	return nil, nil
}
func (m *mockAgentBackend) ReadFile(ctx context.Context, path string) (string, error) { return "", nil }
func (m *mockAgentBackend) SearchText(ctx context.Context, pattern string) ([]backend.SearchResult, error) {
	return nil, nil
}
func (m *mockAgentBackend) Events(ctx context.Context) (<-chan *backend.Event, error) { return nil, nil }

func TestExternalAgent_Execute(t *testing.T) {
	b := &mockAgentBackend{reply: "external reply"}
	agent := NewExternalAgent(b)

	task := &Task{
		ID:      "ext-1",
		Type:    TaskExternal,
		Content: "hello from test",
	}

	result, err := agent.Execute(task)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result != "external reply" {
		t.Errorf("result = %q, want %q", result, "external reply")
	}
}

func TestExternalAgent_RoleName(t *testing.T) {
	agent := NewExternalAgent(&mockAgentBackend{})
	if name := agent.GetRoleName(); name != "外部Agent" {
		t.Errorf("role name = %q, want %q", name, "外部Agent")
	}
}

func TestExternalAgent_SupportedTaskTypes(t *testing.T) {
	agent := NewExternalAgent(&mockAgentBackend{})
	types := agent.GetSupportedTaskTypes()
	if len(types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(types))
	}
	if types[0] != TaskExternal {
		t.Errorf("type = %q, want %q", types[0], TaskExternal)
	}
}
