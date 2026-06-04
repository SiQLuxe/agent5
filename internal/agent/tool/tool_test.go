package tool

import (
	"context"
	"testing"
)

func TestToolSchemaValidation(t *testing.T) {
	schema := ToolSchema{
		Parameters: map[string]ParamSchema{
			"name": {Type: "string", Description: "A name"},
			"count": {Type: "integer", Description: "A count"},
		},
		Required: []string{"name"},
	}

	if len(schema.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(schema.Parameters))
	}
	if len(schema.Required) != 1 {
		t.Fatalf("expected 1 required, got %d", len(schema.Required))
	}
}

func TestToolResult(t *testing.T) {
	r := ToolResult{Success: true, Data: "hello"}
	if !r.Success {
		t.Fatal("expected success")
	}
	if r.Data.(string) != "hello" {
		t.Fatalf("expected 'hello', got %v", r.Data)
	}
}

func TestToolContext(t *testing.T) {
	ctx := context.Background()
	tc := ToolContext{Context: ctx}
	if tc.Context == nil {
		t.Fatal("expected non-nil context")
	}
}
