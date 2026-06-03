package backend

import (
	"context"
	"testing"
	"time"
)

func TestNewProcessManager(t *testing.T) {
	pm := NewProcessManager("echo", []string{"hello"}, nil)
	if pm == nil {
		t.Fatal("expected non-nil ProcessManager")
	}
}

func TestProcessManager_StartStop(t *testing.T) {
	pm := NewProcessManager("echo", []string{"hello"}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := pm.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer pm.Stop(ctx)

	if !pm.Running() {
		t.Errorf("expected process to be running")
	}
}
