package backend

import (
	"testing"
)

func TestNewBackend_UnknownType(t *testing.T) {
	_, err := NewBackend("unknown", BackendConfig{})
	if err == nil {
		t.Errorf("expected error for unknown backend type")
	}
}

func TestNewBackend_Opencode(t *testing.T) {
	if _, ok := backendBuilders[TypeOpencode]; !ok {
		t.Skip("opencode builder not registered: run with opencode subpackage imported")
	}
	b, err := NewBackend(string(TypeOpencode), BackendConfig{
		Type: TypeOpencode, Enabled: true, AutoStart: false,
	})
	if err != nil {
		t.Fatalf("NewBackend(opencode) failed: %v", err)
	}
	if b == nil {
		t.Fatal("expected non-nil backend")
	}
}
