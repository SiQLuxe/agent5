package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

type ChatPanel struct {
	session       *Session
	viewport      viewport.Model
	width         int
	height        int
	colors        ChatPanelColors
	currentTheme  ColorPalette
	searchMode    bool
	searchQuery   string
	searchIdx     int
	searchMatches []int
}

type ChatPanelColors struct {
	Background string
	TextMuted  string
}

func DefaultChatPanelColors() ChatPanelColors {
	return ChatPanelColors{
		Background: "#1e1e1e",
		TextMuted:  "#858585",
	}
}

func NewChatPanel(session *Session) *ChatPanel {
	vp := viewport.New()
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 3

	cp := &ChatPanel{
		session:  session,
		viewport: vp,
		width:    80,
		height:   24,
		colors:   DefaultChatPanelColors(),
	}
	cp.ApplyTheme(DefaultThemes[0].Colors)
	return cp
}

func (cp *ChatPanel) SetSession(s *Session) {
	cp.session = s
	cp.refreshContent()
}

func (cp *ChatPanel) Session() *Session {
	return cp.session
}

func (cp *ChatPanel) SetColors(c ChatPanelColors) {
	cp.colors = c
}

func (cp *ChatPanel) SetSize(width, height int) {
	cp.width = width
	cp.height = height
	cp.viewport.SetWidth(width)
	cp.viewport.SetHeight(height)
}

func (cp *ChatPanel) ScrollUp(lines int) {
	cp.viewport.ScrollUp(lines)
}

func (cp *ChatPanel) ScrollDown(lines int) {
	cp.viewport.ScrollDown(lines)
}

func (cp *ChatPanel) ScrollToBottom() {
	cp.viewport.GotoBottom()
}

func (cp *ChatPanel) ScrollToTop() {
	cp.viewport.GotoTop()
}

func (cp *ChatPanel) View() string {
	cp.refreshContent()
	result := cp.viewport.View()
	style := lipgloss.NewStyle().
		Width(cp.width).
		Height(cp.height).
		Background(lipgloss.Color(cp.colors.Background))
	return style.Render(result)
}

func (cp *ChatPanel) refreshContent() {
	if cp.session == nil || len(cp.session.Messages) == 0 {
		cp.viewport.SetContent("")
		return
	}
	content := cp.session.RenderMessages(cp.width, cp.currentTheme)
	cp.viewport.SetContent(content)
}

func (cp *ChatPanel) ApplyTheme(theme ColorPalette) {
	cp.currentTheme = theme
	cp.viewport.Style = lipgloss.NewStyle().
		Background(lipgloss.Color(theme.Background))
}

// Search methods
func (cp *ChatPanel) EnterSearch() {
	cp.searchMode = true
	cp.searchQuery = ""
	cp.searchIdx = -1
}

func (cp *ChatPanel) ExitSearch() {
	cp.searchMode = false
	cp.searchQuery = ""
	cp.searchMatches = nil
	cp.searchIdx = -1
	cp.viewport.ClearHighlights()
}

func (cp *ChatPanel) IsSearchMode() bool {
	return cp.searchMode
}

func (cp *ChatPanel) SetSearchQuery(q string) {
	cp.searchQuery = q
	cp.searchMatches = nil
	cp.searchIdx = -1
	cp.viewport.ClearHighlights()
	if q == "" || cp.session == nil {
		return
	}

	lower := strings.ToLower(q)
	var matchLines [][]int
	lines := strings.Split(cp.session.RenderMessages(cp.width, cp.currentTheme), "\n")
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), lower) {
			matchLines = append(matchLines, []int{i, i})
			cp.searchMatches = append(cp.searchMatches, i)
		}
	}

	if len(matchLines) > 0 {
		cp.searchIdx = 0
		cp.viewport.SetHighlights(matchLines)
	}
}

func (cp *ChatPanel) NextMatch() {
	if cp.searchIdx < 0 {
		return
	}
	cp.viewport.HighlightNext()
}

func (cp *ChatPanel) PrevMatch() {
	if cp.searchIdx < 0 {
		return
	}
	cp.viewport.HighlightPrevious()
}

func (cp *ChatPanel) Update(msg tea.Msg) {
	cp.viewport.Update(msg)
}
