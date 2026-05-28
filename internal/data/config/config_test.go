package config

import "testing"

func TestGetDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()
	if cfg.DefaultClient == "" {
		t.Errorf("DefaultClient should not be empty")
	}

	if cfg.Theme != "dark" {
		t.Errorf("Theme should be 'dark', got '%s'", cfg.Theme)
	}
	if cfg.MaxSubagents != 3 {
		t.Errorf("MaxSubagents should be 3, got %d", cfg.MaxSubagents)
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("Config path should not be empty")
	}
}

func TestLoadConfig(t *testing.T) {
	_, err := LoadConfig("non-existent.toml")
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}
}

func TestConfigStructure(t *testing.T) {
	cfg := GetDefaultConfig()
	if cfg.DefaultClient != "openai" {
		t.Errorf("expected default client 'openai', got '%s'", cfg.DefaultClient)
	}
	if cfg.Models.OpenAI.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("unexpected openai base url: %s", cfg.Models.OpenAI.BaseURL)
	}
	if cfg.Models.OpenAI.DefaultModel != "gpt-4" {
		t.Errorf("unexpected openai default model: %s", cfg.Models.OpenAI.DefaultModel)
	}
}