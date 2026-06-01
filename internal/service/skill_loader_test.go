package service

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestParseFrontmatterBasic(t *testing.T) {
	input := "---\nname: my-skill\ndescription: My skill\n---\nBody text"
	name, desc, body, err := parseFrontmatter(input)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if name != "my-skill" {
		t.Fatalf("name: %q", name)
	}
	if desc != "My skill" {
		t.Fatalf("desc: %q", desc)
	}
	if body != "Body text" {
		t.Fatalf("body: %q", body)
	}
}

func TestParseFrontmatterNoDelimiter(t *testing.T) {
	_, _, _, err := parseFrontmatter("no delimiters")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseFrontmatterEmptyName(t *testing.T) {
	_, _, _, err := parseFrontmatter("---\nname: \ndescription: test\n---\nbody")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestParseFrontmatterMultiLine(t *testing.T) {
	input := "---\nname: code-review\ndescription: Review code\n---\nLine 1\nLine 2\nLine 3"
	_, _, body, err := parseFrontmatter(input)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if body != "Line 1\nLine 2\nLine 3" {
		t.Fatalf("body: %q", body)
	}
}

func TestParseFrontmatterExtraFields(t *testing.T) {
	input := "---\nname: test\ndescription: A test\nlicense: MIT\nversion: 1.0\n---\nbody"
	name, desc, body, err := parseFrontmatter(input)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if name != "test" || desc != "A test" || body != "body" {
		t.Fatalf("got name=%q desc=%q body=%q", name, desc, body)
	}
}

func TestLoadSkillsFromDir(t *testing.T) {
	baseDir := t.TempDir()
	writeFixture(t, filepath.Join(baseDir, "code-review", "SKILL.md"),
		"---\nname: code-review\ndescription: Review code\n---\nYou are a reviewer.")

	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, baseDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	s, ok := r.Get("code-review")
	if !ok {
		t.Fatal("code-review not registered")
	}
	if s.Name != "code-review" || s.Description != "Review code" || s.Prompt != "You are a reviewer." {
		t.Fatalf("got %+v", s)
	}
	if s.Type != SkillPrompt {
		t.Fatalf("expected SkillPrompt")
	}
}

func TestLoadSkillsInvalidFrontmatter(t *testing.T) {
	baseDir := t.TempDir()
	writeFixture(t, filepath.Join(baseDir, "bad", "SKILL.md"), "no frontmatter here")
	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, baseDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	if _, ok := r.Get("bad"); ok {
		t.Fatal("bad skill should not register")
	}
}

func TestLoadSkillsMissingDir(t *testing.T) {
	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, "/nonexistent/path/xyz"); err != nil {
		t.Fatalf("missing dir should not error: %v", err)
	}
}

func TestLoadSkillsNameMismatch(t *testing.T) {
	baseDir := t.TempDir()
	writeFixture(t, filepath.Join(baseDir, "code-review", "SKILL.md"),
		"---\nname: review\ndescription: review\n---\nbody")
	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, baseDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	if _, ok := r.Get("review"); ok {
		t.Fatal("should not register when name != dir")
	}
}

func TestLoadSkillsNoSKILLMD(t *testing.T) {
	baseDir := t.TempDir()
	os.MkdirAll(filepath.Join(baseDir, "empty"), 0755)
	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, baseDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
}

func TestLoadSkillsDirRecursive(t *testing.T) {
	baseDir := t.TempDir()
	writeFixture(t, filepath.Join(baseDir, "code-review", "SKILL.md"),
		"---\nname: code-review\ndescription: Review\n---\nbody")
	writeFixture(t, filepath.Join(baseDir, "debug", "test", "SKILL.md"),
		"---\nname: test\ndescription: Test\n---\nbody")
	writeFixture(t, filepath.Join(baseDir, "debug", "SKILL.md"),
		"---\nname: debug\ndescription: Debug\n---\nbody")

	r := NewSkillRegistry()
	if err := LoadSkillsDir(r, baseDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	if _, ok := r.Get("code-review"); !ok {
		t.Fatal("code-review not found")
	}
	if _, ok := r.Get("debug"); !ok {
		t.Fatal("debug not found")
	}
	if _, ok := r.Get("test"); !ok {
		t.Fatal("nested test not found")
	}
}

func TestLoadSkillsAutoDetectHandler(t *testing.T) {
	baseDir := t.TempDir()
	writeFixture(t, filepath.Join(baseDir, "ping", "SKILL.md"),
		"---\nname: ping\ndescription: Ping test\n---\nping body")
	r := NewSkillRegistry()
	r.RegisterHandler("ping", func(ctx SkillContext) string { return "pong" })
	if err := LoadSkillsDir(r, baseDir); err != nil {
		t.Fatalf("LoadSkillsDir: %v", err)
	}
	s, ok := r.Get("ping")
	if !ok {
		t.Fatal("ping not found")
	}
	if s.Type != SkillHandler {
		t.Fatalf("expected SkillHandler, got %v", s.Type)
	}
}
