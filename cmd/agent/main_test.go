package main

import (
	"encoding/json"
	"testing"

	"github.com/example/agent-tui/internal/ai"
	"github.com/example/agent-tui/internal/agent/runtime"
)

type mockAIClient struct {
	chatCompletionFn func(req ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error)
	chatStreamFn     func(req ai.ChatCompletionRequest, cb func(string)) error
}

func (m *mockAIClient) ChatCompletion(req ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
	return m.chatCompletionFn(req)
}
func (m *mockAIClient) ChatCompletionStream(req ai.ChatCompletionRequest, cb func(string)) error {
	return m.chatStreamFn(req, cb)
}
func (m *mockAIClient) SetAPIKey(key string)             {}
func (m *mockAIClient) SetBaseURL(url string)            {}
func (m *mockAIClient) GetModel() string                 { return "mock" }
func (m *mockAIClient) ListModels() ([]string, error)    { return []string{"mock"}, nil }

func TestChatWithTools_PassesToolsInRequest(t *testing.T) {
	var capturedReq ai.ChatCompletionRequest
	client := &mockAIClient{
		chatCompletionFn: func(req ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
			capturedReq = req
			return &ai.ChatCompletionResponse{
				Choices: []ai.ResponseChoice{
					{Message: ai.ResponseMessage{Role: "assistant", Content: "hello"}},
				},
			}, nil
		},
	}

	adapter := &aiLLMAdapter{client: client, model: "test-model"}
	tools := []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "write_file",
				"description": "write a file",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path":    map[string]interface{}{"type": "string"},
						"content": map[string]interface{}{"type": "string"},
					},
					"required": []interface{}{"path", "content"},
				},
			},
		},
	}

	_, err := adapter.ChatWithTools(
		[]runtime.Message{{Role: "user", Content: "write hello.py"}},
		tools,
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedReq.Tools) != 1 {
		t.Fatalf("expected 1 tool in request, got %d", len(capturedReq.Tools))
	}
	if capturedReq.Tools[0].Type != "function" {
		t.Fatalf("expected tool type 'function', got %q", capturedReq.Tools[0].Type)
	}
	if capturedReq.Tools[0].Function.Name != "write_file" {
		t.Fatalf("expected function name 'write_file', got %q", capturedReq.Tools[0].Function.Name)
	}
}

