package opencode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupCommandsServer() (*httptest.Server, *OpencodeBackend) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	mux.HandleFunc("/session/s1/command", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"parts": []map[string]interface{}{
				{"type": "text", "text": "cmd output"},
			},
		})
	})
	mux.HandleFunc("/file/content", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"type": "raw", "content": "file data",
		})
	})

	b := NewBackend(Config{AutoStart: false, APIURL: srv.URL})
	return srv, b
}

func TestExecuteCommand(t *testing.T) {
	srv, b := setupCommandsServer()
	defer srv.Close()

	result, err := b.ExecuteCommand(context.Background(), "s1", "/help")
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}
	if !strings.Contains(result.Stdout, "cmd output") {
		t.Errorf("expected cmd output")
	}
}

func TestReadFile(t *testing.T) {
	srv, b := setupCommandsServer()
	defer srv.Close()

	content, err := b.ReadFile(context.Background(), "main.go")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if content != "file data" {
		t.Errorf("content = %q, want %q", content, "file data")
	}
}
