package service

import (
	"errors"
	"testing"

	"github.com/example/agent-tui/internal/ai"
	"github.com/example/agent-tui/internal/data/history"
)

type MockAIClient struct {
	chatCalled     bool
	lastRequest    ai.ChatCompletionRequest
	mockResponse   *ai.ChatCompletionResponse
	mockError      error
	streamCallback func(string)
}

func (m *MockAIClient) ChatCompletion(req ai.ChatCompletionRequest) (*ai.ChatCompletionResponse, error) {
	m.chatCalled = true
	m.lastRequest = req
	if m.mockError != nil {
		return nil, m.mockError
	}
	return m.mockResponse, nil
}

func (m *MockAIClient) ChatCompletionStream(req ai.ChatCompletionRequest, callback func(string)) error {
	m.chatCalled = true
	m.lastRequest = req
	if m.streamCallback != nil {
		m.streamCallback("stream response")
	}
	callback("stream response")
	return m.mockError
}

func (m *MockAIClient) SetAPIKey(key string) {}

func (m *MockAIClient) SetBaseURL(url string) {}

func (m *MockAIClient) GetModel() string {
	return "mock-model"
}

func (m *MockAIClient) ListModels() ([]string, error) {
	return []string{"mock-model"}, nil
}

func TestNewAIAssistant(t *testing.T) {
	h := history.NewHistory("")
	client := &MockAIClient{}

	assistant := NewAIAssistant(client, h)

	if assistant == nil {
		t.Error("NewAIAssistant returned nil")
	}

	if assistant.client != client {
		t.Error("client not set correctly")
	}

	if assistant.history != h {
		t.Error("history not set correctly")
	}
}

func TestAIAssistant_Chat(t *testing.T) {
	h := history.NewHistory("")
	sessionID := h.CreateSession("test session")

	mockClient := &MockAIClient{
		mockResponse: &ai.ChatCompletionResponse{
			Choices: []struct {
				Index        int     `json:"index"`
				Message      ai.Message `json:"message"`
				FinishReason string  `json:"finish_reason"`
			}{
				{
					Message: ai.Message{
						Role:    "assistant",
						Content: "Hello, how can I help you?",
					},
				},
			},
		},
	}

	assistant := NewAIAssistant(mockClient, h)

	response, err := assistant.Chat(sessionID, "Hello")

	if err != nil {
		t.Errorf("Chat returned unexpected error: %v", err)
	}

	if !mockClient.chatCalled {
		t.Error("Chat method did not call the client")
	}

	if response != "Hello, how can I help you?" {
		t.Errorf("Expected response 'Hello, how can I help you?', got '%s'", response)
	}

	messages := h.GetMessages(sessionID)
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages in history, got %d", len(messages))
	}

	if messages[0].Role != "user" {
		t.Errorf("First message role should be 'user', got '%s'", messages[0].Role)
	}

	if messages[0].Content != "Hello" {
		t.Errorf("First message content should be 'Hello', got '%s'", messages[0].Content)
	}

	if messages[1].Role != "assistant" {
		t.Errorf("Second message role should be 'assistant', got '%s'", messages[1].Role)
	}

	if messages[1].Content != "Hello, how can I help you?" {
		t.Errorf("Second message content should be 'Hello, how can I help you?', got '%s'", messages[1].Content)
	}
}

func TestAIAssistant_Chat_Error(t *testing.T) {
	h := history.NewHistory("")
	sessionID := h.CreateSession("test session")

	mockClient := &MockAIClient{
		mockError: errors.New("API error"),
	}

	assistant := NewAIAssistant(mockClient, h)

	response, err := assistant.Chat(sessionID, "Hello")

	if err == nil {
		t.Error("Chat should return error when client fails")
	}

	if response != "" {
		t.Errorf("Expected empty response on error, got '%s'", response)
	}

	if err.Error() != "API error" {
		t.Errorf("Expected error message 'API error', got '%s'", err.Error())
	}
}

func TestAIAssistant_CreateSession(t *testing.T) {
	h := history.NewHistory("")
	client := &MockAIClient{}
	assistant := NewAIAssistant(client, h)

	sessionID := assistant.CreateSession("my session")

	if sessionID == "" {
		t.Error("CreateSession returned empty ID")
	}

	sessions := h.GetSessions()
	if len(sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessions))
	}

	if sessions[0].Name != "my session" {
		t.Errorf("Expected session name 'my session', got '%s'", sessions[0].Name)
	}
}

func TestAIAssistant_ListSessions(t *testing.T) {
	h := history.NewHistory("")
	client := &MockAIClient{}
	assistant := NewAIAssistant(client, h)

	assistant.CreateSession("session 1")
	assistant.CreateSession("session 2")
	assistant.CreateSession("session 3")

	sessions := assistant.ListSessions()

	if len(sessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(sessions))
	}

	sessionNames := make(map[string]bool)
	for _, s := range sessions {
		sessionNames[s.Name] = true
	}

	expectedNames := []string{"session 1", "session 2", "session 3"}
	for _, expected := range expectedNames {
		if !sessionNames[expected] {
			t.Errorf("Expected session '%s' not found", expected)
		}
	}
}

func TestAIAssistant_ChatStream(t *testing.T) {
	h := history.NewHistory("")
	sessionID := h.CreateSession("test stream session")

	streamContent := ""
	mockClient := &MockAIClient{
		mockError: nil,
		streamCallback: func(content string) {
			streamContent += content
		},
	}

	assistant := NewAIAssistant(mockClient, h)

	var callbackContent string
	err := assistant.ChatStream(sessionID, "Hello", func(content string) {
		callbackContent += content
	})

	if err != nil {
		t.Errorf("ChatStream returned unexpected error: %v", err)
	}

	if !mockClient.chatCalled {
		t.Error("ChatStream method did not call the client")
	}

	if streamContent != "stream response" {
		t.Errorf("Expected stream content 'stream response', got '%s'", streamContent)
	}

	if callbackContent != "stream response" {
		t.Errorf("Expected callback content 'stream response', got '%s'", callbackContent)
	}

	messages := h.GetMessages(sessionID)
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages in history after stream, got %d", len(messages))
	}

	if messages[1].Content != "stream response" {
		t.Errorf("Assistant message should be 'stream response', got '%s'", messages[1].Content)
	}
}
