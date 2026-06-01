package service

import (
	"testing"
)

func TestSkillRegistryRegisterAndGet(t *testing.T) {
	r := NewSkillRegistry()
	skill := &Skill{Name: "test-skill", Description: "a test"}
	err := r.Register(skill)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	got, ok := r.Get("test-skill")
	if !ok {
		t.Fatal("Get returned not found")
	}
	if got.Name != "test-skill" {
		t.Fatalf("expected name test-skill, got %s", got.Name)
	}
}

func TestSkillRegistryList(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "a", Description: "skill a"})
	r.Register(&Skill{Name: "b", Description: "skill b"})
	skills := r.List()
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
}

func TestSkillRegistryDuplicate(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "dup", Description: "dup"})
	err := r.Register(&Skill{Name: "dup", Description: "dup"})
	if err != ErrSkillAlreadyExists {
		t.Fatal("expected error on duplicate register")
	}
}

func TestSkillRegistryGetNotFound(t *testing.T) {
	r := NewSkillRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("expected false for nonexistent skill")
	}
}

func TestSkillRegistryDuplicateError(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "dup2", Description: "dup2"})
	err := r.Register(&Skill{Name: "dup2", Description: "dup2"})
	if err != ErrSkillAlreadyExists {
		t.Fatal("expected ErrSkillAlreadyExists for duplicate")
	}
}

func TestRegisterNilSkill(t *testing.T) {
	r := NewSkillRegistry()
	err := r.Register(nil)
	if err == nil {
		t.Fatal("expected error for nil skill")
	}
}

func TestRegisterEmptyName(t *testing.T) {
	r := NewSkillRegistry()
	err := r.Register(&Skill{Name: "", Description: "empty"})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegisterDoesNotMutateCallerSkill(t *testing.T) {
	r := NewSkillRegistry()
	r.RegisterHandler("mutation-test", func(ctx SkillContext) string { return "ok" })
	skill := &Skill{Name: "mutation-test", Description: "test", Type: SkillPrompt}
	originalType := skill.Type
	r.Register(skill)
	if skill.Type != originalType {
		t.Fatal("Register mutated the caller's Skill.Type")
	}
	// But the stored skill should have SkillHandler
	got, _ := r.Get("mutation-test")
	if got.Type != SkillHandler {
		t.Fatal("stored skill should have SkillHandler type")
	}
}

func TestHandlerRegistrationReverseOrder(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "rev-test", Description: "reverse test"})
	r.RegisterHandler("rev-test", func(ctx SkillContext) string { return "ok" })
	got, ok := r.Get("rev-test")
	if !ok {
		t.Fatal("skill not found")
	}
	if got.Type != SkillHandler {
		t.Fatalf("expected SkillHandler, got %v", got.Type)
	}
}

func TestIsHandler(t *testing.T) {
	r := NewSkillRegistry()
	if r.IsHandler("nonexistent") {
		t.Fatal("expected false for nonexistent handler")
	}
	r.RegisterHandler("exists", func(ctx SkillContext) string { return "ok" })
	if !r.IsHandler("exists") {
		t.Fatal("expected true for existing handler")
	}
}

func TestHandlerRegistration(t *testing.T) {
	r := NewSkillRegistry()
	r.RegisterHandler("ping", func(ctx SkillContext) string {
		return "pong"
	})
	r.Register(&Skill{Name: "ping", Description: "ping test"})
	got, ok := r.Get("ping")
	if !ok {
		t.Fatal("skill ping not found")
	}
	if got.Type != SkillHandler {
		t.Fatalf("expected SkillHandler type, got %v", got.Type)
	}
	fn, ok := r.GetHandler("ping")
	if !ok {
		t.Fatal("handler not found")
	}
	result := fn(SkillContext{Input: "hello"})
	if result != "pong" {
		t.Fatalf("expected 'pong', got '%s'", result)
	}
}

func TestListContainsAllSkills(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "a", Description: "a"})
	r.Register(&Skill{Name: "b", Description: "b"})
	got := r.List()
	if len(got) != 2 {
		t.Fatalf("len: expected 2, got %d", len(got))
	}
	names := make(map[string]bool)
	for _, s := range got {
		names[s.Name] = true
	}
	if !names["a"] || !names["b"] {
		t.Fatal("List missing registered skills")
	}
}

func TestUnregister(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "x", Description: "x"})
	r.Unregister("x")
	_, ok := r.Get("x")
	if ok {
		t.Fatal("expected not found after unregister")
	}
}
