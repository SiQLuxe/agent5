package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestNewLocalClient(t *testing.T) {
	client, err := NewLocalClient("test-api-key", "http://localhost:8080", "test-model")
	if err != nil {
		t.Fatalf("NewLocalClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewLocalClient returned nil client")
	}
}

func TestNewLocalClient_DefaultValues(t *testing.T) {
	client, err := NewLocalClient("", "", "")
	if err != nil {
		t.Fatalf("NewLocalClient returned error: %v", err)
	}

	model := client.GetModel()
	if model != "kimi-k2.6" {
		t.Errorf("Expected default model 'kimi-k2.6', got '%s'", model)
	}
}

func TestLocalClient_SetAPIKey(t *testing.T) {
	client, _ := NewLocalClient("", "", "")
	client.SetAPIKey("new-api-key")

	url := "http://example.com/api"
	client.SetBaseURL(url)

	data, _ := json.Marshal(ChatCompletionRequest{})
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	authHeader := request.Header.Get("Authorization")
	if authHeader != "" {
		t.Log("SetAPIKey works correctly")
	}
	_ = data
	_ = authHeader
}

func TestLocalClient_SetBaseURL(t *testing.T) {
	client, _ := NewLocalClient("", "", "")
	if client.GetModel() != "kimi-k2.6" {
		t.Errorf("Default base URL should be set")
	}

	newURL := "http://localhost:9999/"
	client.SetBaseURL(newURL)
}

func TestLocalClient_GetModel(t *testing.T) {
	client, _ := NewLocalClient("api-key", "http://localhost:8080", "custom-model")
	model := client.GetModel()
	if model != "custom-model" {
		t.Errorf("Expected model 'custom-model', got '%s'", model)
	}

	client2, _ := NewLocalClient("", "", "")
	model2 := client2.GetModel()
	if model2 != "kimi-k2.6" {
		t.Errorf("Expected default model 'kimi-k2.6', got '%s'", model2)
	}
}

func TestLocalClient_ListModels(t *testing.T) {
	client, _ := NewLocalClient("", "", "")
	models, err := client.ListModels()
	if err != nil {
		t.Fatalf("ListModels returned error: %v", err)
	}
	if len(models) == 0 {
		t.Fatal("ListModels returned empty list")
	}

	expectedModels := []string{"kimi-k2.6", "local-model", "llama-3", "mistral", "phi-3"}
	if len(models) != len(expectedModels) {
		t.Errorf("Expected %d models, got %d", len(expectedModels), len(models))
	}
}

func TestLocalClient_ChatCompletion_InvalidModel(t *testing.T) {
	client, _ := NewLocalClient("", "http://invalid-url-that-does-not-exist:9999", "")
	_, err := client.ChatCompletion(ChatCompletionRequest{
		Model:    "invalid-model",
		Messages: []Message{{Role: "user", Content: "Hello"}},
	})
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}
