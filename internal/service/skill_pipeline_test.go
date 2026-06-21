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

// TestLoadSkillsDir_ProjectOverridesUser verifies the two-layer loading
// order: when the same skill name exists in both a project-level dir and a
// user-level dir, the one loaded first (project-level) wins.
func TestLoadSkillsDir_ProjectOverridesUser(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project", "skills", "shared")
	userDir := filepath.Join(root, "user", "skills", "shared")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeSkill(filepath.Join(projectDir, "SKILL.md"), "shared", "project version", "PROJECT BODY"); err != nil {
		t.Fatal(err)
	}
	if err := writeSkill(filepath.Join(userDir, "SKILL.md"), "shared", "user version", "USER BODY"); err != nil {
		t.Fatal(err)
	}

	r := NewSkillRegistry()
	// Load project first, then user — project wins.
	if err := LoadSkillsDir(r, filepath.Join(root, "project", "skills")); err != nil {
		t.Fatalf("LoadSkillsDir project: %v", err)
	}
	if err := LoadSkillsDir(r, filepath.Join(root, "user", "skills")); err != nil {
		t.Fatalf("LoadSkillsDir user: %v", err)
	}

	s, ok := r.Get("shared")
	if !ok {
		t.Fatal("expected shared skill registered")
	}
	if !strings.Contains(s.Prompt, "PROJECT BODY") {
		t.Fatalf("expected project body to win, got %q", s.Prompt)
	}
}

// TestLoadSkillsDir_UserOnlyLoadsWhenAbsentFromProject verifies the user
// layer is used when the project layer has no such skill.
func TestLoadSkillsDir_UserOnlyLoadsWhenAbsentFromProject(t *testing.T) {
	root := t.TempDir()
	projectDir := filepath.Join(root, "project", "skills")
	userDir := filepath.Join(root, "user", "skills", "personal")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeSkill(filepath.Join(userDir, "SKILL.md"), "personal", "user only", "USER ONLY BODY"); err != nil {
		t.Fatal(err)
	}

	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, projectDir); err != nil {
		t.Fatalf("LoadSkillsDir project: %v", err)
	}
	if err := LoadSkillsDir(r, filepath.Join(root, "user", "skills")); err != nil {
		t.Fatalf("LoadSkillsDir user: %v", err)
	}

	s, ok := r.Get("personal")
	if !ok {
		t.Fatal("expected personal skill from user layer")
	}
	if !strings.Contains(s.Prompt, "USER ONLY BODY") {
		t.Fatalf("expected user body, got %q", s.Prompt)
	}
}
