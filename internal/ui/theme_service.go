package ui

import (
	"fmt"
	"strings"
	"github.com/example/agent-tui/internal/agent/session"
)

type ThemeService struct {
	sessionMgr *session.Manager
	currentTheme Theme
	themes []Theme
}

func NewThemeService(sm *session.Manager) *ThemeService {
	return &ThemeService{
		sessionMgr: sm,
		currentTheme: DefaultThemes[0],
		themes: DefaultThemes,
	}
}

func (ts *ThemeService) CurrentTheme() Theme {
	return ts.currentTheme
}

func (ts *ThemeService) SetThemeByName(name string) error {
	for _, theme := range ts.themes {
		if strings.EqualFold(theme.Name, name) {
			ts.currentTheme = theme
			return nil
		}
	}
	return fmt.Errorf("theme not found: %s", name)
}

func (ts *ThemeService) NextTheme() {
	currentIndex := -1
	for i, theme := range ts.themes {
		if theme.Name == ts.currentTheme.Name {
			currentIndex = i
			break
		}
	}
	
	if currentIndex >= 0 {
		nextIndex := (currentIndex + 1) % len(ts.themes)
		ts.currentTheme = ts.themes[nextIndex]
	}
}

func (ts *ThemeService) GenerateTheme(preferences string) (*Theme, error) {
	if ts.sessionMgr == nil {
		return nil, fmt.Errorf("session manager not available")
	}
	
	prompt := fmt.Sprintf(`Create a terminal UI color theme with these preferences: %s
Return a JSON with these hex colors:
{
  "Name": "theme name",
  "Description": "brief description",
  "Colors": {
    "Background": "#XXXXXX",
    "PanelBg": "#XXXXXX",
    "Text": "#XXXXXX",
    "TextMuted": "#XXXXXX",
    "UserFg": "#XXXXXX",
    "UserBg": "#XXXXXX",
    "AssistantFg": "#XXXXXX",
    "AssistantBg": "#XXXXXX",
    "SystemFg": "#XXXXXX",
    "SystemBg": "#XXXXXX",
    "Border": "#XXXXXX",
    "Accent": "#XXXXXX",
    "Loading": "#XXXXXX"
  }
}
Use dark theme as base. Ensure good contrast.`, preferences)
	
	_, err := ts.sessionMgr.Chat("theme-generator", prompt)
	if err != nil {
		return nil, err
	}
	
	theme := DefaultThemes[0] // Fallback
	ts.themes = append(ts.themes, theme)
	return &theme, nil
}

func (ts *ThemeService) EvaluateTheme(theme Theme) (score int, suggestions []string, err error) {
	if ts.sessionMgr == nil {
		return 0, []string{}, fmt.Errorf("session manager not available")
	}
	
	prompt := fmt.Sprintf(`Evaluate this terminal UI color theme (score 0-100):
Name: %s
Description: %s
Colors: %#v

Return JSON: {"score": 0-100, "suggestions": ["..."]}`, theme.Name, theme.Description, theme.Colors)
	
	_, err = ts.sessionMgr.Chat("theme-evaluator", prompt)
	if err != nil {
		return 75, []string{"No AI feedback available"}, nil
	}
	
	return 75, []string{}, nil // Default fallback
}
