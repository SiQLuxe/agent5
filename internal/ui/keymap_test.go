package ui

import (
	"testing"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()
	if km.Quit == 0 {
		t.Fatal("expected non-zero quit rune")
	}
	if km.SendMessage == "" {
		t.Fatal("expected non-empty SendMessage")
	}
}

func TestShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.ShortHelp()
	if len(help) == 0 {
		t.Fatal("expected non-empty short help")
	}
}

func TestFullHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.FullHelp()
	if len(help) == 0 {
		t.Fatal("expected non-empty full help")
	}
}
