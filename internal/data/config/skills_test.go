package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultSkillsDirs(t *testing.T) {
	got := DefaultSkillsDirs()
	want := []string{"skills", "~/.config/agent-tui/skills"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DefaultSkillsDirs = %v, want %v", got, want)
	}
}

func TestGetDefaultConfig_SkillsDirs(t *testing.T) {
	cfg := GetDefaultConfig()
	if !reflect.DeepEqual(cfg.Skills.Dirs, DefaultSkillsDirs()) {
		t.Fatalf("default Skills.Dirs = %v, want %v", cfg.Skills.Dirs, DefaultSkillsDirs())
	}
}

func TestLoadConfig_SkillsDirsOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `[skills]
dirs = ["a", "b", "c"]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(cfg.Skills.Dirs, want) {
		t.Fatalf("Skills.Dirs = %v, want %v", cfg.Skills.Dirs, want)
	}
}

func TestLoadConfig_SkillsDirsAbsent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	// No [skills] section at all.
	if err := os.WriteFile(path, []byte(`default_client = "openai"`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	// Absent section → empty slice (main.go falls back to DefaultSkillsDirs).
	if len(cfg.Skills.Dirs) != 0 {
		t.Fatalf("expected empty Skills.Dirs when section absent, got %v", cfg.Skills.Dirs)
	}
}

func TestExpandSkillPath_Relative(t *testing.T) {
	if got := ExpandSkillPath("skills"); got != "skills" {
		t.Fatalf("relative path should be unchanged, got %q", got)
	}
}

func TestExpandSkillPath_Absolute(t *testing.T) {
	in := "/etc/skills"
	if got := ExpandSkillPath(in); got != in {
		t.Fatalf("absolute path should be unchanged, got %q", got)
	}
}

func TestExpandSkillPath_HomeTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	got := ExpandSkillPath("~/foo/bar")
	want := filepath.Join(home, "foo", "bar")
	if got != want {
		t.Fatalf("ExpandSkillPath(~/foo/bar) = %q, want %q", got, want)
	}
}

func TestExpandSkillPath_XDGConfigHome(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	got := ExpandSkillPath("~/.config/agent-tui/skills")
	want := filepath.Join(xdg, "agent-tui/skills")
	if got != want {
		t.Fatalf("with XDG_CONFIG_HOME set: got %q, want %q", got, want)
	}
}

func TestExpandSkillPath_NoXDG_FallsBackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}

	got := ExpandSkillPath("~/.config/agent-tui/skills")
	want := filepath.Join(home, ".config", "agent-tui", "skills")
	if got != want {
		t.Fatalf("without XDG: got %q, want %q", got, want)
	}
}
