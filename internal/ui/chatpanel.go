package ui

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/example/agent-tui/internal/ui/syntax"
)

type ChatPanelColors struct {
	Background     string
	UserFg         string
	AssistantBg    string
	AssistantFg    string
	SystemBg       string
	SystemFg       string
	Text           string
	TextMuted      string
	Timestamp      string
	ThinkingFg     string
	ThinkingBg     string
	ThinkingBorder string
}

func DefaultChatPanelColors() ChatPanelColors {
	return ChatPanelColors{
		Background:     "#1e1e1e",
		UserFg:         "#4ec9b0",
		AssistantBg:    "#28a745",
		AssistantFg:    "#ffffff",
		SystemBg:       "#9370db",
		SystemFg:       "#ffffff",
		Text:           "#d4d4d4",
		TextMuted:      "#858585",
		Timestamp:      "#555555",
		ThinkingFg:     "#dcdcaa",
		ThinkingBg:     "#2a2a1a",
		ThinkingBorder: "#dcdcaa",
	}
}

type ChatPanel struct {
	session        *Session
	width          int
	height         int
	colors         ChatPanelColors
	scrollOffset   int
	maxVisibleLines int // max lines before auto-collapse (0 = default 8)
	searchMode     bool
	searchQuery    string
	searchMatches  []int // message indices matching query
	searchIdx      int   // current match index
}

func NewChatPanel(session *Session) *ChatPanel {
	return &ChatPanel{
		session:         session,
		width:           80,
		height:          24,
		colors:          DefaultChatPanelColors(),
		scrollOffset:    0,
		maxVisibleLines: 8,
	}
}

func (cp *ChatPanel) SetSession(s *Session) {
	cp.session = s
}

func (cp *ChatPanel) Session() *Session {
	return cp.session
}

func (cp *ChatPanel) SetColors(c ChatPanelColors) {
	cp.colors = c
}

func (cp *ChatPanel) SetSize(width, height int) {
	if cp.width != width {
		// Width changed — invalidate all line count caches
		for i := range cp.session.Messages {
			cp.session.Messages[i].lineCount = 0
		}
	}
	cp.width = width
	cp.height = height
}

func (cp *ChatPanel) SetMaxVisibleLines(n int) {
	cp.maxVisibleLines = n
}

func (cp *ChatPanel) ScrollUp(lines int) {
	cp.scrollOffset -= lines
	if cp.scrollOffset < 0 {
		cp.scrollOffset = 0
	}
}

func (cp *ChatPanel) ScrollDown(lines int) {
	cp.scrollOffset += lines
}

func (cp *ChatPanel) ScrollToBottom() {
	cp.scrollOffset = 999999
}

func (cp *ChatPanel) ScrollToTop() {
	cp.scrollOffset = 0
}

// Search methods
func (cp *ChatPanel) EnterSearch() {
	cp.searchMode = true
	cp.searchQuery = ""
	cp.searchMatches = nil
	cp.searchIdx = -1
}

func (cp *ChatPanel) ExitSearch() {
	cp.searchMode = false
	cp.searchQuery = ""
	cp.searchMatches = nil
	cp.searchIdx = -1
}

func (cp *ChatPanel) IsSearchMode() bool {
	return cp.searchMode
}

func (cp *ChatPanel) SetSearchQuery(q string) {
	cp.searchQuery = q
	cp.searchMatches = nil
	cp.searchIdx = -1
	if q == "" || cp.session == nil {
		return
	}
	lower := strings.ToLower(q)
	for i, msg := range cp.session.Messages {
		if strings.Contains(strings.ToLower(msg.Content), lower) {
			cp.searchMatches = append(cp.searchMatches, i)
		}
	}
	if len(cp.searchMatches) > 0 {
		cp.searchIdx = 0
		cp.scrollToMessage(cp.searchMatches[0])
	}
}

func (cp *ChatPanel) NextMatch() {
	if len(cp.searchMatches) == 0 {
		return
	}
	cp.searchIdx = (cp.searchIdx + 1) % len(cp.searchMatches)
	cp.scrollToMessage(cp.searchMatches[cp.searchIdx])
}

func (cp *ChatPanel) PrevMatch() {
	if len(cp.searchMatches) == 0 {
		return
	}
	cp.searchIdx--
	if cp.searchIdx < 0 {
		cp.searchIdx = len(cp.searchMatches) - 1
	}
	cp.scrollToMessage(cp.searchMatches[cp.searchIdx])
}

func (cp *ChatPanel) scrollToMessage(msgIdx int) {
	// Calculate line offset up to this message
	offset := 0
	for i := 0; i < msgIdx && i < len(cp.session.Messages); i++ {
		offset += cp.messageLineCount(&cp.session.Messages[i])
	}
	cp.scrollOffset = offset
}

