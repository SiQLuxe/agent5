package ui

type KeyMap struct {
	NewSession     rune
	CloseSession   rune
	RenameSession  rune
	NextSession    rune
	PrevSession    rune
	ToggleThinking rune
	ToggleCollapse rune
	Search         rune
	ToggleTheme    rune
	ShowHelp       rune
	ScrollUp       string
	ScrollDown     string
	ScrollTop      rune
	ScrollBottom   rune
	SendMessage    string
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		NewSession:     'n',
		CloseSession:   'w',
		RenameSession:  'r',
		NextSession:    '.',
		PrevSession:    ',',
		ToggleThinking: 't',
		ToggleCollapse: 'y',
		Search:         '/',
		ToggleTheme:    'T',
		ShowHelp:       0, // F1 handled via tcell.KeyF1
		ScrollUp:       "pgup",
		ScrollDown:     "pgdn",
		ScrollTop:      'g',
		ScrollBottom:   'G',
		SendMessage:    "ctrl+enter",
	}
}

func (k KeyMap) ShortHelp() []string {
	return []string{
		"Enter: Send",
		"Ctrl+N/Alt+N: New",
		"Ctrl+W/Alt+W: Close",
		"Tab/Alt+.: Switch",
	}
}

func (k KeyMap) FullHelp() []string {
	return []string{
		"Enter         Send message",
		"Ctrl+N/Alt+N  New session",
		"Ctrl+W/Alt+W  Close session",
		"Ctrl+R/Alt+R  Rename session",
		"Tab/Alt+.     Next session",
		"Shift+Tab     Previous session (Alt+,)",
		"Ctrl+T/Alt+T  Toggle thinking",
		"Ctrl+Y/Alt+Y  Toggle collapse",
		"Ctrl+F        Search chat",
		"Ctrl+K/Alt+S  Toggle theme",
		"Ctrl+O        Help",
		"PgUp/PgDn     Scroll chat",
		"g/G           Scroll top/bottom",
		"Ctrl+C        Quit",
	}
}