func TestChatWithTools_ReturnsToolCall(t *testing.T) {
	client := &mockAIClient{
		chatCompletionFn: func(req ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
			return &ai.ChatCompletionResponse{
				Choices: []ai.ResponseChoice{
					{
						Message: ai.ResponseMessage{
							Role:    "assistant",
							Content: "",
							ToolCalls: []ai.ToolCall{
								{
									ID:   "call_123",
									Type: "function",
									Function: ai.ToolCallFunction{
										Name:      "write_file",
										Arguments: `{"path": "hello.py", "content": "print(\"hello\")"}`,
									},
								},
							},
						},
						FinishReason: "tool_calls",
					},
				},
			}, nil
		},
	}

	adapter := &aiLLMAdapter{client: client, model: "test-model"}
	resp, err := adapter.ChatWithTools(
		[]runtime.Message{{Role: "user", Content: "write hello.py"}},
		[]map[string]interface{}{},
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Type != "tool_call" {
		t.Fatalf("expected Type 'tool_call', got %q", resp.Type)
	}
	if resp.ToolCall == nil {
		t.Fatal("expected ToolCall to be non-nil")
	}
	if resp.ToolCall.Name != "write_file" {
		t.Fatalf("expected ToolCall.Name 'write_file', got %q", resp.ToolCall.Name)
	}
	path, _ := resp.ToolCall.Arguments["path"].(string)
	if path != "hello.py" {
		t.Fatalf("expected path 'hello.py', got %q", path)
	}
	content, _ := resp.ToolCall.Arguments["content"].(string)
	if content != `print("hello")` {
		t.Fatalf("expected content 'print(\"hello\")', got %q", content)
	}
}

func TestChatWithTools_ReturnsFinalWhenNoToolCalls(t *testing.T) {
	client := &mockAIClient{
		chatCompletionFn: func(req ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
			return &ai.ChatCompletionResponse{
				Choices: []ai.ResponseChoice{
					{
						Message: ai.ResponseMessage{
							Role:    "assistant",
							Content: "Sure, I'll write a hello world program.",
						},
						FinishReason: "stop",
					},
				},
			}, nil
		},
	}

	adapter := &aiLLMAdapter{client: client, model: "test-model"}
	resp, err := adapter.ChatWithTools(
		[]runtime.Message{{Role: "user", Content: "write hello.py"}},
		[]map[string]interface{}{},
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Type != "final" {
		t.Fatalf("expected Type 'final', got %q", resp.Type)
	}
	if resp.Content != "Sure, I'll write a hello world program." {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
}

func TestChatWithToolsStream_ReturnsToolCallDirectly(t *testing.T) {
	client := &mockAIClient{
		chatCompletionFn: func(req ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
			return &ai.ChatCompletionResponse{
				Choices: []ai.ResponseChoice{
					{
						Message: ai.ResponseMessage{
							Role: "assistant",
							ToolCalls: []ai.ToolCall{
								{
									ID:   "call_123",
									Type: "function",
									Function: ai.ToolCallFunction{
										Name:      "write_file",
										Arguments: `{"path": "hello.py", "content": "print(\"hello\")"}`,
									},
								},
							},
						},
						FinishReason: "tool_calls",
					},
				},
			}, nil
		},
	}

	adapter := &aiLLMAdapter{client: client, model: "test-model"}
	var streamedChunks []string
	resp, err := adapter.ChatWithToolsStream(
		[]runtime.Message{{Role: "user", Content: "write hello.py"}},
		[]map[string]interface{}{},
		"",
		func(chunk string) { streamedChunks = append(streamedChunks, chunk) },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Type != "tool_call" {
		t.Fatalf("expected Type 'tool_call', got %q", resp.Type)
	}
	if len(streamedChunks) != 0 {
		t.Fatalf("expected no streamed chunks for tool_call, got %d", len(streamedChunks))
	}
}

func TestChatWithToolsStream_StreamsFinalText(t *testing.T) {
	client := &mockAIClient{
		chatCompletionFn: func(req ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
			return &ai.ChatCompletionResponse{
				Choices: []ai.ResponseChoice{
					{
						Message: ai.ResponseMessage{
							Role:    "assistant",
							Content: "first call text",
						},
					},
				},
			}, nil
		},
		chatStreamFn: func(req ai.ChatCompletionRequest, cb func(string)) error {
			cb("streamed ")
			cb("response")
			return nil
		},
	}

	adapter := &aiLLMAdapter{client: client, model: "test-model"}
	var streamedChunks []string
	resp, err := adapter.ChatWithToolsStream(
		[]runtime.Message{{Role: "user", Content: "write hello.py"}},
		[]map[string]interface{}{},
		"",
		func(chunk string) { streamedChunks = append(streamedChunks, chunk) },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Type != "final" {
		t.Fatalf("expected Type 'final', got %q", resp.Type)
	}
	if resp.Content != "streamed response" {
		t.Fatalf("expected 'streamed response', got %q", resp.Content)
	}
	if len(streamedChunks) != 2 {
		t.Fatalf("expected 2 streamed chunks, got %d", len(streamedChunks))
	}
}

func TestChatWithToolsStream_StreamingFallbackIncludesTools(t *testing.T) {
	var capturedReq ai.ChatCompletionRequest
	client := &mockAIClient{
		chatCompletionFn: func(req ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
			return &ai.ChatCompletionResponse{
				Choices: []ai.ResponseChoice{
					{
						Message: ai.ResponseMessage{
							Role:    "assistant",
							Content: "I'll help you write that.",
						},
					},
				},
			}, nil
		},
		chatStreamFn: func(req ai.ChatCompletionRequest, cb func(string)) error {
			capturedReq = req
			cb("done")
			return nil
		},
	}

	adapter := &aiLLMAdapter{client: client, model: "test-model"}
	tools := []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name": "write_file",
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
	}

	_, err := adapter.ChatWithToolsStream(
		[]runtime.Message{{Role: "user", Content: "write hello.py"}},
		tools,
		"",
		func(chunk string) {},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedReq.Tools) != 1 {
		t.Fatalf("expected 1 tool in streaming fallback request, got %d", len(capturedReq.Tools))
	}
	if capturedReq.Tools[0].Function.Name != "write_file" {
		t.Fatalf("expected 'write_file' in streaming fallback, got %q", capturedReq.Tools[0].Function.Name)
	}
}

func TestToolDefinitionJSONRoundTrip(t *testing.T) {
	input := map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "write_file",
			"description": "Write or create a file",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":    map[string]interface{}{"type": "string", "description": "Path to the file"},
					"content": map[string]interface{}{"type": "string", "description": "Content to write"},
				},
				"required": []interface{}{"path", "content"},
			},
		},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var def ai.ToolDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if def.Type != "function" {
		t.Fatalf("expected type 'function', got %q", def.Type)
	}
	if def.Function.Name != "write_file" {
		t.Fatalf("expected name 'write_file', got %q", def.Function.Name)
	}
	if def.Function.Description != "Write or create a file" {
		t.Fatalf("expected description 'Write or create a file', got %q", def.Function.Description)
	}
}

func TestToolCallJSONUnmarshal(t *testing.T) {
	jsonData := `{
		"id": "call_abc123",
		"type": "function",
		"function": {
			"name": "write_file",
			"arguments": "{\"path\": \"hello.py\", \"content\": \"print(\\\"hello\\\")\"}"
		}
	}`

	var tc ai.ToolCall
	if err := json.Unmarshal([]byte(jsonData), &tc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if tc.ID != "call_abc123" {
		t.Fatalf("expected ID 'call_abc123', got %q", tc.ID)
	}
	if tc.Type != "function" {
		t.Fatalf("expected type 'function', got %q", tc.Type)
	}
	if tc.Function.Name != "write_file" {
		t.Fatalf("expected name 'write_file', got %q", tc.Function.Name)
	}
	if tc.Function.Arguments != `{"path": "hello.py", "content": "print(\"hello\")"}` {
		t.Fatalf("unexpected arguments: %q", tc.Function.Arguments)
	}
}


