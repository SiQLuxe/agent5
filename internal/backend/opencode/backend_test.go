package opencode

import (
	"testing"

	"github.com/example/agent-tui/internal/backend"
)

func TestNewBackend(t *testing.T) {
	b := NewBackend(Config{
		Binary:    "opencode",
		AutoStart: false,
		APIURL:    "http://127.0.0.1:4096",
	})
	if b == nil {
		t.Fatal("expected non-nil backend")
	}
}

func TestBackend_Name(t *testing.T) {
	b := NewBackend(Config{AutoStart: false, APIURL: "http://127.0.0.1:4096"})
	if b.Name() != "opencode" {
		t.Errorf("Name() = %q, want %q", b.Name(), "opencode")
	}
}

func TestBackendRegistration(t *testing.T) {
	b, err := backend.NewBackend(string(backend.TypeOpencode), backend.BackendConfig{
		Type: backend.TypeOpencode, Enabled: true, AutoStart: false,
	})
	if err != nil {
		t.Fatalf("NewBackend(opencode) failed: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil backend")
	}
}
