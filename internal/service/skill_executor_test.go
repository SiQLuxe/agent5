package service

import "testing"

func TestExecutePromptSkill(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "review", Description: "review", Type: SkillPrompt, Prompt: "Review this: %s"})
	ex := NewSkillExecutor(r, nil)
	result, err := ex.Execute(&ParsedCommand{Name: "review", Args: "main.go"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if result != "Review this: main.go" {
		t.Fatalf("got %q", result)
	}
}

func TestExecuteHandlerSkill(t *testing.T) {
	r := NewSkillRegistry()
	r.RegisterHandler("ping", func(ctx SkillContext) string {
		return "pong:" + ctx.Input
	})
	r.Register(&Skill{Name: "ping", Description: "ping test"})
	ex := NewSkillExecutor(r, nil)
	result, err := ex.Execute(&ParsedCommand{Name: "ping", Args: "hello"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if result != "pong:hello" {
		t.Fatalf("got %q", result)
	}
}

func TestExecuteNotFound(t *testing.T) {
	r := NewSkillRegistry()
	ex := NewSkillExecutor(r, nil)
	_, err := ex.Execute(&ParsedCommand{Name: "nope"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecutePromptWithoutArgs(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "hello", Description: "hello", Type: SkillPrompt, Prompt: "Say hello"})
	ex := NewSkillExecutor(r, nil)
	result, err := ex.Execute(&ParsedCommand{Name: "hello"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if result != "Say hello" {
		t.Fatalf("got %q", result)
	}
}

func TestExecutePromptArgNoPercentS(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "greet", Description: "greet", Type: SkillPrompt, Prompt: "Hello there"})
	ex := NewSkillExecutor(r, nil)
	result, err := ex.Execute(&ParsedCommand{Name: "greet", Args: "world"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if result != "Hello there" {
		t.Fatalf("got %q", result)
	}
}

func TestExecuteHandlerNoHandler(t *testing.T) {
	r := NewSkillRegistry()
	r.Register(&Skill{Name: "orphan", Description: "no handler", Type: SkillHandler})
	ex := NewSkillExecutor(r, nil)
	_, err := ex.Execute(&ParsedCommand{Name: "orphan"})
	if err == nil {
		t.Fatal("expected error for handler type with no registered handler")
	}
}


