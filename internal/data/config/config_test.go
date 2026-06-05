package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestDefaultSandboxDir(t *testing.T) {
	cfg := GetDefaultConfig()
	for _, role := range cfg.AgentRoles {
		if role.SandboxDir != "" {
			t.Fatalf("expected empty sandbox_dir, got %q", role.SandboxDir)
		}
	}
}

func TestSandboxDirDefaultApplied(t *testing.T) {
	content := []byte(`
[[agent_role]]
name = "coder"
enabled = true
model = "test"
system_prompt = "test"
tools = ["read_file", "write_file"]
max_react_loop = 5
sandbox_dir = ""
`)
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.AgentRoles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(cfg.AgentRoles))
	}
	role := cfg.AgentRoles[0]
	if role.SandboxDir == "" {
		t.Fatal("expected SandboxDir to be set to default, got empty")
	}
	if !strings.HasPrefix(role.SandboxDir, os.TempDir()) {
		t.Fatalf("expected %q to start with %q", role.SandboxDir, os.TempDir())
	}
}

func TestSandboxDirPreservesExplicit(t *testing.T) {
	content := []byte(`
[[agent_role]]
name = "coder"
enabled = true
model = "test"
system_prompt = "test"
tools = ["read_file", "write_file"]
max_react_loop = 5
sandbox_dir = "/custom/path"
`)
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AgentRoles[0].SandboxDir != "/custom/path" {
		t.Fatalf("expected /custom/path, got %q", cfg.AgentRoles[0].SandboxDir)
	}
}
