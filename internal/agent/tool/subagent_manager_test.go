package tool

import (
	"context"
	"testing"
)

func TestSubagentManagerSpawn(t *testing.T) {
	mgr := NewSubagentManager(5)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sa := mgr.Spawn(ctx, SubAgentConfig{
		Name:         "test-agent",
		SystemPrompt: "You are a test agent",
		Model:        "test-model",
	})
	if sa.ID == "" {
		t.Fatal("expected non-empty agent ID")
	}
	if sa.Name != "test-agent" {
		t.Errorf("expected name 'test-agent', got %s", sa.Name)
	}
	if sa.Status != StatusRunning {
		t.Errorf("expected status '%s', got '%s'", StatusRunning, sa.Status)
	}

	got := mgr.Get(sa.ID)
	if got == nil {
		t.Fatal("expected to find spawned agent")
	}
	if got.ID != sa.ID {
		t.Errorf("expected ID %s, got %s", sa.ID, got.ID)
	}
}

func TestSubagentManagerConcurrencyLimit(t *testing.T) {
	mgr := NewSubagentManager(2)
	ctx := context.Background()

	sa1 := mgr.Spawn(ctx, SubAgentConfig{Name: "agent1"})
	sa2 := mgr.Spawn(ctx, SubAgentConfig{Name: "agent2"})
	sa3 := mgr.Spawn(ctx, SubAgentConfig{Name: "agent3"})

	if sa1.Status != StatusRunning {
		t.Errorf("sa1 should be '%s', got '%s'", StatusRunning, sa1.Status)
	}
	if sa2.Status != StatusRunning {
		t.Errorf("sa2 should be '%s', got '%s'", StatusRunning, sa2.Status)
	}
	if sa3.Status != StatusFailed {
		t.Errorf("sa3 should be '%s' (limit), got '%s'", StatusFailed, sa3.Status)
	}
	if sa3.Error == "" {
		t.Fatal("expected error message for agent3")
	}
}

func TestSubagentManagerCancel(t *testing.T) {
	mgr := NewSubagentManager(5)
	ctx, cancel := context.WithCancel(context.Background())

	sa := mgr.Spawn(ctx, SubAgentConfig{Name: "test-agent"})
	cancel()
	if !mgr.IsCancelled(sa.ID) {
		t.Fatal("expected agent to be cancelled after context cancel")
	}
}

func TestSubagentManagerList(t *testing.T) {
	mgr := NewSubagentManager(5)
	ctx := context.Background()

	mgr.Spawn(ctx, SubAgentConfig{Name: "alpha"})
	mgr.Spawn(ctx, SubAgentConfig{Name: "beta"})

	agents := mgr.List()
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}
}

func TestSubagentManagerRemove(t *testing.T) {
	mgr := NewSubagentManager(5)
	ctx := context.Background()

	sa := mgr.Spawn(ctx, SubAgentConfig{Name: "test"})
	if sa.Status != StatusRunning {
		t.Fatalf("expected running, got %s", sa.Status)
	}

	mgr.Remove(sa.ID)
	if mgr.Get(sa.ID) != nil {
		t.Fatal("expected agent to be removed")
	}
}

func TestSubagentManagerGetReturnsCopy(t *testing.T) {
	mgr := NewSubagentManager(5)
	ctx := context.Background()

	sa := mgr.Spawn(ctx, SubAgentConfig{Name: "test"})
	got1 := mgr.Get(sa.ID)
	got2 := mgr.Get(sa.ID)

	if got1 == got2 {
		t.Fatal("Get() should return copies, not the same pointer")
	}
}
