package service

import "testing"

func TestParseCommandBasic(t *testing.T) {
	cmd := ParseCommand("/review")
	if cmd == nil {
		t.Fatal("nil result")
	}
	if cmd.Name != "review" {
		t.Fatalf("name: got %q", cmd.Name)
	}
	if cmd.Args != "" {
		t.Fatalf("args: got %q", cmd.Args)
	}
}

func TestParseCommandWithArgs(t *testing.T) {
	cmd := ParseCommand("/review main.go")
	if cmd == nil {
		t.Fatal("nil")
	}
	if cmd.Name != "review" {
		t.Fatalf("name: %q", cmd.Name)
	}
	if cmd.Args != "main.go" {
		t.Fatalf("args: %q", cmd.Args)
	}
}

func TestParseCommandMultiWord(t *testing.T) {
	cmd := ParseCommand("/exec go run main.go")
	if cmd.Name != "exec" {
		t.Fatalf("name: %q", cmd.Name)
	}
	if cmd.Args != "go run main.go" {
		t.Fatalf("args: %q", cmd.Args)
	}
}

func TestParseCommandNoSlash(t *testing.T) {
	if cmd := ParseCommand("hello"); cmd != nil {
		t.Fatal("expected nil")
	}
}

func TestParseCommandEmpty(t *testing.T) {
	if cmd := ParseCommand(""); cmd != nil {
		t.Fatal("expected nil")
	}
}

func TestParseCommandJustSlash(t *testing.T) {
	if cmd := ParseCommand("/"); cmd != nil {
		t.Fatal("expected nil")
	}
}

func TestParseCommandSpaces(t *testing.T) {
	cmd := ParseCommand("  /test  arg1  arg2  ")
	if cmd == nil {
		t.Fatal("nil")
	}
	if cmd.Name != "test" {
		t.Fatalf("name: %q", cmd.Name)
	}
	if cmd.Args != "arg1  arg2" {
		t.Fatalf("args: %q", cmd.Args)
	}
}
