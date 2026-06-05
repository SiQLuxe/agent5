package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIClientToolCallResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatCompletionRequest
		json.NewDecoder(r.Body).Decode(&req)
		if len(req.Tools) == 0 {
			t.Error("expected tools in request")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ChatCompletionResponse{
			Choices: []ResponseChoice{
				{
					Message: ResponseMessage{
						Role:    "assistant",
						Content: "",
						ToolCalls: []ToolCall{
							{
								ID:   "call_123",
								Type: "function",
								Function: ToolCallFunction{
									Name:      "read_file",
									Arguments: `{"path":"/tmp/test.txt"}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
		})
	}))
	defer server.Close()

	client := &OpenAIClient{
		apiKey:  "test",
		baseURL: server.URL,
		model:   "gpt-4",
		client:  &http.Client{},
	}

	resp, err := client.ChatCompletion(ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []Message{{Role: "user", Content: "read a file"}},
		Tools: []ToolDefinition{{
			Type: "function",
			Function: ToolFunction{
				Name:        "read_file",
				Description: "read a file",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{"type": "string"},
					},
					"required": []string{"path"},
				},
			},
		}},
	})
	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}
	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}
	if len(resp.Choices[0].Message.ToolCalls) == 0 {
		t.Fatal("expected tool calls in response")
	}
	if resp.Choices[0].Message.ToolCalls[0].Function.Name != "read_file" {
		t.Fatalf("expected read_file tool call, got %s", resp.Choices[0].Message.ToolCalls[0].Function.Name)
	}
}
