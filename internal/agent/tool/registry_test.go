package tool

import (
	"testing"
)

type mockTool struct{}

func (m *mockTool) Name() string                     { return "mock_tool" }
func (m *mockTool) Description() string              { return "A mock tool" }
func (m *mockTool) Schema() ToolSchema                { return ToolSchema{} }
func (m *mockTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	return ToolResult{Success: true, Data: "ok"}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{})

	got, ok := r.Get("mock_tool")
	if !ok {
		t.Fatal("expected to find mock_tool")
	}
	if got.Name() != "mock_tool" {
		t.Fatalf("expected name mock_tool, got %s", got.Name())
	}
}

func TestRegistryGetMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("expected false for missing tool")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{})

	list := r.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(list))
	}
}

func TestRegistryAsToolDefinitions(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{})

	defs := r.AsToolDefinitions()
	if len(defs) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(defs))
	}

	def := defs[0]
	if def["type"] != "function" {
		t.Fatalf("expected type 'function', got %v", def["type"])
	}

	fn := def["function"].(map[string]interface{})
	if fn["name"] != "mock_tool" {
		t.Fatalf("expected name 'mock_tool', got %v", fn["name"])
	}
}
