package orchestrator

import (
	"testing"

	"github.com/example/agent-tui/internal/agent/runtime"
	"github.com/example/agent-tui/internal/agent/tool"
)

func TestRegistryRegisterAndGet(t *testing.T) {
	reg := NewRegistry()
	agent := runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), nil)
	reg.Register("coder", agent, "code", "review")

	got, ok := reg.Get("coder")
	if !ok {
		t.Fatal("expected to find coder")
	}
	if got.Name != "coder" {
		t.Fatalf("expected name coder, got %s", got.Name)
	}
}

func TestRegistryFindByRole(t *testing.T) {
	reg := NewRegistry()
	reg.Register("planner", runtime.NewAgent(runtime.Config{Name: "planner"}, tool.NewRegistry(), nil), "analyze", "design")
	reg.Register("coder", runtime.NewAgent(runtime.Config{Name: "coder"}, tool.NewRegistry(), nil), "code")

	agents := reg.FindByRole("code")
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent for role 'code', got %d", len(agents))
	}
	agents = reg.FindByRole("analyze")
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent for role 'analyze', got %d", len(agents))
	}
	agents = reg.FindByRole("nonexistent")
	if len(agents) != 0 {
		t.Fatalf("expected 0 agents for missing role, got %d", len(agents))
	}
}
