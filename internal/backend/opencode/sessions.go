package opencode

import (
	"context"

	"github.com/example/agent-tui/internal/backend"
)

type opencodeSession struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (b *OpencodeBackend) CreateSession(ctx context.Context, title string) (*backend.Session, error) {
	body := map[string]string{"title": title}
	var os opencodeSession
	if err := b.doRequest(ctx, "POST", "/session", body, &os); err != nil {
		return nil, err
	}
	return mapSession(&os), nil
}

func (b *OpencodeBackend) ListSessions(ctx context.Context) ([]*backend.Session, error) {
	var sessions []opencodeSession
	if err := b.doRequest(ctx, "GET", "/session", nil, &sessions); err != nil {
		return nil, err
	}
	result := make([]*backend.Session, len(sessions))
	for i, s := range sessions {
		result[i] = mapSession(&s)
	}
	return result, nil
}

func (b *OpencodeBackend) GetSession(ctx context.Context, id string) (*backend.Session, error) {
	var os opencodeSession
	if err := b.doRequest(ctx, "GET", "/session/"+id, nil, &os); err != nil {
		return nil, err
	}
	return mapSession(&os), nil
}

func (b *OpencodeBackend) DeleteSession(ctx context.Context, id string) error {
	return b.doRequest(ctx, "DELETE", "/session/"+id, nil, nil)
}

func mapSession(os *opencodeSession) *backend.Session {
	return &backend.Session{
		ID:    os.ID,
		Title: os.Title,
	}
}
