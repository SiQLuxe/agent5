package tool

import (
	"context"
	"testing"
)

func TestChatLLMToolName(t *testing.T) {
	tool := &ChatLLMTool{Provider: &mockLLMProvider{}}
	if tool.Name() != "chat_llm" {
		t.Fatalf("expected name 'chat_llm', got %s", tool.Name())
	}
}

type mockLLMProvider struct{}

func (m *mockLLMProvider) Chat(model, systemPrompt, userPrompt string) (string, error) {
	return "mock response: " + userPrompt, nil
}

func TestChatLLMToolExecute(t *testing.T) {
	tool := &ChatLLMTool{Provider: &mockLLMProvider{}}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{
		"prompt": "say hi",
	})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Data.(string) != "mock response: say hi" {
		t.Fatalf("unexpected response: %v", result.Data)
	}
}

func TestChatLLMToolMissingPrompt(t *testing.T) {
	tool := &ChatLLMTool{Provider: &mockLLMProvider{}}
	result := tool.Execute(ToolContext{Context: context.Background()}, map[string]interface{}{})
	if result.Success {
		t.Fatal("expected failure for missing prompt")
	}
}
