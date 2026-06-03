package opencode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupSessionsServer() (*httptest.Server, *OpencodeBackend) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			json.NewEncoder(w).Encode([]opencodeSession{
				{ID: "s1", Title: "test", Status: "active"},
			})
		case "POST":
			var s opencodeSession
			json.NewDecoder(r.Body).Decode(&s)
			s.ID = "new-id"
			json.NewEncoder(w).Encode(s)
		}
	})
	mux.HandleFunc("/session/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&opencodeSession{ID: "s1", Title: "test", Status: "active"})
	})

	b := NewBackend(Config{AutoStart: false, APIURL: srv.URL})
	return srv, b
}

func TestCreateSession(t *testing.T) {
	srv, b := setupSessionsServer()
	defer srv.Close()

	sess, err := b.CreateSession(context.Background(), "my session")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if sess.Title != "my session" {
		t.Errorf("Title = %q, want %q", sess.Title, "my session")
	}
}

func TestListSessions(t *testing.T) {
	srv, b := setupSessionsServer()
	defer srv.Close()

	sessions, err := b.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("expected 1, got %d", len(sessions))
	}
}
