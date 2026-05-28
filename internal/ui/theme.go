package ui

type ColorPalette struct {
	Background  string
	PanelBg     string
	Text        string
	TextMuted   string
	Border      string
	Accent      string
	Success     string
	Warning     string
	Error       string
	UserFg      string
	UserBg      string
	AssistantFg string
	AssistantBg string
	SystemFg    string
	SystemBg    string
	Loading     string
	// Tab Dock
	TabActiveBg   string
	TabInactiveFg string
	TabSeparator  string
	TabNewButton  string
	// Thinking
	ThinkingFg     string
	ThinkingBg     string
	ThinkingBorder string
	// Input
	InputBg        string
	InputPrompt    string
	InputSeparator string
	// Timestamp
	Timestamp string
}

type Theme struct {
	Name        string
	Description string
	Colors      ColorPalette
}

var DefaultThemes = []Theme{
	{
		Name:        "Dark",
		Description: "VS Code dark theme",
		Colors: ColorPalette{
			Background:  "#1e1e1e",
			PanelBg:     "#252526",
			Text:        "#d4d4d4",
			TextMuted:   "#858585",
			Border:      "#3c3c3c",
			Accent:      "#569cd6",
			Success:     "#4ec9b0",
			Warning:     "#dcdcaa",
			Error:       "#f14c4c",
			UserFg:      "#ffffff",
			UserBg:      "#0066cc",
			AssistantFg: "#ffffff",
			AssistantBg: "#28a745",
			SystemFg:    "#ffffff",
			SystemBg:    "#9370db",
			Loading:     "#ffa500",
			// Tab Dock
			TabActiveBg:   "#569cd6",
			TabInactiveFg: "#858585",
			TabSeparator:  "#444444",
			TabNewButton:  "#4ec9b0",
			// Thinking
			ThinkingFg:     "#dcdcaa",
			ThinkingBg:     "#2a2a1a",
			ThinkingBorder: "#dcdcaa",
			// Input
			InputBg:        "#1a1a1a",
			InputPrompt:    "#4ec9b0",
			InputSeparator: "#3c3c3c",
			// Timestamp
			Timestamp: "#555555",
		},
	},
}
