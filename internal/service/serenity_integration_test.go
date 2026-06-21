package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockChater captures the prompt sent to the LLM for assertion.
type mockChater struct {
	lastSessionID string
	lastMessage   string
	reply         string
}

func (m *mockChater) Chat(sessionID, message string) (string, error) {
	m.lastSessionID = sessionID
	m.lastMessage = message
	if m.reply != "" {
		return m.reply, nil
	}
	return "MOCK_REPLY", nil
}

// TestSerenitySkill_LoadsFromUserLevelDir verifies serenity-skill is
// discoverable from a user-level skills directory (mirroring
// ~/.config/agent-tui/skills/serenity-skill).
//
// This test is skipped when the real user-level install is absent; it
// otherwise confirms the skill ships with the expected frontmatter and
// a non-empty methodology body.
func TestSerenitySkill_LoadsFromUserLevelDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	userSkillsDir := filepath.Join(home, ".config", "agent-tui", "skills")
	if _, err := os.Stat(filepath.Join(userSkillsDir, "serenity-skill", "SKILL.md")); err != nil {
		t.Skipf("serenity-skill not installed at %s", userSkillsDir)
	}

	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, userSkillsDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}

	s, ok := r.Get("serenity-skill")
	if !ok {
		t.Fatal("expected serenity-skill registered from user-level dir")
	}
	if !strings.Contains(strings.ToLower(s.Description), "supply-chain") {
		t.Fatalf("unexpected description: %q", s.Description)
	}
	if s.Type != SkillPrompt {
		t.Fatalf("expected SkillPrompt, got %v", s.Type)
	}
	if len(s.Prompt) < 500 {
		t.Fatalf("serenity prompt body too short (%d bytes), expected full methodology", len(s.Prompt))
	}
	if !strings.Contains(s.Prompt, "scarce layer") {
		t.Fatalf("expected 'scarce layer' methodology keyword in prompt")
	}
}

// TestSerenitySkill_ArgsNotInjectedWhenNoPlaceholder exposes a design
// gap: serenity-skill's SKILL.md has no %s placeholder, so the executor
// drops the user's topic argument and sends the raw methodology to the
// LLM. This test documents the current behavior so a future fix is
// detectable.
func TestSerenitySkill_ArgsNotInjectedWhenNoPlaceholder(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	userSkillsDir := filepath.Join(home, ".config", "agent-tui", "skills")
	if _, err := os.Stat(filepath.Join(userSkillsDir, "serenity-skill", "SKILL.md")); err != nil {
		t.Skipf("serenity-skill not installed at %s", userSkillsDir)
	}

	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, userSkillsDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	chater := &mockChater{}
	exec := NewSkillExecutor(r, chater)

	topic := "A股AI半导体产业链"
	if _, err := exec.Execute(&ParsedCommand{Name: "serenity-skill", Args: topic}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if chater.lastMessage == "" {
		t.Fatal("expected LLM to receive a prompt")
	}
	// Documents current (flawed) behavior: the topic is NOT in the prompt
	// because serenity-skill has no %s placeholder.
	if strings.Contains(chater.lastMessage, topic) {
		t.Fatalf("BUG-FIXED: topic %q now appears in prompt. Update this test to assert the fix.", topic)
	}
	t.Logf("current behavior: topic %q dropped from prompt (no %%s placeholder in SKILL.md)", topic)
	t.Logf("prompt length sent to LLM: %d bytes", len(chater.lastMessage))
}

// TestSerenitySkill_PromptContainsMethodology verifies the prompt sent
// to the LLM carries the serenity research methodology, so even without
// topic injection the LLM receives usable instructions.
func TestSerenitySkill_PromptContainsMethodology(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	userSkillsDir := filepath.Join(home, ".config", "agent-tui", "skills")
	if _, err := os.Stat(filepath.Join(userSkillsDir, "serenity-skill", "SKILL.md")); err != nil {
		t.Skipf("serenity-skill not installed at %s", userSkillsDir)
	}

	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, userSkillsDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	chater := &mockChater{}
	exec := NewSkillExecutor(r, chater)

	if _, err := exec.Execute(&ParsedCommand{Name: "serenity-skill", Args: "robotics"}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	for _, want := range []string{"scarce layer", "supply-chain", "value chain"} {
		if !strings.Contains(strings.ToLower(chater.lastMessage), want) {
			t.Errorf("expected prompt to contain %q, got (first 200): %.200s", want, chater.lastMessage)
		}
	}
	if !strings.HasPrefix(chater.lastSessionID, "skill-serenity-skill") {
		t.Errorf("expected session ID prefixed with skill-serenity-skill, got %q", chater.lastSessionID)
	}
}
