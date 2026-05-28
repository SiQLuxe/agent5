package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type ModelConfig struct {
	APIKey       string `toml:"api_key"`
	BaseURL      string `toml:"base_url"`
	DefaultModel string `toml:"default_model"`
}

type ModelsConfig struct {
	OpenAI   ModelConfig `toml:"openai"`
	DeepSeek ModelConfig `toml:"deepseek"`
	Local    ModelConfig `toml:"local"`
}

type Config struct {
	Models        ModelsConfig `toml:"models"`
	DefaultClient string       `toml:"default_client"`
	Theme         string       `toml:"theme"`
	ApprovalMode  string       `toml:"approval_mode"`
	MaxSubagents  int          `toml:"max_subagents"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	applyEnvOverrides(&cfg)
	return &cfg, nil
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".agent-tui", "config.toml")
}

func GetDefaultConfig() *Config {
	return &Config{
		Models: ModelsConfig{
			OpenAI: ModelConfig{
				APIKey:       "",
				BaseURL:      "https://api.openai.com/v1",
				DefaultModel: "gpt-4",
			},
			DeepSeek: ModelConfig{
				APIKey:       "",
				BaseURL:      "https://api.deepseek.com",
				DefaultModel: "deepseek-chat",
			},
			Local: ModelConfig{
				BaseURL:      "http://localhost:11434",
				DefaultModel: "llama3",
			},
		},
		DefaultClient: "openai",
		Theme:         "dark",
		ApprovalMode:  "manual",
		MaxSubagents:  3,
	}
}

// ParseModelPrefix parses a default_model value of the form "<client>:<model>".
// Returns client prefix (e.g. "openai", "local") and model string.
func ParseModelPrefix(s string) (client, model string, err error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("invalid default_model format: %q; expected <client>:<model>", s)
	}
	prefix := strings.ToLower(strings.TrimSpace(parts[0]))
	switch prefix {
	case "openai", "deepseek", "anthropic", "local":
		return prefix, strings.TrimSpace(parts[1]), nil
	default:
		return "", "", fmt.Errorf("unknown client prefix: %s", prefix)
	}
}