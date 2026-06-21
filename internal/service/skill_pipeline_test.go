package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadSkillsDir_LoadsCodeReview verifies the built-in code-review skill
// ships in the repo's skills/ directory and parses correctly.
func TestLoadSkillsDir_LoadsCodeReview(t *testing.T) {
	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, "../../skills"); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}

	s, ok := r.Get("code-review")
	if !ok {
		t.Fatal("expected code-review skill to be registered")
	}
	if !strings.Contains(strings.ToLower(s.Description), "review") {
		t.Fatalf("unexpected description: %q", s.Description)
	}
	if s.Type != SkillPrompt {
		t.Fatalf("expected SkillPrompt, got %v", s.Type)
	}
	if !strings.Contains(s.Prompt, "senior Go code reviewer") {
		t.Fatalf("unexpected prompt body: %q", s.Prompt)
	}
}

// TestLoadSkillsDir_LoadsEchoTest verifies the test-only echo-test skill
// (added to exercise the full pipeline) is discovered and parsed.
func TestLoadSkillsDir_LoadsEchoTest(t *testing.T) {
	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, "../../skills"); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}

	s, ok := r.Get("echo-test")
	if !ok {
		t.Fatal("expected echo-test skill to be registered")
	}
	if s.Type != SkillPrompt {
		t.Fatalf("expected SkillPrompt, got %v", s.Type)
	}
	if !strings.Contains(s.Prompt, "%s") {
		t.Fatalf("expected %%s placeholder in prompt, got %q", s.Prompt)
	}
}

// TestSkillExecutor_PromptSubstitution verifies the executor substitutes %s
// with the user argument and returns the rendered prompt when no LLM is wired.
func TestSkillExecutor_PromptSubstitution(t *testing.T) {
	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, "../../skills"); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	// No chater → executor returns the rendered prompt verbatim.
	exec := NewSkillExecutor(r, nil)

	out, err := exec.Execute(&ParsedCommand{Name: "echo-test", Args: "hello world"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "hello world") {
		t.Fatalf("expected args substituted into output, got %q", out)
	}
	if strings.Contains(out, "%s") {
		t.Fatalf("expected %%s placeholder to be replaced, got %q", out)
	}
}

// TestSkillExecutor_MissingSkill verifies error path for unknown skill.
func TestSkillExecutor_MissingSkill(t *testing.T) {
	r := NewSkillRegistry()
	exec := NewSkillExecutor(r, nil)

	if _, err := exec.Execute(&ParsedCommand{Name: "does-not-exist"}); err == nil {
		t.Fatal("expected error for unknown skill")
	}
}

// TestLoadSkillsDir_DirNameMustMatchName verifies the loader rejects a
// SKILL.md whose frontmatter name does not match its parent directory.
func TestLoadSkillsDir_DirNameMustMatchName(t *testing.T) {
	tmp := t.TempDir()
	// skills/wrong-dir/SKILL.md with name: other-name
	dir := tmp + "/wrong-dir"
	if err := mkdirAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := writeSkill(dir+"/SKILL.md", "other-name", "desc", "body"); err != nil {
		t.Fatal(err)
	}

	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, tmp); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	if _, ok := r.Get("other-name"); ok {
		t.Fatal("expected skill to be skipped when dir name != frontmatter name")
	}
	if _, ok := r.Get("wrong-dir"); ok {
		t.Fatal("expected skill to be skipped when dir name != frontmatter name")
	}
}

// TestLoadSkillsDir_MissingDirIsNotError verifies a non-existent skills dir
// is tolerated (returns nil, not an error) so the app app still boots.
func TestLoadSkillsDir_MissingDirIsNotError(t *testing.T) {
	r := NewSkillRegistry()
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	if err := LoadSkillsDir(r, missing); err != nil {
		t.Fatalf("expected nil for missing dir, got %v", err)
	}
}

// --- helpers ---

func mkdirAll(p string) error {
	return os.MkdirAll(p, 0o755)
}

func writeSkill(path, name, desc, body string) error {
	content := "---\nname: " + name + "\ndescription: " + desc + "\n---\n" + body + "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}