// messageLineCount returns cached or computed line count for a message
func (cp *ChatPanel) messageLineCount(msg *Message) int {
	if msg.lineCount > 0 {
		return msg.lineCount
	}
	msg.lineCount = cp.computeLineCount(msg)
	return msg.lineCount
}

func (cp *ChatPanel) computeLineCount(msg *Message) int {
	var sb strings.Builder
	cp.renderMessage(&sb, *msg)
	return len(strings.Split(sb.String(), "\n"))
}

// View renders the chat panel with viewport, folding, and scrollbar
func (cp *ChatPanel) View() string {
	contentWidth := cp.width - 2 // reserve 1 col for scrollbar + 1 for padding
	if contentWidth < 10 {
		contentWidth = 10
	}

	style := lipgloss.NewStyle().
		Width(cp.width).
		Height(cp.height).
		Background(lipgloss.Color(cp.colors.Background))

	if cp.session == nil || len(cp.session.Messages) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(cp.colors.TextMuted)).
			Padding(2, 0)
		return style.Render(emptyStyle.Render("开始新对话..."))
	}

	// Build message line offsets
	type msgRange struct {
		msgIdx     int
		startLine  int
		endLine    int
	}
	var msgRanges []msgRange
	totalLines := 0
	for i := range cp.session.Messages {
		lc := cp.messageLineCount(&cp.session.Messages[i])
		msgRanges = append(msgRanges, msgRange{
			msgIdx:    i,
			startLine: totalLines,
			endLine:   totalLines + lc,
		})
		totalLines += lc
	}

	// Clamp scroll offset
	viewportLines := cp.height - 2 // padding
	if viewportLines < 1 {
		viewportLines = 1
	}
	maxScroll := totalLines - viewportLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if cp.scrollOffset > maxScroll {
		cp.scrollOffset = maxScroll
	}
	if cp.scrollOffset < 0 {
		cp.scrollOffset = 0
	}

	visibleStart := cp.scrollOffset
	visibleEnd := visibleStart + viewportLines

	// Render only visible messages
	var content strings.Builder
	for _, mr := range msgRanges {
		if mr.endLine <= visibleStart {
			continue // above viewport
		}
		if mr.startLine >= visibleEnd {
			break // below viewport
		}
		cp.renderMessage(&content, cp.session.Messages[mr.msgIdx])
	}

	rendered := content.String()
	lines := strings.Split(rendered, "\n")
	// Remove trailing empty line from Split
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Find the first rendered message's start line
	firstRenderedStart := 0
	for _, mr := range msgRanges {
		if mr.endLine > visibleStart {
			firstRenderedStart = mr.startLine
			break
		}
	}

	// Slice to visible portion relative to first rendered message
	start := visibleStart - firstRenderedStart
	if start < 0 {
		start = 0
	}
	end := start + viewportLines
	if end > len(lines) {
		end = len(lines)
	}
	if start > end {
		start = end
	}
	visibleLines := lines[start:end]

	// Pad to viewport height
	for len(visibleLines) < viewportLines {
		visibleLines = append(visibleLines, "")
	}

	// Build scrollbar
	scrollbar := cp.buildScrollbar(totalLines, viewportLines)

	// Combine content + scrollbar
	var result strings.Builder
	for i, line := range visibleLines {
		// Pad content to contentWidth
		lineRunes := []rune(line)
		padded := line
		if len(lineRunes) < contentWidth {
			padded = line + strings.Repeat(" ", contentWidth-len(lineRunes))
		} else if len(lineRunes) > contentWidth {
			padded = string(lineRunes[:contentWidth])
		}
		if i < len(scrollbar) {
			result.WriteString(padded + scrollbar[i] + "\n")
		} else {
			result.WriteString(padded + " \n")
		}
	}

	return style.Padding(1, 1).Render(result.String())
}

func (cp *ChatPanel) buildScrollbar(totalLines, viewportLines int) []string {
	if totalLines <= viewportLines {
		return nil
	}
	bar := make([]string, viewportLines)
	for i := range bar {
		bar[i] = " "
	}

	thumbHeight := viewportLines * viewportLines / totalLines
	if thumbHeight < 1 {
		thumbHeight = 1
	}

	maxScroll := totalLines - viewportLines
	thumbPos := 0
	if maxScroll > 0 {
		thumbPos = cp.scrollOffset * (viewportLines - thumbHeight) / maxScroll
	}
	if thumbPos+thumbHeight > viewportLines {
		thumbPos = viewportLines - thumbHeight
	}

	thumbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))

	for i := 0; i < viewportLines; i++ {
		if i >= thumbPos && i < thumbPos+thumbHeight {
			bar[i] = thumbStyle.Render("█")
		} else {
			bar[i] = trackStyle.Render("░")
		}
	}
	return bar
}

