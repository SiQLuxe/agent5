package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_AppendPrompt(t *testing.T) {
	s := New(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := map[string]string{"text": "hello"}
	data, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/api/tui/append-prompt", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestServer_SubmitPrompt(t *testing.T) {
	s := New(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/api/tui/submit-prompt", "application/json", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestServer_ShowToast(t *testing.T) {
	s := New(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := map[string]string{"message": "done", "variant": "success"}
	data, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/api/tui/show-toast", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestServer_ExecuteCommand(t *testing.T) {
	s := New(nil)
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	body := map[string]string{"command": "/help"}
	data, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/api/tui/execute-command", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
