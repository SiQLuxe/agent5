package bridge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/agent-tui/internal/agent/tool"
)

func TestOpencodeBridgeName(t *testing.T) {
	b := &OpencodeBridge{baseURL: "http://localhost:4096"}
	if b.Name() != "delegate_opencode" {
		t.Fatalf("expected name 'delegate_opencode', got %s", b.Name())
	}
}

func TestOpencodeBridgeSendTask(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"id": "sess_123", "title": "test"})
	})
	mux.HandleFunc("/session/sess_123/message", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"info": map[string]string{"id": "msg_1"},
			"parts": []map[string]string{
				{"type": "text", "text": "task result"},
			},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	b := NewOpencodeBridge(srv.URL, "")
	ctx := tool.ToolContext{Context: context.Background()}
	result := b.Execute(ctx, map[string]interface{}{
		"task": "implement login feature",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Data.(string) != "task result" {
		t.Fatalf("expected 'task result', got %v", result.Data)
	}
}

func TestOpencodeBridgeMissingTask(t *testing.T) {
	b := &OpencodeBridge{baseURL: "http://localhost:4096"}
	ctx := tool.ToolContext{Context: context.Background()}
	result := b.Execute(ctx, map[string]interface{}{})
	if result.Success {
		t.Fatal("expected failure for missing task param")
	}
}
