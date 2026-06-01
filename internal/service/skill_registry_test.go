package service

import (
	"os"
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
	if err == nil {
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

func TestSkillRegistryDuplicateErrorIsErrExist(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "dup2", Description: "dup2"})
	err := r.Register(&Skill{Name: "dup2", Description: "dup2"})
	if !os.IsExist(err) {
		t.Fatal("expected os.ErrExist for duplicate")
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

func TestUnregister(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "x", Description: "x"})
	r.Unregister("x")
	_, ok := r.Get("x")
	if ok {
		t.Fatal("expected not found after unregister")
	}
}
