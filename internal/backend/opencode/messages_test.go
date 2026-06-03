package opencode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/agent-tui/internal/backend"
)

func setupMessagesServer() (*httptest.Server, *OpencodeBackend) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	mux.HandleFunc("/session/s1/message", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(opencodeMessageResponse{
			Parts: []opencodePart{{Type: "text", Text: "reply"}},
		})
	})

	b := NewBackend(Config{AutoStart: false, APIURL: srv.URL})
	return srv, b
}

func TestSendMessage(t *testing.T) {
	srv, b := setupMessagesServer()
	defer srv.Close()

	result, err := b.SendMessage(context.Background(), "s1", &backend.Message{
		Role: backend.RoleUser, Content: "hello",
	})
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if result.Content != "reply" {
		t.Errorf("Content = %q, want %q", result.Content, "reply")
	}
}
