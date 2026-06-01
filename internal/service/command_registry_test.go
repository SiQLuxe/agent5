package service

import (
	"testing"
)

func TestCommandRegistryRegisterAndGet(t *testing.T) {
	r := NewCommandRegistry()
	r.Register(&Command{Name: "test-cmd", Description: "test"})
	cmd := r.Get("test-cmd")
	if cmd == nil {
		t.Fatal("expected command")
	}
	if cmd.Name != "test-cmd" {
		t.Fatalf("name: %q", cmd.Name)
	}
}

func TestCommandRegistryList(t *testing.T) {
	r := NewCommandRegistry()
	r.Register(&Command{Name: "a", Description: "a"})
	r.Register(&Command{Name: "b", Description: "b"})
	cmds := r.List()
	if len(cmds) != 2 {
		t.Fatalf("expected 2, got %d", len(cmds))
	}
}

func TestCommandRegistryListByCategory(t *testing.T) {
	r := NewCommandRegistry()
	r.Register(&Command{Name: "builtin1", Category: CmdBuiltin})
	r.Register(&Command{Name: "skill1", Category: CmdSkill})
	r.Register(&Command{Name: "skill2", Category: CmdSkill})

	builtins := r.ListByCategory(CmdBuiltin)
	if len(builtins) != 1 {
		t.Fatalf("expected 1 builtin, got %d", len(builtins))
	}

	skills := r.ListByCategory(CmdSkill)
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
}

func TestCommandRegistryGetNotFound(t *testing.T) {
	r := NewCommandRegistry()
	cmd := r.Get("nonexistent")
	if cmd != nil {
		t.Fatal("expected nil")
	}
}

func TestCommandRegistrySyncSkills(t *testing.T) {
	cr := NewCommandRegistry()
	sr := NewSkillRegistry()
	sr.Register(&Skill{Name: "review", Description: "review code"})
	sr.Register(&Skill{Name: "analyze", Description: "analyze project"})

	cr.SyncSkills(sr)

	skills := cr.ListByCategory(CmdSkill)
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	review := cr.Get("review")
	if review == nil {
		t.Fatal("review not found")
	}
	if review.Category != CmdSkill {
		t.Fatalf("expected CmdSkill")
	}
}

func TestCommandRegistryActionCalled(t *testing.T) {
	r := NewCommandRegistry()
	called := false
	r.Register(&Command{
		Name: "test-action",
		Action: func(args string) {
			called = true
		},
	})
	cmd := r.Get("test-action")
	if cmd == nil {
		t.Fatal("cmd not found")
	}
	cmd.Action("")
	if !called {
		t.Fatal("action not called")
	}
}
