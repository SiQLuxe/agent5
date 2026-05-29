package ui

type KeyMap struct {
	Quit           rune
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
		Quit:           'q',
		NewSession:     'n',
		CloseSession:   'w',
		RenameSession:  'r',
		NextSession:    '.',
		PrevSession:    ',',
		ToggleThinking: 't',
		ToggleCollapse: 'y',
		Search:         '/',
		ToggleTheme:    'T',
		ShowHelp:       '?',
		ScrollUp:       "pgup",
		ScrollDown:     "pgdn",
		ScrollTop:      'g',
		ScrollBottom:   'G',
		SendMessage:    "ctrl+enter",
	}
}

func (k KeyMap) ShortHelp() []string {
	return []string{
		"Ctrl+Enter: Send",
		"Alt+N: New",
		"Alt+W: Close",
		"?: Help",
	}
}

func (k KeyMap) FullHelp() []string {
	return []string{
		"Ctrl+Enter   Send message",
		"Alt+N        New session",
		"Alt+W        Close session",
		"Alt+R        Rename session",
		"Alt+.        Next session",
		"Alt+,        Previous session",
		"Alt+T        Toggle thinking",
		"Alt+Y        Toggle collapse",
		"/            Search",
		"Alt+Shift+T  Toggle theme",
		"PgUp/PgDn    Scroll chat",
		"g/G          Scroll top/bottom",
		"q            Quit",
	}
}
