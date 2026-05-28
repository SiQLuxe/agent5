package config

import (
	"os"
	"testing"
)

func TestEnvOverrides(t *testing.T) {
    os.Setenv("OPENAI_API_KEY", "test-key")
    defer os.Unsetenv("OPENAI_API_KEY")

    cfg := &Config{
        Models: ModelsConfig{
            OpenAI: ModelConfig{APIKey: "original"},
        },
    }
    applyEnvOverrides(cfg)

    if cfg.Models.OpenAI.APIKey != "test-key" {
        t.Errorf("expected 'test-key', got '%s'", cfg.Models.OpenAI.APIKey)
    }
}

func TestEnvOverridesDefaultClient(t *testing.T) {
    os.Setenv("DEFAULT_CLIENT", "openai")
    defer os.Unsetenv("DEFAULT_CLIENT")

    cfg := &Config{
        DefaultClient: "deepseek",
    }
    applyEnvOverrides(cfg)

    if cfg.DefaultClient != "openai" {
        t.Errorf("expected 'openai', got '%s'", cfg.DefaultClient)
    }
}