// hasCJK returns true if the string contains any CJK characters
func hasCJK(s string) bool {
	for _, r := range s {
		if (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified
			(r >= 0x3400 && r <= 0x4DBF) || // CJK Ext-A
			(r >= 0x20000 && r <= 0x2A6DF) || // CJK Ext-B
			(r >= 0xF900 && r <= 0xFAFF) || // CJK Compatibility
			(r >= 0x3040 && r <= 0x309F) || // Hiragana
			(r >= 0x30A0 && r <= 0x30FF) || // Katakana
			(r >= 0xAC00 && r <= 0xD7AF) { // Hangul
			return true
		}
	}
	return false
}

func (cp *ChatPanel) renderMessage(sb *strings.Builder, msg Message) {
	ts := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cp.colors.Timestamp)).
		Render(msg.Timestamp.Format("15:04"))

	switch msg.Role {
	case RoleUser:
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(cp.colors.UserFg)).
			Foreground(lipgloss.Color("#1e1e1e")).
			Bold(true).
			Padding(0, 1).
			Render("🗣️")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")

		textStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(cp.colors.Text)).
			PaddingLeft(3)
		rendered := cp.highlightSafe(msg.Content)
		if msg.Collapsed {
			rendered = cp.foldContent(rendered, textStyle)
		} else {
			rendered = textStyle.Render(rendered)
		}
		sb.WriteString(rendered)
		sb.WriteString("\n\n")

	case RoleAssistant:
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(cp.colors.AssistantBg)).
			Foreground(lipgloss.Color(cp.colors.AssistantFg)).
			Bold(true).
			Padding(0, 1).
			Render("👾 Chan")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")

		if msg.Thinking != nil {
			cp.renderThinking(sb, msg.Thinking)
		}

		textStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(cp.colors.Text)).
			PaddingLeft(1)
		rendered := cp.highlightSafe(msg.Content)
		if msg.Collapsed {
			rendered = cp.foldContent(rendered, textStyle)
		} else {
			rendered = textStyle.Render(rendered)
		}
		sb.WriteString(rendered)
		sb.WriteString("\n\n")

	case RoleSystem:
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(cp.colors.SystemBg)).
			Foreground(lipgloss.Color(cp.colors.SystemFg)).
			Bold(true).
			Padding(0, 1).
			Render("⚙ sys")
		sb.WriteString(badge)
		sb.WriteString(" ")
		sb.WriteString(ts)
		sb.WriteString("\n")

		textStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(cp.colors.Text)).
			PaddingLeft(1)
		if msg.Collapsed {
			sb.WriteString(cp.foldContent(msg.Content, textStyle))
		} else {
			sb.WriteString(textStyle.Render(msg.Content))
		}
		sb.WriteString("\n\n")
	}
}

// foldContent shows first 3 lines + collapse indicator
func (cp *ChatPanel) foldContent(content string, style lipgloss.Style) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)
	if totalLines <= 3 {
		return style.Render(content)
	}
	preview := strings.Join(lines[:3], "\n")
	indicator := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cp.colors.TextMuted)).
		Render(fmt.Sprintf("▼ [展开 %d 行]", totalLines-3))
	return style.Render(preview) + "\n" + indicator
}

// highlightSafe applies syntax highlighting, skipping chroma for CJK text to avoid duplication
func (cp *ChatPanel) highlightSafe(content string) string {
	if hasCJK(content) {
		return content
	}
	return syntax.Highlight(content, "")
}

func (cp *ChatPanel) renderThinking(sb *strings.Builder, t *Thinking) {
	arrow := "▶"
	if t.Expanded {
		arrow = "▼"
	}

	charCount := utf8.RuneCountInString(t.Content)
	durationStr := ""
	if t.Duration > 0 {
		durationStr = fmt.Sprintf(" · %.1fs", t.Duration.Seconds())
	}

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cp.colors.ThinkingFg)).
		Render(fmt.Sprintf("%s thinking %d chars%s", arrow, charCount, durationStr))
	sb.WriteString(header)
	sb.WriteString("\n")

	if t.Expanded {
		contentStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(cp.colors.ThinkingBg)).
			BorderLeft(true).
			BorderStyle(lipgloss.Border{Left: "│"}).
			BorderForeground(lipgloss.Color(cp.colors.ThinkingBorder)).
			Padding(0, 1).
			MarginLeft(1).
			Width(cp.width - 8)

		sb.WriteString(contentStyle.Render(t.Content))
		sb.WriteString("\n")
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
