package ui

import (
	"testing"

	"charm.land/bubbles/v2/key"
)

func TestKeyMap_RegisteredKeys(t *testing.T) {
	km := DefaultKeyMap

	cases := []struct {
		name string
		b    key.Binding
		keys []string
		help string
	}{
		{"Quit", km.Quit, []string{"ctrl+c"}, "quit"},
		{"NewSession", km.NewSession, []string{"ctrl+n"}, "new session"},
		{"CloseSession", km.CloseSession, []string{"ctrl+q"}, "close session"},
		{"RenameSession", km.RenameSession, []string{"ctrl+e"}, "rename session"},
		{"NextSession", km.NextSession, []string{"alt+n", "alt+right"}, "next session"},
		{"PrevSession", km.PrevSession, []string{"alt+p", "alt+left"}, "prev session"},
		{"ToggleThinking", km.ToggleThinking, []string{"ctrl+y"}, "toggle thinking"},
		{"ToggleCollapse", km.ToggleCollapse, []string{"ctrl+l"}, "toggle collapse"},
		{"Search", km.Search, []string{"ctrl+f"}, "search"},
		{"ToggleTheme", km.ToggleTheme, []string{"ctrl+t"}, "toggle theme"},
		{"ShowHelp", km.ShowHelp, []string{"ctrl+g"}, "show help"},
		{"ScrollUp", km.ScrollUp, []string{"pgup", "ctrl+up"}, "scroll up"},
		{"ScrollDown", km.ScrollDown, []string{"pgdown", "ctrl+down"}, "scroll down"},
		{"ScrollTop", km.ScrollTop, []string{"ctrl+home"}, "scroll to top"},
		{"ScrollBottom", km.ScrollBottom, []string{"ctrl+end"}, "scroll to bottom"},
		{"SendMessage", km.SendMessage, []string{"ctrl+enter"}, "send message"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.b.Help().Key == "" {
				t.Error("binding has empty Help().Key")
			}
			if c.b.Help().Desc != c.help {
				t.Errorf("expected help desc %q, got %q", c.help, c.b.Help().Desc)
			}
		})
	}
}

func TestKeyMap_ShortHelp(t *testing.T) {
	km := DefaultKeyMap
	short := km.ShortHelp()
	if len(short) == 0 {
		t.Fatal("ShortHelp returned empty")
	}
}

func TestKeyMap_FullHelp(t *testing.T) {
	km := DefaultKeyMap
	full := km.FullHelp()
	if len(full) == 0 {
		t.Fatal("FullHelp returned empty")
	}
}
