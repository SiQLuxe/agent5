package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ChatPanel struct {
	*tview.TextView
	session       *Session
	searchQuery   string
	searchResults []int
	currentMatch  int
}

func NewChatPanel() *ChatPanel {
	c := &ChatPanel{
		TextView: tview.NewTextView(),
	}
	c.SetDynamicColors(true)
	c.SetScrollable(true)
	c.SetWordWrap(true)
	c.SetRegions(true)
	return c
}

func (c *ChatPanel) SetSession(session *Session) {
	c.session = session
	c.refresh()
}

func (c *ChatPanel) refresh() {
	if c.session == nil {
		c.SetText("")
		return
	}
	c.SetText(tview.TranslateANSI(c.session.RenderMessages(80, DefaultThemes[0].Colors)))
	c.ScrollToEnd()
}

func (c *ChatPanel) ScrollUp(lines int) {
	row, _ := c.GetScrollOffset()
	c.ScrollTo(row-lines, 0)
}

func (c *ChatPanel) ScrollDown(lines int) {
	row, _ := c.GetScrollOffset()
	c.ScrollTo(row+lines, 0)
}

func (c *ChatPanel) ScrollToTop() {
	c.ScrollTo(0, 0)
}

func (c *ChatPanel) ScrollToBottom() {
	c.ScrollToEnd()
}

func (c *ChatPanel) EnterSearch() {
	c.searchQuery = ""
	c.searchResults = nil
	c.currentMatch = -1
}

func (c *ChatPanel) ExitSearch() {
	c.searchQuery = ""
	c.searchResults = nil
	c.currentMatch = -1
	c.clearHighlights()
}

func (c *ChatPanel) IsSearchMode() bool {
	return false // tracked by App
}

func (c *ChatPanel) SetSearchQuery(query string) {
	c.searchQuery = query
	c.findMatches()
	if len(c.searchResults) > 0 {
		c.currentMatch = 0
		c.Highlight(c.matchRegionID(c.searchResults[0]))
		c.ScrollTo(c.searchResults[0], 0)
	}
}

func (c *ChatPanel) NextMatch() {
	if len(c.searchResults) == 0 {
		return
	}
	c.currentMatch = (c.currentMatch + 1) % len(c.searchResults)
	c.Highlight()
	c.Highlight(c.matchRegionID(c.searchResults[c.currentMatch]))
	c.ScrollTo(c.searchResults[c.currentMatch], 0)
}

func (c *ChatPanel) PrevMatch() {
	if len(c.searchResults) == 0 {
		return
	}
	c.currentMatch--
	if c.currentMatch < 0 {
		c.currentMatch = len(c.searchResults) - 1
	}
	c.Highlight()
	c.Highlight(c.matchRegionID(c.searchResults[c.currentMatch]))
	c.ScrollTo(c.searchResults[c.currentMatch], 0)
}

func (c *ChatPanel) MatchCount() int {
	return len(c.searchResults)
}

func (c *ChatPanel) CurrentMatch() int {
	return c.currentMatch
}

func (c *ChatPanel) ApplyTheme(colors ColorPalette) {
	c.SetBackgroundColor(hexToTCell(colors.Background))
}

func (c *ChatPanel) clearHighlights() {
	c.Highlight()
}

func (c *ChatPanel) findMatches() {
	c.searchResults = nil
	if c.searchQuery == "" || c.session == nil {
		return
	}
	query := strings.ToLower(c.searchQuery)
	content := strings.ToLower(c.GetText(false))
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, query) {
			c.searchResults = append(c.searchResults, i)
		}
	}
}

func (c *ChatPanel) matchRegionID(line int) string {
	return fmt.Sprintf("match-%d", line)
}

func hexToTCell(hex string) tcell.Color {
	if len(hex) == 0 {
		return tcell.ColorDefault
	}
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return tcell.ColorDefault
	}
	r := parseHexPair(hex[0:2])
	g := parseHexPair(hex[2:4])
	b := parseHexPair(hex[4:6])
	return tcell.NewHexColor(int32(r)<<16 | int32(g)<<8 | int32(b))
}

func parseHexPair(s string) int32 {
	if len(s) != 2 {
		return 0
	}
	return hexNibble(s[0])<<4 | hexNibble(s[1])
}

func hexNibble(c byte) int32 {
	switch {
	case c >= '0' && c <= '9':
		return int32(c - '0')
	case c >= 'a' && c <= 'f':
		return int32(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int32(c - 'A' + 10)
	default:
		return 0
	}
}
