package ai

import (
	"testing"

	cfg "github.com/example/agent-tui/internal/data/config"
)

func TestNewClientFromConfig_OpenAI_Success(t *testing.T) {
	c := &cfg.Config{
		DefaultClient: "openai",
		Models: cfg.ModelsConfig{
			OpenAI: cfg.ModelConfig{
				APIKey:       "sk-test",
				BaseURL:      "https://api.openai.com/v1",
				DefaultModel: "gpt-4",
			},
		},
	}
	client, err := NewClientFromConfig(c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatalf("expected client, got nil")
	}
}

func TestNewClientFromConfig_OpenAI_MissingKey(t *testing.T) {
	c := &cfg.Config{
		DefaultClient: "openai",
		Models: cfg.ModelsConfig{
			OpenAI: cfg.ModelConfig{
				APIKey: "",
			},
		},
	}
	_, err := NewClientFromConfig(c)
	if err == nil {
		t.Fatalf("expected error when openai key missing")
	}
}

func TestNewClientFromConfig_Local(t *testing.T) {
	c := &cfg.Config{
		DefaultClient: "local",
		Models: cfg.ModelsConfig{
			Local: cfg.ModelConfig{
				BaseURL:      "http://localhost:4004",
				DefaultModel: "kimi-k2.6",
			},
		},
	}
	client, err := NewClientFromConfig(c)
	if err != nil {
		t.Fatalf("expected no error for local, got %v", err)
	}
	if client == nil {
		t.Fatalf("expected local client, got nil")
	}
}

func TestNewClientFromConfig_UnknownPrefix(t *testing.T) {
	c := &cfg.Config{
		DefaultClient: "foo",
	}
	_, err := NewClientFromConfig(c)
	if err == nil {
		t.Fatalf("expected error for unknown prefix")
	}
}
