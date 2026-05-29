package ui

import (
	"charm.land/bubbles/v2/key"
)

type KeyMap struct {
	Quit           key.Binding
	NewSession     key.Binding
	CloseSession   key.Binding
	RenameSession  key.Binding
	NextSession    key.Binding
	PrevSession    key.Binding
	ToggleThinking key.Binding
	ToggleCollapse key.Binding
	Search         key.Binding
	ToggleTheme    key.Binding
	ShowHelp       key.Binding
	ScrollUp       key.Binding
	ScrollDown     key.Binding
	ScrollTop      key.Binding
	ScrollBottom   key.Binding
	SendMessage    key.Binding
}

var DefaultKeyMap = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("Ctrl+C", "quit"),
	),
	NewSession: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("Ctrl+N", "new session"),
	),
	CloseSession: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("Ctrl+Q", "close session"),
	),
	RenameSession: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("Ctrl+E", "rename session"),
	),
	NextSession: key.NewBinding(
		key.WithKeys("alt+n", "alt+right"),
		key.WithHelp("Alt+N/→", "next session"),
	),
	PrevSession: key.NewBinding(
		key.WithKeys("alt+p", "alt+left"),
		key.WithHelp("Alt+P/←", "prev session"),
	),
	ToggleThinking: key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("Ctrl+Y", "toggle thinking"),
	),
	ToggleCollapse: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("Ctrl+L", "toggle collapse"),
	),
	Search: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("Ctrl+F", "search"),
	),
	ToggleTheme: key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("Ctrl+T", "toggle theme"),
	),
	ShowHelp: key.NewBinding(
		key.WithKeys("ctrl+g"),
		key.WithHelp("Ctrl+G", "show help"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("pgup", "ctrl+up"),
		key.WithHelp("PgUp/Ctrl+↑", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("pgdown", "ctrl+down"),
		key.WithHelp("PgDn/Ctrl+↓", "scroll down"),
	),
	ScrollTop: key.NewBinding(
		key.WithKeys("ctrl+home"),
		key.WithHelp("Ctrl+Home", "scroll to top"),
	),
	ScrollBottom: key.NewBinding(
		key.WithKeys("ctrl+end"),
		key.WithHelp("Ctrl+End", "scroll to bottom"),
	),
	SendMessage: key.NewBinding(
		key.WithKeys("ctrl+enter"),
		key.WithHelp("Ctrl+Enter", "send message"),
	),
}

func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		km.NewSession,
		km.NextSession,
		km.PrevSession,
		km.ToggleThinking,
		km.Search,
		km.ToggleTheme,
		km.ShowHelp,
	}
}

func (km KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{km.NewSession, km.CloseSession, km.RenameSession, km.NextSession, km.PrevSession},
		{km.ToggleThinking, km.ToggleCollapse, km.Search, km.ToggleTheme, km.ShowHelp},
		{km.ScrollUp, km.ScrollDown, km.ScrollTop, km.ScrollBottom},
		{km.SendMessage, km.Quit},
	}
}
